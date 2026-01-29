package econ

import (
    "math"
)

// RewardEconomicsConfig is a workstation-facing parameter bundle for
// what-if simulation. This is NOT yet the final on-chain economics;
// it is a bridge layer so the Reserve System panel can be wired now.
type RewardEconomicsConfig struct {
    // StakeVsPoPAlpha splits the operator reward pool:
    //   stakeBudget = alpha * operatorBudget
    //   popBudget   = (1-alpha) * operatorBudget
    StakeVsPoPAlpha float64 `json:"stake_vs_pop_alpha"`

    CorridorFloor   float64 `json:"corridor_floor"`
    CorridorTarget  float64 `json:"corridor_target"`
    CorridorCeiling float64 `json:"corridor_ceiling"`

    // TreasurySmoothing controls how aggressively the policy shifts
    // rewards to the treasury when coverage is below target.
    TreasurySmoothing float64 `json:"treasury_smoothing"`

    // IssuanceHalfLife is a soft knob that maps to IssuanceParams.K.
    // Higher half-life => slower decay.
    IssuanceHalfLife float64 `json:"issuance_half_life"`

    // PopShare is a convenience mirror of (1-alpha) for UI.
    PopShare float64 `json:"pop_share"`

    Issuance IssuanceParams `json:"issuance"`
}

func DefaultRewardEconomicsConfig() RewardEconomicsConfig {
    iss := DefaultIssuanceParams()
    return RewardEconomicsConfig{
        StakeVsPoPAlpha: 0.55,
        CorridorFloor:   0.90,
        CorridorTarget:  1.05,
        CorridorCeiling: 1.25,
        TreasurySmoothing: 0.25,
        IssuanceHalfLife: 8000,
        PopShare: 0.45,
        Issuance: iss,
    }
}

type PolicySimSeries struct {
    Epochs   []int64   `json:"epochs"`
    Coverage []float64 `json:"coverage"`
    Equity   []float64 `json:"equity"`
    Reserves []float64 `json:"reserves"`
    Liabs    []float64 `json:"liabs"`
    RewardPoolTotal []float64 `json:"reward_pool_total"`
    RewardRSX []float64 `json:"reward_rsx"`
    RewardPoP []float64 `json:"reward_pop"`
    Issuance  []float64 `json:"issuance"`
    Treasury  []float64 `json:"treasury"`
    CorridorFloor []float64 `json:"corridor_floor"`
    CorridorTarget []float64 `json:"corridor_target"`
    CorridorCeiling []float64 `json:"corridor_ceiling"`
}

// RunPolicySimulation executes a conservative what-if simulation over N epochs.
// It is intentionally simple; its goal is to provide a stable UI contract.
func RunPolicySimulation(startEpoch int64, startReserves, startLiabs float64, n int, cfg RewardEconomicsConfig) PolicySimSeries {
    if n <= 0 {
        n = 500
    }

    // Map half-life to K: half-life ~ ln(2)/K for an exponential;
    // our curve is power-law, so we approximate with a bounded mapping.
    if cfg.IssuanceHalfLife > 0 {
        // heuristic: K in [1e-5, 5e-3]
        k := 0.001 * (8000.0 / cfg.IssuanceHalfLife)
        if k < 1e-5 { k = 1e-5 }
        if k > 5e-3 { k = 5e-3 }
        cfg.Issuance.K = k
    }

    // keep PopShare consistent if caller only sets alpha
    cfg.PopShare = 1.0 - cfg.StakeVsPoPAlpha

    epochs := make([]int64, 0, n)
    coverage := make([]float64, 0, n)
    equity := make([]float64, 0, n)
    reserves := make([]float64, 0, n)
    liabs := make([]float64, 0, n)

    rewardPool := make([]float64, 0, n)
    rewardRSX := make([]float64, 0, n)
    rewardPoP := make([]float64, 0, n)
    issuance := make([]float64, 0, n)
    treasury := make([]float64, 0, n)

    cf := make([]float64, 0, n)
    ct := make([]float64, 0, n)
    cc := make([]float64, 0, n)

    R := startReserves
    L := startLiabs

    // Simplified dynamics:
    // - issuance creates operator + treasury rewards in GRC, modelled as USD-1:1 for UI.
    // - some fraction increases liabilities (USDR circulation grows with activity).
    // - some fraction increases reserves (treasury accumulation).
    // - corridor adjustments tilt treasury share up when below target.
    for i := 0; i < n; i++ {
        e := startEpoch + int64(i)
        // issuance curve based on epoch index (cast to uint64 safely)
        t := uint64(0)
        if e > 0 { t = uint64(e) }
        total, op, tr := EpochRewardBudget(t, cfg.Issuance)

        // coverage before applying this epoch
        cov := 0.0
        if L > 0 {
            cov = R / L
        } else if R > 0 {
            cov = 999.0
        }

        // corridor tilt
        tilt := 0.0
        if cov < cfg.CorridorTarget {
            // below target -> increase treasury share (bounded)
            gap := (cfg.CorridorTarget - cov) / math.Max(cfg.CorridorTarget, 1e-9)
            tilt = clamp(gap*cfg.TreasurySmoothing, 0.0, 0.35)
        } else if cov > cfg.CorridorCeiling {
            // above ceiling -> slight decrease treasury share (release)
            gap := (cov - cfg.CorridorCeiling) / math.Max(cfg.CorridorCeiling, 1e-9)
            tilt = -clamp(gap*cfg.TreasurySmoothing, 0.0, 0.20)
        }

        // apply tilt to shares while keeping total=1
        trAdj := clamp(cfg.Issuance.TreasuryShare+tilt, 0.05, 0.50)
        opAdj := clamp(1.0-trAdj, 0.50, 0.95)

        // re-split total
        op = total * opAdj
        tr = total * trAdj

        // split operator budget between stake and pop
        rsxBudget := op * clamp(cfg.StakeVsPoPAlpha, 0.0, 1.0)
        popBudget := op - rsxBudget

        // Update liabilities and reserves.
        // activityMint increases liabilities; treasury accumulates reserves.
        // We use small heuristics so charts move predictably.
        L += 0.30*op + 0.05*tr          // liabilities grow with usage
        R += 0.85*tr + 0.05*op          // reserves mostly from treasury, small from fees

        // store series
        epochs = append(epochs, e)
        reserves = append(reserves, R)
        liabs = append(liabs, L)

        cov2 := 0.0
        if L > 0 { cov2 = R / L } else if R > 0 { cov2 = 999.0 }
        coverage = append(coverage, cov2)
        equity = append(equity, R-L)

        rewardPool = append(rewardPool, total)
        rewardRSX = append(rewardRSX, rsxBudget)
        rewardPoP = append(rewardPoP, popBudget)
        issuance = append(issuance, total)
        treasury = append(treasury, tr)

        cf = append(cf, cfg.CorridorFloor)
        ct = append(ct, cfg.CorridorTarget)
        cc = append(cc, cfg.CorridorCeiling)
    }

    return PolicySimSeries{
        Epochs: epochs,
        Coverage: coverage,
        Equity: equity,
        Reserves: reserves,
        Liabs: liabs,
        RewardPoolTotal: rewardPool,
        RewardRSX: rewardRSX,
        RewardPoP: rewardPoP,
        Issuance: issuance,
        Treasury: treasury,
        CorridorFloor: cf,
        CorridorTarget: ct,
        CorridorCeiling: cc,
    }
}
