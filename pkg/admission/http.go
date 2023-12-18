package admission

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	herrors "github.com/horizoncd/horizon/core/errors"
	"github.com/horizoncd/horizon/pkg/admission/models"
	config "github.com/horizoncd/horizon/pkg/config/admission"
	perror "github.com/horizoncd/horizon/pkg/errors"
	"github.com/horizoncd/horizon/pkg/util/common"
)

type HTTPAdmissionClient struct {
	config config.ClientConfig
	http.Client
}

// NewHTTPAdmissionClient creates a new HTTPAdmissionClient
func NewHTTPAdmissionClient(config config.ClientConfig, timeout time.Duration) *HTTPAdmissionClient {
	var transport = &http.Transport{}
	if config.CABundle != "" {
		ca := config.CABundle
		certPool := x509.NewCertPool()
		certPool.AppendCertsFromPEM([]byte(ca))
		transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: certPool,
			},
		}
	}
	if config.Insecure {
		transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
	}
	return &HTTPAdmissionClient{
		config: config,
		Client: http.Client{
			Timeout:   timeout,
			Transport: transport,
		},
	}
}

// Get sends the admission request to the webhook server and returns the response
func (c *HTTPAdmissionClient) Get(ctx context.Context, admitData *Request) (*Response, error) {
	body, err := json.Marshal(admitData)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.config.URL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, perror.Wrapf(herrors.ErrHTTPRespNotAsExpected, "status code: %d", resp.StatusCode)
	}
	defer resp.Body.Close()

	var response Response
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}
	return &response, nil
}

type ResourceMatcher struct {
	resources  map[string]struct{}
	operations map[models.Operation]struct{}
	versions   map[string]struct{}
}

// NewResourceMatcher creates a new ResourceMatcher
func NewResourceMatcher(rule config.Rule) *ResourceMatcher {
	matcher := &ResourceMatcher{
		resources:  make(map[string]struct{}),
		operations: make(map[models.Operation]struct{}),
		versions:   make(map[string]struct{}),
	}
	for _, resource := range rule.Resources {
		if resource == models.MatchAll {
			matcher.resources = nil
			break
		}
		matcher.resources[resource] = struct{}{}
	}
	for _, operation := range rule.Operations {
		if operation.Eq(models.Operation(models.MatchAll)) {
			matcher.operations = nil
			break
		}
		matcher.operations[operation] = struct{}{}
	}
	for _, version := range rule.Versions {
		if version == models.MatchAll {
			matcher.versions = nil
			break
		}
		matcher.versions[version] = struct{}{}
	}
	return matcher
}

// Match returns true if the request matches the matcher
func (m *ResourceMatcher) Match(req *Request) bool {
	if m.resources != nil {
		resource := req.Resource
		if req.SubResource != "" {
			resource = fmt.Sprintf("%s/%s", resource, req.SubResource)
		}
		if _, ok := m.resources[resource]; !ok {
			return false
		}
	}
	if m.operations != nil {
		if _, ok := m.operations[req.Operation]; !ok {
			return false
		}
	}
	if m.versions != nil {
		if _, ok := m.versions[req.Version]; !ok {
			return false
		}
	}
	return true
}

type ResourceMatchers []*ResourceMatcher

// NewResourceMatchers creates a new ResourceMatchers
func NewResourceMatchers(rules []config.Rule) ResourceMatchers {
	matchers := make(ResourceMatchers, len(rules))
	for i, rule := range rules {
		matchers[i] = NewResourceMatcher(rule)
	}
	return matchers
}

// Match returns true if any matcher matches the request
func (m ResourceMatchers) Match(req *Request) bool {
	for _, matcher := range m {
		if matcher.Match(req) {
			return true
		}
	}
	return false
}

type HTTPAdmissionWebhook struct {
	config     config.Webhook
	httpclient *HTTPAdmissionClient
	matchers   ResourceMatchers
}

// NewHTTPWebhooks registers the webhooks
func NewHTTPWebhooks(config config.Admission) {
	for _, webhook := range config.Webhooks {
		switch webhook.Kind {
		case models.KindValidating:
			Register(models.KindValidating, NewHTTPWebhook(webhook))
		}
	}
}

func NewHTTPWebhook(config config.Webhook) Webhook {
	client := NewHTTPAdmissionClient(config.ClientConfig, time.Duration(config.TimeoutSeconds)*time.Second)
	matchers := NewResourceMatchers(config.Rules)
	return &HTTPAdmissionWebhook{
		config:     config,
		httpclient: client,
		matchers:   matchers,
	}
}

// Handle handles the admission request and returns the response
func (m *HTTPAdmissionWebhook) Handle(ctx context.Context, req *Request) (*Response, error) {
	resp, err := m.httpclient.Get(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// IgnoreError returns true if the webhook is allowed to ignore the error
func (m *HTTPAdmissionWebhook) IgnoreError() bool {
	return m.config.FailurePolicy.Eq(config.FailurePolicyIgnore)
}

// Interest returns true if the request matches the webhook
func (m *HTTPAdmissionWebhook) Interest(req *Request) bool {
	return m.matchers.Match(req)
}

type DummyValidatingWebhookServer struct {
	server *httptest.Server
}

// NewDummyWebhookServer creates a dummy validating webhook server for testing
func NewDummyWebhookServer() *DummyValidatingWebhookServer {
	webhook := &DummyValidatingWebhookServer{}

	mux := http.NewServeMux()
	mux.HandleFunc("/validate", webhook.Validating)

	server := httptest.NewServer(mux)
	webhook.server = server
	return webhook
}

func (*DummyValidatingWebhookServer) ReadAndResponse(resp http.ResponseWriter,
	req *http.Request, fn func(Request, *Response)) {
	bodyBytes, _ := ioutil.ReadAll(req.Body)

	var admissionReq Request
	_ = json.Unmarshal(bodyBytes, &admissionReq)
	var admissionResp Response

	fn(admissionReq, &admissionResp)

	respBytes, _ := json.Marshal(admissionResp)
	resp.WriteHeader(http.StatusOK)
	_, _ = resp.Write(respBytes)
}

func (w *DummyValidatingWebhookServer) Validating(resp http.ResponseWriter, req *http.Request) {
	w.ReadAndResponse(resp, req, w.validating)
}

func (w *DummyValidatingWebhookServer) validating(req Request, resp *Response) {
	obj := req.Object.(map[string]interface{})

	if req.Operation.Eq(models.OperationCreate) {
		// check name
		name, ok := obj["name"].(string)
		if !ok {
			resp.Allowed = common.BoolPtr(false)
			resp.Result = "no name found"
			return
		}
		if strings.Contains(name, "invalid") {
			resp.Allowed = common.BoolPtr(false)
			resp.Result = fmt.Sprintf("name contains invalid: %s", name)
			return
		}
	}

	// check tags
	tagsMap, ok := obj["tags"].([]interface{})
	if !ok {
		// skip tag validation if no tags found
		resp.Allowed = common.BoolPtr(true)
		return
	}
	targetKey := "scope"
	exist := false
	for _, tag := range tagsMap {
		t, ok := tag.(map[string]interface{})
		if !ok {
			continue
		}
		if t["key"] == targetKey {
			exist = true
			break
		}
	}
	if !exist {
		resp.Allowed = common.BoolPtr(false)
		resp.Result = fmt.Sprintf("no tag with key: %s", targetKey)
		return
	}

	resp.Allowed = common.BoolPtr(true)
}

func (w *DummyValidatingWebhookServer) ValidatingURL() string {
	return w.server.URL + "/validate"
}

func (w *DummyValidatingWebhookServer) Stop() {
	w.server.Close()
}
