package types

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

type MigrationSource string

const (
	MigrationSourceAccount		MigrationSource	= "account"
	MigrationSourceStake		MigrationSource	= "stake"
	MigrationSourceValidator	MigrationSource	= "validator"
	MigrationSourceReporter		MigrationSource	= "reporter"
)

type MigrationEntry struct {
	Source		MigrationSource	`json:"source"`
	OriginalIndex	int		`json:"original_index"`
	Account		string		`json:"account"`
	ScoreBefore	uint32		`json:"score_before"`
	ScoreAfter	uint32		`json:"score_after"`
	ConfidenceAfter	uint32		`json:"confidence_after"`
	Merged		bool		`json:"merged"`
	MergeReason	string		`json:"merge_reason,omitempty"`
}

type MigrationReceipt struct {
	Version			uint32			`json:"version"`
	MigrationHeight		uint64			`json:"migration_height"`
	TotalAccounts		int			`json:"total_accounts"`
	TotalStakeRecords	int			`json:"total_stake_records"`
	TotalValidators		int			`json:"total_validators"`
	TotalReporters		int			`json:"total_reporters"`
	IdentitiesCreated	int			`json:"identities_created"`
	IdentitiesMerged	int			`json:"identities_merged"`
	ValidatorScores		int			`json:"validator_scores_created"`
	ServiceTrustScores	int			`json:"service_trust_scores_created"`
	ReportersDropped	int			`json:"reporters_dropped"`
	Entries			[]MigrationEntry	`json:"entries"`
	DeterministicHash	string			`json:"deterministic_hash"`
}

func (r MigrationReceipt) Validate() error {
	if r.TotalAccounts < 0 || r.TotalStakeRecords < 0 || r.TotalValidators < 0 || r.TotalReporters < 0 {
		return fmt.Errorf("migration receipt: source counts must be non-negative")
	}
	if r.IdentitiesCreated < 0 || r.IdentitiesMerged < 0 {
		return fmt.Errorf("migration receipt: identity counts must be non-negative")
	}
	if r.ReportersDropped != r.TotalReporters {
		return fmt.Errorf("migration receipt: all %d reporter records must be dropped (dropped=%d)", r.TotalReporters, r.ReportersDropped)
	}
	expectedHash := ComputeMigrationReceiptHash(r)
	if r.DeterministicHash != expectedHash {
		return fmt.Errorf("migration receipt: deterministic hash mismatch (expected %s, got %s)", expectedHash, r.DeterministicHash)
	}
	return nil
}

func ComputeMigrationReceiptHash(r MigrationReceipt) string {
	sorted := make([]MigrationEntry, len(r.Entries))
	copy(sorted, r.Entries)
	sort.SliceStable(sorted, func(i, j int) bool {
		if sorted[i].Source != sorted[j].Source {
			return string(sorted[i].Source) < string(sorted[j].Source)
		}
		if sorted[i].OriginalIndex != sorted[j].OriginalIndex {
			return sorted[i].OriginalIndex < sorted[j].OriginalIndex
		}
		return strings.ToLower(sorted[i].Account) < strings.ToLower(sorted[j].Account)
	})

	type receiptForHash struct {
		Version			uint32			`json:"version"`
		MigrationHeight		uint64			`json:"migration_height"`
		TotalAccounts		int			`json:"total_accounts"`
		TotalStakeRecords	int			`json:"total_stake_records"`
		TotalValidators		int			`json:"total_validators"`
		TotalReporters		int			`json:"total_reporters"`
		IdentitiesCreated	int			`json:"identities_created"`
		IdentitiesMerged	int			`json:"identities_merged"`
		ValidatorScores		int			`json:"validator_scores_created"`
		ServiceTrustScores	int			`json:"service_trust_scores_created"`
		ReportersDropped	int			`json:"reporters_dropped"`
		Entries			[]MigrationEntry	`json:"entries"`
	}

	data, _ := json.Marshal(receiptForHash{
		Version:		r.Version,
		MigrationHeight:	r.MigrationHeight,
		TotalAccounts:		r.TotalAccounts,
		TotalStakeRecords:	r.TotalStakeRecords,
		TotalValidators:	r.TotalValidators,
		TotalReporters:		r.TotalReporters,
		IdentitiesCreated:	r.IdentitiesCreated,
		IdentitiesMerged:	r.IdentitiesMerged,
		ValidatorScores:	r.ValidatorScores,
		ServiceTrustScores:	r.ServiceTrustScores,
		ReportersDropped:	r.ReportersDropped,
		Entries:		sorted,
	})
	h := sha256.Sum256(data)
	return fmt.Sprintf("%x", h)
}

