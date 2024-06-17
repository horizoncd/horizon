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

package tekton

type Mapper map[string]*Tekton

type Tekton struct {
	Server     string      `yaml:"server"`
	Namespace  string      `yaml:"namespace"`
	Kubeconfig string      `yaml:"kubeconfig"`
	LogStorage *LogStorage `yaml:"logStorage"`
}

type LogStorage struct {
	Type             string `yaml:"type"`
	AccessKey        string `yaml:"accessKey"`
	SecretKey        string `yaml:"secretKey"`
	Region           string `yaml:"region"`
	Endpoint         string `yaml:"endpoint"`
	Bucket           string `yaml:"bucket"`
	Prefix           string `yaml:"prefix"`
	DisableSSL       bool   `yaml:"disableSSL"`
	SkipVerify       bool   `yaml:"skipVerify"`
	S3ForcePathStyle bool   `yaml:"s3ForcePathStyle"`
}
