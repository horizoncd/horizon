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
