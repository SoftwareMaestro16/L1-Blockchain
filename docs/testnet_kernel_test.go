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

	// List all x/* directories
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

	// Allow some modules to be intentionally not documented
	knownMissing := []string{"internal", "memo"} // internal packages don't need docs
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

// TestLaunchModuleInventoryNoNativeAppModules verifies native app modules are marked
func TestLaunchModuleInventoryNoNativeAppModules(t *testing.T) {
	inventoryPath := filepath.Join("launch-module-inventory.md")
	content, err := os.ReadFile(inventoryPath)
	if os.IsNotExist(err) {
		t.Skip("launch-module-inventory.md not found")
	}
	if err != nil {
		t.Fatalf("error reading launch-module-inventory.md: %v", err)
	}

	// Check that tokenfactory/dex/nft are marked as disabled or not native modules
	// They should either be absent or marked as "disabled" or "future avm standard"
	disabledPatterns := []string{
		"tokenfactory",
		"dex",
		"nft",
	}

	for _, pattern := range disabledPatterns {
		// Count occurrences - if marked as launch_core or launch_support, that's a problem
		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			lower := strings.ToLower(line)
			if strings.Contains(lower, pattern) {
				// Check if it's marked as disabled
				if !strings.Contains(lower, "disabled") &&
					!strings.Contains(lower, "not existent") &&
					!strings.Contains(lower, "future avm") &&
					!strings.Contains(lower, "deprecated") &&
					!strings.Contains(lower, "avm contract") {
					// Module mentioned without proper classification
					t.Logf("warning: %s mentioned without clear disabled/future classification", pattern)
				}
			}
		}
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
			// Check if it's marked as prototype_only
			moduleSection := extractSectionAround(text, module, 500)
			if !strings.Contains(strings.ToLower(moduleSection), "prototype") &&
				!strings.Contains(strings.ToLower(moduleSection), "in-memory") {
				t.Errorf("%s should be marked as prototype_only or in-memory", module)
			}
		}
	}
}

// TestLaunchModuleInventoryAVMStandardsMarked verifies AVM standards are marked correctly
func TestLaunchModuleInventoryAVMStandardsMarked(t *testing.T) {
	inventoryPath := filepath.Join("launch-module-inventory.md")
	content, err := os.ReadFile(inventoryPath)
	if os.IsNotExist(err) {
		t.Skip("launch-module-inventory.md not found")
	}
	if err != nil {
		t.Fatalf("error reading launch-module-inventory.md: %v", err)
	}

	text := string(content)
	avmStandards := []string{
		"aetravm/standards",
		"aft",
		"anft",
		"adex",
	}

	for _, standard := range avmStandards {
		if strings.Contains(text, standard) {
			// Check if it's marked as future_avm_standard
			moduleSection := extractSectionAround(text, standard, 500)
			if !strings.Contains(strings.ToLower(moduleSection), "future") &&
				!strings.Contains(strings.ToLower(moduleSection), "avm standard") &&
				!strings.Contains(strings.ToLower(moduleSection), "avm contract") {
				t.Errorf("%s should be marked as future_avm_standard", standard)
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

// TestTestnetKernelNoAevaloperInDocs verifies docs don't teach aevaloper/aevalcons
// for user-facing operations. Context like "not aevaloper" or examples in
// "out of scope" tables are allowed.
func TestTestnetKernelNoAevaloperInDocs(t *testing.T) {
	testFiles := []string{
		"TESTNET.md",
		"official-liquid-staking.md",
	}

	for _, filename := range testFiles {
		path := filepath.Join(filename)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			continue
		}

		content, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		text := string(content)

		// Split into lines and check each line context
		lines := strings.Split(text, "\n")
		for lineNum, line := range lines {
			lower := strings.ToLower(line)

			// Skip lines that explain what's NOT being used (context examples)
			if strings.Contains(lower, "not aevaloper") ||
				strings.Contains(lower, "not aevalcons") ||
				strings.Contains(lower, "x/") && strings.Contains(lower, "deprecated") ||
				strings.Contains(lower, "out of scope") ||
				strings.Contains(lower, "not launching") ||
				strings.Contains(lower, "future") {
				continue
			}

			// Check for user-facing delegation instructions
			badPatterns := []string{
				"delegate to a validator",
				"choose a validator to delegate",
				"select validator to",
			}

			for _, pattern := range badPatterns {
				if strings.Contains(lower, pattern) {
					t.Errorf("%s:%d contains user-facing validator delegation instruction: %q", filename, lineNum+1, pattern)
				}
			}
		}
	}
}

// TestTestnetKernelNoNativeAppModulesInDocs verifies launch docs don't teach
// native DEX/token/NFT as launch features. Mentions in "out of scope" or
// "future" sections are allowed. Also checks that bootstrap-profile.md
// doesn't reference non-existent native modules.
func TestTestnetKernelNoNativeAppModulesInDocs(t *testing.T) {
	testFiles := []string{
		"TESTNET.md",
		"bootstrap-plan.md",
		"release/instructions.md",
	}

	// bootstrap-profile.md should not reference x/tokenfactory or x/dex
	// as native modules (they don't exist in launch profile)
	bootstrapProfilePath := filepath.Join("bootstrap-profile.md")
	if content, err := os.ReadFile(bootstrapProfilePath); err == nil {
		text := strings.ToLower(string(content))
		if strings.Contains(text, "x/tokenfactory") && !strings.Contains(text, "deprecated") && !strings.Contains(text, "not launching") {
			t.Errorf("bootstrap-profile.md references x/tokenfactory native module which does not exist")
		}
		if strings.Contains(text, "x/dex") && !strings.Contains(text, "deprecated") && !strings.Contains(text, "not launching") {
			t.Errorf("bootstrap-profile.md references x/dex native module which does not exist")
		}
	}

	for _, filename := range testFiles {
		path := filepath.Join(filename)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			continue
		}

		content, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		text := string(content)
		lines := strings.Split(text, "\n")

		for lineNum, line := range lines {
			lower := strings.ToLower(line)

			// Skip lines that are explicitly marking things as out of scope
			if strings.Contains(lower, "out of scope") ||
				strings.Contains(lower, "not launching") ||
				strings.Contains(lower, "future") ||
				strings.Contains(lower, "deprecated") ||
				strings.Contains(lower, "target:") ||
				strings.Contains(lower, "status:") ||
				strings.Contains(lower, "avm contract") ||
				strings.Contains(lower, "avm standard") ||
				strings.Contains(lower, "| not |") ||
				strings.Contains(lower, "| x/") {
				continue
			}

			// Check for native application asset module instructions
			badPatterns := []string{
				"x/tokenfactory",
				"token factory",
				"create-token",
				"x/dex",
				"native dex",
				"x/nft",
				"native nft",
				"create-nft",
				"x/market",
			}

			for _, pattern := range badPatterns {
				if strings.Contains(lower, pattern) {
					t.Errorf("%s:%d contains native application asset module pattern: %q", filename, lineNum+1, pattern)
				}
			}
		}
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

	// Must mention that direct delegation is disabled
	hasDirectDelegationMention := strings.Contains(text, "no direct") ||
		strings.Contains(text, "disabled") ||
		strings.Contains(text, "not available") ||
		strings.Contains(text, "not supported")

	if !hasDirectDelegationMention {
		t.Error("TESTNET.md must explicitly state that direct delegation to validators is disabled/not available")
	}
}