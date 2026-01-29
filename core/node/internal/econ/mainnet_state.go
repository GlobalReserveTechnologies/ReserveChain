
package econ

import (
    "sync"
)

// mainnet_state.go
// ---------------------------------------------------------
// Holds the in-memory mainnet monetary state for the current node
// and provides small helper functions for reading and advancing
// that state. In a full deployment this would be backed by a
// durable store or on-chain ledger; here it is an in-memory
// singleton suitable for DevNet/Testnet-style runs.


var (
    mainnetMu       sync.Mutex
    mainnetState    MainnetMonetaryState
    mainnetStateSet bool

    // mainnetHistory holds a rolling window of recent epoch-level
    // monetary snapshots for analytics and operator tooling. In a
    // full deployment this would be backed by a durable store.
    mainnetHistory        []EpochHistoryEntry
    mainnetHistoryMaxSize = 5000
)


// snapshotMainnetStateLocked returns the current mainnet monetary
// state, initialising it if needed. Caller must hold mainnetMu.
func snapshotMainnetStateLocked() MainnetMonetaryState {
    if !mainnetStateSet {
        mainnetState = NewMainnetMonetaryState()
        mainnetStateSet = true
    }
    return mainnetState
}

// SnapshotMainnetState returns a copy of the current mainnet
// monetary state. This is suitable for exposing via read-only
// RPC surfaces.
func SnapshotMainnetState() MainnetMonetaryState {
    mainnetMu.Lock()
    defer mainnetMu.Unlock()
    return snapshotMainnetStateLocked()
}

// SetMainnetState replaces the current mainnet monetary state with
// the provided value. This is primarily intended for testing or
// controlled migrations.
func SetMainnetState(s MainnetMonetaryState) {
    mainnetMu.Lock()
    defer mainnetMu.Unlock()
    mainnetState = s
    mainnetStateSet = true
}

// SettleMainnetEpochBasic applies a conservative USDR policy and then
// advances the mainnet monetary state by one epoch using
// MainnetMonetaryState.SettleEpoch. It returns the new state.
//
// This is intentionally simple: it reuses the DevNet price map helper
// as a stand-in oracle so that the mainnet engine can be exercised
// before a full oracle pipeline is wired.


// EpochHistoryEntry is a compact record of the monetary state at the
// moment an epoch is settled. It is intentionally smaller than the
// full MainnetMonetaryState while still exposing the key economics
// signals used for analytics and dashboards.
type EpochHistoryEntry struct {
    Epoch             int64   `json:"epoch"`
    AtUnix            int64   `json:"at_unix"`
    USDRSupply        float64 `json:"usdr_supply"`
    GRCSupply         float64 `json:"grc_supply"`
    ReserveAssetsUSD  float64 `json:"reserve_assets_usd"`
    USDROutstanding   float64 `json:"usdr_outstanding"`
    EquityUSD         float64 `json:"equity_usd"`
    GRCMode           string  `json:"grc_mode"`
    GRCDelta          float64 `json:"grc_delta"`
    DeltaUSDR         float64 `json:"delta_usdr"`
    DeltaGRC          float64 `json:"delta_grc"`
    DeltaEquityUSD    float64 `json:"delta_equity_usd"`
    StakePortionUSD   float64 `json:"stake_portion_usd"`
    PoPPortionUSD     float64 `json:"pop_portion_usd"`
    StakeShare        float64 `json:"stake_share"`
    PoPShare          float64 `json:"pop_share"`
    PoPStablePortionUSD float64 `json:"pop_stable_portion_usd"`
    PoPVolPortionUSD    float64 `json:"pop_vol_portion_usd"`
    PoPStableShare      float64 `json:"pop_stable_share"`
    PoPVolShare         float64 `json:"pop_vol_share"`
}

