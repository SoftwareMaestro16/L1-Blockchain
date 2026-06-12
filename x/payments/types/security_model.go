package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

type SecurityThreat string

const (
	SecurityThreatStaleStateClose			SecurityThreat	= "STALE_STATE_CLOSE"
	SecurityThreatSameNonceDoubleSign		SecurityThreat	= "SAME_NONCE_DOUBLE_SIGN"
	SecurityThreatReplayAcrossDomain		SecurityThreat	= "REPLAY_ACROSS_CHAIN_OR_CHANNEL_DOMAIN"
	SecurityThreatInvalidConditionResolution	SecurityThreat	= "INVALID_CONDITION_RESOLUTION"
	SecurityThreatPreimageWithholding		SecurityThreat	= "PREIMAGE_WITHHOLDING"
	SecurityThreatTimeoutRace			SecurityThreat	= "TIMEOUT_RACE"
	SecurityThreatRouteGriefing			SecurityThreat	= "ROUTE_GRIEFING"
	SecurityThreatLiquidityExhaustion		SecurityThreat	= "LIQUIDITY_EXHAUSTION"
	SecurityThreatGossipSpam			SecurityThreat	= "GOSSIP_SPAM"
	SecurityThreatChannelOpenSpam			SecurityThreat	= "CHANNEL_OPEN_SPAM"
	SecurityThreatStateBloatUnresolvedPromises	SecurityThreat	= "STATE_BLOAT_UNRESOLVED_PROMISES"
	SecurityThreatWatchServiceDowntime		SecurityThreat	= "WATCH_SERVICE_DOWNTIME"
	SecurityThreatParticipantKeyCompromise		SecurityThreat	= "PARTICIPANT_KEY_COMPROMISE"
	SecurityThreatValidatorCensorship		SecurityThreat	= "VALIDATOR_CENSORSHIP_WITHIN_BOUNDED_WINDOWS"
	SecurityThreatSettlementBatchConflictAmplify	SecurityThreat	= "SETTLEMENT_BATCH_CONFLICT_AMPLIFICATION"
)

type SecurityGuarantee string

const (
	SecurityGuaranteeUnilateralClose		SecurityGuarantee	= "ANY_PARTICIPANT_CAN_UNILATERALLY_CLOSE"
	SecurityGuaranteeLatestStateSupersedesStale	SecurityGuarantee	= "LATEST_VALID_STATE_SUPERSEDES_STALE_CLOSE"
	SecurityGuaranteeSameNonceConflictPunishable	SecurityGuarantee	= "SAME_NONCE_CONFLICTING_SIGNATURES_PUNISHABLE"
	SecurityGuaranteeConditionalProofBeforeTimeout	SecurityGuarantee	= "CONDITIONS_SETTLE_ONLY_WITH_VALID_PROOF_BEFORE_TIMEOUT"
	SecurityGuaranteeExpiredConditionsRelease	SecurityGuarantee	= "EXPIRED_CONDITIONS_RELEASE_RESERVES"
	SecurityGuaranteeReplayRejected			SecurityGuarantee	= "REPLAY_REJECTED_ACROSS_DOMAINS_EPOCHS_AND_FINALIZED_STATES"
	SecurityGuaranteeLockedCollateral		SecurityGuarantee	= "LOCKED_COLLATERAL_WITHDRAWN_ONLY_BY_SETTLEMENT_RULES"
	SecurityGuaranteeOnchainDisputeSufficiency	SecurityGuarantee	= "ONCHAIN_STATE_SUFFICIENT_TO_RESOLVE_DISPUTES"
)

type ThreatModelEntry struct {
	Threat			SecurityThreat
	Controls		[]SecurityGuarantee
	ConsensusCritical	bool
	BoundedWindow		bool
	Description		string
}

type SecurityGuaranteeCheck struct {
	Guarantee	SecurityGuarantee
	Passed		bool
	EvidenceHash	string
	Reason		string
}

type SecurityModelReport struct {
	Threats		[]ThreatModelEntry
	Guarantees	[]SecurityGuaranteeCheck
	ReportHash	string
}

