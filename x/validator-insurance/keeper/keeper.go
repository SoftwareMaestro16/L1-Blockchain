package keeper

import (
	"context"
	"errors"
	"math"
	"strings"

	corestore "cosmossdk.io/core/store"

	"github.com/sovereign-l1/l1/app/addressing"
	"github.com/sovereign-l1/l1/x/internal/prefixgenesis"
	"github.com/sovereign-l1/l1/x/internal/prototype"
	"github.com/sovereign-l1/l1/x/validator-insurance/types"
	validatorregistrytypes "github.com/sovereign-l1/l1/x/validator-registry/types"
)

var genesisKey = []byte{0x01}

type GenesisState struct {
	Version	uint64
	Params	types.Params
	State	types.State
}

type Keeper struct {
	genesis		GenesisState
	storeService	corestore.KVStoreService
	runtimeCtx	context.Context
}

func NewKeeper() Keeper	{ return Keeper{genesis: DefaultGenesis()} }

func NewPersistentKeeper(storeService corestore.KVStoreService) Keeper {
	return Keeper{genesis: DefaultGenesis(), storeService: storeService}
}

func DefaultGenesis() GenesisState {
	params := types.DefaultParams()
	return GenesisState{Version: prototype.CurrentGenesisVersion, Params: params, State: types.DefaultState(params)}
}

func (gs GenesisState) Validate() error {
	if gs.Version != prototype.CurrentGenesisVersion {
		return errors.New("validator insurance unsupported genesis version")
	}
	return gs.State.Validate(gs.Params)
}

func (k *Keeper) InitGenesis(gs GenesisState) error {
	if err := gs.Validate(); err != nil {
		return err
	}
	k.genesis = cloneGenesis(gs)
	return nil
}

func (k *Keeper) InitGenesisState(ctx context.Context, gs GenesisState) error {
	if err := gs.Validate(); err != nil {
		return err
	}
	k.genesis = cloneGenesis(gs)
	k.runtimeCtx = ctx
	if k.storeService == nil {
		return nil
	}
	return prefixgenesis.Save(ctx, k.storeService, genesisKey, k.genesis)
}

func (k Keeper) ExportGenesis() GenesisState	{ return cloneGenesis(k.genesis) }

func (k Keeper) ExportGenesisState(ctx context.Context) (GenesisState, error) {
	if k.storeService == nil {
		return k.ExportGenesis(), nil
	}
	gs, _, err := prefixgenesis.Load(ctx, k.storeService, genesisKey, DefaultGenesis())
	if err != nil {
		return GenesisState{}, err
	}
	if err := gs.Validate(); err != nil {
		return GenesisState{}, err
	}
	return cloneGenesis(gs), nil
}

func (k *Keeper) FundValidatorInsurance(msg types.MsgFundValidatorInsurance) (types.ValidatorInsurance, error) {
	if err := k.genesis.Params.Authorize(msg.Authority); err != nil {
		return types.ValidatorInsurance{}, err
	}
	if msg.Amount == 0 || msg.Height == 0 {
		return types.ValidatorInsurance{}, errors.New("validator insurance fund amount and height must be positive")
	}
	if err := validateActor("validator insurance funder", msg.Funder); err != nil {
		return types.ValidatorInsurance{}, err
	}
	idx, insurance, found := findInsurance(k.genesis.State.Insurances, msg.ValidatorAddress)
	if !found {
		insurance = types.ValidatorInsurance{ValidatorAddress: msg.ValidatorAddress, ValidatorStatus: validatorregistrytypes.StatusCandidate}
		idx = -1
	}
	if math.MaxUint64-insurance.Balance < msg.Amount {
		return types.ValidatorInsurance{}, errors.New("validator insurance funding would overflow balance")
	}
	insurance.Balance += msg.Amount
	if idx == -1 {
		next := cloneGenesis(k.genesis)
		next.State.Insurances = append(next.State.Insurances, insurance)
		if err := k.saveGenesis(next); err != nil {
			return types.ValidatorInsurance{}, err
		}
		return insurance, nil
	}
	return k.saveInsurance(idx, insurance)
}

