package registry

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"g.hz.netease.com/horizon/core/common"
	herrors "g.hz.netease.com/horizon/core/errors"
	perror "g.hz.netease.com/horizon/pkg/errors"
	"github.com/hashicorp/go-retryablehttp"

	"g.hz.netease.com/horizon/pkg/util/errors"
	"g.hz.netease.com/horizon/pkg/util/wlog"
)

const (
	CreateProject    = "createProject"
	AddMembers       = "addMembers"
	DeleteRepository = "deleteRepository"
	ListImage        = "listImage"
	PreHeatProject   = "preheatProject"

	Filters     = "[{\"type\":\"repository\",\"value\":\"**\"},{\"type\":\"tag\",\"value\":\"**\"}]"
	Trigger     = "{\"type\":\"event_based\",\"trigger_setting\":{\"cron\":\"\"}}"
	PreheatName = "kraken"
)

type HarborArtifact struct {
	Tags []HarborArtifactTag `json:"tags"`
}

type HarborArtifactTag struct {
	Name string `json:"name"`
}

type imageList []string

func (l imageList) Len() int {
	return len(l)
}

var timestampPattern = regexp.MustCompile(`\d{14}`)

// Less
// tag的格式为：{分支名}-{commitID}-{timestamp}
// 此处按照timestamp倒序排序，timestamp的格式比如:20210702134536
func (l imageList) Less(i, j int) bool {
	iTimeStr := timestampPattern.FindString(l[i])
	jTimeStr := timestampPattern.FindString(l[j])
	return jTimeStr < iTimeStr
}

func (l imageList) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

func (h *HarborRegistry) CreateProject(ctx context.Context, project string) (_ int, err error) {
	const op = "registry: create project"
	defer wlog.Start(ctx, op).StopPrint()

	url := fmt.Sprintf("%s/api/v2.0/projects", h.server)
	body := map[string]interface{}{
		"project_name": project,
		"metadata": map[string]string{
			"public": "false",
		},
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return -1, perror.Wrap(herrors.ErrParamInvalid, err.Error())
	}
	resp, err := h.sendHTTPRequest(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes), false, CreateProject)
	if err != nil {
		return -1, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusCreated {
		// 创建成功
		location := resp.Header.Get("Location")
		projectIDStr := location[strings.LastIndex(location, "/")+1:]
		projectID, err := strconv.Atoi(projectIDStr)
		if err != nil {
			return -1, perror.Wrap(herrors.ErrParamInvalid, err.Error())
		}
		if err := h.AddMembers(ctx, projectID); err != nil {
			return -1, errors.E(op, err)
		}
		if h.preheatPolicyID != 0 {
			_ = h.PreheatProject(ctx, project, projectID)
		}
		return projectID, nil
	} else if resp.StatusCode == http.StatusConflict {
		// 已经存在
		return -1, nil
	}

	message := common.Response(ctx, resp)
	return -1, errors.E(op, resp.StatusCode, message)
}

func (h *HarborRegistry) AddMembers(ctx context.Context, projectID int) (err error) {
	const op = "registry: add member for project"
	defer wlog.Start(ctx, op).StopPrint()

	url := fmt.Sprintf("%s/api/v2.0/projects/%d/members", h.server, projectID)
	addMember := func(m *HarborMember) error {
		body := map[string]interface{}{
			"role_id": m.Role,
			"member_user": map[string]string{
				"username": m.Username,
			},
		}
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return perror.Wrap(herrors.ErrParamInvalid, err.Error())
		}
		resp, err := h.sendHTTPRequest(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes), true, AddMembers)
		if err != nil {
			return err
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusConflict {
			return nil
		}
		return perror.Wrap(herrors.ErrHTTPRespNotAsExpected, common.Response(ctx, resp))
	}

	for _, member := range h.members {
		if err := addMember(member); err != nil {
			return err
		}
	}
	return nil
}

