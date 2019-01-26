package prometheus

import (
	"context"
	"github.com/micro/go-micro/server"
	"github.com/prometheus/client_golang/prometheus"
)

func NewHandlerWrapper() server.HandlerWrapper {
	opsCounter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "go_micro_requests_total",
			Help: "How many go-miro requests processed, partitioned by method and status",
		},
		[]string{"method", "status"},
	)

	timeCounter := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "go_micro_request_durations_microseconds",
			Help: "Service method request latencies in microseconds",
		},
		[]string{"method"},
	)

	prometheus.MustRegister(opsCounter)
	prometheus.MustRegister(timeCounter)

	return func(fn server.HandlerFunc) server.HandlerFunc {
		return func(ctx context.Context, req server.Request, rsp interface{}) error {
			name := req.Endpoint()

			timer := prometheus.NewTimer(prometheus.ObserverFunc(func(v float64) {
				timeCounter.WithLabelValues(name).Observe(v)
			}))
			defer timer.ObserveDuration()

			err := fn(ctx, req, rsp)
			if err == nil {
				opsCounter.WithLabelValues(name, "success").Inc()
			} else {
				opsCounter.WithLabelValues(name, "fail").Inc()
			}

			return err
		}
	}
}
