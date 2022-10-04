package collectors

import (
	"context"
	"log"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/spotinst/spotinst-sdk-go/service/mcs"
	"github.com/spotinst/spotinst-sdk-go/service/ocean/providers/aws"
	"github.com/spotinst/spotinst-sdk-go/spotinst"
)

// A prometheus collector for the cost of Spotinst Ocean clusters on AWS.
type OceanAWSClusterCostsCollector struct {
	ctx             context.Context
	client          mcs.Service
	clusters        []*aws.Cluster
	clusterCost     *prometheus.Desc
	namespaceCost   *prometheus.Desc
	deploymentCost  *prometheus.Desc
	daemonSetCost   *prometheus.Desc
	statefulSetCost *prometheus.Desc
	jobCost         *prometheus.Desc
}

// Creates a new OceanAWSClusterCostsCollector for collecting the costs of the
// provided list of Ocean clusters.
func NewOceanAWSClusterCostsCollector(ctx context.Context, client mcs.Service, clusters []*aws.Cluster) *OceanAWSClusterCostsCollector {
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
			log.Println(err)
			continue
		}

		c.collectClusterCosts(ch, *cluster.ID, output.ClusterCosts)
	}
}

func (c *OceanAWSClusterCostsCollector) collectClusterCosts(ch chan<- prometheus.Metric, oceanID string, clusters []*mcs.ClusterCost) {
	for _, cluster := range clusters {
		ch <- prometheus.MustNewConstMetric(
			c.clusterCost,
			prometheus.GaugeValue,
			*cluster.TotalCost,
			oceanID,
		)

		c.collectNamespaceCosts(ch, oceanID, cluster.Namespaces)
	}
}

func (c *OceanAWSClusterCostsCollector) collectNamespaceCosts(ch chan<- prometheus.Metric, oceanID string, namespaces []*mcs.Namespace) {
	for _, namespace := range namespaces {
		labelValues := []string{oceanID, *namespace.Namespace}

		ch <- prometheus.MustNewConstMetric(
			c.namespaceCost,
			prometheus.GaugeValue,
			*namespace.Cost,
			labelValues...,
		)

		collectWorkloadCosts(ch, oceanID, namespace.Deployments, c.deploymentCost)
		collectWorkloadCosts(ch, oceanID, namespace.DaemonSets, c.daemonSetCost)
		collectWorkloadCosts(ch, oceanID, namespace.StatefulSets, c.statefulSetCost)
		collectWorkloadCosts(ch, oceanID, namespace.Jobs, c.jobCost)
	}
}

func collectWorkloadCosts(ch chan<- prometheus.Metric, oceanID string, resources []*mcs.Resource, desc *prometheus.Desc) {
	for _, resource := range resources {
		labelValues := []string{oceanID, *resource.Namespace, *resource.Name}

		ch <- prometheus.MustNewConstMetric(
			desc,
			prometheus.GaugeValue,
			*resource.Cost,
			labelValues...,
		)
	}
}
