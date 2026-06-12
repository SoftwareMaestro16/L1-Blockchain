package types

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	nominatorpooltypes "github.com/sovereign-l1/l1/x/nominator-pool/types"
	proofregistrytypes "github.com/sovereign-l1/l1/x/proofregistry/types"
	reputationtypes "github.com/sovereign-l1/l1/x/reputation/types"
	storagerenttypes "github.com/sovereign-l1/l1/x/storage-rent/types"
	validatorregistrytypes "github.com/sovereign-l1/l1/x/validator-registry/types"
)

func TestFullGenesisExportImportPreservesAccountsAndAuthPolicies(t *testing.T) {
	source := newFullGenesisStore(fullGenesisFixture(t))
	exported, err := ExportFullGenesis(source)
	require.NoError(t, err)

	target := newFullGenesisStore(DefaultFullGenesis())
	require.NoError(t, ImportFullGenesis(target, exported))

	roundTrip, err := ExportFullGenesis(target)
	require.NoError(t, err)
	require.Equal(t, exported.Accounts, roundTrip.Accounts)
	require.Equal(t, exported.Accounts[0].AuthPolicy, roundTrip.Accounts[0].AuthPolicy)
	require.Equal(t, "single_key", roundTrip.Accounts[0].AuthPolicy.Mode)
}

func TestFullGenesisExportImportPreservesPoolsRewardsAndUnbondings(t *testing.T) {
	source := newFullGenesisStore(fullGenesisFixture(t))
	exported, err := ExportFullGenesis(source)
	require.NoError(t, err)

	pool := exported.LiquidStakingState.Pools[0]
	require.Equal(t, uint64(2_000), pool.TotalShares)
	require.Equal(t, uint64(1_234_567), pool.RewardIndex)
	require.Equal(t, uint64(77), pool.DelegatorShares[0].PendingRewards)
	require.Len(t, pool.Allocations, 2)
	require.Len(t, pool.PendingWithdrawals, 1)
	require.Len(t, pool.UnbondingQueue, 1)

	target := newFullGenesisStore(DefaultFullGenesis())
	require.NoError(t, ImportFullGenesis(target, exported))
	require.Equal(t, exported.LiquidStakingState, target.state.LiquidStakingState)
}

func TestFullGenesisExportImportPreservesReputationAndStorageRent(t *testing.T) {
	exported, err := ExportFullGenesis(newFullGenesisStore(fullGenesisFixture(t)))
	require.NoError(t, err)

	require.Len(t, exported.ReputationState.ValidatorScores, 1)
	require.Equal(t, uint32(500), exported.ReputationState.ValidatorScores[0].UptimeScore)
	require.Equal(t, uint64(0), exported.ReputationState.ValidatorScores[0].LastUpdateHeight)
	require.Len(t, exported.StorageRentState.Contracts, 1)
	require.Equal(t, uint64(55), exported.StorageRentState.Contracts[0].RentDebt)

	target := newFullGenesisStore(DefaultFullGenesis())
	require.NoError(t, ImportFullGenesis(target, exported))
	require.Equal(t, exported.ReputationState, target.state.ReputationState)
	require.Equal(t, exported.StorageRentState, target.state.StorageRentState)
}

func TestFullGenesisImportRejectsMalformedDuplicateAccountBeforeWrite(t *testing.T) {
	gs := fullGenesisFixture(t)
	duplicate := gs.Accounts[0]
	duplicate.AddressUser = gs.Accounts[1].AddressUser
	duplicate.AddressRaw = gs.Accounts[1].AddressRaw
	gs.Accounts = append(gs.Accounts, duplicate)
	target := newFullGenesisStore(DefaultFullGenesis())

	err := ImportFullGenesis(target, gs)

	require.ErrorContains(t, err, "duplicate native account")
	require.Equal(t, 0, target.writes)
}

