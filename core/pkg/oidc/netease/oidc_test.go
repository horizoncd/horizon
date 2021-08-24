package netease

import (
	"testing"

	"g.hz.netease.com/horizon/core/pkg/oidc"
)

func TestOIDC_GetRedirectURL(t *testing.T) {
	type fields struct {
		config *oidc.Config
	}
	type args struct {
		requestHost string
		state       string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "",
			fields: fields{
				config: &oidc.Config{
					ClientID:     "cccc",
					ClientSecret: "ssss",
					Endpoint: oidc.Endpoint{
						AuthURL: "https://oidc.com/authorize",
					},
					RedirectURI: "/api/v1/login/callback",
					Scopes:      []string{"openid", "fullname", "nickname", "email"},
				}},
			args: args{
				requestHost: "example.com",
				state:       "abcdedf",
			},
			want: "https://oidc.com/authorize?" +
				"client_id=cccc&" +
				"redirect_uri=http%3A%2F%2Fexample.com%2Fapi%2Fv1%2Flogin%2Fcallback&" +
				"response_type=code&scope=openid+fullname+nickname+email&" +
				"state=abcdedf",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &OIDC{
				config: tt.fields.config,
			}
			if got := o.GetRedirectURL(tt.args.requestHost, tt.args.state); got != tt.want {
				t.Errorf("GetRedirectURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
