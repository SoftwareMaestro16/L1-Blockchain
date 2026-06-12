package types

import (
	"errors"
	"fmt"
	"strings"
)

const (
	FeePolicyNaet	= "naet"

	MaxZoneIDLength		= 64
	MaxPolicyValueLength	= 64
	HashHexLength		= 64
)

type ZoneID string

const (
	ZoneIDFinancial		ZoneID	= "FINANCIAL_ZONE"
	ZoneIDIdentity		ZoneID	= "IDENTITY_ZONE"
	ZoneIDApplication	ZoneID	= "APPLICATION_ZONE"
	ZoneIDContract		ZoneID	= "CONTRACT_ZONE"
)

type ZoneKind string

const (
	ZoneKindFinancial	ZoneKind	= "FINANCIAL"
	ZoneKindIdentity	ZoneKind	= "IDENTITY"
	ZoneKindApplication	ZoneKind	= "APPLICATION"
	ZoneKindContract	ZoneKind	= "CONTRACT"
)

type VMPolicy string

const (
	VMPolicyAVM		VMPolicy	= "AVM"
	VMPolicyCosmWasmGated	VMPolicy	= "COSMWASM_GATED"
	VMPolicyNativeModule	VMPolicy	= "NATIVE_MODULE"
)

type UpgradePolicy string

const (
	UpgradePolicyGovernance	UpgradePolicy	= "GOVERNANCE"
	UpgradePolicyScheduled	UpgradePolicy	= "SCHEDULED"
	UpgradePolicyImmutable	UpgradePolicy	= "IMMUTABLE"
)

type DataAvailabilityPolicy string

const (
	DataAvailabilityCoreCommitment	DataAvailabilityPolicy	= "CORE_COMMITMENT"
	DataAvailabilityReplicated	DataAvailabilityPolicy	= "REPLICATED"
)

type AuditStatus string

const (
	AuditStatusExperimental		AuditStatus	= "EXPERIMENTAL"
	AuditStatusInternalReview	AuditStatus	= "INTERNAL_REVIEW"
	AuditStatusAudited		AuditStatus	= "AUDITED"
)

type Zone struct {
	ID			ZoneID
	Kind			ZoneKind
	VMPolicy		VMPolicy
	FeePolicy		string
	GenesisStateHash	string
	StateTransitionID	string
	UpgradePolicy		UpgradePolicy
	DataAvailabilityPolicy	DataAvailabilityPolicy
	AuditStatus		AuditStatus
	ActivationHeight	uint64
}

func (z Zone) Validate() error {
	if err := ValidateZoneID(z.ID); err != nil {
		return err
	}
	if !IsZoneKind(z.Kind) {
		return fmt.Errorf("unknown zone kind %q", z.Kind)
	}
	if !IsVMPolicy(z.VMPolicy) {
		return fmt.Errorf("unknown zone VM policy %q", z.VMPolicy)
	}
	if strings.TrimSpace(z.FeePolicy) != FeePolicyNaet {
		return fmt.Errorf("zone fee policy must use %s", FeePolicyNaet)
	}
	if err := ValidateHash("zone genesis state hash", z.GenesisStateHash); err != nil {
		return err
	}
	if err := validatePolicyID("zone state transition id", z.StateTransitionID); err != nil {
		return err
	}
	if !IsUpgradePolicy(z.UpgradePolicy) {
		return fmt.Errorf("unknown zone upgrade policy %q", z.UpgradePolicy)
	}
	if !IsDataAvailabilityPolicy(z.DataAvailabilityPolicy) {
		return fmt.Errorf("unknown zone data availability policy %q", z.DataAvailabilityPolicy)
	}
	if !IsAuditStatus(z.AuditStatus) {
		return fmt.Errorf("unknown zone audit status %q", z.AuditStatus)
	}
	return nil
}

func ValidateZoneID(id ZoneID) error {
	text := string(id)
	if strings.TrimSpace(text) != text || text == "" {
		return errors.New("zone id is required and must not have surrounding whitespace")
	}
	if len(text) > MaxZoneIDLength {
		return fmt.Errorf("zone id must be <= %d bytes", MaxZoneIDLength)
	}
	for i, r := range text {
		if i == 0 && (r < 'A' || r > 'Z') {
			return errors.New("zone id must start with A-Z")
		}
		if r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '_' {
			continue
		}
		return errors.New("zone id must contain only A-Z, 0-9, or underscore")
	}
	return nil
}

func ValidateHash(fieldName, value string) error {
	if len(value) != HashHexLength {
		return fmt.Errorf("%s must be %d lowercase hex chars", fieldName, HashHexLength)
	}
	for _, r := range value {
		if r >= '0' && r <= '9' || r >= 'a' && r <= 'f' {
			continue
		}
		return fmt.Errorf("%s must be %d lowercase hex chars", fieldName, HashHexLength)
	}
	return nil
}

func IsZoneKind(kind ZoneKind) bool {
	switch kind {
	case ZoneKindFinancial, ZoneKindIdentity, ZoneKindApplication, ZoneKindContract:
		return true
	default:
		return false
	}
}

func IsVMPolicy(policy VMPolicy) bool {
	switch policy {
	case VMPolicyAVM, VMPolicyCosmWasmGated, VMPolicyNativeModule:
		return true
	default:
		return false
	}
}

func IsUpgradePolicy(policy UpgradePolicy) bool {
	switch policy {
	case UpgradePolicyGovernance, UpgradePolicyScheduled, UpgradePolicyImmutable:
		return true
	default:
		return false
	}
}

func IsDataAvailabilityPolicy(policy DataAvailabilityPolicy) bool {
	switch policy {
	case DataAvailabilityCoreCommitment, DataAvailabilityReplicated:
		return true
	default:
		return false
	}
}

func IsAuditStatus(status AuditStatus) bool {
	switch status {
	case AuditStatusExperimental, AuditStatusInternalReview, AuditStatusAudited:
		return true
	default:
		return false
	}
}

func validatePolicyID(fieldName, value string) error {
	if strings.TrimSpace(value) != value || value == "" {
		return fmt.Errorf("%s is required and must not have surrounding whitespace", fieldName)
	}
	if len(value) > MaxPolicyValueLength {
		return fmt.Errorf("%s must be <= %d bytes", fieldName, MaxPolicyValueLength)
	}
	for _, r := range value {
		if r >= 'A' && r <= 'Z' || r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '_' || r == '-' || r == '.' || r == ':' {
			continue
		}
		return fmt.Errorf("%s contains invalid character", fieldName)
	}
	return nil
}