// GetMainnetHistory returns up to `limit` most recent epoch history
// entries in chronological order (oldest first). If limit <= 0, all
// available entries are returned.
func GetMainnetHistory(limit int) []EpochHistoryEntry {
    mainnetMu.Lock()
    defer mainnetMu.Unlock()

    n := len(mainnetHistory)
    if n == 0 {
        return nil
    }
    if limit <= 0 || limit > n {
        limit = n
    }
    start := n - limit
    // Copy to avoid exposing internal slice.
    out := make([]EpochHistoryEntry, limit)
    copy(out, mainnetHistory[start:])
    return out
}
func SettleMainnetEpochBasic() MainnetMonetaryState {
    mainnetMu.Lock()
    defer mainnetMu.Unlock()

    // Capture the state before this epoch settles so we can compute
    // deltas and persist a compact history entry.
    before := snapshotMainnetStateLocked()

    prices := getDevnetPriceMap()
    usdrCfg := DefaultBasicPolicyConfig()
    grcCfg := DefaultGRCPolicyConfig()

    // First enforce USDR coverage / solvency constraints.
    s := ApplyBasicUSDRPolicy(before, prices, usdrCfg)
    // Then derive a hybrid GRC issuance signal and enqueue the
    // corresponding issuance/burn entries before settlement.
    s, sig := ApplyMainnetGRCPolicy(s, prices, grcCfg)

    // Settle the epoch, producing the new state.
    after := s.SettleEpoch()

    // Compute compact deltas for history tracking.
    beforeSupply := before.Supply
    afterSupply := after.Supply

    dUSDR := afterSupply.USDR - beforeSupply.USDR
    dGRC := afterSupply.GRC - beforeSupply.GRC

    // Equity information is available via the existing helpers; for
    // now we reuse the equity estimate from the GRC issuance signal
    // (which is based on reserves and liabilities).
    equityUSD := sig.EquityUSD
    reserveUSD := sig.ReserveAssetsUSD

    // Compute a simple reward split heuristic for this epoch based
    // on the equity estimate. For now we treat positive equity as the
    // aggregate reward pool and split it between RSX stakers (PoS) and
    // PoP operators using a fixed alpha. This keeps the model wired
    // without hard‑coding a particular production ratio.
    totalRewardUSD := equityUSD
    if totalRewardUSD < 0 {
        totalRewardUSD = 0
    }
    split := ComputeEpochRewardSplit(totalRewardUSD, 0.6) // TODO: make alpha configurable

    // For operator rewards (PoP) we currently use a mixed payout:
    // a stable portion (USDR) so operators can cover fiat-denominated
    // costs, and a volatile portion (GRC) for upside exposure. The
    // precise ratios are policy and can be tuned later; for now we
    // start with a 70/30 split in USD terms.
    popStableShare := 0.70
    popVolShare := 0.30
    if popStableShare < 0 {
        popStableShare = 0
    }
    if popVolShare < 0 {
        popVolShare = 0
    }
    if popStableShare+popVolShare == 0 {
        popStableShare = 1
        popVolShare = 0
    }
    popStableUSD := split.PoPPortionUSD * popStableShare
    popVolUSD := split.PoPPortionUSD - popStableUSD


    // Build history entry.
    entry := EpochHistoryEntry{
        Epoch:              after.Epoch,
        AtUnix:             after.At.Unix(),
        USDRSupply:         afterSupply.USDR,
        GRCSupply:          afterSupply.GRC,
        ReserveAssetsUSD:   reserveUSD,
        USDROutstanding:    after.Liabilities.USDROutstanding,
        EquityUSD:          equityUSD,
        GRCMode:            sig.Mode,
        GRCDelta:           sig.RecommendedDelta,
        DeltaUSDR:          dUSDR,
        DeltaGRC:           dGRC,
        DeltaEquityUSD:     equityUSD, // TODO: track true before/after equity when available
        StakePortionUSD:    split.StakePortionUSD,
        PoPPortionUSD:      split.PoPPortionUSD,
        StakeShare:         split.StakeShare,
        PoPShare:           split.PoPShare,
        PoPStablePortionUSD: popStableUSD,
        PoPVolPortionUSD:    popVolUSD,
        PoPStableShare:      popStableShare,
        PoPVolShare:         popVolShare,
    }


// Append to rolling history buffer.
    mainnetHistory = append(mainnetHistory, entry)
    if len(mainnetHistory) > mainnetHistoryMaxSize {
        // Drop oldest entries to enforce the cap.
        trim := len(mainnetHistory) - mainnetHistoryMaxSize
        if trim > 0 && trim < len(mainnetHistory) {
            mainnetHistory = mainnetHistory[trim:]
        }
    }

    mainnetState = after
    mainnetStateSet = true

    return after
}
