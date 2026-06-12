package types

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"
)

func TestUnauthorizedMintRejected(t *testing.T) {
	state := newMintAuthorityState(t)
	decision := emissionDecision("x/attacker", DefaultBaseDenom, 100, 1, 10, true)
	msg := mintMsg("x/attacker", DefaultBaseDenom, 100, 1, 10, decision.DecisionHash)
	_, _, err := ApplyMintProtocolCoins(state, msg, decision, ConstitutionEmergencyAuthorization{})
	require.ErrorContains(t, err, "scheduled mint caller must be x/emissions")
}

func TestAuthorizedEmissionMintSucceeds(t *testing.T) {
	state := newMintAuthorityState(t)
	decision := emissionDecision(DefaultEmissionCaller, DefaultBaseDenom, 250, 3, 20, true)
	next, event, err := ApplyMintProtocolCoins(state, mintMsg(DefaultEmissionCaller, DefaultBaseDenom, 250, 3, 20, decision.DecisionHash), decision, ConstitutionEmergencyAuthorization{})
	require.NoError(t, err)
	require.Equal(t, sdkmath.NewInt(250), event.Amount)
	require.False(t, event.Emergency)
	require.Equal(t, sdkmath.NewInt(250), QueryMintedByEpoch(next, QueryMintedByEpochRequest{Epoch: 3}).Amount)
	require.Equal(t, sdkmath.NewInt(250), QueryMintedLifetime(next, DefaultBaseDenom).Amount)
	require.NoError(t, CheckMintAuthorityInvariants(next))
}

func TestWrongDenomRejected(t *testing.T) {
	state := newMintAuthorityState(t)
	decision := emissionDecision(DefaultEmissionCaller, "uatom", 100, 1, 10, true)
	_, _, err := ApplyMintProtocolCoins(state, mintMsg(DefaultEmissionCaller, "uatom", 100, 1, 10, decision.DecisionHash), decision, ConstitutionEmergencyAuthorization{})
	require.ErrorContains(t, err, "denom must be base denom")
}

func TestCapEnforced(t *testing.T) {
	state := mintAuthorityStateWithCaps(t, 1_000, 1_500)
	decision := emissionDecision(DefaultEmissionCaller, DefaultBaseDenom, 900, 4, 10, true)
	next, _, err := ApplyMintProtocolCoins(state, mintMsg(DefaultEmissionCaller, DefaultBaseDenom, 900, 4, 10, decision.DecisionHash), decision, ConstitutionEmergencyAuthorization{})
	require.NoError(t, err)

	decision = emissionDecision(DefaultEmissionCaller, DefaultBaseDenom, 101, 4, 11, true)
	_, _, err = ApplyMintProtocolCoins(next, mintMsg(DefaultEmissionCaller, DefaultBaseDenom, 101, 4, 11, decision.DecisionHash), decision, ConstitutionEmergencyAuthorization{})
	require.ErrorContains(t, err, "epoch cap")

	decision = emissionDecision(DefaultEmissionCaller, DefaultBaseDenom, 500, 5, 12, true)
	next, _, err = ApplyMintProtocolCoins(next, mintMsg(DefaultEmissionCaller, DefaultBaseDenom, 500, 5, 12, decision.DecisionHash), decision, ConstitutionEmergencyAuthorization{})
	require.NoError(t, err)

	decision = emissionDecision(DefaultEmissionCaller, DefaultBaseDenom, 101, 6, 13, true)
	_, _, err = ApplyMintProtocolCoins(next, mintMsg(DefaultEmissionCaller, DefaultBaseDenom, 101, 6, 13, decision.DecisionHash), decision, ConstitutionEmergencyAuthorization{})
	require.ErrorContains(t, err, "lifetime cap")
}

func TestDirectUserMintRejected(t *testing.T) {
	state := newMintAuthorityState(t)
	decision := emissionDecision("user-address", DefaultBaseDenom, 100, 1, 10, true)
	_, _, err := ApplyMintProtocolCoins(state, mintMsg("user-address", DefaultBaseDenom, 100, 1, 10, decision.DecisionHash), decision, ConstitutionEmergencyAuthorization{})
	require.ErrorContains(t, err, "scheduled mint caller must be x/emissions")
}

