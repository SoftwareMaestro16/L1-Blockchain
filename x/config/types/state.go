package types

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/sovereign-l1/l1/app/addressing"
	appparams "github.com/sovereign-l1/l1/app/params"
	"github.com/sovereign-l1/l1/x/internal/prototype"
)

const (
	MaxConfigEntriesV1	= uint32(128)
	MaxConfigKeyBytesV1	= uint32(96)
	MaxConfigValueBytesV1	= uint32(4096)
	MaxPendingChangesV1	= uint32(128)
	MaxChangeIDBytesV1	= uint32(96)
	DefaultActivationDelay	= uint64(10)
	DefaultCriticalDelay	= uint64(100)
	DefaultEpochLength	= uint64(50)

	OperationSet	= "set"
	OperationDelete	= "delete"

	ChangeStatusPending	= "pending"
	ChangeStatusApproved	= "approved"
	ChangeStatusRejected	= "rejected"
	ChangeStatusExecuted	= "executed"
	ChangeStatusCancelled	= "cancelled"

	KeyConsensusMaxBlockGas		= "consensus/max_block_gas"
	KeyFeeBaseDenom			= "fee/base_denom"
	KeyStorageRentPerByteEpoch	= "storage/rent_per_byte_epoch"
	KeyStorageContractStateActive	= "storage/contract_state_non_empty"
	KeyConstitutionZeroRentAllow	= "constitution/allow_zero_storage_rent"

	MaxConsensusBlockGasV1	= uint64(1_000_000_000)
	MaxConfigUintValueV1	= uint64(1_000_000_000_000_000)
)

type Params struct {
	Authority			string
	MaxEntries			uint32
	MaxPendingChanges		uint32
	MaxKeyBytes			uint32
	MaxValueBytes			uint32
	MaxChangeIDBytes		uint32
	MinActivationDelay		uint64
	CriticalActivationDelay		uint64
	ActivationEpochLength		uint64
	BaseDenom			string
	RequiredSystemAccountKeys	[]string
}

type ConfigEntry struct {
	Key		string
	Value		string
	Owner		string
	Version		uint64
	UpdatedHeight	int64
}

type ConfigChange struct {
	ID					string
	Key					string
	Value					string
	Operation				string
	Status					string
	SubmittedBy				string
	ApprovedBy				string
	RejectedBy				string
	CancelledBy				string
	ExecutedBy				string
	Reason					string
	RequiresConstitutionalException		bool
	Critical				bool
	CreatedHeight				int64
	UpdatedHeight				int64
	ActivationHeight			int64
	ActivationEpoch				uint64
	ExpectedPreviousVersion			uint64
	AllowMissingExpectedPreviousVersion	bool
}

type ConfigState struct {
	Entries		[]ConfigEntry
	PendingChanges	[]ConfigChange
}

func DefaultParams() Params {
	return Params{
		Authority:			prototype.DefaultAuthority,
		MaxEntries:			MaxConfigEntriesV1,
		MaxPendingChanges:		MaxPendingChangesV1,
		MaxKeyBytes:			MaxConfigKeyBytesV1,
		MaxValueBytes:			MaxConfigValueBytesV1,
		MaxChangeIDBytes:		MaxChangeIDBytesV1,
		MinActivationDelay:		DefaultActivationDelay,
		CriticalActivationDelay:	DefaultCriticalDelay,
		ActivationEpochLength:		DefaultEpochLength,
		BaseDenom:			appparams.BaseDenom,
		RequiredSystemAccountKeys:	[]string{"system/account/fee_collector", "system/account/treasury"},
	}
}

