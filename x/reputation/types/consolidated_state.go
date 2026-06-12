package types

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/sovereign-l1/l1/app/addressing"
)

const (
	// ConsolidatedStateVersionV1 is the first version of the consolidated state.
	ConsolidatedStateVersionV1 = uint64(1)
)

// ConsolidatedReputationState is the post-migration unified state.
// No fragmented Account/Validator/Reporter/StakeRecords — only identity records,
// validator scores, and service trust scores. No events, no snapshots.
type ConsolidatedReputationState struct {
	Version			uint64			`json:"version"`
	Params			ReputationParams	`json:"params"`
	EffectParams		ReputationEffectParams	`json:"effect_params"`
	Identities		[]IdentityReputation	`json:"identities"`
	ValidatorScores		[]ValidatorScore	`json:"validator_scores,omitempty"`
	ServiceTrustScores	[]ServiceTrustScore	`json:"service_trust_scores,omitempty"`
	MigrationReceipt	*MigrationReceipt	`json:"migration_receipt,omitempty"`
}

func NewConsolidatedReputationState(params ReputationParams) ConsolidatedReputationState {
	return ConsolidatedReputationState{
		Version:		ConsolidatedStateVersionV1,
		Params:			params,
		EffectParams:		DefaultReputationEffectParams(),
		Identities:		nil,
		ValidatorScores:	nil,
		ServiceTrustScores:	nil,
	}
}

func (state ConsolidatedReputationState) Validate() error {
	state = NormalizeConsolidatedState(state)
	if err := state.Params.Validate(); err != nil {
		return err
	}
	if err := state.EffectParams.Validate(); err != nil {
		return err
	}
	seenIdentities := make(map[string]struct{}, len(state.Identities))
	for i, ip := range state.Identities {
		if err := ValidateIdentityReputation(&ip); err != nil {
			return fmt.Errorf("identity %d invalid: %w", i, err)
		}
		key := addressKeyStr(ip.Account)
		if _, dup := seenIdentities[key]; dup {
			return fmt.Errorf("duplicate identity reputation for %s", ip.Account)
		}
		seenIdentities[key] = struct{}{}
	}
	seenVals := make(map[string]struct{}, len(state.ValidatorScores))
	for i, vs := range state.ValidatorScores {
		if err := validateValidatorScore(&vs); err != nil {
			return fmt.Errorf("validator score %d invalid: %w", i, err)
		}
		key := vs.ValidatorAddress
		if _, dup := seenVals[key]; dup {
			return fmt.Errorf("duplicate validator score for %s", key)
		}
		seenVals[key] = struct{}{}
	}
	seenSvc := make(map[string]struct{}, len(state.ServiceTrustScores))
	for i, sts := range state.ServiceTrustScores {
		if err := validateServiceTrustScore(&sts); err != nil {
			return fmt.Errorf("service trust score %d invalid: %w", i, err)
		}
		key := sts.ServiceAddress
		if _, dup := seenSvc[key]; dup {
			return fmt.Errorf("duplicate service trust score for %s", key)
		}
		seenSvc[key] = struct{}{}
	}
	return nil
}

func NormalizeConsolidatedState(state ConsolidatedReputationState) ConsolidatedReputationState {
	if state.Version == 0 {
		state.Version = ConsolidatedStateVersionV1
	}
	if state.Params.Authority == "" {
		state.Params = DefaultReputationParams()
	}
	state.Params.Authority = strings.TrimSpace(state.Params.Authority)
	if state.EffectParams.MaxFeePremiumBps == 0 {
		state.EffectParams = DefaultReputationEffectParams()
	}
	state.Identities = normalizeIdentities(state.Identities)
	state.ValidatorScores = normalizeValidatorScores(state.ValidatorScores)
	state.ServiceTrustScores = normalizeServiceTrustScores(state.ServiceTrustScores)
	return state
}

func normalizeIdentities(identities []IdentityReputation) []IdentityReputation {
	out := make([]IdentityReputation, len(identities))
	for i, id := range identities {
		if id.Score > IdentityScoreMax {
			id.Score = IdentityScoreMax
		}
		if id.Confidence > ConfidenceMax {
			id.Confidence = ConfidenceMax
		}
		out[i] = id
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].Account < out[j].Account
	})
	return out
}

