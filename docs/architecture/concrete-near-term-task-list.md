# Concrete Near-Term Task List

Near-term implementation should be split into small pull requests:

1. Audit existing validator/economics modules and map them to this spec.
2. Add missing params validation for validator power cap and commission policy.
3. Add effective power and overflow stake queries.
4. Add concentration snapshot query.
5. Add cap math tests for 100/150/200/300 validator scenarios.
6. Add fee split accounting tests.
7. Add inflation bounds tests.
8. Add supply invariant tests.
9. Add validator score state and query tests.
10. Add progressive downtime design or document why standard slashing is enough for v1.
11. Add nomination pool accounting tests.
12. Add AVM smoke and malicious contract tests.
13. Add public testnet finality measurement script.
14. Add documentation for validators and delegators.
15. Add CI gate for critical unit/integration tests.

Every near-term implementation change must state:

- what consensus/economics behavior changes;
- what params are added or changed;
- what tests were added;
- what migration risk exists;
- whether public docs need updates.

## Execution Notes

The list is intentionally ordered from audit and safety toward public testnet readiness. Validator/economics audit comes first because existing modules must be mapped before new behavior is layered onto them. Power-cap, commission, effective-power, overflow-stake, concentration, fee split, inflation, supply, score, downtime, and nomination-pool work should remain small enough to review independently.

AVM, finality measurement, docs, and CI work are public-testnet enablers. They should not bypass P0/P1 correctness gates, but they should start early enough that operators, validators, delegators, and contract developers can test real workflows before public testnet launch.

The implementation gate is `DefaultAetraNearTermTaskListEvidence` in `app/params/near_term_task_list.go`.

## Acceptance Gate

Required behavior:

- missing audit task fails readiness;
- missing params validation task fails readiness;
- missing query tasks fail readiness;
- missing cap, fee, inflation, supply, score, downtime, nomination-pool, AVM, finality, docs, or CI tasks fail readiness;
- missing consensus/economics behavior note fails readiness;
- missing params, tests, migration-risk, or public-docs checklist items fail readiness;
- duplicate or unexpected task/checklist entries fail readiness.
