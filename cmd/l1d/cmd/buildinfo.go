package cmd

import (
	"context"
	"strings"

	cmtversion "github.com/cometbft/cometbft/version"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/version"

	"github.com/sovereign-l1/l1/app/params"
)

const (
	binaryName      = "orbitalisd"
	defaultVersion  = "dev"
	unknownMetadata = "unknown"
)

var (
	// BuildDate and Dirty are set by scripts/build/orbitalisd.ps1 through
	// ldflags. Defaults keep local `go test` and ad hoc builds readable.
	BuildDate = unknownMetadata
	Dirty     = unknownMetadata
)

func configureVersionMetadata(rootCmd *cobra.Command) {
	version.Name = params.ChainName
	version.AppName = binaryName
	if strings.TrimSpace(version.Version) == "" {
		version.Version = defaultVersion
	}
	if strings.TrimSpace(version.Commit) == "" {
		version.Commit = unknownMetadata
	}
	if strings.TrimSpace(version.BuildTags) == "" {
		version.BuildTags = unknownMetadata
	}

	extraInfo := version.ExtraInfo{
		"build_date":       defaultMetadata(BuildDate),
		"dirty":            defaultMetadata(Dirty),
		"cometbft_version": cmtversion.TMCoreSemVer,
	}

	ctx := rootCmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	rootCmd.SetContext(context.WithValue(ctx, version.ContextKey{}, extraInfo))
}

func defaultMetadata(value string) string {
	if strings.TrimSpace(value) == "" {
		return unknownMetadata
	}
	return value
}
