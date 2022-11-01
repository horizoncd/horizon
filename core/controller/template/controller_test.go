package template

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/url"
	"os"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"g.hz.netease.com/horizon/core/common"
	herrors "g.hz.netease.com/horizon/core/errors"
	"g.hz.netease.com/horizon/lib/gitlab"
	"g.hz.netease.com/horizon/lib/orm"
	gitlablibmock "g.hz.netease.com/horizon/mock/lib/gitlab"
	groupmanagermock "g.hz.netease.com/horizon/mock/pkg/group/manager"
	membermock "g.hz.netease.com/horizon/mock/pkg/member/manager"
	tmock "g.hz.netease.com/horizon/mock/pkg/template/manager"
	releasemanagermock "g.hz.netease.com/horizon/mock/pkg/templaterelease/manager"
	trmock "g.hz.netease.com/horizon/mock/pkg/templaterelease/manager"
	trschemamock "g.hz.netease.com/horizon/mock/pkg/templaterelease/schema"
	amodels "g.hz.netease.com/horizon/pkg/application/models"
	userauth "g.hz.netease.com/horizon/pkg/authentication/user"
	cmodels "g.hz.netease.com/horizon/pkg/cluster/models"
	config "g.hz.netease.com/horizon/pkg/config/templaterepo"
	hctx "g.hz.netease.com/horizon/pkg/context"
	perror "g.hz.netease.com/horizon/pkg/errors"
	groupmodels "g.hz.netease.com/horizon/pkg/group/models"
	membermodels "g.hz.netease.com/horizon/pkg/member/models"
	memberservice "g.hz.netease.com/horizon/pkg/member/service"
	"g.hz.netease.com/horizon/pkg/param/managerparam"
	roleservice "g.hz.netease.com/horizon/pkg/rbac/role"
	"g.hz.netease.com/horizon/pkg/server/global"
	"g.hz.netease.com/horizon/pkg/template/models"
	tmodels "g.hz.netease.com/horizon/pkg/template/models"
	trmodels "g.hz.netease.com/horizon/pkg/templaterelease/models"
	trschema "g.hz.netease.com/horizon/pkg/templaterelease/schema"
	reposchema "g.hz.netease.com/horizon/pkg/templaterelease/schema/repo"
	"g.hz.netease.com/horizon/pkg/templaterepo/harbor"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	gitlabapi "github.com/xanzy/go-gitlab"
	"gorm.io/gorm"
)

var (
	ctx = context.Background()
	db  *gorm.DB
	mgr *managerparam.Manager
)