func (p Params) Validate() error {
	if err := addressing.ValidateAuthorityAddress("config authority", p.Authority); err != nil {
		return err
	}
	if p.MaxEntries == 0 || p.MaxEntries > MaxConfigEntriesV1 {
		return fmt.Errorf("config max entries must be between 1 and %d", MaxConfigEntriesV1)
	}
	if p.MaxPendingChanges == 0 || p.MaxPendingChanges > MaxPendingChangesV1 {
		return fmt.Errorf("config max pending changes must be between 1 and %d", MaxPendingChangesV1)
	}
	if p.MaxKeyBytes == 0 || p.MaxKeyBytes > MaxConfigKeyBytesV1 {
		return fmt.Errorf("config max key bytes must be between 1 and %d", MaxConfigKeyBytesV1)
	}
	if p.MaxValueBytes == 0 || p.MaxValueBytes > MaxConfigValueBytesV1 {
		return fmt.Errorf("config max value bytes must be between 1 and %d", MaxConfigValueBytesV1)
	}
	if p.MaxChangeIDBytes == 0 || p.MaxChangeIDBytes > MaxChangeIDBytesV1 {
		return fmt.Errorf("config max change id bytes must be between 1 and %d", MaxChangeIDBytesV1)
	}
	if p.CriticalActivationDelay < p.MinActivationDelay {
		return errors.New("config critical activation delay must be >= minimum activation delay")
	}
	if p.ActivationEpochLength == 0 {
		return errors.New("config activation epoch length must be positive")
	}
	if strings.TrimSpace(p.BaseDenom) != p.BaseDenom || p.BaseDenom == "" {
		return errors.New("config base denom must be canonical")
	}
	if p.BaseDenom != appparams.BaseDenom {
		return fmt.Errorf("config base denom must match native base denom %s", appparams.BaseDenom)
	}
	previous := ""
	for _, key := range SortedStrings(p.RequiredSystemAccountKeys) {
		if err := ValidateConfigKey("required system account key", key, p.MaxKeyBytes); err != nil {
			return err
		}
		if !strings.HasPrefix(key, "system/account/") {
			return fmt.Errorf("required system account key %s must use system/account/ prefix", key)
		}
		if previous == key {
			return fmt.Errorf("required system account key %s is duplicated", key)
		}
		previous = key
	}
	return nil
}

func (p Params) Authorize(authority string) error {
	if err := p.Validate(); err != nil {
		return err
	}
	if err := addressing.ValidateAuthorityAddress("config update authority", authority); err != nil {
		return err
	}
	if authority != p.Authority {
		return errors.New("config update requires governance authority")
	}
	return nil
}

func (e ConfigEntry) Validate(params Params) error {
	if err := params.Validate(); err != nil {
		return err
	}
	if err := ValidateConfigKey("config entry key", e.Key, params.MaxKeyBytes); err != nil {
		return err
	}
	if uint32(len(e.Value)) > params.MaxValueBytes {
		return fmt.Errorf("config entry value exceeds %d bytes", params.MaxValueBytes)
	}
	if err := addressing.ValidateAuthorityAddress("config entry owner", e.Owner); err != nil {
		return err
	}
	if e.Version == 0 {
		return errors.New("config entry version must be positive")
	}
	if e.UpdatedHeight < 0 {
		return errors.New("config entry updated height must be non-negative")
	}
	if err := ValidateSemanticConfigValue(params, e.Key, e.Value, ConfigState{}); err != nil {
		return err
	}
	return nil
}

