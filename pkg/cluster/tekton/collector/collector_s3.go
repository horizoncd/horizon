package collector

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"path"
	"strconv"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws/awserr"
	awss3 "github.com/aws/aws-sdk-go/service/s3"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	"gopkg.in/natefinch/lumberjack.v2"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"

	"g.hz.netease.com/horizon/lib/s3"
	"g.hz.netease.com/horizon/pkg/cluster/tekton"
	"g.hz.netease.com/horizon/pkg/util/errors"
	logutil "g.hz.netease.com/horizon/pkg/util/log"
	"g.hz.netease.com/horizon/pkg/util/wlog"
)

const (
	_envKeyPipelineRunLogDIR  = "PIPELINE_RUN_LOG_DIR"
	_envKeyPipelineRunLogFile = "PIPELINE_RUN_LOG_FILE"

	_defaultPipelineRunLogDir  = "/var/log"
	_defaultPipelineRunLogFile = "build.log"

	// _expireTimeDuration one month
	_expireTimeDuration = time.Hour * 24 * 30
	// _mb 1Mb size
	_mb = 1024 * 1024
	// _limitSize limit size
	_limitSize = _mb * 2.5
)

type S3Collector struct {
	s3     s3.Interface
	tekton tekton.Interface
	logger *log.Logger
}

func NewS3Collector(s3 s3.Interface, tekton tekton.Interface) *S3Collector {
	dir := getEnvOrDefault(_envKeyPipelineRunLogDIR, _defaultPipelineRunLogDir)
	filename := getEnvOrDefault(_envKeyPipelineRunLogFile, _defaultPipelineRunLogFile)
	output := lumberjack.Logger{
		Filename:   path.Join(dir, filename),
		MaxSize:    256,  // 日志文件最大256MB
		MaxAge:     7,    // 最多保留7天的日志文件
		MaxBackups: 7,    // 最多保留7个日志文件
		LocalTime:  true, // 使用本地时间
	}
	logger := log.New(&output, "", log.LstdFlags)
	return &S3Collector{
		s3:     s3,
		tekton: tekton,
		logger: logger,
	}
}

func getEnvOrDefault(envKey string, defaultValue string) string {
	v := os.Getenv(envKey)
	if len(v) > 0 {
		return v
	}
	return defaultValue
}

type LogStruct struct {
	Object *MetadataAndURLStruct `json:"object"`
	Log    *LogAndURLStruct      `json:"log"`
}

type MetadataAndURLStruct struct {
	URL      string      `json:"url"`
	Metadata *ObjectMeta `json:"metadata"`
}

type LogAndURLStruct struct {
	URL     string `json:"url"`
	Content string `json:"content"`
}

func NewLogStruct(prURL string, metadata *ObjectMeta, logURL string, logContent string) *LogStruct {
	return &LogStruct{
		Object: &MetadataAndURLStruct{
			URL:      prURL,
			Metadata: metadata,
		},
		Log: &LogAndURLStruct{
			URL:     logURL,
			Content: logContent,
		},
	}
}

func (c *S3Collector) Collect(ctx context.Context, pr *v1beta1.PipelineRun) error {
	const op = "s3Collector: collect"
	metadata := resolveObjMetadata(pr)
	// 先收集日志，收集日志需要使用client-go访问k8s接口，
	// 如果pipelineRun不存在，直接忽略即可
	logURL, buildLog, err := c.collectLog(ctx, pr, metadata)
	if err != nil {
		return errors.E(op, err)
	}
	prURL, err := c.collectObject(ctx, metadata, pr)
	if err != nil {
		return errors.E(op, err)
	}

	logStruct := NewLogStruct(prURL, metadata, logURL, buildLog)
	b, err := json.Marshal(logStruct)
	if err != nil {
		logutil.Errorf(ctx, "failed to marshal log struct")
		return nil
	}
	//如果日志长度大于2.5mb，则只保留前1mb与后1mb日志数据。
	b = cutByteInMiddle(b, _limitSize, _mb, len(b)-_mb)
	c.logger.Println(string(b))
	return nil
}

func (c *S3Collector) Delete(ctx context.Context, application, cluster string) error {
	const op = "s3Collector: delete"
	var e1, e2 error

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		e1 = c.s3.DeleteObjects(ctx, c.getPrefixForPr(application, cluster))
	}()
	go func() {
		defer wg.Done()
		e2 = c.s3.DeleteObjects(ctx, c.getPrefixForPrLog(application, cluster))
	}()
	wg.Wait()

	if e1 != nil {
		return errors.E(op, e1)
	}
	if e2 != nil {
		return errors.E(op, e2)
	}
	return nil
}