func (k *Keeper) WithdrawValidatorInsurance(msg types.MsgWithdrawValidatorInsurance) (types.PendingInsuranceWithdrawal, error) {
	if err := k.genesis.Params.Authorize(msg.Authority); err != nil {
		return types.PendingInsuranceWithdrawal{}, err
	}
	if msg.Height == 0 {
		return types.PendingInsuranceWithdrawal{}, errors.New("validator insurance withdrawal height must be positive")
	}
	idx, insurance, found := findInsurance(k.genesis.State.Insurances, msg.ValidatorAddress)
	if !found {
		return types.PendingInsuranceWithdrawal{}, errors.New("validator insurance not found")
	}
	if msg.ValidatorStatus != "" {
		insurance.ValidatorStatus = msg.ValidatorStatus
	}
	if insurance.PendingWithdrawal.Status == types.WithdrawalStatusPending {
		if msg.Height < insurance.PendingWithdrawal.CompleteHeight {
			return types.PendingInsuranceWithdrawal{}, errors.New("validator insurance withdrawal lock period has not completed")
		}
		insurance.PendingWithdrawal.Status = types.WithdrawalStatusCompleted
		completed := insurance.PendingWithdrawal
		insurance.PendingWithdrawal = types.PendingInsuranceWithdrawal{}
		if _, err := k.saveInsurance(idx, insurance); err != nil {
			return types.PendingInsuranceWithdrawal{}, err
		}
		return completed, nil
	}
	if msg.Amount == 0 || msg.Amount > insurance.Balance {
		return types.PendingInsuranceWithdrawal{}, errors.New("validator insurance withdrawal amount exceeds balance")
	}
	if err := validateActor("validator insurance withdrawal recipient", msg.Recipient); err != nil {
		return types.PendingInsuranceWithdrawal{}, err
	}
	remaining := insurance.Balance - msg.Amount
	if k.genesis.Params.Enabled && msg.ValidatorStatus == validatorregistrytypes.StatusActive && remaining < k.genesis.Params.MinimumInsurance {
		return types.PendingInsuranceWithdrawal{}, errors.New("validator insurance withdrawal would violate active minimum requirement")
	}
	insurance.Balance = remaining
	insurance.PendingWithdrawal = types.PendingInsuranceWithdrawal{
		Amount:		msg.Amount,
		Recipient:	msg.Recipient,
		RequestHeight:	msg.Height,
		CompleteHeight:	msg.Height + k.genesis.Params.WithdrawalLockBlocks,
		Status:		types.WithdrawalStatusPending,
	}
	if _, err := k.saveInsurance(idx, insurance); err != nil {
		return types.PendingInsuranceWithdrawal{}, err
	}
	return insurance.PendingWithdrawal, nil
}

func (k *Keeper) SubmitInsuranceClaim(msg types.MsgSubmitInsuranceClaim) (types.InsuranceClaim, error) {
	if err := k.genesis.Params.Authorize(msg.Authority); err != nil {
		return types.InsuranceClaim{}, err
	}
	if msg.Amount == 0 || msg.Height == 0 {
		return types.InsuranceClaim{}, errors.New("validator insurance claim amount and height must be positive")
	}
	if _, _, found := findClaim(k.genesis.State.Claims, msg.ClaimID); found {
		return types.InsuranceClaim{}, errors.New("validator insurance claim already exists")
	}
	if _, _, found := findInsurance(k.genesis.State.Insurances, msg.ValidatorAddress); !found {
		return types.InsuranceClaim{}, errors.New("validator insurance not found")
	}
	claim := types.InsuranceClaim{
		ClaimID:		msg.ClaimID,
		ValidatorAddress:	msg.ValidatorAddress,
		Claimant:		msg.Claimant,
		Amount:			msg.Amount,
		Status:			types.ClaimStatusPending,
		Reason:			strings.TrimSpace(msg.Reason),
		SubmittedHeight:	msg.Height,
	}
	if err := claim.Validate(k.genesis.Params); err != nil {
		return types.InsuranceClaim{}, err
	}
	next := cloneGenesis(k.genesis)
	next.State.Claims = append(next.State.Claims, claim)
	if err := k.saveGenesis(next); err != nil {
		return types.InsuranceClaim{}, err
	}
	return claim, nil
}

