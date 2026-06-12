package types

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	aetraaddress "github.com/sovereign-l1/l1/app/addressing"
)

func TestStakingEventGoldenForEveryType(t *testing.T) {
	actor := eventAEAddress(0x11)
	pool := eventAEAddress(0x22)
	validator := eventAEAddress(0x33)
	proof, err := BuildStakingProofMetadata(StakingProofRequest{
		Kind:		StakingProofShare,
		Height:		77,
		PoolID:		"pool-a",
		Account:	actor,
		AppHash:	"app-root-ref",
		RootHash:	"nominator-pool-root-ref",
	})
	require.NoError(t, err)

	cases := []struct {
		name	string
		in	StakingEvent
		hash	string
	}{
		{
			name:	"account activated",
			in: StakingEvent{
				Type:			EventAccountActivated,
				Actor:			actor,
				Height:			77,
				Sequence:		0,
				StateKey:		AccountActivationEventStateKey(actor),
				ProofMetadataHash:	proof.MetadataHash,
			},
			hash:	"cb35cf762ba9afe342a65c6712e190fa8e61765292a10028e6481e7b45d98a6c",
		},
		{
			name:	"pool stake deposited",
			in: StakingEvent{
				Type:			EventPoolStakeDeposited,
				Actor:			actor,
				PoolContract:		pool,
				Amount:			1_000,
				Shares:			990,
				Height:			77,
				Epoch:			3,
				Sequence:		1,
				StateKey:		PoolDepositProofStateKey("pool-a", actor),
				ProofMetadataHash:	proof.MetadataHash,
			},
			hash:	"0578d0bc5949915f193715bb907c6a1d3a195f5fef3073d711ace0130c1e2af4",
		},
		{
			name:	"pool shares minted",
			in: StakingEvent{
				Type:			EventPoolSharesMinted,
				Actor:			actor,
				PoolContract:		pool,
				Shares:			990,
				Height:			77,
				Epoch:			3,
				Sequence:		2,
				StateKey:		PoolShareProofStateKey("pool-a", actor),
				ProofMetadataHash:	proof.MetadataHash,
			},
			hash:	"b910df7323e49e1d2a35056c447653ea49d4fc9c1c8e0f7e03a28214ee3fe91a",
		},
		{
			name:	"pool allocation updated",
			in: StakingEvent{
				Type:			EventPoolAllocationUpdated,
				Actor:			pool,
				PoolContract:		pool,
				Validator:		validator,
				Amount:			700,
				Height:			77,
				Epoch:			3,
				Sequence:		3,
				StateKey:		PoolAllocationProofStateKey("pool-a", 3),
				ProofMetadataHash:	proof.MetadataHash,
			},
			hash:	"22314ed670c83e4580d563c1d35a06b7d963d5b118ccdd7f3e3c55d064b0b7a7",
		},
		{
			name:	"pool unbonding requested",
			in: StakingEvent{
				Type:			EventPoolUnbondingRequested,
				Actor:			actor,
				PoolContract:		pool,
				Shares:			250,
				Height:			77,
				Epoch:			3,
				Sequence:		4,
				StateKey:		string(PoolUnbondingKey("pool-a", actor, "req-1")),
				ProofMetadataHash:	proof.MetadataHash,
			},
			hash:	"35a58960e4010e9dabe969fa74a6470388bcad7a9d52934cbb7e098e448c0279",
		},
		{
			name:	"pool unbonding completed",
			in: StakingEvent{
				Type:			EventPoolUnbondingCompleted,
				Actor:			actor,
				PoolContract:		pool,
				Amount:			245,
				Height:			77,
				Epoch:			3,
				Sequence:		5,
				StateKey:		string(PoolUnbondingKey("pool-a", actor, "req-1")),
				ProofMetadataHash:	proof.MetadataHash,
			},
			hash:	"1b1e11ddafcbda3ef340d583c909a001893365561a972d05e47eccca6d59b5ba",
		},
		{
			name:	"pool rewards claimed",
			in: StakingEvent{
				Type:			EventPoolRewardsClaimed,
				Actor:			actor,
				PoolContract:		pool,
				Amount:			42,
				Height:			77,
				Epoch:			3,
				Sequence:		6,
				StateKey:		PoolRewardProofStateKey("pool-a", actor),
				ProofMetadataHash:	proof.MetadataHash,
			},
			hash:	"756a782c85cdfbbd8a20e83d01933244305d95b37128f4e598ee7a5d56cda942",
		},
		{
			name:	"stake reputation claimed",
			in: StakingEvent{
				Type:			EventStakeReputationClaimed,
				Actor:			actor,
				PoolContract:		pool,
				Amount:			9,
				Height:			77,
				Epoch:			3,
				Sequence:		7,
				StateKey:		StakeReputationProofStateKey(actor),
				ProofMetadataHash:	proof.MetadataHash,
			},
			hash:	"d8c8dfc7b3c7cf461f949d20eb8fb535d8e126ac66f9b7e41aeb6a1e7f59acb4",
		},
		{
			name:	"validator registered",
			in: StakingEvent{
				Type:		EventValidatorRegistered,
				Actor:		validator,
				Validator:	validator,
				Amount:		300_000,
				Height:		77,
				Epoch:		3,
				Sequence:	8,
				StateKey:	string(ValidatorKey(validator)),
			},
			hash:	"182e61b2e422aef770bf83a7535656345f9470ed3c5d9ff2af01a27206e8d4a3",
		},
		{
			name:	"validator updated",
			in: StakingEvent{
				Type:		EventValidatorUpdated,
				Actor:		validator,
				Validator:	validator,
				Height:		77,
				Epoch:		3,
				Sequence:	9,
				StateKey:	string(ValidatorKey(validator)),
			},
			hash:	"ee5e87bd9d5215840beb2f06d10fb80bb325ba8dec6cd1b5397cddf94e660687",
		},
		{
			name:	"advanced stake delegated",
			in: StakingEvent{
				Type:		EventAdvancedStakeDelegated,
				Actor:		actor,
				Validator:	validator,
				Amount:		500,
				Height:		77,
				Epoch:		3,
				Sequence:	10,
				StateKey:	AdvancedStakeEventStateKey(actor, validator),
			},
			hash:	"365df643540507dc7f9a1af51851c3f047ffa5286c4ebc1ce1f37ff5953eefcf",
		},
		{
			name:	"advanced stake undelegated",
			in: StakingEvent{
				Type:		EventAdvancedStakeUndelegated,
				Actor:		actor,
				Validator:	validator,
				Amount:		200,
				Height:		77,
				Epoch:		3,
				Sequence:	11,
				StateKey:	AdvancedStakeEventStateKey(actor, validator),
			},
			hash:	"ba518386cfa0cb9c0de1ffccac14023b582898d803c87437c536b557ecf0cbd9",
		},
		{
			name:	"advanced stake redelegated",
			in: StakingEvent{
				Type:		EventAdvancedStakeRedelegated,
				Actor:		actor,
				Validator:	validator,
				Amount:		150,
				Height:		77,
				Epoch:		3,
				Sequence:	12,
				StateKey:	AdvancedStakeRedelegationEventStateKey(actor, eventAEAddress(0x44), validator),
			},
			hash:	"c5d1dec04139f2824f65094c4aef5a310c0c8068f33118af57045c37051f0d16",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			event, err := NewStakingEvent(tc.in)
			require.NoError(t, err)
			require.Equal(t, tc.hash, event.EventHash)
			require.Equal(t, tc.in.Actor, event.Actor)
			require.Equal(t, tc.in.PoolContract, event.PoolContract)
			require.Equal(t, tc.in.Validator, event.Validator)
			require.Equal(t, tc.in.Amount, event.Amount)
			require.Equal(t, tc.in.Shares, event.Shares)
			require.Equal(t, tc.in.Height, event.Height)
			require.Equal(t, tc.in.Epoch, event.Epoch)
			require.Equal(t, tc.in.StateKey, event.StateKey)
			require.NotContains(t, fmt.Sprint(event.OrderedAttributes()), "private_key")
			require.NotContains(t, fmt.Sprint(event.OrderedAttributes()), "seed phrase")
			require.NotContains(t, fmt.Sprint(event.OrderedAttributes()), "secret")
		})
	}
}

