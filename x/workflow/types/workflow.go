package types

import (
	"errors"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sovereign-l1/l1/app/addressing"
)

const (
	MaxWorkflowSteps	= 16
	MaxWorkflowPayloadBytes	= 4096
)

type StepKind string

const (
	StepKindResolverPayment			StepKind	= "resolver_payment"
	StepKindDomainAuctionFinalization	StepKind	= "domain_auction_finalization"
	StepKindContractDeployFirstMsg		StepKind	= "contract_deploy_first_message"
)

type Workflow struct {
	ID		string
	Authority	sdk.AccAddress
	Atomic		bool
	Steps		[]Step
}

type Step struct {
	ID	string
	Kind	StepKind
	Payload	[]byte
}

func ValidateWorkflow(workflow Workflow) error {
	if strings.TrimSpace(workflow.ID) == "" {
		return errors.New("workflow id is required")
	}
	if len(workflow.Authority) == 0 {
		return errors.New("workflow authority is required")
	}
	if err := addressing.RejectZeroAddress("workflow authority", workflow.Authority); err != nil {
		return err
	}
	if len(workflow.Steps) == 0 {
		return errors.New("workflow must have at least one step")
	}
	if len(workflow.Steps) > MaxWorkflowSteps {
		return fmt.Errorf("workflow must not exceed %d steps", MaxWorkflowSteps)
	}
	seen := make(map[string]struct{}, len(workflow.Steps))
	for i, step := range workflow.Steps {
		if err := ValidateStep(step); err != nil {
			return fmt.Errorf("invalid workflow step %d: %w", i, err)
		}
		if _, ok := seen[step.ID]; ok {
			return fmt.Errorf("duplicate workflow step id %q", step.ID)
		}
		seen[step.ID] = struct{}{}
	}
	return nil
}

func ValidateStep(step Step) error {
	if strings.TrimSpace(step.ID) == "" {
		return errors.New("step id is required")
	}
	if !IsStepKind(step.Kind) {
		return fmt.Errorf("invalid step kind %q", step.Kind)
	}
	if len(step.Payload) > MaxWorkflowPayloadBytes {
		return fmt.Errorf("step payload must not exceed %d bytes", MaxWorkflowPayloadBytes)
	}
	return nil
}

func IsStepKind(kind StepKind) bool {
	switch kind {
	case StepKindResolverPayment,
		StepKindDomainAuctionFinalization,
		StepKindContractDeployFirstMsg:
		return true
	default:
		return false
	}
}
