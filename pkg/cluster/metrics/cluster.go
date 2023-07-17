package metrics

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/horizoncd/horizon/core/common"
	"github.com/horizoncd/horizon/lib/q"
	appmanager "github.com/horizoncd/horizon/pkg/application/manager"
	amodels "github.com/horizoncd/horizon/pkg/application/models"
	userauth "github.com/horizoncd/horizon/pkg/authentication/user"
	clustermanager "github.com/horizoncd/horizon/pkg/cluster/manager"
	cmodels "github.com/horizoncd/horizon/pkg/cluster/models"
	groupmanager "github.com/horizoncd/horizon/pkg/group/manager"
	gmodels "github.com/horizoncd/horizon/pkg/group/models"
	tagmanager "github.com/horizoncd/horizon/pkg/tag/manager"
	tagmodels "github.com/horizoncd/horizon/pkg/tag/models"
	"github.com/horizoncd/horizon/pkg/util/log"
)

const Parrallelism = 100

func NewMetrics(
	appMgr appmanager.Manager,
	clusterMgr clustermanager.Manager,
	tagMgr tagmanager.Manager,
	groupMgr groupmanager.Manager,
) {
	prometheus.MustRegister(&Collector{
		appMgr,
		clusterMgr,
		tagMgr,
		groupMgr,
	})
}

type groupWithFullPath struct {
	*gmodels.Group
	FullPath string
}

type item struct {
	app     *amodels.Application
	cluster *cmodels.Cluster
	group   *groupWithFullPath
	tags    []*tagmodels.Tag
}

type Collector struct {
	// managers *managerparam.Manager
	appMgr     appmanager.Manager
	clusterMgr clustermanager.Manager
	tagMgr     tagmanager.Manager
	groupMgr   groupmanager.Manager
}

const (
	horizonClusterInfo   = "horizon_cluster_info"
	horizonClusterLabels = "horizon_cluster_labels"
)

func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	clusterInfo := prometheus.NewDesc(
		horizonClusterInfo,
		"A metric with a constant '1' value labeled by cluster, application, group, etc.",
		[]string{},
		nil,
	)
	ch <- clusterInfo

	clusterLabels := prometheus.NewDesc(
		horizonClusterLabels,
		"A metric with a constant '1' value labeled by cluster and tags",
		[]string{},
		nil,
	)
	ch <- clusterLabels
}

func (c *Collector) getClusters() []*item {
	ctx := context.Background()

	// dummy user
	// nolint
	ctx = context.WithValue(ctx, common.UserContextKey(), &userauth.DefaultInfo{ID: 0})

	_, deletedClusters, err := c.clusterMgr.List(ctx, &q.Query{
		Keywords: q.KeyWords{
			common.ClusterQueryUpdatedAfter: time.Now().Add(-1 * time.Hour),
			common.ClusterQueryOnlyDeleted:  true,
		},
		Sorts:             []*q.Sort{{Key: "c.created_at", DESC: true}},
		WithoutPagination: true,
	})
	if err != nil {
		log.Errorf(ctx, "Failed to get deleted clusters: %v", err)
		return []*item{}
	}

	_, clusters, err := c.clusterMgr.List(ctx, &q.Query{
		Sorts:             []*q.Sort{{Key: "c.created_at", DESC: true}},
		WithoutPagination: true,
	})
	if err != nil {
		log.Errorf(ctx, "Failed to get clusters: %v", err)
		return []*item{}
	}

	clusters = append(clusters, deletedClusters...)

	m := make(map[uint]struct{})
	appIDs := make([]uint, 0)
	for _, cluster := range clusters {
		if _, ok := m[cluster.ApplicationID]; ok {
			continue
		}
		m[cluster.ApplicationID] = struct{}{}
		appIDs = append(appIDs, cluster.ApplicationID)
	}

	_, apps, err := c.appMgr.List(ctx, []uint{}, &q.Query{
		Keywords: q.KeyWords{
			common.ApplicationQueryID:          appIDs,
			common.ApplicationQueryWithDeleted: true,
		},
		WithoutPagination: true,
	})
	if err != nil {
		log.Errorf(ctx, "Failed to get applications: %v", err)
		return []*item{}
	}

	appMap := make(map[uint]*amodels.Application)
	for _, app := range apps {
		appMap[app.ID] = app
	}

	groups, err := c.groupMgr.GetAll(ctx)
	if err != nil {
		log.Errorf(ctx, "Failed to get groups: %v", err)
		return []*item{}
	}

	groupMap := make(map[uint]*gmodels.Group)
	for _, group := range groups {
		groupMap[group.ID] = group
	}

	items := make([]*item, 0)
	// wg := sync.WaitGroup{}
	// parralleLimit := make(chan struct{}, Parrallelism)
	for _, cluster := range clusters {
		app := appMap[cluster.ApplicationID]
		if app == nil {
			continue
		}
		g := &groupWithFullPath{
			Group: groupMap[app.GroupID],
		}
		if g.Group == nil {
			g = nil
		} else {
			g.genFullPath(groupMap)
		}
		itm := &item{
			app:     app,
			cluster: cluster.Cluster,
			group:   g,
		}
		items = append(items, itm)
		// wg.Add(1)
		// go func(itm *item) {
		// 	parralleLimit <- struct{}{}
		// 	defer func() {
		// 		<-parralleLimit
		// 	}()
		// 	defer wg.Done()
		itm.listTags(c.tagMgr)
		// 	}(itm)
	}
	// wg.Wait()
	return items
}