func (c *S3Collector) GetLatestPipelineRunLog(ctx context.Context,
	application, cluster string) (_ []byte, err error) {
	const op = "s3Collector: getLatestPipelineRunLog"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	path := c.getPathForLatestPrLog(&ObjectMeta{
		Application: application,
		Cluster:     cluster,
	})
	b, err := c.s3.GetObject(ctx, path)
	if err != nil {
		if e, ok := err.(awserr.Error); ok {
			if e.Code() == awss3.ErrCodeNoSuchKey {
				return nil, errors.E(op, http.StatusNotFound, err)
			}
		}
		return nil, errors.E(op, err)
	}
	return b, nil
}

func (c *S3Collector) GetLatestPipelineRunObject(ctx context.Context,
	application, cluster string) (_ *Object, err error) {
	const op = "s3Collector: getLatestPipelineRunObject"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	path := c.getPathForLatestPr(&ObjectMeta{
		Application: application,
		Cluster:     cluster,
	})
	b, err := c.s3.GetObject(ctx, path)
	if err != nil {
		if e, ok := err.(awserr.Error); ok {
			if e.Code() == awss3.ErrCodeNoSuchKey {
				return nil, errors.E(op, http.StatusNotFound, err)
			}
		}
		return nil, errors.E(op, err)
	}
	var obj *Object
	if err := json.Unmarshal(b, &obj); err != nil {
		return nil, errors.E(op, err)
	}
	return obj, nil
}

func (c *S3Collector) collectObject(ctx context.Context, metadata *ObjectMeta,
	pr *v1beta1.PipelineRun) (_ string, err error) {
	const op = "s3Collector: collectObject"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })
	object := &Object{
		Metadata:    metadata,
		PipelineRun: pr,
	}
	b, err := json.Marshal(object)
	if err != nil {
		return "", errors.E(op, err)
	}
	prPath, err := c.getPathForPr(metadata)
	if err != nil {
		return "", errors.E(op, err)
	}
	prURL, err := c.s3.GetSignedObjectURL(prPath, _expireTimeDuration)
	if err != nil {
		return "", errors.E(op, err)
	}
	if err := c.s3.PutObject(ctx, prPath, bytes.NewReader(b), c.resolveMetadata(metadata)); err != nil {
		return "", errors.E(op, err)
	}
	if err := c.s3.CopyObject(ctx, prPath, c.getPathForLatestPr(metadata)); err != nil {
		return "", errors.E(op, err)
	}
	return prURL, nil
}

func (c *S3Collector) collectLog(ctx context.Context,
	pr *v1beta1.PipelineRun, metadata *ObjectMeta) (_ string, _ string, err error) {
	const op = "s3Collector: collectLog"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	logC, errC, err := c.tekton.GetPipelineRunLog(ctx, pr)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			// 如果pipelineRun没有找到，则error code返回http.StatusNotFound
			return "", "", errors.E(op, http.StatusNotFound, err)
		}
		return "", "", errors.E(op, err)
	}
	r, w := io.Pipe()
	go func() {
		defer func() { _ = w.Close() }()
		for logC != nil || errC != nil {
			select {
			case l, ok := <-logC:
				if !ok {
					logC = nil
					continue
				}
				if l.Log == "EOFLOG" {
					_, _ = w.Write([]byte("\n"))
					continue
				}
				_, _ = w.Write([]byte(fmt.Sprintf("[%s : %s] %s\n", l.Task, l.Step, l.Log)))
			case e, ok := <-errC:
				if !ok {
					errC = nil
					continue
				}
				_, _ = w.Write([]byte(fmt.Sprintf("%s\n", e)))
			}
		}
	}()

	prPath, err := c.getPathForPr(metadata)
	if err != nil {
		return "", "", errors.E(op, err)
	}
	logPath, err := c.getPathForPrLog(metadata)
	if err != nil {
		return "", "", errors.E(op, err)
	}

	logURL, err := c.s3.GetSignedObjectURL(logPath, _expireTimeDuration)
	if err != nil {
		return "", "", errors.E(op, err)
	}

	// TODO(demo) 日志先缓存到内存，再上传。如后续遇到内存占用很高的情况，可以考虑先存储到磁盘，再上传
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return "", "", errors.E(op, err)
	}
	if err := c.s3.PutObject(ctx, prPath, bytes.NewReader(b), nil); err != nil {
		return "", "", errors.E(op, err)
	}
	if err := c.s3.CopyObject(ctx, prPath, c.getPathForLatestPrLog(metadata)); err != nil {
		return "", "", errors.E(op, err)
	}
	return logURL, string(b), nil
}

