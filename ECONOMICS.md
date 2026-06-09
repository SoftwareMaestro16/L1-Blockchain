# Aetheris Protocol Economics Specification

Status: Internal design document
Scope: AET economic system evolution
Visibility: Private, not for public repository inclusion

## 1. Baseline Economic Context

### 1.1 Native Asset

- Native denom: `naet`
- Display denom: `AET`
- Precision: `1 AET = 1,000,000,000 naet`
- Primary uses:
  - Validator staking
  - Delegation
  - Transaction fees
  - Minted staking rewards
  - Execution and storage charges

### 1.2 Current Supply Model

- Current supply policy:
  - Uncapped proof-of-stake issuance
  - No fixed maximum supply
  - Inflation bounded by protocol parameters
- Active inflation bounds:
  - Minimum inflation: `1%`
  - Target inflation: `3%`
  - Maximum inflation: `5%`
- Target stake ratio:
  - `67%` of circulating supply bonded or economically locked for consensus security

### 1.3 Current Reward and Fee Flow

- Validator and delegator rewards are distributed through the existing distribution flow.
- Current fee denom is restricted to `naet`.
- Current fee split:
  - `98%` validator and delegator reward flow
  - `2%` community pool allocation
- Existing fee model includes:
  - Base fee
  - Congestion-based fee scaling
  - Transaction gas limits
  - Block gas limits

### 1.4 Current Economic Control Surface

- Inflation controller:
  - Target-based staking participation logic exists.
  - Network activity coupling exists in design but needs stronger production integration.
- Burn controller:
  - Burn ratio model exists.
  - Production integration is incomplete.
- Deflation guard:
  - Design exists.
  - Enforcement needs explicit parameterization, tests, and telemetry.
- AVM execution economy:
  - Storage fee per byte exists.
  - Forwarding fee exists.
  - Deployment cost exists.
  - Full integration into the protocol-wide economic loop is incomplete.

## 2. Economic Goals

### 2.1 Long-Term Goals

- Maintain sufficient bonded economic weight to make validator corruption expensive.
- Keep validator participation economically viable across different network activity regimes.
- Preserve delegation as a simple, liquid, and risk-aware security contribution mechanism.
- Prevent fee markets from becoming unstable under congestion, spam, or low activity.
- Couple issuance, burn, fees, and execution costs into one measurable economic control loop.
- Reduce long-term state growth by making persistent storage economically accountable.
- Improve validator decentralization without weakening liveness or operational quality.
- Make all major economic parameters observable, governable, testable, and bounded.

### 2.2 Optimal Economic State

The protocol is in an optimal economic state when all of the following conditions hold:

- Bonded stake remains near the target stake ratio without forcing excessive issuance.
- Validator rewards cover expected infrastructure, monitoring, key management, and compliance costs for reliable operators.
- Delegators can compare validators using risk, performance, commission, and concentration signals.
- Validator stake is not excessively concentrated in a small active set.
- Fee levels respond predictably to congestion without breaking normal user activity.
- Transaction spam has a direct and escalating economic cost.
- Persistent state usage creates recurring or lifecycle-based economic cost.
- Burns reduce net issuance during sustained high activity without creating uncontrolled deflation.
- Slashing penalties are proportional to security damage, operator fault severity, and recovery requirements.
- Treasury and community pool funding are tied to explicit protocol maintenance needs.

### 2.3 Economic Invariants

- AET must remain the primary economic asset for staking, fees, rewards, slashing, and execution charges.
- No economic subsystem should be able to drive net issuance or net burn outside configured bounds.
- Validator reward distribution must remain deterministic at consensus level.
- Fee computation must be deterministic and bounded per block.
- Slashing must be deterministic, auditable, and resistant to reward manipulation.
- Any adaptive controller must expose parameters, state, and events for observability.
- Storage pricing must make long-lived state more expensive than transient execution.

## 3. System Weaknesses Analysis

### 3.1 Missing or Incomplete Economic Mechanisms

- Burn controller is not fully wired into the production fee and reward flow.
- Inflation controller is only partially coupled to real network activity.
- Deflation guard lacks complete production enforcement and test coverage.
- Extended slashing design is not fully integrated with burn, treasury, and reporter reward paths.
- Epoch-based validator selection exists as a target but is not fully productionized.
- AVM storage and forwarding fees are not fully reconciled with the global fee market.
- No complete state rent or state cleanup incentive exists for long-term state growth.
- Validator reputation and reliability are not first-class inputs to delegation UX or validator selection policy.
- No explicit stake concentration penalty or reward dampening exists.
- No formal economic circuit breaker exists for abnormal activity, fee spikes, or controller instability.

### 3.2 Inflation Model Risks

- Uncapped supply can weaken long-term predictability if issuance is not visibly tied to security need.
- A fixed target inflation rate can overpay security during low-risk, low-activity periods.
- Inflation may underpay security if validator costs rise faster than reward flow.
- Target stake ratio logic may encourage over-delegation to high-commission or high-concentration validators if risk is not priced.
- Inflation changes can become noisy if driven by short-term activity metrics.
- Poor burn integration can create conflicting supply signals between issuance and fee destruction.
- Without a bounded net issuance target, governance may struggle to reason about long-term supply drift.

### 3.3 Validator Incentive Weaknesses

