package types

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"

	sdkmath "cosmossdk.io/math"
)

const (
	ModuleName	= "mint-authority"
	StoreKey	= "xmintauthority"

	DefaultMintAuthorityParamsAuthority	= "4:0000000000000000000000000000000000000000000000000000000000000001"
	DefaultMintAuthorityModuleAccount	= "mint-authority"
	DefaultMintAuthorityAlias		= "AETMint"
	DefaultMintAuthorityPurpose		= "base-denom emission only"
	DefaultBaseDenom			= "naet"
	DefaultEmissionCaller			= "x/emissions"
	BasisPoints				= uint32(10_000)
)

type MintAuthorityParams struct {
	Authority			string
	BaseDenom			string
	MinterModuleAccount		string
	MinterAlias			string
	NormalEmissionCaller		string
	EmergencyCaller			string
	EmergencyMintingEnabled		bool
	EmergencyConstitutionAuthority	string
	RequireSystemRegistry		bool
	MaxMintEvents			uint32
}

type SystemAccountRegistration struct {
	ModuleName	string
	Alias		string
	Purpose		string
	Registered	bool
	UserControlled	bool
}

type AllowedCaller struct {
	Caller		string
	Emergency	bool
	Enabled		bool
}

type MintCap struct {
	Denom		string
	EpochCap	sdkmath.Int
	LifetimeCap	sdkmath.Int
	CapHash		string
}

type MintedByEpoch struct {
	Epoch		uint64
	Denom		string
	Amount		sdkmath.Int
	EpochHash	string
}

type MintedLifetime struct {
	Denom		string
	Amount		sdkmath.Int
	LifetimeHash	string
}

type EmissionDecision struct {
	DecisionHash	string
	Caller		string
	Denom		string
	Amount		sdkmath.Int
	Epoch		uint64
	Height		uint64
	Approved	bool
}

type ConstitutionEmergencyAuthorization struct {
	AuthorizationHash	string
	Caller			string
	Denom			string
	Amount			sdkmath.Int
	Epoch			uint64
	Height			uint64
	Enabled			bool
	Bounded			bool
}

type MintEvent struct {
	EventID				string
	Caller				string
	Recipient			string
	Denom				string
	Amount				sdkmath.Int
	Epoch				uint64
	Height				uint64
	EmissionsDecisionHash		string
	Emergency			bool
	ConstitutionDecisionHash	string
	EventHash			string
}

type MintAuthorityState struct {
	Params		MintAuthorityParams
	Registration	SystemAccountRegistration
	AllowedCallers	[]AllowedCaller
	Caps		[]MintCap
	MintedByEpoch	[]MintedByEpoch
	MintedLifetime	[]MintedLifetime
	Events		[]MintEvent
}

type MsgMintProtocolCoins struct {
	Caller				string
	Recipient			string
	Denom				string
	Amount				sdkmath.Int
	Epoch				uint64
	Height				uint64
	EmissionsDecisionHash		string
	Emergency			bool
	ConstitutionDecisionHash	string
}

type MsgUpdateMintAuthorityParams struct {
	Authority	string
	Params		MintAuthorityParams
	Registration	SystemAccountRegistration
	AllowedCallers	[]AllowedCaller
	Caps		[]MintCap
}

type QueryMintedByEpochRequest struct {
	Epoch	uint64
	Denom	string
}

func DefaultMintAuthorityParams() MintAuthorityParams {
	return MintAuthorityParams{
		Authority:		DefaultMintAuthorityParamsAuthority,
		BaseDenom:		DefaultBaseDenom,
		MinterModuleAccount:	DefaultMintAuthorityModuleAccount,
		MinterAlias:		DefaultMintAuthorityAlias,
		NormalEmissionCaller:	DefaultEmissionCaller,
		RequireSystemRegistry:	true,
		MaxMintEvents:		4096,
	}
}

func DefaultMintAuthorityRegistration() SystemAccountRegistration {
	return SystemAccountRegistration{
		ModuleName:	DefaultMintAuthorityModuleAccount,
		Alias:		DefaultMintAuthorityAlias,
		Purpose:	DefaultMintAuthorityPurpose,
		Registered:	true,
	}
}

func DefaultAllowedMintCallers() []AllowedCaller {
	return []AllowedCaller{{Caller: DefaultEmissionCaller, Enabled: true}}
}

func defaultAllowedMintCallers(params MintAuthorityParams) []AllowedCaller {
	callers := DefaultAllowedMintCallers()
	params = normalizeParams(params)
	if params.EmergencyMintingEnabled && params.EmergencyCaller != "" {
		callers = append(callers, AllowedCaller{Caller: params.EmergencyCaller, Emergency: true, Enabled: true})
	}
	return callers
}

