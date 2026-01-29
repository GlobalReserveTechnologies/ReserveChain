package econ

import (
	"sync"

	"reservechain/internal/core"
	"reservechain/internal/store"
)

// runtime.go
// ------------------------------------------------------------------
// DevNet/Testnet wiring helpers.
// These globals allow the economics epoch settlement scaffolding
// (which is currently in-memory) to credit payouts into the chain
// store and persist payout history into SQLite.
//
// In a production mainnet build, this wiring would be replaced by
// deterministic state transitions and on-chain transactions.

var (
	rtMu    sync.RWMutex
	rtChain *core.Chain
	rtDB    *store.DB
)

// SetRuntime wires the chain + DB into the econ package so that
// epoch settlement helpers can apply staking + PoP payouts.
func SetRuntime(chain *core.Chain, db *store.DB) {
	rtMu.Lock()
	defer rtMu.Unlock()
	rtChain = chain
	rtDB = db
}

func runtimeChain() *core.Chain {
	rtMu.RLock()
	defer rtMu.RUnlock()
	return rtChain
}

func runtimeDB() *store.DB {
	rtMu.RLock()
	defer rtMu.RUnlock()
	return rtDB
}
