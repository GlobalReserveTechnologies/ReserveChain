package core

// WorkWeights controls how the node work score is computed from the
// four major contribution components:
//
//	A = consensus work   (mining / validation / signatures)
//	B = network work     (p2p bytes, latency, uptime)
//	C = storage work     (history depth, archival serving)
//	D = service work     (RPC, explorer, auditor, workstation)
//
// W_raw = wA*A + wB*B + wC*C + wD*D
// W_final is then clamped by the hardware capacity score.
type WorkWeights struct {
	Consensus float64 `yaml:"consensus"`
	Network   float64 `yaml:"network"`
	Storage   float64 `yaml:"storage"`
	Service   float64 `yaml:"service"`
}

// NodeWorkSnapshot is a per-epoch summary of a node's contribution.
type NodeWorkSnapshot struct {
	NodeID      string  // logical node id
	EpochStart  int64   // unix seconds
	EpochEnd    int64   // unix seconds
	Consensus   float64 // normalized consensus work (A)
	Network     float64 // normalized network work (B)
	Storage     float64 // normalized storage work (C)
	Service     float64 // normalized service work (D)
	HardwareCap float64 // capacity ceiling in [0,1]
	IsValidator bool    // whether node is acting as a validator in this epoch
	IsSeed      bool    // whether node is a seed/bootstrapping node
	WorkRaw     float64 // W_raw = sum(weights * components)
	WorkFinal   float64 // W_final = min(W_raw, HardwareCap)
	RewardGRC   float64 // computed GRC payout for this epoch
}

// NodeWorkEpochResult is the outcome of a work-based reward calculation
// across all operators for a single epoch.
type NodeWorkEpochResult struct {
	EpochStart int64
	EpochEnd   int64
	TotalWork  float64
	Nodes      []NodeWorkSnapshot
}

// ComputeOperatorPayouts computes the per-node work scores and GRC payouts
// for a single epoch, given the normalized work components per node and
// a variable reward pool. Floor payouts (USD->GRC) are applied outside
// of this function; here we focus on splitting the variable pool by work.
func ComputeOperatorPayouts(weights WorkWeights, snapshots []NodeWorkSnapshot, variablePoolGRC float64) NodeWorkEpochResult {
	// First pass: compute raw scores and apply hardware caps.
	var total float64
	for i := range snapshots {
		raw := weights.Consensus*snapshots[i].Consensus +
			weights.Network*snapshots[i].Network +
			weights.Storage*snapshots[i].Storage +
			weights.Service*snapshots[i].Service
		snapshots[i].WorkRaw = raw
		cap := snapshots[i].HardwareCap
		if cap <= 0 {
			cap = 1.0
		}
		if raw > cap {
			snapshots[i].WorkFinal = cap
		} else if raw < 0 {
			snapshots[i].WorkFinal = 0
		} else {
			snapshots[i].WorkFinal = raw
		}
		total += snapshots[i].WorkFinal
	}

	// Second pass: split the variable reward pool by work share.
	if total <= 0 || variablePoolGRC <= 0 {
		return NodeWorkEpochResult{
			TotalWork: total,
			Nodes:     snapshots,
		}
	}

	for i := range snapshots {
		share := snapshots[i].WorkFinal / total
		if share < 0 {
			share = 0
		}
		snapshots[i].RewardGRC = share * variablePoolGRC
	}

	return NodeWorkEpochResult{
		TotalWork: total,
		Nodes:     snapshots,
	}
}
