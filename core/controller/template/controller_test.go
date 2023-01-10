package template

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/url"
	"os"
	"reflect"
	"regexp"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/horizoncd/horizon/core/common"
	herrors "github.com/horizoncd/horizon/core/errors"
	"github.com/horizoncd/horizon/lib/orm"
	gitmock "github.com/horizoncd/horizon/mock/pkg/git"
	groupmanagermock "github.com/horizoncd/horizon/mock/pkg/group/manager"
	membermock "github.com/horizoncd/horizon/mock/pkg/member/manager"
	tmock "github.com/horizoncd/horizon/mock/pkg/template/manager"
	releasemanagermock "github.com/horizoncd/horizon/mock/pkg/templaterelease/manager"
	trmock "github.com/horizoncd/horizon/mock/pkg/templaterelease/manager"
	trschemamock "github.com/horizoncd/horizon/mock/pkg/templaterelease/schema"
	amodels "github.com/horizoncd/horizon/pkg/application/models"
	userauth "github.com/horizoncd/horizon/pkg/authentication/user"
	cmodels "github.com/horizoncd/horizon/pkg/cluster/models"
	gitconfig "github.com/horizoncd/horizon/pkg/config/git"
	config "github.com/horizoncd/horizon/pkg/config/templaterepo"
	hctx "github.com/horizoncd/horizon/pkg/context"
	perror "github.com/horizoncd/horizon/pkg/errors"
	"github.com/horizoncd/horizon/pkg/git"
	"github.com/horizoncd/horizon/pkg/git/gitlab"
	groupmodels "github.com/horizoncd/horizon/pkg/group/models"
	membermodels "github.com/horizoncd/horizon/pkg/member/models"
	memberservice "github.com/horizoncd/horizon/pkg/member/service"
	"github.com/horizoncd/horizon/pkg/param/managerparam"
	roleservice "github.com/horizoncd/horizon/pkg/rbac/role"
	"github.com/horizoncd/horizon/pkg/server/global"
	"github.com/horizoncd/horizon/pkg/template/models"
	tmodels "github.com/horizoncd/horizon/pkg/template/models"
	trmodels "github.com/horizoncd/horizon/pkg/templaterelease/models"
	trschema "github.com/horizoncd/horizon/pkg/templaterelease/schema"
	reposchema "github.com/horizoncd/horizon/pkg/templaterelease/schema/repo"
	"github.com/horizoncd/horizon/pkg/templaterepo/chartmuseumbase"
	usermodels "github.com/horizoncd/horizon/pkg/user/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

var (
	ctx = context.Background()
	db  *gorm.DB
	mgr *managerparam.Manager
)

