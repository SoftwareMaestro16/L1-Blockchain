package avmdeterminismgate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDeterminismGateRejectsForbiddenFixtureCalls(t *testing.T) {
	fixture := `package fixture

import (
	cryptorand "crypto/rand"
	"math/rand"
	"net/http"
	"os"
	"time"
)

func ExecuteAVM() {
	_ = time.Now()
	_, _ = rand.Int(cryptorand.Reader, nil)
	_ = os.Getenv("HOME")
	_, _ = os.Open("state.db")
	_, _ = http.Get("https://example.invalid")
	go func() {}()
	var ratio float64 = 1.5
	_ = ratio
}

func ComputeStateRoot(values map[string][]byte) []string {
	out := make([]string, 0, len(values))
	for key := range values {
		out = append(out, key)
	}
	return out
}
`
	path := writeFixture(t, fixture)
	violations, err := (Gate{Root: filepath.Dir(path)}).ScanFile(path)
	require.NoError(t, err)
	output := FormatViolations(violations)
	for _, rule := range []string{
		"forbidden-import",
		"forbidden-time",
		"forbidden-randomness",
		"forbidden-os",
		"forbidden-network",
		"forbidden-goroutine",
		"forbidden-float",
		"unsorted-map-range",
	} {
		require.Contains(t, output, rule)
	}
	require.Contains(t, output, ":")
	require.Contains(t, output, "fixture.go")
}

func TestDeterminismGateAllowsSortedMapStateBuilderFixture(t *testing.T) {
	fixture := `package fixture

import "sort"

func ComputeStateRoot(values map[string][]byte) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	out := make([]string, 0, len(keys))
	for _, key := range keys {
		out = append(out, key)
	}
	return out
}
`
	path := writeFixture(t, fixture)
	violations, err := (Gate{Root: filepath.Dir(path)}).ScanFile(path)
	require.NoError(t, err)
	require.NoError(t, AssertNoViolations(violations), FormatViolations(violations))
}

func TestAVMProductionRuntimePassesDeterminismGate(t *testing.T) {
	root := findRepoRoot(t)
	violations, err := (Gate{Root: root}).ScanProduction()
	require.NoError(t, err)
	require.NoError(t, AssertNoViolations(violations), FormatViolations(violations))
}

func TestProductionScanCoversRequestedRuntimeDirectories(t *testing.T) {
	root := findRepoRoot(t)
	for _, dir := range ProductionDirs {
		info, err := os.Stat(filepath.Join(root, dir))
		require.NoError(t, err, dir)
		require.True(t, info.IsDir(), dir)
	}
}

func writeFixture(t *testing.T, source string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "fixture.go")
	require.NoError(t, os.WriteFile(path, []byte(strings.TrimSpace(source)+"\n"), 0o600))
	return path
}

func findRepoRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	require.NoError(t, err)
	for {
		if _, err := os.Stat(filepath.Join(wd, "go.mod")); err == nil {
			return wd
		}
		parent := filepath.Dir(wd)
		require.NotEqual(t, wd, parent, "go.mod not found")
		wd = parent
	}
}
