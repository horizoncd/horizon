package template

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"testing"

	"g.hz.netease.com/horizon/core/common"
	herrors "g.hz.netease.com/horizon/core/errors"
	"g.hz.netease.com/horizon/lib/orm"
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
	"g.hz.netease.com/horizon/pkg/param/managerparam"
	"g.hz.netease.com/horizon/pkg/server/global"
	"g.hz.netease.com/horizon/pkg/template/models"
	tmodels "g.hz.netease.com/horizon/pkg/template/models"
	trmodels "g.hz.netease.com/horizon/pkg/templaterelease/models"
	trschema "g.hz.netease.com/horizon/pkg/templaterelease/schema"
	reposchema "g.hz.netease.com/horizon/pkg/templaterelease/schema/repo"
	"g.hz.netease.com/horizon/pkg/templaterepo/harbor"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
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

	onlyAdminTrue := true
	onlyAdminFalse := false
	templateMgr.EXPECT().List(gomock.Any()).Return([]*models.Template{
		{
			Model: global.Model{
				ID: 1,
			},
			Name:      "javaapp",
			OnlyAdmin: &onlyAdminTrue,
		}, {
			Model: global.Model{
				ID: 2,
			},
			Name:      "tomcat",
			OnlyAdmin: &onlyAdminFalse,
		},
	}, nil).Times(2)

	recommends := []bool{false, true, false}

	templateReleaseMgr.EXPECT().ListByTemplateName(gomock.Any(), "javaapp").
		Return([]*trmodels.TemplateRelease{
			{
				Model: global.Model{
					ID: 1,
				},
				Template:    1,
				Name:        "v1.0.0",
				Recommended: &recommends[0],
				OnlyAdmin:   &onlyAdminTrue,
			}, {
				Model: global.Model{
					ID: 1,
				},
				Template:    1,
				Name:        "v1.0.1",
				Recommended: &recommends[1],
				OnlyAdmin:   &onlyAdminFalse,
			}, {
				Model: global.Model{
					ID: 1,
				},
				Template:    1,
				Name:        "v1.0.2",
				Recommended: &recommends[2],
				OnlyAdmin:   &onlyAdminFalse,
			},
		}, nil).Times(2)

	ctl := &controller{
		templateMgr:        templateMgr,
		templateReleaseMgr: templateReleaseMgr,
	}

	templates, err := ctl.ListTemplate(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(templates))
	assert.Equal(t, "javaapp", templates[0].Name)
	assert.Equal(t, "tomcat", templates[1].Name)

	templateReleases, err := ctl.ListTemplateRelease(ctx, "javaapp")
	assert.Nil(t, err)
	assert.Equal(t, 3, len(templateReleases))
	assert.Equal(t, "v1.0.1", templateReleases[0].Name)
	assert.Equal(t, "v1.0.2", templateReleases[1].Name)
	assert.Equal(t, "v1.0.0", templateReleases[2].Name)

	ctx = common.WithContext(ctx, &userauth.DefaultInfo{
		Name:  "Jerry",
		ID:    1,
		Admin: false,
	})

	templates, err = ctl.ListTemplate(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(templates))
	assert.Equal(t, "tomcat", templates[0].Name)

	templateReleases, err = ctl.ListTemplateRelease(ctx, "javaapp")
	assert.Nil(t, err)
	assert.Equal(t, 2, len(templateReleases))
	assert.Equal(t, "v1.0.1", templateReleases[0].Name)
	assert.Equal(t, "v1.0.2", templateReleases[1].Name)
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
		Name:      releaseName,
		ChartName: charName,
	}
	templateReleaseMgr.EXPECT().GetByID(gomock.Any(), uint(1)).Return(release, nil)
	templateSchemaGetter.EXPECT().GetTemplateSchema(ctx, charName, releaseName, nil).Return(schemas, nil)

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
	assert.Equal(t, uint(0), template.InGroup)

	tpl, err := mgr.TemplateMgr.GetByID(ctx, 1)
	assert.Nil(t, err)
	assert.Equal(t, templateName, tpl.Name)
	assert.Equal(t, uint(1), tpl.ID)

	releases, err := ctl.ListTemplateReleaseByTemplateID(ctx, 1)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(releases))
	assert.Equal(t, releaseName, releases[0].Name)

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
	assert.Equal(t, uint(1), template.InGroup)

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

	onlyAdminTrue := true

	tplRequest := UpdateTemplateRequest{
		Description: "hello, world",
		Token:       "token",
		Repository:  "repo",
		OnlyAdmin:   &onlyAdminTrue,
	}
	err := ctl.UpdateTemplate(ctx, 1, tplRequest)
	assert.Nil(t, err)

	template, err := ctl.GetTemplate(ctx, 1)
	assert.Nil(t, err)
	assert.Equal(t, tplRequest.Description, template.Description)
	assert.Equal(t, tplRequest.Repository, template.Repository)
	assert.Equal(t, onlyAdminTrue, template.OnlyAdmin)

	oldDescription := tplRequest.Description
	tplRequest.Description = ""
	tplRequest.Token = "token2"

	err = ctl.UpdateTemplate(ctx, 1, tplRequest)
	assert.Nil(t, err)

	template, err = ctl.GetTemplate(ctx, 1)
	assert.Nil(t, err)
	assert.Equal(t, oldDescription, template.Description)
	assert.Equal(t, tplRequest.Repository, template.Repository)
	assert.Equal(t, onlyAdminTrue, template.OnlyAdmin)

	b := true
	trRequest := UpdateReleaseRequest{
		Description: "hello, world",
		Recommended: &b,
		OnlyAdmin:   &onlyAdminTrue,
	}
	err = ctl.UpdateRelease(ctx, 1, trRequest)
	assert.Nil(t, err)

	release, err := ctl.GetRelease(ctx, 1)
	assert.Nil(t, err)
	assert.Equal(t, trRequest.Description, release.Description)
	assert.Equal(t, *trRequest.Recommended, release.Recommended)
	assert.Equal(t, onlyAdminTrue, release.OnlyAdmin)

	oldDescription = trRequest.Description
	trRequest.Description = ""
	trRequest.Recommended = nil

	err = ctl.UpdateRelease(ctx, 1, trRequest)
	assert.Nil(t, err)

	release, err = ctl.GetRelease(ctx, 1)
	assert.Nil(t, err)
	assert.Equal(t, oldDescription, release.Description)
	assert.Equal(t, b, release.Recommended)
	assert.Equal(t, onlyAdminTrue, release.OnlyAdmin)

	ctx = common.WithContext(ctx, &userauth.DefaultInfo{
		Name:  "Jerry",
		ID:    1,
		Admin: false,
	})

	err = ctl.UpdateTemplate(ctx, 1, tplRequest)
	assert.NotNil(t, err)

	err = ctl.UpdateRelease(ctx, 1, trRequest)
	assert.NotNil(t, err)
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

	templates, err = ctl.ListTemplateByGroupID(ctx, 1)
	_, ok = perror.Cause(err).(*herrors.HorizonErrNotFound)
	assert.True(t, ok)
	assert.Equal(t, 0, len(templates))

	templates, err = ctl.ListTemplateByGroupID(ctx, 0)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(templates))

	releases, err = ctl.ListTemplateReleaseByTemplateID(ctx, 1)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(releases))
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
		&tmodels.Template{}, &groupmodels.Group{}); err != nil {
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
	repo, err := harbor.NewTemplateRepo(config.Repo{
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

	getter := reposchema.NewSchemaGetter(context.Background(), repo)

	ctl := &controller{
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
		_, err := mgr.GroupManager.Create(ctx, &groupmodels.Group{
			Name:      "test",
			Path:      "test",
			CreatedBy: 0,
			UpdatedBy: 0,
		})
		assert.Nil(t, err)
	}
	request := CreateTemplateRequest{
		CreateReleaseRequest: CreateReleaseRequest{
			RepoTag: releaseName,
		},
		Name:        templateName,
		Description: "",
		Repository:  templateRepo,
		Token:       templateRepoToken,
	}
	template, err := ctl.CreateTemplate(ctx, groupID, request)
	assert.Nil(t, err)
	_, err = ctl.CreateRelease(ctx, template.ID, request.CreateReleaseRequest)
	assert.Nil(t, err)
}