func normalizeValidatorScores(scores []ValidatorScore) []ValidatorScore {
	out := make([]ValidatorScore, len(scores))
	for i, vs := range scores {
		vs.TotalScore = ComputeValidatorTotalScore(&vs)
		out[i] = vs
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].ValidatorAddress < out[j].ValidatorAddress
	})
	return out
}

func normalizeServiceTrustScores(scores []ServiceTrustScore) []ServiceTrustScore {
	out := make([]ServiceTrustScore, len(scores))
	copy(out, scores)
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].ServiceAddress < out[j].ServiceAddress
	})
	return out
}

func validateValidatorScore(vs *ValidatorScore) error {
	if vs.ValidatorAddress == "" {
		return errors.New("validator score: address must not be empty")
	}
	if vs.UptimeScore > IdentityScoreMax {
		return fmt.Errorf("validator score: uptime %d exceeds max %d", vs.UptimeScore, IdentityScoreMax)
	}
	if vs.MissedBlocksPenalty > IdentityScoreMax {
		return fmt.Errorf("validator score: missed blocks penalty %d exceeds max %d", vs.MissedBlocksPenalty, IdentityScoreMax)
	}
	if vs.SlashingPenalty > IdentityScoreMax {
		return fmt.Errorf("validator score: slashing penalty %d exceeds max %d", vs.SlashingPenalty, IdentityScoreMax)
	}
	if vs.CommissionBehavior > IdentityScoreMax {
		return fmt.Errorf("validator score: commission %d exceeds max %d", vs.CommissionBehavior, IdentityScoreMax)
	}
	if vs.GovernanceParticipation > IdentityScoreMax {
		return fmt.Errorf("validator score: governance %d exceeds max %d", vs.GovernanceParticipation, IdentityScoreMax)
	}
	if vs.PoolAllocationScore > IdentityScoreMax {
		return fmt.Errorf("validator score: pool allocation %d exceeds max %d", vs.PoolAllocationScore, IdentityScoreMax)
	}
	if vs.IsJailed {
		if vs.TotalScore != 0 {
			return fmt.Errorf("validator score: jailed validator must have total_score 0, got %d", vs.TotalScore)
		}
	}
	if vs.IsSlashed {
		if vs.TotalScore != 0 {
			return fmt.Errorf("validator score: slashed validator must have total_score 0, got %d", vs.TotalScore)
		}
	}
	if !vs.IsJailed && !vs.IsSlashed {
		expected := ComputeValidatorTotalScore(vs)
		if vs.TotalScore != expected {
			return fmt.Errorf("validator score: total %d != computed %d", vs.TotalScore, expected)
		}
	}
	return nil
}

func validateServiceTrustScore(sts *ServiceTrustScore) error {
	if sts.ServiceAddress == "" {
		return errors.New("service trust score: address must not be empty")
	}
	if sts.Trust > IdentityScoreMax {
		return fmt.Errorf("service trust score: trust %d exceeds max %d", sts.Trust, IdentityScoreMax)
	}
	if sts.Reliability > IdentityScoreMax {
		return fmt.Errorf("service trust score: reliability %d exceeds max %d", sts.Reliability, IdentityScoreMax)
	}
	return nil
}

func FindIdentity(state ConsolidatedReputationState, account string) (IdentityReputation, bool) {
	state = NormalizeConsolidatedState(state)
	for _, id := range state.Identities {
		if addressKeyStr(id.Account) == addressKeyStr(account) {
			return id, true
		}
	}
	return IdentityReputation{}, false
}

func UpsertIdentity(state ConsolidatedReputationState, identity IdentityReputation) ConsolidatedReputationState {
	state = NormalizeConsolidatedState(state)
	identity.Account = strings.TrimSpace(identity.Account)
	key := addressKeyStr(identity.Account)
	found := false
	for i, existing := range state.Identities {
		if addressKeyStr(existing.Account) == key {
			state.Identities[i] = identity
			found = true
			break
		}
	}
	if !found {
		state.Identities = append(state.Identities, identity)
	}
	return NormalizeConsolidatedState(state)
}

func FindValidatorScore(state ConsolidatedReputationState, addr string) (ValidatorScore, bool) {
	state = NormalizeConsolidatedState(state)
	for _, vs := range state.ValidatorScores {
		if vs.ValidatorAddress == addr {
			return vs, true
		}
	}
	return ValidatorScore{}, false
}

