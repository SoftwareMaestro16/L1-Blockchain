# API, CLI, Query, and Event Surface

Every Aetra module must expose enough surface for validators, wallets, explorers, dashboards, monitoring, governance tooling, and incident response. A module is not production-ready if its keeper logic exists but users cannot inspect state, submit authorized transactions, subscribe to stable events, or reproduce behavior through documented commands.

## 30.1 CLI Requirements

Required command categories:

```text
aetrad query aetra-staking-policy ...
aetrad query aetra-economics ...
aetrad query aetra-validator-score ...
aetrad tx aetra-staking-policy ...
aetrad tx aetra-economics ...
aetrad tx aetra-validator-score ...
```

Commands must support:

- json output;
- height query where applicable;
- pagination where applicable;
- clear errors;
- examples in docs.

## 30.2 gRPC/REST Requirements

Every query must have:

- protobuf definition;
- gRPC service;
- REST gateway mapping if project supports it;
- response examples;
- tests where feasible.

The protobuf definition is the canonical contract for machine clients. It must define request messages, response messages, pagination fields where needed, height-safe behavior where applicable, and stable field names. Query response changes must be treated as public API changes once testnet users, explorers, wallets, or dashboards depend on them.

The gRPC service is the primary typed query API. CLI query commands should call the same query service path rather than implementing separate state logic. REST gateway mappings are required where the project supports gRPC gateway exposure; if a route is intentionally not exposed through REST, the module must document why and keep that exception explicit.

Response examples must include realistic JSON for successful responses and at least one clear error shape for invalid request or not found cases. Tests should cover query handler behavior, pagination, invalid requests, stable response fields, REST route registration where feasible, and CLI examples where the command layer exists.

## Module Surface Contract

Required modules:

- `aetra-staking-policy`;
- `aetra-economics`;
- `aetra-validator-score`.

Each required module must expose:

- CLI query commands;
- CLI tx commands for governance-authorized or signer-authorized messages;
- protobuf query definitions;
- gRPC query services;
- REST gateway routes where applicable;
- response examples;
- query tests where feasible;
- deterministic event names and bounded attributes;
- examples in validator, wallet, explorer, or operator docs.

CLI query commands must support `--output json`. Historical reads must support height queries where the underlying state is safe to query by height. List queries must support pagination and must not force clients to scan unbounded state. Errors must be explicit enough for automation and must not depend on free-form panic text.

CLI tx commands must validate signer, authority, addresses, params, fees, and message shape before broadcast where possible. Tx examples must include command skeletons, required flags, and JSON output examples.

## Query Requirements

Required query behavior:

- protobuf definition for every query;
- gRPC service for every query;
- REST gateway mapping if project supports it;
- response examples for public queries;
- tests where feasible;
- current params query for every module;
- module-specific state queries for validators, economics epochs, score records, and warnings;
- pagination for list endpoints;
- height query support where applicable;
- stable response fields for explorers and wallets;
- clear not-found and invalid-request errors;
- gRPC service definitions are the canonical machine-readable API;
- REST routes are generated or documented where applicable.

## Event Requirements

Required event behavior:

- event names must be stable and module-prefixed;
- attributes must be bounded and deterministic;
- attributes must avoid unbounded user-controlled label values;
- state-changing txs must emit enough data for explorers and indexers;
- parameter updates must emit old value, new value, authority, and activation height or epoch where applicable;
- slashing, reward, fee, treasury, cap, score, and warning events must be indexer-compatible.

## Acceptance Gate

The implementation catalog is `DefaultAPISurfaceModuleSpecs` in `observability/api_surface.go`.

Required catalog properties:

- every required module has query and tx CLI categories;
- every required query surface has protobuf definition;
- every required query surface has gRPC service;
- every required query surface has REST gateway mapping where supported;
- response examples are mandatory;
- query tests are mandatory where feasible;
- every required CLI command supports json output;
- height query support is explicit for query commands;
- pagination support is explicit where list queries exist;
- clear error behavior is explicit;
- examples in docs are mandatory;
- gRPC, REST where applicable, and event surfaces are tracked;
- missing required modules or surfaces fail readiness tests.