func TestFullGenesisImportRejectsDuplicatePoolShareOrAllocationBeforeWrite(t *testing.T) {
	t.Run("duplicate share", func(t *testing.T) {
		gs := fullGenesisFixture(t)
		pool := gs.LiquidStakingState.Pools[0]
		pool.DelegatorShares = append(pool.DelegatorShares, pool.DelegatorShares[0])
		pool.TotalShares += pool.DelegatorShares[0].Shares
		gs.LiquidStakingState.Pools[0] = pool
		target := newFullGenesisStore(DefaultFullGenesis())

		err := ImportFullGenesis(target, gs)

		require.ErrorContains(t, err, "duplicate pool delegator")
		require.Equal(t, 0, target.writes)
	})

	t.Run("duplicate allocation", func(t *testing.T) {
		gs := fullGenesisFixture(t)
		pool := gs.LiquidStakingState.Pools[0]
		duplicate := pool.Allocations[0]
		duplicate.Amount = 1
		pool.Allocations = append(pool.Allocations, duplicate)
		gs.LiquidStakingState.Pools[0] = pool
		target := newFullGenesisStore(DefaultFullGenesis())

		err := ImportFullGenesis(target, gs)

		require.ErrorContains(t, err, "pool allocations must be sorted by unique validator address")
		require.Equal(t, 0, target.writes)
	})
}

func TestFullGenesisRejectsUnsupportedAccountOrStakingVersion(t *testing.T) {
	gs := fullGenesisFixture(t)
	gs.Accounts[0].Version = CurrentAccountVersion + 10
	require.ErrorContains(t, ValidateFullGenesis(NormalizeFullGenesis(gs)), "unsupported native account version")

	gs = fullGenesisFixture(t)
	gs.LiquidStakingVersion = 99
	require.ErrorContains(t, ValidateFullGenesis(NormalizeFullGenesis(gs)), "liquid staking unsupported genesis version")
}

func TestFullGenesisExportOrderDeterministicAcrossRepeatedRuns(t *testing.T) {
	source := newFullGenesisStore(fullGenesisFixture(t))

	first, err := ExportFullGenesis(source)
	require.NoError(t, err)
	second, err := ExportFullGenesis(source)
	require.NoError(t, err)

	firstJSON, err := json.Marshal(first)
	require.NoError(t, err)
	secondJSON, err := json.Marshal(second)
	require.NoError(t, err)
	require.Equal(t, firstJSON, secondJSON)
	require.Less(t, first.Accounts[0].AddressUser, first.Accounts[1].AddressUser)
	require.Equal(t, "pool-a", first.LiquidStakingState.Pools[0].PoolID)
}

func TestFullGenesisExportDoesNotContainPrivateKeyOrSeed(t *testing.T) {
	exported, err := ExportFullGenesis(newFullGenesisStore(fullGenesisFixture(t)))
	require.NoError(t, err)
	bz, err := json.Marshal(exported)
	require.NoError(t, err)
	lower := strings.ToLower(string(bz))
	require.NotContains(t, lower, "private_key")
	require.NotContains(t, lower, "private key")
	require.NotContains(t, lower, "seed_phrase")
	require.NotContains(t, lower, "seed phrase")
	require.NotContains(t, lower, "mnemonic")

	bad := fullGenesisFixture(t)
	bad.Accounts[0].Metadata.MetadataHash = "seed_phrase:do-not-export"
	_, err = ExportFullGenesis(newFullGenesisStore(bad))
	require.ErrorContains(t, err, "seed phrases")
}

type testFullGenesisStore struct {
	state	FullGenesisState
	writes	int
}

func newFullGenesisStore(state FullGenesisState) *testFullGenesisStore {
	return &testFullGenesisStore{state: NormalizeFullGenesis(state)}
}

func (s *testFullGenesisStore) FullGenesisState() (FullGenesisState, error) {
	return s.state, nil
}

func (s *testFullGenesisStore) SetFullGenesisState(state FullGenesisState) error {
	s.writes++
	s.state = state
	return nil
}

func fullGenesisFixture(t *testing.T) FullGenesisState {
	t.Helper()
	gs := DefaultFullGenesis()
	accountA, err := MigrateAccountV1ToV2(v1Account(t, 0x52, 2, 20))
	require.NoError(t, err)
	accountB := v1Account(t, 0x51, 1, 10)
	gs.Accounts = []Account{accountA, accountB}
	gs.ValidatorRegistryState = validatorregistrytypes.State{Validators: []validatorregistrytypes.ValidatorRecord{
		validatorRecord(t, 0x71),
	}}
	gs.LiquidStakingState = nominatorpooltypes.State{Pools: []nominatorpooltypes.NominatorPool{
		liquidStakingPool(t),
	}}
	gs.ReputationState = reputationState(t)
	gs.StorageRentState = storageRentState()
	gs.ProofMetadataState = proofMetadataState(t)
	return gs
}

