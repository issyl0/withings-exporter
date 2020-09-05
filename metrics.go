package main

import "github.com/prometheus/client_golang/prometheus"

var currentWeightMetric = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Name: "withings_current_weight",
		Help: "Shows the latest weight measurement (assumed in kg)",
	},
)

func registerMetrics() {
	prometheus.MustRegister(currentWeightMetric)
}
