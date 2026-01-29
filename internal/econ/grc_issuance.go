
package econ

// DevNet GRC issuance signal model.
//
// This file does NOT actually mint or burn GRC. Instead it produces a
// per-epoch "signal" that an operator console (or later, on-chain
// policy engine) could use to decide whether to:
//   - issue additional GRC (positive delta), or
//   - constrict supply / buy back (negative delta), or
//   - hold (delta ~ 0).
//
// The signal is based on the current TreasuryBalanceSheet snapshot and
// a very conservative coverage-oriented heuristic.

import (
    "time"
)

// GRCIssuanceSignal is a high-level recommendation derived from the
// current DevNet treasury state. All values are in notional units;
// for DevNet we assume 1 GRC ≈ 1 USD for sizing deltas.
type GRCIssuanceSignal struct {
    Epoch            int64     `json:"epoch"`
    At               time.Time `json:"at"`
    CurrentGRCSupply float64   `json:"current_grc_supply"`
    TargetGRCSupply  float64   `json:"target_grc_supply"`
    RecommendedDelta float64   `json:"recommended_delta"`
    GRCCoverage      float64   `json:"grc_coverage"`
    ReserveAssetsUSD float64   `json:"reserve_assets_usd"`
    EquityUSD        float64   `json:"equity_usd"`
    Mode             string    `json:"mode"` // "issue", "constrict", "hold"
}

// ComputeDevnetGRCIssuanceSignal derives a simple, coverage-driven
// issuance recommendation for GRC based on the latest treasury snapshot.
// This is intentionally conservative and meant for exploration, not as
// a final policy.
func ComputeDevnetGRCIssuanceSignal() GRCIssuanceSignal {
    snap := SnapshotTreasury()

    cov := snap.GRCCoverage
    equity := snap.EquityUSD
    base := snap.GRCSupply

    mode := "hold"
    var delta float64

    // If there is no GRC supply yet or equity is non-positive, there is
    // nothing useful to do: recommend holding.
    if base <= 0 || equity <= 0 || cov == 0 {
        return GRCIssuanceSignal{
            Epoch:            devnetCurrentEpoch,
            At:               time.Now().UTC(),
            CurrentGRCSupply: snap.GRCSupply,
            TargetGRCSupply:  snap.GRCSupply,
            RecommendedDelta: 0,
            GRCCoverage:      snap.GRCCoverage,
            ReserveAssetsUSD: snap.ReserveAssetsUSD,
            EquityUSD:        snap.EquityUSD,
            Mode:             mode,
        }
    }

    // Very simple corridor-style heuristic:
    //
    // - If GRCCoverage is comfortably high (> 0.9), allow a small
    //   positive issuance bounded by both outstanding supply and equity.
    //
    // - If GRCCoverage is very low (< 0.4), recommend a negative delta
    //   (constriction) sized as a fraction of supply, but do not allow
    //   the target to drop below zero.
    //
    // - Otherwise, hold.
    if cov > 0.9 {
        mode = "issue"
        // Issue at most 1% of supply, capped at 2% of equity.
        maxFromSupply := base * 0.01
        maxFromEquity := equity * 0.02
        if maxFromSupply < maxFromEquity {
            delta = maxFromSupply
        } else {
            delta = maxFromEquity
        }
    } else if cov < 0.4 {
        mode = "constrict"
        // Constrict up to 2% of supply or 5% of equity, whichever is smaller.
        maxFromSupply := base * 0.02
        maxFromEquity := equity * 0.05
        var magnitude float64
        if maxFromSupply < maxFromEquity {
            magnitude = maxFromSupply
        } else {
            magnitude = maxFromEquity
        }
        // Negative delta for constriction.
        delta = -magnitude
    }

    target := base + delta
    if target < 0 {
        target = 0
        // Clamp delta so that supply never goes negative.
        delta = -base
        mode = "constrict"
    }

    return GRCIssuanceSignal{
        Epoch:            devnetCurrentEpoch,
        At:               time.Now().UTC(),
        CurrentGRCSupply: snap.GRCSupply,
        TargetGRCSupply:  target,
        RecommendedDelta: delta,
        GRCCoverage:      snap.GRCCoverage,
        ReserveAssetsUSD: snap.ReserveAssetsUSD,
        EquityUSD:        snap.EquityUSD,
        Mode:             mode,
    }
}


