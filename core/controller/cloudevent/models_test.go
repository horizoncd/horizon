package cloudevent

import (
	"testing"

	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	"knative.dev/pkg/apis"
	duckv1beta1 "knative.dev/pkg/apis/duck/v1beta1"
)

func TestWrappedPipelineRun_IsFinished(t *testing.T) {
	type fields struct {
		PipelineRun *v1beta1.PipelineRun
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "true1",
			fields: fields{
				PipelineRun: &v1beta1.PipelineRun{
					Status: v1beta1.PipelineRunStatus{
						Status: duckv1beta1.Status{
							Conditions: duckv1beta1.Conditions(
								apis.Conditions{
									{
										Type:   apis.ConditionSucceeded,
										Reason: string(v1beta1.PipelineRunReasonSuccessful),
									},
								},
							),
						},
					},
				},
			},
			want: true,
		},
		{
			name: "true2",
			fields: fields{
				PipelineRun: &v1beta1.PipelineRun{
					Status: v1beta1.PipelineRunStatus{
						Status: duckv1beta1.Status{
							Conditions: duckv1beta1.Conditions(
								apis.Conditions{
									{
										Type:   apis.ConditionSucceeded,
										Reason: string(v1beta1.PipelineRunReasonCancelled),
									},
								},
							),
						},
					},
				},
			},
			want: true,
		},
		{
			name: "true3",
			fields: fields{
				PipelineRun: &v1beta1.PipelineRun{
					Status: v1beta1.PipelineRunStatus{
						Status: duckv1beta1.Status{
							Conditions: duckv1beta1.Conditions(
								apis.Conditions{
									{
										Type:   apis.ConditionSucceeded,
										Reason: string(v1beta1.PipelineRunReasonTimedOut),
									},
								},
							),
						},
					},
				},
			},
			want: true,
		},
		{
			name: "true4",
			fields: fields{
				PipelineRun: &v1beta1.PipelineRun{
					Status: v1beta1.PipelineRunStatus{
						Status: duckv1beta1.Status{
							Conditions: duckv1beta1.Conditions(
								apis.Conditions{
									{
										Type:   apis.ConditionSucceeded,
										Reason: string(v1beta1.PipelineRunReasonFailed),
									},
								},
							),
						},
					},
				},
			},
			want: true,
		},
		{
			name: "true5",
			fields: fields{
				PipelineRun: &v1beta1.PipelineRun{
					Status: v1beta1.PipelineRunStatus{
						Status: duckv1beta1.Status{
							Conditions: duckv1beta1.Conditions(
								apis.Conditions{
									{
										Type:   apis.ConditionSucceeded,
										Reason: string(v1beta1.PipelineRunReasonCompleted),
									},
								},
							),
						},
					},
				},
			},
			want: true,
		},
		{
			name: "false1",
			fields: fields{
				PipelineRun: &v1beta1.PipelineRun{
					Status: v1beta1.PipelineRunStatus{
						Status: duckv1beta1.Status{
							Conditions: duckv1beta1.Conditions(
								apis.Conditions{
									{
										Type:   apis.ConditionSucceeded,
										Reason: string(v1beta1.PipelineRunReasonRunning),
									},
								},
							),
						},
					},
				},
			},
			want: false,
		},
		{
			name: "false2",
			fields: fields{
				PipelineRun: &v1beta1.PipelineRun{
					Status: v1beta1.PipelineRunStatus{
						Status: duckv1beta1.Status{
							Conditions: duckv1beta1.Conditions(
								apis.Conditions{
									{
										Type:   apis.ConditionSucceeded,
										Reason: string(v1beta1.PipelineRunReasonStarted),
									},
								},
							),
						},
					},
				},
			},
			want: false,
		},
		{
			name: "false3",
			fields: fields{
				PipelineRun: &v1beta1.PipelineRun{
					Status: v1beta1.PipelineRunStatus{
						Status: duckv1beta1.Status{
							Conditions: duckv1beta1.Conditions(
								apis.Conditions{
									{
										Type:   apis.ConditionSucceeded,
										Reason: string(v1beta1.PipelineRunReasonStopping),
									},
								},
							),
						},
					},
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wpr := &WrappedPipelineRun{
				PipelineRun: tt.fields.PipelineRun,
			}
			if got := wpr.IsFinished(); got != tt.want {
				t.Errorf("IsFinished() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWrappedPipelineRun_Reason(t *testing.T) {
	type fields struct {
		PipelineRun *v1beta1.PipelineRun
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "empty",
			fields: fields{
				PipelineRun: nil,
			},
			want: "",
		},
		{
			name: "success",
			fields: fields{
				PipelineRun: &v1beta1.PipelineRun{
					Status: v1beta1.PipelineRunStatus{
						Status: duckv1beta1.Status{
							Conditions: duckv1beta1.Conditions(
								apis.Conditions{
									{
										Type:    apis.ConditionSucceeded,
										Reason:  string(v1beta1.PipelineRunReasonSuccessful),
										Message: "Succeeded",
									},
								},
							),
						},
					},
				},
			},
			want: string(v1beta1.PipelineRunReasonSuccessful),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wpr := &WrappedPipelineRun{
				PipelineRun: tt.fields.PipelineRun,
			}
			if got := wpr.Reason(); got != tt.want {
				t.Errorf("Reason() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWrappedPipelineRun_Message(t *testing.T) {
	type fields struct {
		PipelineRun *v1beta1.PipelineRun
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "empty",
			fields: fields{
				PipelineRun: nil,
			},
			want: "",
		},
		{
			name: "success",
			fields: fields{
				PipelineRun: &v1beta1.PipelineRun{
					Status: v1beta1.PipelineRunStatus{
						Status: duckv1beta1.Status{
							Conditions: duckv1beta1.Conditions(
								apis.Conditions{
									{
										Type:    apis.ConditionSucceeded,
										Reason:  string(v1beta1.PipelineRunReasonSuccessful),
										Message: "Succeeded",
									},
								},
							),
						},
					},
				},
			},
			want: "Succeeded",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wpr := &WrappedPipelineRun{
				PipelineRun: tt.fields.PipelineRun,
			}
			if got := wpr.Message(); got != tt.want {
				t.Errorf("Message() = %v, want %v", got, tt.want)
			}
		})
	}
}
