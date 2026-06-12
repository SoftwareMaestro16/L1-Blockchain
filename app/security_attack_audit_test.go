package app

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/stretchr/testify/require"
)

func TestConsensusCriticalSourceRejectsNondeterminismAndExternalNetworkCalls(t *testing.T) {
	repoRoot, err := filepath.Abs("..")
	require.NoError(t, err)

	forbidden := []struct {
		token	string
		risk	string
	}{
		{token: "time.Now(", risk: "wall-clock time is non-deterministic in consensus paths"},
		{token: "rand.", risk: "randomness is non-deterministic in consensus paths"},
		{token: "go func", risk: "goroutines can make consensus execution order non-deterministic"},
		{token: "select {", risk: "select can make consensus execution order non-deterministic"},
		{token: "net/http", risk: "external network calls are forbidden inside state transitions"},
		{token: "http.Get", risk: "external network calls are forbidden inside state transitions"},
		{token: "http.Post", risk: "external network calls are forbidden inside state transitions"},
		{token: "grpc.Dial", risk: "external network calls are forbidden inside state transitions"},
		{token: "net.Dial", risk: "external network calls are forbidden inside state transitions"},
		{token: "os.Getenv", risk: "environment-dependent consensus behavior is non-deterministic"},
	}

	var findings []string
	for _, dir := range []string{"app", "x"} {
		root := filepath.Join(repoRoot, dir)
		require.NoError(t, filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
			require.NoError(t, walkErr)
			if entry.IsDir() {
				name := entry.Name()
				if name == "client" || name == "cli" {
					return filepath.SkipDir
				}
				return nil
			}
			if !isConsensusAuditSourceFile(path) {
				return nil
			}
			bz, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			body := string(bz)
			for _, item := range forbidden {
				if strings.Contains(body, item.token) {
					rel, _ := filepath.Rel(repoRoot, path)
					findings = append(findings, rel+": "+item.risk+" via "+item.token)
				}
			}
			return nil
		}))
	}
	require.Empty(t, findings)
}

func TestFinalizeBlockMalformedTxAttackDoesNotPanicOrSucceed(t *testing.T) {
	app := Setup(t, false)
	height := app.LastBlockHeight() + 1
	malformedTx := bytes.Repeat([]byte{0xff, 0x00, 0x13, 0x37}, 256)

	var (
		res	*abci.ResponseFinalizeBlock
		err	error
	)
	require.NotPanics(t, func() {
		res, err = app.FinalizeBlock(&abci.RequestFinalizeBlock{
			Height:	height,
			Hash:	app.LastCommitID().Hash,
			Txs:	[][]byte{malformedTx},
		})
	})
	if err != nil {
		return
	}
	require.NotNil(t, res)
	require.Len(t, res.TxResults, 1)
	require.NotZero(t, res.TxResults[0].Code, "malformed tx must not be accepted")
}

func TestFinalizeBlockHeightRegressionAttackIsRejected(t *testing.T) {
	app := Setup(t, false)
	_, err := app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height:	1,
		Hash:	app.LastCommitID().Hash,
	})
	require.NoError(t, err)
	_, err = app.Commit()
	require.NoError(t, err)

	currentHeight := app.LastBlockHeight()
	require.Positive(t, currentHeight)

	require.NotPanics(t, func() {
		_, err := app.FinalizeBlock(&abci.RequestFinalizeBlock{
			Height:	currentHeight,
			Hash:	app.LastCommitID().Hash,
		})
		require.Error(t, err)
	})
}

func isConsensusAuditSourceFile(path string) bool {
	if !strings.HasSuffix(path, ".go") {
		return false
	}
	name := filepath.Base(path)
	if strings.HasSuffix(name, "_test.go") ||
		strings.HasSuffix(name, ".pb.go") ||
		strings.HasSuffix(name, ".pb.gw.go") {
		return false
	}
	return true
}
