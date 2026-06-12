package params

import (
	"fmt"
	"time"

	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
)

const (
	SlashingEvidenceStandardCosmos	= "cosmos_sdk_x_slashing_x_evidence"

	DoubleSignSlashMinBps		= int64(500)
	DoubleSignSlashMaxBps		= int64(1_000)
	DoubleSignSlashDefaultBps	= DoubleSignSlashMinBps

	DowntimeFirstSlashMinBps	= int64(5)
	DowntimeFirstSlashMaxBps	= int64(10)
	DowntimeFirstSlashDefaultBps	= DowntimeFirstSlashMinBps

	DowntimeRepeatSlashMinBps	= int64(25)
	DowntimeRepeatSlashMaxBps	= int64(50)
	DowntimeRepeatSlashDefaultBps	= DowntimeRepeatSlashMinBps

	DowntimeChronicSlashMaxBps	= int64(100)
	DowntimeChronicSlashDefaultBps	= DowntimeChronicSlashMaxBps

	DowntimeFirstJailMinMinutes	= int64(60)
	DowntimeFirstJailMaxMinutes	= int64(360)
	DowntimeFirstJailDefaultMinutes	= DowntimeFirstJailMinMinutes
	DowntimeRepeatJailMinutes	= int64(24 * 60)
	DowntimeChronicJailMinutes	= int64(72 * 60)

	RepeatedInvalidProposalSlashBps		= int64(25)
	RepeatedInvalidProposalJailMinutes	= int64(24 * 60)
	RepeatedTimestampViolationSlashBps	= int64(25)
	RepeatedTimestampViolationJailMinutes	= int64(24 * 60)
	TimestampMaxForwardDriftSeconds		= int64(120)

	HeightEvidenceMaxAgeBlocks		= uint64(100_000)
	HeightUnbondingEvidenceWindowBlocks	= uint64(30_000)

	DowntimeOffenseCleanDecayDays		= int64(30)
	DowntimeOffenseMaxPenaltyBps		= DowntimeChronicSlashMaxBps
	DowntimeOffenseMaxJailMinutes		= DowntimeChronicJailMinutes
	DowntimeOffenseQueryStatusMethod	= "QueryDowntimeOffenseStatus"

	InvalidProposalMaxBytesDefault	= uint64(2 * 1024 * 1024)
)