func NewMintAuthorityState(params MintAuthorityParams, registration SystemAccountRegistration, caps []MintCap) (MintAuthorityState, error) {
	if strings.TrimSpace(params.Authority) == "" {
		params = DefaultMintAuthorityParams()
	}
	if strings.TrimSpace(registration.ModuleName) == "" {
		registration = DefaultMintAuthorityRegistration()
	}
	if len(caps) == 0 {
		caps = []MintCap{{
			Denom:		params.BaseDenom,
			EpochCap:	sdkmath.NewInt(1_000_000),
			LifetimeCap:	sdkmath.NewInt(100_000_000),
		}}
	}
	state := MintAuthorityState{
		Params:		params,
		Registration:	registration,
		AllowedCallers:	defaultAllowedMintCallers(params),
		Caps:		caps,
	}
	state = NormalizeMintAuthorityState(state)
	if err := state.Validate(); err != nil {
		return MintAuthorityState{}, err
	}
	return state, nil
}

func (params MintAuthorityParams) Validate() error {
	params = normalizeParams(params)
	if params.Authority == "" {
		return errors.New("mint authority params authority is required")
	}
	if params.BaseDenom == "" {
		return errors.New("mint authority base denom is required")
	}
	if params.MinterModuleAccount != DefaultMintAuthorityModuleAccount {
		return errors.New("mint authority minter module account must be canonical")
	}
	if params.MinterAlias != DefaultMintAuthorityAlias {
		return errors.New("mint authority minter alias must be canonical")
	}
	if params.NormalEmissionCaller != DefaultEmissionCaller {
		return errors.New("mint authority normal emission caller must be x/emissions")
	}
	if params.MaxMintEvents == 0 {
		return errors.New("mint authority max mint events must be positive")
	}
	if params.EmergencyMintingEnabled && params.EmergencyCaller == "" {
		return errors.New("mint authority emergency caller is required when emergency minting is enabled")
	}
	if params.EmergencyMintingEnabled && params.EmergencyConstitutionAuthority == "" {
		return errors.New("mint authority constitutional emergency authority is required when emergency minting is enabled")
	}
	return nil
}

func (registration SystemAccountRegistration) Validate(params MintAuthorityParams) error {
	params = normalizeParams(params)
	registration = normalizeRegistration(registration)
	if params.RequireSystemRegistry && !registration.Registered {
		return errors.New("mint authority minter account must be registered in system-registry")
	}
	if registration.ModuleName != params.MinterModuleAccount {
		return errors.New("mint authority system-registry module account mismatch")
	}
	if registration.Alias != params.MinterAlias {
		return errors.New("mint authority system-registry alias mismatch")
	}
	if registration.Purpose != DefaultMintAuthorityPurpose {
		return errors.New("mint authority system-registry purpose mismatch")
	}
	if registration.UserControlled {
		return errors.New("mint authority minter account cannot be user controlled")
	}
	return nil
}

func (cap MintCap) Validate(params MintAuthorityParams) error {
	params = normalizeParams(params)
	cap = normalizeCap(cap)
	if cap.Denom != params.BaseDenom {
		return errors.New("mint authority cap denom must be base denom")
	}
	if !cap.EpochCap.IsPositive() {
		return errors.New("mint authority epoch cap must be positive")
	}
	if !cap.LifetimeCap.IsPositive() {
		return errors.New("mint authority lifetime cap must be positive")
	}
	if cap.EpochCap.GT(cap.LifetimeCap) {
		return errors.New("mint authority epoch cap cannot exceed lifetime cap")
	}
	if cap.CapHash != ComputeMintCapHash(cap) {
		return errors.New("mint authority cap hash mismatch")
	}
	return nil
}

func (event MintEvent) Validate(params MintAuthorityParams) error {
	params = normalizeParams(params)
	event = normalizeEvent(event)
	if event.EventID == "" {
		return errors.New("mint authority event id is required")
	}
	if event.Caller == "" {
		return errors.New("mint authority event caller is required")
	}
	if event.Recipient == "" {
		return errors.New("mint authority event recipient is required")
	}
	if event.Denom != params.BaseDenom {
		return errors.New("mint authority event denom must be base denom")
	}
	if !event.Amount.IsPositive() {
		return errors.New("mint authority event amount must be positive")
	}
	if !event.Emergency && event.EmissionsDecisionHash == "" {
		return errors.New("mint authority scheduled event requires emissions decision")
	}
	if event.Emergency && event.ConstitutionDecisionHash == "" {
		return errors.New("mint authority emergency event requires constitutional decision")
	}
	if event.EventHash != ComputeMintEventHash(event) {
		return errors.New("mint authority event hash mismatch")
	}
	return nil
}