func DefaultThreatModel() []ThreatModelEntry {
	return []ThreatModelEntry{
		securityThreat(SecurityThreatStaleStateClose, true, true, "stale signed state submitted for unilateral close", SecurityGuaranteeLatestStateSupersedesStale, SecurityGuaranteeUnilateralClose),
		securityThreat(SecurityThreatSameNonceDoubleSign, true, false, "conflicting states signed at the same channel nonce", SecurityGuaranteeSameNonceConflictPunishable),
		securityThreat(SecurityThreatReplayAcrossDomain, true, false, "state or condition replayed across chain, channel, epoch, or tombstone domains", SecurityGuaranteeReplayRejected),
		securityThreat(SecurityThreatInvalidConditionResolution, true, true, "conditional payment resolved with invalid proof or preimage", SecurityGuaranteeConditionalProofBeforeTimeout),
		securityThreat(SecurityThreatPreimageWithholding, true, true, "receiver or downstream hop withholds the preimage until timeout pressure", SecurityGuaranteeExpiredConditionsRelease, SecurityGuaranteeConditionalProofBeforeTimeout),
		securityThreat(SecurityThreatTimeoutRace, true, true, "linked promise timeout ordering leaves an intermediary exposed", SecurityGuaranteeConditionalProofBeforeTimeout, SecurityGuaranteeExpiredConditionsRelease),
		securityThreat(SecurityThreatRouteGriefing, false, true, "off-chain route attempts consume liquidity or retry budget without consensus evidence", SecurityGuaranteeOnchainDisputeSufficiency),
		securityThreat(SecurityThreatLiquidityExhaustion, false, true, "available reserves are exhausted by unresolved or invalid promises", SecurityGuaranteeLockedCollateral, SecurityGuaranteeExpiredConditionsRelease),
		securityThreat(SecurityThreatGossipSpam, false, true, "untrusted topology advertisements attempt to influence settlement", SecurityGuaranteeOnchainDisputeSufficiency, SecurityGuaranteeReplayRejected),
		securityThreat(SecurityThreatChannelOpenSpam, true, true, "small channels attempt to bloat on-chain payment state", SecurityGuaranteeLockedCollateral, SecurityGuaranteeOnchainDisputeSufficiency),
		securityThreat(SecurityThreatStateBloatUnresolvedPromises, true, true, "unresolved promises accumulate beyond bounded channel state", SecurityGuaranteeExpiredConditionsRelease, SecurityGuaranteeLockedCollateral),
		securityThreat(SecurityThreatWatchServiceDowntime, false, true, "watch service misses a stale close during a challenge window", SecurityGuaranteeUnilateralClose, SecurityGuaranteeLatestStateSupersedesStale),
		securityThreat(SecurityThreatParticipantKeyCompromise, true, false, "compromised participant key signs conflicting or replayed state", SecurityGuaranteeSameNonceConflictPunishable, SecurityGuaranteeReplayRejected),
		securityThreat(SecurityThreatValidatorCensorship, true, true, "settlement transactions are censored within bounded dispute windows", SecurityGuaranteeLatestStateSupersedesStale, SecurityGuaranteeOnchainDisputeSufficiency),
		securityThreat(SecurityThreatSettlementBatchConflictAmplify, true, true, "batch settlement is shaped to amplify same-key write conflicts", SecurityGuaranteeLockedCollateral, SecurityGuaranteeOnchainDisputeSufficiency),
	}
}

func ValidateThreatModelCoverage(threats []ThreatModelEntry) error {
	required := requiredSecurityThreats()
	seen := make(map[SecurityThreat]struct{}, len(required))
	for _, entry := range threats {
		entry = entry.Normalize()
		if !IsSecurityThreat(entry.Threat) {
			return fmt.Errorf("unknown payments security threat %q", entry.Threat)
		}
		if _, duplicate := seen[entry.Threat]; duplicate {
			return fmt.Errorf("duplicate payments security threat %q", entry.Threat)
		}
		if len(entry.Controls) == 0 {
			return fmt.Errorf("payments security threat %q has no controls", entry.Threat)
		}
		for _, control := range entry.Controls {
			if !IsSecurityGuarantee(control) {
				return fmt.Errorf("payments security threat %q references unknown guarantee %q", entry.Threat, control)
			}
		}
		seen[entry.Threat] = struct{}{}
	}
	for _, threat := range required {
		if _, found := seen[threat]; !found {
			return fmt.Errorf("missing payments security threat %q", threat)
		}
	}
	return nil
}

