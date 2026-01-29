package econ

// coverage_mainnet.go
// ---------------------------------------------------------
// Helpers for computing concise reserve coverage snapshots from the
// in-memory MainnetMonetaryState and the configured crypto-only
// reserve basket.

// AssetCoverageBreakdown exposes per-asset reserve composition.
type AssetCoverageBreakdown struct {
    Kind  CryptoAssetKind `json:"kind"`
    Role  string          `json:"role"`
    USD   float64         `json:"usd"`
    Share float64         `json:"share"`
}

// MainnetCoverageSnapshot is a compact view suitable for RPC / UI.
type MainnetCoverageSnapshot struct {
    Epoch          int64                        `json:"epoch"`
    USDRSupply     float64                      `json:"usdr_supply"`
    RawReservesUSD float64                      `json:"raw_reserves_usd"`
    EffReservesUSD float64                      `json:"effective_reserves_usd"`
    Coverage       float64                      `json:"coverage_multiple"`
    StableShare    float64                      `json:"stable_share"`
    Basket         MainnetReserveBasketConfig   `json:"basket"`
    Assets         []AssetCoverageBreakdown     `json:"assets"`
}

// ComputeMainnetCoverageSnapshot converts a full mainnet monetary
// state into a compact coverage snapshot by joining it with a price
// map and reserve basket configuration.
func ComputeMainnetCoverageSnapshot(
    s MainnetMonetaryState,
    prices map[CryptoAssetKind]float64,
    cfg MainnetReserveBasketConfig,
) MainnetCoverageSnapshot {

    rawUSD, effUSD, byAsset := ComputeHaircuttedReserveUSDFromMainnetState(s, prices, cfg)

    // Compute USDR supply; when 0 we report coverage as 0 to avoid NaN.
    usdrSupply := s.Supply.USDR
    coverage := 0.0
    if usdrSupply > 0 {
        coverage = effUSD / usdrSupply
    }

    // Build per-asset breakdown and track stable share contribution.
    var assets []AssetCoverageBreakdown
    var stableUSD float64
    totalEffUSD := effUSD
    // Build quick role lookup.
    roleOf := make(map[CryptoAssetKind]string)
    for _, ac := range cfg.Assets {
        roleOf[ac.Kind] = ac.Role
    }

    if totalEffUSD <= 0 {
        // No effective reserves; still populate roles/zero values so
        // the UI can show an empty but well-typed table.
        for k := range byAsset {
            role := roleOf[k]
            assets = append(assets, AssetCoverageBreakdown{
                Kind:  k,
                Role:  role,
                USD:   0,
                Share: 0,
            })
        }
    } else {
        for kind, usdRaw := range byAsset {
            role := roleOf[kind]
            // Apply the same haircut factors to derive effective USD so
            // share computation matches EffReservesUSD.
            eff := usdRaw
            for _, ac := range cfg.Assets {
                if ac.Kind == kind && ac.HaircutBps > 0 {
                    factor := 1.0 - float64(ac.HaircutBps)/10000.0
                    if factor < 0 {
                        factor = 0
                    }
                    eff = usdRaw * factor
                    break
                }
            }
            share := 0.0
            if totalEffUSD > 0 {
                share = eff / totalEffUSD
            }
            if role == "stable" {
                stableUSD += eff
            }
            assets = append(assets, AssetCoverageBreakdown{
                Kind:  kind,
                Role:  role,
                USD:   eff,
                Share: share,
            })
        }
    }

    stableShare := 0.0
    if totalEffUSD > 0 && stableUSD > 0 {
        stableShare = stableUSD / totalEffUSD
    }

    return MainnetCoverageSnapshot{
        Epoch:          s.Epoch,
        USDRSupply:     usdrSupply,
        RawReservesUSD: rawUSD,
        EffReservesUSD: effUSD,
        Coverage:       coverage,
        StableShare:    stableShare,
        Basket:         cfg,
        Assets:         assets,
    }
}

// SnapshotCurrentCoverage uses the in-memory singleton state and
// the default price map + reserve basket config to produce a single
// coverage snapshot suitable for RPC exposure.
func SnapshotCurrentCoverage() MainnetCoverageSnapshot {
    s := SnapshotMainnetState()
    prices := getDevnetPriceMap()
    cfg := DefaultMainnetReserveBasketConfig()
    return ComputeMainnetCoverageSnapshot(s, prices, cfg)
}