func (c ConfigChange) Validate(params Params) error {
	if err := params.Validate(); err != nil {
		return err
	}
	if err := ValidateConfigKey("config change id", c.ID, params.MaxChangeIDBytes); err != nil {
		return err
	}
	if err := ValidateConfigKey("config change key", c.Key, params.MaxKeyBytes); err != nil {
		return err
	}
	if uint32(len(c.Value)) > params.MaxValueBytes {
		return fmt.Errorf("config change value exceeds %d bytes", params.MaxValueBytes)
	}
	if !IsChangeOperation(c.Operation) {
		return fmt.Errorf("config change operation %q is invalid", c.Operation)
	}
	if !IsChangeStatus(c.Status) {
		return fmt.Errorf("config change status %q is invalid", c.Status)
	}
	if err := addressing.ValidateAuthorityAddress("config change submitter", c.SubmittedBy); err != nil {
		return err
	}
	for label, value := range map[string]string{
		"config change approver":	c.ApprovedBy,
		"config change rejector":	c.RejectedBy,
		"config change canceller":	c.CancelledBy,
		"config change executor":	c.ExecutedBy,
	} {
		if strings.TrimSpace(value) != "" {
			if err := addressing.ValidateAuthorityAddress(label, value); err != nil {
				return err
			}
		}
	}
	if c.CreatedHeight < 0 || c.UpdatedHeight < 0 {
		return errors.New("config change heights must be non-negative")
	}
	if c.UpdatedHeight < c.CreatedHeight {
		return errors.New("config change updated height must not precede created height")
	}
	if c.ActivationHeight != 0 {
		if c.ActivationHeight < c.CreatedHeight {
			return errors.New("config change activation height must not precede created height")
		}
		expectedEpoch := ActivationEpoch(uint64(c.ActivationHeight), params.ActivationEpochLength)
		if c.ActivationEpoch != expectedEpoch {
			return fmt.Errorf("config change activation epoch mismatch: expected %d got %d", expectedEpoch, c.ActivationEpoch)
		}
	}
	if c.Operation == OperationDelete && c.Value != "" {
		return errors.New("config delete change value must be empty")
	}
	return nil
}

func (s ConfigState) Validate(params Params) error {
	if err := params.Validate(); err != nil {
		return err
	}
	if uint32(len(s.Entries)) > params.MaxEntries {
		return fmt.Errorf("config entries exceed limit %d", params.MaxEntries)
	}
	if uint32(len(s.PendingChanges)) > params.MaxPendingChanges {
		return fmt.Errorf("config pending changes exceed limit %d", params.MaxPendingChanges)
	}
	var previous string
	for i, entry := range s.Entries {
		if err := entry.Validate(params); err != nil {
			return err
		}
		if i > 0 {
			if previous == entry.Key {
				return fmt.Errorf("config entry key %s is duplicated", entry.Key)
			}
			if previous > entry.Key {
				return errors.New("config entries must be sorted by key")
			}
		}
		previous = entry.Key
	}
	for _, entry := range s.Entries {
		if err := ValidateSemanticConfigValue(params, entry.Key, entry.Value, s); err != nil {
			return err
		}
	}
	previous = ""
	for i, change := range s.PendingChanges {
		if err := change.Validate(params); err != nil {
			return err
		}
		if i > 0 {
			if previous == change.ID {
				return fmt.Errorf("config change id %s is duplicated", change.ID)
			}
			if previous > change.ID {
				return errors.New("config pending changes must be sorted by id")
			}
		}
		previous = change.ID
	}
	if err := s.ValidateRequiredSystemAccounts(params); err != nil {
		return err
	}
	return nil
}

func (s ConfigState) ValidateRequiredSystemAccounts(params Params) error {
	for _, key := range params.RequiredSystemAccountKeys {
		entry, found := FindEntry(s.Entries, key)
		if !found {
			continue
		}
		if err := addressing.ValidateAuthorityAddress("required system account "+key, entry.Value); err != nil {
			return err
		}
	}
	return nil
}

func CloneState(state ConfigState) ConfigState {
	out := ConfigState{
		Entries:	make([]ConfigEntry, len(state.Entries)),
		PendingChanges:	make([]ConfigChange, len(state.PendingChanges)),
	}
	copy(out.Entries, state.Entries)
	copy(out.PendingChanges, state.PendingChanges)
	return out
}

func SortedEntries(entries []ConfigEntry) []ConfigEntry {
	out := make([]ConfigEntry, len(entries))
	copy(out, entries)
	sort.Slice(out, func(i, j int) bool {
		return out[i].Key < out[j].Key
	})
	return out
}

func SortedChanges(changes []ConfigChange) []ConfigChange {
	out := make([]ConfigChange, len(changes))
	copy(out, changes)
	sort.Slice(out, func(i, j int) bool {
		return out[i].ID < out[j].ID
	})
	return out
}