func BuildSecurityModelReport(state PaymentsState) (SecurityModelReport, error) {
	state = state.Export()
	threats := DefaultThreatModel()
	if err := ValidateThreatModelCoverage(threats); err != nil {
		return SecurityModelReport{}, err
	}
	checks := []SecurityGuaranteeCheck{
		buildSecurityCheck(state, SecurityGuaranteeUnilateralClose, securityUnilateralClosePasses(state)),
		buildSecurityCheck(state, SecurityGuaranteeLatestStateSupersedesStale, securityLatestStateSupersedesStalePasses(state)),
		buildSecurityCheck(state, SecurityGuaranteeSameNonceConflictPunishable, securitySameNonceConflictPunishablePasses()),
		buildSecurityCheck(state, SecurityGuaranteeConditionalProofBeforeTimeout, securityConditionalProofBeforeTimeoutPasses(state)),
		buildSecurityCheck(state, SecurityGuaranteeExpiredConditionsRelease, securityExpiredConditionsReleasePasses(state)),
		buildSecurityCheck(state, SecurityGuaranteeReplayRejected, securityReplayRejectedPasses(state)),
		buildSecurityCheck(state, SecurityGuaranteeLockedCollateral, ValidateLockedCollateralForFinality(state) == nil),
		buildSecurityCheck(state, SecurityGuaranteeOnchainDisputeSufficiency, securityOnchainDisputeSufficiencyPasses(state)),
	}
	report := SecurityModelReport{
		Threats:	normalizeThreatModelEntries(threats),
		Guarantees:	normalizeSecurityGuaranteeChecks(checks),
	}
	report.ReportHash = ComputeSecurityModelReportHash(report)
	if err := report.Validate(); err != nil {
		return SecurityModelReport{}, err
	}
	return report, nil
}

func ComputeSecurityModelReportHash(report SecurityModelReport) string {
	report.Threats = normalizeThreatModelEntries(report.Threats)
	report.Guarantees = normalizeSecurityGuaranteeChecks(report.Guarantees)
	parts := []string{"payments-security-model-v1"}
	for _, threat := range report.Threats {
		parts = append(parts, string(threat.Threat), fmt.Sprintf("%t", threat.ConsensusCritical), fmt.Sprintf("%t", threat.BoundedWindow))
		for _, control := range threat.Controls {
			parts = append(parts, string(control))
		}
	}
	for _, check := range report.Guarantees {
		parts = append(parts, string(check.Guarantee), fmt.Sprintf("%t", check.Passed), normalizeHash(check.EvidenceHash), check.Reason)
	}
	return HashParts(parts...)
}

func (report SecurityModelReport) Validate() error {
	report.Threats = normalizeThreatModelEntries(report.Threats)
	report.Guarantees = normalizeSecurityGuaranteeChecks(report.Guarantees)
	if err := ValidateThreatModelCoverage(report.Threats); err != nil {
		return err
	}
	required := requiredSecurityGuarantees()
	seen := make(map[SecurityGuarantee]struct{}, len(required))
	for _, check := range report.Guarantees {
		check = check.Normalize()
		if !IsSecurityGuarantee(check.Guarantee) {
			return fmt.Errorf("unknown payments security guarantee %q", check.Guarantee)
		}
		if _, duplicate := seen[check.Guarantee]; duplicate {
			return fmt.Errorf("duplicate payments security guarantee %q", check.Guarantee)
		}
		if !check.Passed {
			return fmt.Errorf("payments security guarantee %q failed: %s", check.Guarantee, check.Reason)
		}
		if err := ValidateHash("payments security guarantee evidence hash", check.EvidenceHash); err != nil {
			return err
		}
		seen[check.Guarantee] = struct{}{}
	}
	for _, guarantee := range required {
		if _, found := seen[guarantee]; !found {
			return fmt.Errorf("missing payments security guarantee %q", guarantee)
		}
	}
	if err := ValidateHash("payments security model report hash", report.ReportHash); err != nil {
		return err
	}
	if expected := ComputeSecurityModelReportHash(report); report.ReportHash != expected {
		return errors.New("payments security model report hash mismatch")
	}
	return nil
}

