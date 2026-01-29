package econ

import "time"

// InitStateForDevnet ensures the simulated economics engine is
// initialised before ticks begin.
func InitStateForDevnet() {
	// Spin up the synthetic FX / yield / NAV engine and seed a
	// conservative treasury balance sheet so Workstation / Operator
	// Console have meaningful data from the first tick.
	initState()
	InitTreasuryDevnet()
}

// ComputeDevnetTick advances the simulated FX / yield / NAV state,
// advances the window manager, and returns a ValuationTick payload
// suitable for broadcasting over WebSocket.
func ComputeDevnetTick(tickID uint64, leaderID string, wm *WindowManager, now time.Time) ValuationTick {
	fx, y, nav := stepAll()
	wsnap := wm.Advance(now)
	prof := GetProfile()

	return ValuationTick{
		TickID:   tickID,
		LeaderID: leaderID,
		FX:       fx,
		Yield:    y,
		NAV:      nav,
		Windows:  wsnap,
		Profile:  prof,
	}
}
