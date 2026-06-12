package cmd

import (
	"strconv"
	"runtime/debug"

	"github.com/cosmos/cosmos-sdk/version"

	appparams "github.com/sovereign-l1/l1/app/params"
	"github.com/sovereign-l1/l1/x/aetravm/avm"
)

const binaryName = "aetrad"

var (
	appVersion	= "dev"
	gitCommit	= ""
	buildDate	= ""
	buildTags	= ""
	dirty		= ""
)

func initVersionInfo() version.ExtraInfo {
	info, _ := debug.ReadBuildInfo()
	settings := buildSettings(info)

	if version.Name == "" {
		version.Name = appparams.ChainName
	}
	if version.AppName == "" || version.AppName == "<appd>" {
		version.AppName = binaryName
	}
	if version.Version == "" {
		version.Version = firstNonEmpty(appVersion, moduleVersion(info), "dev")
	}
	if version.Commit == "" {
		version.Commit = firstNonEmpty(gitCommit, settings["vcs.revision"], "unknown")
	}
	if version.BuildTags == "" {
		version.BuildTags = buildTags
	}

	buildDateValue := firstNonEmpty(buildDate, settings["vcs.time"], "unknown")
	dirtyValue := firstNonEmpty(dirty, settings["vcs.modified"], "unknown")

	return version.ExtraInfo{
		"build_date":		buildDateValue,
		"dirty":		dirtyValue,
		"cosmos_sdk_version":	dependencyVersion(info, "github.com/cosmos/cosmos-sdk"),
		"cometbft_version":	dependencyVersion(info, "github.com/cometbft/cometbft"),
		"avm_version":		strconv.FormatUint(uint64(avm.Version), 10),
	}
}

func buildSettings(info *debug.BuildInfo) map[string]string {
	settings := map[string]string{}
	if info == nil {
		return settings
	}
	for _, setting := range info.Settings {
		settings[setting.Key] = setting.Value
	}
	return settings
}

func moduleVersion(info *debug.BuildInfo) string {
	if info == nil {
		return ""
	}
	if info.Main.Version == "(devel)" {
		return ""
	}
	return info.Main.Version
}

func dependencyVersion(info *debug.BuildInfo, path string) string {
	if info == nil {
		return "unknown"
	}
	for _, dep := range info.Deps {
		if dep.Path != path {
			continue
		}
		if dep.Replace != nil && dep.Replace.Version != "" && dep.Replace.Version != "(devel)" {
			return dep.Replace.Version
		}
		if dep.Version != "" {
			return dep.Version
		}
		return "unknown"
	}
	return "unknown"
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
