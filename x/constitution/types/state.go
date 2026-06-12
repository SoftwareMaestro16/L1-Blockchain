package types

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/sovereign-l1/l1/app/addressing"
	configtypes "github.com/sovereign-l1/l1/x/config/types"
	"github.com/sovereign-l1/l1/x/internal/prototype"
)

const (
	MaxPendingAmendmentsV1	= uint32(64)
	MaxAmendmentIDBytesV1	= uint32(96)
	MaxReasonBytesV1	= uint32(512)

	AmendmentStatusPending		= "pending"
	AmendmentStatusApproved		= "approved"
	AmendmentStatusExecuted		= "executed"
	AmendmentStatusCancelled	= "cancelled"

	VoteSupportYes	= "yes"
	VoteSupportNo	= "no"

	MaxBasisPoints	= uint32(10_000)
)

type Params struct {
	Authority		string
	MaxPendingAmendments	uint32
	MaxAmendmentIDBytes	uint32
	MinAmendmentDelay	uint64
	MinQuorumBps		uint32
	EmergencyPauseMaxBlocks	uint64
}

type Constitution struct {
	MaxInflationBps			uint32
	MinSlashFractionBps		uint32
	MaxSlashFractionBps		uint32
	MaxValidatorVotingPowerBps	uint32
	MaxBlockGas			uint64
	MaxAVMCodeSizeBytes		uint64
	MaxContractStateSizeBytes	uint64
	MinStorageRentRate		uint64
	TreasurySpendLimitPerEpoch	uint64
	UpgradeDelayBlocks		uint64
	EmergencyPauseMaxBlocks		uint64
	GovernanceQuorumFloorBps	uint32
	ProtectedModules		[]string
	ConstitutionalExceptionKeys	[]string
	EmergencyPauseUntilHeight	uint64
}

type Amendment struct {
	ID			string
	Status			string
	Proposer		string
	Approver		string
	Executor		string
	Canceller		string
	Reason			string
	Proposed		Constitution
	CreatedHeight		uint64
	UpdatedHeight		uint64
	ExecutableHeight	uint64
	YesVotingPowerBps	uint32
	NoVotingPowerBps	uint32
}

type State struct {
	Constitution		Constitution
	PendingAmendments	[]Amendment
}

type ProtectedLimits struct {
	MaxBlockGas			uint64
	MaxAVMCodeSizeBytes		uint64
	MaxContractStateSizeBytes	uint64
	MinStorageRentRate		uint64
	ProtectedModules		[]string
}

type MsgProposeConstitutionAmendment struct {
	Authority	string
	Amendment	Amendment
}

type MsgVoteConstitutionAmendment struct {
	Authority	string
	AmendmentID	string
	Support		string
	VotingPowerBps	uint32
}

type MsgExecuteConstitutionAmendment struct {
	Authority	string
	AmendmentID	string
}

type MsgCancelConstitutionAmendment struct {
	Authority	string
	AmendmentID	string
	Reason		string
}

func DefaultParams() Params {
	return Params{
		Authority:			prototype.DefaultAuthority,
		MaxPendingAmendments:		MaxPendingAmendmentsV1,
		MaxAmendmentIDBytes:		MaxAmendmentIDBytesV1,
		MinAmendmentDelay:		100,
		MinQuorumBps:			6_700,
		EmergencyPauseMaxBlocks:	1_000,
	}
}

func DefaultConstitution() Constitution {
	return Constitution{
		MaxInflationBps:		2_000,
		MinSlashFractionBps:		1,
		MaxSlashFractionBps:		2_000,
		MaxValidatorVotingPowerBps:	2_500,
		MaxBlockGas:			1_000_000_000,
		MaxAVMCodeSizeBytes:		10 * 1024 * 1024,
		MaxContractStateSizeBytes:	1_000 * 1024 * 1024,
		MinStorageRentRate:		1,
		TreasurySpendLimitPerEpoch:	1_000_000_000_000,
		UpgradeDelayBlocks:		10_000,
		EmergencyPauseMaxBlocks:	1_000,
		GovernanceQuorumFloorBps:	6_700,
		ProtectedModules: []string{
			"config",
			"constitution",
			"evidence",
			"fee-collector",
			"mint-authority",
			"slashing",
			"staking",
			"treasury",
			"validator-election",
			"validator-registry",
		},
		ConstitutionalExceptionKeys:	[]string{},
	}
}

