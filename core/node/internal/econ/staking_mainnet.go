package econ

// staking_mainnet.go
// ---------------------------------------------------------
// High-level PoS + PoP staking scaffolding for the mainnet
// monetary engine. This file does not perform any state
// transitions on its own; instead it defines the core data
// structures and pure helpers the chain / epoch runner can
// call into once we wire staking and work accounting into
// the mainnet state.
//
// Design summary
// --------------
// - Consensus starts as PoW and later transitions into a
//   PoS+PoP era once supply, node count, and coverage are
//   sufficient.
// - RSX is the staking token: validators and delegators
//   lock RSX to participate in consensus and earn a share
//   of the protocol revenue bucket allocated to stake.
// - PoP (Proof of Participation) measures real work done
//   by nodes: consensus, networking, storage, and service
//   duties. A separate portion of the epoch reward pool is
//   distributed according to PoP work scores so operators
//   with more capacity and actual usage earn more.
// - Rewards are paid in USDR / GRC; RSX has no inflation.
//   This keeps RSX as a fixed‑supply security / governance
//   asset backed by real protocol cashflows.

// ConsensusMode describes the high‑level consensus regime
// the chain is currently in. The chain starts in PoW and
// later transitions to PoS+PoP.
type ConsensusMode string

const (
    ConsensusModePoW     ConsensusMode = "pow"
    ConsensusModePoS_PoP ConsensusMode = "pos_pop"
)

// NodeCapabilityProfile describes the hardware / capacity
// envelope of a node. This is used as an upper bound on
// how much work a node can credibly claim to do per epoch.
type NodeCapabilityProfile struct {
    NodeID         string  `json:"node_id"`
    Role           string  `json:"role"`             // validator, web, storage, archive, oracle, etc.
    CPUScore       float64 `json:"cpu_score"`        // normalized 0..1
    RAMScore       float64 `json:"ram_score"`        // normalized 0..1
    StorageScore   float64 `json:"storage_score"`    // normalized 0..1
    BandwidthScore float64 `json:"bandwidth_score"`  // normalized 0..1
}

// NodeParticipationMetrics captures per‑epoch work signals
// for a node: how much it actually did in that epoch.
type NodeParticipationMetrics struct {
    Epoch           int64   `json:"epoch"`
    NodeID          string  `json:"node_id"`
    UptimeScore     float64 `json:"uptime_score"`      // 0..1
    RequestsServed  float64 `json:"requests_served"`   // weighted RPC / web requests
    BlocksRelayed   float64 `json:"blocks_relayed"`    // p2p relay contribution
    StorageIO       float64 `json:"storage_io"`        // GB served / proofs served
    LatencyScore    float64 `json:"latency_score"`     // 0..1, higher = better latency
}

// NodeCostProfile describes the operator‑declared monthly
// cost band for a node. We do not trust this blindly for
// protocol safety, but we can use it in analytics and as
// an input when tuning how much of the epoch reward pool
// is allocated to PoP vs stake (to avoid underpaying good
// operators relative to their hosting costs).
type NodeCostProfile struct {
    NodeID         string  `json:"node_id"`
    MonthlyCostUSD float64 `json:"monthly_cost_usd"`
    Tier           string  `json:"tier"` // e.g. "small", "medium", "large"
}

// NodeWorkScore is the final PoP work score for a node in
// a single epoch after combining capability and activity.
type NodeWorkScore struct {
    Epoch      int64   `json:"epoch"`
    NodeID     string  `json:"node_id"`
    WorkScore  float64 `json:"work_score"`
}

// EpochRewardSplit encodes how a single epoch reward pool
// is split between stake (PoS) and work (PoP).
type EpochRewardSplit struct {
    TotalRewardUSD   float64 `json:"total_reward_usd"`
    StakePortionUSD  float64 `json:"stake_portion_usd"`
    PoPPortionUSD    float64 `json:"pop_portion_usd"`
    StakeShare       float64 `json:"stake_share"` // alpha in [0,1]
    PoPShare         float64 `json:"pop_share"`   // 1-alpha
}

// ComputeEpochRewardSplit splits a total epoch reward pool
// (denominated in USD terms for modeling) into a stake and
// PoP portion according to the given alpha parameter. The
// caller is responsible for later converting these amounts
// into specific currency units (USDR / GRC) when applying
// them to ledger balances.
func ComputeEpochRewardSplit(totalRewardUSD float64, alpha float64) EpochRewardSplit {
    if alpha < 0 {
        alpha = 0
    }
    if alpha > 1 {
        alpha = 1
    }
    stake := totalRewardUSD * alpha
    pop := totalRewardUSD - stake
    return EpochRewardSplit{
        TotalRewardUSD:  totalRewardUSD,
        StakePortionUSD: stake,
        PoPPortionUSD:   pop,
        StakeShare:      alpha,
        PoPShare:        1 - alpha,
    }
}

// NodePoPSnapshot is a compact view of a node's work score
// at the end of an epoch, suitable for computing PoP‑based
// reward allocations.
type NodePoPSnapshot struct {
    NodeID    string  `json:"node_id"`
    WorkScore float64 `json:"work_score"`
}

// NodePoPReward describes the portion of the PoP reward
// pool allocated to a single node for an epoch.
type NodePoPReward struct {
    NodeID        string  `json:"node_id"`
    WorkScore     float64 `json:"work_score"`
    RewardUSD     float64 `json:"reward_usd"`
    ShareOfPool   float64 `json:"share_of_pool"`
}

// ComputePoPRewards distributes a PoP reward budget across
// nodes according to their work scores. Nodes with higher
// work scores receive a proportionally larger share of the
// pool. Nodes with zero or negative scores receive nothing.
func ComputePoPRewards(popBudgetUSD float64, snapshots []NodePoPSnapshot) []NodePoPReward {
    if popBudgetUSD <= 0 || len(snapshots) == 0 {
        return nil
    }
    var total float64
    for _, s := range snapshots {
        if s.WorkScore > 0 {
            total += s.WorkScore
        }
    }
    if total <= 0 {
        // No positive work scores recorded; nothing to pay out.
        return nil
    }
    rewards := make([]NodePoPReward, 0, len(snapshots))
    for _, s := range snapshots {
        if s.WorkScore <= 0 {
            continue
        }
        share := s.WorkScore / total
        if share < 0 {
            share = 0
        }
        amount := share * popBudgetUSD
        rewards = append(rewards, NodePoPReward{
            NodeID:      s.NodeID,
            WorkScore:   s.WorkScore,
            RewardUSD:   amount,
            ShareOfPool: share,
        })
    }
    return rewards
}
