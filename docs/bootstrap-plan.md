> Deprecated/migration note: this document contains historical native asset-factory or native exchange references. Those runtime modules have been removed from the active app graph; token, NFT, market, and exchange-style application logic now targets AVM contracts and standards such as AFT-44/ANFT-66.
# Bootstrap Implementation Plan

## Increment 1: Architecture Bootstrap

Commit: `docs(architecture): define sovereign l1 bootstrap plan`

Deliverables:
- Architecture docs.
- Module boundaries.
- Genesis and params model.
- Security and testing strategy.
- Repo structure proposal.

Verification:
- Git status review.
- Markdown file review.
- Privacy guard review.

## Increment 2: Chain Scaffold

Commit: `feat(app): scaffold sovereign cosmos sdk chain`

Planned command:

```powershell
ignite scaffold chain github.com/SoftwareMaestro16/L1-Blockchain --address-prefix ae --default-denom naet --no-module --skip-git --path .
```

Post-scaffold work:
- Pin exact Cosmos SDK version in `go.mod`.
- Confirm generated app uses the expected SDK release family.
- Split app wiring if generated files become too large.
- Add first CI workflow after the Go module exists.
- Apply the engineering governance checklist before and after reviewing generated code.

## Increment 3: Tokenfactory Module

Commit: `feat(tokenfactory): add custom denom lifecycle module`

Deliverables:
- Proto Msg, Query, state, genesis.
- Keeper and MsgServer implementation.
- Unit, adversarial, integration, and genesis tests.

## Increment 4: Fees Module

Commit: `feat(fees): add protocol fee accounting module`

Deliverables:
- Params-controlled distribution policy.
- Bank integration through keeper interfaces.
- Tests for conservation, bounds, and governance authority.

## Increment 5: DEX Module

Commit: `feat(dex): add constant product amm module`

Deliverables:
- Pool lifecycle, swaps, LP accounting.
- Integer-only invariant math.
- Adversarial tests for rounding, reserve solvency, and slippage.

## Increment 6: Security And CI Hardening

Commit: `test(security): add adversarial and determinism test suites`

Deliverables:
- Cross-module integration tests.
- Determinism tests.
- Static checks, lint, and GitHub Actions CI.
- Pre-release audit checklist.
