package cd

import (
	"context"
	"testing"
	"time"

	"g.hz.netease.com/horizon/pkg/util/log"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCompareRevision(t *testing.T) {
	ctx := log.WithContext(context.Background(), "TestCompareRevision")
	type args struct {
		ctx context.Context
		rs1 *appsv1.ReplicaSet
		rs2 *appsv1.ReplicaSet
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "compare with the same rollout",
			args: args{
				ctx: ctx,
				rs1: &appsv1.ReplicaSet{
					ObjectMeta: metav1.ObjectMeta{
						Name: "rs1",
						Annotations: map[string]string{
							"rollout.argoproj.io/revision": "1",
						},
						OwnerReferences: []metav1.OwnerReference{
							{
								UID: "111",
							},
						},
					},
				},
				rs2: &appsv1.ReplicaSet{
					ObjectMeta: metav1.ObjectMeta{
						Name: "rs2",
						Annotations: map[string]string{
							"rollout.argoproj.io/revision": "2",
						},
						OwnerReferences: []metav1.OwnerReference{
							{
								UID: "111",
							},
						},
					},
				},
			},
			want: false,
		},
		{
			name: "compare with the same deployment",
			args: args{
				ctx: ctx,
				rs1: &appsv1.ReplicaSet{
					ObjectMeta: metav1.ObjectMeta{
						Name: "rs1",
						Annotations: map[string]string{
							"deployment.kubernetes.io/revision": "2",
						},
						OwnerReferences: []metav1.OwnerReference{
							{
								UID: "111",
							},
						},
					},
				},
				rs2: &appsv1.ReplicaSet{
					ObjectMeta: metav1.ObjectMeta{
						Name: "rs2",
						Annotations: map[string]string{
							"deployment.kubernetes.io/revision": "1",
						},
						OwnerReferences: []metav1.OwnerReference{
							{
								UID: "111",
							},
						},
					},
				},
			},
			want: true,
		},
		{
			name: "compare with different deployment",
			args: args{
				ctx: ctx,
				rs1: &appsv1.ReplicaSet{
					ObjectMeta: metav1.ObjectMeta{
						Name: "rs1",
						Annotations: map[string]string{
							"deployment.kubernetes.io/revision": "1",
						},
						CreationTimestamp: metav1.Time{Time: time.Now().Add(time.Hour)},
						OwnerReferences: []metav1.OwnerReference{
							{
								UID: "111",
							},
						},
					},
				},
				rs2: &appsv1.ReplicaSet{
					ObjectMeta: metav1.ObjectMeta{
						Name: "rs2",
						Annotations: map[string]string{
							"deployment.kubernetes.io/revision": "2",
						},
						CreationTimestamp: metav1.Time{Time: time.Now()},
						OwnerReferences: []metav1.OwnerReference{
							{
								UID: "222",
							},
						},
					},
				},
			},
			want: true,
		},
		{
			name: "compare with different deployment 2",
			args: args{
				ctx: ctx,
				rs1: &appsv1.ReplicaSet{
					ObjectMeta: metav1.ObjectMeta{
						Name: "rs1",
						Annotations: map[string]string{
							"deployment.kubernetes.io/revision": "1",
						},
						CreationTimestamp: metav1.Time{Time: time.Now().Add(time.Hour)},
						OwnerReferences: []metav1.OwnerReference{
							{
								UID: "111",
							},
						},
					},
				},
				rs2: &appsv1.ReplicaSet{
					ObjectMeta: metav1.ObjectMeta{
						Name: "rs2",
						Annotations: map[string]string{
							"deployment.kubernetes.io/revision": "2",
						},
						CreationTimestamp: metav1.Time{Time: time.Now().Add(2 * time.Hour)},
						OwnerReferences: []metav1.OwnerReference{
							{
								UID: "222",
							},
						},
					},
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CompareRevision(tt.args.ctx, tt.args.rs1, tt.args.rs2); got != tt.want {
				t.Errorf("CompareRevision() = %v, want %v", got, tt.want)
			}
		})
	}
}
