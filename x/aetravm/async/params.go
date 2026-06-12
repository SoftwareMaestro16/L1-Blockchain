package async

import (
	"errors"
	"fmt"

	sdkmath "cosmossdk.io/math"
)

func DefaultParams() Params {
	return Params{
		MaxMessagesPerTx:		32,
		MaxMessagesPerBlock:		128,
		MaxQueuedMessagesPerContract:	1024,
		MaxProcessingAttempts:		4,
		MaxRecursionDepth:		8,
		MaxBodySize:			4096,
		MaxStateSize:			64 * 1024,
		MaxContractDeploysPerTx:	4,
		MaxContractDeploysPerBlock:	16,
		MaxEmittedMessagesPerExec:	16,
		MaxStorageWritesPerExec:	64,
		MaxActionsPerExecution:		256,
		MaxRetriesPerMessage:		3,
		DefaultRetryDelayBlocks:	1,
		MaxRetryDelayBlocks:		64,
		MaxDeadLetters:			1024,
		ExecutionGasPerMessage:		10_000,
		StorageFeePerByte:		sdkmath.NewInt(1),
		ForwardingFee:			sdkmath.NewInt(1),
		ContractDeploymentCost:		sdkmath.NewInt(1_000),
	}
}

func (p Params) Validate() error {
	if p.MaxMessagesPerTx == 0 {
		return errors.New("max messages per tx must be positive")
	}
	if p.MaxMessagesPerBlock == 0 {
		return errors.New("max messages per block must be positive")
	}
	if p.MaxQueuedMessagesPerContract == 0 {
		return errors.New("max queued messages per contract must be positive")
	}
	if p.MaxProcessingAttempts == 0 {
		return errors.New("max processing attempts must be positive")
	}
	if p.MaxRecursionDepth == 0 {
		return errors.New("max recursion depth must be positive")
	}
	if p.MaxBodySize == 0 {
		return errors.New("max body size must be positive")
	}
	if p.MaxStateSize == 0 {
		return errors.New("max state size must be positive")
	}
	if p.MaxContractDeploysPerTx == 0 {
		return errors.New("max contract deploys per tx must be positive")
	}
	if p.MaxContractDeploysPerBlock == 0 {
		return errors.New("max contract deploys per block must be positive")
	}
	if p.MaxEmittedMessagesPerExec == 0 {
		return errors.New("max emitted messages per execution must be positive")
	}
	if p.MaxStorageWritesPerExec == 0 {
		return errors.New("max storage writes per execution must be positive")
	}
	if p.MaxActionsPerExecution == 0 {
		return errors.New("max actions per execution must be positive")
	}
	if p.MaxRetriesPerMessage == 0 {
		return errors.New("max retries per message must be positive")
	}
	if p.DefaultRetryDelayBlocks == 0 {
		return errors.New("default retry delay blocks must be positive")
	}
	if p.MaxRetryDelayBlocks == 0 {
		return errors.New("max retry delay blocks must be positive")
	}
	if p.DefaultRetryDelayBlocks > p.MaxRetryDelayBlocks {
		return errors.New("default retry delay blocks must not exceed max retry delay blocks")
	}
	if p.MaxDeadLetters == 0 {
		return errors.New("max dead letters must be positive")
	}
	if p.ExecutionGasPerMessage == 0 {
		return errors.New("execution gas per message must be positive")
	}
	for _, item := range []struct {
		name	string
		value	sdkmath.Int
	}{
		{name: "storage fee per byte", value: p.StorageFeePerByte},
		{name: "forwarding fee", value: p.ForwardingFee},
		{name: "contract deployment cost", value: p.ContractDeploymentCost},
	} {
		if item.value.IsNil() || item.value.IsNegative() {
			return fmt.Errorf("%s must be non-negative", item.name)
		}
	}
	return nil
}