func TestList(t *testing.T) {
	createContext()

	mockCtl := gomock.NewController(t)
	templateMgr := tmock.NewMockManager(mockCtl)
	templateReleaseMgr := trmock.NewMockManager(mockCtl)
	groupMgr := groupmanagermock.NewMockManager(mockCtl)
	memberMgr := membermock.NewMockManager(mockCtl)

	bts, err := ioutil.ReadFile("../../../roles.yaml")
	if err != nil {
		panic(err)
	}
	roleService, err := roleservice.NewFileRole(context.Background(), bytes.NewReader(bts))
	if err != nil {
		panic(err)
	}
	memberService := memberservice.NewService(roleService, nil, mgr)
	if err != nil {
		panic(err)
	}

	onlyOwnerTrue := true
	onlyOwnerFalse := false
	//        G0
	//       /  \
	//      G1   T1
	//     /
	//    T2
	templateMgr.EXPECT().List(gomock.Any()).Return([]*models.Template{
		{
			Model: global.Model{
				ID: 1,
			},
			GroupID:   0,
			Name:      "javaapp",
			OnlyOwner: &onlyOwnerTrue,
		}, {
			Model: global.Model{
				ID: 2,
			},
			GroupID:   1,
			Name:      "tomcat",
			OnlyOwner: &onlyOwnerFalse,
		},
	}, nil).Times(2)
	group1 := &groupmodels.Group{
		Model:        global.Model{ID: 1},
		Name:         "group1",
		Path:         "group1",
		TraversalIDs: "1",
	}
	groupMgr.EXPECT().GetByID(gomock.Any(), uint(1)).Return(group1, nil)
	groupMgr.EXPECT().GetByIDs(gomock.Any(), []uint{1}).Return([]*groupmodels.Group{group1}, nil)

	recommends := []bool{false, true, false}

	templateMgr.EXPECT().GetByName(gomock.Any(), "javaapp").Return(
		&tmodels.Template{
			Name:       "javaapp",
			ChartName:  "7-javaapp_test3",
			Repository: "https://g.hz.netease.com/music-cloud-native/horizon/horizon.git",
		}, nil,
	)
	templateMgr.EXPECT().GetByName(gomock.Any(), "javaapp").Return(
		&tmodels.Template{
			Model:      global.Model{ID: 1},
			Name:       "javaapp",
			OnlyOwner:  &onlyOwnerTrue,
			ChartName:  "7-javaapp_test3",
			Repository: "https://g.hz.netease.com/music-cloud-native/horizon/horizon",
		}, nil,
	)

	tags := []string{"v1.0.0", "v1.0.1", "v1.0.2"}
	templateReleaseMgr.EXPECT().ListByTemplateName(gomock.Any(), "javaapp").
		Return([]*trmodels.TemplateRelease{
			{
				Model: global.Model{
					ID: 1,
				},
				Template:    1,
				Name:        tags[0],
				CommitID:    "test",
				SyncStatus:  trmodels.StatusSucceed,
				Recommended: &recommends[0],
				OnlyOwner:   &onlyOwnerTrue,
			}, {
				Model: global.Model{
					ID: 1,
				},
				Template:    1,
				Name:        tags[1],
				CommitID:    "test",
				SyncStatus:  trmodels.StatusSucceed,
				Recommended: &recommends[1],
				OnlyOwner:   &onlyOwnerFalse,
			}, {
				Model: global.Model{
					ID: 1,
				},
				Template:    1,
				Name:        tags[2],
				CommitID:    "test3",
				SyncStatus:  trmodels.StatusSucceed,
				Recommended: &recommends[2],
				OnlyOwner:   &onlyOwnerFalse,
			},
		}, nil).Times(2)

	templateReleaseMgr.EXPECT().UpdateByID(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	gitlabLib := gitlablibmock.NewMockInterface(mockCtl)

	gitlabLib.EXPECT().
		GetRepositoryArchive(gomock.Any(), gomock.Any(), gomock.Any()).Return([]byte{}, nil).Times(6)

	gitlabLib.EXPECT().GetTag(gomock.Any(), "music-cloud-native/horizon/horizon", tags[0]).
		Return(&gitlabapi.Tag{
			Commit: &gitlabapi.Commit{ShortID: "test"},
		}, nil)
	gitlabLib.EXPECT().GetTag(gomock.Any(), "music-cloud-native/horizon/horizon", tags[1]).
		Return(&gitlabapi.Tag{
			Commit: &gitlabapi.Commit{ShortID: "test"},
		}, nil)
	gitlabLib.EXPECT().GetTag(gomock.Any(), "music-cloud-native/horizon/horizon", tags[2]).
		Return(&gitlabapi.Tag{
			Commit: &gitlabapi.Commit{ShortID: "test3"},
		}, nil).Times(1)

	gitlabLib.EXPECT().GetTag(gomock.Any(), "music-cloud-native/horizon/horizon", tags[0]).
		Return(&gitlabapi.Tag{
			Commit: &gitlabapi.Commit{ShortID: "test"},
		}, nil)
	gitlabLib.EXPECT().GetTag(gomock.Any(), "music-cloud-native/horizon/horizon", tags[1]).
		Return(&gitlabapi.Tag{
			Commit: &gitlabapi.Commit{ShortID: "test"},
		}, nil)
	gitlabLib.EXPECT().GetTag(gomock.Any(), "music-cloud-native/horizon/horizon", tags[2]).
		Return(nil, errors.New("test")).Times(1)

	ctl := &controller{
		templateMgr:        templateMgr,
		templateReleaseMgr: templateReleaseMgr,
		groupMgr:           groupMgr,
		gitlabLib:          gitlabLib,
		memberSvc:          memberService,
		memberMgr:          memberMgr,
	}

	c := context.WithValue(ctx, hctx.TemplateWithFullPath, true)
	templates, err := ctl.ListTemplate(c)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(templates))
	assert.Equal(t, "javaapp", templates[0].Name)
	assert.Equal(t, "/javaapp", templates[0].FullPath)
	assert.Equal(t, "tomcat", templates[1].Name)
	assert.Equal(t, "/group1/tomcat", templates[1].FullPath)

	templateReleases, err := ctl.ListTemplateRelease(ctx, "javaapp")
	assert.Nil(t, err)
	assert.Equal(t, 3, len(templateReleases))
	assert.Equal(t, "v1.0.1", templateReleases[0].Name)
	assert.Equal(t, "v1.0.2", templateReleases[1].Name)
	assert.Equal(t, "v1.0.0", templateReleases[2].Name)
	assert.Equal(t, uint8(trmodels.StatusSucceed), templateReleases[0].SyncStatusCode)
	assert.Equal(t, uint8(trmodels.StatusSucceed), templateReleases[1].SyncStatusCode)
	assert.Equal(t, uint8(trmodels.StatusSucceed), templateReleases[2].SyncStatusCode)

	ctx = common.WithContext(ctx, &userauth.DefaultInfo{
		Name:  "Jerry",
		ID:    1,
		Admin: false,
	})

	templates, err = ctl.ListTemplate(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(templates))
	assert.Equal(t, "tomcat", templates[0].Name)

	m, err := mgr.MemberManager.Create(ctx, &membermodels.Member{
		ResourceType: common.ResourceTemplate,
		ResourceID:   1,
		Role:         roleservice.Owner,
		MemberType:   membermodels.MemberUser,
		MemberNameID: 1,
	})
	assert.Nil(t, err)
	assert.NotNil(t, m)

	ctx = context.WithValue(ctx, hctx.MemberDirectMemberOnly, true)
	templateReleases, err = ctl.ListTemplateRelease(ctx, "javaapp")
	assert.Nil(t, err)
	assert.Equal(t, 3, len(templateReleases))
}

