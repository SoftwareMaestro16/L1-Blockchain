package observability

import "time"

func NewRegistry() *Registry {
	r := &Registry{
		enabled:	true,
		startedAt:	time.Now(),
		counters:	make(map[metricKey]sample),
		gauges:		make(map[metricKey]sample),
		summaries:	make(map[metricKey]observation),
	}
	r.SetGauge(MetricLocalnetHealth, Labels{"component": "metrics"}, 1)
	return r
}

func ResetForTesting() {
	DefaultRegistry = NewRegistry()
}

func SetEnabled(enabled bool) {
	DefaultRegistry.SetEnabled(enabled)
}

func (r *Registry) SetEnabled(enabled bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.enabled = enabled
}

func (r *Registry) Enabled() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.enabled
}

func IncCounter(name string, labels Labels, delta float64) {
	DefaultRegistry.IncCounter(name, labels, delta)
}

func SetGauge(name string, labels Labels, value float64) {
	DefaultRegistry.SetGauge(name, labels, value)
}

func AddGauge(name string, labels Labels, delta float64) {
	DefaultRegistry.AddGauge(name, labels, delta)
}

func Observe(name string, labels Labels, value float64) {
	DefaultRegistry.Observe(name, labels, value)
}

func (r *Registry) IncCounter(name string, labels Labels, delta float64) {
	if delta <= 0 || !knownMetric(name) {
		return
	}
	normalized, labelKey := normalizeLabels(labels)
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.enabled {
		return
	}
	key := metricKey{name: name, labelKey: labelKey}
	current := r.counters[key]
	current.labels = normalized
	current.value += delta
	r.counters[key] = current
}

func (r *Registry) SetGauge(name string, labels Labels, value float64) {
	if !knownMetric(name) || invalidNumber(value) {
		return
	}
	normalized, labelKey := normalizeLabels(labels)
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.enabled && name != MetricTelemetryEnabled {
		return
	}
	r.gauges[metricKey{name: name, labelKey: labelKey}] = sample{labels: normalized, value: value}
}

func (r *Registry) AddGauge(name string, labels Labels, delta float64) {
	if !knownMetric(name) || invalidNumber(delta) {
		return
	}
	normalized, labelKey := normalizeLabels(labels)
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.enabled {
		return
	}
	key := metricKey{name: name, labelKey: labelKey}
	current := r.gauges[key]
	current.labels = normalized
	current.value += delta
	r.gauges[key] = current
}

func (r *Registry) Observe(name string, labels Labels, value float64) {
	if value < 0 || !knownMetric(name) || invalidNumber(value) {
		return
	}
	normalized, labelKey := normalizeLabels(labels)
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.enabled {
		return
	}
	key := metricKey{name: name, labelKey: labelKey}
	current := r.summaries[key]
	current.labels = normalized
	current.count++
	current.sum += value
	r.summaries[key] = current
}