- Commission cap alone does not ensure good validator behavior or decentralization.
- Rewards are primarily stake-weighted, so large validators compound faster unless concentration is controlled.
- Baseline slashing covers core faults but does not fully price softer reliability failures or repeated marginal performance.
- Reporter rewards are not fully production wired, reducing incentives to surface slashable evidence quickly.
- Validator selection improvements are not fully connected to economic outcomes.
- Operators with weak infrastructure may remain profitable if delegation inflows ignore performance risk.
- New validators face a bootstrap disadvantage against existing high-stake validators.

### 3.4 Staking Centralization Risks

- Delegators may concentrate stake into the largest or most visible validators.
- Commission competition can produce a race toward unsustainable validator pricing.
- High self-delegation requirements can improve alignment but may reduce operator diversity if too high.
- Low self-delegation requirements can increase weak-commitment validator participation.
- Redelegation behavior may lag behind validator risk changes.
- Delegators may not have enough structured information to price validator failure risk.
- Concentrated voting power can increase collusion and governance capture risk.

### 3.5 Fee Model Inefficiencies

- A static fee split may overpay validators during high activity and underfund public maintenance.
- Base fee and congestion scaling require stronger parameter bounds and simulation coverage.
- Native-denom-only fees simplify accounting but concentrate all fee pressure into AET liquidity.
- Anti-spam costs may be too weak if congestion multipliers react slowly.
- Fee burn is not fully integrated, limiting the ability to offset issuance during high demand.
- Storage and execution fees are not fully normalized against block-level transaction fees.
- Block gas and transaction gas limits need continuous calibration against validator hardware expectations.

## 4. Improvement Roadmap

### 4.1 Validator Economy Improvements

#### 4.1.1 Validator Selection Logic

Objectives:

- Productionize epoch-based validator selection.
- Reduce active-set concentration.
- Preserve deterministic, auditable validator admission.

Actionable tasks:

- Define validator selection inputs:
  - Bonded stake
  - Self-delegation
  - Uptime
  - Missed block history
  - Slash history
  - Commission rate
  - Stake concentration score
  - Minimum operational metadata completeness
- Add a validator eligibility score with deterministic calculation.
- Add epoch boundary validation for active-set changes.
- Add simulation tests for validator churn under normal, adversarial, and low-participation conditions.
- Add a maximum per-validator voting power soft cap.
- Add reward dampening above the soft cap rather than forced exclusion.
- Emit epoch selection events with all score components.
- Add governance parameters for:
  - Minimum self-delegation
  - Maximum active set
  - Concentration soft cap
  - Score weights
  - Epoch length

Acceptance criteria:

- Active-set selection is deterministic across nodes.
- Validator set transitions are bounded per epoch.
- A validator above the concentration soft cap receives reduced marginal rewards.
- Tests cover stake splitting, stake concentration, and validator churn scenarios.

#### 4.1.2 Staking Incentives

Objectives:

- Reward reliable validators.
- Reduce excessive concentration.
- Maintain viable economics for smaller operators.

Actionable tasks:

- Add reliability-weighted reward adjustment within bounded limits.
- Add a new validator bootstrap band for qualified low-stake validators.
- Add minimum performance thresholds for full reward eligibility.
- Add a reward curve that gradually reduces marginal reward share above concentration thresholds.
- Add explicit modeling for validator operating cost assumptions.
- Add telemetry for validator reward per unit of voting power.
- Add validator profitability dashboards for active, standby, and candidate validators.

Acceptance criteria:

- Reward adjustments cannot exceed configured bounds.
- Validators cannot gain more reward by intentionally reducing uptime.
- Reward dampening is predictable and visible before delegation.
- Bootstrap incentives expire automatically after configured conditions.

#### 4.1.3 Slashing Model

Objectives:

- Make penalties proportional to security harm.
- Integrate burn, treasury, and reporter paths.
- Penalize repeated poor behavior more strongly.

Actionable tasks:

- Define severity classes:
  - Minor downtime
  - Major downtime
  - Repeated downtime
  - Equivocation
  - Evidence manipulation
  - Validator key compromise response failure
- Extend slashing parameters per severity class.
- Route slashed funds according to deterministic split:
  - Burn allocation
  - Treasury allocation
  - Reporter allocation where evidence is accepted
- Add reporter reward caps to prevent evidence farming incentives.
- Add repeat-offense multipliers with decay over time.
- Add slashing event telemetry for penalty amount, reason, validator, and fund routing.
- Add integration tests for each slashing class and routing path.

Acceptance criteria:

- Slashing cannot mint or misroute funds.
- Reporter reward is paid only for accepted, non-duplicate evidence.
- Repeat-offense multiplier is deterministic and bounded.
- Slashing fund routing is covered by invariant tests.

#### 4.1.4 Validator Decentralization

Objectives:

- Reduce correlated control of consensus weight.
- Improve operator diversity.
- Make concentration visible to delegators and governance.

Actionable tasks:

- Add concentration metrics:
  - Validator voting power share
  - Top-N voting power share
  - Delegator concentration per validator
  - Self-delegation ratio
  - Commission concentration by voting power
- Add concentration warnings to validator metadata and API responses.
- Add reward dampening for validators above configured voting power thresholds.
- Add minimum self-delegation ratio or absolute minimum, selected by governance parameter.
- Add stake movement incentives for delegating away from highly concentrated validators.
- Add dashboards for active-set and standby-set concentration.

Acceptance criteria:

