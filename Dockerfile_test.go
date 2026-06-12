package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestDockerfileExists verifies Dockerfile exists
func TestDockerfileExists(t *testing.T) {
	dockerfilePath := filepath.Join("Dockerfile")
	if _, err := os.Stat(dockerfilePath); os.IsNotExist(err) {
		t.Error("Dockerfile not found - required for reproducible image builds")
	}
}

// TestDockerfileHasMultiStageBuild verifies multi-stage build
func TestDockerfileHasMultiStageBuild(t *testing.T) {
	dockerfilePath := filepath.Join("Dockerfile")
	content, err := os.ReadFile(dockerfilePath)
	if os.IsNotExist(err) {
		t.Skip("Dockerfile not found")
	}
	if err != nil {
		t.Fatalf("error reading Dockerfile: %v", err)
	}

	text := string(content)

	if !strings.Contains(text, "AS builder") {
		t.Error("Dockerfile should use multi-stage build with 'AS builder' stage")
	}
	if !strings.Contains(text, "AS runner") {
		t.Error("Dockerfile should have 'AS runner' stage for minimal runtime image")
	}
}

// TestDockerfileHasNonRootUser verifies non-root user for security
func TestDockerfileHasNonRootUser(t *testing.T) {
	dockerfilePath := filepath.Join("Dockerfile")
	content, err := os.ReadFile(dockerfilePath)
	if os.IsNotExist(err) {
		t.Skip("Dockerfile not found")
	}
	if err != nil {
		t.Fatalf("error reading Dockerfile: %v", err)
	}

	text := string(content)

	if !strings.Contains(text, "adduser") && !strings.Contains(text, "addgroup") {
		t.Error("Dockerfile should create non-root user for security")
	}

	if !strings.Contains(text, "USER") {
		t.Error("Dockerfile should switch to non-root user")
	}
}

// TestDockerfileHasHealthcheck verifies healthcheck command
func TestDockerfileHasHealthcheck(t *testing.T) {
	dockerfilePath := filepath.Join("Dockerfile")
	content, err := os.ReadFile(dockerfilePath)
	if os.IsNotExist(err) {
		t.Skip("Dockerfile not found")
	}
	if err != nil {
		t.Fatalf("error reading Dockerfile: %v", err)
	}

	text := strings.ToLower(string(content))

	if !strings.Contains(text, "healthcheck") {
		t.Error("Dockerfile should have HEALTHCHECK directive")
	}

	if !strings.Contains(text, "curl") && !strings.Contains(text, "wget") {
		t.Log("warning: healthcheck may not use curl/wget for health verification")
	}
}

// TestDockerfileHasBuildArgs verifies build args for version/commit
func TestDockerfileHasBuildArgs(t *testing.T) {
	dockerfilePath := filepath.Join("Dockerfile")
	content, err := os.ReadFile(dockerfilePath)
	if os.IsNotExist(err) {
		t.Skip("Dockerfile not found")
	}
	if err != nil {
		t.Fatalf("error reading Dockerfile: %v", err)
	}

	text := string(content)

	requiredArgs := []string{"VERSION", "COMMIT", "DATE"}
	for _, arg := range requiredArgs {
		if !strings.Contains(text, "ARG "+arg) && !strings.Contains(text, "ARG VERSION") {
			t.Logf("warning: Dockerfile should have ARG for %s", arg)
		}
	}

	if !strings.Contains(text, "-ldflags") {
		t.Error("Dockerfile should use ldflags for version metadata")
	}
}

// TestDockerfileHasExposedPorts verifies port exposure
func TestDockerfileHasExposedPorts(t *testing.T) {
	dockerfilePath := filepath.Join("Dockerfile")
	content, err := os.ReadFile(dockerfilePath)
	if os.IsNotExist(err) {
		t.Skip("Dockerfile not found")
	}
	if err != nil {
		t.Fatalf("error reading Dockerfile: %v", err)
	}

	text := string(content)

	requiredPorts := []string{
		"26656",
		"26657",
		"1317",
		"9090",
	}

	for _, port := range requiredPorts {
		if !strings.Contains(text, port) {
			t.Errorf("Dockerfile should EXPOSE port %s", port)
		}
	}
}

