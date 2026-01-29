package econ

// reserve_mainnet_basket.go
// ---------------------------------------------------------
// Holds the high-level configuration for the mainnet crypto-only
// reserve basket and convenience helpers for computing haircutted
// reserve value from the in-memory MainnetMonetaryState.
//
// This does not yet enforce the configuration inside ApplyBasicUSDRPolicy;
// instead it provides a focused helper that can be consumed by policy
// and analytics layers as they evolve.

type MainnetReserveAssetConfig struct {
    Kind       CryptoAssetKind `json:"kind"`
    Role       string          `json:"role"`        // e.g. "stable", "base", "btc", "lst"
    HaircutBps int             `json:"haircut_bps"` // risk haircut in basis points
    MaxShare   float64         `json:"max_share"`   // optional soft cap on share of total reserves (0 disables)
}

type MainnetReserveBasketConfig struct {
    Assets        []MainnetReserveAssetConfig   `json:"assets"`
    MinStableShare float64                      `json:"min_stable_share"` // optional minimum share for stables (0 disables)
}

// DefaultMainnetReserveBasketConfig returns a conservative starting
// configuration for the crypto-only reserve basket.
//
// Your choice for Phase A was:
//   - USDC
//   - ETH
//   - WBTC
//   - STETH (LST)
// This function encodes that selection and assigns simple, explainable
// haircuts and soft max share caps.
func DefaultMainnetReserveBasketConfig() MainnetReserveBasketConfig {
    return MainnetReserveBasketConfig{
        Assets: []MainnetReserveAssetConfig{
            {
                Kind:       AssetUSDC,
                Role:       "stable",
                HaircutBps: 25,     // 0.25%% haircut on fiat-backed stable
                MaxShare:   0.80,   // avoid >80%% concentration
            },
            {
                Kind:       AssetETH,
                Role:       "base",
                HaircutBps: 500,    // 5%% haircut to reflect volatility
                MaxShare:   0.60,
            },
            {
                Kind:       AssetWBTC,
                Role:       "btc",
                HaircutBps: 600,    // 6%% haircut; wraps BTC + bridge risk
                MaxShare:   0.50,
            },
            {
                Kind:       AssetSTETH,
                Role:       "lst",
                HaircutBps: 800,    // 8%% haircut for LST slashing + liquidity
                MaxShare:   0.40,
            },
        },
        // Encourage a non-trivial stablecoin base share when possible.
        MinStableShare: 0.20,
    }
}

// ComputeHaircuttedReserveUSDFromMainnetState walks the mainnet
// reserve pools, joins them with the provided price map and basket
// configuration, and returns both the raw and haircutted USD totals.
//
// The per-asset breakdown is returned to allow callers to inspect
// concentration, compute coverage multiples per asset, etc.
func ComputeHaircuttedReserveUSDFromMainnetState(
    s MainnetMonetaryState,
    prices map[CryptoAssetKind]float64,
    cfg MainnetReserveBasketConfig,
) (rawTotalUSD float64, haircuttedTotalUSD float64, byAsset map[CryptoAssetKind]float64) {

    byAsset = make(map[CryptoAssetKind]float64)

    // First aggregate raw USD value per asset.
    for _, pool := range s.Reserves {
        for _, bal := range pool.Balances {
            if bal.Amount <= 0 {
                continue
            }
            // Convert the asset code into our CryptoAssetKind universe.
            kind := CryptoAssetKind(bal.Asset)
            px, ok := prices[kind]
            if !ok {
                continue
            }
            usd := bal.Amount * px
            if usd <= 0 {
                continue
            }
            rawTotalUSD += usd
            byAsset[kind] += usd
        }
    }

    if rawTotalUSD <= 0 {
        return rawTotalUSD, 0, byAsset
    }

    // Build haircut factors from config.
    haircutFactor := map[CryptoAssetKind]float64{}
    for _, ac := range cfg.Assets {
        factor := 1.0
        if ac.HaircutBps > 0 {
            factor = 1.0 - float64(ac.HaircutBps)/10000.0
            if factor < 0 {
                factor = 0
            }
        }
        haircutFactor[ac.Kind] = factor
    }

    // Apply per-asset haircuts to derive the effective reserve value.
    for kind, usd := range byAsset {
        factor, ok := haircutFactor[kind]
        if !ok {
            // If an asset is not explicitly configured, treat it as
            // fully haircutted (0 contribution) to be conservative.
            continue
        }
        haircuttedTotalUSD += usd * factor
    }

    return rawTotalUSD, haircuttedTotalUSD, byAsset
}
