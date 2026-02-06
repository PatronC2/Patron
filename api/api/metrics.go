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

	redirectorsTotal = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "patron",
		Subsystem: "redirectors",
		Name:      "redirectors_total",
		Help:      "Total number of redirectors",
	})
	redirectorsOnline = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "patron",
		Subsystem: "redirectors",
		Name:      "redirectors_online",
		Help:      "Number of online redirectors",
	})
	redirectorsOffline = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "patron",
		Subsystem: "redirectors",
		Name:      "redirectors_offline",
		Help:      "Number of offline redirectors",
	})

	redirectorCountsLastSuccess = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "patron",
		Subsystem: "redirectors",
		Name:      "redirector_counts_last_success_unixtime",
		Help:      "Unix timestamp of last successful redirector count update",
	})

	redirectorCountsErrors = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "patron",
		Subsystem: "redirectors",
		Name:      "redirector_counts_update_errors_total",
		Help:      "Total errors while updating redirector counts",
	})
)

func RegisterMetrics() {
	prometheus.MustRegister(
		agentsTotal,
		agentsOnline,
		agentsOffline,
		agentCountsLastSuccess,
		agentCountsErrors,
		redirectorsTotal,
		redirectorsOnline,
		redirectorsOffline,
		redirectorCountsLastSuccess,
		redirectorCountsErrors,
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

func StartRedirectorCountUpdater(ctx context.Context, interval time.Duration) (stop func()) {
	if interval <= 0 {
		interval = 10 * time.Second
	}

	updateOnce := func() {
		c, err := data.GetRedirectorCounts()
		if err != nil {
			redirectorCountsErrors.Inc()
			return
		}
		redirectorsTotal.Set(float64(c.Total))
		redirectorsOnline.Set(float64(c.Online))
		redirectorsOffline.Set(float64(c.Offline))
		redirectorCountsLastSuccess.Set(float64(time.Now().Unix()))
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