type SlashingAccountabilityPolicy struct {
	EvidenceStandard			string
	ObjectiveCryptographicEvidenceOnly	bool
	SubjectiveSlashingAllowed		bool
	DoubleSignSlashBps			int64
	DoubleSignJailImmediate			bool
	DoubleSignPermanentTombstone		bool
	ConsensusKeyReuseForbidden		bool
	UsesCosmosSlashingAndEvidence		bool
	BaseFaultsUseCometBFTEvidence		bool
	StandardDoubleSignIntegrated		bool
	StandardLivenessDowntimeIntegrated	bool
	StandardTombstoneIntegrated		bool
	StandardJailUnjailIntegrated		bool
	CustomLogicWrapsStandardOnly		bool
	CoreSlashingForkForbidden		bool
	ProgressiveDowntimeEnabled		bool
	StandardDowntimeStatePreserved		bool
	CustomDowntimeOverlayRequired		bool
	DowntimeOffenseTracksValidatorConsAddr	bool
	DowntimeOffenseTracksOffenseCount	bool
	DowntimeOffenseTracksFirstOffenseTime	bool
	DowntimeOffenseTracksLastOffenseTime	bool
	DowntimeOffenseTracksLastSlashFraction	bool
	DowntimeOffenseTracksCurrentJail	bool
	DowntimeOffenseCleanDecayDays		int64
	DowntimeOffenseMaxPenaltyBps		int64
	DowntimeOffenseMaxJailMinutes		int64
	DowntimeOffenseDelegatorRiskInherited	bool
	DowntimeOffenseQueryStatusEnabled	bool
	DowntimeOffenseUnjailKeepsHistory	bool
	DowntimeFirstSlashBps			int64
	DowntimeFirstJailMinutes		int64
	DowntimeRepeatSlashBps			int64
	DowntimeRepeatJailMinutes		int64
	DowntimeChronicSlashBps			int64
	DowntimeChronicJailMinutes		int64
	DowntimeGovernanceReputationFlag	bool
	InvalidProposalDeterministicReject	bool
	InvalidProposalAutoSlash		bool
	InvalidProposalRepeatEvidenceOnly	bool
	ProcessProposalExternalInputs		bool
	ProcessProposalTestsRequired		bool
	RepeatedInvalidProposalSlashBps		int64
	RepeatedInvalidProposalJailMinutes	int64
	TimestampRejectOutsideBounds		bool
	TimestampCometBFTCompatible		bool
	TimestampCustomWallClockLogic		bool
	TimestampSlashObjectiveEvidenceOnly	bool
	TimestampRepeatedViolationsSlashBps	int64
	TimestampRepeatedViolationsJailMinutes	int64
	TimestampMaxForwardDriftSeconds		int64
	HeightConsensusControlled		bool
	SingleValidatorHeightControlForbidden	bool
	SameHeightDoubleSignCovered		bool
	EquivocationCovered			bool
	InvalidProposalHeightChecked		bool
	NonDeterministicAppValidationForbidden	bool
	EvidenceExpirationChecked		bool
	UnbondingEvidenceTimingChecked		bool
	HeightEvidenceMaxAgeBlocks		uint64
	HeightUnbondingEvidenceWindowBlocks	uint64
	EvidenceWhileValidatorBondedTest	bool
	EvidenceWhileValidatorUnbondingTest	bool
	EvidenceAfterUnbondingBeforeExpiryTest	bool
	EvidenceAfterExpirationRejectedTest	bool
	DelegatorInfractionHeightSlashTest	bool
	TombstoneCapBehaviorTest		bool
	InvalidTxProposalRejectedTest		bool
	OversizedProposalRejectedTest		bool
	MalformedSpecialTxRejectedTest		bool
	ValidProposalAcceptedTest		bool
	AllValidatorsProposalDeterminismTest	bool
	InvalidProposalWallClockForbidden	bool
	InvalidProposalExternalAPIsForbidden	bool
	ProcessProposalFragilityForbidden	bool
	InvalidProposalMaxBytes			uint64
}

