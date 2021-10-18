package template

import (
	"context"
	"reflect"
	"testing"

	tmock "g.hz.netease.com/horizon/mock/pkg/template/manager"
	trmock "g.hz.netease.com/horizon/mock/pkg/templaterelease/manager"
	trschemamock "g.hz.netease.com/horizon/mock/pkg/templaterelease/schema"
	"g.hz.netease.com/horizon/pkg/template/models"
	trmodels "g.hz.netease.com/horizon/pkg/templaterelease/models"
	tsvc "g.hz.netease.com/horizon/pkg/templaterelease/schema"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

var (
	ctx          = context.Background()
	templateName = "javaapp"
	releaseName  = "v1.0.0"
)

func TestList(t *testing.T) {
	mockCtl := gomock.NewController(t)
	templateMgr := tmock.NewMockManager(mockCtl)
	templateReleaseMgr := trmock.NewMockManager(mockCtl)

	templateMgr.EXPECT().List(ctx).Return([]models.Template{
		{
			Model: gorm.Model{
				ID: 1,
			},
			Name: "javaapp",
		}, {
			Model: gorm.Model{
				ID: 2,
			},
			Name: "tomcat",
		},
	}, nil)

	templateReleaseMgr.EXPECT().ListByTemplateName(ctx, "javaapp").
		Return([]*trmodels.TemplateRelease{
			{
				Model: gorm.Model{
					ID: 1,
				},
				TemplateName: "javaapp",
				Name:         "v1.0.0",
				Recommended:  false,
			}, {
				Model: gorm.Model{
					ID: 1,
				},
				TemplateName: "javaapp",
				Name:         "v1.0.1",
				Recommended:  true,
			}, {
				Model: gorm.Model{
					ID: 1,
				},
				TemplateName: "javaapp",
				Name:         "v1.0.2",
				Recommended:  false,
			},
		}, nil)

	ctl := &controller{
		templateReleaseMgr: templateReleaseMgr,
		templateMgr:        templateMgr,
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
}

func TestGetSchema(t *testing.T) {
	mockCtl := gomock.NewController(t)
	templateSchemaGetter := trschemamock.NewMockSchemaGetter(mockCtl)
	schema := map[string]interface{}{
		"type": "object",
	}
	schemas := &tsvc.Schemas{
		Application: &tsvc.Schema{
			JSONSchema: schema,
			UISchema:   schema,
		},
		Pipeline: &tsvc.Schema{
			JSONSchema: schema,
			UISchema:   schema,
		},
	}
	templateSchemaGetter.EXPECT().GetTemplateSchema(ctx, templateName, releaseName).Return(schemas, nil)

	ctl := &controller{
		templateSchemaGetter: templateSchemaGetter,
	}

	ss, err := ctl.GetTemplateSchema(ctx, templateName, releaseName)
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
