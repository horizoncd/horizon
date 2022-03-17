package cmdb

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"g.hz.netease.com/horizon/core/common"
	"g.hz.netease.com/horizon/pkg/config/cmdb"
	"g.hz.netease.com/horizon/pkg/util/log"
	"g.hz.netease.com/horizon/pkg/util/wlog"
)

type Controller interface {
	CreateApplication(ctx context.Context, req CreateApplicationRequest) error
	DeleteApplication(ctx context.Context, appName string) error
	CreateCluster(ctx context.Context, req CreateClusterRequest) error
	DeleteCluster(ctx context.Context, clusterName string) error
}

func NewController(config cmdb.Config) Controller {
	return &controller{
		cli:    &http.Client{},
		config: config,
	}
}

type controller struct {
	cli    *http.Client
	config cmdb.Config
}

const (
	CreateApplicationFormat string = "http://%s/api/v2/createApplication?signature=%s&client=%s"
	DeleteApplicationFormat string = "http://%s/api/v2/deleteApplication?signature=%s&client=%s&applicationName=%s"
	CreateClusterFormat     string = "http://%s/api/v2/createCluster?signature=%s&client=%s"
	DeleteClusterFormat     string = "http://%s/api/v2/deleteCluster?signature=%s&client=%s&clusterName=%s" +
		"&forceDelete=true"
)

func (c *controller) getSignature() (string, error) {
	now := time.Now()
	year := now.Year()
	month := now.Month()
	day := now.Day()
	date := fmt.Sprintf("%d%02d%02d", year, month, day)
	content := fmt.Sprintf("%v%v%v", c.config.ClientID, date, c.config.SecretCode)

	hashFunc := md5.New()
	var b []byte
	_, err := io.WriteString(hashFunc, content)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hashFunc.Sum(b)), nil
}

func (c *controller) CreateApplication(ctx context.Context, req CreateApplicationRequest) (err error) {
	const op = "cmdb CreateApplication"
	defer wlog.Start(ctx, op).StopPrint()

	// 1. build request
	signature, err := c.getSignature()
	if err != nil {
		return err
	}

	URL := fmt.Sprintf(CreateApplicationFormat, c.config.URL, signature, c.config.ClientID)

	req.ParentID = c.config.ParentID
	content, err := json.Marshal(req)
	if err != nil {
		return err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, URL, bytes.NewReader(content))
	if err != nil {
		return err
	}
	httpReq.Header.Add("Content-Type", "application/json")

	// 2. do request
	resp, err := c.cli.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		message := common.Response(ctx, resp)
		return fmt.Errorf("%s, code  = %d, err = %s", op, resp.StatusCode, message)
	}

	var cResp CommonResp
	responseBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = json.NewDecoder(bytes.NewReader(responseBytes)).Decode(&cResp)
	if err != nil {
		return err
	}
	if cResp.Code != 200 && cResp.Code != 601 {
		return fmt.Errorf("%s, code = %+v, rt = %+v, body = %s", op, cResp.Code, cResp.RT, string(responseBytes))
	}
	log.WithFiled(ctx, "op", op).WithField("responseBody", string(responseBytes)).Info()
	return nil
}

func (c *controller) DeleteApplication(ctx context.Context, appName string) (err error) {
	const op = "cmdb DeleteApplication"
	defer wlog.Start(ctx, op).StopPrint()

	// 1. build request
	signature, err := c.getSignature()
	if err != nil {
		return err
	}
	URL := fmt.Sprintf(DeleteApplicationFormat, c.config.URL, signature, c.config.ClientID, appName)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, URL, nil)
	if err != nil {
		return err
	}

	// 2. do request
	resp, err := c.cli.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		message := common.Response(ctx, resp)
		return fmt.Errorf("code  = %d, err = %s", resp.StatusCode, message)
	}
	var cResp CommonResp
	responseBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	err = json.NewDecoder(bytes.NewReader(responseBytes)).Decode(&cResp)
	if err != nil {
		return err
	}
	if cResp.Code != 200 && cResp.Code != 1030 && cResp.Code != 601 {
		return fmt.Errorf("%s, code = %+v, rt = %+v, body = %s", op, cResp.Code, cResp.RT, string(responseBytes))
	}
	log.WithFiled(ctx, "op", op).WithField("responseBody", string(responseBytes)).Info()
	return nil
}

func (c *controller) CreateCluster(ctx context.Context, req CreateClusterRequest) (err error) {
	const op = "cmdb CreateCluster"
	defer wlog.Start(ctx, op).StopPrint()

	// 1. build request
	signature, err := c.getSignature()
	if err != nil {
		return err
	}
	URL := fmt.Sprintf(CreateClusterFormat, c.config.URL, signature, c.config.ClientID)

	req.ClusterStyle = Docker
	content, err := json.Marshal(req)
	if err != nil {
		return err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, URL, bytes.NewReader(content))
	httpReq.Header.Add("Content-Type", "application/json")
	if err != nil {
		return err
	}

	// 2. do request
	resp, err := c.cli.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		message := common.Response(ctx, resp)
		return fmt.Errorf("code  = %d, err = %s", resp.StatusCode, message)
	}
	var cResp CommonResp
	responseBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	err = json.NewDecoder(bytes.NewReader(responseBytes)).Decode(&cResp)
	if err != nil {
		return err
	}
	if cResp.Code != 200 && cResp.Code != 601 {
		return fmt.Errorf("%s, code = %+v, rt = %+v, body = %s", op, cResp.Code, cResp.RT, string(responseBytes))
	}
	log.WithFiled(ctx, "op", op).WithField("responseBody", string(responseBytes)).Info()
	return nil
}
func (c *controller) DeleteCluster(ctx context.Context, clusterName string) (err error) {
	const op = "cmdb DeleteCluster"
	defer wlog.Start(ctx, op).StopPrint()

	// 1. build request
	signature, err := c.getSignature()
	if err != nil {
		return err
	}
	URL := fmt.Sprintf(DeleteClusterFormat, c.config.URL, signature, c.config.ClientID, clusterName)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, URL, nil)
	if err != nil {
		return err
	}

	// 2. do request
	resp, err := c.cli.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		message := common.Response(ctx, resp)
		return fmt.Errorf("code  = %d, err = %s", resp.StatusCode, message)
	}
	var cResp CommonResp
	responseBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	err = json.NewDecoder(bytes.NewReader(responseBytes)).Decode(&cResp)
	if err != nil {
		return err
	}
	if cResp.Code != 200 && cResp.Code != 1030 && cResp.Code != 601 {
		return fmt.Errorf("%s, code = %+v, rt = %+v, body = %s", op, cResp.Code, cResp.RT, string(responseBytes))
	}
	log.WithFiled(ctx, "op", op).WithField("responseBody", string(responseBytes)).Info()
	return nil
}
