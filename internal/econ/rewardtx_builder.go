package econ

import (
    "time"

    "reservechain/internal/core"
)

// BuildRewardTx constructs a core.RewardTx from an epoch work result and
// the issuance budget for that epoch. This is intended to be called by
// the economic leader once per epoch.
//
// Parameters:
//   - epochIndex:        the 0-based epoch index.
//   - epochStartUnixSec: unix seconds for the start of the epoch window.
//   - epochEndUnixSec:   unix seconds for the end of the epoch window.
//   - leaderValidatorID: validator ID of the node assembling this RewardTx.
//   - workEpoch:         NodeWorkEpochResult containing per-node RewardGRC.
//   - totalBudgetGRC:    total reward budget for the epoch (operator +
//                         treasury) as returned by EpochRewardBudget.
//   - workRoot:          optional commitment to the underlying work metrics.
//
// The returned RewardTx does not include any signature; that is the job
// of the consensus / validator layer.
func BuildRewardTx(
    epochIndex uint64,
    epochStartUnixSec int64,
    epochEndUnixSec int64,
    leaderValidatorID string,
    workEpoch core.NodeWorkEpochResult,
    totalBudgetGRC float64,
    workRoot [32]byte,
) core.RewardTx {
    entries := make([]core.RewardEntry, 0, len(workEpoch.Nodes))

    for _, n := range workEpoch.Nodes {
        if n.RewardGRC <= 0 {
            continue
        }
        entries = append(entries, core.RewardEntry{
            OperatorID: n.NodeID,
            AmountGRC:  n.RewardGRC,
        })
    }

    // Clamp times if the caller passes zero; this keeps the struct
    // usable for observability even with minimal inputs.
    now := time.Now().Unix()
    if epochStartUnixSec == 0 {
        epochStartUnixSec = now
    }
    if epochEndUnixSec == 0 {
        epochEndUnixSec = epochStartUnixSec
    }

    return core.RewardTx{
        EpochIndex:        epochIndex,
        EpochStartUnix:    epochStartUnixSec,
        EpochEndUnix:      epochEndUnixSec,
        LeaderValidatorID: leaderValidatorID,
        TotalRewardGRC:    totalBudgetGRC,
        WorkRoot:          workRoot,
        Entries:           entries,
    }
}
