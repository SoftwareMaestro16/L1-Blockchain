package types

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAVMExecutionModelComparisonMatchesSection16(t *testing.T) {
	model, err := DefaultAVMExecutionModelComparison()
	require.NoError(t, err)
	require.NoError(t, model.Validate())
	require.Equal(t, "Classic Cosmos SDK vs AVM", model.ModelName)
	require.Equal(t, ComputeAVMExecutionModelComparisonHash(model), model.MatrixHash)
	require.Len(t, model.Rows, 6)

	rows := map[AVMExecutionComparisonFeature]AVMExecutionComparisonRow{}
	for _, row := range model.Rows {
		require.NoError(t, row.Validate())
		require.True(t, row.RequiresAVMExtension)
		rows[row.Feature] = row
	}
	require.Equal(t, "synchronous", rows[AVMComparisonFeatureExecution].ClassicCosmosSDK)
	require.Equal(t, "sync + async", rows[AVMComparisonFeatureExecution].AVM)
	require.Equal(t, "KVStore", rows[AVMComparisonFeatureState].ClassicCosmosSDK)
	require.Equal(t, "KVStore + zone roots", rows[AVMComparisonFeatureState].AVM)
	require.Equal(t, "tx-only", rows[AVMComparisonFeatureMessaging].ClassicCosmosSDK)
	require.Equal(t, "message-driven", rows[AVMComparisonFeatureMessaging].AVM)
	require.Equal(t, "block-bound", rows[AVMComparisonFeatureScheduling].ClassicCosmosSDK)
	require.Equal(t, "cross-block", rows[AVMComparisonFeatureScheduling].AVM)
	require.Equal(t, "module-based", rows[AVMComparisonFeatureContracts].ClassicCosmosSDK)
	require.Equal(t, "actor + module hybrid", rows[AVMComparisonFeatureContracts].AVM)
	require.Equal(t, "CLI/API", rows[AVMComparisonFeatureUXModel].ClassicCosmosSDK)
	require.Equal(t, "interface-driven", rows[AVMComparisonFeatureUXModel].AVM)
}

func TestAVMExecutionModelComparisonRendersValidMarkdownTable(t *testing.T) {
	model, err := DefaultAVMExecutionModelComparison()
	require.NoError(t, err)

	markdown, err := RenderAVMExecutionModelComparisonMarkdown(model)
	require.NoError(t, err)
	lines := strings.Split(markdown, "\n")
	require.Equal(t, "| Feature | Classic Cosmos SDK | AVM |", lines[0])
	require.Equal(t, "| --- | --- | --- |", lines[1])
	require.Contains(t, markdown, "| Execution | synchronous | sync + async |")
	require.Contains(t, markdown, "| UX model | CLI/API | interface-driven |")
	require.NotContains(t, markdown, "| --- | --- | --- | --- |")
	require.Len(t, lines, 8)
}

func TestAVMExecutionModelComparisonRejectsMissingFeatureAndHashMismatch(t *testing.T) {
	model, err := DefaultAVMExecutionModelComparison()
	require.NoError(t, err)

	missing := model
	missing.Rows = missing.Rows[:len(missing.Rows)-1]
	missing.MatrixHash = ComputeAVMExecutionModelComparisonHash(missing)
	require.ErrorContains(t, missing.Validate(), "every section 16 feature")

	mutated := model
	mutated.Rows[0].AVM = "unexpected"
	require.ErrorContains(t, mutated.Validate(), "row hash mismatch")
}

func TestAVMExecutionComparisonRejectsUndifferentiatedRows(t *testing.T) {
	_, err := NewAVMExecutionComparisonRow(AVMExecutionComparisonRow{
		Feature:		AVMComparisonFeatureExecution,
		ClassicCosmosSDK:	"synchronous",
		AVM:			"synchronous",
		RequiresAVMExtension:	false,
	})
	require.ErrorContains(t, err, "distinguish AVM")

	row, err := NewAVMExecutionComparisonRow(AVMExecutionComparisonRow{
		Feature:		AVMComparisonFeatureExecution,
		ClassicCosmosSDK:	"synchronous",
		AVM:			"sync + async",
		RequiresAVMExtension:	true,
	})
	require.NoError(t, err)
	row.AVM = "sync | async"
	row.RowHash = ComputeAVMExecutionComparisonRowHash(row)
	require.ErrorContains(t, row.Validate(), "invalid character")
}
