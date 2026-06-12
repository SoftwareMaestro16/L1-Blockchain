package types

import (
	"sort"
	"strings"
)

const (
	ModuleName	= "nativeaccount"
	ModulePath	= "x/native-account"
	StoreKey	= ModuleName
)

type Boundary struct {
	Path		string
	Owner		string
	OwnedState	[]string
	Interfaces	[]string
	RejectedWrites	[]string
}

type RejectedCrossModuleWrite struct {
	From	string
	To	string
	State	string
}

type AssetRoute struct {
	Behavior		string
	Route			string
	NativeModuleAllowed	bool
}

func DefaultBoundaries() []Boundary {
	return []Boundary{
		{
			Path:		"app/addressing",
			Owner:		"AE user-facing and 4: raw/internal address codecs",
			OwnedState:	nil,
			Interfaces:	[]string{"format AE account/validator/consensus address", "parse raw 4: system address"},
			RejectedWrites: []string{
				"account state",
				"auth policy",
				"balances",
				"staking state",
			},
		},
		{
			Path:	ModulePath,
			Owner:	"native account state, activation, auth policy, account status, storage-rent account references",
			OwnedState: []string{
				"account/by_user",
				"account/by_raw",
				"account/number",
				"account/reputation",
				"account/storage",
			},
			Interfaces:	[]string{"bank balance lookup", "reputation id lookup", "storage-rent debt settlement"},
			RejectedWrites: []string{
				"private keys",
				"seed phrases",
				"domain registry records",
			},
		},
		{
			Path:		"x/identity",
			Owner:		".aet domain registry, resolver state, and optional domain binding proofs",
			OwnedState:	[]string{"domain records", "resolver records", "domain ownership indexes", "domain binding records"},
			Interfaces:	[]string{"account owner address validation", "contract resolver address validation"},
			RejectedWrites: []string{
				"native account status",
				"native account auth policy",
				"native account sequence",
				"bank balances",
			},
		},
		{
			Path:		"x/reputation",
			Owner:		"deterministic account and contract reputation scores and accumulators",
			OwnedState:	[]string{"reputation record", "stake-time accumulator", "score metadata"},
			Interfaces:	[]string{"native account reputation id reference", "staking exposure inputs"},
			RejectedWrites: []string{
				"account auth policy",
				"account sequence",
				"bank balances",
				"transferable reputation assets",
			},
		},
		{
			Path:		"x/storage-rent",
			Owner:		"rent params, debt accounting, state-size records, and protocol-payer accounting",
			OwnedState:	[]string{"rent params", "rent debt records", "state size records", "system storage reserve"},
			Interfaces:	[]string{"charge account or contract rent", "report frozen/frozen_limited eligibility"},
			RejectedWrites: []string{
				"automatic wallet deletion",
				"protocol-critical freeze due to rent",
				"account address derivation",
			},
		},
		{
			Path:		"x/pos",
			Owner:		"PoS calculation specs and compatibility checks around SDK staking/slashing",
			OwnedState:	nil,
			Interfaces:	[]string{"validator score inputs", "staking policy calculations"},
			RejectedWrites: []string{
				"direct user delegation enablement",
				"native account state",
				"bank balances",
			},
		},
		{
			Path:		"x/nominator-pool",
			Owner:		"legacy/custom nominator pool records and pool accounting boundaries",
			OwnedState:	[]string{"nominator pool state", "pool accounting records"},
			Interfaces:	[]string{"official liquid staking integration boundary"},
			RejectedWrites: []string{
				"native account auth policy",
				"validator selection by normal users",
				"bank balances outside bank keeper movement",
			},
		},
		{
			Path:		"x/single-nominator-pool",
			Owner:		"single-nominator pool state and accounting boundaries",
			OwnedState:	[]string{"single-nominator pool state"},
			Interfaces:	[]string{"official liquid staking integration boundary"},
			RejectedWrites: []string{
				"native account auth policy",
				"validator selection by normal users",
				"bank balances outside bank keeper movement",
			},
		},
		{
			Path:		"x/validator-*",
			Owner:		"validator registry, election, insurance, and validator lifecycle metadata",
			OwnedState:	[]string{"validator registry entries", "validator election records", "validator insurance records"},
			Interfaces:	[]string{"pool allocation registry queries", "validator admission policy"},
			RejectedWrites: []string{
				"native account auth policy",
				"user-facing legacy validator addresses",
				"normal-user direct validator delegation",
			},
		},
		{
			Path:		"x/stake-concentration",
			Owner:		"validator concentration, power-cap snapshots, and anti-centralization metrics",
			OwnedState:	[]string{"concentration params", "network concentration snapshot"},
			Interfaces:	[]string{"validator registry inputs", "allocation engine cap signals"},
			RejectedWrites: []string{
				"native account state",
				"bank balances",
				"validator registry ownership records",
			},
		},
		{
			Path:		"x/fees",
			Owner:		"protocol fee policy, minimum fee checks, and fee distribution accounting",
			OwnedState:	[]string{"fee params", "protocol fee state", "sender tx count index"},
			Interfaces:	[]string{"bank keeper fee movement", "burn and treasury distribution hooks"},
			RejectedWrites: []string{
				"duplicated wallet balances",
				"account auth policy",
				"contract asset state",
			},
		},
		{
			Path:		"x/burn",
			Owner:		"burn accounting and visible burned-supply records",
			OwnedState:	[]string{"burn params", "burned denom totals", "burn reason records"},
			Interfaces:	[]string{"bank keeper burn movement"},
			RejectedWrites: []string{
				"wallet balances",
				"account auth policy",
				"treasury allocation records",
			},
		},
		{
			Path:		"x/treasury",
			Owner:		"treasury buckets, spend proposals, and protocol allocation records",
			OwnedState:	[]string{"treasury params", "allocation buckets", "spend records"},
			Interfaces:	[]string{"bank keeper treasury movement", "governance authorization"},
			RejectedWrites: []string{
				"wallet balances",
				"account auth policy",
				"burn totals",
			},
		},
		{
			Path:		"x/contracts, x/vm, x/aetravm/*",
			Owner:		"contract code/data, VM routing, contract standards, async messages, and contract-owned app assets",
			OwnedState:	[]string{"contract code", "contract data", "contract queues", "contract application state"},
			Interfaces:	[]string{"native account endpoint validation", "bank value movement", "storage-rent contract charging"},
			RejectedWrites: []string{
				"native account private keys",
				"native account seed phrases",
				"native account sequence bypass",
				"SDK staking direct user delegation",
			},
		},
	}
}

