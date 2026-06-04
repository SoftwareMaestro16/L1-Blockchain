package observability

import (
	"bufio"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestMetricNamesSnapshot(t *testing.T) {
	reg := NewRegistry()
	reg.RecordTestSamples()

	out := renderRegistry(t, reg)
	expected := []string{
		MetricTelemetryEnabled,
		MetricBlockHeight,
		MetricBlockTimeSeconds,
		MetricBlockProcessing,
		MetricTxLatency,
		MetricModuleErrors,
		MetricDexPoolCount,
		MetricDexLiquidityNorb,
		MetricDexSwaps,
		MetricFeesAccepted,
		MetricFeesRejected,
		MetricLocalnetHealth,
		MetricProcessUptimeSeconds,
		MetricProcessMemoryBytes,
		MetricProcessGoroutines,
	}
	for _, name := range expected {
		require.Contains(t, out, "# HELP "+name+" ")
		require.Contains(t, out, "# TYPE "+name+" ")
	}
}

func TestDisabledTelemetryNoPanic(t *testing.T) {
	reg := NewRegistry()
	reg.SetEnabled(false)

	require.NotPanics(t, func() {
		reg.IncCounter(MetricFeesAccepted, nil, 1)
		reg.SetGauge(MetricBlockHeight, nil, 7)
		reg.Observe(MetricTxLatency, nil, time.Millisecond.Seconds())
		_ = renderRegistry(t, reg)
	})

	out := renderRegistry(t, reg)
	require.Contains(t, out, MetricTelemetryEnabled+" 0")
	require.NotContains(t, out, MetricBlockHeight+" 7")
}

func TestLabelsAreBoundedAndRedacted(t *testing.T) {
	reg := NewRegistry()
	reg.IncCounter(MetricModuleErrors, Labels{
		"module":  "dex",
		"action":  "swap",
		"address": "orb1notallowed",
		"reason":  "contains@secret",
	}, 1)

	out := renderRegistry(t, reg)
	require.Contains(t, out, `module="dex"`)
	require.Contains(t, out, `action="swap"`)
	require.Contains(t, out, `reason="redacted"`)
	require.NotContains(t, out, "orb1notallowed")
	require.NotContains(t, out, "secret")
}

func TestMetricsHandler(t *testing.T) {
	reg := NewRegistry()
	reg.SetGauge(MetricBlockHeight, nil, 10)
	server := httptest.NewServer(Handler(reg))
	defer server.Close()

	res, err := http.Get(server.URL + MetricsPath)
	require.NoError(t, err)
	defer res.Body.Close()
	require.Equal(t, http.StatusOK, res.StatusCode)

	scanner := bufio.NewScanner(res.Body)
	found := false
	for scanner.Scan() {
		if scanner.Text() == MetricBlockHeight+" 10" {
			found = true
			break
		}
	}
	require.NoError(t, scanner.Err())
	require.True(t, found)
}

func (r *Registry) RecordTestSamples() {
	r.SetGauge(MetricBlockHeight, nil, 1)
	r.SetGauge(MetricBlockTimeSeconds, nil, 1700000000)
	r.Observe(MetricBlockProcessing, Labels{"result": "finalized"}, 0.01)
	r.Observe(MetricTxLatency, Labels{"result": "finalized"}, 0.001)
	r.IncCounter(MetricModuleErrors, Labels{"module": "dex", "action": "swap", "reason": "invalid"}, 1)
	r.AddGauge(MetricDexPoolCount, nil, 1)
	r.AddGauge(MetricDexLiquidityNorb, Labels{"denom": "norb"}, 100)
	r.IncCounter(MetricDexSwaps, Labels{"result": "success"}, 1)
	r.IncCounter(MetricFeesAccepted, Labels{"result": "accepted"}, 1)
	r.IncCounter(MetricFeesRejected, Labels{"reason": "invalid_fee"}, 1)
}

func renderRegistry(t *testing.T, reg *Registry) string {
	t.Helper()
	var out strings.Builder
	require.NoError(t, reg.WritePrometheus(&out))
	return out.String()
}