func (k *Keeper) ResolveInsuranceClaim(msg types.MsgResolveInsuranceClaim) (types.InsuranceClaim, error) {
	if err := k.genesis.Params.Authorize(msg.Authority); err != nil {
		return types.InsuranceClaim{}, err
	}
	if msg.Height == 0 {
		return types.InsuranceClaim{}, errors.New("validator insurance claim resolution height must be positive")
	}
	claimIdx, claim, found := findClaim(k.genesis.State.Claims, msg.ClaimID)
	if !found {
		return types.InsuranceClaim{}, errors.New("validator insurance claim not found")
	}
	if claim.Status != types.ClaimStatusPending || claim.Paid {
		return types.InsuranceClaim{}, errors.New("validator insurance claim already resolved")
	}
	next := cloneGenesis(k.genesis)
	if !msg.Approved {
		claim.Status = types.ClaimStatusRejected
		claim.ResolvedHeight = msg.Height
		next.State.Claims[claimIdx] = claim
		if err := k.saveGenesis(next); err != nil {
			return types.InsuranceClaim{}, err
		}
		return claim, nil
	}
	insuranceIdx, insurance, found := findInsurance(next.State.Insurances, claim.ValidatorAddress)
	if !found {
		return types.InsuranceClaim{}, errors.New("validator insurance not found")
	}
	payout := claim.Amount
	if payout > insurance.Balance {
		payout = insurance.Balance
	}
	if payout == 0 {
		return types.InsuranceClaim{}, errors.New("validator insurance balance is empty")
	}
	insurance.Balance -= payout
	claim.PayoutAmount = payout
	claim.Status = types.ClaimStatusPaid
	claim.ResolvedHeight = msg.Height
	claim.Paid = true
	next.State.Insurances[insuranceIdx] = insurance
	next.State.Claims[claimIdx] = claim
	if err := k.saveGenesis(next); err != nil {
		return types.InsuranceClaim{}, err
	}
	return claim, nil
}

func (k *Keeper) ApplyValidatorSlash(validatorAddress string, slashAmount uint64, faultType string) (types.SlashCoverageResult, error) {
	if slashAmount == 0 {
		return types.SlashCoverageResult{}, errors.New("validator insurance slash amount must be positive")
	}
	idx, insurance, found := findInsurance(k.genesis.State.Insurances, validatorAddress)
	if !found {
		return types.SlashCoverageResult{}, errors.New("validator insurance not found")
	}
	coverageBps := k.coverageBps(faultType)
	covered := slashAmount/uint64(types.MaxBasisPoints)*uint64(coverageBps) + slashAmount%uint64(types.MaxBasisPoints)*uint64(coverageBps)/uint64(types.MaxBasisPoints)
	if coverageBps != 0 && covered == 0 {
		covered = 1
	}
	if covered > insurance.Balance {
		covered = insurance.Balance
	}
	insurance.Balance -= covered
	if _, err := k.saveInsurance(idx, insurance); err != nil {
		return types.SlashCoverageResult{}, err
	}
	return types.SlashCoverageResult{
		ValidatorAddress:	validatorAddress,
		SlashAmount:		slashAmount,
		CoveredAmount:		covered,
		RemainingPenalty:	slashAmount - covered,
		CoverageBps:		coverageBps,
	}, nil
}

