package types

import (
	"errors"
	"fmt"
	"sort"
)

const (
	EconomicsAuthoritativePolicyModule	= "x/aetra-economics"
	EconomicsFeesModule			= "x/fees"
	EconomicsFeeCollectorModule		= "x/fee-collector"
	EconomicsBurnModule			= "x/burn"
	EconomicsTreasuryModule			= "x/treasury"
	EconomicsEmissionsModule		= "x/emissions"
	EconomicsMintAuthorityModule		= "x/mint-authority"
)

type EconomicsModuleRole struct {
	Module			string	`json:"module"`
	AuthoritativePolicy	bool	`json:"authoritative_policy"`
	ExecutionAccounting	bool	`json:"execution_accounting"`
}

func DefaultEconomicsAuthorityPath() []EconomicsModuleRole {
	return []EconomicsModuleRole{
		{Module: EconomicsAuthoritativePolicyModule, AuthoritativePolicy: true, ExecutionAccounting: false},
		{Module: EconomicsFeesModule, AuthoritativePolicy: false, ExecutionAccounting: true},
		{Module: EconomicsFeeCollectorModule, AuthoritativePolicy: false, ExecutionAccounting: true},
		{Module: EconomicsBurnModule, AuthoritativePolicy: false, ExecutionAccounting: true},
		{Module: EconomicsTreasuryModule, AuthoritativePolicy: false, ExecutionAccounting: true},
		{Module: EconomicsEmissionsModule, AuthoritativePolicy: false, ExecutionAccounting: true},
		{Module: EconomicsMintAuthorityModule, AuthoritativePolicy: false, ExecutionAccounting: true},
	}
}

func ValidateEconomicsAuthorityPath(path []EconomicsModuleRole) error {
	if len(path) == 0 {
		return errors.New("economics authority path is required")
	}
	seen := map[string]struct{}{}
	authoritative := ""
	executionModules := map[string]bool{
		EconomicsFeesModule:		false,
		EconomicsFeeCollectorModule:	false,
		EconomicsBurnModule:		false,
		EconomicsTreasuryModule:	false,
		EconomicsEmissionsModule:	false,
		EconomicsMintAuthorityModule:	false,
	}
	for _, role := range path {
		if role.Module == "" {
			return errors.New("economics module role requires module")
		}
		if _, found := seen[role.Module]; found {
			return fmt.Errorf("duplicate economics module role %s", role.Module)
		}
		seen[role.Module] = struct{}{}
		if role.AuthoritativePolicy {
			if authoritative != "" {
				return errors.New("economics authority path must have exactly one authoritative policy module")
			}
			authoritative = role.Module
		}
		if _, expectedExecution := executionModules[role.Module]; expectedExecution {
			if role.AuthoritativePolicy {
				return fmt.Errorf("%s must not be an authoritative economics policy module", role.Module)
			}
			executionModules[role.Module] = role.ExecutionAccounting
		}
	}
	if authoritative != EconomicsAuthoritativePolicyModule {
		return fmt.Errorf("economics authoritative policy module must be %s", EconomicsAuthoritativePolicyModule)
	}
	missing := make([]string, 0)
	for module, present := range executionModules {
		if !present {
			missing = append(missing, module)
		}
	}
	sort.Strings(missing)
	if len(missing) > 0 {
		return fmt.Errorf("economics execution modules missing or not marked accounting: %v", missing)
	}
	return nil
}