- Concentration metrics are queryable through node APIs.
- Reward dampening cannot reduce rewards below the configured safety floor.
- Delegators can identify concentration risk before delegation.

### 4.2 Nominator and Delegation Economy Improvements

#### 4.2.1 Delegation Incentives

Objectives:

- Improve delegator risk-adjusted yield.
- Make validator choice economically informed.
- Reduce passive concentration.

Actionable tasks:

- Publish validator risk scores as queryable data.
- Add estimated net yield fields after commission, performance adjustment, and concentration adjustment.
- Add delegation concentration warnings.
- Add optional delegation policy templates for wallets and interfaces:
  - Low-risk
  - High-availability
  - Decentralization-supporting
  - Maximum yield within configured risk bounds
- Add reward preview calculations for redelegation decisions.
- Add minimum disclosure fields for validators:
  - Commission
  - Max commission change
  - Uptime
  - Slash history
  - Self-delegation
  - Concentration status

Acceptance criteria:

- Delegator-facing yield estimates use the same reward inputs as the distribution system.
- Risk score components are transparent and queryable.
- Delegation policy templates are advisory and do not custody or move stake automatically.

#### 4.2.2 Protection Against Validator Capture

Objectives:

- Reduce risk of delegator funds amplifying captured validators.
- Make validator control changes economically visible.
- Increase cost of coordinated capture attempts.

Actionable tasks:

- Add validator metadata change tracking.
- Add cooldown or warning period for material validator identity changes.
- Add a commission-change risk flag when validators rapidly increase commission.
- Add capture-risk indicators:
  - Sudden delegation inflow
  - Rapid commission change
  - Recent slash
  - Self-delegation withdrawal
  - Operator metadata inconsistency
- Add optional delegator alert events for high-risk validator changes.
- Add governance-controlled limits for maximum commission change per interval.

Acceptance criteria:

- Validator identity and commission changes emit machine-readable events.
- Delegators can react before material commission changes take effect.
- Risk flags do not directly slash validators unless tied to existing slashable behavior.

#### 4.2.3 Yield-Risk Balancing

Objectives:

- Make reward expectations reflect validator risk.
- Reduce yield chasing into fragile validators.
- Improve delegation distribution.

Actionable tasks:

- Define risk-adjusted yield formula:
  - Gross reward rate
  - Commission
  - Historical uptime adjustment
  - Slash probability proxy
  - Concentration adjustment
  - Unbonding liquidity cost
- Add query endpoints for projected reward range.
- Add historical validator reward variance metrics.
- Add redelegation cost visibility.
- Add delegation simulator using recent chain state.

Acceptance criteria:

- Risk-adjusted yield is reproducible from on-chain and node-queryable inputs.
- Reward projections include uncertainty bands.
- The simulator handles commission changes, slashing, and reward dampening.

### 4.3 Fee Market Evolution

#### 4.3.1 Dynamic Fee Optimization

Objectives:

- Keep blockspace usable under normal load.
- Increase costs under congestion.
- Make fee changes predictable.

Actionable tasks:

- Define target block utilization for fee control.
- Add bounded fee adjustment per block or per window.
- Add smoothing windows to prevent short-lived activity spikes from causing excessive fee changes.
- Add minimum and maximum base fee parameters.
- Add fee estimation queries for wallets and services.
- Add simulation tests for low load, steady load, burst load, and spam load.

Acceptance criteria:

- Base fee adjustment is deterministic.
- Fee changes are bounded by governance parameters.
- Fee estimator tracks actual inclusion costs within a defined tolerance.

#### 4.3.2 Congestion Response

Objectives:

- Respond quickly to spam and congestion.
- Avoid unnecessary cost spikes during brief bursts.
- Protect critical protocol transactions.

Actionable tasks:

- Separate congestion signals:
  - Block gas utilization
  - Mempool pressure
  - Failed execution rate
  - Repeated sender activity
  - State-write pressure
- Add per-resource fee multipliers for compute, storage writes, message forwarding, and deployment.
- Add short-window spam surcharge for repeated failed transactions.
- Add maximum transaction gas limits by message class.
- Add reserve policy for critical protocol operations if required by consensus safety.

Acceptance criteria:

- Congestion fee increase occurs before sustained full blocks persist.
- Failed spam transactions become progressively more expensive.
- Critical protocol operations remain bounded and auditable.

#### 4.3.3 Fee Distribution Models

Objectives:

- Align fee revenue with validator security, public maintenance, burn, and long-term state cost.
- Avoid a static split that ignores network conditions.

Actionable tasks:

- Replace static split with parameterized distribution buckets:
  - Validator and delegator rewards
  - Community pool
  - Burn
  - State maintenance reserve
  - Security reserve
- Define base allocation and activity-dependent allocation.
- Increase burn allocation during sustained high activity.
- Increase state maintenance allocation during high state growth.
- Preserve validator reward floor during low activity.
- Add accounting events for each fee allocation bucket.

Acceptance criteria:

- Fee distribution sums exactly to collected fees.
- Validator reward floor is enforced.
- Burn and reserve allocations are queryable and test-covered.

#### 4.3.4 Anti-Spam Economic Design

Objectives:

- Make spam costly before it damages liveness.
- Price repeated resource abuse directly.
- Preserve access for normal users.

Actionable tasks:

