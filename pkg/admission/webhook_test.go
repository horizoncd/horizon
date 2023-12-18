package admission

import (
	"context"
	"testing"

	clusterctrl "github.com/horizoncd/horizon/core/controller/cluster"
	"github.com/horizoncd/horizon/pkg/admission/models"
	codemodels "github.com/horizoncd/horizon/pkg/cluster/code"
	admissionconfig "github.com/horizoncd/horizon/pkg/config/admission"
	tagmodels "github.com/horizoncd/horizon/pkg/tag/models"
	"github.com/stretchr/testify/assert"
)

func TestWebhook(t *testing.T) {
	ctx := context.Background()

	server := NewDummyWebhookServer()
	defer server.Stop()
	validatingURL := server.ValidatingURL()

	config := admissionconfig.Admission{
		Webhooks: []admissionconfig.Webhook{
			{
				Name:           "test1",
				Kind:           models.KindValidating,
				FailurePolicy:  admissionconfig.FailurePolicyFail,
				TimeoutSeconds: 5,
				Rules: []admissionconfig.Rule{
					{
						Resources: []string{
							"applications/clusters",
						},
						Operations: []models.Operation{
							models.OperationCreate,
						},
						Versions: []string{"v2"},
					},
				},
				ClientConfig: admissionconfig.ClientConfig{
					URL: validatingURL,
				},
			},
			{
				Name:           "test2",
				Kind:           models.KindValidating,
				FailurePolicy:  admissionconfig.FailurePolicyIgnore,
				TimeoutSeconds: 5,
				Rules: []admissionconfig.Rule{
					{
						Resources: []string{
							"clusters",
						},
						Operations: []models.Operation{
							models.OperationUpdate,
						},
						Versions: []string{"v2"},
					},
				},
				ClientConfig: admissionconfig.ClientConfig{
					URL: validatingURL,
				},
			},
		},
	}
	NewHTTPWebhooks(config)

	createBody := clusterctrl.CreateClusterRequestV2{
		Name:        "cluster-1",
		Description: "xxx",
		Priority:    "P0",
		Git: &codemodels.Git{
			URL:    "https://github.com/horizoncd/horizon.git",
			Branch: "main",
		},
		TemplateInfo: &codemodels.TemplateInfo{
			Name:    "javaapp",
			Release: "v1.0.0",
		},
	}

	createRequest := &Request{
		Operation:    models.OperationCreate,
		Resource:     "applications",
		ResourceName: "1",
		SubResource:  "clusters",
		Version:      "v2",
		Object:       createBody,
		OldObject:    nil,
		Options: map[string]interface{}{
			"scope": []string{"online/hz1"},
		},
	}
	err := Validating(ctx, createRequest)
	assert.NoError(t, err)

	createBody.Name = "cluster-invalid"
	createRequest.Object = createBody
	err = Validating(ctx, createRequest)
	assert.Error(t, err)
	t.Logf("error: %v", err)

	updateBody := clusterctrl.UpdateClusterRequestV2{
		Description: "yyy",
		Tags: tagmodels.TagsBasic{
			{
				Key:   "k1",
				Value: "v1",
			},
			{
				Key:   "scope",
				Value: "online/hz1",
			},
		},
	}
	updateRequest := &Request{
		Operation:    models.OperationUpdate,
		Resource:     "clusters",
		ResourceName: "1",
		SubResource:  "",
		Version:      "v2",
		Object:       updateBody,
		OldObject:    nil,
		Options:      nil,
	}
	err = Validating(ctx, updateRequest)
	assert.NoError(t, err)

	updateBody.Tags = tagmodels.TagsBasic{
		{
			Key:   "k1",
			Value: "v1",
		},
	}
	updateRequest.Object = updateBody
	err = Validating(ctx, updateRequest)
	assert.Error(t, err)
	t.Logf("error: %v", err)
}
