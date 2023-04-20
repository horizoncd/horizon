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

package angular

import (
	"testing"
)

type Body struct {
	Image   *string `json:"image"`
	Replica *int    `json:"replica"`
}

func TestCommitMessage(t *testing.T) {
	type args struct {
		scope   string
		subject Subject
		body    Body
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "create application",
			args: args{
				scope: "a.md",
				subject: Subject{
					Operator:    "alice",
					Action:      "create application",
					Application: StringPtr("application-test-1"),
				},
				body: Body{
					Image:   func() *string { a := "image"; return &a }(),
					Replica: func() *int { i := 1; return &i }(),
				},
			},
			want: "change(a.md): alice " + "create application" + " application-test-1" + "\n\n" +
				"{\"header\":{\"kind\":\"change\",\"scope\":\"a.md\",\"subject\":{\"operator\":\"alice\"," +
				"\"action\":\"create application\",\"application\":\"application-test-1\"}}," +
				"\"body\":{\"image\":\"image\",\"replica\":1}}",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CommitMessage(tt.args.scope, tt.args.subject, &tt.args.body); got != tt.want {
				t.Errorf("CommitMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}