func (p Params) Validate() error {
	if err := addressing.ValidateAuthorityAddress("constitution authority", p.Authority); err != nil {
		return err
	}
	if p.MaxPendingAmendments == 0 || p.MaxPendingAmendments > MaxPendingAmendmentsV1 {
		return fmt.Errorf("constitution max pending amendments must be between 1 and %d", MaxPendingAmendmentsV1)
	}
	if p.MaxAmendmentIDBytes == 0 || p.MaxAmendmentIDBytes > MaxAmendmentIDBytesV1 {
		return fmt.Errorf("constitution max amendment id bytes must be between 1 and %d", MaxAmendmentIDBytesV1)
	}
	if p.MinAmendmentDelay == 0 {
		return errors.New("constitution amendment delay must be positive")
	}
	if p.MinQuorumBps == 0 || p.MinQuorumBps > MaxBasisPoints {
		return fmt.Errorf("constitution quorum floor must be between 1 and %d", MaxBasisPoints)
	}
	if p.EmergencyPauseMaxBlocks == 0 {
		return errors.New("constitution emergency pause max blocks must be positive")
	}
	return nil
}

func (p Params) Authorize(authority string) error {
	if err := p.Validate(); err != nil {
		return err
	}
	if err := addressing.ValidateAuthorityAddress("constitution update authority", authority); err != nil {
		return err
	}
	if authority != p.Authority {
		return errors.New("constitution update requires governance authority")
	}
	return nil
}

func (c Constitution) Validate() error {
	if c.MaxInflationBps > MaxBasisPoints {
		return fmt.Errorf("constitution max inflation must be <= %d bps", MaxBasisPoints)
	}
	if c.MinSlashFractionBps > c.MaxSlashFractionBps || c.MaxSlashFractionBps > MaxBasisPoints {
		return fmt.Errorf("constitution slash fractions must be ordered and <= %d bps", MaxBasisPoints)
	}
	if c.MaxValidatorVotingPowerBps == 0 || c.MaxValidatorVotingPowerBps > MaxBasisPoints {
		return fmt.Errorf("constitution max validator voting power must be between 1 and %d bps", MaxBasisPoints)
	}
	if c.MaxBlockGas == 0 {
		return errors.New("constitution max block gas must be positive")
	}
	if c.MaxAVMCodeSizeBytes == 0 {
		return errors.New("constitution max AVM code size must be positive")
	}
	if c.MaxContractStateSizeBytes == 0 {
		return errors.New("constitution max contract state size must be positive")
	}
	if c.MinStorageRentRate == 0 {
		return errors.New("constitution minimum storage rent rate must be positive")
	}
	if c.TreasurySpendLimitPerEpoch == 0 {
		return errors.New("constitution treasury spend limit must be positive")
	}
	if c.UpgradeDelayBlocks == 0 {
		return errors.New("constitution upgrade delay must be positive")
	}
	if c.EmergencyPauseMaxBlocks == 0 {
		return errors.New("constitution emergency pause max blocks must be positive")
	}
	if c.GovernanceQuorumFloorBps == 0 || c.GovernanceQuorumFloorBps > MaxBasisPoints {
		return fmt.Errorf("constitution quorum floor must be between 1 and %d bps", MaxBasisPoints)
	}
	if len(c.ProtectedModules) == 0 {
		return errors.New("constitution protected module list must be non-empty")
	}
	if err := validateSortedUniqueTokens("constitution protected module", c.ProtectedModules); err != nil {
		return err
	}
	if err := validateSortedUniqueTokens("constitution exception key", c.ConstitutionalExceptionKeys); err != nil {
		return err
	}
	return nil
}

