package econ

import (
    "log"
    "time"

    "reservechain/internal/core"
)

// DevnetRewardLoopConfig controls the behaviour of the standalone
// DevNet reward loop. This is intentionally simple and is NOT meant to
// represent the final mainnet economics pipeline; it is a harness for
// exercising the issuance curve + work-scoring + RewardTx plumbing on a
// single node.
type DevnetRewardLoopConfig struct {
    // EpochSeconds controls how often a reward epoch is executed.
    EpochSeconds int64

    // TreasuryAddr is the payout address for the treasury / reserve
    // bucket. If empty, the treasury share for each epoch is skipped.
    TreasuryAddr string
}

// DefaultDevnetRewardLoopConfig returns a conservative configuration
// suitable for local testing.
func DefaultDevnetRewardLoopConfig() DevnetRewardLoopConfig {
    return DevnetRewardLoopConfig{
        EpochSeconds: 60,        // one reward epoch per minute
        TreasuryAddr: "treasury", // simple well-known bucket
    }
}

// RunDevnetRewardLoop starts a background goroutine that periodically:
//
//   1. Builds a single-node NodeWorkSnapshot for this node.
//   2. Computes the operator + treasury reward budgets for the current epoch.
//   3. Splits the operator budget across nodes (in DevNet: just this node).
//   4. Assembles a RewardTx.
//   5. Applies it to the chain via Chain.ApplyRewardTx.
//
// All errors are logged; the loop continues unless the stopCh is closed.
//
// This helper is deliberately opinionated and is expected to be used
// only for DevNet-style experiments. The mainnet design will instead
// have the consensus / economic-leader logic call the lower-level
// EpochOperatorPayouts + BuildRewardTx + ApplyRewardTx helpers directly.
func RunDevnetRewardLoop(chain *core.Chain, nodeID string, stopCh <-chan struct{}) {
    cfg := DefaultDevnetRewardLoopConfig()
    rewardCfg := DefaultRewardEngineConfig()

    if cfg.EpochSeconds <= 0 {
        cfg.EpochSeconds = 60
    }

    ticker := time.NewTicker(time.Duration(cfg.EpochSeconds) * time.Second)
    defer ticker.Stop()

    log.Printf("[devnet-rewards] starting reward loop: epoch=%ds treasury=%q", cfg.EpochSeconds, cfg.TreasuryAddr)

    var epochIndex uint64
    var lastEpochEnd int64 = time.Now().Unix()

    for {
        select {
        case <-stopCh:
            log.Printf("[devnet-rewards] stopping reward loop")
            return
        case now := <-ticker.C:
            epochIndex++

            epochStart := lastEpochEnd
            epochEnd := now.Unix()
            lastEpochEnd = epochEnd

            // For DevNet we fabricate a simple, single-node snapshot
            // that treats this node as a healthy validator.
            snapshot := core.NodeWorkSnapshot{
                NodeID:      nodeID,
                EpochStart:  epochStart,
                EpochEnd:    epochEnd,
                Consensus:   1.0,
                Network:     1.0,
                Storage:     0.0,
                Service:     0.0,
                HardwareCap: 1.0,
                IsValidator: true,
                IsSeed:      false,
            }

            // Compute work-based payouts for this epoch and derive the
            // operator + treasury budgets.
            workEpoch, opBudget, treasuryBudget := EpochOperatorPayouts(epochIndex, []core.NodeWorkSnapshot{snapshot}, rewardCfg)
            totalBudget := opBudget + treasuryBudget

            // Build a RewardTx with an empty work-root commitment for now.
            var zeroRoot [32]byte
            rewardTx := BuildRewardTx(
                epochIndex,
                epochStart,
                epochEnd,
                nodeID,
                workEpoch,
                totalBudget,
                zeroRoot,
            )

            // Apply the RewardTx to the chain. We pass nil for the
            // operatorToAccount map so that OperatorID doubles as the
            // payout address in DevNet.
            blk, hash, err := chain.ApplyRewardTx(
                rewardTx,
                epochIndex,
                totalBudget,
                nil,
                cfg.TreasuryAddr,
                treasuryBudget,
            )
            if err != nil {
                log.Printf("[devnet-rewards] ApplyRewardTx failed: %v", err)
                continue
            }

            log.Printf("[devnet-rewards] epoch=%d total=%.4f op=%.4f treasury=%.4f block=%s logIndex=%d",
                epochIndex, totalBudget, opBudget, treasuryBudget, hash, blk.Index)
        }
    }
}