- Add account-level short-window activity scoring.
- Add sender-local surcharge for repeated failed transactions.
- Add state-write surcharge for high-frequency state expansion.
- Add deployment surcharge during deployment congestion.
- Add minimum fee enforcement for all executable messages.
- Add mempool admission rules aligned with final fee rules.
- Add tests for fee bypass attempts.

Acceptance criteria:

- Mempool and block execution use compatible fee validation.
- Repeated failed transaction spam increases sender cost.
- Anti-spam logic is deterministic and bounded.

### 4.4 Inflation and Burn Model Redesign

#### 4.4.1 Inflation Coupled to Network Activity

Objectives:

- Issue rewards according to security need and actual fee revenue.
- Avoid overpaying or underpaying validator security.
- Smooth issuance changes.

Actionable tasks:

- Define controller inputs:
  - Bonded stake ratio
  - Validator operating cost index
  - Fee revenue
  - Active validator count
  - Slashing risk events
  - Network activity score
  - Treasury reserve health
- Define inflation adjustment function with:
  - Minimum inflation bound
  - Maximum inflation bound
  - Per-window change limit
  - Smoothing window
  - Emergency freeze parameter
- Add net issuance reporting:
  - Gross minted AET
  - Burned AET
  - Net supply change
  - Security spend per block
- Add simulation tests across low, normal, high, and adversarial activity.

Acceptance criteria:

- Inflation cannot exceed configured bounds.
- Inflation changes are explainable from queryable inputs.
- Net issuance is reported per epoch and per accounting period.

#### 4.4.2 Fully Integrated Burn Mechanics

Objectives:

- Make burn a production accounting path.
- Offset issuance during sustained activity.
- Avoid uncontrolled deflation.

Actionable tasks:

- Wire burn allocation into fee distribution.
- Wire burn allocation into slashing distribution where configured.
- Add burn module accounting events.
- Add burn caps by epoch.
- Add burn floor disabled by default unless governance defines a minimum burn requirement.
- Add invariant tests for burn accounting.
- Add queries for cumulative burned supply and recent burn rate.

Acceptance criteria:

- Burned funds are removed from spendable supply.
- Burn accounting cannot conflict with reward distribution.
- Burn cap prevents excessive deflation during abnormal fee spikes.

#### 4.4.3 Deflation Control Mechanisms

Objectives:

- Prevent uncontrolled supply contraction.
- Preserve security budget.
- Keep supply policy predictable.

Actionable tasks:

- Define net issuance lower bound per epoch.
- Add deflation guard:
  - If burn exceeds configured net issuance target, reduce burn allocation before reducing security rewards.
  - If fees spike abnormally, route excess to reserves before further burn.
  - If bonded stake falls below safety threshold, prioritize validator rewards.
- Add controller telemetry for guard activation.
- Add governance parameters for:
  - Net issuance floor
  - Burn cap
  - Reserve cap
  - Security reward floor

Acceptance criteria:

- Deflation guard activation is deterministic and visible.
- Security reward floor has priority over burn allocation.
- Net issuance cannot fall below configured lower bound unless governance explicitly permits it.

#### 4.4.4 Long-Term Supply Stabilization

Objectives:

- Reduce long-term supply drift.
- Use fee revenue to offset issuance where activity supports it.
- Preserve validator security through low-activity periods.

Actionable tasks:

- Define target net issuance range per year.
- Define conditions for moving toward lower net issuance:
  - Sustained fee revenue
  - Stable bonded stake
  - Healthy validator set
  - Low slashing rate
  - Adequate reserves
- Define conditions for temporarily higher issuance:
  - Low bonded stake
  - Validator attrition
  - Low fee revenue
  - Elevated security risk
- Add supply projection reports for governance.
- Add stress tests for 1-year, 3-year, and 5-year supply scenarios.

Acceptance criteria:

- Supply projections are generated from protocol parameters and recent activity.
- Governance can reason about net issuance without manual off-chain accounting.
- Stabilization policy never bypasses consensus reward accounting.

### 4.5 Storage and Execution Economy

#### 4.5.1 Storage Pricing

Objectives:

- Charge persistent state according to long-term cost.
- Discourage unnecessary state growth.
- Preserve predictable application deployment costs.

Actionable tasks:

- Define state write fee per byte.
- Define state update fee separate from first-write fee.
- Define state delete refund policy with caps.
- Add storage footprint metering per account or contract.
- Add storage usage query endpoints.
- Add storage fee events for write, update, delete, and refund.

Acceptance criteria:

- Storage fees are deterministic from state delta.
- Delete refunds cannot exceed original economic cost after configured decay.
- Storage footprint is queryable for accounts and contracts.

#### 4.5.2 Long-Term State Rent

Objectives:

- Prevent permanent unpaid state accumulation.
- Create lifecycle cost for persistent data.
- Allow graceful cleanup before state becomes inactive.

Actionable tasks:

- Define rent unit:
  - Bytes stored
  - Duration
  - Account or contract class
- Add prepaid storage balance model.
- Add warning state before rent exhaustion.
- Add inactive state handling:
  - Freeze
  - Limited execution
  - Cleanup eligibility
  - Recovery path if funded
- Define exempt or special-state categories only where required for protocol correctness.
- Add governance parameters for rent rate and grace periods.

Acceptance criteria:

- State rent cannot delete consensus-critical state accidentally.
- Accounts and contracts can query rent status.
- Rent exhaustion behavior is deterministic and test-covered.

#### 4.5.3 Execution Cost Optimization

Objectives:

- Align execution fees with validator resource costs.
- Reduce inefficient workload patterns.
- Keep execution metering predictable.

Actionable tasks:

- Calibrate gas costs for compute-heavy operations.
- Separate compute, storage read, storage write, and message forwarding cost classes.
- Add execution traces for fee debugging.
- Add deployment cost scaling by code size and initialization state.
- Add fee estimation for contract deployment and async message flows.
- Add benchmarks tied to protocol gas tables.

Acceptance criteria:

- Gas costs are backed by repeatable benchmarks.
- Contract deployment fee estimates are within defined tolerance.
- Message forwarding fees cover actual routed workload.

#### 4.5.4 State Bloat Prevention

Objectives:

- Make state expansion economically bounded.
- Improve cleanup incentives.
- Detect abnormal state growth early.

Actionable tasks:

- Add state growth telemetry:
  - Bytes added per block
  - Bytes removed per block
  - Net state growth per epoch
  - Top state-growth accounts or contracts
- Add state-growth surcharge during high-growth periods.
- Add delete refund decay to prevent storage churn attacks.
- Add reserve allocation for state maintenance.
- Add alerts for abnormal state growth.

Acceptance criteria:

- State growth is visible at block and epoch levels.
- High-growth periods increase storage expansion cost.
- Refund logic is resistant to write-delete fee extraction.

### 4.6 Security and Game-Theory Hardening

#### 4.6.1 Economic Attack Prevention

Objectives:

- Increase cost of economically rational attacks.
- Detect abnormal reward, fee, and stake behavior.
- Reduce attack profitability.

Actionable tasks:

- Model attack classes:
  - Stake concentration
  - Commission bait-and-switch
  - Reward manipulation
  - Fee spam
  - State bloat
  - Evidence spam
  - Delegation capture
- Add automated economic invariant tests.
- Add simulations for validator cartel behavior.
- Add monitoring for abnormal stake movements.
- Add circuit breaker parameters for fee controller instability.

Acceptance criteria:

- Each attack class has a measurable cost and mitigation.
- Invariant tests run in CI.
- Circuit breaker activation is deterministic and governed by parameters.

#### 4.6.2 Sybil Resistance

Objectives:

- Reduce benefit from splitting one operator into many validators.
- Preserve room for legitimate new validators.
- Avoid identity assumptions that cannot be enforced at protocol level.

Actionable tasks:

- Add stake-splitting simulation tests.
- Use economic constraints instead of unverifiable identity rules.
- Apply bootstrap rewards only once per economic eligibility window.
- Add correlated behavior detection as advisory telemetry.
- Add reward dampening for validator groups only where correlation proof is deterministic.
- Add minimum self-delegation and performance requirements for active-set eligibility.

Acceptance criteria:

- Splitting stake across validators does not bypass concentration policy.
- Advisory correlation metrics do not affect consensus unless deterministic.
- New legitimate validators retain a clear path into the active set.

#### 4.6.3 Validator Collusion Resistance

Objectives:

- Increase cost of coordinated validator behavior.
- Reduce reward from collusive concentration.
- Improve evidence and monitoring paths.

Actionable tasks:

- Add top-N voting power concentration limits or reward dampening.
- Add collusion scenario simulations.
- Add monitoring for synchronized commission changes.
- Add monitoring for correlated downtime.
- Add evidence routing for consensus faults.
- Add governance reports for concentration and correlated behavior.

Acceptance criteria:

- Collusion scenarios are included in economic stress tests.
- Concentration dampening reduces marginal reward from coordinated stake accumulation.
- Monitoring outputs are queryable and auditable.

#### 4.6.4 Incentive Alignment

Objectives:

- Align validators, delegators, users, and protocol reserves.
- Ensure each participant pays for the resources or risks they create.
- Make incentives transparent enough for governance review.

Actionable tasks:

- Define participant-level incentive maps.
- Add explicit reward and penalty accounting for each participant class.
- Add per-epoch economic reports:
  - Issuance
  - Burns
  - Fees collected
  - Rewards distributed
  - Slashed funds
  - Reserve changes
  - State growth
  - Validator concentration
- Add governance dashboards for parameter impact.
- Add pre-upgrade economic simulation requirement for parameter changes.

Acceptance criteria:

- Economic reports reconcile with supply accounting.
- Parameter changes include projected impact before activation.
- Participant incentives are documented and testable.

## 5. Protocol Economic Module Design

### 5.1 Staking Enhancements Module

Purpose:

- Extend staking economics without replacing the existing staking system.
- Add validator scoring, concentration controls, and reward adjustment inputs.

Inputs:

- Validator bonded stake
- Self-delegation
- Commission parameters
- Uptime and missed block data
- Slash history
- Active-set size parameters
- Concentration thresholds
- Epoch parameters

Outputs:

- Validator eligibility status
- Validator score
- Concentration score
- Reward adjustment factor
- Active-set recommendation for epoch transition
- Validator economic events

Failure modes:

- Score instability causes validator churn.
- Incorrect weighting over-penalizes small validators.
- Concentration dampening reduces validator viability too aggressively.
- Metadata-dependent inputs become stale or inconsistent.

Integration points:

- `x/staking` validator set and delegation state
- `x/distribution` reward calculation
- Slashing event stream
- Epoch transition logic
- Governance parameter updates
- Telemetry and query endpoints

Required tasks:

- Define score formula and parameter bounds.
- Add deterministic score calculation.
- Add simulation suite for validator distribution.
- Add invariant tests for reward adjustment bounds.

### 5.2 Validator Reputation System

Purpose:

- Provide risk and reliability signals for delegators, validator selection, and governance.
- Keep reputation advisory unless all inputs are deterministic and consensus-safe.

Inputs:

- Uptime history
- Missed blocks
- Slash events
- Commission changes
- Self-delegation changes
- Validator metadata changes
- Delegation inflow and outflow
- Historical reward performance

Outputs:

- Validator risk score
- Reliability score
- Commission stability score
- Concentration warning
- Capture-risk warning
- Delegator-facing metadata

Failure modes:

- Advisory score is mistaken for a guarantee.
- Poor scoring weights create misleading yield estimates.
- Non-deterministic inputs are accidentally used in consensus-critical logic.
- Operators optimize for score metrics rather than real reliability.

Integration points:

- Validator query APIs
- Wallet and explorer interfaces
- Delegation simulator
- Governance dashboards
- Staking enhancements module

Required tasks:

- Separate consensus-safe and advisory-only score components.
- Define scoring versioning.
- Add score explanation output.
- Add tests for score stability under missing data.

### 5.3 Adaptive Inflation Module

Purpose:

- Adjust gross issuance based on bonded stake, network activity, fee revenue, and security budget requirements.
- Keep inflation within governance-defined bounds.

Inputs:

- Bonded stake ratio
- Target stake ratio
- Fee revenue
- Burn amount
- Validator count
- Validator reward floor
- Network activity score
- Treasury and reserve health
- Inflation bounds
- Per-window adjustment limit

Outputs:

- Inflation rate for next epoch
- Mint amount
- Net issuance report
- Controller state
- Deflation guard status
- Economic accounting events

Failure modes:

- Controller oscillation due to noisy inputs.
- Inflation remains high after security need declines.
- Inflation drops too low during low-fee periods.
- Bad activity score allows manipulation.
- Parameter misconfiguration creates reward instability.

Integration points:

- Minting flow
- Distribution flow
- Fee market optimizer
- Burn controller
- Governance parameters
- Epoch accounting
- Supply query APIs

Required tasks:

- Define controller formula.
- Add smoothing and adjustment bounds.
- Add manipulation-resistant activity score.
- Add epoch-level accounting reconciliation.
- Add stress tests for supply and security budget.

### 5.4 Economic Security Module

Purpose:

- Enforce economic invariants and coordinate security-related penalties, reserves, and circuit breakers.

Inputs:

- Slashing events
- Evidence submissions
- Validator concentration metrics
- Fee market instability signals
- State growth metrics
- Abnormal stake movement metrics
- Security reserve balance
- Governance-defined thresholds

Outputs:

- Security alerts
- Circuit breaker status
- Penalty routing decisions
- Reserve funding requests
- Invariant violation events
- Governance reports

Failure modes:

- False positives trigger unnecessary restrictions.
- Weak thresholds miss real economic attacks.
- Circuit breaker creates unexpected fee or reward behavior.
- Security reserve accounting becomes inconsistent.

Integration points:

- Slashing flow
- Fee distribution
- Burn controller
- Treasury and community pool accounting
- Governance
- Monitoring stack

Required tasks:

- Define economic invariants.
- Add invariant checking at epoch boundaries.
- Add security reserve accounting.
- Add deterministic circuit breaker rules.
- Add incident telemetry and audit logs.

### 5.5 Fee Market Optimizer Module

Purpose:

- Coordinate base fee, congestion multipliers, anti-spam surcharges, and fee distribution.

Inputs:

- Block gas utilization
- Transaction gas usage
- Mempool pressure
- Failed execution rate
- State write volume
- Deployment volume
- Message forwarding volume
- Fee split parameters
- Burn and reserve caps

Outputs:

- Base fee
- Resource multipliers
- Sender-local surcharge
- Fee allocation by bucket
- Fee estimate data
- Congestion events

Failure modes:

- Fee spikes price out normal usage.
- Slow response allows spam to persist too long.
- Mempool fee checks diverge from execution fee checks.
- Fee bucket accounting fails to sum exactly.
- Resource multipliers become manipulable.

Integration points:

- Ante handler fee checks
- Mempool admission
- Block execution
- Distribution flow
- Burn controller
- State rent and storage accounting
- Wallet fee estimation queries

Required tasks:

- Define base fee update formula.
- Define resource multiplier formula.
- Align mempool and execution fee validation.
- Add fee allocation accounting.
- Add congestion simulations.

## 6. Prioritization Matrix

Scoring:

- Impact on security: `1` low, `10` high
- Impact on decentralization: `1` low, `10` high
- Implementation complexity: `1` low, `10` high
- Urgency: `Low`, `Medium`, `High`