func (state MintAuthorityState) Validate() error {
	state = NormalizeMintAuthorityState(state)
	if err := state.Params.Validate(); err != nil {
		return err
	}
	if err := state.Registration.Validate(state.Params); err != nil {
		return err
	}
	if len(state.Caps) != 1 {
		return errors.New("mint authority must have exactly one base denom cap")
	}
	for _, cap := range state.Caps {
		if err := cap.Validate(state.Params); err != nil {
			return err
		}
	}
	if len(state.Events) > int(state.Params.MaxMintEvents) {
		return errors.New("mint authority mint event history exceeds configured max")
	}
	seenCallers := make(map[string]struct{}, len(state.AllowedCallers))
	normalCallerEnabled := false
	emergencyCallerEnabled := !state.Params.EmergencyMintingEnabled
	for _, caller := range state.AllowedCallers {
		caller = normalizeAllowedCaller(caller)
		if caller.Caller == "" {
			return errors.New("mint authority allowed caller is required")
		}
		key := fmt.Sprintf("%s/%t", caller.Caller, caller.Emergency)
		if _, ok := seenCallers[key]; ok {
			return errors.New("mint authority duplicate allowed caller")
		}
		seenCallers[key] = struct{}{}
		if caller.Enabled && !caller.Emergency && caller.Caller != state.Params.NormalEmissionCaller {
			return errors.New("mint authority only x/emissions can be normal mint caller")
		}
		if caller.Enabled && !caller.Emergency && caller.Caller == state.Params.NormalEmissionCaller {
			normalCallerEnabled = true
		}
		if caller.Enabled && caller.Emergency && caller.Caller == state.Params.EmergencyCaller {
			emergencyCallerEnabled = true
		}
	}
	if !normalCallerEnabled {
		return errors.New("mint authority x/emissions caller must be enabled")
	}
	if !emergencyCallerEnabled {
		return errors.New("mint authority emergency caller must be allowlisted")
	}
	seenEpochs := make(map[string]struct{}, len(state.MintedByEpoch))
	for _, minted := range state.MintedByEpoch {
		minted = normalizeMintedByEpoch(minted)
		if minted.Denom != state.Params.BaseDenom {
			return errors.New("mint authority epoch minted denom must be base denom")
		}
		if minted.Amount.IsNegative() {
			return errors.New("mint authority epoch minted amount cannot be negative")
		}
		if minted.EpochHash != ComputeMintedByEpochHash(minted) {
			return errors.New("mint authority epoch minted hash mismatch")
		}
		key := fmt.Sprintf("%020d/%s", minted.Epoch, minted.Denom)
		if _, ok := seenEpochs[key]; ok {
			return errors.New("mint authority duplicate epoch minted counter")
		}
		seenEpochs[key] = struct{}{}
		if minted.Amount.GT(state.baseCap().EpochCap) {
			return errors.New("mint authority epoch minted amount exceeds cap")
		}
	}
	if len(state.MintedLifetime) != 1 {
		return errors.New("mint authority must track exactly one lifetime counter")
	}
	lifetime := state.MintedLifetime[0]
	if lifetime.Denom != state.Params.BaseDenom {
		return errors.New("mint authority lifetime minted denom must be base denom")
	}
	if lifetime.Amount.IsNegative() {
		return errors.New("mint authority lifetime minted amount cannot be negative")
	}
	if lifetime.Amount.GT(state.baseCap().LifetimeCap) {
		return errors.New("mint authority lifetime minted amount exceeds cap")
	}
	if lifetime.LifetimeHash != ComputeMintedLifetimeHash(lifetime) {
		return errors.New("mint authority lifetime minted hash mismatch")
	}
	sum := sdkmath.ZeroInt()
	seenEvents := make(map[string]struct{}, len(state.Events))
	for _, event := range state.Events {
		if err := event.Validate(state.Params); err != nil {
			return err
		}
		if _, ok := seenEvents[event.EventID]; ok {
			return errors.New("mint authority duplicate mint event")
		}
		seenEvents[event.EventID] = struct{}{}
		sum = sum.Add(event.Amount)
	}
	if !sum.Equal(lifetime.Amount) {
		return errors.New("mint authority lifetime minted counter must equal mint events")
	}
	return nil
}

func CheckMintAuthorityInvariants(state MintAuthorityState) error {
	return state.Validate()
}