func UpsertValidatorScore(state ConsolidatedReputationState, vs ValidatorScore) ConsolidatedReputationState {
	state = NormalizeConsolidatedState(state)
	vs.ValidatorAddress = strings.TrimSpace(vs.ValidatorAddress)
	key := vs.ValidatorAddress
	found := false
	for i, existing := range state.ValidatorScores {
		if existing.ValidatorAddress == key {
			state.ValidatorScores[i] = vs
			found = true
			break
		}
	}
	if !found {
		state.ValidatorScores = append(state.ValidatorScores, vs)
	}
	return NormalizeConsolidatedState(state)
}

func FindServiceTrustScore(state ConsolidatedReputationState, addr string) (ServiceTrustScore, bool) {
	state = NormalizeConsolidatedState(state)
	for _, sts := range state.ServiceTrustScores {
		if sts.ServiceAddress == addr {
			return sts, true
		}
	}
	return ServiceTrustScore{}, false
}

func UpsertServiceTrustScore(state ConsolidatedReputationState, sts ServiceTrustScore) ConsolidatedReputationState {
	state = NormalizeConsolidatedState(state)
	sts.ServiceAddress = strings.TrimSpace(sts.ServiceAddress)
	key := sts.ServiceAddress
	found := false
	for i, existing := range state.ServiceTrustScores {
		if existing.ServiceAddress == key {
			state.ServiceTrustScores[i] = sts
			found = true
			break
		}
	}
	if !found {
		state.ServiceTrustScores = append(state.ServiceTrustScores, sts)
	}
	return NormalizeConsolidatedState(state)
}

func MigrateFromReputationState(old ReputationState) ConsolidatedReputationState {
	old = NormalizeReputationState(old)
	params := old.Params

	identities := make([]IdentityReputation, 0, len(old.Accounts))
	for _, acc := range old.Accounts {
		id := identityFromAccountRecord(acc)
		identities = append(identities, *id)
	}

	for _, stakeRec := range old.StakeRecords {
		userAddr := addressing.FormatAccAddress(stakeRec.Account)
		idx := findIdentityByIdx(identities, userAddr)
		var id *IdentityReputation
		if idx >= 0 {
			id = &identities[idx]
		} else {
			id = NewIdentityReputation(userAddr)
		}
		id.RecordStakeTime(stakeRec.StakeWeightedSeconds, stakeRec.LastUpdatedUnix)
		id.Score = ComputeIdentityScore(id)
		id.Confidence = ComputeConfidence(id)
		if idx < 0 {
			identities = append(identities, *id)
		}
	}

	validatorScores := make([]ValidatorScore, 0, len(old.Validators))
	for _, val := range old.Validators {
		vs := validatorScoreFromRecord(val)
		validatorScores = append(validatorScores, *vs)
	}

	return ConsolidatedReputationState{
		Version:		ConsolidatedStateVersionV1,
		Params:			params,
		Identities:		normalizeIdentities(identities),
		ValidatorScores:	normalizeValidatorScores(validatorScores),
		ServiceTrustScores:	nil,
	}
}

func identityFromAccountRecord(record ReputationRecord) *IdentityReputation {
	userAddr := addressing.FormatAccAddress(record.Account)
	id := NewIdentityReputation(userAddr)
	if record.TxSuccessScore > 0 {
		for i := uint16(0); i < record.TxSuccessScore; i++ {
			id.RecordSuccessfulTx(record.LastUpdatedEpoch)
		}
	}
	if record.SpamPenalty > 0 {
		for i := uint16(0); i < record.SpamPenalty; i++ {
			id.RecordSpam(record.LastUpdatedEpoch)
		}
	}
	if record.SlashPenalty > 0 {
		for i := uint16(0); i < record.SlashPenalty; i++ {
			id.RecordSlashEvent(record.LastUpdatedEpoch)
		}
	}
	id.Score = uint32(record.Score) * 10000 / 255
	if id.Score > IdentityScoreMax {
		id.Score = IdentityScoreMax
	}
	id.Confidence = ConfidenceDefault
	return id
}

func validatorScoreFromRecord(record ReputationRecord) *ValidatorScore {
	addr := addressing.FormatAccAddress(record.Account)
	vs := NewValidatorScore(addr)
	if record.TxSuccessScore > 0 {
		vs.UptimeScore = uint32(record.TxSuccessScore) * 100
	}
	if record.FailedTxPenalty > 0 {
		vs.MissedBlocksPenalty = uint32(record.FailedTxPenalty) * 10
	}
	if record.SlashPenalty > 0 {
		vs.SlashingPenalty = uint32(record.SlashPenalty) * 10
	}
	vs.TotalScore = ComputeValidatorTotalScore(vs)
	return vs
}

