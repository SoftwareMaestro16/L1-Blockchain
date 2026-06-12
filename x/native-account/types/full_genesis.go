package types

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/sovereign-l1/l1/x/internal/prototype"
	nominatorpooltypes "github.com/sovereign-l1/l1/x/nominator-pool/types"
	proofregistrytypes "github.com/sovereign-l1/l1/x/proofregistry/types"
	reputationtypes "github.com/sovereign-l1/l1/x/reputation/types"
	storagerenttypes "github.com/sovereign-l1/l1/x/storage-rent/types"
	validatorregistrytypes "github.com/sovereign-l1/l1/x/validator-registry/types"
)

type FullGenesisState struct {
	Version	uint64

	Accounts	[]Account

	ValidatorRegistryVersion	uint64
	ValidatorRegistryParams		validatorregistrytypes.Params
	ValidatorRegistryState		validatorregistrytypes.State

	LiquidStakingVersion	uint64
	LiquidStakingParams	nominatorpooltypes.Params
	LiquidStakingState	nominatorpooltypes.State

	ReputationVersion	uint64
	ReputationState		reputationtypes.ConsolidatedReputationState

	StorageRentVersion	uint64
	StorageRentParams	storagerenttypes.StorageRentParams
	StorageRentState	storagerenttypes.StorageRentState

	ProofMetadataVersion	uint64
	ProofMetadataState	proofregistrytypes.ProofRegistryState
}

type FullGenesisReader interface {
	FullGenesisState() (FullGenesisState, error)
}

type FullGenesisWriter interface {
	SetFullGenesisState(FullGenesisState) error
}

func DefaultFullGenesis() FullGenesisState {
	reputationState := reputationtypes.NewConsolidatedReputationState(reputationtypes.DefaultReputationParams())
	proofState, _ := proofregistrytypes.NewProofRegistryState(proofregistrytypes.DefaultHistoryWindow)
	return FullGenesisState{
		Version:			prototype.CurrentGenesisVersion,
		Accounts:			nil,
		ValidatorRegistryVersion:	prototype.CurrentGenesisVersion,
		ValidatorRegistryParams:	validatorregistrytypes.DefaultParams(),
		ValidatorRegistryState:		validatorregistrytypes.State{},
		LiquidStakingVersion:		prototype.CurrentGenesisVersion,
		LiquidStakingParams:		nominatorpooltypes.DefaultParams(),
		LiquidStakingState:		nominatorpooltypes.State{},
		ReputationVersion:		prototype.CurrentGenesisVersion,
		ReputationState:		reputationState,
		StorageRentVersion:		prototype.CurrentGenesisVersion,
		StorageRentParams:		storagerenttypes.DefaultStorageRentParams(),
		StorageRentState:		storagerenttypes.EmptyStorageRentState(),
		ProofMetadataVersion:		prototype.CurrentGenesisVersion,
		ProofMetadataState:		proofState,
	}
}

func ExportFullGenesis(reader FullGenesisReader) (FullGenesisState, error) {
	if reader == nil {
		return FullGenesisState{}, errors.New("native account full genesis reader is required")
	}
	gs, err := reader.FullGenesisState()
	if err != nil {
		return FullGenesisState{}, err
	}
	gs = NormalizeFullGenesis(gs)
	if err := ValidateFullGenesis(gs); err != nil {
		return FullGenesisState{}, err
	}
	return cloneFullGenesis(gs), nil
}

func ImportFullGenesis(writer FullGenesisWriter, gs FullGenesisState) error {
	if writer == nil {
		return errors.New("native account full genesis writer is required")
	}
	gs = NormalizeFullGenesis(gs)
	if err := ValidateFullGenesis(gs); err != nil {
		return err
	}
	return writer.SetFullGenesisState(cloneFullGenesis(gs))
}

