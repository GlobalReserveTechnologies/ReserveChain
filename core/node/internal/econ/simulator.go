package econ

import (
	"math"
	"math/rand"
	"sync"
	"time"
)

var (
	mu        sync.Mutex
	inited    bool
	lastFX    FXRates
	lastYield YieldSnapshot
	lastNAV   NAVSnapshot
	rng       *rand.Rand
)

// initState seeds the DevNet economic simulator with
// deterministic-but-random-looking starting values.
func initState() {
	mu.Lock()
	defer mu.Unlock()
	if inited {
		return
	}

	src := rand.NewSource(time.Now().UnixNano())
	rng = rand.New(src)

	lastFX = FXRates{
		EURUSD:  1.05,
		USDCUSD: 1.00,
	}
	lastYield = YieldSnapshot{
		TBillYield: 0.035, // 3.5% annualised
	}
	lastNAV = NAVSnapshot{
		GRC: 1.00,
	}
	inited = true
}

// gbmStep performs a tiny geometric Brownian motion step with
// volatility "vol" around the previous value.
func gbmStep(prev, vol float64) float64 {
	if rng == nil {
		src := rand.NewSource(time.Now().UnixNano())
		rng = rand.New(src)
	}
	dt := 1.0 / (365.0 * 24.0 * 3600.0)
	drift := 0.0
	noise := vol * math.Sqrt(dt) * rng.NormFloat64()
	next := prev * (1.0 + drift + noise)
	return next
}

func clamp(x, lo, hi float64) float64 {
	if x < lo {
		return lo
	}
	if x > hi {
		return hi
	}
	return x
}

// simpleNavStep nudges NAV based on annual yield plus a small noise term.
func simpleNavStep(prev float64, annualYield float64) float64 {
	if rng == nil {
		src := rand.NewSource(time.Now().UnixNano())
		rng = rand.New(src)
	}
	dt := 1.0 / (365.0 * 24.0 * 3600.0)
	drift := prev * annualYield * dt
	noise := prev * 0.0005 * rng.NormFloat64() * math.Sqrt(dt)
	next := prev + drift + noise
	return clamp(next, 0.99, 1.05)
}

// stepAll advances FX, yield and NAV by one tick and returns the
// new snapshots. It is safe to call from multiple goroutines via
// the internal mutex.
func stepAll() (FXRates, YieldSnapshot, NAVSnapshot) {
	mu.Lock()
	defer mu.Unlock()
	if !inited {
		initState()
	}

	// FX: keep EURUSD in a plausible band.
	lastFX.EURUSD = clamp(gbmStep(lastFX.EURUSD, 0.10), 0.95, 1.15)
	lastFX.USDCUSD = 1.0 // DevNet keeps USDC ~ 1.0

	// Yields: light mean-reverting random walk around 3–5%.
	target := 0.04
	speed := 0.05
	shock := 0.002 * rng.NormFloat64()
	lastYield.TBillYield += speed*(target-lastYield.TBillYield) + shock
	lastYield.TBillYield = clamp(lastYield.TBillYield, 0.01, 0.08)

	// NAV: derived from yield with small noise.
	lastNAV.GRC = simpleNavStep(lastNAV.GRC, lastYield.TBillYield)

	return lastFX, lastYield, lastNAV
}

// GetLastNAV returns the latest NAV snapshot (for read-only use).
func GetLastNAV() NAVSnapshot {
	mu.Lock()
	defer mu.Unlock()
	if !inited {
		initState()
	}
	return lastNAV
}

// GetLastFX returns the latest FX snapshot (for read-only use).
func GetLastFX() FXRates {
	mu.Lock()
	defer mu.Unlock()
	if !inited {
		initState()
	}
	return lastFX
}

// GetLastYield returns the latest yield snapshot (for read-only use).
func GetLastYield() YieldSnapshot {
	mu.Lock()
	defer mu.Unlock()
	if !inited {
		initState()
	}
	return lastYield
}
