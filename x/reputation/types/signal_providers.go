package types

import (
	"fmt"
)

const (
	SignalTypeStakeTime		= "stake_time"
	SignalTypeTxSuccess		= "tx_success"
	SignalTypeTxFailure		= "tx_failure"
	SignalTypeTxSpam		= "tx_spam"
	SignalTypeContractOk		= "contract_ok"
	SignalTypeContractFail		= "contract_fail"
	SignalTypeRecovery		= "recovery"
	SignalTypeFreeze		= "freeze"
	SignalTypeUnfreeze		= "unfreeze"
	SignalTypeValidatorUptime	= "validator_uptime"
	SignalTypeValidatorMissed	= "validator_missed"
	SignalTypeValidatorSlashed	= "validator_slashed"
	SignalTypeValidatorCommission	= "validator_commission"
	SignalTypeServiceAvailable	= "service_available"
	SignalTypeServiceUnavailable	= "service_unavailable"
	SignalTypeServiceProofQuality	= "service_proof_quality"
	SignalTypeDomainRegistration	= "domain_registration"
)

type SignalProviderType string

const (
	SignalProviderUser	SignalProviderType	= "user"
	SignalProviderValidator	SignalProviderType	= "validator"
	SignalProviderService	SignalProviderType	= "service"
)

type ReputationSignal struct {
	ProviderType	SignalProviderType	`json:"provider_type"`
	ProviderID	string			`json:"provider_id"`
	SignalType	string			`json:"signal_type"`
	Height		uint64			`json:"height"`
	Amount		uint64			`json:"amount"`
	Metadata	string			`json:"metadata,omitempty"`
}

func (s ReputationSignal) Validate() error {
	if s.ProviderID == "" {
		return fmt.Errorf("reputation signal: provider_id must not be empty")
	}
	switch s.ProviderType {
	case SignalProviderUser, SignalProviderValidator, SignalProviderService:
	default:
		return fmt.Errorf("reputation signal: invalid provider_type %q", s.ProviderType)
	}
	switch s.SignalType {
	case SignalTypeStakeTime, SignalTypeTxSuccess, SignalTypeTxFailure,
		SignalTypeTxSpam, SignalTypeContractOk, SignalTypeContractFail,
		SignalTypeRecovery, SignalTypeFreeze, SignalTypeUnfreeze,
		SignalTypeDomainRegistration:
		if s.ProviderType != SignalProviderUser {
			return fmt.Errorf("reputation signal: signal_type %q requires provider_type user", s.SignalType)
		}
	case SignalTypeValidatorUptime, SignalTypeValidatorMissed,
		SignalTypeValidatorSlashed, SignalTypeValidatorCommission:
		if s.ProviderType != SignalProviderValidator {
			return fmt.Errorf("reputation signal: signal_type %q requires provider_type validator", s.SignalType)
		}
	case SignalTypeServiceAvailable, SignalTypeServiceUnavailable, SignalTypeServiceProofQuality:
		if s.ProviderType != SignalProviderService {
			return fmt.Errorf("reputation signal: signal_type %q requires provider_type service", s.SignalType)
		}
	default:
		return fmt.Errorf("reputation signal: unknown signal_type %q", s.SignalType)
	}
	return nil
}

func ApplyIdentitySignal(identity *IdentityReputation, signal ReputationSignal, effectParams ReputationEffectParams) (*IdentityReputation, error) {
	if identity == nil {
		return nil, fmt.Errorf("reputation signal: identity must not be nil")
	}
	if err := signal.Validate(); err != nil {
		return nil, err
	}

	oldScore := identity.Score
	oldConf := identity.Confidence

	switch signal.SignalType {
	case SignalTypeStakeTime:
		identity.RecordStakeTime(signal.Amount, signal.Height)
	case SignalTypeTxSuccess:
		identity.RecordSuccessfulTx(signal.Height)
	case SignalTypeTxFailure:
		identity.RecordFailedTx(signal.Height)
	case SignalTypeTxSpam:
		identity.RecordSpam(signal.Height)
	case SignalTypeContractOk:
		identity.RecordContractInteraction(signal.Height)
	case SignalTypeContractFail:
		identity.RecordContractFailure(signal.Height)
	case SignalTypeRecovery:
		identity.RecordRecoveryEvent(signal.Height)
	case SignalTypeFreeze:
		identity.RecordRecoveryEvent(signal.Height)
	case SignalTypeUnfreeze:
		identity.RecordRecoveryEvent(signal.Height)
	case SignalTypeDomainRegistration:
		identity.RecordDomainRegistration(signal.Height)
	default:
		return nil, fmt.Errorf("reputation signal: unsupported identity signal_type %q", signal.SignalType)
	}

	newScore := ComputeIdentityScore(identity)
	newConf := ComputeConfidence(identity)

	identity.Score = ApplyPerEpochScoreCap(oldScore, newScore, effectParams.PerEpochScoreCap)
	identity.Confidence = ApplyPerEpochConfidenceCap(oldConf, newConf, effectParams.PerEpochConfidenceCap)

	return identity, nil
}

