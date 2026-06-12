package async

import (
	"errors"
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	aetraaddress "github.com/sovereign-l1/l1/app/addressing"
)

func NewExecutor(params Params) (*Executor, error) {
	if err := params.Validate(); err != nil {
		return nil, err
	}
	return &Executor{
		params:		params,
		contracts:	make(map[string]ContractAccount),
		inbox:		make(map[string][]QueuedMessage),
		outbox:		make(map[string][]QueuedMessage),
		handlers:	make(map[string]Handler),
	}, nil
}

func (e *Executor) RegisterHandler(address sdk.AccAddress, handler Handler) error {
	if err := aetraaddress.RejectZeroAddress("contract handler", address); err != nil {
		return err
	}
	if _, ok := e.contracts[string(address)]; !ok {
		return errors.New("cannot register handler for missing contract")
	}
	e.handlers[string(address)] = handler
	return nil
}

func (e *Executor) DeployContract(deployer sdk.AccAddress, codeHash []byte, salt []byte, state []byte, balance sdkmath.Int) (sdk.AccAddress, error) {
	return e.DeployContracts(deployer, []DeploySpec{{CodeHash: codeHash, Salt: salt, State: state, BalanceNaet: balance}})
}

func (e *Executor) DeployContracts(deployer sdk.AccAddress, specs []DeploySpec) (sdk.AccAddress, error) {
	if len(specs) == 0 {
		return nil, errors.New("deploy count must be positive")
	}
	if len(specs) > int(e.params.MaxContractDeploysPerTx) {
		return nil, fmt.Errorf("contract deploys per tx must be <= %d", e.params.MaxContractDeploysPerTx)
	}
	if e.deploysInBlock+uint32(len(specs)) > e.params.MaxContractDeploysPerBlock {
		return nil, fmt.Errorf("contract deploys per block must be <= %d", e.params.MaxContractDeploysPerBlock)
	}
	var first sdk.AccAddress
	for _, spec := range specs {
		if spec.BalanceNaet.IsNil() || spec.BalanceNaet.LT(e.params.ContractDeploymentCost) {
			return nil, errors.New("contract deployment requires naet deployment cost")
		}
		address, err := DeriveContractAddress(deployer, spec.CodeHash, spec.Salt)
		if err != nil {
			return nil, err
		}
		if _, exists := e.contracts[string(address)]; exists {
			return nil, errors.New("contract already deployed")
		}
		contract := ContractAccount{
			Address:			address,
			CodeHash:			append([]byte(nil), spec.CodeHash...),
			State:				append([]byte(nil), spec.State...),
			BalanceNaet:			spec.BalanceNaet.Sub(e.params.ContractDeploymentCost),
			Status:				ContractStatusActive,
			StorageRentDebtNaet:		sdkmath.ZeroInt(),
			LastStorageChargeHeight:	e.blockHeight,
		}
		if err := contract.Validate(e.params); err != nil {
			return nil, err
		}
		e.contracts[string(address)] = contract
		e.metrics.DeploymentCostsNaet = addNaetMetric(e.metrics.DeploymentCostsNaet, e.params.ContractDeploymentCost)
		if first == nil {
			first = address
		}
	}
	e.deploysInBlock += uint32(len(specs))
	return first, nil
}

func (e *Executor) Contract(address sdk.AccAddress) (ContractAccount, bool) {
	contract, ok := e.contracts[string(address)]
	return cloneContract(contract), ok
}

func (e *Executor) Queue() []QueuedMessage {
	return cloneQueuedMessages(e.queue)
}

func (e *Executor) Inbox(address sdk.AccAddress) []QueuedMessage {
	return cloneQueuedMessages(e.inbox[inboxKey(address)])
}

func (e *Executor) Outbox(address sdk.AccAddress) []QueuedMessage {
	return cloneQueuedMessages(e.outbox[outboxKey(address)])
}

func (e *Executor) DeadLetters() []DeadLetter {
	return cloneDeadLetters(e.deadLetters)
}

func (e *Executor) Receipts() []ExecutionReceipt {
	return cloneReceipts(e.receipts)
}

func (e *Executor) Metrics() Observability {
	return e.metrics
}
