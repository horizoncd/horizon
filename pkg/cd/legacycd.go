package cd

import "context"

// nolint
//
//go:generate mockgen -source=$GOFILE -destination=../../mock/pkg/cd/lagacycd_mock.go -package=mock_cd -aux_files=github.com/horizoncd/horizon/pkg/cd=cd.go
type LegacyCD interface {
	CD

	GetClusterStateV1(ctx context.Context, params *GetClusterStateParams) (*ClusterState, error)
}