// GRCPolicyConfig controls the mainnet GRC issuance model. This is a
// hybrid of corridor, equity, and demand impulses.
type GRCPolicyConfig struct {
    // Maximum fraction of current GRC supply that can be issued in a
    // single epoch when the signal is strongly positive (e.g. 0.02 = 2%).
    MaxIssueFraction float64
    // Maximum fraction of current GRC supply that can be burned in a
    // single epoch when the signal is strongly negative.
    MaxBurnFraction float64

    // Corridor weight: how strongly coverage deviations influence the
    // issuance decision.
    WeightCoverage float64
    // Equity weight: how strongly changes in equity influence issuance.
    WeightEquity float64
    // Demand weight: how strongly net USDR demand influences issuance.
    WeightDemand float64

    // Soft corridor bounds for USDR coverage used in the signal model.
    // Coverage significantly above HighCoverage drives issuance; below
    // LowCoverage drives constriction.
    LowCoverage  float64
    HighCoverage float64
}

// DefaultGRCPolicyConfig returns conservative, hard-coded defaults for
// early mainnet iterations. These can be tuned later or governed.
func DefaultGRCPolicyConfig() GRCPolicyConfig {
    return GRCPolicyConfig{
        MaxIssueFraction: 0.02, // up to 2% of GRC supply per epoch
        MaxBurnFraction:  0.02,

        WeightCoverage: 0.5,
        WeightEquity:   0.3,
        WeightDemand:   0.2,

        LowCoverage:  0.98,
        HighCoverage: 1.02,
    }
}


