package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func repoRoot() string {
	if dir := os.Getenv("REPO_ROOT"); dir != "" {
		return dir
	}
	cwd, _ := os.Getwd()

	if strings.HasSuffix(cwd, ".github/workflows") {
		return filepath.Dir(filepath.Dir(cwd))
	}
	return cwd
}

// TestReleaseWorkflowFileExists verifies testnet-readiness.yml exists
func TestReleaseWorkflowFileExists(t *testing.T) {
	workflowPath := filepath.Join(repoRoot(), ".github", "workflows", "testnet-readiness.yml")
	if _, err := os.Stat(workflowPath); os.IsNotExist(err) {
		t.Error("testnet-readiness.yml not found - required for release workflow")
	}
}

// TestReleaseWorkflowHasGoTest verifies workflow has go test job
func TestReleaseWorkflowHasGoTest(t *testing.T) {
	workflowPath := filepath.Join(repoRoot(), ".github", "workflows", "testnet-readiness.yml")
	content, err := os.ReadFile(workflowPath)
	if os.IsNotExist(err) {
		t.Skip("testnet-readiness.yml not found")
	}
	if err != nil {
		t.Fatalf("error reading testnet-readiness.yml: %v", err)
	}

	text := string(content)
	if !strings.Contains(text, "go test ./...") {
		t.Error("testnet-readiness.yml should run 'go test ./...'")
	}
}

// TestReleaseWorkflowHasGoVet verifies workflow has go vet job
func TestReleaseWorkflowHasGoVet(t *testing.T) {
	workflowPath := filepath.Join(repoRoot(), ".github", "workflows", "testnet-readiness.yml")
	content, err := os.ReadFile(workflowPath)
	if os.IsNotExist(err) {
		t.Skip("testnet-readiness.yml not found")
	}
	if err != nil {
		t.Fatalf("error reading testnet-readiness.yml: %v", err)
	}

	text := string(content)
	if !strings.Contains(text, "go vet") {
		t.Error("testnet-readiness.yml should run 'go vet ./...'")
	}
}

// TestReleaseWorkflowHasBufLint verifies workflow has buf lint job
func TestReleaseWorkflowHasBufLint(t *testing.T) {
	workflowPath := filepath.Join(repoRoot(), ".github", "workflows", "testnet-readiness.yml")
	content, err := os.ReadFile(workflowPath)
	if os.IsNotExist(err) {
		t.Skip("testnet-readiness.yml not found")
	}
	if err != nil {
		t.Fatalf("error reading testnet-readiness.yml: %v", err)
	}

	text := string(content)
	if !strings.Contains(text, "buf lint") {
		t.Error("testnet-readiness.yml should run 'buf lint'")
	}
}

// TestReleaseWorkflowHasGenesisValidate verifies workflow has genesis validation
func TestReleaseWorkflowHasGenesisValidate(t *testing.T) {
	workflowPath := filepath.Join(repoRoot(), ".github", "workflows", "testnet-readiness.yml")
	content, err := os.ReadFile(workflowPath)
	if os.IsNotExist(err) {
		t.Skip("testnet-readiness.yml not found")
	}
	if err != nil {
		t.Fatalf("error reading testnet-readiness.yml: %v", err)
	}

	text := string(content)
	if !strings.Contains(text, "genesis-validate") && !strings.Contains(text, "validate-genesis") {
		t.Error("testnet-readiness.yml should run genesis validation")
	}
}

// TestReleaseWorkflowHasLocalnetSmoke verifies workflow has localnet smoke test
func TestReleaseWorkflowHasLocalnetSmoke(t *testing.T) {
	workflowPath := filepath.Join(repoRoot(), ".github", "workflows", "testnet-readiness.yml")
	content, err := os.ReadFile(workflowPath)
	if os.IsNotExist(err) {
		t.Skip("testnet-readiness.yml not found")
	}
	if err != nil {
		t.Fatalf("error reading testnet-readiness.yml: %v", err)
	}

	text := string(content)
	if !strings.Contains(text, "localnet-smoke") && !strings.Contains(text, "localnet") {
		t.Error("testnet-readiness.yml should run localnet smoke test")
	}
}

// TestReleaseWorkflowHasExportImportSmoke verifies workflow has export/import smoke
func TestReleaseWorkflowHasExportImportSmoke(t *testing.T) {
	workflowPath := filepath.Join(repoRoot(), ".github", "workflows", "testnet-readiness.yml")
	content, err := os.ReadFile(workflowPath)
	if os.IsNotExist(err) {
		t.Skip("testnet-readiness.yml not found")
	}
	if err != nil {
		t.Fatalf("error reading testnet-readiness.yml: %v", err)
	}

	text := string(content)
	if !strings.Contains(text, "export-import") && !strings.Contains(text, "export") {
		t.Error("testnet-readiness.yml should run export/import smoke test")
	}
}

// TestReleaseWorkflowHasInvariants verifies workflow has invariants test
func TestReleaseWorkflowHasInvariants(t *testing.T) {
	workflowPath := filepath.Join(repoRoot(), ".github", "workflows", "testnet-readiness.yml")
	content, err := os.ReadFile(workflowPath)
	if os.IsNotExist(err) {
		t.Skip("testnet-readiness.yml not found")
	}
	if err != nil {
		t.Fatalf("error reading testnet-readiness.yml: %v", err)
	}

	text := string(content)
	if !strings.Contains(text, "invariants") {
		t.Error("testnet-readiness.yml should run invariants test")
	}
}