func (k Keeper) ValidateValidatorActivation(validatorAddress string) error {
	if !k.genesis.Params.Enabled {
		return nil
	}
	_, insurance, found := findInsurance(k.genesis.State.Insurances, validatorAddress)
	if !found || insurance.Balance < k.genesis.Params.MinimumInsurance {
		return errors.New("validator activation requires minimum insurance")
	}
	return nil
}

func (k *Keeper) MarkValidatorStatus(validatorAddress, status string) (types.ValidatorInsurance, error) {
	idx, insurance, found := findInsurance(k.genesis.State.Insurances, validatorAddress)
	if !found {
		return types.ValidatorInsurance{}, errors.New("validator insurance not found")
	}
	insurance.ValidatorStatus = status
	if status == validatorregistrytypes.StatusActive {
		if err := k.ValidateValidatorActivation(validatorAddress); err != nil {
			return types.ValidatorInsurance{}, err
		}
	}
	return k.saveInsurance(idx, insurance)
}

func (k Keeper) ValidatorInsurance(validatorAddress string) (types.ValidatorInsurance, bool) {
	_, insurance, found := findInsurance(k.genesis.State.Insurances, validatorAddress)
	return insurance, found
}

func (k Keeper) InsuranceClaims(validatorAddress string) []types.InsuranceClaim {
	claims := make([]types.InsuranceClaim, 0)
	for _, claim := range k.genesis.State.Claims {
		if validatorAddress == "" || claim.ValidatorAddress == validatorAddress {
			claims = append(claims, claim)
		}
	}
	return types.SortClaims(claims)
}

func (k Keeper) InsuranceParams() types.Params	{ return k.genesis.Params }

type Migrator struct{ keeper *Keeper }

func NewMigrator(k *Keeper) Migrator	{ return Migrator{keeper: k} }
func (m Migrator) Migrate1to2() error	{ return m.keeper.ExportGenesis().Validate() }
func (k Keeper) Migrate1to2State(ctx context.Context) error {
	_, err := k.ExportGenesisState(ctx)
	return err
}

func (k Keeper) coverageBps(faultType string) uint32 {
	faultType = strings.ToLower(strings.TrimSpace(faultType))
	for _, rule := range k.genesis.State.CoverageRules {
		if strings.ToLower(rule.FaultType) == faultType {
			return rule.CoverageBps
		}
	}
	return k.genesis.Params.DefaultSlashCoverageBps
}

func (k *Keeper) saveInsurance(idx int, insurance types.ValidatorInsurance) (types.ValidatorInsurance, error) {
	next := cloneGenesis(k.genesis)
	next.State.Insurances[idx] = insurance
	if err := k.saveGenesis(next); err != nil {
		return types.ValidatorInsurance{}, err
	}
	return insurance, nil
}

func (k *Keeper) saveGenesis(next GenesisState) error {
	next.State = next.State.Normalize(next.Params)
	if err := next.Validate(); err != nil {
		return err
	}
	k.genesis = next
	if k.storeService == nil || k.runtimeCtx == nil {
		return nil
	}
	return prefixgenesis.Save(k.runtimeCtx, k.storeService, genesisKey, next)
}

func cloneGenesis(gs GenesisState) GenesisState {
	gs.State = gs.State.Normalize(gs.Params)
	return gs
}

func findInsurance(values []types.ValidatorInsurance, validatorAddress string) (int, types.ValidatorInsurance, bool) {
	for idx, insurance := range values {
		if insurance.ValidatorAddress == validatorAddress {
			return idx, insurance, true
		}
	}
	return -1, types.ValidatorInsurance{}, false
}

func findClaim(values []types.InsuranceClaim, claimID string) (int, types.InsuranceClaim, bool) {
	for idx, claim := range values {
		if claim.ClaimID == claimID {
			return idx, claim, true
		}
	}
	return -1, types.InsuranceClaim{}, false
}

func validateActor(label, value string) error {
	return addressing.ValidateAuthorityAddress(label, value)
}
