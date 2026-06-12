package types

import (
	"errors"
	"fmt"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sovereign-l1/l1/app/addressing"
)

const (
	OpNoop		= "noop"
	OpStorageRead	= "storage_read"
	OpStorageWrite	= "storage_write"
	OpCryptoVerify	= "crypto_verify"
	OpContractCall	= "contract_call"
	OpHeavyCompute	= "heavy_compute"
	DefaultBlockCap	= uint64(10_000_000)
)

type Params struct {
	DefaultOpCost		uint64
	MaxBlockComputeUnits	uint64
	MaxContractComputeUnits	uint64
	OpCosts			map[string]uint64
}

type Operation struct {
	Contract	sdk.AccAddress
	Op		string
	Count		uint64
}

type ContractStats struct {
	Contract	sdk.AccAddress
	Used		uint64
	Ops		uint64
}

type BlockMeter struct {
	params	Params
	used	uint64
	stats	map[string]ContractStats
}

func DefaultParams() Params {
	return Params{
		DefaultOpCost:			1,
		MaxBlockComputeUnits:		DefaultBlockCap,
		MaxContractComputeUnits:	DefaultBlockCap / 10,
		OpCosts: map[string]uint64{
			OpNoop:		1,
			OpStorageRead:	5,
			OpStorageWrite:	20,
			OpCryptoVerify:	500,
			OpContractCall:	100,
			OpHeavyCompute:	10_000,
		},
	}
}

func NewBlockMeter(params Params) (*BlockMeter, error) {
	if err := params.Validate(); err != nil {
		return nil, err
	}
	return &BlockMeter{params: params.Clone(), stats: make(map[string]ContractStats)}, nil
}

func (p Params) Clone() Params {
	out := p
	out.OpCosts = make(map[string]uint64, len(p.OpCosts))
	for op, cost := range p.OpCosts {
		out.OpCosts[op] = cost
	}
	return out
}

func (p Params) Validate() error {
	if p.DefaultOpCost == 0 {
		return errors.New("default compute op cost must be positive")
	}
	if p.MaxBlockComputeUnits == 0 {
		return errors.New("max block compute units must be positive")
	}
	if p.MaxContractComputeUnits == 0 {
		return errors.New("max contract compute units must be positive")
	}
	if p.MaxContractComputeUnits > p.MaxBlockComputeUnits {
		return errors.New("max contract compute units cannot exceed block cap")
	}
	for op, cost := range p.OpCosts {
		if op == "" {
			return errors.New("compute op name is required")
		}
		if cost == 0 {
			return fmt.Errorf("compute op %q cost must be positive", op)
		}
	}
	return nil
}

func (p Params) CostFor(op string) uint64 {
	if cost, ok := p.OpCosts[op]; ok {
		return cost
	}
	return p.DefaultOpCost
}

func ComputeCost(params Params, ops []Operation) (uint64, error) {
	if err := params.Validate(); err != nil {
		return 0, err
	}
	var total uint64
	for _, op := range ops {
		cost, err := operationCost(params, op)
		if err != nil {
			return 0, err
		}
		total += cost
	}
	return total, nil
}

func (m *BlockMeter) Charge(op Operation) error {
	cost, err := operationCost(m.params, op)
	if err != nil {
		return err
	}
	if m.used+cost > m.params.MaxBlockComputeUnits {
		return fmt.Errorf("block compute cap exceeded: %d > %d", m.used+cost, m.params.MaxBlockComputeUnits)
	}
	key := string(op.Contract)
	stats := m.stats[key]
	stats.Contract = append(sdk.AccAddress(nil), op.Contract...)
	if stats.Used+cost > m.params.MaxContractComputeUnits {
		return fmt.Errorf("contract compute cap exceeded: %d > %d", stats.Used+cost, m.params.MaxContractComputeUnits)
	}
	stats.Used += cost
	stats.Ops += op.Count
	m.stats[key] = stats
	m.used += cost
	return nil
}

func (m *BlockMeter) Used() uint64 {
	return m.used
}

func (m *BlockMeter) Stats() []ContractStats {
	out := make([]ContractStats, 0, len(m.stats))
	for _, stats := range m.stats {
		stats.Contract = append(sdk.AccAddress(nil), stats.Contract...)
		out = append(out, stats)
	}
	sort.SliceStable(out, func(i, j int) bool {
		return string(out[i].Contract) < string(out[j].Contract)
	})
	return out
}

func operationCost(params Params, op Operation) (uint64, error) {
	if len(op.Contract) == 0 {
		return 0, errors.New("compute contract address is required")
	}
	if err := addressing.RejectZeroAddress("compute contract", op.Contract); err != nil {
		return 0, err
	}
	if op.Op == "" {
		return 0, errors.New("compute op is required")
	}
	if op.Count == 0 {
		return 0, errors.New("compute op count must be positive")
	}
	return params.CostFor(op.Op) * op.Count, nil
}