func TestListTemplates(t *testing.T) {
	createContext()

	mockCtl := gomock.NewController(t)
	groupMgr := groupmanagermock.NewMockManager(mockCtl)
	memberMgr := membermock.NewMockManager(mockCtl)
	mgr.MemberManager = memberMgr
	mgr.GroupManager = groupMgr

	ctrl := &controller{
		groupMgr:           groupMgr,
		templateMgr:        mgr.TemplateMgr,
		templateReleaseMgr: mgr.TemplateReleaseManager,
		memberMgr:          memberMgr,
	}

	ctx = common.WithContext(ctx, &userauth.DefaultInfo{
		Name:  "Jerry",
		ID:    1,
		Admin: false,
	})

	//        G0
	//       /  \
	//      G1   T1
	//     /
	//    T2
	tpl1 := &models.Template{
		Model: global.Model{
			ID: 1,
		},
		GroupID: 0,
		Name:    "javaapp",
	}
	tpl2 := &models.Template{
		Model: global.Model{
			ID: 2,
		},
		GroupID: 1,
		Name:    "tomcat",
	}
	tpl1, err := mgr.TemplateMgr.Create(ctx, tpl1)
	assert.Nil(t, err)
	assert.NotNil(t, tpl1)
	tpl2, err = mgr.TemplateMgr.Create(ctx, tpl2)
	assert.Nil(t, err)
	assert.NotNil(t, tpl2)

	// for Jerry is owner of group 1,
	// result should be "tpl2"
	groupID := uint(1)
	userID := uint(1)
	memberMgr.EXPECT().
		ListResourceOfMemberInfoByRole(gomock.Any(), membermodels.TypeGroup, userID, roleservice.Owner).
		Return([]uint{1}, nil).Times(1)
	groupMgr.EXPECT().GetSubGroupsByGroupIDs(gomock.Any(), []uint{1}).
		Return([]*groupmodels.Group{{Model: global.Model{ID: 1}}}, nil).Times(1)
	memberMgr.EXPECT().Get(gomock.Any(), membermodels.TypeGroup,
		groupID, membermodels.MemberUser, userID).
		Return(&membermodels.Member{Role: roleservice.Owner}, nil).Times(1)
	memberMgr.EXPECT().
		ListResourceOfMemberInfoByRole(gomock.Any(), membermodels.TypeTemplate, userID, roleservice.Owner).
		Return([]uint{}, nil).Times(1)

	ctx = context.WithValue(ctx, hctx.TemplateListSelfOnly, true)
	templates, err := ctrl.ListTemplate(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(templates))
	assert.Equal(t, tpl2.ID, templates[0].ID)
	assert.Equal(t, tpl2.Name, templates[0].Name)
}

