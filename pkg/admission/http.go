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

	"github.com/mattbaird/jsonpatch"

	herrors "github.com/horizoncd/horizon/core/errors"
	"github.com/horizoncd/horizon/pkg/admission/models"
	config "github.com/horizoncd/horizon/pkg/config/admission"
	perror "github.com/horizoncd/horizon/pkg/errors"
)

type HTTPAdmissionClient struct {
	config config.ClientConfig
	http.Client
}

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

func NewResourceMatcher(rule config.Rule) *ResourceMatcher {
	matcher := &ResourceMatcher{
		resources:  make(map[string]struct{}),
		operations: make(map[models.Operation]struct{}),
		versions:   make(map[string]struct{}),
	}
	for _, resource := range rule.Resources {
		if resource == "*" {
			matcher.resources = nil
			break
		}
		matcher.resources[resource] = struct{}{}
	}
	for _, operation := range rule.Operations {
		if operation.Eq(models.OperationAll) {
			matcher.operations = nil
			break
		}
		matcher.operations[operation] = struct{}{}
	}
	for _, version := range rule.Versions {
		if version == "*" {
			matcher.versions = nil
			break
		}
		matcher.versions[version] = struct{}{}
	}
	return matcher
}

func (m *ResourceMatcher) Match(req *Request) bool {
	if m.resources != nil {
		if _, ok := m.resources[req.Resource]; !ok {
			return false
		}
	}
	if m.operations != nil {
		if _, ok := m.operations[models.Operation(req.Operation)]; !ok {
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

func NewResourceMatchers(rules []config.Rule) ResourceMatchers {
	matchers := make(ResourceMatchers, len(rules))
	for i, rule := range rules {
		matchers[i] = NewResourceMatcher(rule)
	}
	return matchers
}

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

func NewHTTPWebhooks(config config.Admission) {
	for _, webhook := range config.Webhooks {
		switch webhook.Kind {
		case models.KindMutating:
			Register(models.KindMutating, NewHTTPWebhook(webhook))
		case models.KindValidating:
			Register(models.KindValidating, NewHTTPWebhook(webhook))
		}
	}
}

func NewHTTPWebhook(config config.Webhook) *HTTPAdmissionWebhook {
	client := NewHTTPAdmissionClient(config.ClientConfig, time.Duration(config.TimeoutSeconds)*time.Second)
	matchers := NewResourceMatchers(config.Rules)
	return &HTTPAdmissionWebhook{
		config:     config,
		httpclient: client,
		matchers:   matchers,
	}
}

func (m *HTTPAdmissionWebhook) Handle(ctx context.Context, req *Request) (*Response, error) {
	resp, err := m.httpclient.Get(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *HTTPAdmissionWebhook) IgnoreError() bool {
	return m.config.FailurePolicy.Eq(config.FailurePolicyIgnore)
}

func (m *HTTPAdmissionWebhook) Interest(req *Request) bool {
	return m.matchers.Match(req)
}

type DummyMutatingWebhookServer struct {
	server *httptest.Server
}

func NewDummyWebhookServer() *DummyMutatingWebhookServer {
	webhook := &DummyMutatingWebhookServer{}

	mux := http.NewServeMux()
	mux.HandleFunc("/mutate", webhook.Mutating)
	mux.HandleFunc("/validate", webhook.Validating)

	server := httptest.NewServer(mux)
	webhook.server = server
	return webhook
}

func (*DummyMutatingWebhookServer) ReadAndResponse(resp http.ResponseWriter,
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

func (w *DummyMutatingWebhookServer) Mutating(resp http.ResponseWriter, req *http.Request) {
	w.ReadAndResponse(resp, req, w.mutating)
}

type Tag struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func (*DummyMutatingWebhookServer) mutating(req Request, resp *Response) {
	obj := req.Object.(map[string]interface{})

	jsonObj, _ := json.Marshal(obj)

	var newObj map[string]interface{}
	_ = json.Unmarshal(jsonObj, &newObj)
	if obj["tags"] != nil {
		tags := obj["tags"].([]interface{})
		tags = append(tags, map[string]interface{}{"key": "insertByWebhook", "value": "insertByWebhook"})
		newObj["tags"] = tags
	}

	newObj["name"] = fmt.Sprintf("%v-%s", obj["name"], "mutated")

	jsonNewObj, _ := json.Marshal(newObj)

	patch, _ := jsonpatch.CreatePatch(jsonObj, jsonNewObj)

	patchJSON, _ := json.Marshal(patch)

	resp.Patch = patchJSON
	resp.PatchType = models.PatchTypeJSONPatch
}

func (w *DummyMutatingWebhookServer) Validating(resp http.ResponseWriter, req *http.Request) {
	w.ReadAndResponse(resp, req, w.validating)
}

func (w *DummyMutatingWebhookServer) validating(req Request, resp *Response) {
	obj := req.Object.(map[string]interface{})

	allow := true

	name, ok := obj["name"].(string)
	if !ok {
		allow = false
		resp.Result = "no name found"
	}

	if strings.Contains(name, "invalid") {
		allow = false
		resp.Result = fmt.Sprintf("name contains invalid: %s", name)
	}

	if obj["tags"] != nil {
		tags := obj["tags"].([]interface{})
		for _, tag := range tags {
			tag := tag.(map[string]interface{})
			tagKey := tag["key"].(string)
			if strings.Contains(tagKey, "invalid") {
				allow = false
				resp.Result = fmt.Sprintf("tag key contains invalid: %s", tagKey)
				break
			}
		}
	}
	if !allow {
		allow = false
		resp.Allowed = &allow
		return
	}
	resp.Allowed = &allow
}

func (w *DummyMutatingWebhookServer) MutatingURL() string {
	return w.server.URL + "/mutate"
}

func (w *DummyMutatingWebhookServer) ValidatingURL() string {
	return w.server.URL + "/validate"
}

func (w *DummyMutatingWebhookServer) Stop() {
	w.server.Close()
}
