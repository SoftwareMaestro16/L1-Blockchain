package app

import aetracorewiring "github.com/sovereign-l1/l1/app/wiring/aetracore"

func aetherCorePreBlockerOrder() []string {
	return aetracorewiring.PreBlockerOrder()
}

func aetherCoreBeginBlockerOrder() []string {
	return aetracorewiring.BeginBlockerOrder()
}

func aetherCoreEndBlockerOrder() []string {
	return aetracorewiring.EndBlockerOrder()
}

func aetherCoreInitGenesisOrder() []string {
	return aetracorewiring.InitGenesisOrder()
}

func aetherCoreExportGenesisOrder() []string {
	return aetracorewiring.ExportGenesisOrder()
}