func (entry ThreatModelEntry) Normalize() ThreatModelEntry {
	entry.Description = strings.TrimSpace(entry.Description)
	entry.Controls = normalizeSecurityGuarantees(entry.Controls)
	return entry
}

func (check SecurityGuaranteeCheck) Normalize() SecurityGuaranteeCheck {
	check.EvidenceHash = normalizeHash(check.EvidenceHash)
	check.Reason = strings.TrimSpace(check.Reason)
	return check
}

func IsSecurityThreat(value SecurityThreat) bool {
	for _, threat := range requiredSecurityThreats() {
		if value == threat {
			return true
		}
	}
	return false
}

func IsSecurityGuarantee(value SecurityGuarantee) bool {
	for _, guarantee := range requiredSecurityGuarantees() {
		if value == guarantee {
			return true
		}
	}
	return false
}

func securityThreat(threat SecurityThreat, consensusCritical, boundedWindow bool, description string, controls ...SecurityGuarantee) ThreatModelEntry {
	return ThreatModelEntry{
		Threat:			threat,
		Controls:		normalizeSecurityGuarantees(controls),
		ConsensusCritical:	consensusCritical,
		BoundedWindow:		boundedWindow,
		Description:		description,
	}
}

func buildSecurityCheck(state PaymentsState, guarantee SecurityGuarantee, passed bool) SecurityGuaranteeCheck {
	reason := "verified"
	if !passed {
		reason = "state violates guarantee"
	}
	return SecurityGuaranteeCheck{
		Guarantee:	guarantee,
		Passed:		passed,
		EvidenceHash:	securityGuaranteeEvidenceHash(state, guarantee),
		Reason:		reason,
	}
}

func securityUnilateralClosePasses(state PaymentsState) bool {
	for _, channel := range state.Channels {
		channel = channel.Normalize()
		if channel.Status == ChannelStatusOpen {
			if len(channel.Participants) == 0 || channel.DisputePeriod == 0 {
				return false
			}
		}
	}
	return true
}

func securityLatestStateSupersedesStalePasses(state PaymentsState) bool {
	for _, channel := range state.Channels {
		channel = channel.Normalize()
		if channel.Status != ChannelStatusPendingClose {
			continue
		}
		if channel.PendingClose.SettleAfterHeight <= channel.PendingClose.SubmittedHeight {
			return false
		}
		if channel.DisputedNonce > 0 && channel.DisputedNonce < channel.PendingClose.State.Nonce {
			return false
		}
	}
	return true
}

func securitySameNonceConflictPunishablePasses() bool {
	entry, err := PenaltyMatrixEntryForProof(FraudProofTypeDoubleSign, DefaultPenaltyMatrix())
	return err == nil && entry.Class == PenaltyClassDoubleSign
}

func securityConditionalProofBeforeTimeoutPasses(state PaymentsState) bool {
	for _, channel := range state.Channels {
		channel = channel.Normalize()
		if len(channel.LatestState.Conditions) > MaxConditionsPerState {
			return false
		}
		for _, condition := range channel.LatestState.Normalize().Conditions {
			if condition.ConditionType == ConditionTypeHashLock && condition.HashLock == "" {
				return false
			}
			if condition.TimeoutHeight == 0 {
				return false
			}
		}
	}
	return true
}

func securityExpiredConditionsReleasePasses(state PaymentsState) bool {
	if DefaultTimeoutMargin == 0 || MaxConditionsPerState == 0 {
		return false
	}
	for _, claim := range state.ConditionClaims {
		if normalizeHash(claim.ConditionID) == "" {
			return false
		}
	}
	return true
}