func TestGetSchema(t *testing.T) {
	createContext()
	groupID := 0
	charName := fmt.Sprintf(ChartNameFormat, groupID, templateName)

	mockCtl := gomock.NewController(t)
	// templateMgr := tmock.NewMockManager(mockCtl)
	templateReleaseMgr := releasemanagermock.NewMockManager(mockCtl)
	templateSchemaGetter := trschemamock.NewMockGetter(mockCtl)
	schema := map[string]interface{}{
		"type": "object",
	}
	schemas := &trschema.Schemas{
		Application: &trschema.Schema{
			JSONSchema: schema,
			UISchema:   schema,
		},
		Pipeline: &trschema.Schema{
			JSONSchema: schema,
			UISchema:   schema,
		},
	}
	release := &trmodels.TemplateRelease{
		Name:         releaseName,
		TemplateName: templateName,
		ChartVersion: releaseName,
		ChartName:    charName,
	}
	templateReleaseMgr.EXPECT().GetByID(gomock.Any(), uint(1)).Return(release, nil)
	templateSchemaGetter.EXPECT().GetTemplateSchema(ctx, templateName, releaseName, nil).Return(schemas, nil)

	ctl := &controller{
		templateSchemaGetter: templateSchemaGetter,
		templateReleaseMgr:   templateReleaseMgr,
	}

	ss, err := ctl.GetTemplateSchema(ctx, 1, nil)
	assert.Nil(t, err)
	if !reflect.DeepEqual(ss, &Schemas{
		Application: &Schema{
			JSONSchema: schema,
			UISchema:   schema,
		},
		Pipeline: &Schema{
			JSONSchema: schema,
			UISchema:   schema,
		},
	}) {
		t.Fatal("not equal")
	}
}

func TestCreateTemplate(t *testing.T) {
	createContext()
	checkSkip(t)
	ctl := createController(t)

	var err error

	createChart(t, ctl, 0)
	assert.Nil(t, err)
	defer func() { _ = ctl.DeleteRelease(ctx, 1) }()

	template, err := ctl.GetTemplate(ctx, 1)
	assert.Nil(t, err)
	assert.Equal(t, templateName, template.Name)
	assert.Equal(t, templateRepo, template.Repository)
	assert.Equal(t, uint(0), template.GroupID)

	tpl, err := mgr.TemplateMgr.GetByID(ctx, 1)
	assert.Nil(t, err)
	assert.Equal(t, templateName, tpl.Name)
	assert.Equal(t, uint(1), tpl.ID)

	releases, err := ctl.ListTemplateReleaseByTemplateID(ctx, 1)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(releases))
	assert.Equal(t, releaseName, releases[0].Name)
	versionPattern := regexp.MustCompile(`^v(\d\.){2}\d-(.+)$`)
	assert.True(t, versionPattern.MatchString(releases[0].ChartVersion))

	err = ctl.SyncReleaseToRepo(ctx, 1)
	assert.Nil(t, err)
}