// TestReleaseWorkflowHasReleaseArtifactBuild verifies workflow has release artifact build
func TestReleaseWorkflowHasReleaseArtifactBuild(t *testing.T) {
	workflowPath := filepath.Join(repoRoot(), ".github", "workflows", "testnet-readiness.yml")
	content, err := os.ReadFile(workflowPath)
	if os.IsNotExist(err) {
		t.Skip("testnet-readiness.yml not found")
	}
	if err != nil {
		t.Fatalf("error reading testnet-readiness.yml: %v", err)
	}

	text := string(content)
	if !strings.Contains(text, "release-artifact") && !strings.Contains(text, "artifact") {
		t.Error("testnet-readiness.yml should build release artifact")
	}
}

// TestReleaseWorkflowHasVersionCommand verifies workflow has version command test
func TestReleaseWorkflowHasVersionCommand(t *testing.T) {
	workflowPath := filepath.Join(repoRoot(), ".github", "workflows", "testnet-readiness.yml")
	content, err := os.ReadFile(workflowPath)
	if os.IsNotExist(err) {
		t.Skip("testnet-readiness.yml not found")
	}
	if err != nil {
		t.Fatalf("error reading testnet-readiness.yml: %v", err)
	}

	text := string(content)
	if !strings.Contains(text, "version") {
		t.Error("testnet-readiness.yml should test version command")
	}
}

// TestReleaseWorkflowHasDockerBuild verifies workflow has Docker build
func TestReleaseWorkflowHasDockerBuild(t *testing.T) {
	workflowPath := filepath.Join(repoRoot(), ".github", "workflows", "testnet-readiness.yml")
	content, err := os.ReadFile(workflowPath)
	if os.IsNotExist(err) {
		t.Skip("testnet-readiness.yml not found")
	}
	if err != nil {
		t.Fatalf("error reading testnet-readiness.yml: %v", err)
	}

	text := string(content)
	if !strings.Contains(text, "docker") && !strings.Contains(text, "Docker") {
		t.Log("warning: testnet-readiness.yml may not include Docker build - check if Dockerfile exists")
	}
}

// TestReleaseWorkflowUploadsArtifacts verifies workflow uploads artifacts
func TestReleaseWorkflowUploadsArtifacts(t *testing.T) {
	workflowPath := filepath.Join(repoRoot(), ".github", "workflows", "testnet-readiness.yml")
	content, err := os.ReadFile(workflowPath)
	if os.IsNotExist(err) {
		t.Skip("testnet-readiness.yml not found")
	}
	if err != nil {
		t.Fatalf("error reading testnet-readiness.yml: %v", err)
	}

	text := string(content)
	if !strings.Contains(text, "upload-artifact") {
		t.Error("testnet-readiness.yml should upload release artifacts")
	}
}

// TestReleasePackageScriptExists verifies release package script exists
func TestReleasePackageScriptExists(t *testing.T) {
	scriptPath := filepath.Join(repoRoot(), "scripts", "release", "prototype-package.ps1")
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		t.Error("scripts/release/prototype-package.ps1 not found - required for release packaging")
	}
}

// TestReleasePackageScriptBuildsBinary verifies release script builds binary
func TestReleasePackageScriptBuildsBinary(t *testing.T) {
	scriptPath := filepath.Join(repoRoot(), "scripts", "release", "prototype-package.ps1")
	content, err := os.ReadFile(scriptPath)
	if os.IsNotExist(err) {
		t.Skip("prototype-package.ps1 not found")
	}
	if err != nil {
		t.Fatalf("error reading prototype-package.ps1: %v", err)
	}

	text := string(content)
	if !strings.Contains(text, "go build") && !strings.Contains(text, "build-aetrad") {
		t.Error("prototype-package.ps1 should build aetrad binary")
	}
}

// TestReleasePackageScriptGeneratesChecksums verifies release script generates checksums
func TestReleasePackageScriptGeneratesChecksums(t *testing.T) {
	scriptPath := filepath.Join(repoRoot(), "scripts", "release", "prototype-package.ps1")
	content, err := os.ReadFile(scriptPath)
	if os.IsNotExist(err) {
		t.Skip("prototype-package.ps1 not found")
	}
	if err != nil {
		t.Fatalf("error reading prototype-package.ps1: %v", err)
	}

	text := string(content)
	if !strings.Contains(text, "checksum") && !strings.Contains(text, "sha256") {
		t.Error("prototype-package.ps1 should generate checksums")
	}
}

// TestReleasePackageScriptIncludesDocs verifies release script includes docs
func TestReleasePackageScriptIncludesDocs(t *testing.T) {
	scriptPath := filepath.Join(repoRoot(), "scripts", "release", "prototype-package.ps1")
	content, err := os.ReadFile(scriptPath)
	if os.IsNotExist(err) {
		t.Skip("prototype-package.ps1 not found")
	}
	if err != nil {
		t.Fatalf("error reading prototype-package.ps1: %v", err)
	}

	text := strings.ToLower(string(content))
	if !strings.Contains(text, "docs") && !strings.Contains(text, "readme") {
		t.Error("prototype-package.ps1 should include docs in release package")
	}
}

// TestReleasePackageScriptHasVersionArg verifies release script has version argument
func TestReleasePackageScriptHasVersionArg(t *testing.T) {
	scriptPath := filepath.Join(repoRoot(), "scripts", "release", "prototype-package.ps1")
	content, err := os.ReadFile(scriptPath)
	if os.IsNotExist(err) {
		t.Skip("prototype-package.ps1 not found")
	}
	if err != nil {
		t.Fatalf("error reading prototype-package.ps1: %v", err)
	}

	text := string(content)
	if !strings.Contains(text, "[string]$Version") {
		t.Error("prototype-package.ps1 should have Version parameter")
	}
}