func ApplyMintProtocolCoins(state MintAuthorityState, msg MsgMintProtocolCoins, decision EmissionDecision, emergencyAuth ConstitutionEmergencyAuthorization) (MintAuthorityState, MintEvent, error) {
	next := cloneState(state)
	if err := next.Validate(); err != nil {
		return MintAuthorityState{}, MintEvent{}, err
	}
	msg = normalizeMintMsg(msg)
	if msg.Caller == "" {
		return MintAuthorityState{}, MintEvent{}, errors.New("mint authority mint caller is required")
	}
	if msg.Recipient == "" {
		return MintAuthorityState{}, MintEvent{}, errors.New("mint authority mint recipient is required")
	}
	if msg.Denom != next.Params.BaseDenom {
		return MintAuthorityState{}, MintEvent{}, errors.New("mint authority minted denom must be base denom")
	}
	if !msg.Amount.IsPositive() {
		return MintAuthorityState{}, MintEvent{}, errors.New("mint authority mint amount must be positive")
	}
	if msg.Emergency {
		if err := authorizeEmergencyMint(next, msg, emergencyAuth); err != nil {
			return MintAuthorityState{}, MintEvent{}, err
		}
	} else {
		if err := authorizeScheduledMint(next, msg, decision); err != nil {
			return MintAuthorityState{}, MintEvent{}, err
		}
	}
	epochCounter := next.epochCounter(msg.Epoch)
	lifetimeCounter := next.lifetimeCounter()
	if epochCounter.Amount.Add(msg.Amount).GT(next.baseCap().EpochCap) {
		return MintAuthorityState{}, MintEvent{}, errors.New("mint authority minted amount exceeds epoch cap")
	}
	if lifetimeCounter.Amount.Add(msg.Amount).GT(next.baseCap().LifetimeCap) {
		return MintAuthorityState{}, MintEvent{}, errors.New("mint authority minted amount exceeds lifetime cap")
	}
	event := MintEvent{
		Caller:				msg.Caller,
		Recipient:			msg.Recipient,
		Denom:				msg.Denom,
		Amount:				msg.Amount,
		Epoch:				msg.Epoch,
		Height:				msg.Height,
		EmissionsDecisionHash:		msg.EmissionsDecisionHash,
		Emergency:			msg.Emergency,
		ConstitutionDecisionHash:	msg.ConstitutionDecisionHash,
	}
	event.EventID = ComputeMintEventID(event)
	event.EventHash = ComputeMintEventHash(event)
	next.Events = append(next.Events, event)
	next.setEpochCounter(MintedByEpoch{
		Epoch:	msg.Epoch,
		Denom:	msg.Denom,
		Amount:	epochCounter.Amount.Add(msg.Amount),
	})
	next.setLifetimeCounter(MintedLifetime{
		Denom:	msg.Denom,
		Amount:	lifetimeCounter.Amount.Add(msg.Amount),
	})
	next = NormalizeMintAuthorityState(next)
	if err := next.Validate(); err != nil {
		return MintAuthorityState{}, MintEvent{}, err
	}
	return next, event, nil
}

func ApplyUpdateMintAuthorityParams(state MintAuthorityState, msg MsgUpdateMintAuthorityParams) (MintAuthorityState, error) {
	next := cloneState(state)
	if err := next.Validate(); err != nil {
		return MintAuthorityState{}, err
	}
	if strings.TrimSpace(msg.Authority) != next.Params.Authority {
		return MintAuthorityState{}, errors.New("mint authority params update requires authority")
	}
	next.Params = msg.Params
	next.Registration = msg.Registration
	next.AllowedCallers = msg.AllowedCallers
	next.Caps = msg.Caps
	next = NormalizeMintAuthorityState(next)
	if err := next.Validate(); err != nil {
		return MintAuthorityState{}, err
	}
	return next, nil
}

func QueryMintAuthority(state MintAuthorityState) (MintAuthorityParams, SystemAccountRegistration, []AllowedCaller) {
	state = NormalizeMintAuthorityState(state)
	return state.Params, state.Registration, append([]AllowedCaller(nil), state.AllowedCallers...)
}

func QueryMintedByEpoch(state MintAuthorityState, req QueryMintedByEpochRequest) MintedByEpoch {
	state = NormalizeMintAuthorityState(state)
	req.Denom = strings.TrimSpace(req.Denom)
	if req.Denom == "" {
		req.Denom = state.Params.BaseDenom
	}
	for _, minted := range state.MintedByEpoch {
		if minted.Epoch == req.Epoch && minted.Denom == req.Denom {
			return minted
		}
	}
	return MintedByEpoch{Epoch: req.Epoch, Denom: req.Denom, Amount: sdkmath.ZeroInt(), EpochHash: ComputeMintedByEpochHash(MintedByEpoch{Epoch: req.Epoch, Denom: req.Denom})}
}

func QueryMintedLifetime(state MintAuthorityState, denom string) MintedLifetime {
	state = NormalizeMintAuthorityState(state)
	denom = strings.TrimSpace(denom)
	if denom == "" {
		denom = state.Params.BaseDenom
	}
	for _, minted := range state.MintedLifetime {
		if minted.Denom == denom {
			return minted
		}
	}
	return MintedLifetime{Denom: denom, Amount: sdkmath.ZeroInt(), LifetimeHash: ComputeMintedLifetimeHash(MintedLifetime{Denom: denom})}
}