func findIdentityByIdx(identities []IdentityReputation, account string) int {
	key := addressKeyStr(account)
	for i, id := range identities {
		if addressKeyStr(id.Account) == key {
			return i
		}
	}
	return -1
}

func addressKeyStr(addr string) string {
	return strings.TrimSpace(strings.ToLower(addr))
}

func ExportConsolidatedState(state ConsolidatedReputationState) (ConsolidatedReputationState, error) {
	state = NormalizeConsolidatedState(state)
	if err := state.Validate(); err != nil {
		return ConsolidatedReputationState{}, err
	}
	return cloneConsolidatedState(state), nil
}

func ImportConsolidatedState(exported ConsolidatedReputationState) (ConsolidatedReputationState, error) {
	exported = NormalizeConsolidatedState(exported)
	if err := exported.Validate(); err != nil {
		return ConsolidatedReputationState{}, err
	}
	return cloneConsolidatedState(exported), nil
}

func cloneConsolidatedState(state ConsolidatedReputationState) ConsolidatedReputationState {
	ids := make([]IdentityReputation, len(state.Identities))
	copy(ids, state.Identities)
	vs := make([]ValidatorScore, len(state.ValidatorScores))
	copy(vs, state.ValidatorScores)
	sts := make([]ServiceTrustScore, len(state.ServiceTrustScores))
	copy(sts, state.ServiceTrustScores)
	var receipt *MigrationReceipt
	if state.MigrationReceipt != nil {
		r := *state.MigrationReceipt
		receipt = &r
	}
	return ConsolidatedReputationState{
		Version:		state.Version,
		Params:			state.Params,
		EffectParams:		state.EffectParams,
		Identities:		ids,
		ValidatorScores:	vs,
		ServiceTrustScores:	sts,
		MigrationReceipt:	receipt,
	}
}

// Legacy serialization support — allow marshal/unmarshal for genesis.
var _ json.Marshaler = ConsolidatedReputationState{}
var _ json.Unmarshaler = &ConsolidatedReputationState{}

func (state ConsolidatedReputationState) MarshalJSON() ([]byte, error) {
	clone := NormalizeConsolidatedState(state)
	raw := struct {
		Version			uint64			`json:"version"`
		Params			ReputationParams	`json:"params"`
		EffectParams		ReputationEffectParams	`json:"effect_params"`
		Identities		[]IdentityReputation	`json:"identities"`
		ValidatorScores		[]ValidatorScore	`json:"validator_scores,omitempty"`
		ServiceTrustScores	[]ServiceTrustScore	`json:"service_trust_scores,omitempty"`
		MigrationReceipt	*MigrationReceipt	`json:"migration_receipt,omitempty"`
	}{
		Version:		clone.Version,
		Params:			clone.Params,
		EffectParams:		clone.EffectParams,
		Identities:		clone.Identities,
		ValidatorScores:	clone.ValidatorScores,
		ServiceTrustScores:	clone.ServiceTrustScores,
		MigrationReceipt:	clone.MigrationReceipt,
	}
	return json.Marshal(raw)
}

func (state *ConsolidatedReputationState) UnmarshalJSON(bz []byte) error {
	var raw struct {
		Version			uint64			`json:"version"`
		Params			ReputationParams	`json:"params"`
		EffectParams		ReputationEffectParams	`json:"effect_params"`
		Identities		[]IdentityReputation	`json:"identities"`
		ValidatorScores		[]ValidatorScore	`json:"validator_scores,omitempty"`
		ServiceTrustScores	[]ServiceTrustScore	`json:"service_trust_scores,omitempty"`
		MigrationReceipt	*MigrationReceipt	`json:"migration_receipt,omitempty"`
	}
	if err := json.Unmarshal(bz, &raw); err != nil {
		return err
	}
	state.Version = raw.Version
	state.Params = raw.Params
	state.EffectParams = raw.EffectParams
	state.Identities = raw.Identities
	state.ValidatorScores = raw.ValidatorScores
	state.ServiceTrustScores = raw.ServiceTrustScores
	state.MigrationReceipt = raw.MigrationReceipt
	*state = NormalizeConsolidatedState(*state)
	return nil
}