func (a Amendment) Validate(params Params) error {
	if err := params.Validate(); err != nil {
		return err
	}
	if err := validateToken("constitution amendment id", a.ID, params.MaxAmendmentIDBytes); err != nil {
		return err
	}
	if !IsAmendmentStatus(a.Status) {
		return fmt.Errorf("constitution amendment status %q is invalid", a.Status)
	}
	if err := addressing.ValidateAuthorityAddress("constitution amendment proposer", a.Proposer); err != nil {
		return err
	}
	for label, value := range map[string]string{
		"constitution amendment approver":	a.Approver,
		"constitution amendment executor":	a.Executor,
		"constitution amendment canceller":	a.Canceller,
	} {
		if strings.TrimSpace(value) != "" {
			if err := addressing.ValidateAuthorityAddress(label, value); err != nil {
				return err
			}
		}
	}
	if uint32(len(a.Reason)) > MaxReasonBytesV1 {
		return fmt.Errorf("constitution amendment reason exceeds %d bytes", MaxReasonBytesV1)
	}
	if a.CreatedHeight == 0 {
		return errors.New("constitution amendment created height must be positive")
	}
	if a.UpdatedHeight < a.CreatedHeight {
		return errors.New("constitution amendment updated height must not precede created height")
	}
	if a.ExecutableHeight < a.CreatedHeight+params.MinAmendmentDelay {
		return errors.New("constitution amendment executable height violates required delay")
	}
	if uint64(a.YesVotingPowerBps)+uint64(a.NoVotingPowerBps) > uint64(MaxBasisPoints) {
		return fmt.Errorf("constitution amendment voting power must be <= %d bps", MaxBasisPoints)
	}
	return a.Proposed.Validate()
}

func (s State) Validate(params Params) error {
	if err := params.Validate(); err != nil {
		return err
	}
	if err := s.Constitution.Validate(); err != nil {
		return err
	}
	if uint32(len(s.PendingAmendments)) > params.MaxPendingAmendments {
		return fmt.Errorf("constitution pending amendments exceed limit %d", params.MaxPendingAmendments)
	}
	previous := ""
	for _, amendment := range s.PendingAmendments {
		if err := amendment.Validate(params); err != nil {
			return err
		}
		if previous == amendment.ID {
			return fmt.Errorf("constitution amendment %s is duplicated", amendment.ID)
		}
		if previous > amendment.ID {
			return errors.New("constitution amendments must be sorted by id")
		}
		previous = amendment.ID
	}
	return nil
}

func (c Constitution) ProtectedLimits() ProtectedLimits {
	return ProtectedLimits{
		MaxBlockGas:			c.MaxBlockGas,
		MaxAVMCodeSizeBytes:		c.MaxAVMCodeSizeBytes,
		MaxContractStateSizeBytes:	c.MaxContractStateSizeBytes,
		MinStorageRentRate:		c.MinStorageRentRate,
		ProtectedModules:		append([]string(nil), c.ProtectedModules...),
	}
}

func (c Constitution) IsEmergencyPaused(height uint64) bool {
	return c.EmergencyPauseUntilHeight != 0 && height <= c.EmergencyPauseUntilHeight
}

func (c Constitution) Normalize() Constitution {
	c.ProtectedModules = NormalizeTokens(c.ProtectedModules)
	c.ConstitutionalExceptionKeys = NormalizeTokens(c.ConstitutionalExceptionKeys)
	return c
}

func (a Amendment) Normalize(params Params, authority string, height uint64) Amendment {
	a.ID = strings.TrimSpace(a.ID)
	a.Status = AmendmentStatusPending
	a.Proposer = authority
	a.Approver = ""
	a.Executor = ""
	a.Canceller = ""
	a.CreatedHeight = height
	a.UpdatedHeight = height
	a.ExecutableHeight = height + params.MinAmendmentDelay
	a.Proposed = a.Proposed.Normalize()
	return a
}

func SortedAmendments(amendments []Amendment) []Amendment {
	out := make([]Amendment, len(amendments))
	copy(out, amendments)
	sort.Slice(out, func(i, j int) bool {
		return out[i].ID < out[j].ID
	})
	return out
}

func UpsertAmendment(amendments []Amendment, amendment Amendment) []Amendment {
	out := make([]Amendment, 0, len(amendments)+1)
	replaced := false
	for _, existing := range amendments {
		if existing.ID == amendment.ID {
			out = append(out, amendment)
			replaced = true
			continue
		}
		out = append(out, existing)
	}
	if !replaced {
		out = append(out, amendment)
	}
	return SortedAmendments(out)
}

func FindAmendment(amendments []Amendment, id string) (int, Amendment, bool) {
	amendments = SortedAmendments(amendments)
	idx := sort.Search(len(amendments), func(i int) bool {
		return amendments[i].ID >= id
	})
	if idx >= len(amendments) || amendments[idx].ID != id {
		return -1, Amendment{}, false
	}
	return idx, amendments[idx], true
}