func securityReplayRejectedPasses(state PaymentsState) bool {
	for _, tombstone := range state.ClosedChannels {
		tombstone = tombstone.Normalize()
		if tombstone.ChannelID == "" || tombstone.FinalizedNonce == 0 || tombstone.ExpiresHeight == 0 {
			return false
		}
		if tombstone.ExpiresHeight <= tombstone.ClosedHeight {
			return false
		}
	}
	return true
}

func securityOnchainDisputeSufficiencyPasses(state PaymentsState) bool {
	for _, channel := range state.Channels {
		channel = channel.Normalize()
		if channel.Status != ChannelStatusPendingClose {
			continue
		}
		if channel.PendingClose.State.StateHash == "" || len(channel.PendingClose.State.Signatures) == 0 {
			return false
		}
	}
	return true
}

func securityGuaranteeEvidenceHash(state PaymentsState, guarantee SecurityGuarantee) string {
	parts := []string{"payments-security-guarantee-v1", string(guarantee)}
	for _, channel := range state.Channels {
		channel = channel.Normalize()
		parts = append(parts, channel.ChannelID, string(channel.Status), string(channel.Finality), channel.LatestState.StateHash, channel.PendingClose.State.StateHash)
	}
	for _, tombstone := range state.ClosedChannels {
		tombstone = tombstone.Normalize()
		parts = append(parts, tombstone.ChannelID, fmt.Sprintf("%d", tombstone.FinalizedNonce), fmt.Sprintf("%d", tombstone.ExpiresHeight))
	}
	for _, claim := range state.ConditionClaims {
		parts = append(parts, normalizeHash(claim.ChannelID), normalizeHash(claim.ConditionID), normalizeOptionalHash(claim.PreimageHash), normalizeOptionalHash(claim.EvidenceHash))
	}
	return HashParts(parts...)
}

func requiredSecurityThreats() []SecurityThreat {
	return []SecurityThreat{
		SecurityThreatStaleStateClose,
		SecurityThreatSameNonceDoubleSign,
		SecurityThreatReplayAcrossDomain,
		SecurityThreatInvalidConditionResolution,
		SecurityThreatPreimageWithholding,
		SecurityThreatTimeoutRace,
		SecurityThreatRouteGriefing,
		SecurityThreatLiquidityExhaustion,
		SecurityThreatGossipSpam,
		SecurityThreatChannelOpenSpam,
		SecurityThreatStateBloatUnresolvedPromises,
		SecurityThreatWatchServiceDowntime,
		SecurityThreatParticipantKeyCompromise,
		SecurityThreatValidatorCensorship,
		SecurityThreatSettlementBatchConflictAmplify,
	}
}

func requiredSecurityGuarantees() []SecurityGuarantee {
	return []SecurityGuarantee{
		SecurityGuaranteeUnilateralClose,
		SecurityGuaranteeLatestStateSupersedesStale,
		SecurityGuaranteeSameNonceConflictPunishable,
		SecurityGuaranteeConditionalProofBeforeTimeout,
		SecurityGuaranteeExpiredConditionsRelease,
		SecurityGuaranteeReplayRejected,
		SecurityGuaranteeLockedCollateral,
		SecurityGuaranteeOnchainDisputeSufficiency,
	}
}

func normalizeThreatModelEntries(entries []ThreatModelEntry) []ThreatModelEntry {
	out := make([]ThreatModelEntry, len(entries))
	for i, entry := range entries {
		out[i] = entry.Normalize()
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Threat < out[j].Threat })
	return out
}

func normalizeSecurityGuaranteeChecks(checks []SecurityGuaranteeCheck) []SecurityGuaranteeCheck {
	out := make([]SecurityGuaranteeCheck, len(checks))
	for i, check := range checks {
		out[i] = check.Normalize()
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Guarantee < out[j].Guarantee })
	return out
}

func normalizeSecurityGuarantees(guarantees []SecurityGuarantee) []SecurityGuarantee {
	out := make([]SecurityGuarantee, len(guarantees))
	copy(out, guarantees)
	sort.SliceStable(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}
