package econ

import (
    "reservechain/internal/core"
)

// RewardEngineConfig ties together the issuance curve and the node work
// scoring weights. This layer does not talk to the ledger directly; it
// just decides how much GRC is available for rewards in a given epoch
// and how that pool is split across nodes.
//
// The actual minting / crediting is handled by the chain / state
// transition layer, which can call into this package as a pure helper.
type RewardEngineConfig struct {
    Issuance IssuanceParams
    Weights  core.WorkWeights
}

// DefaultRewardEngineConfig returns a conservative, devnet-safe set of
// parameters for splitting work-based rewards. The WorkWeights here are
// intentionally simple; they can be overridden from config or governance
// once the mainnet economics are locked in.
func DefaultRewardEngineConfig() RewardEngineConfig {
    return RewardEngineConfig{
        Issuance: DefaultIssuanceParams(),
        Weights: core.WorkWeights{
            Consensus: 0.70,
            Network:   0.30,
            // Storage / Service components will be wired in later once
            // those metrics are actually collected on-chain.
        },
    }
}

// EpochOperatorPayouts computes the work-based reward split for a given
// epoch:
//
//   * epochIndex: monotonic epoch counter (0-based).
//   * snapshots:  per-node work snapshots for this epoch.
//   * cfg:        reward engine parameters (issuance + weights).
//
// It returns:
//
//   * NodeWorkEpochResult from core.ComputeOperatorPayouts, containing
//     the per-node RewardGRC allocations;
//   * operatorBudget: the total GRC allocated to operators for this
//     epoch (before any rounding in ComputeOperatorPayouts);
//   * treasuryBudget: the total GRC allocated to the treasury / reserve
//     bucket for this epoch.
//
// This function is deliberately pure: it does not mutate chain state or
// perform any minting. The caller is expected to:
//
//   1. Use the returned NodeWorkEpochResult to construct a RewardTx
//      (economic leader path), or
//   2. In DevNet, apply the rewards directly to an in-memory store.
func EpochOperatorPayouts(epochIndex uint64, snapshots []core.NodeWorkSnapshot, cfg RewardEngineConfig) (core.NodeWorkEpochResult, float64, float64) {
    // 1) Compute the epoch reward budget from the issuance curve.
    _, opBudget, treasuryBudget := EpochRewardBudget(epochIndex, cfg.Issuance)

    // 2) Split the operator budget across nodes according to their
    // work scores using the core-level helper.
    result := core.ComputeOperatorPayouts(cfg.Weights, snapshots, opBudget)

    return result, opBudget, treasuryBudget
}