func validatorRecord(t *testing.T, fill byte) validatorregistrytypes.ValidatorRecord {
	t.Helper()
	operator, _ := testAddressPair(t, fill)
	treasury, _ := testAddressPair(t, fill+1)
	withdrawal, _ := testAddressPair(t, fill+2)
	emergency, _ := testAddressPair(t, fill+3)
	return validatorregistrytypes.ValidatorRecord{
		OperatorAddress:	operator,
		ConsensusPublicKey:	"ed25519-consensus-key",
		TreasuryAddress:	treasury,
		WithdrawalAddress:	withdrawal,
		EmergencyAddress:	emergency,
		CommissionPolicy:	validatorregistrytypes.DefaultCommissionPolicy(),
		Status:			validatorregistrytypes.StatusActive,
		Capabilities:		[]string{"liquid-staking"},
		SelfBond:		validatorregistrytypes.DefaultSoloValidatorMinSelfStake,
	}
}

func liquidStakingPool(t *testing.T) nominatorpooltypes.NominatorPool {
	t.Helper()
	contractUser, contractRaw := testAddressPair(t, 0x61)
	operator, _ := testAddressPair(t, 0x62)
	delegator, _ := testAddressPair(t, 0x63)
	validatorA, _ := testAddressPair(t, 0x64)
	validatorB, _ := testAddressPair(t, 0x65)
	return nominatorpooltypes.NominatorPool{
		PoolID:			"pool-a",
		ContractAddressUser:	contractUser,
		ContractAddressRaw:	contractRaw,
		OfficialLiquidStaking:	true,
		PoolOperator:		operator,
		TotalShares:		2_000,
		TotalBondedStake:	3_000,
		Allocations: []nominatorpooltypes.PoolAllocation{
			{ValidatorAddress: validatorA, Amount: 1_000, Height: 7},
			{ValidatorAddress: validatorB, Amount: 1_500, Height: 8},
		},
		DelegatorShares: []nominatorpooltypes.DelegatorShare{
			{Delegator: delegator, Shares: 2_000, RewardIndexCheckpoint: 1_000_000, PendingRewards: 77, SlashIndexCheckpoint: 3},
		},
		PendingDeposits: []nominatorpooltypes.PendingDeposit{
			{Delegator: delegator, Amount: 10, Height: 5},
		},
		PendingWithdrawals: []nominatorpooltypes.PendingWithdrawal{
			{WithdrawalID: "withdraw-a", Delegator: delegator, Shares: 100, Amount: 150, RequestHeight: 9, CompleteHeight: 99, Status: nominatorpooltypes.WithdrawalStatusPending},
		},
		RewardIndex:		1_234_567,
		SlashIndex:		3,
		PoolCommissionBps:	100,
		Status:			nominatorpooltypes.PoolStatusActive,
		UnbondingQueue: []nominatorpooltypes.UnbondingEntry{
			{WithdrawalID: "withdraw-a", Delegator: delegator, Amount: 150, CompleteHeight: 99, Status: nominatorpooltypes.WithdrawalStatusPending},
		},
	}
}

func reputationState(t *testing.T) reputationtypes.ConsolidatedReputationState {
	t.Helper()
	params := reputationtypes.DefaultReputationParams()
	state := reputationtypes.NewConsolidatedReputationState(params)
	vs := reputationtypes.NewValidatorScore("4:0000000000000000000000007171717171717171717171717171717171717171")
	vs.UptimeScore = 500
	vs.TotalScore = reputationtypes.ComputeValidatorTotalScore(vs)
	state.ValidatorScores = []reputationtypes.ValidatorScore{*vs}
	return reputationtypes.NormalizeConsolidatedState(state)
}

func storageRentState() storagerenttypes.StorageRentState {
	return storagerenttypes.StorageRentState{
		Contracts: []storagerenttypes.ContractRentRecord{
			{
				ContractAddress:	"contract-a",
				ActorID:		"actor-a",
				StorageBytes:		10_000,
				PrepaidRentBalance:	20,
				RentDebt:		55,
				LastChargedHeight:	12,
				Status:			storagerenttypes.ContractStatusActive,
				ArchivalProofRoot:	storagerenttypes.DefaultProofRoot,
			},
		},
		Distributions: []storagerenttypes.RentDistributionRecord{
			{ContractAddress: "contract-a", Height: 13, Amount: 10, FeeCollectorAmount: 5, TreasuryAmount: 4, BurnAmount: 1},
		},
	}
}

func proofMetadataState(t *testing.T) proofregistrytypes.ProofRegistryState {
	t.Helper()
	state, err := proofregistrytypes.NewProofRegistryState(proofregistrytypes.DefaultHistoryWindow)
	require.NoError(t, err)
	return state
}