func ValidateFullGenesis(gs FullGenesisState) error {
	if err := validateFullGenesisVersions(gs); err != nil {
		return err
	}
	if err := (GenesisState{Version: prototype.CurrentGenesisVersion, Accounts: gs.Accounts}).Validate(); err != nil {
		return err
	}
	if err := gs.ValidatorRegistryState.Validate(gs.ValidatorRegistryParams); err != nil {
		return err
	}
	if err := gs.LiquidStakingState.Validate(gs.LiquidStakingParams); err != nil {
		return err
	}
	if err := gs.ReputationState.Validate(); err != nil {
		return err
	}
	if err := gs.StorageRentState.Validate(gs.StorageRentParams); err != nil {
		return err
	}
	if err := gs.ProofMetadataState.Validate(); err != nil {
		return err
	}
	return ValidateNoSecretsInFullGenesis(gs)
}

func NormalizeFullGenesis(gs FullGenesisState) FullGenesisState {
	gs.Accounts = SortAccounts(gs.Accounts)
	gs.ValidatorRegistryState = gs.ValidatorRegistryState.Normalize(gs.ValidatorRegistryParams)
	gs.LiquidStakingState = gs.LiquidStakingState.Normalize(gs.LiquidStakingParams)
	gs.ReputationState = reputationtypes.NormalizeConsolidatedState(gs.ReputationState)
	gs.StorageRentState = gs.StorageRentState.Export()
	gs.ProofMetadataState = normalizeProofMetadataState(gs.ProofMetadataState)
	return gs
}

func ValidateNoSecretsInFullGenesis(gs FullGenesisState) error {
	bz, err := json.Marshal(gs)
	if err != nil {
		return err
	}
	text := strings.ToLower(string(bz))
	if strings.Contains(text, "private key") ||
		strings.Contains(text, "private_key") ||
		strings.Contains(text, "seed phrase") ||
		strings.Contains(text, "seed_phrase") ||
		strings.Contains(text, "mnemonic") {
		return errors.New("native account full genesis must not export private keys or seed phrases")
	}
	return nil
}

func validateFullGenesisVersions(gs FullGenesisState) error {
	versions := map[string]uint64{
		"native account full genesis":	gs.Version,
		"validator registry":		gs.ValidatorRegistryVersion,
		"liquid staking":		gs.LiquidStakingVersion,
		"reputation":			gs.ReputationVersion,
		"storage rent":			gs.StorageRentVersion,
		"proof metadata":		gs.ProofMetadataVersion,
	}
	for name, version := range versions {
		if version != prototype.CurrentGenesisVersion {
			return fmt.Errorf("%s unsupported genesis version %d", name, version)
		}
	}
	return nil
}

func cloneFullGenesis(gs FullGenesisState) FullGenesisState {
	return NormalizeFullGenesis(gs)
}

func normalizeProofMetadataState(state proofregistrytypes.ProofRegistryState) proofregistrytypes.ProofRegistryState {
	state = state.Clone()
	sort.SliceStable(state.Snapshots, func(i, j int) bool {
		return state.Snapshots[i].Height < state.Snapshots[j].Height
	})
	for idx := range state.Snapshots {
		sort.SliceStable(state.Snapshots[idx].Roots, func(i, j int) bool {
			left := state.Snapshots[idx].Roots[i]
			right := state.Snapshots[idx].Roots[j]
			if left.RootType != right.RootType {
				return left.RootType < right.RootType
			}
			if left.ZoneID != right.ZoneID {
				return left.ZoneID < right.ZoneID
			}
			return left.Source < right.Source
		})
		sort.SliceStable(state.Snapshots[idx].Metadata, func(i, j int) bool {
			left := state.Snapshots[idx].Metadata[i]
			right := state.Snapshots[idx].Metadata[j]
			if left.RootType != right.RootType {
				return left.RootType < right.RootType
			}
			return left.Source < right.Source
		})
	}
	sort.SliceStable(state.Entries, func(i, j int) bool {
		return state.Entries[i].EntryHash < state.Entries[j].EntryHash
	})
	sort.SliceStable(state.TestVectors, func(i, j int) bool {
		return state.TestVectors[i].VectorID < state.TestVectors[j].VectorID
	})
	return state
}
