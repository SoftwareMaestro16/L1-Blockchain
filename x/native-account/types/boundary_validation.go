package types

import (
	"errors"
	"fmt"
	"strings"
)

func ValidateBoundaries(boundaries []Boundary) error {
	if len(boundaries) == 0 {
		return errors.New("native account boundaries are required")
	}
	seen := make(map[string]struct{}, len(boundaries))
	required := map[string]bool{
		"app/addressing":			false,
		ModulePath:				false,
		"x/identity":				false,
		"x/reputation":				false,
		"x/storage-rent":			false,
		"x/pos":				false,
		"x/nominator-pool":			false,
		"x/single-nominator-pool":		false,
		"x/validator-*":			false,
		"x/stake-concentration":		false,
		"x/fees":				false,
		"x/burn":				false,
		"x/treasury":				false,
		"x/contracts, x/vm, x/aetravm/*":	false,
	}
	for _, boundary := range boundaries {
		if strings.TrimSpace(boundary.Path) == "" {
			return errors.New("boundary path is required")
		}
		if strings.TrimSpace(boundary.Owner) == "" {
			return fmt.Errorf("boundary %s owner is required", boundary.Path)
		}
		if _, found := seen[boundary.Path]; found {
			return fmt.Errorf("duplicate boundary %s", boundary.Path)
		}
		seen[boundary.Path] = struct{}{}
		if _, found := required[boundary.Path]; found {
			required[boundary.Path] = true
		}
	}
	for path, found := range required {
		if !found {
			return fmt.Errorf("missing boundary %s", path)
		}
	}
	return nil
}

func ValidateAssetRoutes(routes []AssetRoute) error {
	if len(routes) == 0 {
		return errors.New("asset routes are required")
	}
	for _, route := range routes {
		if strings.TrimSpace(route.Behavior) == "" || strings.TrimSpace(route.Route) == "" {
			return errors.New("asset route behavior and route are required")
		}
	}
	return nil
}

func ValidateNoNativeAssetModules(moduleNames []string) error {
	denied := make(map[string]struct{})
	for _, moduleName := range NativeAssetModuleDenylist() {
		denied[moduleName] = struct{}{}
	}
	for _, moduleName := range moduleNames {
		if _, found := denied[moduleName]; found {
			return fmt.Errorf("native asset module %s is not allowed", moduleName)
		}
	}
	return nil
}