// getPathForPr 计算pr的路径
// 由于s3的list接口返回的数据是按照字母序排序，故使用 maxInt64-{当前时间} 作为前缀，
// 这样能保证后续上传的对象在list接口中排在前面，实现按照creationTime倒序排序
func (c *S3Collector) getPathForPr(metadata *ObjectMeta) (string, error) {
	const op = "s3Collector: getPathForPr"
	creationTime, err := strconv.Atoi(metadata.CreationTimestamp)
	if err != nil {
		return "", errors.E(op, err)
	}
	prefix := math.MaxInt64 - creationTime
	return fmt.Sprintf("pr/%s/%s/%d-%s",
		metadata.Application, metadata.Cluster,
		prefix, metadata.PipelineRun.Name), nil
}

// getPathForLatestPr 计算pr的路径
func (c *S3Collector) getPathForLatestPr(metadata *ObjectMeta) string {
	return fmt.Sprintf("pr/%s/%s/latest",
		metadata.Application, metadata.Cluster)
}

// getPathForPr 计算pr的路径
// 由于s3的list接口返回的数据是按照字母序排序，故使用 maxInt64-{当前时间} 作为前缀，
// 这样能保证后续上传的对象在list接口中排在前面，实现按照creationTime倒序排序
func (c *S3Collector) getPathForPrLog(metadata *ObjectMeta) (string, error) {
	const op = "s3Collector: getPathForPrLog"
	creationTime, err := strconv.Atoi(metadata.CreationTimestamp)
	if err != nil {
		return "", errors.E(op, err)
	}
	prefix := math.MaxInt64 - creationTime
	return fmt.Sprintf("pr-log/%s/%s/%d-%s",
		metadata.Application, metadata.Cluster,
		prefix, metadata.PipelineRun.Name), nil
}

// getPathForLatestPr 计算pr的路径
func (c *S3Collector) getPathForLatestPrLog(metadata *ObjectMeta) string {
	return fmt.Sprintf("pr-log/%s/%s/latest",
		metadata.Application, metadata.Cluster)
}

func (c *S3Collector) getPrefixForPr(application, cluster string) string {
	return fmt.Sprintf("pr/%s/%s", application, cluster)
}

func (c *S3Collector) getPrefixForPrLog(application, cluster string) string {
	return fmt.Sprintf("pr-log/%s/%s", application, cluster)
}

/**
*  @brief 从开始至结束位置的Byte数组截断。
*
*  @param data 要处理的数组
*  @param limit 当大于传入的大小才会截断，否则不截断。传入-1则都会截断。以字节为单位，即1kb为1024,1Mb为1024*1024
*  @param begin 截断开始位置
*  @param end 截断结束位置

*
*  @return 处理过的数组
 */
func cutByteInMiddle(data []byte, limit int, begin int, end int) []byte {
	l := len(data)
	if limit > 0 && l < limit {
		return data
	}
	b1 := data[:begin]
	b2 := data[end:]

	b1 = append(b1, []byte("......")...)
	b1 = append(b1, b2...)
	return b1
}

const (
	application       = "application"
	cluster           = "cluster"
	environment       = "environment"
	operator          = "operator"
	pipelineRun       = "pipelineRun"
	pipeline          = "pipeline"
	result            = "result"
	duration          = "duration"
	creationTimestamp = "creationTimestamp"
)

func (c *S3Collector) resolveMetadata(metadata *ObjectMeta) map[string]string {
	metaMap := make(map[string]string)
	if metadata == nil {
		return metaMap
	}

	metaMap[application] = metadata.Application
	metaMap[cluster] = metadata.Cluster
	metaMap[environment] = metadata.Environment
	metaMap[operator] = metadata.Operator
	metaMap[pipelineRun] = metadata.PipelineRun.Name
	metaMap[pipeline] = metadata.PipelineRun.Pipeline
	metaMap[result] = metadata.PipelineRun.Result
	metaMap[duration] = strconv.Itoa(int(metadata.PipelineRun.DurationSeconds))
	metaMap[creationTimestamp] = metadata.CreationTimestamp

	return metaMap
}
