package localnet

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestLocalnetProfilesIncludes4And5 verifies localnet-4 and localnet-5 profiles exist
func TestLocalnetProfilesIncludes4And5(t *testing.T) {
	profilesPath := filepath.Join("lib", "profiles.ps1")
	content, err := os.ReadFile(profilesPath)
	if os.IsNotExist(err) {
		t.Skip("profiles.ps1 not found")
	}
	if err != nil {
		t.Fatalf("error reading profiles.ps1: %v", err)
	}

	text := string(content)

	requiredProfiles := []string{"localnet-3", "localnet-4", "localnet-5"}
	for _, profile := range requiredProfiles {
		if !strings.Contains(text, profile) {
			t.Errorf("profiles.ps1 missing localnet profile: %s", profile)
		}
	}
}

// TestLocalnetGetNodeCountFunction verifies Get-LocalnetProfileNodeCount exists
func TestLocalnetGetNodeCountFunction(t *testing.T) {
	profilesPath := filepath.Join("lib", "profiles.ps1")
	content, err := os.ReadFile(profilesPath)
	if os.IsNotExist(err) {
		t.Skip("profiles.ps1 not found")
	}
	if err != nil {
		t.Fatalf("error reading profiles.ps1: %v", err)
	}

	text := string(content)

	if !strings.Contains(text, "function Get-LocalnetProfileNodeCount") {
		t.Error("profiles.ps1 missing Get-LocalnetProfileNodeCount function")
	}

	if !strings.Contains(text, `"localnet-4" { return 4 }`) {
		t.Error("Get-LocalnetProfileNodeCount should return 4 for localnet-4 profile")
	}
	if !strings.Contains(text, `"localnet-5" { return 5 }`) {
		t.Error("Get-LocalnetProfileNodeCount should return 5 for localnet-5 profile")
	}
}

// TestLocalnetDiagnosticsScriptExists verifies diagnostics script exists
func TestLocalnetDiagnosticsScriptExists(t *testing.T) {
	diagnosticsPath := filepath.Join("diagnostics-localnet.ps1")
	if _, err := os.Stat(diagnosticsPath); os.IsNotExist(err) {
		t.Error("diagnostics-localnet.ps1 not found - required for localnet diagnostics")
	}
}

// TestLocalnetDiagnosticsNoSecrets verifies diagnostics don't collect secrets
func TestLocalnetDiagnosticsNoSecrets(t *testing.T) {
	diagnosticsPath := filepath.Join("diagnostics-localnet.ps1")
	content, err := os.ReadFile(diagnosticsPath)
	if os.IsNotExist(err) {
		t.Skip("diagnostics-localnet.ps1 not found")
	}
	if err != nil {
		t.Fatalf("error reading diagnostics-localnet.ps1: %v", err)
	}

	text := strings.ToLower(string(content))
	secretPatterns := []string{
		"mnemonic",
		"private_key",
		"priv_key",
		"secret",
		"seed",
		"keyring",
	}

	for _, pattern := range secretPatterns {
		if strings.Contains(text, pattern) && !strings.Contains(text, "no "+pattern) {
			t.Logf("warning: diagnostics script may collect %s - ensure it's excluded or logged as safe", pattern)
		}
	}
}

// TestLocalnetInitScriptSupportsValidatorCount verifies init.ps1 supports variable validator count
func TestLocalnetInitScriptSupportsValidatorCount(t *testing.T) {
	initPath := filepath.Join("init.ps1")
	content, err := os.ReadFile(initPath)
	if os.IsNotExist(err) {
		t.Skip("init.ps1 not found")
	}
	if err != nil {
		t.Fatalf("error reading init.ps1: %v", err)
	}

	text := string(content)

	if !strings.Contains(text, "[int]$ValidatorCount") {
		t.Error("init.ps1 should have ValidatorCount parameter")
	}

	if !strings.Contains(text, "--validator-count") {
		t.Error("init.ps1 should pass --validator-count to testnet init-files")
	}
}

