
package econ

// Mainnet-oriented monetary ledger primitives.
// ---------------------------------------------------------
// This file begins the transition away from the ad‑hoc DevNet
// in‑memory balance sheet helpers toward a stable, mainnet‑style
// monetary state representation.
//
// The goal is that all future issuance / mint / redeem logic is
// expressed in terms of these primitives so the same engine can
// run identically on Testnet and Mainnet, with DevNet using a
// thin wrapper or adapter.

import "time"

// AssetCode is a chain‑level identifier for reserve assets that
// can sit on the ReserveChain balance sheet (e.g. USDC, ETH, WBTC,
// staked derivatives, LSTs, etc.).
type AssetCode string

// MainnetSupplyState tracks total supply for the two core monetary
// instruments: GRC (base money) and USDR (stablecoin).
type MainnetSupplyState struct {
    GRC  float64 `json:"grc"`
    USDR float64 `json:"usdr"`
}

// MainnetReserveBalance represents a single asset position inside a
// reserve pool, expressed in native units (not USD).
type MainnetReserveBalance struct {
    Asset  AssetCode `json:"asset"`
    Amount float64   `json:"amount"`
}

// MainnetReservePool is a logical grouping of reserve balances.
// In mainnet this can map 1:1 to wallets, custody arrangements,
// or on‑chain vault contracts.
type MainnetReservePool struct {
    PoolID   string                  `json:"pool_id"`
    Balances []MainnetReserveBalance `json:"balances"`
}

// MainnetLiabilityState tracks outstanding obligations to holders.
// For now this is aggregated; later this can be extended with
// per‑tranche or per‑instrument breakdowns.
type MainnetLiabilityState struct {
    USDROutstanding float64 `json:"usdr_outstanding"`
}

// MainnetEquityState is the residual interest of the ReserveChain
// treasury after accounting for assets and liabilities.
type MainnetEquityState struct {
    EquityUSD float64 `json:"equity_usd"`
}

// MainnetPendingKind describes a pending settlement entry type
// that will be finalized at an epoch boundary.
type MainnetPendingKind string

const (
    PendingMintUSDR   MainnetPendingKind = "mint_usdr"
    PendingRedeemUSDR MainnetPendingKind = "redeem_usdr"
    PendingIssueGRC   MainnetPendingKind = "issue_grc"
    PendingBurnGRC    MainnetPendingKind = "burn_grc"
)

// MainnetPendingEntry represents a single pending monetary action
// (mint, redeem, issue, burn) that has been accepted at the edge
// of the system but not yet finalized at an epoch boundary.
type MainnetPendingEntry struct {
    Kind      MainnetPendingKind `json:"kind"`
    Amount    float64            `json:"amount"`
    AssetHint string             `json:"asset_hint,omitempty"` // optional: backing asset or pool
    Account   string             `json:"account,omitempty"`     // optional: logical account ref
    CreatedAt time.Time          `json:"created_at"`
}

// MainnetMonetaryState is the canonical snapshot of the ReserveChain
// monetary system at a given point in time. All issuance policy,
// coverage checks, and settlement logic should operate on this
// structure (or diffs of it).
type MainnetMonetaryState struct {
    Epoch        int64                    `json:"epoch"`
    At           time.Time                `json:"at"`
    Supply       MainnetSupplyState       `json:"supply"`
    Reserves     []MainnetReservePool     `json:"reserves"`
    Liabilities  MainnetLiabilityState    `json:"liabilities"`
    Equity       MainnetEquityState       `json:"equity"`
    Pending      []MainnetPendingEntry    `json:"pending"`
    LastEpochAt  *time.Time               `json:"last_epoch_at,omitempty"`
}

// NewMainnetMonetaryState builds an empty monetary state at epoch 0.
// Callers can then seed an initial reserve / equity position.
func NewMainnetMonetaryState() MainnetMonetaryState {
    return MainnetMonetaryState{
        Epoch:       0,
        At:          time.Now().UTC(),
        Supply:      MainnetSupplyState{},
        Reserves:    nil,
        Liabilities: MainnetLiabilityState{},
        Equity:      MainnetEquityState{},
        Pending:     nil,
        LastEpochAt: nil,
    }
}

// Clone returns a deep copy of the monetary state suitable for
// what‑if simulation and policy evaluation.
func (s MainnetMonetaryState) Clone() MainnetMonetaryState {
    cp := s
    if s.Reserves != nil {
        cp.Reserves = make([]MainnetReservePool, len(s.Reserves))
        copy(cp.Reserves, s.Reserves)
        for i := range s.Reserves {
            if s.Reserves[i].Balances != nil {
                cp.Reserves[i].Balances = make([]MainnetReserveBalance, len(s.Reserves[i].Balances))
                copy(cp.Reserves[i].Balances, s.Reserves[i].Balances)
            }
        }
    }
    if s.Pending != nil {
        cp.Pending = make([]MainnetPendingEntry, len(s.Pending))
        copy(cp.Pending, s.Pending)
    }
    if s.LastEpochAt != nil {
        t := *s.LastEpochAt
        cp.LastEpochAt = &t
    }
    return cp
}

