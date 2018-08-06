package resource

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
)

type Metrics interface {
	Success()
	Failure()
}

func NewMetrics(request Request) Metrics {
	return &prometheusMetrics{
		request:   request,
		startTime: time.Now(),
		successCounter: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "halfpipe_cf_plugin_success",
			Help: "Successful invocation of halfpipe cf plugin",
		}),
		failureCounter: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "halfpipe_cf_plugin_failure",
			Help: "Unsuccessful invocation of halfpipe cf plugin",
		}),
		timerHistogram: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "halfpipe_cf_plugin_duration_seconds",
			Help:    "Time taken in seconds for successful invocation of halfpipe cf plugin",
			Buckets: []float64{5, 10, 20, 30, 40, 50, 60, 90, 120, 180},
		}),
	}
}

type prometheusMetrics struct {
	request        Request
	startTime      time.Time
	successCounter prometheus.Counter
	failureCounter prometheus.Counter
	timerHistogram prometheus.Histogram
}

func (p *prometheusMetrics) Success() {
	p.successCounter.Inc()
	p.timerHistogram.Observe(time.Now().Sub(p.startTime).Seconds())
	p.push(p.successCounter, p.timerHistogram)
}

func (p *prometheusMetrics) Failure() {
	p.failureCounter.Inc()
	p.push(p.failureCounter)
}

func (p *prometheusMetrics) push(metrics ...prometheus.Collector) {
	if p.request.Source.PrometheusGatewayURL != "" {
		pusher := push.New(p.request.Source.PrometheusGatewayURL, p.request.Params.Command)
		pusher.Grouping("cf_api", p.request.Source.API)
		pusher.Grouping("cf_org", p.request.Source.Org)
		for _, m := range metrics {
			pusher.Collector(m)
		}
		pusher.Add()
	}
}
