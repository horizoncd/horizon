package slo

import (
	"context"
	"reflect"
	"testing"

	"g.hz.netease.com/horizon/pkg/config/grafana"
)

func Test_controller_GetDashboards(t *testing.T) {
	type fields struct {
		GrafanaSLO grafana.SLO
	}
	type args struct {
		in0 context.Context
		env string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Dashboards
	}{
		{
			name: "get",
			fields: fields{
				GrafanaSLO: grafana.SLO{
					OverviewDashboard: "http://grafana.mockserver.org/d/Wz3GSBank/slo-gai-lan?orgId=1" +
						"&kiosk&theme=light&var-env=%s",
					HistoryDashboard: "http://grafana.mockserver.org/d/tKjaD2-nk/horizon-slo-history?orgId=1&kiosk" +
						"&theme=light&var-env=%s",
				},
			},
			args: args{
				env: "test",
			},
			want: &Dashboards{
				Overview: "http://grafana.mockserver.org/d/Wz3GSBank/slo-gai-lan?orgId=1&kiosk&theme=light&var-env=test",
				History: "http://grafana.mockserver.org/d/tKjaD2-nk/horizon-slo-history?orgId=1" +
					"&kiosk&theme=light&var-env=test",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &controller{
				GrafanaSLO: tt.fields.GrafanaSLO,
			}
			if got := c.GetDashboards(tt.args.in0, tt.args.env); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetDashboards() = %v, want %v", got, tt.want)
			}
		})
	}
}
