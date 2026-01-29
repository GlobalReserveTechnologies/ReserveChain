package analytics

// Package analytics provides higher-level, read-only projections
// on top of the core economics engine. This is where Analytics
// views (NAV curves, corridor volume, treasury history, etc.)
// read from in the workstation.

import "reservechain/internal/econ"

// NAVPoint is a single point in a NAV time series.
type NAVPoint struct {
	Height uint64  `json:"height"`
	NAV    float64 `json:"nav"`
}

// WindowFlow summarises the economic state of a settlement window.
type WindowFlow struct {
	WindowID         int     `json:"window_id"`
	OpenedAtUnix     int64   `json:"opened_at_unix"`
	NextCloseUnix    int64   `json:"next_close_unix"`
	PendingVolumeUSD float64 `json:"pending_volume_usd"`
	SettledVolumeUSD float64 `json:"settled_volume_usd"`
	NetFlowGRC       float64 `json:"net_flow_grc"`
	SettlementNAV    float64 `json:"settlement_nav"`
	Profile          string  `json:"profile"`
	Status           string  `json:"status"`
}

// TreasurySnapshot is a DevNet-friendly view of the reserve basket,
// broken down into liquidity tiers and Tier-2 buckets. For now this is
// derived from NAV and a synthetic notional and does not yet reflect
// real ledger balances (that will come in the settlement phase).
type TreasurySnapshot struct {
	TotalUSD float64        `json:"total_usd"`
	Tier1    Tier1Liquidity `json:"tier1"`
	Tier2    []Tier2Bucket  `json:"tier2"`
}

type Tier1Liquidity struct {
	CashUSD float64 `json:"cash_usd"`
}

type Tier2Bucket struct {
	Type        string  `json:"type"`
	NotionalUSD float64 `json:"notional_usd"`
	Rate        float64 `json:"rate"`
	DurationD   int     `json:"duration_days"`
}

// BuildNAVSeries constructs a simple NAV time series around the current
// snapshot. For DevNet we generate a short synthetic history using the
// current NAV as an anchor.
func BuildNAVSeries(nav econ.NAVSnapshot, points int) []NAVPoint {
	if points <= 0 {
		points = 30
	}
	series := make([]NAVPoint, points)
	base := nav.GRC
	for i := range series {
		// For now we keep it flat; the live WS feed provides motion.
		series[i] = NAVPoint{
			Height: uint64(i),
			NAV:    base,
		}
	}
	return series
}

// BuildWindowFlows converts an econ.WindowSnapshot into a higher-level
// analytics representation. Future versions will add historical flows.
func BuildWindowFlows(ws econ.WindowSnapshot) []WindowFlow {
	return []WindowFlow{{
		WindowID:         ws.WindowID,
		OpenedAtUnix:     ws.OpenedAtUnix,
		NextCloseUnix:    ws.NextCloseUnix,
		PendingVolumeUSD: ws.PendingVolumeUSD,
		SettledVolumeUSD: 0,
		NetFlowGRC:       0,
		SettlementNAV:    0,
		Profile:          string(ws.Profile),
		Status:           string(ws.Status),
	}}
}

// BuildTreasurySnapshot constructs a tiered, MMF-style view of the
// treasury based on the current NAV. This is intentionally simple for
// DevNet: we assume a synthetic supply and derive a basket that is
// split between Tier-1 liquidity and Tier-2 short-duration assets.
func BuildTreasurySnapshot(nav econ.NAVSnapshot, yield econ.YieldSnapshot) TreasurySnapshot {
	// Synthetic supply used for DevNet visualisation only.
	baseSupply := 1_000_000.0
	total := baseSupply * nav.GRC

	// MMF-style split (Option 2 that you selected):
	//   Tier1: 50% cash-like
	//   Tier2: 25% T-Bills, 25% Overnight repo
	tier1 := Tier1Liquidity{
		CashUSD: total * 0.50,
	}

	tBills := Tier2Bucket{
		Type:        "t-bills",
		NotionalUSD: total * 0.25,
		Rate:        yield.TBillYield,
		DurationD:   30,
	}
	repo := Tier2Bucket{
		Type:        "overnight_repo",
		NotionalUSD: total * 0.25,
		Rate:        yield.TBillYield + 0.0025,
		DurationD:   7,
	}

	return TreasurySnapshot{
		TotalUSD: total,
		Tier1:    tier1,
		Tier2:    []Tier2Bucket{tBills, repo},
	}
}
