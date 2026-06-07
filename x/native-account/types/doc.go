// Package types documents the native account boundary for Aetra.
//
// The native account model is intentionally assigned to x/native-account rather
// than x/identity or SDK x/auth. SDK x/auth remains the compatibility account
// keeper, SDK x/bank remains the native balance layer, and x/identity remains
// the domain and resolver registry. Future implementation files in this module
// should keep state, keys, keeper, messages, queries, genesis, migrations,
// ante/auth policy, events, and tests split by responsibility.
package types