func TestStakingReceiptEventOrderDeterministicForMultiMessageTx(t *testing.T) {
	actor := eventAEAddress(0x55)
	pool := eventAEAddress(0x66)
	first := mustEvent(t, StakingEvent{
		Type:		EventPoolStakeDeposited,
		Actor:		actor,
		PoolContract:	pool,
		Amount:		1_000,
		Shares:		1_000,
		Height:		90,
		Epoch:		4,
		Sequence:	0,
		StateKey:	PoolDepositProofStateKey("pool-b", actor),
	})
	second := mustEvent(t, StakingEvent{
		Type:		EventPoolSharesMinted,
		Actor:		actor,
		PoolContract:	pool,
		Shares:		1_000,
		Height:		90,
		Epoch:		4,
		Sequence:	1,
		StateKey:	PoolShareProofStateKey("pool-b", actor),
	})

	a, err := NewStakingReceipt("TX-ABC", 90, []StakingEvent{second, first})
	require.NoError(t, err)
	b, err := NewStakingReceipt("tx-abc", 90, []StakingEvent{first, second})
	require.NoError(t, err)

	require.Equal(t, first.EventHash, a.Events[0].EventHash)
	require.Equal(t, second.EventHash, a.Events[1].EventHash)
	require.Equal(t, b.ReceiptHash, a.ReceiptHash)
	require.Equal(t, "66554346fe964d6a61ba78e565ee20d73c16d4d08e77f90f21636495280592f2", a.ReceiptHash)
}