func ValidateOrdinaryConfigChange(constitution Constitution, change configtypes.ConfigChange) error {
	constitution = constitution.Normalize()
	if change.Operation == configtypes.OperationDelete && isProtectedModuleKey(change.Key, constitution.ProtectedModules) {
		return errors.New("protected modules cannot be disabled through ordinary config updates")
	}
	if change.Operation != configtypes.OperationSet {
		return nil
	}
	switch change.Key {
	case configtypes.KeyConsensusMaxBlockGas:
		value, err := parseUint(change.Value)
		if err != nil {
			return err
		}
		if value > constitution.MaxBlockGas {
			return errors.New("ordinary config change exceeds constitutional max block gas")
		}
	case "avm/max_code_size":
		value, err := parseUint(change.Value)
		if err != nil {
			return err
		}
		if value > constitution.MaxAVMCodeSizeBytes {
			return errors.New("ordinary config change exceeds constitutional max AVM code size")
		}
	case "storage/max_contract_state_size":
		value, err := parseUint(change.Value)
		if err != nil {
			return err
		}
		if value > constitution.MaxContractStateSizeBytes {
			return errors.New("ordinary config change exceeds constitutional max contract state size")
		}
	case configtypes.KeyStorageRentPerByteEpoch:
		value, err := parseUint(change.Value)
		if err != nil {
			return err
		}
		if value < constitution.MinStorageRentRate && !contains(constitution.ConstitutionalExceptionKeys, change.Key) {
			return errors.New("ordinary config change violates constitutional minimum storage rent")
		}
	case "inflation/max_bps":
		value, err := parseUint(change.Value)
		if err != nil {
			return err
		}
		if value > uint64(constitution.MaxInflationBps) {
			return errors.New("ordinary config change exceeds constitutional max inflation")
		}
	case "validator/max_voting_power_bps":
		value, err := parseUint(change.Value)
		if err != nil {
			return err
		}
		if value > uint64(constitution.MaxValidatorVotingPowerBps) {
			return errors.New("ordinary config change exceeds constitutional max validator voting power")
		}
	}
	if isProtectedModuleDisable(change.Key, change.Value, constitution.ProtectedModules) {
		return errors.New("protected modules cannot be disabled through ordinary config updates")
	}
	return nil
}

func IsAmendmentStatus(value string) bool {
	switch value {
	case AmendmentStatusPending, AmendmentStatusApproved, AmendmentStatusExecuted, AmendmentStatusCancelled:
		return true
	default:
		return false
	}
}

func NormalizeTokens(values []string) []string {
	out := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, found := seen[value]; found {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

func validateSortedUniqueTokens(label string, values []string) error {
	if len(values) == 0 {
		return nil
	}
	previous := ""
	for _, value := range values {
		if err := validateToken(label, value, MaxAmendmentIDBytesV1); err != nil {
			return err
		}
		if previous == value {
			return fmt.Errorf("%s %s is duplicated", label, value)
		}
		if previous > value {
			return fmt.Errorf("%s list must be sorted", label)
		}
		previous = value
	}
	return nil
}

func validateToken(field string, value string, maxBytes uint32) error {
	if strings.TrimSpace(value) != value || value == "" {
		return fmt.Errorf("%s must be canonical", field)
	}
	if uint32(len(value)) > maxBytes {
		return fmt.Errorf("%s exceeds %d bytes", field, maxBytes)
	}
	if strings.ContainsAny(value, " \t\r\n") {
		return fmt.Errorf("%s must not contain whitespace", field)
	}
	return nil
}

func parseUint(value string) (uint64, error) {
	if strings.TrimSpace(value) != value || value == "" {
		return 0, errors.New("constitution config value must be a canonical unsigned integer")
	}
	return strconv.ParseUint(value, 10, 64)
}

func isProtectedModuleKey(key string, modules []string) bool {
	for _, moduleName := range modules {
		if key == "module/enabled/"+moduleName || key == "module/authority/"+moduleName {
			return true
		}
	}
	return false
}

func isProtectedModuleDisable(key string, value string, modules []string) bool {
	for _, moduleName := range modules {
		if key == "module/enabled/"+moduleName && strings.EqualFold(value, "false") {
			return true
		}
	}
	return false
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
