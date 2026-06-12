package keeper

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sovereign-l1/l1/x/internal/prototype"
	"github.com/sovereign-l1/l1/x/reporter/types"
)

func TestDefaultGenesisValidates(t *testing.T) {
	require.NoError(t, DefaultGenesis().Validate())
}

func TestRegisterAndBondReporter(t *testing.T) {
	k := NewKeeper()
	reporter := registerReporter(t, &k, "11", 1)
	require.Equal(t, types.StatusActive, reporter.Status)
	require.Zero(t, reporter.BondedAmount)

	bonded, err := k.BondReporter(types.MsgBondReporter{
		Authority:		prototype.DefaultAuthority,
		ReporterAddress:	reporter.ReporterAddress,
		Amount:			types.DefaultMinBondAmount,
		Height:			2,
	})
	require.NoError(t, err)
	require.Equal(t, types.DefaultMinBondAmount, bonded.BondedAmount)
}

func TestSubmitValidReport(t *testing.T) {
	k := NewKeeper()
	reporter := registerAndBondReporter(t, &k, "11")

	report, err := k.SubmitReport(validReport(reporter.ReporterAddress, "report-valid", 3, withAccepted(123)))
	require.NoError(t, err)
	require.Equal(t, types.ReportStatusAccepted, report.Status)
	require.Equal(t, uint64(123), report.RewardAmount)

	stored, found := k.Reporter(reporter.ReporterAddress)
	require.True(t, found)
	require.Equal(t, uint64(1), stored.AcceptedReports)
	require.Len(t, stored.RewardHistory, 1)
	require.False(t, stored.RewardHistory[0].Claimed)
}

func TestMaliciousReportSlashesReporter(t *testing.T) {
	k := NewKeeper()
	reporter := registerAndBondReporter(t, &k, "11")

	report, err := k.SubmitReport(validReport(reporter.ReporterAddress, "report-malicious", 3, withMalicious()))
	require.NoError(t, err)
	require.Equal(t, types.ReportStatusMalicious, report.Status)
	require.Positive(t, report.SlashAmount)

	stored, found := k.Reporter(reporter.ReporterAddress)
	require.True(t, found)
	require.Equal(t, report.SlashAmount, stored.SlashedReporterBond)
	require.Equal(t, types.DefaultMinBondAmount-report.SlashAmount, stored.BondedAmount)
	require.Equal(t, uint64(1), stored.RejectedReports)
}

func TestRewardPaidOnce(t *testing.T) {
	k := NewKeeper()
	reporter := registerAndBondReporter(t, &k, "11")
	_, err := k.SubmitReport(validReport(reporter.ReporterAddress, "report-reward", 3, withAccepted(321)))
	require.NoError(t, err)

	reward, err := k.ClaimReporterReward(types.MsgClaimReporterReward{
		Authority:		prototype.DefaultAuthority,
		ReporterAddress:	reporter.ReporterAddress,
		ReportID:		"report-reward",
		Height:			4,
	})
	require.NoError(t, err)
	require.True(t, reward.Claimed)
	require.Equal(t, uint64(4), reward.ClaimedAt)

	_, err = k.ClaimReporterReward(types.MsgClaimReporterReward{
		Authority:		prototype.DefaultAuthority,
		ReporterAddress:	reporter.ReporterAddress,
		ReportID:		"report-reward",
		Height:			5,
	})
	require.ErrorContains(t, err, "already claimed")
}

