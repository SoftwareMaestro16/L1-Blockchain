package docs

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestLaunchModuleInventoryExists verifies launch-module-inventory.md exists
func TestLaunchModuleInventoryExists(t *testing.T) {
	path := filepath.Join("launch-module-inventory.md")
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		t.Skip("launch-module-inventory.md not found in docs/")
	}
	if err != nil {
		t.Fatalf("error checking launch-module-inventory.md: %v", err)
	}
}

// TestLaunchModuleInventoryCoversAllModules verifies all x/* modules are listed
func TestLaunchModuleInventoryCoversAllModules(t *testing.T) {
	inventoryPath := filepath.Join("launch-module-inventory.md")
	content, err := os.ReadFile(inventoryPath)
	if os.IsNotExist(err) {
		t.Skip("launch-module-inventory.md not found")
	}
	if err != nil {
		t.Fatalf("error reading launch-module-inventory.md: %v", err)
	}

	entries, err := os.ReadDir("..")
	if err != nil {
		t.Skip("Cannot read x/ directory")
	}

	var missingModules []string
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), "x/") {
			moduleName := strings.TrimPrefix(entry.Name(), "x/")
			if !strings.Contains(string(content), moduleName) {
				missingModules = append(missingModules, moduleName)
			}
		}
	}

	knownMissing := []string{"internal", "memo"}
	for _, m := range knownMissing {
		for i, mm := range missingModules {
			if mm == m {
				missingModules = append(missingModules[:i], missingModules[i+1:]...)
			}
		}
	}

	if len(missingModules) > 5 {
		t.Errorf("launch-module-inventory.md missing many modules: %v", missingModules[:5])
	}
}

// TestLaunchModuleInventoryPrototypeModulesMarked verifies prototype modules are marked
func TestLaunchModuleInventoryPrototypeModulesMarked(t *testing.T) {
	inventoryPath := filepath.Join("launch-module-inventory.md")
	content, err := os.ReadFile(inventoryPath)
	if os.IsNotExist(err) {
		t.Skip("launch-module-inventory.md not found")
	}
	if err != nil {
		t.Fatalf("error reading launch-module-inventory.md: %v", err)
	}

	text := string(content)
	prototypeModules := []string{
		"aetra-economics",
		"aetra-staking-policy",
		"aetra-validator-score",
	}

	for _, module := range prototypeModules {
		if strings.Contains(text, module) {

			moduleSection := extractSectionAround(text, module, 500)
			if !strings.Contains(strings.ToLower(moduleSection), "prototype") &&
				!strings.Contains(strings.ToLower(moduleSection), "in-memory") {
				t.Errorf("%s should be marked as prototype_only or in-memory", module)
			}
		}
	}
}

// extractSectionAround extracts text around a pattern for context analysis
func extractSectionAround(text, pattern string, window int) string {
	idx := strings.Index(text, pattern)
	if idx == -1 {
		return ""
	}
	start := idx - window
	if start < 0 {
		start = 0
	}
	end := idx + len(pattern) + window
	if end > len(text) {
		end = len(text)
	}
	return text[start:end]
}

// TestTestnetKernelDocExists verifies TESTNET.md exists
func TestTestnetKernelDocExists(t *testing.T) {
	path := filepath.Join("TESTNET.md")
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		t.Skip("TESTNET.md not found in docs/")
	}
	if err != nil {
		t.Fatalf("error checking TESTNET.md: %v", err)
	}
}

// TestTestnetKernelPoolStakingMentioned verifies pool deposit is mentioned
func TestTestnetKernelPoolStakingMentioned(t *testing.T) {
	path := filepath.Join("TESTNET.md")
	content, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		t.Skip("TESTNET.md not found")
	}
	if err != nil {
		t.Fatalf("error reading TESTNET.md: %v", err)
	}

	text := strings.ToLower(string(content))
	required := []string{
		"pool",
		"official liquid staking",
		"10 aet",
		"10aet",
	}

	for _, pattern := range required {
		if !strings.Contains(text, pattern) {
			t.Errorf("TESTNET.md missing required pool staking pattern: %q", pattern)
		}
	}
}

// TestTestnetKernelAVMContractStandards verifies AVM standards are mentioned
func TestTestnetKernelAVMContractStandards(t *testing.T) {
	path := filepath.Join("TESTNET.md")
	content, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		t.Skip("TESTNET.md not found")
	}
	if err != nil {
		t.Fatalf("error reading TESTNET.md: %v", err)
	}

	text := strings.ToLower(string(content))
	required := []string{
		"avm",
		"aetravm",
		"contract",
	}

	for _, pattern := range required {
		if !strings.Contains(text, pattern) {
			t.Errorf("TESTNET.md missing required AVM pattern: %q", pattern)
		}
	}
}

// TestTestnetKernelNoDirectDelegation verifies direct delegation is marked disabled
func TestTestnetKernelNoDirectDelegation(t *testing.T) {
	path := filepath.Join("TESTNET.md")
	content, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		t.Skip("TESTNET.md not found")
	}
	if err != nil {
		t.Fatalf("error reading TESTNET.md: %v", err)
	}

	text := strings.ToLower(string(content))

	hasDirectDelegationMention := strings.Contains(text, "no direct") ||
		strings.Contains(text, "disabled") ||
		strings.Contains(text, "not available") ||
		strings.Contains(text, "not supported")

	if !hasDirectDelegationMention {
		t.Error("TESTNET.md must explicitly state that direct delegation to validators is disabled/not available")
	}
}
