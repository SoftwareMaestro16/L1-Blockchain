# x/market Migration Marker

Status: deprecated-native-app-logic.

Target: AVM market/resource contract or scheduler fee policy.

Market product logic should move to AVM contracts. Any protocol fee-market or
scheduler policy that must remain native should be split into explicit protocol
modules before this package is removed.