func MigrateFromReputationStateWithReceipt(old ReputationState, migrationHeight uint64) (ConsolidatedReputationState, MigrationReceipt) {
	old = NormalizeReputationState(old)
	params := old.Params
	params.Authority = strings.TrimSpace(params.Authority)

	entries := make([]MigrationEntry, 0)
	identities := make([]IdentityReputation, 0)
	identitiesMerged := 0
	stakeTimeClaimed := make(map[string]uint64)

	for i, acc := range old.Accounts {
		id := identityFromAccountRecord(acc)
		entry := MigrationEntry{
			Source:			MigrationSourceAccount,
			OriginalIndex:		i,
			Account:		id.Account,
			ScoreBefore:		uint32(acc.Score) * 10000 / 255,
			ScoreAfter:		id.Score,
			ConfidenceAfter:	id.Confidence,
			Merged:			false,
		}
		entries = append(entries, entry)
		identities = append(identities, *id)
	}

	for i, stakeRec := range old.StakeRecords {
		userAddr := formatAddress(stakeRec.Account)
		if userAddr == "" {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(userAddr))
		if _, alreadyClaimed := stakeTimeClaimed[key]; alreadyClaimed {
			entries = append(entries, MigrationEntry{
				Source:		MigrationSourceStake,
				OriginalIndex:	i,
				Account:	userAddr,
				Merged:		false,
				MergeReason:	"duplicate_stake_time_rejected",
			})
			continue
		}
		idx := findIdentityByIdx(identities, userAddr)
		var id *IdentityReputation
		if idx >= 0 {
			id = &identities[idx]
			identitiesMerged++
		} else {
			newId := NewIdentityReputation(userAddr)
			id = newId
		}

		stakeSeconds := stakeRec.StakeWeightedSeconds
		id.RecordStakeTime(stakeSeconds, stakeRec.LastUpdatedUnix)
		id.Score = ComputeIdentityScore(id)
		id.Confidence = ComputeConfidence(id)

		stakeTimeClaimed[key] = stakeSeconds

		entries = append(entries, MigrationEntry{
			Source:			MigrationSourceStake,
			OriginalIndex:		i,
			Account:		userAddr,
			ScoreBefore:		id.Score,
			ScoreAfter:		ComputeIdentityScore(id),
			ConfidenceAfter:	id.Confidence,
			Merged:			idx >= 0,
			MergeReason:		"stake_time_merged_into_identity",
		})

		if idx < 0 {
			identities = append(identities, *id)
		} else {
			identities[idx] = *id
		}
	}

	validatorScores := make([]ValidatorScore, 0)
	for i, val := range old.Validators {
		vs := validatorScoreFromRecord(val)
		entries = append(entries, MigrationEntry{
			Source:		MigrationSourceValidator,
			OriginalIndex:	i,
			Account:	vs.ValidatorAddress,
			ScoreBefore:	uint32(val.Score) * 10000 / 255,
			ScoreAfter:	vs.TotalScore,
			Merged:		false,
		})
		validatorScores = append(validatorScores, *vs)
	}

	reportersDropped := len(old.Reporters)
	for i, rep := range old.Reporters {
		entries = append(entries, MigrationEntry{
			Source:		MigrationSourceReporter,
			OriginalIndex:	i,
			Account:	formatAddress(rep.Account),
			Merged:		false,
			MergeReason:	"reporter_dropped_no_separate_reputation",
		})
	}

	state := ConsolidatedReputationState{
		Version:		ConsolidatedStateVersionV1,
		Params:			params,
		Identities:		normalizeIdentities(identities),
		ValidatorScores:	normalizeValidatorScores(validatorScores),
		ServiceTrustScores:	nil,
	}

	receipt := MigrationReceipt{
		Version:		uint32(ConsolidatedStateVersionV1),
		MigrationHeight:	migrationHeight,
		TotalAccounts:		len(old.Accounts),
		TotalStakeRecords:	len(old.StakeRecords),
		TotalValidators:	len(old.Validators),
		TotalReporters:		len(old.Reporters),
		IdentitiesCreated:	len(state.Identities) - identitiesMerged,
		IdentitiesMerged:	identitiesMerged,
		ValidatorScores:	len(state.ValidatorScores),
		ServiceTrustScores:	0,
		ReportersDropped:	reportersDropped,
		Entries:		entries,
	}
	receipt.DeterministicHash = ComputeMigrationReceiptHash(receipt)

	return state, receipt
}

func formatAddress(addr interface{}) string {
	switch v := addr.(type) {
	case string:
		return strings.TrimSpace(v)
	default:
		return fmt.Sprintf("%v", v)
	}
}
