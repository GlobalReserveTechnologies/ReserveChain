package econ

type FXRates struct {
    EURUSD  float64 `json:"eur_usd"`
    USDCUSD float64 `json:"usdc_usd"`
}

type YieldSnapshot struct {
    TBillYield float64 `json:"t_bill_yield"`
}

type NAVSnapshot struct {
    GRC float64 `json:"grc"`
}

type ValuationTick struct {
    TickID   uint64         `json:"tick_id"`
    LeaderID string         `json:"leader_id"`
    FX       FXRates        `json:"fx"`
    Yield    YieldSnapshot  `json:"yield"`
    NAV      NAVSnapshot    `json:"nav"`
    Windows  WindowSnapshot `json:"windows"`
    Profile  ProfileMode    `json:"profile"`
}

// ComputeCorridorBounds returns the lower and upper bounds of the
// corridor around a given NAV using a symmetric band specified in
// basis points (1 basis point = 0.01%%). For example, bandBps=10
// represents a ±0.10%% corridor. If nav is non-positive, both bounds
// are returned as 0.
func ComputeCorridorBounds(nav float64, bandBps float64) (lower, upper float64) {
    if nav <= 0 || bandBps < 0 {
        return 0, 0
    }
    // Convert bps to fractional width, e.g. 10 bps => 0.0010
    width := bandBps / 10000.0
    lower = nav * (1.0 - width)
    upper = nav * (1.0 + width)
    return
}

