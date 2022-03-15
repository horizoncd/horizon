package manager

import (
	"context"
	"fmt"

	he "g.hz.netease.com/horizon/core/errors"

	harbordao "g.hz.netease.com/horizon/pkg/harbor/dao"
	harbormodels "g.hz.netease.com/horizon/pkg/harbor/models"
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
	// ListRegionEntities list all region entity
	ListRegionEntities(ctx context.Context) ([]*models.RegionEntity, error)
	// GetRegionEntity get region entity, todo(gjq) add cache
	GetRegionEntity(ctx context.Context, regionName string) (*models.RegionEntity, error)
}

type manager struct {
	regionDAO     regiondao.DAO
	k8sClusterDAO k8sclusterdao.DAO
	harborDAO     harbordao.DAO
}

func New() Manager {
	return &manager{
		regionDAO:     regiondao.NewDAO(),
		k8sClusterDAO: k8sclusterdao.NewDAO(),
		harborDAO:     harbordao.NewDAO(),
	}
}

func (m *manager) Create(ctx context.Context, region *models.Region) (*models.Region, error) {
	return m.regionDAO.Create(ctx, region)
}

func (m *manager) ListAll(ctx context.Context) ([]*models.Region, error) {
	return m.regionDAO.ListAll(ctx)
}

func (m *manager) ListRegionEntities(ctx context.Context) (ret []*models.RegionEntity, err error) {
	var regions []*models.Region
	regions, err = m.regionDAO.ListAll(ctx)
	if err != nil {
		return
	}

	k8sClusterMap, err := m.getK8SClusterMap(ctx)
	if err != nil {
		return nil, err
	}

	harborMap, err := m.getHarborMap(ctx)
	if err != nil {
		return nil, err
	}

	for _, region := range regions {
		k8sCluster, ok := k8sClusterMap[region.K8SClusterID]
		if !ok {
			return nil, fmt.Errorf("k8sCluster with ID: %v of region: %v is not found",
				region.K8SClusterID, region.Name)
		}
		harbor, ok := harborMap[region.HarborID]
		if !ok {
			return nil, fmt.Errorf("harbor with ID: %v of region: %v is not found",
				region.HarborID, region.Name)
		}
		ret = append(ret, &models.RegionEntity{
			Region:     region,
			K8SCluster: k8sCluster,
			Harbor:     harbor,
		})
	}
	return
}

func (m *manager) GetRegionEntity(ctx context.Context,
	regionName string) (*models.RegionEntity, error) {
	region, err := m.regionDAO.GetRegion(ctx, regionName)
	if err != nil {
		return nil, err
	}

	k8sClusterMap, err := m.getK8SClusterMap(ctx)
	if err != nil {
		return nil, err
	}

	harborMap, err := m.getHarborMap(ctx)
	if err != nil {
		return nil, err
	}

	k8sCluster, ok := k8sClusterMap[region.K8SClusterID]
	if !ok {
		return nil, he.NewErrNotFound(he.K8SCluster,
			fmt.Sprintf("k8sCluster with ID: %v of region: %v is not found",
				region.K8SClusterID, region.Name))
	}

	harbor, ok := harborMap[region.HarborID]
	if !ok {
		return nil, he.NewErrNotFound(he.Harbor,
			fmt.Sprintf("harbor with ID: %v of region: %v is not found",
				region.HarborID, region.Name))
	}

	return &models.RegionEntity{
		Region:     region,
		K8SCluster: k8sCluster,
		Harbor:     harbor,
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

func (m *manager) getHarborMap(ctx context.Context) (map[uint]*harbormodels.Harbor, error) {
	harbors, err := m.harborDAO.ListAll(ctx)
	if err != nil {
		return nil, err
	}

	harborMap := make(map[uint]*harbormodels.Harbor)
	for _, harbor := range harbors {
		harborMap[harbor.ID] = harbor
	}
	return harborMap, nil
}