// To ease the migration from the DevNet helpers to the mainnet
// ledger, we provide a bridge that can derive a TreasuryBalanceSheet
// style snapshot from a MainnetMonetaryState. This lets existing
// Workstation panels continue to function while new policy code
// moves to the mainnet structures.
func (s MainnetMonetaryState) ToTreasuryBalanceSheet(prices map[CryptoAssetKind]float64) TreasuryBalanceSheet {
    // Compute reserve assets MTM in USD.
    var reservesUSD float64
    for _, pool := range s.Reserves {
        for _, b := range pool.Balances {
            // Map AssetCode -> CryptoAssetKind where possible.
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

    var usdrCoverage float64
    if usdrSupply > 0 {
        usdrCoverage = reservesUSD / usdrSupply
    }

    // EquityUSD is carried separately on the state; if it is zero
    // callers may choose to recompute it as (assets - liabilities)
    // for their own views.
    equityUSD := s.Equity.EquityUSD

    // For now we reuse the existing TreasuryBalanceSheet struct so
    // dashboards do not need to change immediately.
    return TreasuryBalanceSheet{
        At:                s.At,
        ReserveAssetsUSD:  reservesUSD,
        USDRSupply:        usdrSupply,
        GRCSupply:         grcSupply,
        USDReserveCoverage: usdrCoverage,
        GRCCoverage:       0, // will be refined as the GRC model is migrated
        EquityUSD:         equityUSD,
        PendingUSDRRedemptions: 0, // pending queues will be wired in next stages
    }
}

// AppendPending adds a new pending monetary entry to the state and
// returns the updated state. This is a pure functional helper so
// the caller decides when to commit the new state.
func (s MainnetMonetaryState) AppendPending(kind MainnetPendingKind, amount float64, assetHint, account string) MainnetMonetaryState {
    if amount <= 0 {
        return s
    }
    entry := MainnetPendingEntry{
        Kind:      kind,
        Amount:    amount,
        AssetHint: assetHint,
        Account:   account,
        CreatedAt: time.Now().UTC(),
    }
    s.Pending = append(s.Pending, entry)
    return s
}

// WithPendingMintUSDR registers a pending USDR mint for the given
// account and (optionally) backing asset.
func (s MainnetMonetaryState) WithPendingMintUSDR(amount float64, account string, asset AssetCode) MainnetMonetaryState {
    return s.AppendPending(PendingMintUSDR, amount, string(asset), account)
}

// WithPendingRedeemUSDR registers a pending USDR redemption.
func (s MainnetMonetaryState) WithPendingRedeemUSDR(amount float64, account string) MainnetMonetaryState {
    return s.AppendPending(PendingRedeemUSDR, amount, "", account)
}

// WithPendingIssueGRC registers a pending GRC issuance entry.
func (s MainnetMonetaryState) WithPendingIssueGRC(amount float64, reason string) MainnetMonetaryState {
    return s.AppendPending(PendingIssueGRC, amount, "", reason)
}

// WithPendingBurnGRC registers a pending GRC burn entry.
func (s MainnetMonetaryState) WithPendingBurnGRC(amount float64, reason string) MainnetMonetaryState {
    return s.AppendPending(PendingBurnGRC, amount, "", reason)
}

// SettleEpoch applies all pending monetary actions to the state and
// advances the epoch counter. This is deliberately conservative: it
// assumes that policy / risk checks have already been performed on
// the pending entries. Callers that need to reject or reshape the
// pending set should do so before invoking SettleEpoch.
func (s MainnetMonetaryState) SettleEpoch() MainnetMonetaryState {
    now := time.Now().UTC()
    next := s.Clone()

    // Apply pending entries.
    for _, p := range s.Pending {
        switch p.Kind {
        case PendingMintUSDR:
            next.Supply.USDR += p.Amount
            next.Liabilities.USDROutstanding += p.Amount
        case PendingRedeemUSDR:
            // Burn USDR supply and reduce liabilities, with floors at zero.
            next.Supply.USDR -= p.Amount
            if next.Supply.USDR < 0 {
                next.Supply.USDR = 0
            }
            next.Liabilities.USDROutstanding -= p.Amount
            if next.Liabilities.USDROutstanding < 0 {
                next.Liabilities.USDROutstanding = 0
            }
        case PendingIssueGRC:
            next.Supply.GRC += p.Amount
        case PendingBurnGRC:
            next.Supply.GRC -= p.Amount
            if next.Supply.GRC < 0 {
                next.Supply.GRC = 0
            }
        default:
            // Unknown kinds are ignored for now.
        }
    }

    // Clear pending set and advance epoch.
    next.Pending = nil
    next.Epoch = s.Epoch + 1
    next.At = now
    next.LastEpochAt = &now

    return next
}
