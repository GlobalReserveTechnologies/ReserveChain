package core

import "time"

// EpochScheduler provides a simple mapping between wall‑clock time and
// monotonically increasing epoch indices. It is intentionally minimal
// and suitable for DevNet / testnet style networks where epochs are
// defined as fixed‑length windows since a configured genesis time.
//
// In a full consensus implementation this would be derived from block
// height and/or slots; for now we keep it time‑based so that standalone
// processes (like the DevNet reward loop) can stay in sync without
// referencing chain height.
type EpochScheduler struct {
	// GenesisUnix is the Unix timestamp (seconds) of epoch 0 start.
	GenesisUnix int64

	// EpochSeconds is the fixed length of each epoch window in seconds.
	EpochSeconds int64
}

// EpochIndexForTime returns the 0‑based epoch index for a given point
// in time. If t is before the genesis time, 0 is returned.
func (s EpochScheduler) EpochIndexForTime(t time.Time) uint64 {
	if s.EpochSeconds <= 0 {
		return 0
	}
	delta := t.Unix() - s.GenesisUnix
	if delta <= 0 {
		return 0
	}
	return uint64(delta / s.EpochSeconds)
}

// EpochWindow returns the [start,end) Unix timestamps (seconds) for a
// given epoch index. If EpochSeconds is not set, all epochs collapse to
// the genesis instant.
func (s EpochScheduler) EpochWindow(epochIndex uint64) (startUnix, endUnix int64) {
	if s.EpochSeconds <= 0 {
		return s.GenesisUnix, s.GenesisUnix
	}
	startUnix = s.GenesisUnix + int64(epochIndex)*s.EpochSeconds
	endUnix = startUnix + s.EpochSeconds
	return
}
