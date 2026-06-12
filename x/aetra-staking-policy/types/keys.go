package types

const (
	ModuleName	= "aetrastakingpolicy"
	StoreKey	= ModuleName

	BasisPoints	uint32	= 10_000

	PhaseOneValidatorSetMax	uint32	= 150
	PhaseTwoValidatorSetMax	uint32	= 250

	PhaseOnePowerCapBps	uint32	= 300
	PhaseTwoPowerCapBps	uint32	= 250
	MatureSetPowerCapBps	uint32	= 200

	Top10ConcentrationTargetBps	uint32	= 2_500
	Top20ConcentrationTargetBps	uint32	= 4_000
	Top33ConcentrationTargetBps	uint32	= 5_000

	DelegationWarningNone		= "none"
	DelegationWarningNearCap	= "near_cap"
	DelegationWarningOverloaded	= "overloaded"

	EventTypeUpdateParams		= "aetra_staking_policy_update_params"
	EventTypeRegisterIdentity	= "aetra_staking_policy_register_identity"
	EventTypeAcknowledgeWarning	= "aetra_staking_policy_acknowledge_warning"
)
