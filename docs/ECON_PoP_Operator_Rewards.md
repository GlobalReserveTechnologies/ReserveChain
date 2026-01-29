# PoP Operator Rewards — Design Summary

This document captures the high‑level design for Proof of Participation
(PoP) operator rewards on ReserveChain mainnet. It is the bridge between
the abstract economic model and the concrete Go helpers in
`internal/econ/staking_mainnet.go` and `internal/core/work.go`.

## Goals

- Reward node operators **based on real work**, not equal per‑node splits.
- Ensure properly provisioned nodes can **earn more than their hosting
  costs** at equilibrium, given protocol fees and demand.
- Keep RSX as a **fixed‑supply staking / security token** with no block
  inflation; rewards are paid in USDR / GRC from protocol cashflows.
- Cleanly separate:
  - PoW bootstrap era
  - PoS+PoP steady‑state era

## Work Model

Each node has:

- A **capability profile** (`NodeCapabilityProfile`):
  - CPUScore, RAMScore, StorageScore, BandwidthScore, Role.
- **Participation metrics** (`NodeParticipationMetrics`):
  - UptimeScore, RequestsServed, BlocksRelayed, StorageIO, LatencyScore.
- A **cost profile** (`NodeCostProfile`):
  - Operator‑declared MonthlyCostUSD and Tier.

From these inputs and the weights in `internal/core/work.go`, the core
layer computes a **work score** per epoch for each node. The econ layer
then wraps this into `NodeWorkScore` / `NodePoPSnapshot` and uses it to
split the PoP reward pool with `ComputePoPRewards`.

## Epoch Reward Split

For each epoch, the global reward pool (denominated in USD terms for
modeling) is split as:

- Stake portion (PoS): `R_stake = alpha * R_total`
- Work portion (PoP): `R_pop   = (1-alpha) * R_total`

where `alpha` is typically in the 0.5–0.7 range and may be adjusted over
time based on observed node cost coverage.

`ComputeEpochRewardSplit` in `staking_mainnet.go` encapsulates this
split and returns an `EpochRewardSplit` struct the epoch runner can use
when applying rewards to ledger balances.

## Operator Payout Currency (USDR + GRC mix)

Operator rewards are paid in a **mixture of USDR and GRC**:

- A **stable portion in USDR** so operators can reliably cover fiat‑denominated
  infrastructure costs (servers, bandwidth, storage).
- A **volatile portion in GRC** so operators share in the upside of the protocol.

The exact mix is a policy parameter. In the current engine wiring we start
with a 70/30 split **by USD value**:

- 70% of the PoP reward pool is earmarked as a stable USDR payout.
- 30% of the PoP reward pool is earmarked as a volatile GRC payout.

These appear in the epoch history as `pop_stable_portion_usd` and
`pop_vol_portion_usd` fields, plus their corresponding shares
`pop_stable_share` and `pop_vol_share`. The future staking / payout logic
will convert these USD amounts into concrete token flows once the RSX and
per‑node PoP staking modules are fully wired.


## Node Cost Coverage Target

For each node, the economics layer can estimate cost coverage as:

- `EstimatedMonthlyReward_i ≈ Reward_pop_i_per_epoch * epochs_per_month`
- `CostCoverage_i = EstimatedMonthlyReward_i / MonthlyCostUSD_i`

The policy target is to keep **healthy** nodes (good uptime and work
scores) around `CostCoverage >= 1.2` at equilibrium. If the system is
systematically under‑paying nodes, future policy revisions can:

- Increase the share of `R_total` allocated to PoP vs stake (adjust
  `alpha`), or
- Adjust protocol fees / spreads to grow `R_total`.

## PoW → PoS+PoP

The chain begins life in **PoW** mode and transitions to **PoS+PoP**
once supply, node count, and coverage are sufficient. The transition is
governed by `ConsensusMode`:

- `ConsensusModePoW`   — blocks come from mining; PoP may be tracked but
  is not yet the dominant reward path.
- `ConsensusModePoS_PoP` — blocks come from RSX staking (PoS) and epoch
  rewards are split between stakers and PoP operators as per the split
  above.

This document is intentionally high‑level; the precise parameter values
(alpha, weights, thresholds) are stored in configuration and can be
tuned without changing the core algorithm.
