
package econ

// Treasury balance sheet + coverage model for DevNet.
// This is a simplified, in-memory representation that can be driven
// either by simulated deposits or (later) by real on-chain telemetry.
// It is intentionally conservative: USDR is treated as fully-reserved
// against the crypto reserve basket, while GRC may be issued on a
// hybrid / partially-fractional basis under the issuance + gating
// rules discussed in the design docs.

import (
    "sync"
    "time"
)

// TreasuryBalanceSheet is a high-level, central-bank-style snapshot of
// the ReserveChain DevNet treasury. All values are expressed in USD
// terms using mark-to-market pricing for the reserve assets.
//
// In mainnet this will likely evolve into a richer structure that is
// backed directly by chain state rather than the in-memory DevNet model.
type TreasuryBalanceSheet struct {
    At time.Time `json:"at"`

    // ReserveAssetsUSD is the total USD mark-to-market value of the
    // crypto reserve basket (BTC/ETH/USDC/USDT/DAI/…).
    ReserveAssetsUSD float64 `json:"reserve_assets_usd"`

    // USDRSupply is the total outstanding supply of USDR, expressed in
    // whole-token units. On DevNet this is simulated; on mainnet it will
    // be wired to the actual chain ledger.
    USDRSupply float64 `json:"usdr_supply"`

    // GRCSupply is the total outstanding supply of GRC (reserve
    // currency). This may be partially fractional relative to the
    // reserve basket, but the coverage ratio is still a useful risk
    // indicator for operators.
    GRCSupply float64 `json:"grc_supply"`

    // PendingUSDRRedemptions represents the queued, not-yet-settled
    // redemption volume (in USDR units) that will be processed at the
    // next epoch boundary.
    PendingUSDRRedemptions float64 `json:"pending_usdr_redemptions"`

    // EquityUSD is a simple Assets - Liabilities view from the
    // perspective of USDR + pending redemption obligations. It is
    // intentionally conservative and does not attempt to mark any
    // seigniorage or senior capital structure.
    EquityUSD float64 `json:"equity_usd"`

    // ReserveCoverage is ReserveAssetsUSD / USDRSupply. For USDR to be
    // fully reserved, this should be >= 1.0. This is the primary safety
    // metric for the stablecoin layer.
    ReserveCoverage float64 `json:"reserve_coverage"`

    // GRCCoverage is ReserveAssetsUSD / GRCSupply. This is expected to
    // be <= 1.0 in a hybrid issuance regime, but is still useful for
    // risk dashboards and long-horizon policy planning.
    GRCCoverage float64 `json:"grc_coverage"`
}

var (
    treasuryMu sync.RWMutex

    // In DevNet we maintain a simple in-memory representation of the
    // crypto reserve pools plus aggregate token supplies. On mainnet
    // these will be populated from chain state + oracle feeds instead.
    treasuryPools            []ReservePoolSnapshot
    treasuryUSDRSupply       float64
    treasuryGRCSupply        float64
    treasuryPendingUSDRQueue float64
)

// InitTreasuryDevnet seeds a conservative, plausible treasury state for
// DevNet so that the Workstation and Operator Console have something
// meaningful to display even before any simulated deposits occur.
func InitTreasuryDevnet() {
    treasuryMu.Lock()
    defer treasuryMu.Unlock()

    if len(treasuryPools) == 0 {
        // Single synthetic pool that roughly resembles a diversified
        // crypto reserve basket. The specific numbers are not critical;
        // they just need to be internally consistent for coverage math.
        treasuryPools = []ReservePoolSnapshot{
            {
                PoolID: ReservePoolID("devnet-main"),
                At:     time.Now().UTC(),
                Balances: []ReservePoolBalance{
                    {Asset: AssetUSDC, Amount: 500_000},
                    {Asset: AssetUSDT, Amount: 250_000},
                    {Asset: AssetDAI, Amount: 250_000},
                    {Asset: AssetETH, Amount: 1_000},
                    {Asset: AssetWBTC, Amount: 40},
                },
            },
        }
    }

    // Seed a conservative starting point: USDR is fully reserved and
    // GRC is issued on top with a slightly lower coverage ratio.
    if treasuryUSDRSupply <= 0 {
        treasuryUSDRSupply = 1_500_000
    }
    if treasuryGRCSupply <= 0 {
        treasuryGRCSupply = 10_000_000
    }
    // No redemptions queued initially.
    if treasuryPendingUSDRQueue < 0 {
        treasuryPendingUSDRQueue = 0
    }
}