func (h *HarborRegistry) DeleteRepository(ctx context.Context, project string, repository string) (err error) {
	const op = "registry: delete repository"
	defer wlog.Start(ctx, op).StopPrint()

	url := fmt.Sprintf("%s/api/v2.0/projects/%s/repositories/%s", h.server, project, repository)
	resp, err := h.sendHTTPRequest(ctx, http.MethodDelete, url, nil, true, DeleteRepository)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNotFound {
		return nil
	}

	return perror.Wrap(herrors.ErrHTTPRespNotAsExpected, common.Response(ctx, resp))
}

func (h *HarborRegistry) ListImage(ctx context.Context,
	project string, repository string) (images []string, err error) {
	const op = "registry: list image tag"
	defer wlog.Start(ctx, op).StopPrint()

	const defaultImageCount = 10

	url := fmt.Sprintf("%s/api/v2.0/projects/%s/repositories/%s/artifacts?with_tag=true",
		h.server, project, repository) + "&q=tags%3d*"
	resp, err := h.sendHTTPRequest(ctx, http.MethodGet, url, nil, true, ListImage)

	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, perror.Wrap(herrors.ErrHTTPRespNotAsExpected, common.Response(ctx, resp))
	}

	var harborArtifacts []HarborArtifact
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, perror.Wrap(herrors.ErrReadFailed, err.Error())
	}
	if err = json.Unmarshal(body, &harborArtifacts); err != nil {
		return nil, perror.Wrap(herrors.ErrParamInvalid, err.Error())
	}

	for _, artifact := range harborArtifacts {
		if len(artifact.Tags) > 0 {
			images = append(images, fmt.Sprintf("%s/%s/%s:%s",
				strings.TrimPrefix(strings.TrimPrefix(h.server, "http://"), "https://"),
				project, repository, artifact.Tags[0].Name))
			if len(images) == defaultImageCount {
				break
			}
		}
	}

	sort.Sort(imageList(images))

	return images, nil
}

func (h *HarborRegistry) PreheatProject(ctx context.Context, project string,
	projectID int) (err error) {
	const op = "registry: preheat project"
	defer wlog.Start(ctx, op).StopPrint()

	preheatURL := fmt.Sprintf("%s/api/v2.0/projects/%s/preheat/policies", h.server, project)
	body := map[string]interface{}{
		"name":        PreheatName,
		"project_id":  projectID,
		"provider_id": h.preheatPolicyID,
		"filters":     Filters,
		"trigger":     Trigger,
		"enabled":     true,
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return perror.Wrap(herrors.ErrParamInvalid, err.Error())
	}
	resp, err := h.sendHTTPRequest(ctx, http.MethodPost, preheatURL, bytes.NewReader(bodyBytes), false, PreHeatProject)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusConflict {
		return nil
	}
	return perror.Wrap(herrors.ErrHTTPRespNotAsExpected, common.Response(ctx, resp))
}

func (h *HarborRegistry) GetServer(ctx context.Context) string {
	return h.server
}

func (h *HarborRegistry) sendHTTPRequest(ctx context.Context, method string,
	url string, body io.Reader, retry bool, operation string) (*http.Response, error) {
	begin := time.Now()
	var rsp *http.Response
	var err error
	defer func() {
		duration := time.Since(begin)
		server := h.server
		uri := strings.TrimPrefix(url, server)
		server = strings.TrimPrefix(strings.TrimPrefix(h.server, "http://"), "https://")
		// statuscode rsp可能为nil，默认为空字符串
		statuscode := ""
		if rsp != nil {
			statuscode = strconv.Itoa(rsp.StatusCode)
		}
		observe(server, method, uri, statuscode, operation, duration)
	}()
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, perror.Wrap(herrors.ErrHTTPRequestFailed, err.Error())
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", h.token))
	if !retry {
		rsp, err = h.client.Do(req)
		if err != nil {
			return nil, perror.Wrap(herrors.ErrHTTPRequestFailed, err.Error())
		}
		return rsp, nil
	}
	r, err := retryablehttp.FromRequest(req)
	if err != nil {
		return nil, perror.Wrap(herrors.ErrHTTPRequestFailed, err.Error())
	}
	rsp, err = h.retryableClient.Do(r)
	if err != nil {
		return nil, perror.Wrap(herrors.ErrHTTPRequestFailed, err.Error())
	}
	return rsp, nil
}