func TestNonEmissionsModuleRejectedUnlessEmergencyAllowlisted(t *testing.T) {
	state := newMintAuthorityState(t)
	decision := emissionDecision("x/treasury", DefaultBaseDenom, 100, 1, 10, true)
	_, _, err := ApplyMintProtocolCoins(state, mintMsg("x/treasury", DefaultBaseDenom, 100, 1, 10, decision.DecisionHash), decision, ConstitutionEmergencyAuthorization{})
	require.ErrorContains(t, err, "scheduled mint caller must be x/emissions")

	state = emergencyMintAuthorityState(t)
	auth := emergencyAuthorization("x/constitution", DefaultBaseDenom, 100, 2, 20, true, true)
	msg := mintMsg("x/constitution", DefaultBaseDenom, 100, 2, 20, "")
	msg.Emergency = true
	msg.ConstitutionDecisionHash = auth.AuthorizationHash
	next, event, err := ApplyMintProtocolCoins(state, msg, EmissionDecision{}, auth)
	require.NoError(t, err)
	require.True(t, event.Emergency)
	require.Equal(t, sdkmath.NewInt(100), QueryMintedLifetime(next, DefaultBaseDenom).Amount)
	require.NoError(t, CheckMintAuthorityInvariants(next))
}

func TestEmergencyMintMustBeBoundedByConstitution(t *testing.T) {
	state := emergencyMintAuthorityState(t)
	auth := emergencyAuthorization("x/constitution", DefaultBaseDenom, 100, 2, 20, true, false)
	msg := mintMsg("x/constitution", DefaultBaseDenom, 100, 2, 20, "")
	msg.Emergency = true
	msg.ConstitutionDecisionHash = auth.AuthorizationHash
	_, _, err := ApplyMintProtocolCoins(state, msg, EmissionDecision{}, auth)
	require.ErrorContains(t, err, "bounded by constitution")
}

func TestMinterAccountRegistrationRequiredAtGenesis(t *testing.T) {
	params := DefaultMintAuthorityParams()
	registration := DefaultMintAuthorityRegistration()
	registration.Registered = false
	_, err := NewMintAuthorityState(params, registration, []MintCap{defaultCap(params, 1_000, 10_000)})
	require.ErrorContains(t, err, "registered in system-registry")

	registration = DefaultMintAuthorityRegistration()
	registration.UserControlled = true
	_, err = NewMintAuthorityState(params, registration, []MintCap{defaultCap(params, 1_000, 10_000)})
	require.ErrorContains(t, err, "cannot be user controlled")
}

func TestMintAmountMustMatchEmissionDecision(t *testing.T) {
	state := newMintAuthorityState(t)
	decision := emissionDecision(DefaultEmissionCaller, DefaultBaseDenom, 100, 1, 10, true)
	msg := mintMsg(DefaultEmissionCaller, DefaultBaseDenom, 101, 1, 10, decision.DecisionHash)
	_, _, err := ApplyMintProtocolCoins(state, msg, decision, ConstitutionEmergencyAuthorization{})
	require.ErrorContains(t, err, "must match emissions decision")

	decision = emissionDecision(DefaultEmissionCaller, DefaultBaseDenom, 100, 1, 10, false)
	msg = mintMsg(DefaultEmissionCaller, DefaultBaseDenom, 100, 1, 10, decision.DecisionHash)
	_, _, err = ApplyMintProtocolCoins(state, msg, decision, ConstitutionEmergencyAuthorization{})
	require.ErrorContains(t, err, "not approved")
}

func TestUpdateParamsRejectsNonCanonicalNormalCaller(t *testing.T) {
	state := newMintAuthorityState(t)
	params := state.Params
	params.NormalEmissionCaller = "x/treasury"
	_, err := ApplyUpdateMintAuthorityParams(state, MsgUpdateMintAuthorityParams{
		Authority:	state.Params.Authority,
		Params:		params,
		Registration:	state.Registration,
		AllowedCallers:	[]AllowedCaller{{Caller: "x/treasury", Enabled: true}},
		Caps:		state.Caps,
	})
	require.ErrorContains(t, err, "normal emission caller must be x/emissions")

	_, err = ApplyUpdateMintAuthorityParams(state, MsgUpdateMintAuthorityParams{
		Authority:	"wrong",
		Params:		state.Params,
		Registration:	state.Registration,
		AllowedCallers:	state.AllowedCallers,
		Caps:		state.Caps,
	})
	require.ErrorContains(t, err, "requires authority")
}

func TestExportImportPreservesLifetimeMintedCounter(t *testing.T) {
	state := newMintAuthorityState(t)
	next := state
	for i := int64(0); i < 3; i++ {
		decision := emissionDecision(DefaultEmissionCaller, DefaultBaseDenom, 100+i, uint64(i+1), uint64(10+i), true)
		var err error
		next, _, err = ApplyMintProtocolCoins(next, mintMsg(DefaultEmissionCaller, DefaultBaseDenom, 100+i, uint64(i+1), uint64(10+i), decision.DecisionHash), decision, ConstitutionEmergencyAuthorization{})
		require.NoError(t, err)
	}
	exported, err := ExportMintAuthorityState(next)
	require.NoError(t, err)
	imported, err := ImportMintAuthorityState(exported)
	require.NoError(t, err)
	require.Equal(t, exported, imported)
	require.Equal(t, sdkmath.NewInt(303), QueryMintedLifetime(imported, DefaultBaseDenom).Amount)
	require.NoError(t, CheckMintAuthorityInvariants(imported))
}