func TestUnbondingDelayEnforced(t *testing.T) {
	k := NewKeeper()
	gs := k.ExportGenesis()
	gs.Params.ChallengePeriodBlocks = 10
	require.NoError(t, k.InitGenesis(gs))
	reporter := registerAndBondReporter(t, &k, "11")

	unbonding, err := k.UnbondReporter(types.MsgUnbondReporter{
		Authority:		prototype.DefaultAuthority,
		ReporterAddress:	reporter.ReporterAddress,
		Height:			7,
	})
	require.NoError(t, err)
	require.Equal(t, types.StatusUnbonding, unbonding.Status)
	require.Equal(t, uint64(17), unbonding.UnbondingCompleteHeight)

	_, err = k.UnbondReporter(types.MsgUnbondReporter{
		Authority:		prototype.DefaultAuthority,
		ReporterAddress:	reporter.ReporterAddress,
		Height:			16,
	})
	require.ErrorContains(t, err, "challenge period")

	completed, err := k.UnbondReporter(types.MsgUnbondReporter{
		Authority:		prototype.DefaultAuthority,
		ReporterAddress:	reporter.ReporterAddress,
		Height:			17,
	})
	require.NoError(t, err)
	require.Equal(t, types.StatusActive, completed.Status)
	require.Zero(t, completed.BondedAmount)
}

func TestExportImportPreservesUnclaimedRewards(t *testing.T) {
	source := NewKeeper()
	reporter := registerAndBondReporter(t, &source, "11")
	_, err := source.SubmitReport(validReport(reporter.ReporterAddress, "report-unclaimed", 3, withAccepted(777)))
	require.NoError(t, err)

	exported := source.ExportGenesis()
	require.NoError(t, exported.Validate())
	target := NewKeeper()
	require.NoError(t, target.InitGenesis(exported))
	require.Equal(t, exported, target.ExportGenesis())
	require.NoError(t, NewMigrator(&target).Migrate1to2())

	rewards := target.ReporterRewards(reporter.ReporterAddress)
	require.Len(t, rewards, 1)
	require.Equal(t, uint64(777), rewards[0].Amount)
	require.False(t, rewards[0].Claimed)
}

func TestReporterMustBeBondedToSubmitSlashableReports(t *testing.T) {
	k := NewKeeper()
	reporter := registerReporter(t, &k, "11", 1)

	_, err := k.SubmitReport(validReport(reporter.ReporterAddress, "report-unbonded", 2, withAccepted(100)))
	require.ErrorContains(t, err, "must be bonded")
}

func registerAndBondReporter(t *testing.T, k *Keeper, fill string) types.ReporterRecord {
	t.Helper()
	reporter := registerReporter(t, k, fill, 1)
	bonded, err := k.BondReporter(types.MsgBondReporter{
		Authority:		prototype.DefaultAuthority,
		ReporterAddress:	reporter.ReporterAddress,
		Amount:			types.DefaultMinBondAmount,
		Height:			2,
	})
	require.NoError(t, err)
	return bonded
}

func registerReporter(t *testing.T, k *Keeper, fill string, height uint64) types.ReporterRecord {
	t.Helper()
	reporter, err := k.RegisterReporter(types.MsgRegisterReporter{
		Authority:		prototype.DefaultAuthority,
		ReporterAddress:	rawReporterAddress(fill),
		Height:			height,
	})
	require.NoError(t, err)
	return reporter
}

func validReport(reporter string, reportID string, height uint64, opts ...func(*types.MsgSubmitReport)) types.MsgSubmitReport {
	msg := types.MsgSubmitReport{
		Authority:		prototype.DefaultAuthority,
		ReporterAddress:	reporter,
		ReportID:		reportID,
		ReportType:		types.ReportTypeFault,
		Subject:		"validator-fault",
		PayloadHash:		payloadHash(reportID),
		PayloadSizeBytes:	128,
		Height:			height,
	}
	for _, opt := range opts {
		opt(&msg)
	}
	return msg
}

func withAccepted(reward uint64) func(*types.MsgSubmitReport) {
	return func(msg *types.MsgSubmitReport) {
		msg.Accepted = true
		msg.RewardAmount = reward
	}
}

func withMalicious() func(*types.MsgSubmitReport) {
	return func(msg *types.MsgSubmitReport) {
		msg.Malicious = true
	}
}

func payloadHash(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func rawReporterAddress(hexByte string) string {
	return "4:000000000000000000000000" + fmt.Sprintf("%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s", hexByte, hexByte, hexByte, hexByte, hexByte, hexByte, hexByte, hexByte, hexByte, hexByte, hexByte, hexByte, hexByte, hexByte, hexByte, hexByte, hexByte, hexByte, hexByte, hexByte)
}