func QueryMintCaps(state MintAuthorityState) []MintCap {
	state = NormalizeMintAuthorityState(state)
	return append([]MintCap(nil), state.Caps...)
}

func ExportMintAuthorityState(state MintAuthorityState) (MintAuthorityState, error) {
	state = cloneState(state)
	if err := state.Validate(); err != nil {
		return MintAuthorityState{}, err
	}
	return state, nil
}

func ImportMintAuthorityState(state MintAuthorityState) (MintAuthorityState, error) {
	state = NormalizeMintAuthorityState(state)
	if err := state.Validate(); err != nil {
		return MintAuthorityState{}, err
	}
	return cloneState(state), nil
}

func authorizeScheduledMint(state MintAuthorityState, msg MsgMintProtocolCoins, decision EmissionDecision) error {
	if msg.Caller != state.Params.NormalEmissionCaller {
		return errors.New("mint authority scheduled mint caller must be x/emissions")
	}
	if !state.isCallerAllowed(msg.Caller, false) {
		return errors.New("mint authority mint caller is not authorized")
	}
	decision = normalizeDecision(decision)
	if !decision.Approved {
		return errors.New("mint authority emissions decision is not approved")
	}
	expectedHash := ComputeEmissionDecisionHash(decision)
	if msg.EmissionsDecisionHash != expectedHash || decision.DecisionHash != expectedHash {
		return errors.New("mint authority emissions decision hash mismatch")
	}
	if decision.Caller != msg.Caller || decision.Denom != msg.Denom || !decision.Amount.Equal(msg.Amount) || decision.Epoch != msg.Epoch {
		return errors.New("mint authority mint amount must match emissions decision")
	}
	if msg.ConstitutionDecisionHash != "" {
		return errors.New("mint authority scheduled mint cannot carry constitutional emergency decision")
	}
	return nil
}

func authorizeEmergencyMint(state MintAuthorityState, msg MsgMintProtocolCoins, emergencyAuth ConstitutionEmergencyAuthorization) error {
	if !state.Params.EmergencyMintingEnabled {
		return errors.New("mint authority emergency minting is disabled")
	}
	if !state.isCallerAllowed(msg.Caller, true) {
		return errors.New("mint authority emergency caller is not allowlisted")
	}
	emergencyAuth = normalizeEmergencyAuth(emergencyAuth)
	if !emergencyAuth.Enabled || !emergencyAuth.Bounded {
		return errors.New("mint authority emergency minting must be bounded by constitution")
	}
	expectedHash := ComputeConstitutionEmergencyAuthorizationHash(emergencyAuth)
	if msg.ConstitutionDecisionHash != expectedHash || emergencyAuth.AuthorizationHash != expectedHash {
		return errors.New("mint authority constitutional emergency decision hash mismatch")
	}
	if emergencyAuth.Caller != msg.Caller || emergencyAuth.Denom != msg.Denom || !emergencyAuth.Amount.Equal(msg.Amount) || emergencyAuth.Epoch != msg.Epoch {
		return errors.New("mint authority emergency mint amount must match constitutional authorization")
	}
	if msg.EmissionsDecisionHash != "" {
		return errors.New("mint authority emergency mint cannot carry scheduled emissions decision")
	}
	return nil
}

func NormalizeMintAuthorityState(state MintAuthorityState) MintAuthorityState {
	state.Params = normalizeParams(state.Params)
	state.Registration = normalizeRegistration(state.Registration)
	state.AllowedCallers = normalizeAllowedCallers(state.AllowedCallers)
	state.Caps = normalizeCaps(state.Caps)
	if len(state.MintedLifetime) == 0 {
		state.MintedLifetime = []MintedLifetime{{Denom: state.Params.BaseDenom}}
	}
	state.MintedByEpoch = normalizeEpochCounters(state.MintedByEpoch)
	state.MintedLifetime = normalizeLifetimeCounters(state.MintedLifetime)
	state.Events = normalizeEvents(state.Events)
	return state
}

func (state MintAuthorityState) isCallerAllowed(caller string, emergency bool) bool {
	caller = strings.TrimSpace(caller)
	for _, allowed := range state.AllowedCallers {
		allowed = normalizeAllowedCaller(allowed)
		if allowed.Enabled && allowed.Caller == caller && allowed.Emergency == emergency {
			return true
		}
	}
	return false
}

func (state MintAuthorityState) baseCap() MintCap {
	for _, cap := range state.Caps {
		if cap.Denom == state.Params.BaseDenom {
			return cap
		}
	}
	return MintCap{Denom: state.Params.BaseDenom, EpochCap: sdkmath.ZeroInt(), LifetimeCap: sdkmath.ZeroInt(), CapHash: ComputeMintCapHash(MintCap{Denom: state.Params.BaseDenom})}
}