// ComputeMainnetGRCIssuanceSignal derives a hybrid issuance
// recommendation for GRC based on the mainnet monetary state and
// crypto reserve prices. It combines:
//   - USDR coverage (corridor impulse),
//   - treasury equity (equity impulse),
//   - net USDR demand from pending mints/redemptions (demand impulse).
func ComputeMainnetGRCIssuanceSignal(
    s MainnetMonetaryState,
    prices map[CryptoAssetKind]float64,
    cfg GRCPolicyConfig,
) GRCIssuanceSignal {
    // Compute reserve USD value and infer a simple equity view as:
    //   equity ≈ reserves - USDR liabilities
    var reservesUSD float64
    for _, pool := range s.Reserves {
        for _, b := range pool.Balances {
            var kind CryptoAssetKind
            switch string(b.Asset) {
            case string(AssetUSDC):
                kind = AssetUSDC
            case string(AssetUSDT):
                kind = AssetUSDT
            case string(AssetDAI):
                kind = AssetDAI
            case string(AssetETH):
                kind = AssetETH
            case string(AssetWBTC):
                kind = AssetWBTC
            default:
                kind = ""
            }
            if kind == "" {
                continue
            }
            px, ok := prices[kind]
            if !ok {
                continue
            }
            reservesUSD += b.Amount * px
        }
    }

    usdrSupply := s.Supply.USDR
    grcSupply := s.Supply.GRC

    liabilitiesUSDR := s.Liabilities.USDROutstanding
    if liabilitiesUSDR < 0 {
        liabilitiesUSDR = 0
    }
    equityUSD := reservesUSD - liabilitiesUSDR

    // Coverage impulse: how far are we from the desired band?
    var coverage float64
    if usdrSupply > 0 {
        coverage = reservesUSD / usdrSupply
    } else {
        coverage = 0
    }

    var covImpulse float64
    if coverage > 0 && (coverage > cfg.HighCoverage || coverage < cfg.LowCoverage) {
        mid := (cfg.HighCoverage + cfg.LowCoverage) / 2.0
        span := (cfg.HighCoverage - cfg.LowCoverage) / 2.0
        if span > 0 {
            covImpulse = (coverage - mid) / span
            // Clamp to [-1, 1]
            if covImpulse > 1 {
                covImpulse = 1
            } else if covImpulse < -1 {
                covImpulse = -1
            }
        }
    }

    // Equity impulse: normalized equity vs reserves.
    var eqImpulse float64
    if reservesUSD > 0 {
        ratio := equityUSD / reservesUSD
        if ratio > 1 {
            ratio = 1
        } else if ratio < -1 {
            ratio = -1
        }
        eqImpulse = ratio
    }

    // Demand impulse: net USDR demand from pending mints/redemptions.
    var netUSDRDemand float64
    for _, p := range s.Pending {
        switch p.Kind {
        case PendingMintUSDR:
            netUSDRDemand += p.Amount
        case PendingRedeemUSDR:
            netUSDRDemand -= p.Amount
        }
    }
    var demandImpulse float64
    if reservesUSD > 0 {
        ratio := netUSDRDemand / reservesUSD
        if ratio > 1 {
            ratio = 1
        } else if ratio < -1 {
            ratio = -1
        }
        demandImpulse = ratio
    }

    // Combine the impulses into a single signal in [-1, 1].
    combined := cfg.WeightCoverage*covImpulse +
        cfg.WeightEquity*eqImpulse +
        cfg.WeightDemand*demandImpulse

    if combined > 1 {
        combined = 1
    } else if combined < -1 {
        combined = -1
    }

    mode := "hold"
    var delta float64
    target := grcSupply

    if grcSupply <= 0 || reservesUSD <= 0 {
        // Nothing meaningful to do yet.
        return GRCIssuanceSignal{
            Epoch:            s.Epoch,
            At:               s.At,
            CurrentGRCSupply: grcSupply,
            TargetGRCSupply:  target,
            RecommendedDelta: 0,
            GRCCoverage:      0,
            ReserveAssetsUSD: reservesUSD,
            EquityUSD:        equityUSD,
            Mode:             mode,
        }
    }

    if combined > 0.01 {
        mode = "issue"
        maxDelta := cfg.MaxIssueFraction * grcSupply
        if maxDelta <= 0 {
            maxDelta = grcSupply * 0.01
        }
        delta = combined * maxDelta
        target = grcSupply + delta
    } else if combined < -0.01 {
        mode = "constrict"
        maxDelta := cfg.MaxBurnFraction * grcSupply
        if maxDelta <= 0 {
            maxDelta = grcSupply * 0.01
        }
        delta = combined * maxDelta // negative
        target = grcSupply + delta
        if target < 0 {
            target = 0
        }
    }

    return GRCIssuanceSignal{
        Epoch:            s.Epoch,
        At:               s.At,
        CurrentGRCSupply: grcSupply,
        TargetGRCSupply:  target,
        RecommendedDelta: delta,
        GRCCoverage:      0, // will be refined as a dedicated GRC coverage metric
        ReserveAssetsUSD: reservesUSD,
        EquityUSD:        equityUSD,
        Mode:             mode,
    }
}

// ApplyMainnetGRCPolicy takes a mainnet state and price map, computes
// a hybrid GRC issuance signal, and enqueues the corresponding pending
// issuance or burn entries. It returns the updated state and signal.
func ApplyMainnetGRCPolicy(
    s MainnetMonetaryState,
    prices map[CryptoAssetKind]float64,
    cfg GRCPolicyConfig,
) (MainnetMonetaryState, GRCIssuanceSignal) {
    sig := ComputeMainnetGRCIssuanceSignal(s, prices, cfg)
    if sig.RecommendedDelta > 0 && sig.Mode == "issue" {
        s = s.WithPendingIssueGRC(sig.RecommendedDelta, "mainnet_grc_policy_issue")
    } else if sig.RecommendedDelta < 0 && sig.Mode == "constrict" {
        s = s.WithPendingBurnGRC(-sig.RecommendedDelta, "mainnet_grc_policy_constrict")
    }
    return s, sig
}

// ComputeMainnetGRCIssuanceSignalAuto is a convenience wrapper that
// pulls the current mainnet state and DevNet price map to produce a
// signal without the caller needing to pass them explicitly. This is
// primarily intended for RPC surfaces and operator tools.
func ComputeMainnetGRCIssuanceSignalAuto() GRCIssuanceSignal {
    s := SnapshotMainnetState()
    prices := getDevnetPriceMap()
    cfg := DefaultGRCPolicyConfig()
    _, sig := ApplyMainnetGRCPolicy(s, prices, cfg)
    return sig
}