func TestMintAuthorityInvariantsRejectTampering(t *testing.T) {
	state := newMintAuthorityState(t)
	decision := emissionDecision(DefaultEmissionCaller, DefaultBaseDenom, 100, 1, 10, true)
	next, event, err := ApplyMintProtocolCoins(state, mintMsg(DefaultEmissionCaller, DefaultBaseDenom, 100, 1, 10, decision.DecisionHash), decision, ConstitutionEmergencyAuthorization{})
	require.NoError(t, err)

	tampered := next
	tampered.MintedLifetime[0].Amount = sdkmath.NewInt(99)
	tampered.MintedLifetime[0].LifetimeHash = ComputeMintedLifetimeHash(tampered.MintedLifetime[0])
	require.ErrorContains(t, CheckMintAuthorityInvariants(tampered), "counter must equal mint events")

	tampered = next
	tampered.Events = append(tampered.Events, event)
	require.ErrorContains(t, CheckMintAuthorityInvariants(tampered), "duplicate mint event")

	tampered = next
	tampered.Caps = []MintCap{defaultCap(tampered.Params, 1_000, 10_000), {Denom: "uatom", EpochCap: sdkmath.NewInt(1), LifetimeCap: sdkmath.NewInt(1)}}
	tampered.Caps = normalizeCaps(tampered.Caps)
	require.ErrorContains(t, CheckMintAuthorityInvariants(tampered), "exactly one base denom cap")
}

func newMintAuthorityState(t *testing.T) MintAuthorityState {
	t.Helper()
	params := DefaultMintAuthorityParams()
	state, err := NewMintAuthorityState(params, DefaultMintAuthorityRegistration(), []MintCap{defaultCap(params, 1_000, 10_000)})
	require.NoError(t, err)
	return state
}

func mintAuthorityStateWithCaps(t *testing.T, epochCap int64, lifetimeCap int64) MintAuthorityState {
	t.Helper()
	params := DefaultMintAuthorityParams()
	state, err := NewMintAuthorityState(params, DefaultMintAuthorityRegistration(), []MintCap{defaultCap(params, epochCap, lifetimeCap)})
	require.NoError(t, err)
	return state
}

func emergencyMintAuthorityState(t *testing.T) MintAuthorityState {
	t.Helper()
	params := DefaultMintAuthorityParams()
	params.EmergencyMintingEnabled = true
	params.EmergencyCaller = "x/constitution"
	params.EmergencyConstitutionAuthority = "x/constitution"
	state, err := NewMintAuthorityState(params, DefaultMintAuthorityRegistration(), []MintCap{defaultCap(params, 1_000, 10_000)})
	require.NoError(t, err)
	state = NormalizeMintAuthorityState(state)
	require.NoError(t, CheckMintAuthorityInvariants(state))
	return state
}

func defaultCap(params MintAuthorityParams, epochCap int64, lifetimeCap int64) MintCap {
	cap := MintCap{Denom: params.BaseDenom, EpochCap: sdkmath.NewInt(epochCap), LifetimeCap: sdkmath.NewInt(lifetimeCap)}
	cap.CapHash = ComputeMintCapHash(cap)
	return cap
}

func mintMsg(caller string, denom string, amount int64, epoch uint64, height uint64, decisionHash string) MsgMintProtocolCoins {
	return MsgMintProtocolCoins{
		Caller:			caller,
		Recipient:		"module:fee-distributor",
		Denom:			denom,
		Amount:			sdkmath.NewInt(amount),
		Epoch:			epoch,
		Height:			height,
		EmissionsDecisionHash:	decisionHash,
	}
}

func emissionDecision(caller string, denom string, amount int64, epoch uint64, height uint64, approved bool) EmissionDecision {
	decision := EmissionDecision{
		Caller:		caller,
		Denom:		denom,
		Amount:		sdkmath.NewInt(amount),
		Epoch:		epoch,
		Height:		height,
		Approved:	approved,
	}
	decision.DecisionHash = ComputeEmissionDecisionHash(decision)
	return decision
}

func emergencyAuthorization(caller string, denom string, amount int64, epoch uint64, height uint64, enabled bool, bounded bool) ConstitutionEmergencyAuthorization {
	auth := ConstitutionEmergencyAuthorization{
		Caller:		caller,
		Denom:		denom,
		Amount:		sdkmath.NewInt(amount),
		Epoch:		epoch,
		Height:		height,
		Enabled:	enabled,
		Bounded:	bounded,
	}
	auth.AuthorizationHash = ComputeConstitutionEmergencyAuthorizationHash(auth)
	return auth
}
