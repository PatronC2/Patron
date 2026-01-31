package api

import (
	"context"
	"time"

	"github.com/PatronC2/Patron/data"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	agentsTotal = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "patron",
		Subsystem: "api",
		Name:      "agents_total",
		Help:      "Total number of agents",
	})
	agentsOnline = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "patron",
		Subsystem: "api",
		Name:      "agents_online",
		Help:      "Number of online agents",
	})
	agentsOffline = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "patron",
		Subsystem: "api",
		Name:      "agents_offline",
		Help:      "Number of offline agents",
	})

	agentCountsLastSuccess = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "patron",
		Subsystem: "api",
		Name:      "agent_counts_last_success_unixtime",
		Help:      "Unix timestamp of last successful agent count update",
	})

	agentCountsErrors = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "patron",
		Subsystem: "api",
		Name:      "agent_counts_update_errors_total",
		Help:      "Total errors while updating agent counts",
	})
)

func RegisterMetrics() {
	prometheus.MustRegister(
		agentsTotal,
		agentsOnline,
		agentsOffline,
		agentCountsLastSuccess,
		agentCountsErrors,
	)
}

func StartAgentCountUpdater(ctx context.Context, interval time.Duration) (stop func()) {
	if interval <= 0 {
		interval = 10 * time.Second
	}

	updateOnce := func() {
		c, err := data.GetAgentCounts()
		if err != nil {
			agentCountsErrors.Inc()
			return
		}
		agentsTotal.Set(float64(c.Total))
		agentsOnline.Set(float64(c.Online))
		agentsOffline.Set(float64(c.Offline))
		agentCountsLastSuccess.Set(float64(time.Now().Unix()))
	}

	updateOnce()

	ticker := time.NewTicker(interval)
	done := make(chan struct{})

	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				updateOnce()
			case <-ctx.Done():
				close(done)
				return
			}
		}
	}()

	return func() {
		select {
		case <-done:
		default:
		}
	}
}