func (state MintAuthorityState) epochCounter(epoch uint64) MintedByEpoch {
	for _, counter := range state.MintedByEpoch {
		if counter.Epoch == epoch && counter.Denom == state.Params.BaseDenom {
			return counter
		}
	}
	return MintedByEpoch{Epoch: epoch, Denom: state.Params.BaseDenom, Amount: sdkmath.ZeroInt(), EpochHash: ComputeMintedByEpochHash(MintedByEpoch{Epoch: epoch, Denom: state.Params.BaseDenom})}
}

func (state MintAuthorityState) lifetimeCounter() MintedLifetime {
	for _, counter := range state.MintedLifetime {
		if counter.Denom == state.Params.BaseDenom {
			return counter
		}
	}
	return MintedLifetime{Denom: state.Params.BaseDenom, Amount: sdkmath.ZeroInt(), LifetimeHash: ComputeMintedLifetimeHash(MintedLifetime{Denom: state.Params.BaseDenom})}
}

func (state *MintAuthorityState) setEpochCounter(counter MintedByEpoch) {
	counter = normalizeMintedByEpoch(counter)
	for i, existing := range state.MintedByEpoch {
		if existing.Epoch == counter.Epoch && existing.Denom == counter.Denom {
			state.MintedByEpoch[i] = counter
			return
		}
	}
	state.MintedByEpoch = append(state.MintedByEpoch, counter)
}

func (state *MintAuthorityState) setLifetimeCounter(counter MintedLifetime) {
	counter = normalizeMintedLifetime(counter)
	for i, existing := range state.MintedLifetime {
		if existing.Denom == counter.Denom {
			state.MintedLifetime[i] = counter
			return
		}
	}
	state.MintedLifetime = append(state.MintedLifetime, counter)
}

func cloneState(state MintAuthorityState) MintAuthorityState {
	state = NormalizeMintAuthorityState(state)
	return MintAuthorityState{
		Params:		state.Params,
		Registration:	state.Registration,
		AllowedCallers:	append([]AllowedCaller(nil), state.AllowedCallers...),
		Caps:		append([]MintCap(nil), state.Caps...),
		MintedByEpoch:	append([]MintedByEpoch(nil), state.MintedByEpoch...),
		MintedLifetime:	append([]MintedLifetime(nil), state.MintedLifetime...),
		Events:		append([]MintEvent(nil), state.Events...),
	}
}

func normalizeParams(params MintAuthorityParams) MintAuthorityParams {
	if strings.TrimSpace(params.Authority) == "" {
		params = DefaultMintAuthorityParams()
	}
	params.Authority = strings.TrimSpace(params.Authority)
	params.BaseDenom = strings.TrimSpace(params.BaseDenom)
	params.MinterModuleAccount = strings.TrimSpace(params.MinterModuleAccount)
	params.MinterAlias = strings.TrimSpace(params.MinterAlias)
	params.NormalEmissionCaller = strings.TrimSpace(params.NormalEmissionCaller)
	params.EmergencyCaller = strings.TrimSpace(params.EmergencyCaller)
	params.EmergencyConstitutionAuthority = strings.TrimSpace(params.EmergencyConstitutionAuthority)
	return params
}

func normalizeRegistration(registration SystemAccountRegistration) SystemAccountRegistration {
	registration.ModuleName = strings.TrimSpace(registration.ModuleName)
	registration.Alias = strings.TrimSpace(registration.Alias)
	registration.Purpose = strings.TrimSpace(registration.Purpose)
	return registration
}

func normalizeAllowedCaller(caller AllowedCaller) AllowedCaller {
	caller.Caller = strings.TrimSpace(caller.Caller)
	return caller
}

func normalizeAllowedCallers(callers []AllowedCaller) []AllowedCaller {
	out := make([]AllowedCaller, len(callers))
	for i, caller := range callers {
		out[i] = normalizeAllowedCaller(caller)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Emergency != out[j].Emergency {
			return !out[i].Emergency
		}
		return out[i].Caller < out[j].Caller
	})
	return out
}

func normalizeCap(cap MintCap) MintCap {
	cap.Denom = strings.TrimSpace(cap.Denom)
	cap.EpochCap = normalizeInt(cap.EpochCap)
	cap.LifetimeCap = normalizeInt(cap.LifetimeCap)
	cap.CapHash = strings.ToLower(strings.TrimSpace(cap.CapHash))
	if cap.CapHash == "" {
		cap.CapHash = ComputeMintCapHash(cap)
	}
	return cap
}

func normalizeCaps(caps []MintCap) []MintCap {
	out := make([]MintCap, len(caps))
	for i, cap := range caps {
		out[i] = normalizeCap(cap)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Denom < out[j].Denom })
	return out
}

