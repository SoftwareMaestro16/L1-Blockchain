package types

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"
)

func TestDomainCommitmentV2BindsNameCommitterSaltChainVersionAndIntent(t *testing.T) {
	ctx := commitmentV2TestContext(20)
	commitment, err := NewDomainCommitmentV2("Alice.AET", addr(1), "salt", sdkmath.NewInt(100), 10, ctx, true)
	require.NoError(t, err)
	require.NoError(t, ValidateDomainCommitmentV2Reveal(commitment, "alice.aet", "salt", ctx))
	require.Len(t, commitment.CommitmentHash, 64)
	require.Len(t, commitment.SaltHashOptional, 64)

	wrongChain := ctx
	wrongChain.ChainID = "other-chain"
	require.ErrorContains(t, ValidateDomainCommitmentV2Reveal(commitment, "alice.aet", "salt", wrongChain), "hash mismatch")

	wrongIntent := ctx
	wrongIntent.RegistrationIntent = "renew"
	require.ErrorContains(t, ValidateDomainCommitmentV2Reveal(commitment, "alice.aet", "salt", wrongIntent), "hash mismatch")
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

func commitmentV2TestContext(currentHeight uint64) DomainCommitmentV2Context {
	return DomainCommitmentV2Context{
		CurrentHeight:      currentHeight,
		RevealWindowBlocks: 20,
		ChainID:            "aetheris-local-1",
		ModuleVersion:      DefaultDomainCommitmentVersion,
		RegistrationIntent: DomainRegistrationIntent,
	}
}