| Improvement | Security Impact | Decentralization Impact | Complexity | Urgency |
| --- | ---: | ---: | ---: | --- |
| Wire burn controller into production fee flow | 8 | 4 | 6 | High |
| Add net issuance accounting and supply reports | 8 | 3 | 4 | High |
| Add deflation guard enforcement | 8 | 3 | 5 | High |
| Productionize epoch-based validator selection | 9 | 8 | 8 | High |
| Add validator concentration metrics | 8 | 9 | 4 | High |
| Add reward dampening above concentration thresholds | 8 | 9 | 7 | High |
| Extend slashing fund routing to burn, treasury, and reporters | 9 | 5 | 7 | High |
| Add repeat-offense slashing multipliers | 8 | 5 | 6 | Medium |
| Add validator risk and reputation queries | 6 | 8 | 5 | High |
| Add risk-adjusted delegation yield estimates | 6 | 8 | 5 | Medium |
| Add validator metadata and commission-change warnings | 6 | 7 | 4 | Medium |
| Add dynamic base fee adjustment bounds | 8 | 4 | 5 | High |
| Add congestion simulations and fee controller tests | 8 | 3 | 4 | High |
| Add sender-local anti-spam surcharge | 8 | 3 | 6 | High |
| Add resource-specific fee multipliers | 7 | 3 | 7 | Medium |
| Replace static fee split with bucketed allocation | 7 | 5 | 6 | Medium |
| Add state write and update pricing | 8 | 4 | 6 | High |
| Add storage footprint queries | 6 | 4 | 4 | Medium |
| Design and implement state rent | 9 | 5 | 9 | Medium |
| Add state delete refund policy | 7 | 4 | 6 | Medium |
| Add state growth telemetry and alerts | 7 | 4 | 4 | High |
| Add adaptive inflation controller | 9 | 5 | 8 | Medium |
| Add inflation smoothing and per-window limits | 8 | 4 | 5 | High |
| Add economic invariant tests | 9 | 5 | 5 | High |
| Add economic attack simulations | 9 | 6 | 6 | High |
| Add security reserve accounting | 7 | 4 | 6 | Medium |
| Add fee market circuit breaker | 8 | 3 | 7 | Medium |
| Add delegation simulator | 5 | 7 | 5 | Medium |
| Add validator bootstrap band | 6 | 8 | 7 | Medium |
| Add governance parameter impact reports | 7 | 5 | 6 | Medium |

## 7. Implementation Sequencing

### Phase 0: Measurement and Accounting

Goal:

- Make existing economics observable before changing incentives.

Tasks:

- Add net issuance accounting.
- Add cumulative burn accounting.
- Add fee bucket accounting.
- Add validator concentration metrics.
- Add state growth telemetry.
- Add validator reward per voting-power telemetry.
- Add epoch economic report generation.

Exit criteria:

- Operators and governance can reconcile issuance, burn, fees, rewards, and slashing per epoch.
- Concentration and state growth are queryable.
- Accounting invariants are covered by tests.

### Phase 1: Production Safety

Goal:

- Close incomplete production paths that affect supply, fees, and penalties.

Tasks:

- Wire burn controller into fee distribution.
- Enforce deflation guard.
- Add burn caps.
- Add fee controller bounds.
- Align mempool and execution fee validation.
- Add slashing fund routing.
- Add invariant tests for all fund movement.

Exit criteria:

- No economic fund path is partially disconnected.
- Fee, burn, mint, distribution, and slash accounting reconcile in tests.
- Controller behavior remains inside configured bounds under stress tests.

### Phase 2: Validator and Delegation Incentives

Goal:

- Improve security and decentralization of staking.

Tasks:

- Add validator scoring.
- Productionize epoch-based selection.
- Add concentration reward dampening.
- Add validator risk score queries.
- Add commission-change warnings.
- Add risk-adjusted yield estimates.
- Add validator bootstrap band.

Exit criteria:

- Delegators have queryable risk and yield data.
- Validator concentration incentives are active and bounded.
- Active-set transitions are deterministic and tested.

### Phase 3: Fee, Storage, and Execution Optimization

Goal:

- Price resource usage more accurately.

Tasks:

- Add resource-specific fee multipliers.
- Add sender-local spam surcharge.
- Add storage write and update pricing.
- Add storage footprint queries.
- Add delete refund policy.
- Add deployment and forwarding fee estimation.
- Add state growth surcharge.

Exit criteria:

- Persistent state and high-frequency spam have direct economic costs.
- Fee estimator supports transaction, deployment, and async message flows.
- State growth is bounded by pricing and telemetry.

### Phase 4: Adaptive Controllers and Long-Term Stabilization

Goal:

- Couple issuance, burns, security budget, and activity into a stable economic loop.

Tasks:

- Implement adaptive inflation controller.
- Add supply projection reports.
- Add economic security module.
- Add fee market circuit breaker.
- Add security reserve accounting.
- Add pre-upgrade economic simulation requirement.

Exit criteria:

- Gross issuance, burns, and net issuance are controlled by explicit policy.
- Economic controllers are simulation-tested before activation.
- Governance receives parameter impact reports before economic changes.

## 8. Required Test Coverage

### 8.1 Invariant Tests

- Minted supply equals distributed rewards plus module balances where applicable.
- Burned supply is removed from spendable supply.
- Fee allocation buckets sum exactly to collected fees.
- Slashed funds route exactly according to configured splits.
- Reward adjustment factors remain within configured bounds.
- Inflation remains within configured bounds.
- Deflation guard enforces net issuance floor.
- Storage refunds cannot exceed configured maximum.
- Mempool and execution fee checks cannot diverge.

### 8.2 Simulation Tests

- Low activity with low fee revenue.
- Normal activity near target stake ratio.
- High activity with sustained congestion.
- High burn pressure.
- Low bonded stake below safety threshold.
- Validator concentration above soft cap.
- Stake split across multiple validators.
- Repeated validator downtime.
- Equivocation with reporter reward.
- Fee spam from one account.
- State bloat attack.
- Deployment congestion.
- Rapid commission increase.
- Sudden delegation inflow to one validator.

