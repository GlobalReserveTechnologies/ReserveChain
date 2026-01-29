package econ

import "math"

// IssuanceParams controls the shape of the epoch reward curve.
// This is a first, devnet-friendly implementation of the monetary
// issuance function R(t) = R0 / (1 + k*t)^alpha with a simple split
// between operator and treasury buckets.
//
// All values are expressed in whole GRC units. On mainnet this will
// likely be expressed in the smallest atomic unit instead.
type IssuanceParams struct {
    // R0 is the initial per-epoch reward budget.
    R0 float64

    // K controls the decay speed; higher K -> faster decay.
    K float64

    // Alpha controls the curve shape; Alpha > 1 yields a finite
    // long-term total monetary base.
    Alpha float64

    // OperatorShare is the fraction of the epoch reward that is
    // allocated to node operators (validators / service nodes).
    OperatorShare float64

    // TreasuryShare is the fraction of the epoch reward that is
    // allocated to the treasury / reserve bucket.
    TreasuryShare float64
}

// DefaultIssuanceParams returns a conservative, devnet-safe set of
// parameters. These are deliberately hard-coded for now; later we
// will map them from config and, eventually, governance.
func DefaultIssuanceParams() IssuanceParams {
    return IssuanceParams{
        R0:            1000.0,  // 1000 GRC in the first epoch
        K:             0.001,   // mild decay over time
        Alpha:         1.5,     // >1 => finite long-term supply
        OperatorShare: 0.80,    // 80%% of issuance to operators
        TreasuryShare: 0.20,    // 20%% to treasury / reserve
    }
}

// EpochRewardBudget computes the total reward budget for a given
// epoch index t (0-based). It returns:
//
//   total    - total epoch reward R(t)
//   operator - portion allocated to operators
//   treasury - portion allocated to treasury / reserve
//
// For now this is entirely stateless and deterministic; the epoch
// index is expected to be derived from chain height / time in the
// consensus layer.
func EpochRewardBudget(t uint64, p IssuanceParams) (total, operator, treasury float64) {
    // Basic power-law style decay:
    base := p.R0 / math.Pow(1.0+p.K*float64(t), p.Alpha)
    if base < 0 {
        base = 0
    }

    // Simple split into operator / treasury buckets.
    op := base * p.OperatorShare
    tr := base * p.TreasuryShare

    // Guard against misconfigured shares that don't sum to 1.0 by
    // snapping total to the actual sum of the buckets.
    total = op + tr
    if total < 0 {
        total = 0
    }
    return total, op, tr
}