func TestCreateTemplateInNonRootGroup(t *testing.T) {
	createContext()
	checkSkip(t)
	ctl := createController(t)

	var err error

	ctx = context.WithValue(ctx, hctx.ReleaseSyncToRepo, false)
	createChart(t, ctl, 1)

	template, err := ctl.GetTemplate(ctx, 1)
	assert.Nil(t, err)
	assert.Equal(t, templateName, template.Name)
	assert.Equal(t, templateRepo, template.Repository)
	assert.Equal(t, uint(1), template.GroupID)

	tpl, err := mgr.TemplateMgr.GetByID(ctx, 1)
	assert.Nil(t, err)
	assert.Equal(t, templateName, tpl.Name)
	assert.Equal(t, uint(1), tpl.ID)

	release, err := mgr.TemplateReleaseManager.GetByID(ctx, 1)
	assert.Nil(t, err)
	assert.Equal(t, releaseName, release.Name)
	assert.Equal(t, templateName, release.TemplateName)
	assert.Equal(t, tpl.ID, release.Template)
}

func TestDeleteTemplate(t *testing.T) {
	createContext()
	checkSkip(t)
	ctl := createController(t)

	createChart(t, ctl, 0)

	err := ctl.DeleteTemplate(ctx, 1)
	assert.NotNil(t, err)

	err = ctl.DeleteRelease(ctx, 1)
	assert.Nil(t, err)

	release, err := mgr.TemplateReleaseManager.GetByID(ctx, 1)
	assert.NotNil(t, err)
	assert.Nil(t, release)

	schemas, err := ctl.GetTemplateSchema(ctx, 1, map[string]string{})
	assert.NotNil(t, err)
	assert.Nil(t, schemas)

	err = ctl.DeleteTemplate(ctx, 1)
	assert.Nil(t, err)

	template, err := mgr.TemplateMgr.GetByID(ctx, 1)
	assert.NotNil(t, err)
	assert.Nil(t, template)
}

func TestGetTemplate(t *testing.T) {
	createContext()
	checkSkip(t)
	ctl := createController(t)

	createChart(t, ctl, 0)
	defer func() { _ = ctl.DeleteRelease(ctx, 1) }()

	schemas, err := ctl.GetTemplateSchema(ctx, 1, map[string]string{})
	assert.Nil(t, err)
	assert.NotNil(t, schemas)
}

