package econ

import (
	"sync"
	"time"

	"reservechain/internal/config"
)

// WindowStatus represents the lifecycle state of a settlement window.
type WindowStatus string

const (
	WindowStatusOpen    WindowStatus = "open"
	WindowStatusClosing WindowStatus = "closing"
	WindowStatusSettled WindowStatus = "settled"
)

// WindowSnapshot is a read-only view of the current window state.
type WindowSnapshot struct {
	WindowID         int               `json:"window_id"`
	Profile          ProfileMode       `json:"profile"`
	Status           WindowStatus      `json:"status"`
	Mode             config.WindowMode `json:"mode"`
	OpenedAtUnix     int64             `json:"opened_at_unix"`
	NextCloseUnix    int64             `json:"next_close_unix"`
	PendingVolumeUSD float64           `json:"pending_volume_usd"`
	LastSettleUnix   int64             `json:"last_settle_unix"`
}

// WindowManager tracks the active window and applies a simple
// DevNet-friendly windowing policy:
//
//   - windows have a fixed length in seconds (cfg.Windows.FixedLengthSeconds)
//   - when a window reaches its end, it is "settled" and a new one opens
//   - pending USD volume is carried into the settlement logic later (Step C)
//
// More advanced behaviours (trigger-based windows, min-volume triggers, etc.)
// can be layered on top in later phases.
type WindowManager struct {
	cfg     *config.NodeConfig
	mu      sync.Mutex
	current WindowSnapshot
}

// NewWindowManager creates a WindowManager seeded from the node config.
func NewWindowManager(cfg *config.NodeConfig) *WindowManager {
	now := time.Now().Unix()
	w := &WindowManager{
		cfg: cfg,
		current: WindowSnapshot{
			WindowID:         1,
			Profile:          GetProfile(),
			Status:           WindowStatusOpen,
			Mode:             cfg.Windows.Mode,
			OpenedAtUnix:     now,
			NextCloseUnix:    now + int64(cfg.Windows.FixedLengthSeconds),
			PendingVolumeUSD: 0,
			LastSettleUnix:   0,
		},
	}
	return w
}

// RecordVolume adds USD notional into the current window's pending volume.
func (wm *WindowManager) RecordVolume(usdNotional float64) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	if wm.current.Status == WindowStatusOpen {
		wm.current.PendingVolumeUSD += usdNotional
	}
}

// Snapshot returns a copy of the current window state.
func (wm *WindowManager) Snapshot() WindowSnapshot {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	return wm.current
}

// Advance progresses the window state based on time and returns a snapshot.
// For DevNet we implement a fixed-length policy only; the Mode field is kept
// so that later phases can branch on it.
func (wm *WindowManager) Advance(now time.Time) WindowSnapshot {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	// Ensure profile tracks whatever the economics layer is using.
	wm.current.Profile = GetProfile()

	elapsed := now.Unix() - wm.current.OpenedAtUnix
	fixedLen := int64(wm.cfg.Windows.FixedLengthSeconds)

	if wm.current.Status == WindowStatusOpen && elapsed >= fixedLen {
		// Mark this window as settled and immediately open a new one.
		wm.current.Status = WindowStatusSettled
		wm.current.LastSettleUnix = now.Unix()

		newID := wm.current.WindowID + 1
		mode := wm.current.Mode

		wm.current = WindowSnapshot{
			WindowID:         newID,
			Profile:          GetProfile(),
			Status:           WindowStatusOpen,
			Mode:             mode,
			OpenedAtUnix:     now.Unix(),
			NextCloseUnix:    now.Unix() + fixedLen,
			PendingVolumeUSD: 0,
			LastSettleUnix:   0,
		}
	} else {
		// Still inside the current window; just keep the next-close estimate fresh.
		wm.current.NextCloseUnix = wm.current.OpenedAtUnix + fixedLen
	}

	return wm.current
}
