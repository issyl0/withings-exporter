package main

import "github.com/prometheus/client_golang/prometheus"

var currentWeightMetric = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Name: "withings_current_weight",
		Help: "Shows the latest weight measurement (in kg)",
	},
)

var hydrationMetric = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Name: "withings_current_hydration",
		Help: "Shows the latest hydration measurement (in kg)",
	},
)

func registerMetrics() {
	prometheus.MustRegister(currentWeightMetric)
	prometheus.MustRegister(hydrationMetric)
}
