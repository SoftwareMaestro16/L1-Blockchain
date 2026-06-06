# Financial Zone Migration Marker

Status: extract-app-specific-financial-logic-to-avm-contracts.

`x/zones` remains native system zone infrastructure. The app-specific financial
logic currently represented by contract-assets roots, DEX pool/swap state, and
payment channel state in `financial_zone.go` is marked for extraction to AVM
contracts after replacement contracts and compatibility tests exist.
