package collectors

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/spotinst/spotinst-sdk-go/service/mcs"
	"github.com/spotinst/spotinst-sdk-go/service/ocean/providers/aws"
	"github.com/spotinst/spotinst-sdk-go/spotinst"
)

// OceanAWSClusterCostsFetcher is the interface for fetching Ocean cluster costs.
//
// It is implemented by the Spotinst *mcs.ServiceOp client.
type OceanAWSClusterCostsFetcher interface {
	GetClusterCosts(context.Context, *mcs.ClusterCostInput) (*mcs.ClusterCostOutput, error)
}

// OceanAWSClusterCostsCollector is a prometheus collector for the cost of
// Spotinst Ocean clusters on AWS.
type OceanAWSClusterCostsCollector struct {
	ctx             context.Context
	client          OceanAWSClusterCostsFetcher
	clusters        []*aws.Cluster
	clusterCost     *prometheus.Desc
	namespaceCost   *prometheus.Desc
	deploymentCost  *prometheus.Desc
	daemonSetCost   *prometheus.Desc
	statefulSetCost *prometheus.Desc
	jobCost         *prometheus.Desc
}

// NewOceanAWSClusterCostsCollector creates a new OceanAWSClusterCostsCollector
// for collecting the costs of the provided list of Ocean clusters.
func NewOceanAWSClusterCostsCollector(
	ctx context.Context,
	client mcs.Service,
	clusters []*aws.Cluster,
) *OceanAWSClusterCostsCollector {
	collector := &OceanAWSClusterCostsCollector{
		ctx:      ctx,
		client:   client,
		clusters: clusters,
		clusterCost: prometheus.NewDesc(
			"spotinst_ocean_aws_cluster_cost_total",
			"Total cost of an ocean cluster",
			[]string{"ocean"},
			nil,
		),
		namespaceCost: prometheus.NewDesc(
			"spotinst_ocean_aws_namespace_cost_total",
			"Total cost of a namespace",
			[]string{"ocean", "namespace"},
			nil,
		),
		deploymentCost: prometheus.NewDesc(
			"spotinst_ocean_aws_deployment_cost_total",
			"Total cost of a deployment",
			[]string{"ocean", "namespace", "name"},
			nil,
		),
		daemonSetCost: prometheus.NewDesc(
			"spotinst_ocean_aws_daemonset_cost_total",
			"Total cost of a daemonset",
			[]string{"ocean", "namespace", "name"},
			nil,
		),
		statefulSetCost: prometheus.NewDesc(
			"spotinst_ocean_aws_statefulset_cost_total",
			"Total cost of a statefulset",
			[]string{"ocean", "namespace", "name"},
			nil,
		),
		jobCost: prometheus.NewDesc(
			"spotinst_ocean_aws_job_cost_total",
			"Total cost of a job",
			[]string{"ocean", "namespace", "name"},
			nil,
		),
	}

	return collector
}

// Describe implements the prometheus.Collector interface.
func (c *OceanAWSClusterCostsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.clusterCost
	ch <- c.namespaceCost
	ch <- c.deploymentCost
	ch <- c.daemonSetCost
	ch <- c.statefulSetCost
	ch <- c.jobCost
}

// Collect implements the prometheus.Collector interface.
func (c *OceanAWSClusterCostsCollector) Collect(ch chan<- prometheus.Metric) {
	now := time.Now()
	firstDayOfCurrentMonth := now.AddDate(0, 0, -now.Day()+1)
	firstDayOfNextMonth := now.AddDate(0, 1, -now.Day()+1)

	fromDate := spotinst.String(firstDayOfCurrentMonth.Format("2006-01-02"))
	toDate := spotinst.String(firstDayOfNextMonth.Format("2006-01-02"))

	for _, cluster := range c.clusters {
		input := &mcs.ClusterCostInput{
			ClusterID: cluster.ControllerClusterID,
			FromDate:  fromDate,
			ToDate:    toDate,
		}

		output, err := c.client.GetClusterCosts(c.ctx, input)
		if err != nil {
			logger.Error(err, "failed to fetch cluster costs", "ocean", *cluster.ID)
			continue
		}

		c.collectClusterCosts(ch, output.ClusterCosts, *cluster.ID)
	}
}

func (c *OceanAWSClusterCostsCollector) collectClusterCosts(
	ch chan<- prometheus.Metric,
	clusters []*mcs.ClusterCost,
	oceanID string,
) {
	labelValues := []string{oceanID}

	for _, cluster := range clusters {
		collectGaugeValue(ch, c.clusterCost, *cluster.TotalCost, labelValues)

		c.collectNamespaceCosts(ch, cluster.Namespaces, labelValues)
	}
}

func (c *OceanAWSClusterCostsCollector) collectNamespaceCosts(
	ch chan<- prometheus.Metric,
	namespaces []*mcs.Namespace,
	clusterLabelValues []string,
) {
	for _, namespace := range namespaces {
		labelValues := append(clusterLabelValues, *namespace.Namespace)

		collectGaugeValue(ch, c.namespaceCost, *namespace.Cost, labelValues)

		collectWorkloadCosts(ch, c.deploymentCost, namespace.Deployments, labelValues)
		collectWorkloadCosts(ch, c.daemonSetCost, namespace.DaemonSets, labelValues)
		collectWorkloadCosts(ch, c.statefulSetCost, namespace.StatefulSets, labelValues)
		collectWorkloadCosts(ch, c.jobCost, namespace.Jobs, labelValues)
	}
}

func collectWorkloadCosts(
	ch chan<- prometheus.Metric,
	desc *prometheus.Desc,
	resources []*mcs.Resource,
	namespaceLabelValues []string,
) {
	for _, resource := range resources {
		labelValues := append(namespaceLabelValues, *resource.Name)

		collectGaugeValue(ch, desc, *resource.Cost, labelValues)
	}
}