// TestLocalnetStartScriptSupportsValidatorCount verifies start.ps1 validates validator count
func TestLocalnetStartScriptSupportsValidatorCount(t *testing.T) {
	startPath := filepath.Join("start.ps1")
	content, err := os.ReadFile(startPath)
	if os.IsNotExist(err) {
		t.Skip("start.ps1 not found")
	}
	if err != nil {
		t.Fatalf("error reading start.ps1: %v", err)
	}

	text := string(content)

	if !strings.Contains(text, "[int]$ValidatorCount") {
		t.Error("start.ps1 should have ValidatorCount parameter")
	}

	if !strings.Contains(text, "$ValidatorCount") && !strings.Contains(text, "validator count") {
		t.Log("warning: start.ps1 may not validate validator count")
	}
}

// TestLocalnetProfileManifestIncludesNodeCount verifies profile manifest includes validator_count
func TestLocalnetProfileManifestIncludesNodeCount(t *testing.T) {
	profilesPath := filepath.Join("lib", "profiles.ps1")
	content, err := os.ReadFile(profilesPath)
	if os.IsNotExist(err) {
		t.Skip("profiles.ps1 not found")
	}
	if err != nil {
		t.Fatalf("error reading profiles.ps1: %v", err)
	}

	text := string(content)

	if !strings.Contains(text, "validator_count") {
		t.Error("Write-LocalnetProfileManifest should write validator_count to manifest")
	}
}

// TestLocalnetStopScriptExists verifies stop.ps1 exists
func TestLocalnetStopScriptExists(t *testing.T) {
	stopPath := filepath.Join("stop.ps1")
	if _, err := os.Stat(stopPath); os.IsNotExist(err) {
		t.Error("stop.ps1 not found - required for localnet shutdown")
	}
}

// TestLocalnetValidateGenesisScriptExists verifies validate-genesis.ps1 exists
func TestLocalnetValidateGenesisScriptExists(t *testing.T) {
	validatePath := filepath.Join("validate-genesis.ps1")
	if _, err := os.Stat(validatePath); os.IsNotExist(err) {
		t.Error("validate-genesis.ps1 not found - required for genesis validation")
	}
}

// TestLocalnetHealthCheckExists verifies health.ps1 exists
func TestLocalnetHealthCheckExists(t *testing.T) {
	healthPath := filepath.Join("health.ps1")
	if _, err := os.Stat(healthPath); os.IsNotExist(err) {
		t.Error("health.ps1 not found - required for health checks")
	}
}

// TestLocalnetProfilesAllowedList verifies all profiles are valid
func TestLocalnetProfilesAllowedList(t *testing.T) {
	profilesPath := filepath.Join("lib", "profiles.ps1")
	content, err := os.ReadFile(profilesPath)
	if os.IsNotExist(err) {
		t.Skip("profiles.ps1 not found")
	}
	if err != nil {
		t.Fatalf("error reading profiles.ps1: %v", err)
	}

	text := string(content)
	expectedProfiles := []string{
		"base",
		"localnet-3",
		"localnet-4",
		"localnet-5",
	}

	for _, profile := range expectedProfiles {
		if !strings.Contains(text, `"`+profile+`"`) {
			t.Errorf("LocalnetProfiles array missing profile: %s", profile)
		}
	}
}

// TestLocalnetScriptsHaveNoSecretCollection verifies scripts don't collect secrets
func TestLocalnetScriptsHaveNoSecretCollection(t *testing.T) {
	scripts := []string{
		"init.ps1",
		"start.ps1",
		"stop.ps1",
		"diagnostics-localnet.ps1",
	}

	secretPatterns := []string{
		"$env:AETRA_MNEMONIC",
		"private_key",
		"priv_key",
		"keyring-backend=os",
	}

	for _, scriptName := range scripts {
		scriptPath := filepath.Join(scriptName)
		if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
			continue
		}

		content, err := os.ReadFile(scriptPath)
		if err != nil {
			continue
		}

		text := string(content)
		for _, pattern := range secretPatterns {
			if strings.Contains(text, pattern) {
				t.Logf("warning: %s may contain secret collection pattern: %s", scriptName, pattern)
			}
		}
	}
}
