package manager

import (
	"context"
	"fmt"

	k8sclusterdao "g.hz.netease.com/horizon/pkg/k8scluster/dao"
	k8sclustermodels "g.hz.netease.com/horizon/pkg/k8scluster/models"
	regiondao "g.hz.netease.com/horizon/pkg/region/dao"
	"g.hz.netease.com/horizon/pkg/region/models"
)

var (
	// Mgr is the global region manager
	Mgr = New()
)

type Manager interface {
	// Create a region
	Create(ctx context.Context, region *models.Region) (*models.Region, error)
	// ListAll list all regions
	ListAll(ctx context.Context) ([]*models.Region, error)
	// ListAllWithK8SCluster list all regions with K8S cluster
	ListAllWithK8SCluster(ctx context.Context) ([]*models.RegionWithK8SCluster, error)
	// GetRegionWithK8SCluster get region with k8s cluster
	GetRegionWithK8SCluster(ctx context.Context, regionName string) (*models.RegionWithK8SCluster, error)
}

type manager struct {
	regionDAO     regiondao.DAO
	k8sClusterDAO k8sclusterdao.DAO
}

func New() Manager {
	return &manager{
		regionDAO:     regiondao.NewDAO(),
		k8sClusterDAO: k8sclusterdao.NewDAO(),
	}
}

func (m *manager) Create(ctx context.Context, region *models.Region) (*models.Region, error) {
	return m.regionDAO.Create(ctx, region)
}

func (m *manager) ListAll(ctx context.Context) ([]*models.Region, error) {
	return m.regionDAO.ListAll(ctx)
}

func (m *manager) ListAllWithK8SCluster(ctx context.Context) (ret []*models.RegionWithK8SCluster, err error) {
	var regions []*models.Region
	regions, err = m.regionDAO.ListAll(ctx)
	if err != nil {
		return
	}

	k8sClusterMap, err := m.getK8SClusterMap(ctx)
	if err != nil {
		return nil, err
	}

	for _, region := range regions {
		k8sCluster, ok := k8sClusterMap[region.K8SClusterID]
		if !ok {
			return nil, fmt.Errorf("k8sCluster with ID: %v of region: %v is not found",
				region.K8SClusterID, region.Name)
		}
		ret = append(ret, &models.RegionWithK8SCluster{
			Model:       region.Model,
			Name:        region.Name,
			DisplayName: region.DisplayName,
			K8SCluster:  k8sCluster,
			CreatedBy:   region.CreatedBy,
			UpdatedBy:   region.UpdatedBy,
		})
	}
	return
}

func (m *manager) GetRegionWithK8SCluster(ctx context.Context,
	regionName string) (*models.RegionWithK8SCluster, error) {
	region, err := m.regionDAO.GetRegion(ctx, regionName)
	if err != nil {
		return nil, err
	}

	k8sClusterMap, err := m.getK8SClusterMap(ctx)
	if err != nil {
		return nil, err
	}

	k8sCluster, ok := k8sClusterMap[region.K8SClusterID]
	if !ok {
		return nil, fmt.Errorf("k8sCluster with ID: %v of region: %v is not found",
			region.K8SClusterID, region.Name)
	}

	return &models.RegionWithK8SCluster{
		Model:       region.Model,
		Name:        region.Name,
		DisplayName: region.DisplayName,
		K8SCluster:  k8sCluster,
		CreatedBy:   region.CreatedBy,
		UpdatedBy:   region.UpdatedBy,
	}, nil
}

func (m *manager) getK8SClusterMap(ctx context.Context) (map[uint]*k8sclustermodels.K8SCluster, error) {
	k8sClusters, err := m.k8sClusterDAO.ListAll(ctx)
	if err != nil {
		return nil, err
	}

	k8sClusterMap := make(map[uint]*k8sclustermodels.K8SCluster)
	for _, k8sCluster := range k8sClusters {
		k8sClusterMap[k8sCluster.ID] = k8sCluster
	}
	return k8sClusterMap, nil
}
