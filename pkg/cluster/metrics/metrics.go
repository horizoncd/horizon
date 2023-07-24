package metrics

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/horizoncd/horizon/core/common"
	"github.com/horizoncd/horizon/lib/q"
	amodels "github.com/horizoncd/horizon/pkg/application/models"
	userauth "github.com/horizoncd/horizon/pkg/authentication/user"
	cmodels "github.com/horizoncd/horizon/pkg/cluster/models"
	gmodels "github.com/horizoncd/horizon/pkg/group/models"
	"github.com/horizoncd/horizon/pkg/param/managerparam"
	tagmodels "github.com/horizoncd/horizon/pkg/tag/models"
	"github.com/horizoncd/horizon/pkg/util/log"
)

var (
	replacer = strings.NewReplacer("-", "_", ".", "_", "/", "_")
	pattern  = regexp.MustCompile(`[a-zA-Z_][a-zA-Z0-9_]{0,239}`)
)

func NewMetrics(
	managers *managerparam.Manager,
) {
	prometheus.MustRegister(&Collector{
		managers: managers,
	})
}

type groupWithFullPath struct {
	*gmodels.Group
	FullPath string
}

type appWithTags struct {
	*amodels.Application
	Tags []*tagmodels.Tag
}

type clusterWithTags struct {
	*cmodels.Cluster
	Tags []*tagmodels.Tag
}

type item struct {
	app     *amodels.Application
	cluster *clusterWithTags
	group   *groupWithFullPath
}

type Collector struct {
	managers *managerparam.Manager
}

const (
	horizonClusterInfo   = "horizon_cluster_info"
	horizonClusterLabels = "horizon_cluster_labels"
	horizonAppLabels     = "horizon_application_labels"
)

func (collector *Collector) Describe(ch chan<- *prometheus.Desc) {
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

	appLabels := prometheus.NewDesc(
		horizonAppLabels,
		"A metric with a constant '1' value labeled by application and tags",
		[]string{},
		nil,
	)
	ch <- appLabels
}

func (collector *Collector) getTagsMap(resourceIDs []uint, resourceType string) (map[uint][]*tagmodels.Tag, error) {
	tags, err := collector.managers.TagMgr.ListByResourceTypeIDs(context.Background(),
		resourceType, resourceIDs, false)
	if err != nil {
		return nil, err
	}

	tagsMap := make(map[uint][]*tagmodels.Tag)
	for _, tag := range tags {
		tagsMap[tag.ResourceID] = append(tagsMap[tag.ResourceID], tag)
	}
	return tagsMap, nil
}

func (collector *Collector) getClustersAndApps() ([]*item, []*appWithTags) {
	ctx := context.Background()

	// dummy user
	// nolint
	ctx = context.WithValue(ctx, common.UserContextKey(), &userauth.DefaultInfo{ID: 0})

	_, clusters, err := collector.managers.ClusterMgr.List(ctx, &q.Query{
		Sorts:             []*q.Sort{{Key: "collector.created_at", DESC: true}},
		WithoutPagination: true,
	})
	if err != nil {
		log.Errorf(ctx, "Failed to get clusters: %v", err)
		return []*item{}, []*appWithTags{}
	}

	clusterIDs := make([]uint, 0)
	for _, cluster := range clusters {
		clusterIDs = append(clusterIDs, cluster.ID)
	}

	clusterTagsMap, err := collector.getTagsMap(clusterIDs, common.ResourceCluster)
	if err != nil {
		log.Errorf(ctx, "Failed to get tags: %v", err)
		return []*item{}, []*appWithTags{}
	}

	_, apps, err := collector.managers.ApplicationMgr.List(ctx, []uint{}, &q.Query{
		WithoutPagination: true,
	})
	if err != nil {
		log.Errorf(ctx, "Failed to get applications: %v", err)
		return []*item{}, []*appWithTags{}
	}

	appIDs := make([]uint, 0)
	appMap := make(map[uint]*amodels.Application)
	for _, app := range apps {
		appIDs = append(appIDs, app.ID)
		if _, ok := appMap[app.ID]; ok {
			continue
		}
		appMap[app.ID] = app
	}

	appTagsMap, err := collector.getTagsMap(appIDs, common.ResourceApplication)
	if err != nil {
		log.Errorf(ctx, "Failed to get tags: %v", err)
		return []*item{}, []*appWithTags{}
	}

	groups, err := collector.managers.GroupMgr.GetAll(ctx)
	if err != nil {
		log.Errorf(ctx, "Failed to get groups: %v", err)
		return []*item{}, []*appWithTags{}
	}

	groupMap := make(map[uint]*gmodels.Group)
	for _, group := range groups {
		groupMap[group.ID] = group
	}

	items := make([]*item, 0)
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
			app: app,
			cluster: &clusterWithTags{
				Cluster: cluster.Cluster,
				Tags:    clusterTagsMap[cluster.ID],
			},
			group: g,
		}
		items = append(items, itm)
	}

	appsWithTags := make([]*appWithTags, 0, len(apps))
	for _, app := range apps {
		appsWithTags = append(appsWithTags, &appWithTags{
			Application: app,
			Tags:        appTagsMap[app.ID],
		})
	}
	return items, appsWithTags
}

