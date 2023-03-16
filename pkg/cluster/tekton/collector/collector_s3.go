package collector

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strconv"
	"time"

	herrors "github.com/horizoncd/horizon/pkg/core/errors"
	perror "github.com/horizoncd/horizon/pkg/errors"
	"github.com/horizoncd/horizon/pkg/server/global"

	"github.com/aws/aws-sdk-go/aws/awserr"
	awss3 "github.com/aws/aws-sdk-go/service/s3"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	"gopkg.in/natefinch/lumberjack.v2"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/horizoncd/horizon/lib/s3"
	"github.com/horizoncd/horizon/pkg/cluster/tekton"
	logutil "github.com/horizoncd/horizon/pkg/util/log"
	"github.com/horizoncd/horizon/pkg/util/wlog"
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
		MaxSize:    256,
		MaxAge:     7,
		MaxBackups: 7,
		LocalTime:  true,
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

type CollectResult struct {
	Bucket         string
	LogObject      string
	PrObject       string
	Result         string
	StartTime      *metav1.Time
	CompletionTime *metav1.Time
}

func (c *S3Collector) Collect(ctx context.Context, pr *v1beta1.PipelineRun, horizonMetaData *global.HorizonMetaData) (
	*CollectResult, error) {
	const op = "s3Collector: collect"
	defer wlog.Start(ctx, op).StopPrint()
	metadata := resolveObjMetadata(pr, horizonMetaData)
	collectLogResult, err := c.collectLog(ctx, pr, metadata)
	if err != nil {
		return nil, err
	}
	collectObjectResult, err := c.collectObject(ctx, metadata, pr)
	if err != nil {
		return nil, err
	}

	logStruct := NewLogStruct(collectObjectResult.PrURL,
		metadata, collectLogResult.LogURL, collectLogResult.LogContent)
	b, err := json.Marshal(logStruct)
	if err != nil {
		logutil.Errorf(ctx, "failed to marshal log struct")
		return nil, perror.Wrap(herrors.ErrParamInvalid, err.Error())
	}
	b = cutByteInMiddle(b, _limitSize, _mb, len(b)-_mb)
	c.logger.Println(string(b))
	return &CollectResult{
		Bucket:         c.s3.GetBucket(ctx),
		LogObject:      collectLogResult.LogObject,
		PrObject:       collectObjectResult.PrObject,
		Result:         metadata.PipelineRun.Result,
		StartTime:      metadata.PipelineRun.StartTime,
		CompletionTime: metadata.PipelineRun.CompletionTime,
	}, nil
}

func (c *S3Collector) GetPipelineRunLog(ctx context.Context, logObject string) (_ []byte, err error) {
	const op = "s3Collector: getPipelineRunLog"
	defer wlog.Start(ctx, op).StopPrint()

	b, err := c.s3.GetObject(ctx, logObject)
	if err != nil {
		if e, ok := err.(awserr.Error); ok {
			if e.Code() == awss3.ErrCodeNoSuchKey {
				return nil, herrors.NewErrNotFound(herrors.PipelinerunLog, err.Error())
			}
		}
		return nil, perror.Wrap(herrors.ErrS3GetObjFailed, err.Error())
	}
	return b, nil
}

func (c *S3Collector) GetPipelineRunObject(ctx context.Context, object string) (_ *Object, err error) {
	const op = "s3Collector: getPipelineRunObject"
	defer wlog.Start(ctx, op).StopPrint()

	b, err := c.s3.GetObject(ctx, object)
	if err != nil {
		if e, ok := err.(awserr.Error); ok {
			if e.Code() == awss3.ErrCodeNoSuchKey {
				return nil, herrors.NewErrNotFound(herrors.PipelinerunObj, err.Error())
			}
		}
		return nil, perror.Wrap(herrors.ErrS3GetObjFailed, err.Error())
	}
	var obj *Object
	if err := json.Unmarshal(b, &obj); err != nil {
		return nil, perror.Wrap(herrors.ErrParamInvalid, err.Error())
	}
	return obj, nil
}

type CollectObjectResult struct {
	PrObject string
	PrURL    string
}

func (c *S3Collector) collectObject(ctx context.Context, metadata *ObjectMeta,
	pr *v1beta1.PipelineRun) (_ *CollectObjectResult, err error) {
	const op = "s3Collector: collectObject"
	defer wlog.Start(ctx, op).StopPrint()
	object := &Object{
		Metadata:    metadata,
		PipelineRun: pr,
	}
	b, err := json.Marshal(object)
	if err != nil {
		return nil, perror.Wrap(herrors.ErrParamInvalid, err.Error())
	}
	prPath := c.getPathForPr(metadata)

	prURL, err := c.s3.GetSignedObjectURL(prPath, _expireTimeDuration)
	if err != nil {
		return nil, perror.Wrap(herrors.ErrS3SignFailed, err.Error())
	}
	if err := c.s3.PutObject(ctx, prPath, bytes.NewReader(b), c.resolveMetadata(metadata)); err != nil {
		return nil, perror.Wrap(herrors.ErrS3PutObjFailed, err.Error())
	}
	return &CollectObjectResult{
		PrObject: prPath,
		PrURL:    prURL,
	}, nil
}

type CollectLogResult struct {
	LogObject  string
	LogURL     string
	LogContent string
}

func (c *S3Collector) collectLog(ctx context.Context,
	pr *v1beta1.PipelineRun, metadata *ObjectMeta) (_ *CollectLogResult, err error) {
	const op = "s3Collector: collectLog"
	defer wlog.Start(ctx, op).StopPrint()

	logC, errC, err := c.tekton.GetPipelineRunLog(ctx, pr)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil, herrors.NewErrNotFound(herrors.Pipelinerun, "")
		}
		return nil, herrors.NewErrGetFailed(herrors.Pipelinerun, "")
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

	logPath := c.getPathForPrLog(metadata)

	logURL, err := c.s3.GetSignedObjectURL(logPath, _expireTimeDuration)
	if err != nil {
		return nil, perror.Wrap(herrors.ErrS3SignFailed, err.Error())
	}

	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, perror.Wrap(herrors.ErrReadFailed, err.Error())
	}
	if err := c.s3.PutObject(ctx, logPath, bytes.NewReader(b), nil); err != nil {
		return nil, perror.Wrap(herrors.ErrS3PutObjFailed, err.Error())
	}
	return &CollectLogResult{
		LogObject:  logPath,
		LogURL:     logURL,
		LogContent: string(b),
	}, nil
}

func (c *S3Collector) getPathForPr(metadata *ObjectMeta) string {
	timeFormat := "200601"
	timeStr := time.Now().Format(timeFormat)
	return fmt.Sprintf("%s/pr/%s-%s/%s-%s/%s", timeStr,
		metadata.Application, metadata.ApplicationID, metadata.Cluster, metadata.ClusterID,
		metadata.PipelineRun.Name)
}

func (c *S3Collector) getPathForPrLog(metadata *ObjectMeta) string {
	timeFormat := "200601"
	timeStr := time.Now().Format(timeFormat)
	return fmt.Sprintf("%s/pr-log/%s-%s/%s-%s/%s", timeStr,
		metadata.Application, metadata.ApplicationID, metadata.Cluster, metadata.ClusterID,
		metadata.PipelineRun.Name)
}

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
