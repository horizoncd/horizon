package common

import "testing"

func TestInternalSSHToHTTPURL(t *testing.T) {
	type args struct {
		sshURL string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "normal",
			args: args{
				sshURL: "ssh://git@g.hz.netease.com:22222/cloudmusic/music-arc/knative-sink-test.git",
			},
			want: "https://g.hz.netease.com/cloudmusic/music-arc/knative-sink-test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := InternalSSHToHTTPURL(tt.args.sshURL); got != tt.want {
				t.Errorf("InternalSSHToHTTPURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
