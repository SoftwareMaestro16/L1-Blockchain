package app

import aethercorewiring "github.com/sovereign-l1/l1/app/wiring/aethercore"

func aetherCorePreBlockerOrder() []string {
	return aethercorewiring.PreBlockerOrder()
}

func aetherCoreBeginBlockerOrder() []string {
	return aethercorewiring.BeginBlockerOrder()
}

func aetherCoreEndBlockerOrder() []string {
	return aethercorewiring.EndBlockerOrder()
}

func aetherCoreInitGenesisOrder() []string {
	return aethercorewiring.InitGenesisOrder()
}

func aetherCoreExportGenesisOrder() []string {
	return aethercorewiring.ExportGenesisOrder()
}