func TestUpdateTemplate(t *testing.T) {
	createContext()
	checkSkip(t)
	ctl := createController(t)

	ctx = context.WithValue(ctx, hctx.ReleaseSyncToRepo, false)
	createChart(t, ctl, 0)

	// nolint
	ctx = common.WithContext(ctx, &userauth.DefaultInfo{
		Name:  "Jerry",
		ID:    2,
		Admin: true,
	})

	defer func() { _ = ctl.DeleteRelease(ctx, 1) }()

	onlyOwnerTrue := true

	tplRequest := UpdateTemplateRequest{
		Name:        "javaapp",
		Description: "hello, world",
		Repository:  "repo",
		OnlyOwner:   onlyOwnerTrue,
	}
	err := ctl.UpdateTemplate(ctx, 1, tplRequest)
	assert.NotNil(t, err)

	tplRequest = UpdateTemplateRequest{
		Name:        "javaapp",
		Description: "hello, world",
		Repository:  templateRepo,
		OnlyOwner:   onlyOwnerTrue,
	}
	err = ctl.UpdateTemplate(ctx, 1, tplRequest)
	assert.Nil(t, err)

	oldDescription := tplRequest.Description
	tplRequest.Description = ""
	tplRequest.Repository = templateRepo

	err = ctl.UpdateTemplate(ctx, 1, tplRequest)
	assert.Nil(t, err)

	template, err := ctl.GetTemplate(ctx, 1)
	assert.Nil(t, err)
	assert.Equal(t, oldDescription, template.Description)
	assert.Equal(t, tplRequest.Repository, template.Repository)
	assert.Equal(t, onlyOwnerTrue, template.OnlyOwner)

	b := true
	trRequest := UpdateReleaseRequest{
		Description: "hello, world",
		Recommended: &b,
		OnlyOwner:   onlyOwnerTrue,
	}
	err = ctl.UpdateRelease(ctx, 1, trRequest)
	assert.Nil(t, err)

	release, err := ctl.GetRelease(ctx, 1)
	assert.Nil(t, err)
	assert.Equal(t, trRequest.Description, release.Description)
	assert.Equal(t, *trRequest.Recommended, release.Recommended)
	assert.Equal(t, onlyOwnerTrue, release.OnlyOwner)

	oldDescription = trRequest.Description
	trRequest.Description = ""
	trRequest.Recommended = nil

	err = ctl.UpdateRelease(ctx, 1, trRequest)
	assert.Nil(t, err)

	release, err = ctl.GetRelease(ctx, 1)
	assert.Nil(t, err)
	assert.Equal(t, oldDescription, release.Description)
	assert.Equal(t, b, release.Recommended)
	assert.Equal(t, onlyOwnerTrue, release.OnlyOwner)
}

func TestListTemplate(t *testing.T) {
	createContext()
	checkSkip(t)
	ctl := createController(t)

	templates, err := ctl.ListTemplate(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(templates))

	releases, err := ctl.ListTemplateReleaseByTemplateID(ctx, 1)

	_, ok := perror.Cause(err).(*herrors.HorizonErrNotFound)
	assert.True(t, ok)
	assert.Equal(t, 0, len(releases))

	ctx = context.WithValue(ctx, hctx.ReleaseSyncToRepo, false)
	createChart(t, ctl, 0)

	templates, err = ctl.ListTemplateByGroupID(ctx, 1, false)
	_, ok = perror.Cause(err).(*herrors.HorizonErrNotFound)
	assert.True(t, ok)
	assert.Equal(t, 0, len(templates))

	templates, err = ctl.ListTemplateByGroupID(ctx, 0, false)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(templates))

	releases, err = ctl.ListTemplateReleaseByTemplateID(ctx, 1)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(releases))

	createChart(t, ctl, 1)

	ctx = context.WithValue(ctx, hctx.TemplateListRecursively, true)
	templates, err = ctl.ListTemplateByGroupID(ctx, 0, false)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(templates))

	_, err = mgr.GroupManager.Create(ctx,
		&groupmodels.Group{
			Model:        global.Model{ID: 2},
			Name:         "test2",
			Path:         "test2",
			ParentID:     1,
			TraversalIDs: "1,2",
		})
	createChart(t, ctl, 2)

	assert.Nil(t, err)
	templates, err = ctl.ListTemplateByGroupID(ctx, 2, false)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(templates))
}

const (
	EnvHarborHost        = "HARBOR_HOST"
	EnvHarborUser        = "HARBOR_USER"
	EnvHarborPasswd      = "HARBOR_PASSWD"
	EnvHarborRepoName    = "HARBOR_REPO_NAME"
	EnvTemplateName      = "TEMPLATE_NAME"
	EnvTemplateTag       = "TEMPLATE_TAG"
	EnvTemplateRepo      = "TEMPLATE_REPO"
	EnvTemplateRepoToken = "TEMPLATE_REPO_TOKEN"
)

var (
	harborHost        string
	harborAdmin       string
	harborPasswd      string
	harborRepoName    string
	templateName      string
	releaseName       string
	templateRepo      string
	templateRepoToken string
)

