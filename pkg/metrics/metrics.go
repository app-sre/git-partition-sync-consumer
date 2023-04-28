package metrics

import (
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	SUCCESS = "success"
	FAILURE = "failure"
)

var (
	reconcileSuccessCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "qontract_reconcile_execution_counter_total",
			Help: "Increment by one for each successful reconcile. Used to alert on 'stuck' instance reconciles",
		},
		[]string{
			"shard_id",
			"integration",
		},
	)
	lastReconcileSuccessGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "qontract_reconcile_last_run_status",
			Help: `Whether or not last reconcile run was successful. ` +
				`A reconcile is successful if no errors occur. 0 = success. 1 = failure.`,
		},
		[]string{
			"shard_id",
			"integration",
		},
	)
	executionDurationGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "qontract_reconcile_last_run_seconds",
			Help: "Execution duration of this job in seconds.",
		},
		[]string{
			"shard_id",
			"integration",
		},
	)
)

// register custom metrics and start metrics server
func Start(port string) {
	prometheus.MustRegister(reconcileSuccessCounter)
	prometheus.MustRegister(lastReconcileSuccessGauge)
	prometheus.MustRegister(executionDurationGauge)
	// configure prometheus metrics handler
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
	}()
}

func RecordMetrics(instance string, status int, duration time.Duration) {
	const INTEGRATION = "git-partition-sync-consumer"

	lastReconcileSuccessGauge.With(
		prometheus.Labels{
			"shard_id":    instance,
			"integration": INTEGRATION,
		}).Set(float64(status))

	// only inc counter metric for successful reconciles
	if status == 0 {
		reconcileSuccessCounter.With(
			prometheus.Labels{
				"shard_id":    instance,
				"integration": INTEGRATION,
			}).Inc()
	}

	executionDurationGauge.With(
		prometheus.Labels{
			"shard_id":    instance,
			"integration": INTEGRATION,
		}).Set(duration.Seconds())
}
