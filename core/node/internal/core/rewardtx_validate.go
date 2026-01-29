package core

import (
    "errors"
    "fmt"
)

// BasicRewardTxValidation performs a set of stateless and semi-stateful
// checks over a RewardTx. This helper is intended to be used by the
// consensus / block validation layer as a first line of defence before
// any more sophisticated economic-leader / signature rules are applied.
//
// The function expects:
//   - expectedEpoch:  epoch index this block is supposed to be paying.
//   - expectedBudget: total epoch reward budget from the issuance curve.
//
// It enforces:
//   * epoch index matches expectedEpoch
//   * TotalRewardGRC matches expectedBudget within a small epsilon
//   * all entry amounts are non-negative
//   * no duplicate OperatorID entries
//   * sum(entry amounts) does not exceed TotalRewardGRC (within epsilon)
func BasicRewardTxValidation(tx RewardTx, expectedEpoch uint64, expectedBudget float64) error {
    if tx.EpochIndex != expectedEpoch {
        return fmt.Errorf("rewardtx: epoch mismatch: have %d want %d", tx.EpochIndex, expectedEpoch)
    }

    // allow a tiny epsilon to account for floating point rounding
    const eps = 1e-6
    if tx.TotalRewardGRC < 0 {
        return errors.New("rewardtx: negative TotalRewardGRC")
    }
    if absFloat(tx.TotalRewardGRC-expectedBudget) > eps {
        return fmt.Errorf("rewardtx: budget mismatch: have %.8f want %.8f", tx.TotalRewardGRC, expectedBudget)
    }

    if len(tx.Entries) == 0 {
        return errors.New("rewardtx: no entries")
    }

    seen := make(map[string]struct{}, len(tx.Entries))
    var sum float64

    for _, e := range tx.Entries {
        if e.OperatorID == "" {
            return errors.New("rewardtx: empty operator id")
        }
        if _, ok := seen[e.OperatorID]; ok {
            return fmt.Errorf("rewardtx: duplicate operator id %q", e.OperatorID)
        }
        seen[e.OperatorID] = struct{}{}

        if e.AmountGRC < 0 {
            return fmt.Errorf("rewardtx: negative amount for operator %q", e.OperatorID)
        }
        sum += e.AmountGRC
    }

    if sum < 0 {
        return errors.New("rewardtx: negative total entry sum")
    }
    if sum-tx.TotalRewardGRC > eps {
        return fmt.Errorf("rewardtx: entries exceed budget: entries=%.8f budget=%.8f", sum, tx.TotalRewardGRC)
    }

    return nil
}

func absFloat(v float64) float64 {
    if v < 0 {
        return -v
    }
    return v
}