func SortedStrings(values []string) []string {
	out := append([]string(nil), values...)
	sort.Strings(out)
	return out
}

func FindEntry(entries []ConfigEntry, key string) (ConfigEntry, bool) {
	entries = SortedEntries(entries)
	idx := sort.Search(len(entries), func(i int) bool {
		return entries[i].Key >= key
	})
	if idx >= len(entries) || entries[idx].Key != key {
		return ConfigEntry{}, false
	}
	return entries[idx], true
}

func FindChange(changes []ConfigChange, id string) (int, ConfigChange, bool) {
	changes = SortedChanges(changes)
	idx := sort.Search(len(changes), func(i int) bool {
		return changes[i].ID >= id
	})
	if idx >= len(changes) || changes[idx].ID != id {
		return -1, ConfigChange{}, false
	}
	return idx, changes[idx], true
}

func UpsertChange(changes []ConfigChange, change ConfigChange) []ConfigChange {
	out := make([]ConfigChange, 0, len(changes)+1)
	replaced := false
	for _, existing := range changes {
		if existing.ID == change.ID {
			out = append(out, change)
			replaced = true
			continue
		}
		out = append(out, existing)
	}
	if !replaced {
		out = append(out, change)
	}
	return SortedChanges(out)
}

func IsChangeOperation(value string) bool {
	switch value {
	case OperationSet, OperationDelete:
		return true
	default:
		return false
	}
}

func IsChangeStatus(value string) bool {
	switch value {
	case ChangeStatusPending, ChangeStatusApproved, ChangeStatusRejected, ChangeStatusExecuted, ChangeStatusCancelled:
		return true
	default:
		return false
	}
}

func ValidateConfigKey(field, key string, maxBytes uint32) error {
	raw := key
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return fmt.Errorf("%s must be set", field)
	}
	if uint32(len(raw)) > maxBytes {
		return fmt.Errorf("%s exceeds %d bytes", field, maxBytes)
	}
	if trimmed != raw {
		return fmt.Errorf("%s must be canonical", field)
	}
	if strings.ContainsAny(raw, " \t\r\n") {
		return fmt.Errorf("%s must not contain whitespace", field)
	}
	return nil
}

func ValidateSemanticConfigValue(params Params, key string, value string, state ConfigState) error {
	switch {
	case key == KeyConsensusMaxBlockGas:
		amount, err := parseBoundedUint(key, value, true)
		if err != nil {
			return err
		}
		if amount > MaxConsensusBlockGasV1 {
			return fmt.Errorf("config cannot set unlimited block gas; max is %d", MaxConsensusBlockGasV1)
		}
	case key == KeyFeeBaseDenom:
		if value != params.BaseDenom {
			return fmt.Errorf("config fee denom must match base denom policy %s", params.BaseDenom)
		}
	case key == KeyStorageRentPerByteEpoch:
		amount, err := parseBoundedUint(key, value, false)
		if err != nil {
			return err
		}
		if amount == 0 && contractStateNotEmpty(state) && !constitutionalZeroRentAllowed(state) {
			return errors.New("config cannot set zero storage rent for non-empty contract state without constitutional allowance")
		}
	case strings.HasPrefix(key, "avm/gas/"):
		amount, err := parseBoundedUint(key, value, true)
		if err != nil {
			return err
		}
		if amount == 0 {
			return errors.New("config gas schedule cannot contain zero-cost execution paths")
		}
	case strings.HasPrefix(key, "system/account/") || strings.HasPrefix(key, "module/authority/"):
		if err := addressing.ValidateAuthorityAddress("config address value "+key, value); err != nil {
			return err
		}
	}
	return nil
}

