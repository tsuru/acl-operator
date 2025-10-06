package controllers

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var subReconcilerTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
	Name: "subreconciler_reconcile_total",
	Help: "Total number of reconciliations per subcontroller",
}, []string{"controller", "subcontroller", "result"})

var subReconcilerTime = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "subreconciler_reconcile_time_seconds",
	Help:    "Length of time per reconciliation per subcontroller per controller",
	Buckets: prometheus.ExponentialBuckets(0.001, 2, 20),
}, []string{"controller", "subccontroller"})

func init() {
	metrics.Registry.MustRegister(subReconcilerTotal)
	metrics.Registry.MustRegister(subReconcilerTime)
}