func TestMain(m *testing.M) {
	harborHost = os.Getenv(EnvHarborHost)
	harborHost = strings.TrimPrefix(harborHost, "https://")
	harborHost = strings.TrimPrefix(harborHost, "http://")
	harborAdmin = os.Getenv(EnvHarborUser)
	harborPasswd = os.Getenv(EnvHarborPasswd)
	harborRepoName = os.Getenv(EnvHarborRepoName)
	templateName = os.Getenv(EnvTemplateName)
	releaseName = os.Getenv(EnvTemplateTag)
	templateRepo = os.Getenv(EnvTemplateRepo)
	templateRepoToken = os.Getenv(EnvTemplateRepoToken)

	os.Exit(m.Run())
}

func checkSkip(t *testing.T) {
	if harborHost == "" ||
		harborAdmin == "" ||
		harborPasswd == "" ||
		harborRepoName == "" ||
		templateName == "" ||
		releaseName == "" ||
		templateRepo == "" ||
		templateRepoToken == "" {
		t.Skip()
	}
}

func createContext() {
	db, _ = orm.NewSqliteDB("")
	if err := db.AutoMigrate(&trmodels.TemplateRelease{},
		&amodels.Application{}, &cmodels.Cluster{}, &membermodels.Member{},
		&tmodels.Template{}, &membermodels.Member{}, &groupmodels.Group{}); err != nil {
		panic(err)
	}
	mgr = managerparam.InitManager(db)
	ctx = context.Background()
	// nolint
	ctx = common.WithContext(ctx, &userauth.DefaultInfo{
		Name:  "Jerry",
		ID:    1,
		Admin: true,
	})
}

func createController(t *testing.T) Controller {
	repo, err := harbor.NewRepo(config.Repo{
		Scheme:   "https",
		Host:     harborHost,
		Username: harborAdmin,
		Password: harborPasswd,
		Insecure: true,
		CertFile: "",
		KeyFile:  "",
		CAFile:   "",
		RepoName: harborRepoName,
	})
	assert.Nil(t, err)

	getter := reposchema.NewSchemaGetter(context.Background(), repo, mgr)

	URL, err := url.Parse(templateRepo)
	assert.Nil(t, err)
	gitlabLib, err := gitlab.New(templateRepoToken,
		fmt.Sprintf("%s://%s", URL.Scheme, URL.Host), "")
	assert.Nil(t, err)

	ctl := &controller{
		gitlabLib:            gitlabLib,
		templateRepo:         repo,
		groupMgr:             mgr.GroupManager,
		templateMgr:          mgr.TemplateMgr,
		templateReleaseMgr:   mgr.TemplateReleaseManager,
		memberMgr:            mgr.MemberManager,
		templateSchemaGetter: getter,
	}
	return ctl
}

func createChart(t *testing.T, ctl Controller, groupID uint) {
	if groupID != 0 {
		_, err := mgr.GroupManager.GetByID(ctx, groupID)
		if err != nil {
			if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
				name := fmt.Sprintf("test-%d", rand.Uint32())
				_, err := mgr.GroupManager.Create(ctx, &groupmodels.Group{
					Name:         name,
					Path:         name,
					ParentID:     0,
					TraversalIDs: fmt.Sprintf("%d", groupID),
					CreatedBy:    0,
					UpdatedBy:    0,
				})
				assert.Nil(t, err)
			} else {
				assert.Nil(t, err)
			}
		}
	}
	request := CreateTemplateRequest{
		CreateReleaseRequest: CreateReleaseRequest{
			Name: releaseName,
		},
		Name:        templateName,
		Description: "",
		Repository:  templateRepo,
	}
	template, err := ctl.CreateTemplate(ctx, groupID, request)
	assert.Nil(t, err)
	_, err = ctl.CreateRelease(ctx, template.ID, request.CreateReleaseRequest)
	assert.Nil(t, err)
}
