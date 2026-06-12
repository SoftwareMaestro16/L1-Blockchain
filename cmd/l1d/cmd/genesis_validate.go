package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"

	appparams "github.com/sovereign-l1/l1/app/params"
)

const (
	localGenesisValidatorMin	= 1
	localGenesisValidatorMax	= 10
)

func wrapAetraGenesisValidation(cmd *cobra.Command) {
	validateCmd := findChildCommand(cmd, "validate-genesis")
	if validateCmd == nil || validateCmd.RunE == nil {
		return
	}
	originalRunE := validateCmd.RunE
	validateCmd.RunE = func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			if err := validateAetraGenesisFile(args[0]); err != nil {
				return err
			}
		}
		return originalRunE(cmd, args)
	}
}

func findChildCommand(cmd *cobra.Command, name string) *cobra.Command {
	for _, child := range cmd.Commands() {
		if child.Name() == name {
			return child
		}
		if nested := findChildCommand(child, name); nested != nil {
			return nested
		}
	}
	return nil
}

func validateAetraGenesisFile(path string) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if containsGenesisSecretLikeMaterial(string(raw)) {
		return fmt.Errorf("genesis contains secret-like material")
	}
	appGenesis, err := genutiltypes.AppGenesisFromFile(path)
	if err != nil {
		return err
	}
	if err := appGenesis.ValidateAndComplete(); err != nil {
		return err
	}
	if err := appparams.ValidateAetraTestnetChainID(appGenesis.ChainID); err != nil {
		return fmt.Errorf("invalid genesis chain-id: %w", err)
	}
	validatorCount, err := countGenesisValidators(appGenesis.AppState)
	if err != nil {
		return err
	}
	return validateGenesisValidatorCount(appGenesis.ChainID, validatorCount)
}

func containsGenesisSecretLikeMaterial(raw string) bool {
	lower := strings.ToLower(raw)
	for _, token := range []string{
		"mnemonic",
		"private_key",
		"private-key",
		"priv_validator",
		"secret",
		"seed",
		"wallet",
	} {
		if strings.Contains(lower, token) {
			return true
		}
	}
	return false
}

func countGenesisValidators(appState json.RawMessage) (int, error) {
	var state map[string]json.RawMessage
	if err := json.Unmarshal(appState, &state); err != nil {
		return 0, fmt.Errorf("invalid genesis app_state: %w", err)
	}
	raw := state[genutiltypes.ModuleName]
	if raw != nil {
		var genutilGenesis genutiltypes.GenesisState
		if err := json.Unmarshal(raw, &genutilGenesis); err != nil {
			return 0, fmt.Errorf("invalid %s genesis state: %w", genutiltypes.ModuleName, err)
		}
		if len(genutilGenesis.GenTxs) > 0 {
			return len(genutilGenesis.GenTxs), nil
		}
	}
	if raw := state["staking"]; raw != nil {
		var stakingGenesis struct {
			Validators []json.RawMessage `json:"validators"`
		}
		if err := json.Unmarshal(raw, &stakingGenesis); err != nil {
			return 0, fmt.Errorf("invalid staking genesis state: %w", err)
		}
		if len(stakingGenesis.Validators) > 0 {
			return len(stakingGenesis.Validators), nil
		}
	}
	if raw == nil {
		return 0, fmt.Errorf("missing %s genesis state", genutiltypes.ModuleName)
	}
	return 0, nil
}

func validateGenesisValidatorCount(chainID string, count int) error {
	if strings.Contains(chainID, "-local-") {
		if count < localGenesisValidatorMin || count > localGenesisValidatorMax {
			return fmt.Errorf("local genesis validator count must be %d-%d, got %d", localGenesisValidatorMin, localGenesisValidatorMax, count)
		}
		return nil
	}
	profile := appparams.DefaultNetworkProfile()
	if _, err := profile.ValidatorSetPhase(count); err != nil {
		return fmt.Errorf("public testnet genesis validator count invalid: %w", err)
	}
	return nil
}