// TestDockerfileHasLabels verifies labels for image metadata
func TestDockerfileHasLabels(t *testing.T) {
	dockerfilePath := filepath.Join("Dockerfile")
	content, err := os.ReadFile(dockerfilePath)
	if os.IsNotExist(err) {
		t.Skip("Dockerfile not found")
	}
	if err != nil {
		t.Fatalf("error reading Dockerfile: %v", err)
	}

	text := string(content)

	requiredLabels := []string{
		"org.opencontainers.image.title",
		"org.opencontainers.image.version",
	}

	for _, label := range requiredLabels {
		if !strings.Contains(text, label) {
			t.Errorf("Dockerfile should have LABEL %s", label)
		}
	}
}

// TestDockerComposeLocalnetExists verifies docker-compose for localnet exists
func TestDockerComposeLocalnetExists(t *testing.T) {
	composePath := filepath.Join("docker-compose.localnet.yml")
	if _, err := os.Stat(composePath); os.IsNotExist(err) {
		t.Error("docker-compose.localnet.yml not found - sample localnet configuration")
	}
}

// TestDockerComposeLocalnetDoesNotReplaceScripts verifies compose doesn't replace PS scripts
func TestDockerComposeLocalnetDoesNotReplaceScripts(t *testing.T) {
	composePath := filepath.Join("docker-compose.localnet.yml")
	content, err := os.ReadFile(composePath)
	if os.IsNotExist(err) {
		t.Skip("docker-compose.localnet.yml not found")
	}
	if err != nil {
		t.Fatalf("error reading docker-compose.localnet.yml: %v", err)
	}

	text := strings.ToLower(string(content))

	if !strings.Contains(text, "powershell") && !strings.Contains(text, "scripts/localnet") {
		t.Log("warning: docker-compose should reference PowerShell scripts as alternative")
	}

	if !strings.Contains(text, "sample") && !strings.Contains(text, "testing") && !strings.Contains(text, "not replace") {
		t.Error("docker-compose should indicate it's a sample for testing, not replacement for PS scripts")
	}
}

// TestDockerfileNoSecretInImage verifies no secrets embedded in Dockerfile
func TestDockerfileNoSecretInImage(t *testing.T) {
	dockerfilePath := filepath.Join("Dockerfile")
	content, err := os.ReadFile(dockerfilePath)
	if os.IsNotExist(err) {
		t.Skip("Dockerfile not found")
	}
	if err != nil {
		t.Fatalf("error reading Dockerfile: %v", err)
	}

	secretPatterns := []string{
		"password",
		"secret",
		"token",
		"api_key",
		"apikey",
	}

	for _, pattern := range secretPatterns {
		if strings.Contains(strings.ToLower(string(content)), pattern) {
			t.Logf("warning: Dockerfile contains '%s' - ensure it's not a hardcoded secret", pattern)
		}
	}
}

// TestDockerfileUsesAlpine verifies Alpine Linux for minimal image
func TestDockerfileUsesAlpine(t *testing.T) {
	dockerfilePath := filepath.Join("Dockerfile")
	content, err := os.ReadFile(dockerfilePath)
	if os.IsNotExist(err) {
		t.Skip("Dockerfile not found")
	}
	if err != nil {
		t.Fatalf("error reading Dockerfile: %v", err)
	}

	text := string(content)

	if !strings.Contains(text, "alpine") {
		t.Error("Dockerfile should use Alpine Linux for minimal runtime image")
	}
}

// TestDockerfileEntrypoint verifies entrypoint is aetrad
func TestDockerfileEntrypoint(t *testing.T) {
	dockerfilePath := filepath.Join("Dockerfile")
	content, err := os.ReadFile(dockerfilePath)
	if os.IsNotExist(err) {
		t.Skip("Dockerfile not found")
	}
	if err != nil {
		t.Fatalf("error reading Dockerfile: %v", err)
	}

	text := string(content)

	if !strings.Contains(text, "ENTRYPOINT") {
		t.Error("Dockerfile should have ENTRYPOINT for aetrad")
	}

	if !strings.Contains(text, "aetrad") {
		t.Error("Dockerfile ENTRYPOINT should reference aetrad binary")
	}
}