func testList(t *testing.T) {
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
	templateMgr.EXPECT().ListTemplate(gomock.Any()).Return([]*models.Template{
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
			Repository: "https://cloudnative.com/music-cloud-native/horizon/horizon.git",
		}, nil,
	)
	templateMgr.EXPECT().GetByName(gomock.Any(), "javaapp").Return(
		&tmodels.Template{
			Model:      global.Model{ID: 1},
			Name:       "javaapp",
			OnlyOwner:  &onlyOwnerTrue,
			ChartName:  "7-javaapp_test3",
			Repository: "https://cloudnative.com/music-cloud-native/horizon/horizon",
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

	gitlabLib := gitmock.NewMockHelper(mockCtl)

	gitlabLib.EXPECT().GetTagArchive(gomock.Any(), gomock.Any(), "v1.0.2").
		Return(&git.Tag{ShortID: "test3", Name: "", ArchiveData: []byte{}}, nil).Times(2)
	gitlabLib.EXPECT().GetTagArchive(gomock.Any(), gomock.Any(), "v1.0.1").
		Return(&git.Tag{ShortID: "test", Name: "", ArchiveData: []byte{}}, nil).Times(2)
	gitlabLib.EXPECT().GetTagArchive(gomock.Any(), gomock.Any(), "v1.0.0").
		Return(&git.Tag{ShortID: "test", Name: "", ArchiveData: []byte{}}, nil).Times(2)

	ctl := &controller{
		templateMgr:        templateMgr,
		templateReleaseMgr: templateReleaseMgr,
		groupMgr:           groupMgr,
		gitgetter:          gitlabLib,
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

	currentUser, err := common.UserFromContext(ctx)
	assert.Nil(t, err)
	ctx = common.WithContext(ctx, &userauth.DefaultInfo{
		Name: currentUser.GetName(),
		ID:   currentUser.GetID(),
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
		MemberNameID: currentUser.GetID(),
	})
	assert.Nil(t, err)
	assert.NotNil(t, m)

	ctx = context.WithValue(ctx, hctx.MemberDirectMemberOnly, true)
	templateReleases, err = ctl.ListTemplateRelease(ctx, "javaapp")
	assert.Nil(t, err)
	assert.Equal(t, 3, len(templateReleases))
}

func testListTemplates(t *testing.T) {
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

func testGetSchema(t *testing.T) {
	createContext()
	charName := repoConfig.TemplateName

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
		Name:         repoConfig.TemplateTag,
		TemplateName: repoConfig.TemplateName,
		ChartVersion: repoConfig.TemplateTag,
		ChartName:    charName,
	}
	templateReleaseMgr.EXPECT().GetByID(gomock.Any(), uint(1)).Return(release, nil)
	templateSchemaGetter.EXPECT().GetTemplateSchema(ctx,
		repoConfig.TemplateName, repoConfig.TemplateTag, nil).Return(schemas, nil)

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

func testCreateTemplate(t *testing.T) {
	createContext()
	ctl := createController(t)

	var err error

	createChart(t, ctl, 0)
	assert.Nil(t, err)
	defer func() { _ = ctl.DeleteRelease(ctx, 1) }()

	template, err := ctl.GetTemplate(ctx, 1)
	assert.Nil(t, err)
	assert.Equal(t, repoConfig.TemplateName, template.Name)
	assert.Equal(t, repoConfig.TemplateRepo, template.Repository)
	assert.Equal(t, uint(0), template.GroupID)

	tpl, err := mgr.TemplateMgr.GetByID(ctx, 1)
	assert.Nil(t, err)
	assert.Equal(t, repoConfig.TemplateName, tpl.Name)
	assert.Equal(t, uint(1), tpl.ID)

	releases, err := ctl.ListTemplateReleaseByTemplateID(ctx, 1)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(releases))
	assert.Equal(t, repoConfig.TemplateTag, releases[0].Name)
	versionPattern := regexp.MustCompile(`^v(\d\.){2}\d-(.+)$`)
	assert.True(t, versionPattern.MatchString(releases[0].ChartVersion))

	err = ctl.SyncReleaseToRepo(ctx, 1)
	assert.Nil(t, err)
}

func testCreateTemplateInNonRootGroup(t *testing.T) {
	createContext()
	ctl := createController(t)

	var err error

	ctx = context.WithValue(ctx, hctx.ReleaseSyncToRepo, false)
	createChart(t, ctl, 1)

	template, err := ctl.GetTemplate(ctx, 1)
	assert.Nil(t, err)
	assert.Equal(t, repoConfig.TemplateName, template.Name)
	assert.Equal(t, repoConfig.TemplateRepo, template.Repository)
	assert.Equal(t, uint(1), template.GroupID)

	tpl, err := mgr.TemplateMgr.GetByID(ctx, 1)
	assert.Nil(t, err)
	assert.Equal(t, repoConfig.TemplateName, tpl.Name)
	assert.Equal(t, uint(1), tpl.ID)

	release, err := mgr.TemplateReleaseManager.GetByID(ctx, 1)
	assert.Nil(t, err)
	assert.Equal(t, repoConfig.TemplateTag, release.Name)
	assert.Equal(t, repoConfig.TemplateName, release.TemplateName)
	assert.Equal(t, tpl.ID, release.Template)
}

func testDeleteTemplate(t *testing.T) {
	createContext()
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

func testGetTemplate(t *testing.T) {
	createContext()
	ctl := createController(t)

	createChart(t, ctl, 0)
	defer func() { _ = ctl.DeleteRelease(ctx, 1) }()

	schemas, err := ctl.GetTemplateSchema(ctx, 1, map[string]string{})
	assert.Nil(t, err)
	assert.NotNil(t, schemas)
}

func testUpdateTemplate(t *testing.T) {
	createContext()
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
		Repository:  repoConfig.TemplateRepo,
		OnlyOwner:   onlyOwnerTrue,
	}
	err = ctl.UpdateTemplate(ctx, 1, tplRequest)
	assert.Nil(t, err)

	oldDescription := tplRequest.Description
	tplRequest.Description = ""
	tplRequest.Repository = repoConfig.TemplateRepo

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

func testListTemplate(t *testing.T) {
	createContext()
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

	templates, err = ctl.ListTemplateByGroupID(ctx, 1, true)
	_, ok = perror.Cause(err).(*herrors.HorizonErrNotFound)
	assert.True(t, ok)
	assert.Equal(t, 0, len(templates))

	templates, err = ctl.ListTemplateByGroupID(ctx, 0, true)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(templates))

	releases, err = ctl.ListTemplateReleaseByTemplateID(ctx, 1)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(releases))
}

const (
	EnvTemplateRepos = "TEMPLATE_REPOS"
)

type RepoConfig struct {
	Kind              string `json:"kind"`
	Host              string `json:"host"`
	Passwd            string `json:"passwd"`
	RepoName          string `json:"repoName"`
	Username          string `json:"username"`
	TemplateName      string `json:"templateName"`
	TemplateRepo      string `json:"templateRepo"`
	TemplateRepoToken string `json:"templateRepoToken"`
	TemplateTag       string `json:"templateTag"`
}

var repoConfig *RepoConfig

func Test(t *testing.T) {
	templateRepos := os.Getenv(EnvTemplateRepos)
	if templateRepos == "" {
		return
	}

	configs := make([]RepoConfig, 0)

	if err := json.Unmarshal([]byte(templateRepos), &configs); err != nil {
		panic(err)
	}

	for _, cfg := range configs {
		repoConfig = &cfg

		t.Run(fmt.Sprintf("TestList_%s", repoConfig.Kind), testList)
		t.Run(fmt.Sprintf("TestListTemplates_%s", repoConfig.Kind), testListTemplates)
		t.Run(fmt.Sprintf("TestGetSchema_%s", repoConfig.Kind), testGetSchema)
		t.Run(fmt.Sprintf("TestCreateTemplate_%s", repoConfig.Kind), testCreateTemplate)
		t.Run(fmt.Sprintf("TestCreateTemplateInNonRootGroup_%s", repoConfig.Kind), testCreateTemplateInNonRootGroup)
		t.Run(fmt.Sprintf("TestDeleteTemplate_%s", repoConfig.Kind), testDeleteTemplate)
		t.Run(fmt.Sprintf("TestGetTemplate_%s", repoConfig.Kind), testGetTemplate)
		t.Run(fmt.Sprintf("TestUpdateTemplate_%s", repoConfig.Kind), testUpdateTemplate)
		t.Run(fmt.Sprintf("TestListTemplate_%s", repoConfig.Kind), testListTemplate)
	}
}

func createContext() {
	db, _ = orm.NewSqliteDB("")
	if err := db.AutoMigrate(&trmodels.TemplateRelease{},
		&amodels.Application{}, &cmodels.Cluster{}, &membermodels.Member{},
		&tmodels.Template{}, &membermodels.Member{}, &groupmodels.Group{}, &usermodels.User{}); err != nil {
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

	currentUser := usermodels.User{
		Name: "Jerry",
		Model: global.Model{
			ID: 1,
		},
	}
	_, err := mgr.UserManager.Create(ctx, &currentUser)
	if err != nil {
		panic(err)
	}
}

func createController(t *testing.T) Controller {
	repo, err := chartmuseumbase.NewRepo(config.Repo{
		Kind:     repoConfig.Kind,
		Host:     repoConfig.Host,
		Username: repoConfig.Username,
		Password: repoConfig.Passwd,
		Insecure: true,
		CertFile: "",
		KeyFile:  "",
		CAFile:   "",
		RepoName: repoConfig.RepoName,
	})
	assert.Nil(t, err)

	getter := reposchema.NewSchemaGetter(context.Background(), repo, mgr)

	URL, err := url.Parse(repoConfig.TemplateRepo)
	assert.Nil(t, err)

	gitlabLib, err := gitlab.New(ctx, &gitconfig.Repo{Kind: "gitlab",
		URL: fmt.Sprintf("%s://%s", URL.Scheme, URL.Host), Token: repoConfig.TemplateRepoToken})
	assert.Nil(t, err)

	ctl := &controller{
		gitgetter:            gitlabLib,
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
			Name: repoConfig.TemplateTag,
		},
		Name:        repoConfig.TemplateName,
		Description: "",
		Repository:  repoConfig.TemplateRepo,
	}
	template, err := ctl.CreateTemplate(ctx, groupID, request)
	assert.Nil(t, err)
	_, err = ctl.CreateRelease(ctx, template.ID, request.CreateReleaseRequest)
	assert.Nil(t, err)
}
