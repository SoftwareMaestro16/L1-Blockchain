# Final Target Formula

```text
Aetra =
  CometBFT BFT PoS
  + Cosmos SDK
  + AVM-only genesis smart contracts
  + 100-300 active validators over time
  + 5-8 second block time
  + <= 120 second worst acceptable finality target
  + strict objective slashing
  + validator effective power cap
  + anti-concentration rewards
  + dynamic low/moderate inflation
  + fee burn
  + protocol treasury
  + mandatory tests for every feature
```

The most important product decision: Aetra should be a chain people can trust, not a chain optimized only for speed or short-term APR.

## Implementation Contract

The implementation gate is `app/params/final_target_formula.go`.

Required properties:

- consensus target is CometBFT BFT PoS;
- application base is Cosmos SDK;
- smart contract VM target is AVM-only at genesis;
- active validator range is 100-300 over time;
- block time target is 5-8 seconds;
- worst acceptable finality target is less than or equal to 120 seconds;
- slashing is strict and objective;
- validator effective power cap is required;
- anti-concentration rewards are required;
- inflation is dynamic and low/moderate;
- fee burn and protocol treasury are required;
- every feature must have tests;
- product direction prioritizes trust over speed-only or short-term APR positioning.
