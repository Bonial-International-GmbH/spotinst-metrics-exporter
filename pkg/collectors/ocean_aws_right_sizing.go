package collectors

import (
	"context"
	"log"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/spotinst/spotinst-sdk-go/service/ocean/providers/aws"
)

// A prometheus collector for the right sizing suggestions of Spotinst Ocean clusters on AWS.
type OceanAWSRightSizingCollector struct {
	ctx                      context.Context
	client                   aws.Service
	clusters                 []*aws.Cluster
	requestedCPU             *prometheus.Desc
	suggestedCPU             *prometheus.Desc
	requestedMemory          *prometheus.Desc
	suggestedMemory          *prometheus.Desc
	requestedContainerCPU    *prometheus.Desc
	suggestedContainerCPU    *prometheus.Desc
	requestedContainerMemory *prometheus.Desc
	suggestedContainerMemory *prometheus.Desc
}

// Creates a new OceanAWSRightSizingCollector for collecting the right sizing
// suggestions for the provided list of Ocean clusters.
func NewOceanAWSRightSizingCollector(
	ctx context.Context,
	client aws.Service,
	clusters []*aws.Cluster,
) *OceanAWSRightSizingCollector {
	collector := &OceanAWSRightSizingCollector{
		ctx:      ctx,
		client:   client,
		clusters: clusters,
		requestedCPU: prometheus.NewDesc(
			"spotinst_ocean_aws_right_sizing_cpu_requested",
			"The number of actual CPU units requested by a workload",
			[]string{"ocean", "resource", "namespace", "name"},
			nil,
		),
		suggestedCPU: prometheus.NewDesc(
			"spotinst_ocean_aws_right_sizing_cpu_suggested",
			"The number of CPU units suggested for a workload",
			[]string{"ocean", "resource", "namespace", "name"},
			nil,
		),
		requestedMemory: prometheus.NewDesc(
			"spotinst_ocean_aws_right_sizing_memory_requested",
			"The number of actual memory units requested by a workload",
			[]string{"ocean", "resource", "namespace", "name"},
			nil,
		),
		suggestedMemory: prometheus.NewDesc(
			"spotinst_ocean_aws_right_sizing_memory_suggested",
			"The number of memory units suggested for a workload",
			[]string{"ocean", "resource", "namespace", "name"},
			nil,
		),
		requestedContainerCPU: prometheus.NewDesc(
			"spotinst_ocean_aws_right_sizing_container_cpu_requested",
			"The number of actual CPU units requested by a workload's container",
			[]string{"ocean", "resource", "namespace", "name", "container"},
			nil,
		),
		suggestedContainerCPU: prometheus.NewDesc(
			"spotinst_ocean_aws_right_sizing_container_cpu_suggested",
			"The number of CPU units suggested for a workload's container",
			[]string{"ocean", "resource", "namespace", "name", "container"},
			nil,
		),
		requestedContainerMemory: prometheus.NewDesc(
			"spotinst_ocean_aws_right_sizing_container_memory_requested",
			"The number of actual memory units requested by a workload's container",
			[]string{"ocean", "resource", "namespace", "name", "container"},
			nil,
		),
		suggestedContainerMemory: prometheus.NewDesc(
			"spotinst_ocean_aws_right_sizing_container_memory_suggested",
			"The number of memory units suggested for a workload's container",
			[]string{"ocean", "resource", "namespace", "name", "container"},
			nil,
		),
	}

	return collector
}

// Describe implements the prometheus.Collector interface.
func (c *OceanAWSRightSizingCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.requestedCPU
	ch <- c.suggestedCPU
	ch <- c.requestedMemory
	ch <- c.suggestedMemory
	ch <- c.requestedContainerCPU
	ch <- c.suggestedContainerCPU
	ch <- c.requestedContainerMemory
	ch <- c.suggestedContainerMemory
}

// Collect implements the prometheus.Collector interface.
func (c *OceanAWSRightSizingCollector) Collect(ch chan<- prometheus.Metric) {
	for _, cluster := range c.clusters {
		input := &aws.ListOceanResourceSuggestionsInput{
			OceanID: cluster.ID,
		}

		output, err := c.client.ListOceanResourceSuggestions(c.ctx, input)
		if err != nil {
			log.Println(err)
			continue
		}

		c.collectSuggestions(ch, output.Suggestions, *cluster.ID)
	}
}

func (c *OceanAWSRightSizingCollector) collectSuggestions(
	ch chan<- prometheus.Metric,
	suggestions []*aws.ResourceSuggestion,
	oceanID string,
) {
	for _, suggestion := range suggestions {
		labelValues := []string{oceanID, *suggestion.ResourceType, *suggestion.Namespace, *suggestion.ResourceName}

		collectGaugeValue(ch, c.requestedCPU, *suggestion.RequestedCPU, labelValues)
		collectGaugeValue(ch, c.suggestedCPU, *suggestion.SuggestedCPU, labelValues)
		collectGaugeValue(ch, c.requestedMemory, *suggestion.RequestedMemory, labelValues)
		collectGaugeValue(ch, c.suggestedMemory, *suggestion.SuggestedMemory, labelValues)

		c.collectContainerSuggestions(ch, suggestion.Containers, labelValues)
	}
}

func (c *OceanAWSRightSizingCollector) collectContainerSuggestions(
	ch chan<- prometheus.Metric,
	suggestions []*aws.ContainerResourceSuggestion,
	workloadLabelValues []string,
) {
	for _, suggestion := range suggestions {
		labelValues := append(workloadLabelValues, *suggestion.Name)

		collectGaugeValue(ch, c.requestedContainerCPU, *suggestion.RequestedCPU, labelValues)
		collectGaugeValue(ch, c.suggestedContainerCPU, *suggestion.SuggestedCPU, labelValues)
		collectGaugeValue(ch, c.requestedContainerMemory, *suggestion.RequestedMemory, labelValues)
		collectGaugeValue(ch, c.suggestedContainerMemory, *suggestion.SuggestedMemory, labelValues)
	}
}

func collectGaugeValue(
	ch chan<- prometheus.Metric,
	desc *prometheus.Desc,
	value float64,
	labelValues []string,
) {
	ch <- prometheus.MustNewConstMetric(
		desc,
		prometheus.GaugeValue,
		value,
		labelValues...,
	)
}
