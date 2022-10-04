package collectors

import (
	"github.com/Bonial-International-GmbH/spotinst-metrics-exporter/pkg/log"
	"github.com/prometheus/client_golang/prometheus"
)

var logger = log.Logger().WithName("collectors")

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