func (collector *Collector) Collect(ch chan<- prometheus.Metric) {
	clusters, apps := collector.getClustersAndApps()

	for _, item := range clusters {
		collector.CollectClusterInfo(item, ch)

		collector.CollectClusterLabel(item, ch)
	}

	for _, app := range apps {
		collector.CollectAppLabel(app, ch)
	}
}

func (collector *Collector) CollectAppLabel(app *appWithTags, ch chan<- prometheus.Metric) {
	if app == nil || app.Application == nil {
		return
	}

	tagLabels, tagLabelKeys := collector.tagsToMetrics(app.Tags)
	tagLabels["application"] = app.Name

	horizonClusterLabels := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: horizonAppLabels,
		Help: "A metric with a constant '1' value labeled by application and tags",
	},
		append([]string{"application"}, tagLabelKeys...),
	)

	horizonClusterLabels.With(tagLabels).Set(1)
	horizonClusterLabels.Collect(ch)
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

func (collector *Collector) CollectClusterLabel(itm *item, ch chan<- prometheus.Metric) {
	c, a, g := itm.cluster, itm.app, itm.group
	if c == nil || a == nil || g == nil {
		return
	}
	tagLabels, tagLabelKeys := collector.tagsToMetrics(c.Tags)
	tagLabels["cluster"] = c.Name
	tagLabels["application"] = a.Name
	horizonClusterLabels := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: horizonClusterLabels,
		Help: "A metric with a constant '1' value labeled by cluster and tags",
	},
		append([]string{"cluster", "application"}, tagLabelKeys...),
	)

	horizonClusterLabels.With(tagLabels).Set(1)
	horizonClusterLabels.Collect(ch)
}

func (*Collector) tagsToMetrics(tags []*tagmodels.Tag) (prometheus.Labels, []string) {
	tagLabels := prometheus.Labels{}
	for _, tag := range tags {
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
	return tagLabels, tagLabelKeys
}

func (g *groupWithFullPath) genFullPath(groups map[uint]*gmodels.Group) {
	if g == nil {
		return
	}
	ids := strings.SplitN(g.Group.TraversalIDs, ",", -1)
	builder := strings.Builder{}
	for i, idStr := range ids {
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			log.Warningf(context.Background(), "Failed to parse group id: %v", err)
			return
		}
		g := groups[uint(id)]
		if i > 0 {
			builder.WriteByte('/')
		}
		if g != nil {
			builder.WriteString(groups[uint(id)].Name)
		}
	}
	g.FullPath = builder.String()
}

func validLabel(value string) bool {
	return pattern.FindString(value) == value
}
