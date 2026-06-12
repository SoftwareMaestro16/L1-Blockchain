package types

import appparams "github.com/sovereign-l1/l1/app/params"

const (
	ModuleName	= "feecollector"
	StoreKey	= ModuleName
	RouterKey	= ModuleName

	CollectorModuleName	= ModuleName
	TreasuryModuleName	= "feecollector_treasury"
	ProtectionModuleName	= "feecollector_protection"

	ValidatorInsuranceModuleName	= "feecollector_validator_insurance"
	EcosystemGrantsModuleName	= "feecollector_ecosystem_grants"
	StorageRentReserveModuleName	= "feecollector_storage_rent_reserve"
	BurnModuleName			= "feecollector_burn"
	ReporterRewardsModuleName	= "feecollector_reporter_rewards"
)

var (
	ParamsKey		= []byte{0x01}
	FeeBalancesKey		= []byte{0x02}
	PendingDistributionKey	= []byte{0x03}
	FeeHistoryPrefix	= []byte{0x04}
	ProtocolIncomePolicyKey	= []byte{0x05}
)

const (
	BaseDenom		= appparams.BaseDenom
	BasisPoints	uint32	= 10_000

	FeeTypeGas		= "gas"
	FeeTypeForwarding	= "forwarding"
	FeeTypeProtocol		= "protocol"
)