func DefaultSlashingAccountabilityPolicy() SlashingAccountabilityPolicy {
	return SlashingAccountabilityPolicy{
		EvidenceStandard:			SlashingEvidenceStandardCosmos,
		ObjectiveCryptographicEvidenceOnly:	true,
		SubjectiveSlashingAllowed:		false,
		DoubleSignSlashBps:			DoubleSignSlashDefaultBps,
		DoubleSignJailImmediate:		true,
		DoubleSignPermanentTombstone:		true,
		ConsensusKeyReuseForbidden:		true,
		UsesCosmosSlashingAndEvidence:		true,
		BaseFaultsUseCometBFTEvidence:		true,
		StandardDoubleSignIntegrated:		true,
		StandardLivenessDowntimeIntegrated:	true,
		StandardTombstoneIntegrated:		true,
		StandardJailUnjailIntegrated:		true,
		CustomLogicWrapsStandardOnly:		true,
		CoreSlashingForkForbidden:		true,
		ProgressiveDowntimeEnabled:		true,
		StandardDowntimeStatePreserved:		true,
		CustomDowntimeOverlayRequired:		true,
		DowntimeOffenseTracksValidatorConsAddr:	true,
		DowntimeOffenseTracksOffenseCount:	true,
		DowntimeOffenseTracksFirstOffenseTime:	true,
		DowntimeOffenseTracksLastOffenseTime:	true,
		DowntimeOffenseTracksLastSlashFraction:	true,
		DowntimeOffenseTracksCurrentJail:	true,
		DowntimeOffenseCleanDecayDays:		DowntimeOffenseCleanDecayDays,
		DowntimeOffenseMaxPenaltyBps:		DowntimeOffenseMaxPenaltyBps,
		DowntimeOffenseMaxJailMinutes:		DowntimeOffenseMaxJailMinutes,
		DowntimeOffenseDelegatorRiskInherited:	true,
		DowntimeOffenseQueryStatusEnabled:	true,
		DowntimeOffenseUnjailKeepsHistory:	true,
		DowntimeFirstSlashBps:			DowntimeFirstSlashDefaultBps,
		DowntimeFirstJailMinutes:		DowntimeFirstJailDefaultMinutes,
		DowntimeRepeatSlashBps:			DowntimeRepeatSlashDefaultBps,
		DowntimeRepeatJailMinutes:		DowntimeRepeatJailMinutes,
		DowntimeChronicSlashBps:		DowntimeChronicSlashDefaultBps,
		DowntimeChronicJailMinutes:		DowntimeChronicJailMinutes,
		DowntimeGovernanceReputationFlag:	true,
		InvalidProposalDeterministicReject:	true,
		InvalidProposalAutoSlash:		false,
		InvalidProposalRepeatEvidenceOnly:	true,
		ProcessProposalExternalInputs:		false,
		ProcessProposalTestsRequired:		true,
		RepeatedInvalidProposalSlashBps:	RepeatedInvalidProposalSlashBps,
		RepeatedInvalidProposalJailMinutes:	RepeatedInvalidProposalJailMinutes,
		TimestampRejectOutsideBounds:		true,
		TimestampCometBFTCompatible:		true,
		TimestampCustomWallClockLogic:		false,
		TimestampSlashObjectiveEvidenceOnly:	true,
		TimestampRepeatedViolationsSlashBps:	RepeatedTimestampViolationSlashBps,
		TimestampRepeatedViolationsJailMinutes:	RepeatedTimestampViolationJailMinutes,
		TimestampMaxForwardDriftSeconds:	TimestampMaxForwardDriftSeconds,
		HeightConsensusControlled:		true,
		SingleValidatorHeightControlForbidden:	true,
		SameHeightDoubleSignCovered:		true,
		EquivocationCovered:			true,
		InvalidProposalHeightChecked:		true,
		NonDeterministicAppValidationForbidden:	true,
		EvidenceExpirationChecked:		true,
		UnbondingEvidenceTimingChecked:		true,
		HeightEvidenceMaxAgeBlocks:		HeightEvidenceMaxAgeBlocks,
		HeightUnbondingEvidenceWindowBlocks:	HeightUnbondingEvidenceWindowBlocks,
		EvidenceWhileValidatorBondedTest:	true,
		EvidenceWhileValidatorUnbondingTest:	true,
		EvidenceAfterUnbondingBeforeExpiryTest:	true,
		EvidenceAfterExpirationRejectedTest:	true,
		DelegatorInfractionHeightSlashTest:	true,
		TombstoneCapBehaviorTest:		true,
		InvalidTxProposalRejectedTest:		true,
		OversizedProposalRejectedTest:		true,
		MalformedSpecialTxRejectedTest:		true,
		ValidProposalAcceptedTest:		true,
		AllValidatorsProposalDeterminismTest:	true,
		InvalidProposalWallClockForbidden:	true,
		InvalidProposalExternalAPIsForbidden:	true,
		ProcessProposalFragilityForbidden:	true,
		InvalidProposalMaxBytes:		InvalidProposalMaxBytesDefault,
	}
}

func AetraSlashingParams() slashingtypes.Params {
	params := slashingtypes.DefaultParams()
	params.SlashFractionDoubleSign = BpsToLegacyDec(DoubleSignSlashDefaultBps)
	params.SlashFractionDowntime = BpsToLegacyDec(DowntimeFirstSlashDefaultBps)
	params.DowntimeJailDuration = time.Duration(DowntimeFirstJailDefaultMinutes) * time.Minute
	return params
}

