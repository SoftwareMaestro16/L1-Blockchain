// Package types documents the native account boundary for Aetra.
//
// The native account model is intentionally assigned to x/native-account rather
// than x/identity or SDK x/auth. SDK x/auth remains the compatibility account
// keeper, SDK x/bank remains the native balance layer, and x/identity remains
// the domain and resolver registry. Future implementation files in this module
// should keep state, keys, keeper, messages, queries, genesis, migrations,
// ante/auth policy, events, and tests split by responsibility.
//
// Account state keys are prefix based and deterministic:
// account/by_user/{AE...}, account/by_raw/{4:...}, account/number/{zero-padded
// account_number}, account/reputation/{reputation_id}, and
// account/storage/{AE...}. User-facing query inputs accept AE addresses only;
// raw 4: addresses are accepted only on raw/internal query boundaries.
//
// Account activation initializes sequence at 0. This is a compatibility
// contract: future upgrades must preserve sequence semantics or ship an
// explicit migration with compatibility tests.
package types
