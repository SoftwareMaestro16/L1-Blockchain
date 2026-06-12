package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/sovereign-l1/l1/observability"
)

const (
	flagObservabilityMetrics	= "observability-metrics"
	flagObservabilityMetricsAddr	= "observability-metrics-addr"
)

func startObservabilityMetrics(cmd *cobra.Command) error {
	if cmd.Name() != "start" {
		return nil
	}
	enabled, err := cmd.Flags().GetBool(flagObservabilityMetrics)
	if err != nil {
		return err
	}
	if !enabled {
		observability.SetEnabled(false)
		return nil
	}
	addr, err := cmd.Flags().GetString(flagObservabilityMetricsAddr)
	if err != nil {
		return err
	}
	if err := observability.StartMetricsServer(cmd.Context(), addr, observability.DefaultRegistry); err != nil {
		return fmt.Errorf("failed to start observability metrics endpoint: %w", err)
	}
	observability.SetEnabled(true)
	return nil
}
