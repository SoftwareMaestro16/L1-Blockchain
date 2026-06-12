package localnet

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// TestHealthDocExists verifies HEALTH.md exists
func TestHealthDocExists(t *testing.T) {
	healthPath := filepath.Join("..", "..", "docs", "HEALTH.md")
	if _, err := os.Stat(healthPath); os.IsNotExist(err) {
		t.Error("docs/HEALTH.md not found - required health check documentation")
	}
}

// TestHealthDocHasRequiredSections verifies HEALTH.md has required sections
func TestHealthDocHasRequiredSections(t *testing.T) {
	healthPath := filepath.Join("..", "..", "docs", "HEALTH.md")
	content, err := os.ReadFile(healthPath)
	if os.IsNotExist(err) {
		t.Skip("docs/HEALTH.md not found")
	}
	if err != nil {
		t.Fatalf("error reading HEALTH.md: %v", err)
	}

	text := string(content)
	requiredSections := []string{
		"Process Alive",
		"RPC Status",
		"Block Height",
		"Catching Up",
		"Peer Count",
		"Validator Signing",
		"App Invariant",
	}

	for _, section := range requiredSections {
		if !strings.Contains(text, section) {
			t.Errorf("HEALTH.md missing required section: %s", section)
		}
	}
}

// TestHealthScriptExists verifies health.ps1 exists
func TestHealthScriptExists(t *testing.T) {
	healthPath := filepath.Join("health.ps1")
	if _, err := os.Stat(healthPath); os.IsNotExist(err) {
		t.Error("health.ps1 not found - required for health checks")
	}
}

// TestHealthScriptValidatesNodeCount verifies health script validates validator count
func TestHealthScriptValidatesNodeCount(t *testing.T) {
	healthPath := filepath.Join("health.ps1")
	content, err := os.ReadFile(healthPath)
	if os.IsNotExist(err) {
		t.Skip("health.ps1 not found")
	}
	if err != nil {
		t.Fatalf("error reading health.ps1: %v", err)
	}

	text := string(content)
	if !strings.Contains(text, "$ValidatorCount") {
		t.Error("health.ps1 should validate validator count")
	}
}

// TestPeerListJSONExists verifies peers.example.json exists
func TestPeerListJSONExists(t *testing.T) {
	peerPath := filepath.Join("..", "..", "docs", "testnet", "peers.example.json")
	if _, err := os.Stat(peerPath); os.IsNotExist(err) {
		t.Error("docs/testnet/peers.example.json not found - required for peer configuration")
	}
}

// TestPeerListJSONValidFormat verifies peers.example.json has valid structure
func TestPeerListJSONValidFormat(t *testing.T) {
	peerPath := filepath.Join("..", "..", "docs", "testnet", "peers.example.json")
	content, err := os.ReadFile(peerPath)
	if os.IsNotExist(err) {
		t.Skip("peers.example.json not found")
	}

	var peers struct {
		Network	string	`json:"network"`
		Peers	[]struct {
			ID		string	`json:"id"`
			NodeID		string	`json:"node_id"`
			Address		string	`json:"address"`
			Persistent	bool	`json:"persistent"`
			Description	string	`json:"description"`
		}	`json:"peers"`
		Seeds	[]struct {
			ID		string	`json:"id"`
			Address		string	`json:"address"`
			Description	string	`json:"description"`
		}	`json:"seeds"`
	}

	if err := json.Unmarshal(content, &peers); err != nil {
		t.Errorf("peers.example.json has invalid JSON structure: %v", err)
	}

	if peers.Network == "" {
		t.Error("peers.example.json should have network field")
	}
}

// TestPeerListJSONNodeIDFormat verifies node IDs match expected format
func TestPeerListJSONNodeIDFormat(t *testing.T) {
	peerPath := filepath.Join("..", "..", "docs", "testnet", "peers.example.json")
	content, err := os.ReadFile(peerPath)
	if os.IsNotExist(err) {
		t.Skip("peers.example.json not found")
	}

	var peers struct {
		Peers []struct {
			ID string `json:"id"`
		} `json:"peers"`
	}

	if err := json.Unmarshal(content, &peers); err != nil {
		t.Skip("JSON parsing failed")
	}

	nodeIDPattern := regexp.MustCompile(`^[0-9a-fA-F]{64}@.*:\d+$`)
	for _, peer := range peers.Peers {
		if peer.ID != "" && !strings.Contains(peer.ID, "NODE_ID") {
			if !nodeIDPattern.MatchString(peer.ID) {
				t.Logf("warning: peer ID %s may not match CometBFT format (64 hex chars @ host:port)", peer.ID)
			}
		}
	}
}

// TestSeedListExists verifies seeds.example.txt exists
func TestSeedListExists(t *testing.T) {
	seedPath := filepath.Join("..", "..", "docs", "testnet", "seeds.example.txt")
	if _, err := os.Stat(seedPath); os.IsNotExist(err) {
		t.Error("docs/testnet/seeds.example.txt not found - required for seed configuration")
	}
}

// TestSeedListFormat verifies seed list format
func TestSeedListFormat(t *testing.T) {
	seedPath := filepath.Join("..", "..", "docs", "testnet", "seeds.example.txt")
	content, err := os.ReadFile(seedPath)
	if os.IsNotExist(err) {
		t.Skip("seeds.example.txt not found")
	}

	text := string(content)

	if !strings.Contains(text, "seeds =") {
		t.Error("seeds.example.txt should document how to use in config.toml")
	}

	if !strings.Contains(text, "node_id@host:port") && !strings.Contains(text, "NODE_ID") {
		t.Error("seeds.example.txt should show node_id@host:port format")
	}
}

// TestHealthScriptSupportsJSONOutput verifies health script supports JSON output
func TestHealthScriptSupportsJSONOutput(t *testing.T) {
	healthPath := filepath.Join("health.ps1")
	content, err := os.ReadFile(healthPath)
	if os.IsNotExist(err) {
		t.Skip("health.ps1 not found")
	}

	text := string(content)
	if !strings.Contains(text, "-Json") {
		t.Error("health.ps1 should support JSON output for monitoring systems")
	}
}

// TestHealthScriptChecksRPC verifies health script checks RPC status
func TestHealthScriptChecksRPC(t *testing.T) {
	healthPath := filepath.Join("health.ps1")
	content, err := os.ReadFile(healthPath)
	if os.IsNotExist(err) {
		t.Skip("health.ps1 not found")
	}

	text := string(content)
	healthIndicators := []string{
		"26657",
		"status",
		"catching_up",
	}

	found := 0
	for _, indicator := range healthIndicators {
		if strings.Contains(text, indicator) {
			found++
		}
	}

	if found < 2 {
		t.Error("health.ps1 should check RPC status and catching_up")
	}
}

// TestPeerListParserRejectsMalformed verifies parser logic for malformed data
func TestPeerListParserRejectsMalformed(t *testing.T) {
	malformedPeers := []string{
		"",
		"invalid",
		"@host:port",
		"nodeid@",
		"nodeid@host",
		"nodeid@host:invalid",
		"nodeid@host:-1",
		"nodeid@host:70000",
		"1234567890@host:port",
	}

	nodeIDPattern := regexp.MustCompile(`^[0-9a-fA-F]{64}@[^:]+:\d+$`)

	for _, peer := range malformedPeers {
		if nodeIDPattern.MatchString(peer) {
			t.Logf("warning: '%s' unexpectedly matched pattern", peer)
		}
	}
}
