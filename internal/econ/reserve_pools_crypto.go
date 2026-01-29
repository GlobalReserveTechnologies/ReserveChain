package econ

import (
    "time"
)

// CryptoAssetKind enumerates the types of on-chain assets that can be
// used as backing for GRC in the crypto-only reserve mode. These are
// symbolic identifiers; pricing and actual token addresses live in a
// higher-level config / oracle process.
type CryptoAssetKind string

const (
    AssetUSDC CryptoAssetKind = "USDC"
    AssetUSDT CryptoAssetKind = "USDT"
    AssetDAI  CryptoAssetKind = "DAI"
    AssetETH  CryptoAssetKind = "ETH"
    AssetWBTC CryptoAssetKind = "WBTC"
)

// ReservePoolID identifies a logical reserve pool. In crypto-only mode a
// pool is typically backed by a smart contract or multisig address that
// holds the underlying assets on-chain.
type ReservePoolID string

// ReservePoolBalance captures the balance of a single asset within a
// reserve pool at a given point in time.
type ReservePoolBalance struct {
    Asset  CryptoAssetKind `json:"asset"`
    Amount float64         `json:"amount"`
}

// ReservePoolSnapshot is a time-stamped view of a single reserve pool.
// In practice these snapshots would be derived from on-chain balances
// plus oracle pricing data.
type ReservePoolSnapshot struct {
    PoolID   ReservePoolID        `json:"pool_id"`
    At       time.Time            `json:"at"`
    Balances []ReservePoolBalance `json:"balances"`
}

// CryptoReserveConfig describes which assets are allowed in the crypto
// reserve basket and what their target weights are. All weights are
// expressed as fractions that should sum to 1.0 for a fully specified
// basket.
type CryptoReserveConfig struct {
    EnabledAssets []CryptoAssetKind        `json:"enabled_assets"`
    TargetWeights map[CryptoAssetKind]float64 `json:"target_weights"`
    // MaxGRCLeverage defines an upper bound on the ratio of GRC supply
    // to total reserve value (e.g. 1.0 for fully collateralized, 0.9 for
    // over-collateralized). The economic layer should refuse to mint if
    // this constraint would be violated.
    MaxGRCLeverage float64 `json:"max_grc_leverage"`
}

// ComputeReserveNAVUSD computes the total USD value of a set of reserve
// pools given a simple price map. prices is expected to map each asset
// kind to a USD price (e.g. 1.0 for USDC/USDT/DAI, ETHUSD, WBTCUSD).
func ComputeReserveNAVUSD(pools []ReservePoolSnapshot, prices map[CryptoAssetKind]float64) float64 {
    var total float64
    for _, p := range pools {
        for _, b := range p.Balances {
            price, ok := prices[b.Asset]
            if !ok {
                continue
            }
            total += b.Amount * price
        }
    }
    return total
}

// ComputeGRCNAV returns the implied NAV (in USD) of 1.0 GRC given the
// total reserve value in USD and the current total supply of GRC units.
// If totalGRCSupply is zero, 0 is returned.
func ComputeGRCNAV(totalReserveUSD float64, totalGRCSupply float64) float64 {
    if totalGRCSupply <= 0 {
        return 0
    }
    return totalReserveUSD / totalGRCSupply
}
