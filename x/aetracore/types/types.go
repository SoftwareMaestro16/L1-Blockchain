package types

import (
	"errors"
	"fmt"
	"strings"

	"github.com/sovereign-l1/l1/app/addressing"
)

const (
	DefaultAuthority		= "4:0000000000000000000000000000000000000000000000000000000000000001"
	DefaultQueryLimit		= uint64(50)
	MaxQueryLimit			= uint64(200)
	DefaultRootHistory		= uint64(4096)
	MaxRootHistory			= uint64(1_000_000)
	DefaultMaxZones			= uint32(32)
	MaxZones			= uint32(1024)
	DefaultMaxShards		= uint32(64)
	MaxShardsPerZone		= uint32(65_536)
	DefaultProposalItems		= uint32(10_000)
	MaxProposalItems		= uint32(250_000)
	NativeFeePolicyID		= "naet"
	DefaultMempoolPolicy		= "core-default"
	DefaultMessagePolicy		= "async-message-v1"
	DefaultGasPolicy		= "deterministic-gas-v1"
	DefaultProofRootType		= "global"
	NextArchitectureRootType	= "aetra_next"
	AccountProofRootType		= "account"
	IdentityProofRootType		= "identity"
	ResolverProofRootType		= "resolver"
	ZoneCommitmentsRoot		= "zone_commitments"
	ShardLayoutRootType		= "shard_layouts"
	RoutingTableRootType		= "routing_table"
	MessageProofRootType		= "messages"
	ReceiptProofRootType		= "receipts"
	PaymentsProofRootType		= "payments"
	VMProofRootType			= "vm"
	StateProofRootType		= "state"
)

type ZoneID string
type ShardID string
type RootType string

const (
	ZoneIDAetraCore		= "AETHER_CORE"
	ZoneIDFinancial		= "FINANCIAL_ZONE"
	ZoneIDIdentity		= "IDENTITY_ZONE"
	ZoneIDPayment		= "PAYMENT_ZONE"
	ZoneIDApplication	= "APPLICATION_ZONE"
	ZoneIDContract		= "CONTRACT_ZONE"
)

type AetraCoreParams struct {
	Enabled				bool
	Authority			string
	DefaultQueryLimit		uint64
	MaxQueryLimit			uint64
	MaxZones			uint32
	MaxShardsPerZone		uint32
	MaxProposalItemsPerBlock	uint32
	RootHistoryWindow		uint64
	CrossZoneFinalityDelay		uint64
	DeterministicProposalGrouping	bool
	ProductionVersionGate		string
}

func DefaultParams() AetraCoreParams {
	return AetraCoreParams{
		Authority:			DefaultAuthority,
		DefaultQueryLimit:		DefaultQueryLimit,
		MaxQueryLimit:			MaxQueryLimit,
		MaxZones:			DefaultMaxZones,
		MaxShardsPerZone:		DefaultMaxShards,
		MaxProposalItemsPerBlock:	DefaultProposalItems,
		RootHistoryWindow:		DefaultRootHistory,
		CrossZoneFinalityDelay:		1,
		DeterministicProposalGrouping:	true,
	}
}

func TestnetParams() AetraCoreParams {
	params := DefaultParams()
	params.Enabled = true
	params.ProductionVersionGate = "testnet-profile"
	return params
}

func (p AetraCoreParams) Validate() error {
	if err := addressing.ValidateAuthorityAddress("aetracore authority", p.Authority); err != nil {
		return err
	}
	if p.DefaultQueryLimit == 0 {
		return errors.New("aetracore default query limit must be positive")
	}
	if p.MaxQueryLimit == 0 || p.MaxQueryLimit > MaxQueryLimit {
		return fmt.Errorf("aetracore max query limit must be between 1 and %d", MaxQueryLimit)
	}
	if p.DefaultQueryLimit > p.MaxQueryLimit {
		return errors.New("aetracore default query limit must not exceed max query limit")
	}
	if p.MaxZones == 0 || p.MaxZones > MaxZones {
		return fmt.Errorf("aetracore max zones must be between 1 and %d", MaxZones)
	}
	if p.MaxShardsPerZone == 0 || p.MaxShardsPerZone > MaxShardsPerZone {
		return fmt.Errorf("aetracore max shards per zone must be between 1 and %d", MaxShardsPerZone)
	}
	if p.MaxProposalItemsPerBlock == 0 || p.MaxProposalItemsPerBlock > MaxProposalItems {
		return fmt.Errorf("aetracore max proposal items per block must be between 1 and %d", MaxProposalItems)
	}
	if p.RootHistoryWindow == 0 || p.RootHistoryWindow > MaxRootHistory {
		return fmt.Errorf("aetracore root history window must be between 1 and %d", MaxRootHistory)
	}
	if p.Enabled && !p.DeterministicProposalGrouping {
		return errors.New("aetracore deterministic proposal grouping must remain enabled")
	}
	if p.Enabled && p.ProductionVersionGate == "" {
		return errors.New("aetracore production enablement requires software version gate")
	}
	return nil
}

func (p AetraCoreParams) RequireEnabled() error {
	if err := p.Validate(); err != nil {
		return err
	}
	if !p.Enabled {
		return errors.New("aetracore feature gate is disabled")
	}
	return nil
}

func (p AetraCoreParams) Authorize(authority string) error {
	if err := p.Validate(); err != nil {
		return err
	}
	if err := addressing.ValidateAuthorityAddress("aetracore update authority", authority); err != nil {
		return err
	}
	if authority != p.Authority {
		return errors.New("aetracore update requires governance authority")
	}
	return nil
}

func ValidateZoneID(id ZoneID) error {
	return validateZoneID(string(id))
}

func ValidateShardID(id ShardID) error {
	text := string(id)
	if strings.TrimSpace(text) != text || text == "" {
		return errors.New("aetracore shard id is required and must not have surrounding whitespace")
	}
	if len(text) > MaxScopeLength {
		return fmt.Errorf("aetracore shard id must be <= %d bytes", MaxScopeLength)
	}
	for _, r := range text {
		if r <= ' ' || r == 0x7f {
			return errors.New("aetracore shard id must not contain whitespace or control characters")
		}
	}
	return nil
}