func ValidateChangeAgainstState(params Params, state ConfigState, change ConfigChange) error {
	if err := change.Validate(params); err != nil {
		return err
	}
	if isRequiredSystemAccount(params, change.Key) {
		if change.Operation == OperationDelete {
			return errors.New("config cannot remove required system account addresses")
		}
		if err := addressing.ValidateAuthorityAddress("required system account "+change.Key, change.Value); err != nil {
			return err
		}
	}
	if change.Operation == OperationSet {
		if err := ValidateSemanticConfigValue(params, change.Key, change.Value, state); err != nil {
			return err
		}
		if change.Key == KeyStorageRentPerByteEpoch && change.Value == "0" && contractStateNotEmpty(state) && !change.RequiresConstitutionalException {
			return errors.New("zero storage rent change requires explicit constitutional exception")
		}
	}
	if change.Operation == OperationDelete {
		if _, found := FindEntry(state.Entries, change.Key); !found && !change.AllowMissingExpectedPreviousVersion {
			return errors.New("config delete change target is missing")
		}
	}
	if change.ExpectedPreviousVersion != 0 {
		entry, found := FindEntry(state.Entries, change.Key)
		if !found {
			return errors.New("config expected previous version target is missing")
		}
		if entry.Version != change.ExpectedPreviousVersion {
			return fmt.Errorf("config expected previous version mismatch: expected %d got %d", change.ExpectedPreviousVersion, entry.Version)
		}
	}
	return nil
}

func IsCriticalConfigKey(key string) bool {
	return key == KeyConsensusMaxBlockGas ||
		key == KeyStorageRentPerByteEpoch ||
		key == KeyStorageContractStateActive ||
		strings.HasPrefix(key, "avm/") ||
		strings.HasPrefix(key, "consensus/") ||
		strings.HasPrefix(key, "module/") ||
		strings.HasPrefix(key, "system/")
}

func ActivationHeight(params Params, key string, createdHeight int64) (int64, uint64) {
	if createdHeight < 0 {
		createdHeight = 0
	}
	delay := params.MinActivationDelay
	if IsCriticalConfigKey(key) {
		delay = params.CriticalActivationDelay
	}
	height := uint64(createdHeight) + delay
	if params.ActivationEpochLength > 0 {
		remainder := height % params.ActivationEpochLength
		if remainder != 0 {
			height += params.ActivationEpochLength - remainder
		}
	}
	return int64(height), ActivationEpoch(height, params.ActivationEpochLength)
}

func ActivationEpoch(height uint64, epochLength uint64) uint64 {
	if epochLength == 0 {
		return 0
	}
	return height / epochLength
}

func isRequiredSystemAccount(params Params, key string) bool {
	for _, required := range params.RequiredSystemAccountKeys {
		if required == key {
			return true
		}
	}
	return false
}

func contractStateNotEmpty(state ConfigState) bool {
	entry, found := FindEntry(state.Entries, KeyStorageContractStateActive)
	return found && strings.EqualFold(entry.Value, "true")
}

func constitutionalZeroRentAllowed(state ConfigState) bool {
	entry, found := FindEntry(state.Entries, KeyConstitutionZeroRentAllow)
	return found && strings.EqualFold(entry.Value, "true")
}

func parseBoundedUint(key string, value string, requirePositive bool) (uint64, error) {
	if strings.TrimSpace(value) != value || value == "" {
		return 0, fmt.Errorf("config %s must be a canonical unsigned integer", key)
	}
	amount, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("config %s must be a canonical unsigned integer: %w", key, err)
	}
	if requirePositive && amount == 0 {
		return 0, fmt.Errorf("config %s must be positive", key)
	}
	if amount > MaxConfigUintValueV1 {
		return 0, fmt.Errorf("config %s exceeds max %d", key, MaxConfigUintValueV1)
	}
	return amount, nil
}

type MsgSubmitConfigChange struct {
	Authority	string
	Change		ConfigChange
}

type MsgApproveConfigChange struct {
	Authority	string
	ChangeID	string
}

type MsgRejectConfigChange struct {
	Authority	string
	ChangeID	string
	Reason		string
}

type MsgExecuteConfigChange struct {
	Authority	string
	ChangeID	string
}

type MsgCancelConfigChange struct {
	Authority	string
	ChangeID	string
	Reason		string
}