func ApplyValidatorSignal(vs *ValidatorScore, signal ReputationSignal, effectParams ReputationEffectParams) (*ValidatorScore, error) {
	if vs == nil {
		return nil, fmt.Errorf("reputation signal: validator score must not be nil")
	}
	if err := signal.Validate(); err != nil {
		return nil, err
	}

	if vs.IsJailed || vs.IsSlashed {
		switch signal.SignalType {
		case SignalTypeValidatorSlashed:
			vs.SlashingPenalty += uint32(signal.Amount)
			vs.IsSlashed = true
		case SignalTypeValidatorMissed:
			vs.MissedBlocksPenalty += uint32(signal.Amount)
		default:
		}
		vs.TotalScore = ComputeValidatorTotalScore(vs)
		vs.LastUpdateHeight = signal.Height
		return vs, nil
	}

	switch signal.SignalType {
	case SignalTypeValidatorUptime:
		vs.UptimeScore += uint32(signal.Amount)
		if vs.UptimeScore > IdentityScoreMax {
			vs.UptimeScore = IdentityScoreMax
		}
	case SignalTypeValidatorMissed:
		vs.MissedBlocksPenalty += uint32(signal.Amount)
	case SignalTypeValidatorSlashed:
		vs.SlashingPenalty += uint32(signal.Amount)
		vs.IsSlashed = true
	case SignalTypeValidatorCommission:
		vs.CommissionBehavior += uint32(signal.Amount)
		if vs.CommissionBehavior > IdentityScoreMax {
			vs.CommissionBehavior = IdentityScoreMax
		}
	}

	vs.TotalScore = ComputeValidatorTotalScore(vs)
	vs.LastUpdateHeight = signal.Height
	return vs, nil
}

func ApplyServiceSignal(sts *ServiceTrustScore, signal ReputationSignal) (*ServiceTrustScore, error) {
	if sts == nil {
		return nil, fmt.Errorf("reputation signal: service trust score must not be nil")
	}
	if err := signal.Validate(); err != nil {
		return nil, err
	}

	switch signal.SignalType {
	case SignalTypeServiceAvailable:
		sts.Trust += uint32(signal.Amount)
		if sts.Trust > IdentityScoreMax {
			sts.Trust = IdentityScoreMax
		}
		sts.Reliability += uint32(signal.Amount) / 2
		if sts.Reliability > IdentityScoreMax {
			sts.Reliability = IdentityScoreMax
		}
	case SignalTypeServiceUnavailable:
		if sts.Trust > uint32(signal.Amount) {
			sts.Trust -= uint32(signal.Amount)
		} else {
			sts.Trust = 0
		}
		if sts.Reliability > uint32(signal.Amount) {
			sts.Reliability -= uint32(signal.Amount)
		} else {
			sts.Reliability = 0
		}
	case SignalTypeServiceProofQuality:
		sts.Reliability += uint32(signal.Amount)
		if sts.Reliability > IdentityScoreMax {
			sts.Reliability = IdentityScoreMax
		}
	}

	sts.LastUpdateHeight = signal.Height
	return sts, nil
}

func ValidateSignalRoutesToCallerNotContract(signal ReputationSignal) error {
	if signal.ProviderType == SignalProviderUser && signal.SignalType == SignalTypeContractFail {
		return nil
	}
	if signal.ProviderType == SignalProviderUser && signal.SignalType == SignalTypeContractOk {
		return nil
	}
	return nil
}

func ValidateNoContractReputationState(contractAddr string) error {
	return fmt.Errorf("contracts do not have reputation state: %s", contractAddr)
}
