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

package schema

import (
	"bytes"
	"context"
	"encoding/json"
	"sync"
	"text/template"

	"github.com/Masterminds/sprig"
	herrors "github.com/horizoncd/horizon/core/errors"
	perror "github.com/horizoncd/horizon/pkg/errors"
)

type Schemas struct {
	Application *Schema
	Pipeline    *Schema
}

type Schema struct {
	JSONSchema map[string]interface{}
	UISchema   map[string]interface{}
}

// params
const (
	ClusterIDKey    string = "clusterID"
	ResourceTypeKey string = "resourceType"
)

// Getter provides some functions for template schema
// nolint
//
//go:generate mockgen -source=$GOFILE -destination=../../../mock/pkg/templaterelease/schema/mock_schema.go -package=mock_schema
type Getter interface {
	// GetTemplateSchema get schema for specified template release. todo(gjq) add cache
	GetTemplateSchema(ctx context.Context, templateName, releaseName string, params map[string]string) (*Schemas, error)
}

func RenderFiles(params map[string]string, files ...[]byte) (retFiles [][]byte, _ error) {
	for _, file := range files {
		if file != nil {
			var b bytes.Buffer
			doTemplate := template.Must(template.New("").Funcs(sprig.TxtFuncMap()).Parse(string(file)))
			err := doTemplate.ExecuteTemplate(&b, "", params)
			if err != nil {
				return nil, perror.Wrap(herrors.ErrParamInvalid, err.Error())
			}
			retFiles = append(retFiles, b.Bytes())
		} else {
			retFiles = append(retFiles, nil)
		}
	}
	return retFiles, nil
}

func ParseFiles(params map[string]string,
	pipelineSchemaBytes, applicationSchemaBytes,
	pipelineUISchemaBytes, applicationUISchemaBytes []byte,
) (*Schemas, error) {
	readerSchemas, err := RenderFiles(params, pipelineSchemaBytes, applicationSchemaBytes)
	if err != nil {
		return nil, err
	}
	pipelineSchemaBytes = readerSchemas[0]
	applicationSchemaBytes = readerSchemas[1]

	var err1, err2, err3, err4 error
	// 2. unmarshal concurrently
	var pipelineSchema, applicationSchema, pipelineUISchema, applicationUISchema map[string]interface{}
	var wgUnmarshal sync.WaitGroup
	wgUnmarshal.Add(4)
	unmarshal := func(b []byte, m *map[string]interface{}, err *error) {
		defer wgUnmarshal.Done()
		if b != nil {
			if e := json.Unmarshal(b, &m); e != nil {
				*err = perror.Wrap(herrors.ErrParamInvalid, e.Error())
			}
		}
	}
	go unmarshal(pipelineSchemaBytes, &pipelineSchema, &err1)
	go unmarshal(applicationSchemaBytes, &applicationSchema, &err2)
	go unmarshal(pipelineUISchemaBytes, &pipelineUISchema, &err3)
	go unmarshal(applicationUISchemaBytes, &applicationUISchema, &err4)
	wgUnmarshal.Wait()

	for _, err := range []error{err1, err2, err3, err4} {
		if err != nil {
			return nil, err
		}
	}

	return &Schemas{
		Application: &Schema{
			JSONSchema: applicationSchema,
			UISchema:   applicationUISchema,
		},
		Pipeline: &Schema{
			JSONSchema: pipelineSchema,
			UISchema:   pipelineUISchema,
		},
	}, nil
}
