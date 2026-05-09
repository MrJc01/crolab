package gateway

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	ActiveSessionsGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "crolab_active_sessions",
			Help: "Current number of active WebSocket to ZeroMQ sessions",
		},
	)
	MessagesRoutedCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "crolab_messages_routed_total",
			Help: "Total number of messages routed through the gateway",
		},
		[]string{"direction"}, // "frontend_to_kernel", "kernel_to_frontend"
	)
)

func init() {
	prometheus.MustRegister(ActiveSessionsGauge)
	prometheus.MustRegister(MessagesRoutedCounter)
}

// MetricsHandler retorna o endpoint HTTP para o Prometheus raspar os dados
func MetricsHandler() http.Handler {
	return promhttp.Handler()
}
