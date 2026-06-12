package types

const (
	ModuleName	= "aetravalidatorscore"
	StoreKey	= ModuleName

	BasisPoints	uint32	= 10_000

	ConcentrationStatusNormal	= "normal"
	ConcentrationStatusNearCap	= "near_cap"
	ConcentrationStatusOverloaded	= "overloaded"

	EventTypeUpdateParams	= "aetra_validator_score_update_params"
	EventTypeUpdateScores	= "aetra_validator_score_update_scores"
)
