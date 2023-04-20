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

package build

import (
	"golang.org/x/net/context"
)

// Controller get build schema
type Controller interface {
	GetSchema(ctx context.Context) (*Schema, error)
}

type controller struct {
	schema *Schema
}

func NewController(schema *Schema) Controller {
	return &controller{schema: schema}
}

func (c controller) GetSchema(ctx context.Context) (*Schema, error) {
	return c.schema, nil
}

var _ Controller = (*controller)(nil)