### 8.3 Upgrade Tests

- Parameter migration preserves existing balances.
- Existing delegations remain valid.
- Existing validators remain queryable.
- Fee distribution changes do not strand module balances.
- Burn controller activation does not break reward distribution.
- State pricing activation has a defined starting state.
- Inflation controller upgrade starts from current inflation parameters.

## 9. Observability Requirements

### 9.1 Metrics

- Current inflation rate.
- Gross minted AET per epoch.
- Burned AET per epoch.
- Net supply change per epoch.
- Total bonded stake ratio.
- Active validator count.
- Standby validator count.
- Top-N voting power share.
- Per-validator reward per voting power.
- Fee revenue by bucket.
- Base fee and congestion multiplier.
- Failed transaction surcharge totals.
- State bytes added, removed, and net changed.
- Storage rent balances and exhaustion warnings.
- Slashing amount by severity.
- Reporter rewards paid.
- Deflation guard activation count.
- Circuit breaker activation count.

### 9.2 Events

- Inflation update event.
- Fee allocation event.
- Burn event.
- Deflation guard event.
- Validator score update event.
- Validator concentration warning event.
- Commission change warning event.
- Delegation risk warning event.
- Slashing route event.
- Reporter reward event.
- Storage fee event.
- State rent warning event.
- Circuit breaker event.

### 9.3 Queries

- Current economic parameters.
- Current and historical inflation state.
- Net issuance by epoch.
- Fee distribution by epoch.
- Burn history.
- Validator score and score components.
- Validator concentration metrics.
- Delegator risk-adjusted yield estimate.
- Storage footprint by account or contract.
- State rent status.
- Fee estimate for transaction class.
- Supply projection under current parameters.

## 10. Governance Parameter Surface

### 10.1 Inflation Parameters

- Minimum inflation.
- Target inflation.
- Maximum inflation.
- Per-window adjustment limit.
- Smoothing window.
- Target stake ratio.
- Validator reward floor.
- Net issuance floor.

### 10.2 Burn Parameters

- Fee burn allocation.
- Slashing burn allocation.
- Burn cap per epoch.
- Burn activation threshold.
- Deflation guard threshold.

### 10.3 Fee Parameters

- Minimum base fee.
- Maximum base fee.
- Target block utilization.
- Maximum fee adjustment per window.
- Congestion multiplier bounds.
- Sender-local surcharge parameters.
- Resource-specific multipliers.
- Fee allocation bucket weights.

### 10.4 Validator Parameters

- Minimum self-delegation.
- Maximum validator commission.
- Maximum commission change per interval.
- Active validator target.
- Active validator maximum.
- Epoch length.
- Validator score weights.
- Concentration soft cap.
- Reward dampening curve.
- Bootstrap eligibility parameters.

### 10.5 Storage Parameters

- State write fee per byte.
- State update fee per byte.
- Delete refund cap.
- Delete refund decay.
- Rent rate.
- Rent grace period.
- State growth surcharge threshold.
- State maintenance reserve allocation.

### 10.6 Security Parameters

- Slashing severity rates.
- Repeat-offense multiplier.
- Repeat-offense decay.
- Reporter reward allocation.
- Reporter reward cap.
- Security reserve allocation.
- Circuit breaker thresholds.

## 11. Open Design Decisions

- Whether validator reward dampening should affect only validator commission, total validator-delegator rewards, or both.
- Whether state rent should apply immediately to all state or only to state created after activation.
- Whether storage delete refunds should be paid immediately or credited against future storage fees.
- Whether the validator bootstrap band should be funded by inflation, fee allocation, or reward redistribution.
- Whether risk-adjusted yield estimates should be chain-native queries or maintained by indexers using chain-native data.
- Whether fee bucket weights should be static governance parameters or controller-adjusted within bounds.
- Whether security reserve spending should require governance approval, automatic trigger conditions, or both.

## 12. Near-Term Engineering Backlog

High priority:

- Add `/ECONOMICS.md` to local git exclude.
- Implement epoch economic report data model.
- Add net issuance accounting.
- Add burn accounting and queries.
- Wire burn allocation into fee distribution.
- Enforce deflation guard.
- Add fee allocation invariant tests.
- Add slashing route invariant tests.
- Add validator concentration queries.
- Add state growth telemetry.

Medium priority:

- Implement validator risk score query.
- Implement commission-change warning event.
- Add dynamic base fee simulation tests.
- Add sender-local spam surcharge design.
- Add storage pricing specification for first write, update, delete, and refund.
- Add supply projection command or query.
- Add governance parameter impact report.

Lower priority:

- Implement full state rent lifecycle.
- Implement adaptive inflation controller.
- Implement security reserve module.
- Implement fee market circuit breaker.
- Implement delegation simulator.
- Implement validator bootstrap band.

## 13. Non-Goals

- Do not introduce a second staking asset.
- Do not make external assets part of validator rewards.
- Do not make fee accounting depend on off-chain data.
- Do not use non-deterministic reputation inputs in consensus-critical calculations.
- Do not rely on unverifiable validator identity assumptions.
- Do not reduce slashing determinism for discretionary penalty handling.
- Do not make burn priority higher than validator security budget.
- Do not allow economic controllers to operate without bounds, telemetry, and tests.