func (p SlashingAccountabilityPolicy) Validate() error {
	if p.EvidenceStandard != SlashingEvidenceStandardCosmos {
		return fmt.Errorf("slashing evidence standard must use Cosmos SDK x/slashing and x/evidence")
	}
	if !p.ObjectiveCryptographicEvidenceOnly {
		return fmt.Errorf("slashing must require objective cryptographic evidence")
	}
	if p.SubjectiveSlashingAllowed {
		return fmt.Errorf("subjective slashing must not be enabled")
	}
	if err := validateSlashingBpsValue("double_sign_slash", p.DoubleSignSlashBps, DoubleSignSlashMinBps, DoubleSignSlashMaxBps); err != nil {
		return err
	}
	if !p.DoubleSignJailImmediate {
		return fmt.Errorf("double-sign evidence must jail immediately")
	}
	if !p.DoubleSignPermanentTombstone {
		return fmt.Errorf("double-sign evidence must permanently tombstone the validator")
	}
	if !p.ConsensusKeyReuseForbidden {
		return fmt.Errorf("double-sign tombstone must forbid consensus key reuse")
	}
	if !p.UsesCosmosSlashingAndEvidence {
		return fmt.Errorf("slashing policy must use Cosmos SDK slashing and evidence modules")
	}
	if !p.BaseFaultsUseCometBFTEvidence {
		return fmt.Errorf("base slashing faults must use CometBFT evidence")
	}
	if !p.StandardDoubleSignIntegrated {
		return fmt.Errorf("standard double-sign handling must integrate x/slashing")
	}
	if !p.StandardLivenessDowntimeIntegrated {
		return fmt.Errorf("standard liveness/downtime handling must integrate x/slashing")
	}
	if !p.StandardTombstoneIntegrated {
		return fmt.Errorf("standard tombstone handling must integrate x/slashing")
	}
	if !p.StandardJailUnjailIntegrated {
		return fmt.Errorf("standard jail/unjail handling must integrate x/slashing")
	}
	if !p.CustomLogicWrapsStandardOnly {
		return fmt.Errorf("custom slashing logic must wrap or extend standard behavior only where necessary")
	}
	if !p.CoreSlashingForkForbidden {
		return fmt.Errorf("core x/slashing logic must not be forked unless no safer option exists")
	}
	if !p.ProgressiveDowntimeEnabled {
		return fmt.Errorf("progressive downtime penalties must be enabled")
	}
	if !p.StandardDowntimeStatePreserved {
		return fmt.Errorf("progressive downtime must preserve standard x/slashing signing state")
	}
	if !p.CustomDowntimeOverlayRequired {
		return fmt.Errorf("progressive downtime requires custom overlay when x/slashing is insufficient")
	}
	if !p.DowntimeOffenseTracksValidatorConsAddr {
		return fmt.Errorf("DowntimeOffense must track ValidatorConsAddr")
	}
	if !p.DowntimeOffenseTracksOffenseCount {
		return fmt.Errorf("DowntimeOffense must track OffenseCount")
	}
	if !p.DowntimeOffenseTracksFirstOffenseTime {
		return fmt.Errorf("DowntimeOffense must track FirstOffenseTime")
	}
	if !p.DowntimeOffenseTracksLastOffenseTime {
		return fmt.Errorf("DowntimeOffense must track LastOffenseTime")
	}
	if !p.DowntimeOffenseTracksLastSlashFraction {
		return fmt.Errorf("DowntimeOffense must track LastSlashFraction")
	}
	if !p.DowntimeOffenseTracksCurrentJail {
		return fmt.Errorf("DowntimeOffense must track CurrentJailDuration")
	}
	if p.DowntimeOffenseCleanDecayDays <= 0 {
		return fmt.Errorf("downtime offense count must decay after a positive clean period")
	}
	if p.DowntimeOffenseMaxPenaltyBps <= 0 || p.DowntimeOffenseMaxPenaltyBps > DowntimeChronicSlashMaxBps {
		return fmt.Errorf("downtime offense maximum penalty must be capped at <= 1 percent")
	}
	if p.DowntimeOffenseMaxJailMinutes < p.DowntimeRepeatJailMinutes {
		return fmt.Errorf("downtime offense maximum jail must be at least repeat jail duration")
	}
	if !p.DowntimeOffenseDelegatorRiskInherited {
		return fmt.Errorf("delegators must inherit validator downtime risk")
	}
	if !p.DowntimeOffenseQueryStatusEnabled {
		return fmt.Errorf("validator downtime offense status query must be enabled")
	}
	if !p.DowntimeOffenseUnjailKeepsHistory {
		return fmt.Errorf("unjail must not erase downtime slash history immediately")
	}
	if err := validateSlashingBpsValue("downtime_first_slash", p.DowntimeFirstSlashBps, DowntimeFirstSlashMinBps, DowntimeFirstSlashMaxBps); err != nil {
		return err
	}
	if p.DowntimeFirstJailMinutes < DowntimeFirstJailMinMinutes || p.DowntimeFirstJailMinutes > DowntimeFirstJailMaxMinutes {
		return fmt.Errorf("downtime first jail must stay within 1-6 hours")
	}
	if err := validateSlashingBpsValue("downtime_repeat_slash", p.DowntimeRepeatSlashBps, DowntimeRepeatSlashMinBps, DowntimeRepeatSlashMaxBps); err != nil {
		return err
	}
	if p.DowntimeRepeatJailMinutes != DowntimeRepeatJailMinutes {
		return fmt.Errorf("downtime repeat jail must be 24 hours")
	}
	if p.DowntimeChronicSlashBps <= p.DowntimeRepeatSlashBps || p.DowntimeChronicSlashBps > DowntimeChronicSlashMaxBps {
		return fmt.Errorf("downtime chronic slash must be above repeat slash and <= 1 percent")
	}
	if p.DowntimeChronicJailMinutes <= p.DowntimeRepeatJailMinutes {
		return fmt.Errorf("downtime chronic jail must be longer than repeat jail")
	}
	if !p.DowntimeGovernanceReputationFlag {
		return fmt.Errorf("chronic downtime must expose governance or reputation flag")
	}
	if !p.InvalidProposalDeterministicReject {
		return fmt.Errorf("invalid proposals must be rejected deterministically")
	}
	if p.InvalidProposalAutoSlash {
		return fmt.Errorf("invalid proposals must not auto-slash without objective repeat evidence")
	}
	if !p.InvalidProposalRepeatEvidenceOnly {
		return fmt.Errorf("invalid proposal slashing requires repeated objective evidence")
	}
	if p.ProcessProposalExternalInputs {
		return fmt.Errorf("ProcessProposal must not use non-deterministic external inputs")
	}
	if !p.ProcessProposalTestsRequired {
		return fmt.Errorf("ProcessProposal deterministic accept/reject tests are required")
	}
	if err := validateSlashingBpsValue("repeated_invalid_proposal_slash", p.RepeatedInvalidProposalSlashBps, DowntimeRepeatSlashMinBps, DowntimeRepeatSlashMaxBps); err != nil {
		return err
	}
	if p.RepeatedInvalidProposalJailMinutes < DowntimeRepeatJailMinutes {
		return fmt.Errorf("repeated invalid proposal jail must be at least 24 hours")
	}
	if !p.TimestampRejectOutsideBounds {
		return fmt.Errorf("timestamp policy must reject blocks outside consensus/application bounds")
	}
	if !p.TimestampCometBFTCompatible {
		return fmt.Errorf("timestamp policy must remain CometBFT-compatible")
	}
	if p.TimestampCustomWallClockLogic {
		return fmt.Errorf("timestamp policy must not use custom wall-clock logic")
	}
	if !p.TimestampSlashObjectiveEvidenceOnly {
		return fmt.Errorf("timestamp slashing requires objective reproducible signed evidence")
	}
	if err := validateSlashingBpsValue("repeated_timestamp_violation_slash", p.TimestampRepeatedViolationsSlashBps, DowntimeRepeatSlashMinBps, DowntimeRepeatSlashMaxBps); err != nil {
		return err
	}
	if p.TimestampRepeatedViolationsJailMinutes < DowntimeRepeatJailMinutes {
		return fmt.Errorf("repeated timestamp violation jail must be at least 24 hours")
	}
	if p.TimestampMaxForwardDriftSeconds <= 0 {
		return fmt.Errorf("timestamp max forward drift must be positive")
	}
	if !p.HeightConsensusControlled {
		return fmt.Errorf("height must remain consensus-controlled")
	}
	if !p.SingleValidatorHeightControlForbidden {
		return fmt.Errorf("single validator height control must be forbidden")
	}
	if !p.SameHeightDoubleSignCovered {
		return fmt.Errorf("same-height double-sign must be covered")
	}
	if !p.EquivocationCovered {
		return fmt.Errorf("height manipulation policy must cover equivocation")
	}
	if !p.InvalidProposalHeightChecked {
		return fmt.Errorf("invalid proposal height must be checked deterministically")
	}
	if !p.NonDeterministicAppValidationForbidden {
		return fmt.Errorf("non-deterministic app validation must be forbidden")
	}
	if !p.EvidenceExpirationChecked {
		return fmt.Errorf("evidence expiration edge cases must be checked")
	}
	if !p.UnbondingEvidenceTimingChecked {
		return fmt.Errorf("unbonding and evidence timing must be checked")
	}
	if p.HeightEvidenceMaxAgeBlocks == 0 {
		return fmt.Errorf("height evidence max age must be positive")
	}
	if p.HeightUnbondingEvidenceWindowBlocks == 0 {
		return fmt.Errorf("height unbonding evidence window must be positive")
	}
	if p.HeightUnbondingEvidenceWindowBlocks > p.HeightEvidenceMaxAgeBlocks {
		return fmt.Errorf("unbonding evidence window must not exceed evidence max age")
	}
	if !p.EvidenceWhileValidatorBondedTest {
		return fmt.Errorf("tests must cover evidence submitted while validator bonded")
	}
	if !p.EvidenceWhileValidatorUnbondingTest {
		return fmt.Errorf("tests must cover evidence submitted while validator unbonding")
	}
	if !p.EvidenceAfterUnbondingBeforeExpiryTest {
		return fmt.Errorf("tests must cover evidence after unbonding before evidence expiration")
	}
	if !p.EvidenceAfterExpirationRejectedTest {
		return fmt.Errorf("tests must cover evidence submitted after expiration")
	}
	if !p.DelegatorInfractionHeightSlashTest {
		return fmt.Errorf("tests must cover slashing delegators bonded at infraction height")
	}
	if !p.TombstoneCapBehaviorTest {
		return fmt.Errorf("tests must cover tombstone cap behavior")
	}
	if !p.InvalidTxProposalRejectedTest {
		return fmt.Errorf("tests must cover invalid tx proposal rejection")
	}
	if !p.OversizedProposalRejectedTest {
		return fmt.Errorf("tests must cover oversized proposal rejection")
	}
	if !p.MalformedSpecialTxRejectedTest {
		return fmt.Errorf("tests must cover malformed special tx rejection")
	}
	if !p.ValidProposalAcceptedTest {
		return fmt.Errorf("tests must cover valid proposal acceptance")
	}
	if !p.AllValidatorsProposalDeterminismTest {
		return fmt.Errorf("tests must cover identical proposal decisions across validators")
	}
	if !p.InvalidProposalWallClockForbidden {
		return fmt.Errorf("invalid proposal handling must not depend on local wall clock")
	}
	if !p.InvalidProposalExternalAPIsForbidden {
		return fmt.Errorf("invalid proposal handling must not depend on external APIs")
	}
	if !p.ProcessProposalFragilityForbidden {
		return fmt.Errorf("ProcessProposal must not be fragile")
	}
	if p.InvalidProposalMaxBytes == 0 {
		return fmt.Errorf("invalid proposal max bytes must be positive")
	}
	return nil
}

func validateSlashingBpsValue(name string, value, allowedMin, allowedMax int64) error {
	if value < allowedMin || value > allowedMax {
		return fmt.Errorf("%s must stay within %d-%d bps", name, allowedMin, allowedMax)
	}
	return nil
}
