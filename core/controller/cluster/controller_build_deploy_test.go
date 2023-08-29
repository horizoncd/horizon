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

package cluster

import (
	"strings"
	"testing"

	regionmodels "github.com/horizoncd/horizon/pkg/region/models"
	registrymodels "github.com/horizoncd/horizon/pkg/registry/models"

	"github.com/mozillazg/go-pinyin"
)

func testImageURL(t *testing.T) {
	type args struct {
		regionEntity *regionmodels.RegionEntity
		application  string
		cluster      string
		branch       string
		commit       string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "normal1",
			args: args{
				regionEntity: &regionmodels.RegionEntity{
					Registry: &registrymodels.Registry{
						Path:   "path",
						Server: "https://harbor.com",
					},
				},
				application: "app",
				cluster:     "cluster",
				branch:      "master",
				commit:      "117651f0c06486ba50a01eb0ed82be46ef3b528e",
			},
			want: "harbor.com/path/app/cluster:master-117651f0",
		},
		{
			name: "normal2",
			args: args{
				regionEntity: &regionmodels.RegionEntity{
					Registry: &registrymodels.Registry{
						Path:   "path",
						Server: "https://harbor.com",
					},
				},
				application: "app",
				cluster:     "cluster",
				branch:      "测试",
				commit:      "117651f0c06486ba50a01eb0ed82be46ef3b528e",
			},
			want: "harbor.com/path/app/cluster:ceshi-117651f0",
		},
		{
			name: "normal3",
			args: args{
				regionEntity: &regionmodels.RegionEntity{
					Registry: &registrymodels.Registry{
						Path:   "path",
						Server: "https://harbor.com",
					},
				},
				application: "app",
				cluster:     "cluster",
				branch:      "测试$90",
				commit:      "117651f0c06486ba50a01eb0ed82be46ef3b528e",
			},
			want: "harbor.com/path/app/cluster:ceshi_90-117651f0",
		},
		{
			name: "normal4",
			args: args{
				regionEntity: &regionmodels.RegionEntity{
					Registry: &registrymodels.Registry{
						Path:   "path",
						Server: "https://harbor.com",
					},
				},
				application: "app",
				cluster:     "cluster",
				branch:      "测试@中国hello*",
				commit:      "117651f0c06486ba50a01eb0ed82be46ef3b528e",
			},
			want: "harbor.com/path/app/cluster:ceshi_zhongguohello_-117651f0",
		},
		{
			name: "normal5",
			args: args{
				regionEntity: &regionmodels.RegionEntity{
					Registry: &registrymodels.Registry{
						Path:   "path",
						Server: "https://harbor.com",
					},
				},
				application: "app",
				cluster:     "cluster",
				branch:      "fix/bug",
				// commit will be branch name if the repo could not be fetched
				commit: "fix/bug",
			},
			want: "harbor.com/path/app/cluster:fix_bug-fix_bug",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := assembleImageURL(tt.args.regionEntity, tt.args.application,
				tt.args.cluster, tt.args.branch, tt.args.commit); !strings.HasPrefix(got, tt.want) {
				t.Errorf("imageURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func testPinyin(t *testing.T) {
	args := pinyin.Args{
		Fallback: func(r rune, a pinyin.Args) []string {
			return []string{string(r)}
		},
	}
	t.Logf("%v", pinyin.LazyPinyin("中国", args))
	t.Logf("%v", pinyin.LazyPinyin("中国ren", args))
	t.Logf("%v", pinyin.LazyPinyin("123中国ren", args))
	t.Logf("%v", pinyin.LazyPinyin("hello", args))
	t.Logf("%v", pinyin.LazyPinyin("hello-|.>()", args))
}