func normalizeMintedByEpoch(minted MintedByEpoch) MintedByEpoch {
	minted.Denom = strings.TrimSpace(minted.Denom)
	minted.Amount = normalizeInt(minted.Amount)
	minted.EpochHash = strings.ToLower(strings.TrimSpace(minted.EpochHash))
	if minted.EpochHash == "" {
		minted.EpochHash = ComputeMintedByEpochHash(minted)
	}
	return minted
}

func normalizeEpochCounters(counters []MintedByEpoch) []MintedByEpoch {
	out := make([]MintedByEpoch, len(counters))
	for i, counter := range counters {
		out[i] = normalizeMintedByEpoch(counter)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Epoch != out[j].Epoch {
			return out[i].Epoch < out[j].Epoch
		}
		return out[i].Denom < out[j].Denom
	})
	return out
}

func normalizeMintedLifetime(minted MintedLifetime) MintedLifetime {
	minted.Denom = strings.TrimSpace(minted.Denom)
	minted.Amount = normalizeInt(minted.Amount)
	minted.LifetimeHash = strings.ToLower(strings.TrimSpace(minted.LifetimeHash))
	if minted.LifetimeHash == "" {
		minted.LifetimeHash = ComputeMintedLifetimeHash(minted)
	}
	return minted
}

func normalizeLifetimeCounters(counters []MintedLifetime) []MintedLifetime {
	out := make([]MintedLifetime, len(counters))
	for i, counter := range counters {
		out[i] = normalizeMintedLifetime(counter)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Denom < out[j].Denom })
	return out
}

func normalizeDecision(decision EmissionDecision) EmissionDecision {
	decision.DecisionHash = strings.ToLower(strings.TrimSpace(decision.DecisionHash))
	decision.Caller = strings.TrimSpace(decision.Caller)
	decision.Denom = strings.TrimSpace(decision.Denom)
	decision.Amount = normalizeInt(decision.Amount)
	if decision.DecisionHash == "" {
		decision.DecisionHash = ComputeEmissionDecisionHash(decision)
	}
	return decision
}

func normalizeEmergencyAuth(auth ConstitutionEmergencyAuthorization) ConstitutionEmergencyAuthorization {
	auth.AuthorizationHash = strings.ToLower(strings.TrimSpace(auth.AuthorizationHash))
	auth.Caller = strings.TrimSpace(auth.Caller)
	auth.Denom = strings.TrimSpace(auth.Denom)
	auth.Amount = normalizeInt(auth.Amount)
	if auth.AuthorizationHash == "" {
		auth.AuthorizationHash = ComputeConstitutionEmergencyAuthorizationHash(auth)
	}
	return auth
}

func normalizeMintMsg(msg MsgMintProtocolCoins) MsgMintProtocolCoins {
	msg.Caller = strings.TrimSpace(msg.Caller)
	msg.Recipient = strings.TrimSpace(msg.Recipient)
	msg.Denom = strings.TrimSpace(msg.Denom)
	msg.Amount = normalizeInt(msg.Amount)
	msg.EmissionsDecisionHash = strings.ToLower(strings.TrimSpace(msg.EmissionsDecisionHash))
	msg.ConstitutionDecisionHash = strings.ToLower(strings.TrimSpace(msg.ConstitutionDecisionHash))
	return msg
}

func normalizeEvent(event MintEvent) MintEvent {
	event.EventID = strings.ToLower(strings.TrimSpace(event.EventID))
	event.Caller = strings.TrimSpace(event.Caller)
	event.Recipient = strings.TrimSpace(event.Recipient)
	event.Denom = strings.TrimSpace(event.Denom)
	event.Amount = normalizeInt(event.Amount)
	event.EmissionsDecisionHash = strings.ToLower(strings.TrimSpace(event.EmissionsDecisionHash))
	event.ConstitutionDecisionHash = strings.ToLower(strings.TrimSpace(event.ConstitutionDecisionHash))
	event.EventHash = strings.ToLower(strings.TrimSpace(event.EventHash))
	if event.EventID == "" {
		event.EventID = ComputeMintEventID(event)
	}
	if event.EventHash == "" {
		event.EventHash = ComputeMintEventHash(event)
	}
	return event
}

func normalizeEvents(events []MintEvent) []MintEvent {
	out := make([]MintEvent, len(events))
	for i, event := range events {
		out[i] = normalizeEvent(event)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Height != out[j].Height {
			return out[i].Height < out[j].Height
		}
		return out[i].EventID < out[j].EventID
	})
	return out
}

func ComputeMintCapHash(cap MintCap) string {
	cap = normalizeCapForHash(cap)
	return mintAuthorityHashParts("mint-authority-cap-v1", cap.Denom, cap.EpochCap.String(), cap.LifetimeCap.String())
}

func ComputeMintedByEpochHash(minted MintedByEpoch) string {
	minted = normalizeMintedByEpochForHash(minted)
	return mintAuthorityHashParts("mint-authority-epoch-v1", fmt.Sprint(minted.Epoch), minted.Denom, minted.Amount.String())
}

