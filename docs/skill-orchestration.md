# Skill Orchestration

This repository treats skills as behavior overrides, not code. Skills influence architecture, implementation, testing, review, and security posture.

## Mandatory Pre-Action Checks

Before implementation, refactoring, debugging, or review:

1. Skill decision
   - Check whether an installed skill applies.
   - Prefer installed, task-specific skills over generic behavior.
   - If a required skill is missing, output `INSTALL REQUIRED SKILL` and the exact install command.
   - Do not install new skills without explicit user confirmation.
2. MCP tool decision
   - Check whether filesystem, GitHub, browser, blockchain RPC, or other external tools are required.
   - Use only the minimum tool surface needed for the task.
3. Architecture decision
   - Classify the target as L1, L2, service, module, docs, tooling, or test.
   - Identify system boundaries and prevent monolithic drift.

## Installed Skill Baseline

For this Cosmos L1 project, the expected local skill set includes:

- `software-crypto-web3`: broad blockchain/Web3 architecture and security guidance.
- `cosmos-vulnerability-scanner`: Cosmos SDK and CosmWasm vulnerability review patterns.
- Existing Cosmos/TON development skills already present in the Codex profile remain available but must not override Cosmos L1 requirements unless directly relevant.

If these skills are missing in a fresh environment, install them with the user's approval:

```powershell
npx skills add https://github.com/trailofbits/skills --skill cosmos-vulnerability-scanner
npx add-skill software-crypto-web3
```

## Skill Use Rules

- Load relevant skill rules before generating code or audit findings.
- Skill rules override generic preferences when they are more specific and do not conflict with repository governance.
- Repository governance still applies: SOLID, DRY, KISS, modular boundaries, no hardcoded consensus values, deterministic behavior, and testability.

## Required Output Order

For implementation or review tasks:

1. Skill decision
2. Architecture decision
3. Module design
4. Implementation
5. Tests
6. Security review
7. Git commit plan
8. Refactor suggestions

## Git Integration

- Group changes into logical PR-sized commits.
- Do not mix skill-policy changes with chain scaffold, module implementation, or generated code.
- Use Conventional Commits with a clear scope.
- Push completed commits to `origin/main` unless explicitly told not to.

## Security Model

Every skill-guided change must still consider:

- Malicious input.
- Replay or duplicate execution.
- State corruption.
- Invalid transactions.
- Governance abuse.
- Consensus non-determinism.
