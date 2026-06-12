package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPersistentStateRentAccruesForActiveContractAndLongLivedRecords(t *testing.T) {
	params := DefaultStorageRentParams()
	params.FreeStorageAllowance = 0
	params.RentRatePerByteSecond = 2

	fixtures := []PersistentStateRecord{
		{SubjectID: "contract-1", Class: StateClassContract, CodeBytes: 10, DataBytes: 20, IndexBytes: 5, Persistent: true, Status: ContractStatusActive, LastChargedHeight: 7},
		{SubjectID: "domain-1", Class: StateClassDomainRecord, DataBytes: 30, IndexBytes: 5, Persistent: true, Status: ContractStatusActive, LastChargedHeight: 7},
		{SubjectID: "rep-1", Class: StateClassStakingReputation, DataBytes: 30, IndexBytes: 5, Persistent: true, Status: ContractStatusActive, LastChargedHeight: 7},
	}

	for _, fixture := range fixtures {
		result, err := AccruePersistentStateRent(fixture, params, 9)
		require.NoError(t, err, fixture.SubjectID)
		require.Equal(t, uint64(30), result.StorageBytes, fixture.SubjectID)
		require.Equal(t, uint64(120), result.RentDelta, fixture.SubjectID)
		require.Equal(t, uint64(120), result.Subject.RentDebt, fixture.SubjectID)
		require.Equal(t, uint64(9), result.Subject.LastChargedHeight, fixture.SubjectID)
	}
}

func TestPersistentStorageSizeIsCodeBytesPlusDataBytesOnly(t *testing.T) {
	size, err := PersistentStorageSize(PersistentStateRecord{
		CodeBytes:	11,
		DataBytes:	17,
		IndexBytes:	99,
	})
	require.NoError(t, err)
	require.Equal(t, uint64(28), size)
}

func TestRentDeltaUsesRatePerByteSecondAndElapsedSeconds(t *testing.T) {
	params := DefaultStorageRentParams()
	params.FreeStorageAllowance = 0
	params.RentRatePerByteSecond = 3

	delta, err := RentForSeconds(7, params, 11)
	require.NoError(t, err)
	require.Equal(t, uint64(231), delta)
}

func TestPersistentStateRentSkipsUnactivatedEmptyAndDeletedState(t *testing.T) {
	params := DefaultStorageRentParams()
	params.FreeStorageAllowance = 0
	params.RentRatePerByteSecond = 10

	unactivated := PersistentStateRecord{SubjectID: "virtual-ae", Class: StateClassWallet, Persistent: false, LastChargedHeight: 1}
	result, err := AccruePersistentStateRent(unactivated, params, 50)
	require.NoError(t, err)
	require.Zero(t, result.RentDelta)
	require.Zero(t, result.Subject.RentDebt)

	empty := PersistentStateRecord{SubjectID: "empty", Class: StateClassContract, Persistent: true, Status: ContractStatusActive, LastChargedHeight: 1}
	result, err = AccruePersistentStateRent(empty, params, 50)
	require.NoError(t, err)
	require.Zero(t, result.RentDelta)
	require.Zero(t, result.Subject.RentDebt)

	deleted := PersistentStateRecord{SubjectID: "deleted", Class: StateClassContract, DataBytes: 100, Persistent: true, Status: ContractStatusDeleted, LastChargedHeight: 1}
	result, err = AccruePersistentStateRent(deleted, params, 50)
	require.NoError(t, err)
	require.Zero(t, result.RentDelta)
	require.Zero(t, result.Subject.RentDebt)
}

func TestZeroBalanceNoStateHasZeroRentDebtAndPersistentStateAccruesDebt(t *testing.T) {
	params := DefaultStorageRentParams()
	params.FreeStorageAllowance = 0
	params.RentRatePerByteSecond = 4

	noState := PersistentStateRecord{SubjectID: "empty", Class: StateClassWallet, Persistent: true, Status: ContractStatusActive, LastChargedHeight: 10}
	result, err := AccruePersistentStateRent(noState, params, 20)
	require.NoError(t, err)
	require.Zero(t, result.RentDelta)
	require.Zero(t, result.Subject.RentDebt)

	withState := noState
	withState.SubjectID = "wallet-with-state"
	withState.DataBytes = 5
	result, err = AccruePersistentStateRent(withState, params, 20)
	require.NoError(t, err)
	require.Equal(t, uint64(200), result.RentDelta)
	require.Equal(t, uint64(200), result.Subject.RentDebt)
}