func ComputeMintedLifetimeHash(minted MintedLifetime) string {
	minted = normalizeMintedLifetimeForHash(minted)
	return mintAuthorityHashParts("mint-authority-lifetime-v1", minted.Denom, minted.Amount.String())
}

func ComputeEmissionDecisionHash(decision EmissionDecision) string {
	decision = normalizeDecisionForHash(decision)
	return mintAuthorityHashParts("mint-authority-emission-decision-v1", decision.Caller, decision.Denom, decision.Amount.String(), fmt.Sprint(decision.Epoch), fmt.Sprint(decision.Height), fmt.Sprint(decision.Approved))
}

func ComputeConstitutionEmergencyAuthorizationHash(auth ConstitutionEmergencyAuthorization) string {
	auth = normalizeEmergencyAuthForHash(auth)
	return mintAuthorityHashParts("mint-authority-constitution-emergency-v1", auth.Caller, auth.Denom, auth.Amount.String(), fmt.Sprint(auth.Epoch), fmt.Sprint(auth.Height), fmt.Sprint(auth.Enabled), fmt.Sprint(auth.Bounded))
}

func ComputeMintEventID(event MintEvent) string {
	event = normalizeEventForHash(event)
	return mintAuthorityHashParts("mint-authority-event-id-v1", event.Caller, event.Recipient, event.Denom, event.Amount.String(), fmt.Sprint(event.Epoch), fmt.Sprint(event.Height), event.EmissionsDecisionHash, fmt.Sprint(event.Emergency), event.ConstitutionDecisionHash)
}

func ComputeMintEventHash(event MintEvent) string {
	event = normalizeEventForHash(event)
	return mintAuthorityHashParts("mint-authority-event-v1", event.EventID, event.Caller, event.Recipient, event.Denom, event.Amount.String(), fmt.Sprint(event.Epoch), fmt.Sprint(event.Height), event.EmissionsDecisionHash, fmt.Sprint(event.Emergency), event.ConstitutionDecisionHash)
}

func normalizeCapForHash(cap MintCap) MintCap {
	cap.Denom = strings.TrimSpace(cap.Denom)
	cap.EpochCap = normalizeInt(cap.EpochCap)
	cap.LifetimeCap = normalizeInt(cap.LifetimeCap)
	cap.CapHash = ""
	return cap
}

func normalizeMintedByEpochForHash(minted MintedByEpoch) MintedByEpoch {
	minted.Denom = strings.TrimSpace(minted.Denom)
	minted.Amount = normalizeInt(minted.Amount)
	minted.EpochHash = ""
	return minted
}

func normalizeMintedLifetimeForHash(minted MintedLifetime) MintedLifetime {
	minted.Denom = strings.TrimSpace(minted.Denom)
	minted.Amount = normalizeInt(minted.Amount)
	minted.LifetimeHash = ""
	return minted
}

func normalizeDecisionForHash(decision EmissionDecision) EmissionDecision {
	decision.DecisionHash = ""
	decision.Caller = strings.TrimSpace(decision.Caller)
	decision.Denom = strings.TrimSpace(decision.Denom)
	decision.Amount = normalizeInt(decision.Amount)
	return decision
}

func normalizeEmergencyAuthForHash(auth ConstitutionEmergencyAuthorization) ConstitutionEmergencyAuthorization {
	auth.AuthorizationHash = ""
	auth.Caller = strings.TrimSpace(auth.Caller)
	auth.Denom = strings.TrimSpace(auth.Denom)
	auth.Amount = normalizeInt(auth.Amount)
	return auth
}

func normalizeEventForHash(event MintEvent) MintEvent {
	event.EventID = strings.ToLower(strings.TrimSpace(event.EventID))
	event.Caller = strings.TrimSpace(event.Caller)
	event.Recipient = strings.TrimSpace(event.Recipient)
	event.Denom = strings.TrimSpace(event.Denom)
	event.Amount = normalizeInt(event.Amount)
	event.EmissionsDecisionHash = strings.ToLower(strings.TrimSpace(event.EmissionsDecisionHash))
	event.ConstitutionDecisionHash = strings.ToLower(strings.TrimSpace(event.ConstitutionDecisionHash))
	event.EventHash = ""
	return event
}

func mintAuthorityHashParts(parts ...string) string {
	h := sha256.New()
	for _, part := range parts {
		data := []byte(part)
		var lenBuf [8]byte
		for i := uint(0); i < 8; i++ {
			lenBuf[7-i] = byte(uint64(len(data)) >> (i * 8))
		}
		h.Write(lenBuf[:])
		h.Write(data)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func normalizeInt(value sdkmath.Int) sdkmath.Int {
	if value.IsNil() {
		return sdkmath.ZeroInt()
	}
	return value
}
