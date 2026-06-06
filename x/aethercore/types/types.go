package types

import (
	"errors"
	"fmt"
	"strings"

	"github.com/sovereign-l1/l1/app/addressing"
)

const (
	DefaultAuthority     = "4:0000000000000000000000000000000000000000000000000000000000000001"
	DefaultQueryLimit    = uint64(50)
	MaxQueryLimit        = uint64(200)
	DefaultRootHistory   = uint64(4096)
	MaxRootHistory       = uint64(1_000_000)
	DefaultMaxZones      = uint32(32)
	MaxZones             = uint32(1024)
	DefaultMaxShards     = uint32(64)
	MaxShardsPerZone     = uint32(65_536)
	DefaultProposalItems = uint32(10_000)
	MaxProposalItems     = uint32(250_000)
	NativeFeePolicyID    = "naet"
	DefaultMempoolPolicy = "core-default"
	DefaultMessagePolicy = "async-message-v1"
	DefaultGasPolicy     = "deterministic-gas-v1"
	DefaultProofRootType = "global"
	ZoneCommitmentsRoot  = "zone_commitments"
	ShardLayoutRootType  = "shard_layouts"
	RoutingTableRootType = "routing_table"
	MessageProofRootType = "messages"
	ReceiptProofRootType = "receipts"
	StateProofRootType   = "state"
)

type ZoneID string
type ShardID string
type RootType string

const (
	ZoneIDAetherCore  = "AETHER_CORE"
	ZoneIDFinancial   = "FINANCIAL_ZONE"
	ZoneIDIdentity    = "IDENTITY_ZONE"
	ZoneIDPayment     = "PAYMENT_ZONE"
	ZoneIDApplication = "APPLICATION_ZONE"
	ZoneIDContract    = "CONTRACT_ZONE"
)

type AetherCoreParams struct {
	Enabled                       bool
	Authority                     string
	DefaultQueryLimit             uint64
	MaxQueryLimit                 uint64
	MaxZones                      uint32
	MaxShardsPerZone              uint32
	MaxProposalItemsPerBlock      uint32
	RootHistoryWindow             uint64
	CrossZoneFinalityDelay        uint64
	DeterministicProposalGrouping bool
	ProductionVersionGate         string
}

func DefaultParams() AetherCoreParams {
	return AetherCoreParams{
		Authority:                     DefaultAuthority,
		DefaultQueryLimit:             DefaultQueryLimit,
		MaxQueryLimit:                 MaxQueryLimit,
		MaxZones:                      DefaultMaxZones,
		MaxShardsPerZone:              DefaultMaxShards,
		MaxProposalItemsPerBlock:      DefaultProposalItems,
		RootHistoryWindow:             DefaultRootHistory,
		CrossZoneFinalityDelay:        1,
		DeterministicProposalGrouping: true,
	}
}

func TestnetParams() AetherCoreParams {
	params := DefaultParams()
	params.Enabled = true
	params.ProductionVersionGate = "testnet-profile"
	return params
}

func (p AetherCoreParams) Validate() error {
	if err := addressing.ValidateAuthorityAddress("aethercore authority", p.Authority); err != nil {
		return err
	}
	if p.DefaultQueryLimit == 0 {
		return errors.New("aethercore default query limit must be positive")
	}
	if p.MaxQueryLimit == 0 || p.MaxQueryLimit > MaxQueryLimit {
		return fmt.Errorf("aethercore max query limit must be between 1 and %d", MaxQueryLimit)
	}
	if p.DefaultQueryLimit > p.MaxQueryLimit {
		return errors.New("aethercore default query limit must not exceed max query limit")
	}
	if p.MaxZones == 0 || p.MaxZones > MaxZones {
		return fmt.Errorf("aethercore max zones must be between 1 and %d", MaxZones)
	}
	if p.MaxShardsPerZone == 0 || p.MaxShardsPerZone > MaxShardsPerZone {
		return fmt.Errorf("aethercore max shards per zone must be between 1 and %d", MaxShardsPerZone)
	}
	if p.MaxProposalItemsPerBlock == 0 || p.MaxProposalItemsPerBlock > MaxProposalItems {
		return fmt.Errorf("aethercore max proposal items per block must be between 1 and %d", MaxProposalItems)
	}
	if p.RootHistoryWindow == 0 || p.RootHistoryWindow > MaxRootHistory {
		return fmt.Errorf("aethercore root history window must be between 1 and %d", MaxRootHistory)
	}
	if p.Enabled && !p.DeterministicProposalGrouping {
		return errors.New("aethercore deterministic proposal grouping must remain enabled")
	}
	if p.Enabled && p.ProductionVersionGate == "" {
		return errors.New("aethercore production enablement requires software version gate")
	}
	return nil
}

func (p AetherCoreParams) RequireEnabled() error {
	if err := p.Validate(); err != nil {
		return err
	}
	if !p.Enabled {
		return errors.New("aethercore feature gate is disabled")
	}
	return nil
}

func (p AetherCoreParams) Authorize(authority string) error {
	if err := p.Validate(); err != nil {
		return err
	}
	if err := addressing.ValidateAuthorityAddress("aethercore update authority", authority); err != nil {
		return err
	}
	if authority != p.Authority {
		return errors.New("aethercore update requires governance authority")
	}
	return nil
}

func ValidateZoneID(id ZoneID) error {
	return validateZoneID(string(id))
}

func ValidateShardID(id ShardID) error {
	text := string(id)
	if strings.TrimSpace(text) != text || text == "" {
		return errors.New("aethercore shard id is required and must not have surrounding whitespace")
	}
	if len(text) > MaxScopeLength {
		return fmt.Errorf("aethercore shard id must be <= %d bytes", MaxScopeLength)
	}
	for _, r := range text {
		if r <= ' ' || r == 0x7f {
			return errors.New("aethercore shard id must not contain whitespace or control characters")
		}
	}
	return nil
}