func TestWalletEffectiveFeeIncludesGasRentAndUnpaidDebt(t *testing.T) {
	fee, err := EffectiveWalletFee(3, 5, 7)
	require.NoError(t, err)
	require.Equal(t, uint64(15), fee)

	_, err = EffectiveWalletFee(^uint64(0), 1, 0)
	require.ErrorContains(t, err, "overflow")
}

func TestPoolRentUsesProtocolReservesAndFrozenLimitedAllowsOnlyRecoveryActions(t *testing.T) {
	payer, remaining := PayPoolRent(PoolRentPayer{
		ProtocolFeeReserve:	30,
		GovernanceReserve:	15,
		UserFacingCharge:	99,
	}, 40)
	require.Zero(t, remaining)
	require.Equal(t, uint64(0), payer.ProtocolFeeReserve)
	require.Equal(t, uint64(5), payer.GovernanceReserve)
	require.Equal(t, uint64(99), payer.UserFacingCharge)

	pool := PersistentStateRecord{Class: StateClassPoolContract, OfficialPool: true, Status: ContractStatusActive}
	require.Equal(t, ContractStatusFrozenLimited, OfficialPoolStatusForDebt(pool, 1))
	require.False(t, FrozenLimitedPoolAllows(PoolActionDeposit))
	require.True(t, FrozenLimitedPoolAllows(PoolActionClaim))
	require.True(t, FrozenLimitedPoolAllows(PoolActionUnbond))
	require.True(t, FrozenLimitedPoolAllows(PoolActionMaturedWithdrawal))
	require.True(t, FrozenLimitedPoolAllows(PoolActionGovernanceRecovery))
}

func TestNormalAccountAndContractFreezeOnUnpaidDebt(t *testing.T) {
	wallet := PersistentStateRecord{Class: StateClassWallet, Status: ContractStatusActive}
	contract := PersistentStateRecord{Class: StateClassContract, Status: ContractStatusActive}

	require.Equal(t, ContractStatusFrozen, StatusAfterRentDebt(wallet, 1))
	require.Equal(t, ContractStatusFrozen, StatusAfterRentDebt(contract, 1))
	require.Equal(t, ContractStatusActive, StatusAfterRentDebt(contract, 0))
}

func TestProtocolCriticalSystemStateIsProtocolPaidAndCannotFreezeFromRent(t *testing.T) {
	params := DefaultStorageRentParams()
	params.FreeStorageAllowance = 0
	params.RentRatePerByteSecond = 1
	system := PersistentStateRecord{
		SubjectID:		"system/module",
		Class:			StateClassSystemModule,
		DataBytes:		10,
		Persistent:		true,
		Status:			ContractStatusActive,
		ProtocolCritical:	true,
		LastChargedHeight:	1,
	}

	result, err := AccruePersistentStateRent(system, params, 6)
	require.NoError(t, err)
	require.Equal(t, uint64(50), result.RentDelta)
	require.True(t, result.ProtocolPaid)
	require.False(t, result.UserFacingFee)
	require.Equal(t, ContractStatusActive, OfficialPoolStatusForDebt(result.Subject, result.Subject.RentDebt))
}

