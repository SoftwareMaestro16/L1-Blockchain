# Engineering Governance

This standard applies to design, implementation, refactoring, testing, debugging, and review work in this repository.

## Pre-Code Checklist

Before writing or changing code:

1. Architecture check
   - Define the system boundary.
   - Identify required modules and dependencies.
   - Confirm the change is modular and not drifting toward a monolith.
2. Design principles check
   - Apply SOLID, DRY, KISS, separation of concerns, and composition-first design.
   - For consensus or distributed logic, prove deterministic behavior.
3. Scalability check
   - Avoid shared mutable state and inappropriate global singletons.
   - Bound loops, state growth, and gas-heavy paths.
4. Security check
   - Identify attacker-controlled inputs.
   - Check state corruption, bypass, replay, duplication, and malformed data paths.
5. Test strategy check
   - Define unit, integration, adversarial, edge, and failure scenarios before implementation.

## Code Structure Rules

- No file should exceed 300-500 lines; split earlier when responsibilities diverge.
- No magic numbers or embedded consensus configuration in logic.
- Blockchain configuration must live in params, genesis, or explicit app configuration.
- Each module owns one responsibility and must be testable in isolation.
- Use explicit interfaces and dependency injection where possible.
- Prefer small functions with predictable control flow and no hidden side effects.

## Testing Requirements

Every feature must include the applicable layers:

- Unit tests for normal cases and boundary conditions.
- Adversarial tests for invalid input, corrupted state, replayed actions, malformed data, and malicious behavior.
- Integration tests for module interaction and full lifecycle flows.
- Determinism tests for consensus or distributed-system code.

## Post-Code Review Loop

After implementation:

1. Review the design.
2. Identify complexity hotspots.
3. Detect duplication.
4. Detect oversized modules or files.
5. Check SOLID and boundary violations.
6. Run or document tests.
7. Improve structure immediately when risk is low and scope is local.
8. Record remaining refactor opportunities.

## Git Workflow

- Treat every change as a PR-sized increment.
- Do not mix unrelated changes in one commit.
- Use Conventional Commits with clear scope.
- Include the reason and improvement in the implementation summary.
- Run relevant checks before commit.
- Push completed commits to a topic branch and open a PR; direct pushes to `origin/main` are reserved for explicitly approved emergency recovery.
- Critical/high security findings must be fixed before push/merge, or documented with a linked issue, severity assessment, owner, and decision.

## Required Response Order

Implementation summaries must use this order:

1. Architecture decision
2. Module breakdown
3. Implementation
4. Tests
5. Security review
6. Refactor suggestions
7. Next iteration plan

## Architecture Evolution

Every iteration must classify future improvements:

- MUST FIX: correctness, security, determinism, or build blockers.
- SHOULD FIX: maintainability, performance, or test-depth improvements.
- NICE TO HAVE: future ergonomics or polish.

Each refactor plan should identify affected files, expected change, risk level, and migration steps when relevant.
