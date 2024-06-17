// Copyright © 2023 Horizoncd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package s3

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	pathutil "path"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	awss3 "github.com/aws/aws-sdk-go/service/s3"

	"github.com/horizoncd/horizon/pkg/util/errors"
)

type Interface interface {
	PutObject(ctx context.Context, path string, content io.ReadSeeker, metadata map[string]string) error
	GetObject(ctx context.Context, path string) ([]byte, error)
	CopyObject(ctx context.Context, srcPath, destPath string) error
	// ListObjects NOTE: The returned results of the func are sorted alphabetically by key, not by upload time
	// Ref: https://docs.aws.amazon.com/AmazonS3/latest/API/API_ListObjects.html
	ListObjects(ctx context.Context, prefix string, maxKeys int64) ([]*awss3.Object, error)
	DeleteObjects(ctx context.Context, prefix string) error
	GetSignedObjectURL(path string, expire time.Duration) (string, error)
	GetBucket(ctx context.Context) string
}

type Params struct {
	AccessKey        string
	SecretKey        string
	Region           string
	Endpoint         string
	Bucket           string
	Prefix           string
	DisableSSL       bool
	SkipVerify       bool
	S3ForcePathStyle bool
	ContentType      string
	LogLevel         *aws.LogLevelType
}

func (params *Params) check() error {
	const op = "s3 params check"
	if len(params.AccessKey) == 0 {
		return errors.E(op, "AccessKey must be specified")
	}
	if len(params.SecretKey) == 0 {
		return errors.E(op, "SecretKey must be specified")
	}
	if len(params.Region) == 0 {
		return errors.E(op, "Region must be specified")
	}
	if len(params.Bucket) == 0 {
		return errors.E(op, "Bucket must be specified")
	}
	return nil
}

type Driver struct {
	Params
	S3 *awss3.S3
}

func NewDriver(params Params) (Interface, error) {
	const op = "new s3 driver"
	if err := params.check(); err != nil {
		return nil, err
	}
	d := &Driver{Params: params}
	d.Prefix = cleanPrefix(d.Prefix)

	awsConfig := aws.NewConfig()
	cred := credentials.NewStaticCredentials(params.AccessKey, params.SecretKey, "")
	awsConfig.WithCredentials(cred)
	awsConfig.WithRegion(params.Region)
	awsConfig.WithS3ForcePathStyle(params.S3ForcePathStyle)
	if len(params.Endpoint) > 0 {
		awsConfig.WithEndpoint(params.Endpoint)
	}
	awsConfig.WithDisableSSL(params.DisableSSL)
	if params.SkipVerify {
		awsConfig.WithHTTPClient(&http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		})
	}
	if params.LogLevel != nil {
		awsConfig.WithLogLevel(*params.LogLevel)
	}
	sess, err := session.NewSession(awsConfig)
	if err != nil {
		return nil, errors.E(op, err)
	}

	d.S3 = awss3.New(sess)

	return d, nil
}

func (d *Driver) PutObject(ctx context.Context, path string, content io.ReadSeeker, metadata map[string]string) error {
	_, err := d.S3.PutObjectWithContext(ctx, &awss3.PutObjectInput{
		Body:        content,
		Bucket:      aws.String(d.Bucket),
		ContentType: aws.String(d.ContentType),
		Key:         aws.String(pathutil.Join(d.Prefix, path)),
		Metadata: func() map[string]*string {
			if metadata == nil {
				return nil
			}
			ret := make(map[string]*string)
			for k, v := range metadata {
				ret[k] = aws.String(v)
			}
			return ret
		}(),
	})
	return err
}

func (d *Driver) GetObject(ctx context.Context, path string) ([]byte, error) {
	output, err := d.S3.GetObjectWithContext(ctx, &awss3.GetObjectInput{
		Bucket: aws.String(d.Bucket),
		Key:    aws.String(pathutil.Join(d.Prefix, path)),
	})
	if err != nil {
		return nil, err
	}
	defer func() { _ = output.Body.Close() }()

	content, err := ioutil.ReadAll(output.Body)
	if err != nil {
		return nil, err
	}
	return content, nil
}

func (d *Driver) GetSignedObjectURL(path string, expire time.Duration) (string, error) {
	req, _ := d.S3.GetObjectRequest(&awss3.GetObjectInput{
		Bucket: aws.String(d.Bucket),
		Key:    aws.String(pathutil.Join(d.Prefix, path)),
	})
	urlStr, err := req.Presign(expire)

	if err != nil {
		return "", err
	}
	return urlStr, nil
}

func (d *Driver) CopyObject(ctx context.Context, srcPath, destPath string) error {
	_, err := d.S3.CopyObjectWithContext(ctx, &awss3.CopyObjectInput{
		Bucket:     aws.String(d.Bucket),
		CopySource: aws.String(fmt.Sprintf("/%s/%s", d.Bucket, pathutil.Join(d.Prefix, srcPath))),
		Key:        aws.String(pathutil.Join(d.Prefix, destPath)),
	})
	return err
}

func (d *Driver) ListObjects(ctx context.Context, prefix string, maxKeys int64) ([]*awss3.Object, error) {
	var objects []*awss3.Object
	output, err := d.S3.ListObjectsWithContext(ctx, &awss3.ListObjectsInput{
		Bucket:  aws.String(d.Bucket),
		MaxKeys: aws.Int64(maxKeys),
		Prefix:  aws.String(pathutil.Join(d.Prefix, prefix)),
	})
	if err != nil {
		return nil, err
	}
	for _, obj := range output.Contents {
		path := removePrefixFromObjectPath(d.Prefix, *obj.Key)
		obj.Key = &path
		objects = append(objects, obj)
	}
	return objects, nil
}

func (d *Driver) DeleteObjects(ctx context.Context, prefix string) error {
	maxKeys := int64(1000)
	var objects []*awss3.Object
	var err error
	for {
		objects, err = d.ListObjects(ctx, prefix, maxKeys)
		if err != nil {
			return err
		}
		if objects == nil || len(objects) <= 0 {
			return nil
		}
		if _, err := d.S3.DeleteObjectsWithContext(ctx, &awss3.DeleteObjectsInput{
			Bucket: aws.String(d.Bucket),
			Delete: &awss3.Delete{
				Objects: func() []*awss3.ObjectIdentifier {
					identifiers := make([]*awss3.ObjectIdentifier, 0)
					for _, obj := range objects {
						objKey := pathutil.Join(d.Prefix, *obj.Key)
						identifiers = append(identifiers, &awss3.ObjectIdentifier{
							Key: &objKey,
						})
					}
					return identifiers
				}(),
			},
		}); err != nil {
			return err
		}
	}
}

func (d *Driver) GetBucket(ctx context.Context) string {
	return d.Bucket
}

func cleanPrefix(prefix string) string {
	return strings.Trim(prefix, "/")
}

func removePrefixFromObjectPath(prefix string, path string) string {
	if prefix == "" {
		return path
	}
	path = strings.TrimPrefix(path, fmt.Sprintf("%s/", prefix))
	return path
}