func RejectedCrossModuleWrites() []RejectedCrossModuleWrite {
	return []RejectedCrossModuleWrite{
		{From: "app/addressing", To: ModulePath, State: "account state"},
		{From: "x/identity", To: ModulePath, State: "auth policy"},
		{From: "x/identity", To: ModulePath, State: "account sequence"},
		{From: "x/reputation", To: ModulePath, State: "account status"},
		{From: "x/storage-rent", To: ModulePath, State: "automatic wallet deletion"},
		{From: "x/storage-rent", To: "protocol-critical/system state", State: "rent freeze"},
		{From: "x/nominator-pool", To: "x/validator-*", State: "user-selected validator delegation"},
		{From: "x/single-nominator-pool", To: "x/validator-*", State: "user-selected validator delegation"},
		{From: "x/validator-*", To: ModulePath, State: "auth policy"},
		{From: "x/stake-concentration", To: ModulePath, State: "account state"},
		{From: "x/fees", To: ModulePath, State: "duplicated wallet balance"},
		{From: "x/burn", To: ModulePath, State: "wallet balance"},
		{From: "x/treasury", To: ModulePath, State: "wallet balance"},
		{From: "x/contracts, x/vm, x/aetravm/*", To: ModulePath, State: "sequence bypass"},
	}
}

func DefaultAssetRoutes() []AssetRoute {
	return []AssetRoute{
		{Behavior: "native AET balance", Route: "SDK bank/native balance layer", NativeModuleAllowed: true},
	}
}

func NativeAssetModuleDenylist() []string {
	return []string{}
}

func IsRejectedCrossModuleWrite(from, to, state string) bool {
	from = strings.TrimSpace(from)
	to = strings.TrimSpace(to)
	state = strings.TrimSpace(state)
	for _, rejected := range RejectedCrossModuleWrites() {
		if rejected.From == from && rejected.To == to && rejected.State == state {
			return true
		}
	}
	return false
}

func BoundaryManifestLines() []string {
	boundaries := DefaultBoundaries()
	lines := make([]string, 0, len(boundaries))
	for _, boundary := range boundaries {
		owned := append([]string(nil), boundary.OwnedState...)
		rejected := append([]string(nil), boundary.RejectedWrites...)
		sort.Strings(owned)
		sort.Strings(rejected)
		lines = append(lines, boundary.Path+"|"+boundary.Owner+"|"+strings.Join(owned, ",")+"|"+strings.Join(rejected, ","))
	}
	sort.Strings(lines)
	return lines
}
