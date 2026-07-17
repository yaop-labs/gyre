package gyre

import "time"

// MetricsSink is a minimal integration point for products' metrics systems.
// Implementations should keep recording non-blocking and allocation-light.
type MetricsSink interface {
	Observe(name, operation string, duration time.Duration, err error)
}

// RuntimeMetrics wraps a sink and can be used by lifecycle adapters.
type RuntimeMetrics struct{ Sink MetricsSink }

func (m RuntimeMetrics) Observe(name, operation string, started time.Time, err error) {
	if m.Sink != nil {
		m.Sink.Observe(name, operation, time.Since(started), err)
	}
}
