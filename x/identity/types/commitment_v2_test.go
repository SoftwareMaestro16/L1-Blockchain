package types

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"
)

func TestDomainCommitmentV2BindsNameCommitterSaltChainVersionAndIntent(t *testing.T) {
	ctx := commitmentV2TestContext(20)
	ctx.RegistrationClass = "standard"
	ctx.MaxPrice = sdkmath.NewInt(500)
	commitment, err := NewDomainCommitmentV2("Alice.AET", addr(1), "salt", sdkmath.NewInt(100), 10, ctx, true)
	require.NoError(t, err)
	require.NoError(t, ValidateDomainCommitmentV2Reveal(commitment, "alice.aet", "salt", ctx))
	require.Len(t, commitment.CommitmentHash, 64)
	require.Len(t, commitment.SaltHashOptional, 64)

	wrongChain := ctx
	wrongChain.ChainID = "other-chain"
	require.ErrorContains(t, ValidateDomainCommitmentV2Reveal(commitment, "alice.aet", "salt", wrongChain), "hash mismatch")

	wrongIntent := ctx
	wrongIntent.RegistrationClass = "renew"
	require.ErrorContains(t, ValidateDomainCommitmentV2Reveal(commitment, "alice.aet", "salt", wrongIntent), "hash mismatch")

	wrongPrice := ctx
	wrongPrice.MaxPrice = sdkmath.NewInt(501)
	require.ErrorContains(t, ValidateDomainCommitmentV2Reveal(commitment, "alice.aet", "salt", wrongPrice), "hash mismatch")

	wrongModule := ctx
	wrongModule.ModuleName = "x/not-identity"
	require.ErrorContains(t, ValidateDomainCommitmentV2Reveal(commitment, "alice.aet", "salt", wrongModule), "hash mismatch")
}

func TestDomainCommitmentV2ExpiresAfterRevealWindowAndRejectsReplay(t *testing.T) {
	ctx := commitmentV2TestContext(20)
	commitment, err := NewDomainCommitmentV2("alice.aet", addr(1), "salt", sdkmath.NewInt(100), 10, ctx, false)
	require.NoError(t, err)
	require.Equal(t, uint64(30), commitment.ExpiresAtHeight)

	expired := ctx
	expired.CurrentHeight = commitment.ExpiresAtHeight + 1
	require.ErrorContains(t, ValidateDomainCommitmentV2Reveal(commitment, "alice.aet", "salt", expired), "expired")

	replayed := ctx
	replayed.RevealedCommitmentHashes = []string{commitment.CommitmentHash}
	require.ErrorContains(t, ValidateDomainCommitmentV2(commitment, replayed), "already revealed")
}

func TestDomainCommitmentV2RevealWindowBoundaries(t *testing.T) {
	ctx := commitmentV2TestContext(10)
	commitment, err := NewDomainCommitmentV2("alice.aet", addr(1), "salt", sdkmath.NewInt(100), 10, ctx, false)
	require.NoError(t, err)

	atStart := ctx
	atStart.CurrentHeight = commitment.CreatedAtHeight
	require.NoError(t, ValidateDomainCommitmentV2Reveal(commitment, "alice.aet", "salt", atStart))

	atEnd := ctx
	atEnd.CurrentHeight = commitment.ExpiresAtHeight
	require.NoError(t, ValidateDomainCommitmentV2Reveal(commitment, "alice.aet", "salt", atEnd))

	tooEarly := ctx
	tooEarly.CurrentHeight = commitment.CreatedAtHeight - 1
	require.ErrorContains(t, ValidateDomainCommitmentV2Reveal(commitment, "alice.aet", "salt", tooEarly), "before commitment")

	tooLate := ctx
	tooLate.CurrentHeight = commitment.ExpiresAtHeight + 1
	require.ErrorContains(t, ValidateDomainCommitmentV2Reveal(commitment, "alice.aet", "salt", tooLate), "expired")
}

