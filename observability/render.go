package observability

import (
	"fmt"
	"io"
	"sort"
	"strings"
)

func Render() string {
	var b strings.Builder
	_ = DefaultRegistry.WritePrometheus(&b)
	return b.String()
}

func (r *Registry) WritePrometheus(w io.Writer) error {
	r.collectRuntime()

	r.mu.RLock()
	enabled := r.enabled
	counters := cloneSamples(r.counters)
	gauges := cloneSamples(r.gauges)
	summaries := cloneObservations(r.summaries)
	r.mu.RUnlock()

	gauges[metricKey{name: MetricTelemetryEnabled}] = sample{value: boolFloat(enabled)}

	for _, def := range Definitions {
		if _, err := fmt.Fprintf(w, "# HELP %s %s\n# TYPE %s %s\n", def.Name, def.Help, def.Name, def.Type); err != nil {
			return err
		}
		switch def.Type {
		case kindCounter:
			if err := writeSamples(w, def.Name, counters); err != nil {
				return err
			}
		case kindGauge:
			if err := writeSamples(w, def.Name, gauges); err != nil {
				return err
			}
		case kindSummary:
			if err := writeObservations(w, def.Name, summaries); err != nil {
				return err
			}
		}
	}
	return nil
}

func cloneSamples(in map[metricKey]sample) map[metricKey]sample {
	out := make(map[metricKey]sample, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
}

func cloneObservations(in map[metricKey]observation) map[metricKey]observation {
	out := make(map[metricKey]observation, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
}

func writeSamples(w io.Writer, name string, samples map[metricKey]sample) error {
	keys := sortedKeys(name, samples)
	for _, key := range keys {
		s := samples[key]
		if _, err := fmt.Fprintf(w, "%s%s %s\n", name, formatLabels(s.labels), formatFloat(s.value)); err != nil {
			return err
		}
	}
	return nil
}

func writeObservations(w io.Writer, name string, observations map[metricKey]observation) error {
	keys := sortedKeys(name, observations)
	for _, key := range keys {
		o := observations[key]
		labels := formatLabels(o.labels)
		if _, err := fmt.Fprintf(w, "%s_sum%s %s\n%s_count%s %d\n", name, labels, formatFloat(o.sum), name, labels, o.count); err != nil {
			return err
		}
	}
	return nil
}

type keyed interface {
	sample | observation
}

func sortedKeys[T keyed](name string, values map[metricKey]T) []metricKey {
	keys := make([]metricKey, 0, len(values))
	for key := range values {
		if key.name == name {
			keys = append(keys, key)
		}
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i].labelKey < keys[j].labelKey
	})
	return keys
}
