# x/identity Migration Marker

Status: legacy identity specification.

This package is kept as specification and compatibility surface for tests. Domain
NFT, resolver, subdomain manager, domain auction, DAO-controlled domain logic,
and NFT binding customization are contract-target logic and must migrate to AVM
contracts.

Root-only native logic belongs in `x/identity-root`: `.aet` root coordination,
name uniqueness, reserved names, deterministic normalization, expiry bounds, and
root registry policy.