func TestDomainCommitmentV2RefundRules(t *testing.T) {
	ctx := commitmentV2TestContext(20)
	ctx.RefundValidRevealDeposit = true
	ctx.RefundCleanExpiryDeposit = true
	commitment, err := NewDomainCommitmentV2("alice.aet", addr(1), "salt", sdkmath.NewInt(100), 10, ctx, false)
	require.NoError(t, err)

	refundable, err := DomainCommitmentV2Refundable(commitment, DomainCommitmentRefundValidReveal, ctx)
	require.NoError(t, err)
	require.True(t, refundable)

	expired := ctx
	expired.CurrentHeight = commitment.ExpiresAtHeight + 1
	refundable, err = DomainCommitmentV2Refundable(commitment, DomainCommitmentRefundCleanExpiry, expired)
	require.NoError(t, err)
	require.True(t, refundable)

	replayed := expired
	replayed.RevealedCommitmentHashes = []string{commitment.CommitmentHash}
	_, err = DomainCommitmentV2Refundable(commitment, DomainCommitmentRefundCleanExpiry, replayed)
	require.ErrorContains(t, err, "cannot be refunded again")
}

func TestRegistrationCommitRevealCreatesUsedTombstoneAndRejectsReplay(t *testing.T) {
	state, _ := registerSpecDomain(t, "alice", addr(1), "salt", 10)
	require.Len(t, state.UsedCommitments, 1)
	used := state.UsedCommitments[0]
	require.Equal(t, "alice.aet", used.Name)
	require.Equal(t, addr(1), used.Owner)
	require.Equal(t, uint64(11), used.RevealedHeight)
	require.Equal(t, DefaultIdentityModuleName, used.ModuleName)
	require.Equal(t, DefaultRegistrationClass, used.RegistrationClass)
	require.Equal(t, "0", used.MaxPrice)

	_, err := CommitDomainRegistration(state, "bob.aet", addr(1), used.CommitmentHash, 20)
	require.ErrorContains(t, err, "already used")
}

func FuzzDomainCommitmentV2RevealReplayProtection(f *testing.F) {
	f.Add("alice", "salt-a", uint64(10), uint64(20), uint64(100), "standard")
	f.Add("service.api", "salt-b", uint64(2), uint64(3), uint64(0), "premium")
	f.Fuzz(func(t *testing.T, label string, salt string, createdHeight uint64, revealWindow uint64, maxPrice uint64, registrationClass string) {
		if createdHeight == 0 {
			createdHeight = 1
		}
		if revealWindow == 0 || revealWindow > 1000 {
			revealWindow = 10
		}
		if registrationClass == "" {
			registrationClass = "standard"
		}
		name := label + ".aet"
		if _, err := NormalizeAETDomain(name); err != nil {
			t.Skip()
		}
		ctx := commitmentV2TestContext(createdHeight)
		ctx.RevealWindowBlocks = revealWindow
		ctx.RegistrationClass = registrationClass
		ctx.MaxPrice = sdkmath.NewIntFromUint64(maxPrice)
		commitment, err := NewDomainCommitmentV2(name, addr(1), salt, sdkmath.NewInt(1), createdHeight, ctx, true)
		if err != nil {
			t.Skip()
		}
		ctx.CurrentHeight = commitment.ExpiresAtHeight
		require.NoError(t, ValidateDomainCommitmentV2Reveal(commitment, name, salt, ctx))

		replayed := ctx
		replayed.RevealedCommitmentHashes = []string{commitment.CommitmentHash}
		require.ErrorContains(t, ValidateDomainCommitmentV2Reveal(commitment, name, salt, replayed), "already revealed")
	})
}

func commitmentV2TestContext(currentHeight uint64) DomainCommitmentV2Context {
	return DomainCommitmentV2Context{
		CurrentHeight:		currentHeight,
		RevealWindowBlocks:	20,
		ChainID:		"aetra-local-1",
		ModuleVersion:		DefaultDomainCommitmentVersion,
		RegistrationIntent:	DomainRegistrationIntent,
		RegistrationClass:	DefaultRegistrationClass,
		MaxPrice:		sdkmath.ZeroInt(),
	}
}
