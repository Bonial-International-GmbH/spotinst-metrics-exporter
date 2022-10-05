package collectors

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/spotinst/spotinst-sdk-go/service/ocean/providers/aws"
)

// OceanAWSResourceSuggestionsLister is the interface for something that can
// list Ocean resource suggestions.
//
// It is implemented by the Spotinst *aws.ServiceOp client.
type OceanAWSResourceSuggestionsLister interface {
	ListOceanResourceSuggestions(
		context.Context,
		*aws.ListOceanResourceSuggestionsInput,
	) (*aws.ListOceanResourceSuggestionsOutput, error)
}

// OceanAWSRightSizingCollector is a prometheus collector for the right sizing
// suggestions of Spotinst Ocean clusters on AWS.
type OceanAWSRightSizingCollector struct {
	ctx                      context.Context
	client                   OceanAWSResourceSuggestionsLister
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

// NewOceanAWSRightSizingCollector creates a new OceanAWSRightSizingCollector
// for collecting the right sizing suggestions for the provided list of Ocean
// clusters.
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
			prometheus.BuildFQName("spotinst", "ocean_aws", "cpu_requested"),
			"The number of actual CPU units requested by a workload",
			[]string{"ocean", "resource", "namespace", "name"},
			nil,
		),
		suggestedCPU: prometheus.NewDesc(
			prometheus.BuildFQName("spotinst", "ocean_aws", "cpu_suggested"),
			"The number of CPU units suggested for a workload",
			[]string{"ocean", "resource", "namespace", "name"},
			nil,
		),
		requestedMemory: prometheus.NewDesc(
			prometheus.BuildFQName("spotinst", "ocean_aws", "memory_requested"),
			"The number of actual memory units requested by a workload",
			[]string{"ocean", "resource", "namespace", "name"},
			nil,
		),
		suggestedMemory: prometheus.NewDesc(
			prometheus.BuildFQName("spotinst", "ocean_aws", "memory_suggested"),
			"The number of memory units suggested for a workload",
			[]string{"ocean", "resource", "namespace", "name"},
			nil,
		),
		requestedContainerCPU: prometheus.NewDesc(
			prometheus.BuildFQName("spotinst", "ocean_aws", "container_cpu_requested"),
			"The number of actual CPU units requested by a workload's container",
			[]string{"ocean", "resource", "namespace", "name", "container"},
			nil,
		),
		suggestedContainerCPU: prometheus.NewDesc(
			prometheus.BuildFQName("spotinst", "ocean_aws", "container_cpu_suggested"),
			"The number of CPU units suggested for a workload's container",
			[]string{"ocean", "resource", "namespace", "name", "container"},
			nil,
		),
		requestedContainerMemory: prometheus.NewDesc(
			prometheus.BuildFQName("spotinst", "ocean_aws", "container_memory_requested"),
			"The number of actual memory units requested by a workload's container",
			[]string{"ocean", "resource", "namespace", "name", "container"},
			nil,
		),
		suggestedContainerMemory: prometheus.NewDesc(
			prometheus.BuildFQName("spotinst", "ocean_aws", "container_memory_suggested"),
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
			logger.Error(err, "failed to list resource suggestions", "ocean", *cluster.ID)
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
