package cmd

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	l1app "github.com/sovereign-l1/l1/app"
)

func TestInvariantListCommandReturnsCriticalRoutes(t *testing.T) {
	out, err := executeAVMCommand(NewInvariantsCmd(), "list")
	require.NoError(t, err)

	var res struct {
		Command	string		`json:"command"`
		Routes	[]string	`json:"routes"`
	}
	require.NoError(t, json.Unmarshal([]byte(out), &res), out)
	require.Equal(t, "invariants list", res.Command)
	require.ElementsMatch(t, l1app.CriticalAppInvariantRoutes(), res.Routes)
}

func TestInvariantCheckCommandRunsDefaultGenesisRunner(t *testing.T) {
	out, err := executeAVMCommand(NewInvariantsCmd(), "check")
	require.NoError(t, err)

	var report invariantCheckReport
	require.NoError(t, json.Unmarshal([]byte(out), &report), out)
	require.Equal(t, "invariants check", report.Command)
	require.Equal(t, "default-genesis", report.Mode)
	require.True(t, report.Passed, out)
	require.Empty(t, report.Failures)
	require.Contains(t, report.Skipped, "aetra/"+l1app.AppInvariantGenesisExport)
	require.ElementsMatch(t, l1app.CriticalAppInvariantRoutes(), report.Routes)
}