func TestStakingEventsRejectSecretsAndMisplacedValidator(t *testing.T) {
	actor := eventAEAddress(0x77)
	pool := eventAEAddress(0x88)
	validator := eventAEAddress(0x99)

	_, err := NewStakingEvent(StakingEvent{
		Type:		EventPoolRewardsClaimed,
		Actor:		actor,
		PoolContract:	pool,
		Validator:	validator,
		Amount:		10,
		Height:		1,
		StateKey:	PoolRewardProofStateKey("pool-c", actor),
	})
	require.ErrorContains(t, err, "validator is only allowed")

	_, err = NewStakingEvent(StakingEvent{
		Type:		EventPoolRewardsClaimed,
		Actor:		actor,
		PoolContract:	pool,
		Amount:		10,
		Height:		1,
		StateKey:	"staking/rewards/private_key/leak",
	})
	require.ErrorContains(t, err, "secret material")

	_, err = NewStakingEvent(StakingEvent{
		Type:		EventPoolRewardsClaimed,
		Actor:		"4:1111111111111111111111111111111111111111111111111111111111111111",
		PoolContract:	pool,
		Amount:		10,
		Height:		1,
		StateKey:	PoolRewardProofStateKey("pool-c", actor),
	})
	require.ErrorContains(t, err, "AE")
}

func TestAccountActivationAndRewardClaimEventsAreStable(t *testing.T) {
	actor := eventAEAddress(0xaa)
	pool := eventAEAddress(0xbb)
	account := mustEvent(t, StakingEvent{
		Type:		EventAccountActivated,
		Actor:		actor,
		Height:		12,
		Sequence:	0,
		StateKey:	AccountActivationEventStateKey(actor),
	})
	reward := mustEvent(t, StakingEvent{
		Type:		EventPoolRewardsClaimed,
		Actor:		actor,
		PoolContract:	pool,
		Amount:		777,
		Height:		12,
		Epoch:		2,
		Sequence:	1,
		StateKey:	PoolRewardProofStateKey("pool-reward", actor),
	})

	accountAgain := mustEvent(t, account)
	rewardAgain := mustEvent(t, reward)
	require.Equal(t, account.EventHash, accountAgain.EventHash)
	require.Equal(t, reward.EventHash, rewardAgain.EventHash)
	require.Equal(t, "4a8f689c06b1a466f44ff48d36d7a7dc3b52fc559d3b79e29a0cb78ed3e72f1f", account.EventHash)
	require.Equal(t, "c9bd403df613c311bc81b0425a8f4319e0c7c43b9ca701d3ebae80dd70080402", reward.EventHash)
}

func mustEvent(t *testing.T, event StakingEvent) StakingEvent {
	t.Helper()
	out, err := NewStakingEvent(event)
	require.NoError(t, err)
	return out
}

func eventAEAddress(fill byte) string {
	bz := make([]byte, 20)
	for i := range bz {
		bz[i] = fill
	}
	return aetraaddress.FormatAccAddress(sdk.AccAddress(bz))
}