// getDevnetPriceMap returns a simple, hard-coded price map for DevNet
// mark-to-market valuation. In a real deployment this would be replaced
// by an oracle-derived price map and likely live outside this package.
func getDevnetPriceMap() map[CryptoAssetKind]float64 {
    return map[CryptoAssetKind]float64{
        AssetUSDC: 1.0,
        AssetUSDT: 1.0,
        AssetDAI:  1.0,
        AssetETH:  3_000.0,
        AssetWBTC: 60_000.0,
    }
}

// SnapshotTreasury builds a point-in-time balance sheet using the
// current in-memory DevNet model plus mark-to-market pricing.
func SnapshotTreasury() TreasuryBalanceSheet {
    treasuryMu.RLock()
    defer treasuryMu.RUnlock()

    prices := getDevnetPriceMap()
    reserveUSD := ComputeReserveNAVUSD(treasuryPools, prices)

    bs := TreasuryBalanceSheet{
        At:                   time.Now().UTC(),
        ReserveAssetsUSD:     reserveUSD,
        USDRSupply:           treasuryUSDRSupply,
        GRCSupply:            treasuryGRCSupply,
        PendingUSDRRedemptions: treasuryPendingUSDRQueue,
    }

    // Equity is assets minus explicit USDR + pending redemption
    // obligations. GRC is treated as base money rather than hard debt.
    bs.EquityUSD = reserveUSD - (treasuryUSDRSupply + treasuryPendingUSDRQueue)

    if treasuryUSDRSupply > 0 {
        bs.ReserveCoverage = reserveUSD / treasuryUSDRSupply
    }
    if treasuryGRCSupply > 0 {
        bs.GRCCoverage = reserveUSD / treasuryGRCSupply
    }

    return bs
}

// The following setters form the DevNet simulation surface. They allow
// higher layers (CLI, HTTP endpoints, test harnesses) to adjust the
// synthetic treasury state without wiring to real on-chain data yet.

// SimulateUSDRSupply sets the total outstanding USDR supply in the
// DevNet treasury model.
func SimulateUSDRSupply(supply float64) {
    treasuryMu.Lock()
    defer treasuryMu.Unlock()
    if supply < 0 {
        supply = 0
    }
    treasuryUSDRSupply = supply
}

// SimulateGRCSupply sets the total outstanding GRC supply in the DevNet
// treasury model.
func SimulateGRCSupply(supply float64) {
    treasuryMu.Lock()
    defer treasuryMu.Unlock()
    if supply < 0 {
        supply = 0
    }
    treasuryGRCSupply = supply
}

// SimulateReservePools replaces the current reserve pool snapshot set
// with a new slice. Callers are responsible for constructing plausible
// balances; this function just installs them.
func SimulateReservePools(pools []ReservePoolSnapshot) {
    treasuryMu.Lock()
    defer treasuryMu.Unlock()
    // Make a shallow copy so callers cannot mutate our slice in-place.
    cp := make([]ReservePoolSnapshot, len(pools))
    copy(cp, pools)
    treasuryPools = cp
}

// SimulatePendingUSDRRedemptions sets the queued-but-not-yet-settled
// redemption volume (in USDR units).
func SimulatePendingUSDRRedemptions(amount float64) {
    treasuryMu.Lock()
    defer treasuryMu.Unlock()
    if amount < 0 {
        amount = 0
    }
    treasuryPendingUSDRQueue = amount
}