func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	items := c.getClusters()

	for _, item := range items {
		c.CollectClusterInfo(item, ch)

		c.CollectClusterLabel(item, ch)
	}
}

func (*Collector) CollectClusterInfo(itm *item, ch chan<- prometheus.Metric) {
	infoKeys := []string{
		"cluster", "application", "group", "region", "template", "environment",
	}
	c, a, g := itm.cluster, itm.app, itm.group
	if c == nil || a == nil || g == nil {
		return
	}
	infoLabels := prometheus.Labels{
		"cluster":     c.Name,
		"application": a.Name,
		"group":       g.FullPath,
		"region":      c.RegionName,
		"template":    c.Template,
		"environment": c.EnvironmentName,
	}
	horizonClusterLabels := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: horizonClusterInfo,
		Help: "A metric with a constant '1' value labeled by cluster, application, group, etc.",
	},
		infoKeys,
	)

	horizonClusterLabels.With(infoLabels).Set(1)
	horizonClusterLabels.Collect(ch)
}

func (*Collector) CollectClusterLabel(itm *item, ch chan<- prometheus.Metric) {
	replacer := strings.NewReplacer("-", "_", ".", "_", "/", "_")
	pattern := regexp.MustCompile(`[a-zA-Z_][a-zA-Z0-9_]{0,239}`)
	validLabel := func(value string) bool {
		return pattern.FindString(value) == value
	}

	c, a, g := itm.cluster, itm.app, itm.group
	if c == nil || a == nil || g == nil {
		return
	}
	tagLabels := prometheus.Labels{
		"cluster":     c.Name,
		"application": a.Name,
	}
	for _, tag := range itm.tags {
		label := fmt.Sprintf("label_%s", tag.Key)
		label = replacer.Replace(label)
		if !validLabel(label) {
			log.Warningf(context.Background(), "Invalid label name: %s", label)
			continue
		}
		tagLabels[label] = tag.Value
	}
	tagLabelKeys := make([]string, 0, len(tagLabels))
	for k := range tagLabels {
		tagLabelKeys = append(tagLabelKeys, k)
	}
	horizonClusterLabels := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: horizonClusterLabels,
		Help: "A metric with a constant '1' value labeled by cluster and tags",
	},
		tagLabelKeys,
	)

	horizonClusterLabels.With(tagLabels).Set(1)
	horizonClusterLabels.Collect(ch)
}

func (g *groupWithFullPath) genFullPath(groups map[uint]*gmodels.Group) {
	if g == nil {
		return
	}
	ids := strings.SplitN(g.Group.TraversalIDs, ",", -1)
	builder := strings.Builder{}
	for _, idStr := range ids {
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			log.Warningf(context.Background(), "Failed to parse group id: %v", err)
			return
		}
		g := groups[uint(id)]
		builder.WriteByte('/')
		if g != nil {
			builder.WriteString(groups[uint(id)].Name)
		}
	}
	g.FullPath = builder.String()
}

func (i *item) listTags(tagMgr tagmanager.Manager) {
	if i == nil || i.cluster == nil {
		return
	}
	ctx := context.Background()
	tags, err := tagMgr.ListByResourceTypeID(ctx, common.ResourceCluster, i.cluster.ID)
	if err != nil {
		log.Errorf(ctx, "Failed to get tags: %v", err)
		return
	}
	i.tags = tags
}
