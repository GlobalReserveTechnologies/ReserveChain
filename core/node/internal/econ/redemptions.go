package econ

// DevNet mint / redeem queue + epoch scaffolding.
//
// This file introduces a simple, in-memory model for USDR redemption
// requests and epoch-based settlement. For DevNet this is intentionally
// conservative and operator-focused; mainnet wiring will eventually
// attach this logic to real chain state and policy primitives.

import (
	"fmt"
	"sync"
	"time"
)

type RedemptionStatus string

const (
	RedemptionPending  RedemptionStatus = "pending"
	RedemptionSettled  RedemptionStatus = "settled"
	RedemptionRejected RedemptionStatus = "rejected"
)

// DevnetRedemptionRequest represents a single USDR redemption request
// in the DevNet treasury model. It is deliberately high-level: the
// account identifier is opaque and asset-level breakdown is deferred
// to future iterations when the on-chain ledger is fully wired.
type DevnetRedemptionRequest struct {
	ID         string           `json:"id"`
	CreatedAt  time.Time        `json:"created_at"`
	Epoch      int64            `json:"epoch"`
	AccountRef string           `json:"account_ref"`
	Tier       string           `json:"tier"`
	AmountUSDR float64          `json:"amount_usdr"`
	Status     RedemptionStatus `json:"status"`
}

// DevnetRedemptionSnapshot is a read-only view suitable for RPC / UI.
type DevnetRedemptionSnapshot struct {
	CurrentEpoch       int64                     `json:"current_epoch"`
	Pending            []DevnetRedemptionRequest `json:"pending"`
	LastEpochSettledAt *time.Time                `json:"last_epoch_settled_at,omitempty"`
	TotalPendingUSDR   float64                   `json:"total_pending_usdr"`
}

var (
	redeemMu                 sync.RWMutex
	devnetCurrentEpoch       int64 = 1
	devnetRedemptionsPend          = make([]DevnetRedemptionRequest, 0)
	devnetLastEpochSettledAt *time.Time
)


// CurrentDevnetEpoch returns the current DevNet epoch index.
func CurrentDevnetEpoch() int64 {
	redeemMu.RLock()
	defer redeemMu.RUnlock()
	return devnetCurrentEpoch
}

// EnqueueDevnetRedemption registers a new USDR redemption request to be
// settled at or after the current epoch. For now this only updates the
// in-memory DevNet model; mainnet wiring will eventually connect this
// to actual vault / account state.
func EnqueueDevnetRedemption(accountRef, tier string, amountUSDR float64) DevnetRedemptionRequest {
	redeemMu.Lock()
	defer redeemMu.Unlock()

	if amountUSDR <= 0 {
		amountUSDR = 0
	}

	req := DevnetRedemptionRequest{
		ID:         makeSimpleID("rdm"),
		CreatedAt:  time.Now().UTC(),
		Epoch:      devnetCurrentEpoch,
		AccountRef: accountRef,
		Tier:       tier,
		AmountUSDR: amountUSDR,
		Status:     RedemptionPending,
	}

	devnetRedemptionsPend = append(devnetRedemptionsPend, req)
	// Update the treasury-facing pending queue total.
	SimulatePendingUSDRRedemptions(totalPendingUSDRLocked())

	return req
}

// SnapshotDevnetRedemptions returns a read-only view of the current
// redemption queue state. This is what will be exposed over RPC to the
// Workstation and Operator Console.
func SnapshotDevnetRedemptions() DevnetRedemptionSnapshot {
	redeemMu.RLock()
	defer redeemMu.RUnlock()

	// Copy pending slice to avoid exposing internal backing array.
	pendCopy := make([]DevnetRedemptionRequest, len(devnetRedemptionsPend))
	copy(pendCopy, devnetRedemptionsPend)

	snap := DevnetRedemptionSnapshot{
		CurrentEpoch:     devnetCurrentEpoch,
		Pending:          pendCopy,
		TotalPendingUSDR: totalPendingUSDRLocked(),
	}
	if devnetLastEpochSettledAt != nil {
		snap.LastEpochSettledAt = devnetLastEpochSettledAt
	}
	return snap
}

// AdvanceDevnetEpoch settles the current pending redemption queue and
// moves the DevNet epoch counter forward by one. For now settlement is
// modeled as a single aggregate burn of USDR corresponding to the sum
// of all pending requests.
//
// In a future iteration this function will:
//
/*
   - compute a TWAP-based FX price for the epoch,
   - perform corridor / coverage checks,
   - determine how much of the queue can be honored,
   - adjust the treasury reserve pools accordingly, and
   - emit a richer per-request settlement result.
*/
func AdvanceDevnetEpoch() {
	// Ensure epoch settlement is serialized.
	epochMu.Lock()
	defer epochMu.Unlock()

	// First settle any mint requests that are queued for the current
	// epoch and update the treasury pools / USDR supply.
	settleDevnetMintsForEpoch(devnetCurrentEpoch)

	redeemMu.Lock()
	defer redeemMu.Unlock()

	if len(devnetRedemptionsPend) == 0 {
		devnetCurrentEpoch++
		now := time.Now().UTC()
		devnetLastEpochSettledAt = &now
		// No change to USDR supply / pending queue.
		return
	}

	total := totalPendingUSDRLocked()

	// Adjust USDR supply and clear pending queue at the treasury layer.
	// This is intentionally conservative and assumes full settlement
	// of the queued requests for DevNet.
	treasuryMu.Lock()
	if treasuryUSDRSupply < total {
		treasuryUSDRSupply = 0
	} else {
		treasuryUSDRSupply -= total
	}
	treasuryPendingUSDRQueue = 0
	treasuryMu.Unlock()

	// Mark all pending as settled for this epoch.
	for i := range devnetRedemptionsPend {
		devnetRedemptionsPend[i].Status = RedemptionSettled
	}

	// In DevNet we simply discard settled entries after accounting for
	// the aggregate effect. A more detailed history can be added later.
	devnetRedemptionsPend = devnetRedemptionsPend[:0]

	// After mint/redeem settlement, apply operator reward payouts for this epoch.
	// This wires RSX staking reward splits and PoP scoring/payouts into the DevNet loop.
	cfg := DefaultRewardEconomicsConfig()
	total, op, tr := EpochRewardBudget(uint64(devnetCurrentEpoch), cfg.Issuance)
	_ = total // kept for future dashboard display
	stakeBudget := op * clamp(cfg.StakeVsPoPAlpha, 0.0, 1.0)
	popBudget := op - stakeBudget
	SettleEpochRewardsDevnet(uint64(devnetCurrentEpoch), stakeBudget, popBudget, tr)

	devnetCurrentEpoch++
	now := time.Now().UTC()
	devnetLastEpochSettledAt = &now
}

// totalPendingUSDRLocked assumes redeemMu is already held.
func totalPendingUSDRLocked() float64 {
	var total float64
	for _, r := range devnetRedemptionsPend {
		total += r.AmountUSDR
	}
	return total
}

var (
	simpleIDMu  sync.Mutex
	simpleIDSeq int64
)

// makeSimpleID builds a very small, DevNet-only unique identifier
// using a prefix + Unix timestamp + monotonic counter. This is not
// intended for production use; it is just enough for demo / UI wiring.
func makeSimpleID(prefix string) string {
	simpleIDMu.Lock()
	defer simpleIDMu.Unlock()
	simpleIDSeq++
	return fmt.Sprintf("%s-%d-%d", prefix, time.Now().Unix(), simpleIDSeq)
}