func TestSystemRentAccountingWarnsTopsUpDeterministicallyAndKeepsCriticalExecutable(t *testing.T) {
	warning := ComputeSystemRentAccounting(SystemRentAccounting{
		AvailableFunds:			60,
		ProjectedRentPerBlock:		10,
		WarningRunwayBlocks:		10,
		CriticalRunwayBlocks:		5,
		RequiredTopUp:			20,
		ProtocolCriticalExecutable:	true,
	})
	require.Equal(t, uint64(6), warning.RunwayBlocks)
	require.Equal(t, SystemRentAlertWarning, warning.Alert)
	require.Zero(t, warning.TopUpAmount)

	critical := ComputeSystemRentAccounting(SystemRentAccounting{
		AvailableFunds:			40,
		ProjectedRentPerBlock:		10,
		WarningRunwayBlocks:		10,
		CriticalRunwayBlocks:		5,
		FeeCollectorBalance:		20,
		RequiredTopUp:			20,
		ProtocolCriticalExecutable:	true,
	})
	require.Equal(t, uint64(4), critical.RunwayBlocks)
	require.Equal(t, SystemRentAlertCritical, critical.Alert)
	require.Equal(t, uint64(20), critical.TopUpAmount)

	result := ComputeSystemRentAccounting(SystemRentAccounting{
		AvailableFunds:				10,
		ProjectedRentPerBlock:			5,
		WarningRunwayBlocks:			10,
		CriticalRunwayBlocks:			3,
		FeeCollectorBalance:			4,
		TreasuryBalance:			3,
		GovernanceConfiguredPayerBalance:	2,
		RequiredTopUp:				10,
		ProtocolCriticalExecutable:		true,
	})

	require.Equal(t, uint64(2), result.RunwayBlocks)
	require.Equal(t, SystemRentAlertInvariant, result.Alert)
	require.Equal(t, uint64(9), result.TopUpAmount)
	require.Equal(t, []SystemRentTopUpSource{
		{Source: SystemRentPayerFeeCollector, Amount: 4},
		{Source: SystemRentPayerTreasury, Amount: 3},
		{Source: SystemRentPayerGovernance, Amount: 2},
	}, result.TopUpSources)
	require.Equal(t, uint64(1), result.RemainingDebt)
	require.True(t, result.FreezeForbidden)
	require.True(t, result.Executable)
}

func TestSystemTopUpRunsBeforeUserFreezeProcessingAndProtocolCriticalKeepsExecuting(t *testing.T) {
	params := DefaultStorageRentParams()
	params.FreeStorageAllowance = 0
	params.RentRatePerByteSecond = 1

	result, err := ProcessStorageRent(RentProcessingInput{
		Params:	params,
		System: SystemRentAccounting{
			AvailableFunds:			0,
			ProjectedRentPerBlock:		10,
			WarningRunwayBlocks:		10,
			CriticalRunwayBlocks:		5,
			FeeCollectorBalance:		3,
			TreasuryBalance:		2,
			RequiredTopUp:			10,
			ProtocolCriticalExecutable:	true,
		},
		CurrentUnixSeconds:	20,
		FreezeDebtThreshold:	0,
		Subjects: []PersistentStateRecord{
			{SubjectID: "wallet", Class: StateClassWallet, DataBytes: 4, Persistent: true, Status: ContractStatusActive, LastChargedHeight: 10},
			{SubjectID: "system", Class: StateClassSystemModule, DataBytes: 4, Persistent: true, Status: ContractStatusActive, ProtocolCritical: true, LastChargedHeight: 10},
		},
	})
	require.NoError(t, err)
	require.Equal(t, []string{RentProcessingStepSystemTopUp, RentProcessingStepUserFreeze}, result.Order)
	require.Equal(t, SystemRentAlertInvariant, result.System.Alert)
	require.Equal(t, uint64(5), result.System.TopUpAmount)
	require.Equal(t, uint64(5), result.System.RemainingDebt)
	require.True(t, result.System.Executable)
	require.Equal(t, ContractStatusFrozen, result.Subjects[0].Status)
	require.Equal(t, ContractStatusActive, result.Subjects[1].Status)
	require.True(t, CanExecutePersistentStateAction(result.Subjects[1], "begin_block"))
}

func TestFrozenContractPreservesReadAndProofButBlocksNormalExecute(t *testing.T) {
	frozen := PersistentStateRecord{SubjectID: "contract", Class: StateClassContract, Status: ContractStatusFrozen, RentDebt: 20}
	require.True(t, CanReadPersistentState(frozen))
	require.True(t, CanQueryPersistentStateProof(frozen))
	require.False(t, CanExecutePersistentStateAction(frozen, "execute"))

	limited := PersistentStateRecord{SubjectID: "official-pool", Class: StateClassPoolContract, Status: ContractStatusFrozenLimited, OfficialPool: true}
	require.False(t, CanExecutePersistentStateAction(limited, PoolActionDeposit))
	require.True(t, CanExecutePersistentStateAction(limited, PoolActionMaturedWithdrawal))
}

func TestPersistentStateClassValidationRejectsUnknownClasses(t *testing.T) {
	require.NoError(t, ValidatePersistentStateClass(StateClassPoolRewardIndex))
	require.ErrorContains(t, ValidatePersistentStateClass("native_token_module"), "unsupported")
}
