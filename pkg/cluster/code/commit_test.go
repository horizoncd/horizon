package code

import "testing"

func Test_extractProjectPathFromSSHURL(t *testing.T) {
	type args struct {
		gitURL string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "normally",
			args: args{
				gitURL: "ssh://git@g.hz.netease.com:22222/music-cloud-native/horizon/horizon.git",
			},
			want:    "music-cloud-native/horizon/horizon",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractProjectPathFromSSHURL(tt.args.gitURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractProjectPathFromSSHURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("extractProjectPathFromSSHURL() got = %v, want %v", got, tt.want)
			}
		})
	}
}
