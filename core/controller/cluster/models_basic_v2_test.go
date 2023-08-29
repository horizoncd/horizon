package cluster

import (
	"testing"

	appmodels "github.com/horizoncd/horizon/pkg/application/models"
	codemodels "github.com/horizoncd/horizon/pkg/cluster/code"
	"github.com/horizoncd/horizon/pkg/cluster/models"
	envregionmodels "github.com/horizoncd/horizon/pkg/environmentregion/models"
	templatemodels "github.com/horizoncd/horizon/pkg/template/models"
	"github.com/horizoncd/horizon/pkg/util/common"
	"github.com/stretchr/testify/assert"
)

func TestModelsBasicV2(t *testing.T) {
	type args struct {
		params        *CreateClusterParamsV2
		application   *appmodels.Application
		er            *envregionmodels.EnvironmentRegion
		info          *BuildTemplateInfo
		template      *templatemodels.Template
		expireSeconds uint
	}
	gitInstanceParams := &CreateClusterParamsV2{
		CreateClusterRequestV2: &CreateClusterRequestV2{
			Name:        "git-instance",
			Description: "git instance",
			Git: &codemodels.Git{
				URL:       "ssh://git@localhost:7999/horizon/instance.git",
				Subfolder: "subfolder",
				Branch:    "main",
			},
		},
	}
	imageInstanceParams := &CreateClusterParamsV2{
		CreateClusterRequestV2: &CreateClusterRequestV2{
			Name:        "image-instance",
			Description: "image instance",
			Image:       common.StringPtr("horizon/instance:v1.0"),
		},
	}
	chartInstanceParams := &CreateClusterParamsV2{
		CreateClusterRequestV2: &CreateClusterRequestV2{
			Name:        "chart-instance",
			Description: "chart instance",
		},
	}
	gitApplication := &appmodels.Application{
		GitURL:       "ssh://git@localhost:7999/horizon/application.git",
		GitSubfolder: "",
		GitRef:       "master",
		GitRefType:   codemodels.GitRefTypeBranch,
	}
	imageApplication := &appmodels.Application{
		Image: "horizon/application:v1.0",
	}
	er := &envregionmodels.EnvironmentRegion{}
	rolloutBuildInfo := &BuildTemplateInfo{
		TemplateInfo: &codemodels.TemplateInfo{
			Name:    "rollout",
			Release: "v1.0",
		},
	}
	workloadTemplate := &templatemodels.Template{
		Type: templatemodels.TemplateTypeWorkload,
	}
	chartBuildInfo := &BuildTemplateInfo{
		TemplateInfo: &codemodels.TemplateInfo{
			Name:    "chart",
			Release: "v1.0",
		},
	}
	chartTemplate := &templatemodels.Template{
		Type: "chart",
	}
	tests := []struct {
		name    string
		args    args
		desired *models.Cluster
	}{
		{
			name: "create git instance in git application",
			args: args{
				params:        gitInstanceParams,
				application:   gitApplication,
				er:            er,
				info:          rolloutBuildInfo,
				template:      workloadTemplate,
				expireSeconds: 0,
			},
			desired: &models.Cluster{
				GitURL:       "ssh://git@localhost:7999/horizon/instance.git",
				GitSubfolder: "subfolder",
				GitRef:       "main",
				GitRefType:   codemodels.GitRefTypeBranch,
			},
		},
		{
			name: "create git instance in git application and use default info",
			args: args{
				params:        chartInstanceParams,
				application:   gitApplication,
				er:            er,
				info:          rolloutBuildInfo,
				template:      workloadTemplate,
				expireSeconds: 0,
			},
			desired: &models.Cluster{
				GitURL:       "ssh://git@localhost:7999/horizon/application.git",
				GitSubfolder: "",
				GitRef:       "master",
				GitRefType:   codemodels.GitRefTypeBranch,
			},
		},
		{
			name: "create git instance in git application and only specify git branch",
			args: args{
				params: &CreateClusterParamsV2{
					CreateClusterRequestV2: &CreateClusterRequestV2{
						Name:        "git-instance",
						Description: "git instance",
						Git: &codemodels.Git{
							Branch: "feature",
						},
					},
				},
				application:   gitApplication,
				er:            er,
				info:          rolloutBuildInfo,
				template:      workloadTemplate,
				expireSeconds: 0,
			},
			desired: &models.Cluster{
				GitURL:       "ssh://git@localhost:7999/horizon/application.git",
				GitSubfolder: "",
				GitRef:       "feature",
				GitRefType:   codemodels.GitRefTypeBranch,
			},
		},
		{
			name: "create image instance in git application",
			args: args{
				params:        imageInstanceParams,
				application:   gitApplication,
				er:            er,
				info:          rolloutBuildInfo,
				template:      workloadTemplate,
				expireSeconds: 0,
			},
			desired: &models.Cluster{
				Image: "horizon/instance:v1.0",
			},
		},
		{
			name: "create chart instance in git application",
			args: args{
				params:        chartInstanceParams,
				application:   gitApplication,
				er:            er,
				info:          chartBuildInfo,
				template:      chartTemplate,
				expireSeconds: 0,
			},
			desired: &models.Cluster{},
		},
		{
			name: "create git instance in image application",
			args: args{
				params:        gitInstanceParams,
				application:   imageApplication,
				er:            er,
				info:          rolloutBuildInfo,
				template:      workloadTemplate,
				expireSeconds: 0,
			},
			desired: &models.Cluster{
				GitURL:       "ssh://git@localhost:7999/horizon/instance.git",
				GitSubfolder: "subfolder",
				GitRef:       "main",
				GitRefType:   codemodels.GitRefTypeBranch,
			},
		},
		{
			name: "create image instance in image application",
			args: args{
				params:        imageInstanceParams,
				application:   imageApplication,
				er:            er,
				info:          rolloutBuildInfo,
				template:      workloadTemplate,
				expireSeconds: 0,
			},
			desired: &models.Cluster{
				Image: "horizon/instance:v1.0",
			},
		},
		{
			name: "create image instance in image application, use default info",
			args: args{
				params:        chartInstanceParams,
				application:   imageApplication,
				er:            er,
				info:          rolloutBuildInfo,
				template:      workloadTemplate,
				expireSeconds: 0,
			},
			desired: &models.Cluster{
				Image: "horizon/application:v1.0",
			},
		},
		{
			name: "create chart instance in image application",
			args: args{
				params:        chartInstanceParams,
				application:   imageApplication,
				er:            er,
				info:          chartBuildInfo,
				template:      chartTemplate,
				expireSeconds: 0,
			},
			desired: &models.Cluster{},
		},
	}
	for _, test := range tests {
		t.Logf("toClusterModel test: %s", test.name)
		cluster, _ := test.args.params.toClusterModel(test.args.application, test.args.er,
			test.args.info, test.args.template, test.args.expireSeconds)
		assert.Equal(t, test.desired.GitURL, cluster.GitURL)
		assert.Equal(t, test.desired.GitSubfolder, cluster.GitSubfolder)
		assert.Equal(t, test.desired.GitRef, cluster.GitRef)
		assert.Equal(t, test.desired.GitRefType, cluster.GitRefType)
		assert.Equal(t, test.desired.Image, cluster.Image)
	}
}
