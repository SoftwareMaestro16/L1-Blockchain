package types

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateWorkflow(t *testing.T) {
	workflow := Workflow{
		ID:		"deploy-and-call",
		Authority:	[]byte{1, 2, 3},
		Atomic:		true,
		Steps: []Step{
			{ID: "deploy", Kind: StepKindContractDeployFirstMsg, Payload: []byte("deploy")},
			{ID: "resolve", Kind: StepKindResolverPayment, Payload: []byte("pay")},
		},
	}
	require.NoError(t, ValidateWorkflow(workflow))

	workflow.Authority = make([]byte, 20)
	require.ErrorContains(t, ValidateWorkflow(workflow), "workflow authority")
	workflow.Authority = []byte{1, 2, 3}

	workflow.Steps[1].ID = "deploy"
	require.ErrorContains(t, ValidateWorkflow(workflow), "duplicate workflow step id")
	workflow.Steps[1].ID = "resolve"

	workflow.Steps = append(workflow.Steps, Step{ID: "bad", Kind: "unknown"})
	require.ErrorContains(t, ValidateWorkflow(workflow), "invalid step kind")
}

func TestValidateWorkflowBounds(t *testing.T) {
	steps := make([]Step, MaxWorkflowSteps+1)
	for i := range steps {
		steps[i] = Step{ID: fmt.Sprintf("step-%d", i), Kind: StepKindResolverPayment}
	}
	err := ValidateWorkflow(Workflow{ID: "too-many", Authority: []byte{1}, Steps: steps})
	require.ErrorContains(t, err, "must not exceed")

	err = ValidateStep(Step{
		ID:		"oversized",
		Kind:		StepKindContractDeployFirstMsg,
		Payload:	bytes.Repeat([]byte{1}, MaxWorkflowPayloadBytes+1),
	})
	require.ErrorContains(t, err, "payload")
}
