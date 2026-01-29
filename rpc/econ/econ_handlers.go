package econ

// Package econ exposes economics + analytics views over RPC.
// This is a thin adapter on top of the internal econ + analytics
// packages so that Workstation / Analytics Terminal / Operator Console
// can consume the same data model.

import (
	"encoding/json"
	"net/http"
	"strconv"

	intecon "reservechain/internal/econ"
)

// AttachHTTP mounts the economics/analytics handlers under a given mux prefix.
// In DevNet this is intentionally minimal; main wiring happens in http_rpc.go.
func AttachHTTP(mux *http.ServeMux, prefix string) {
	mux.HandleFunc(prefix+"/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":  true,
			"msg": "econ RPC online",
		})
	})

	// /econ/treasury returns a live, mark-to-market treasury balance
	// sheet snapshot suitable for the Workstation / Operator Console
	// dashboards.
	mux.HandleFunc(prefix+"/treasury", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		snap := intecon.SnapshotTreasury()
		_ = json.NewEncoder(w).Encode(snap)
	})


	// /econ/coverage returns a compact reserve coverage snapshot based
	// on the current mainnet monetary state, the configured crypto-only
	// reserve basket, and the internal price map. This is intended for
	// operator / analytics use.
	mux.HandleFunc(prefix+"/coverage", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		snap := intecon.SnapshotCurrentCoverage()
		_ = json.NewEncoder(w).Encode(snap)
	})

	// /econ/redemptions returns a summary of the DevNet redemption queue
	// and current epoch state.
	mux.HandleFunc(prefix+"/redemptions", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		snap := intecon.SnapshotDevnetRedemptions()
		_ = json.NewEncoder(w).Encode(snap)
	})

	// /econ/mints returns a summary of the DevNet mint queue state.
	mux.HandleFunc(prefix+"/mints", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		snap := intecon.SnapshotDevnetMints()
		_ = json.NewEncoder(w).Encode(snap)
	})

	
	// /econ/grc-issuance returns the current DevNet GRC issuance
	// recommendation derived from the treasury snapshot.
	mux.HandleFunc(prefix+"/grc-issuance", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		sig := intecon.ComputeMainnetGRCIssuanceSignalAuto()
		_ = json.NewEncoder(w).Encode(sig)
	})

// /econ/advance-epoch forces a DevNet epoch advance, settling both
	// mint and redemption queues. This is intended for operator-only use.
	mux.HandleFunc(prefix+"/advance-epoch", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		intecon.AdvanceDevnetEpoch()
		w.WriteHeader(http.StatusNoContent)
	})
}
	// /econ/mainnet-state exposes the current mainnet monetary state.
	mux.HandleFunc(prefix+"/mainnet-state", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		state := intecon.SnapshotMainnetState()
		_ = json.NewEncoder(w).Encode(state)
	})

	

	// /econ/history returns a rolling window of recent epoch-level
	// monetary history entries for analytics and dashboards. Results
	// are returned in chronological order (oldest first).
	mux.HandleFunc(prefix+"/history", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		limit := 0
		if ls := r.URL.Query().Get("limit"); ls != "" {
			if n, err := strconv.Atoi(ls); err == nil && n > 0 {
				limit = n
			}
		}

		h := intecon.GetMainnetHistory(limit)
		_ = json.NewEncoder(w).Encode(h)
	})
// /econ/settle-mainnet-epoch applies basic mainnet USDR policy and
	// advances the mainnet monetary state by a single epoch. Intended
	// for DevNet/Testnet operator use while wiring the full node.
	mux.HandleFunc(prefix+"/settle-mainnet-epoch", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		state := intecon.SettleMainnetEpochBasic()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(state)
	})


	// /econ/mint-usdr registers a pending USDR mint on the mainnet
	// monetary state. This does not immediately change supply; the
	// mint will only be finalized when a mainnet epoch is settled.
	mux.HandleFunc(prefix+"/mint-usdr", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			Amount  float64 `json:"amount"`
			Account string  `json:"account"`
			Asset   string  `json:"asset"` // optional asset code hint (e.g. "USDC", "ETH")
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"error": "invalid JSON body",
			})
			return
		}
		if req.Amount <= 0 {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"error": "amount must be positive",
			})
			return
		}

		// Enqueue a pending mint on the mainnet monetary state.
		state := intecon.SnapshotMainnetState()
		state = state.WithPendingMintUSDR(req.Amount, req.Account, intecon.AssetCode(req.Asset))
		intecon.SetMainnetState(state)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(state)
	})

	// /econ/redeem-usdr registers a pending USDR redemption on the
	// mainnet monetary state. The redemption will be capped by policy
	// and finalized at epoch settlement time.
	mux.HandleFunc(prefix+"/redeem-usdr", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			Amount  float64 `json:"amount"`
			Account string  `json:"account"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"error": "invalid JSON body",
			})
			return
		}
		if req.Amount <= 0 {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"error": "amount must be positive",
			})
			return
		}

		state := intecon.SnapshotMainnetState()
		state = state.WithPendingRedeemUSDR(req.Amount, req.Account)
		intecon.SetMainnetState(state)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(state)
	})

	// /econ/issue-grc registers a pending GRC issuance entry on the
	// mainnet monetary state. This does not immediately change the
	// GRC supply; it will be applied at the next mainnet epoch
	// settlement after policy evaluation.
	mux.HandleFunc(prefix+"/issue-grc", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			Amount float64 `json:"amount"`
			Reason string  `json:"reason"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"error": "invalid JSON body",
			})
			return
		}
		if req.Amount <= 0 {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"error": "amount must be positive",
			})
			return
		}

		state := intecon.SnapshotMainnetState()
		state = state.WithPendingIssueGRC(req.Amount, req.Reason)
		intecon.SetMainnetState(state)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(state)
	})

	// /econ/burn-grc registers a pending GRC burn entry on the mainnet
	// monetary state. The burn will be applied at the next epoch
	// settlement.
	mux.HandleFunc(prefix+"/burn-grc", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			Amount float64 `json:"amount"`
			Reason string  `json:"reason"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"error": "invalid JSON body",
			})
			return
		}
		if req.Amount <= 0 {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"error": "amount must be positive",
			})
			return
		}

		state := intecon.SnapshotMainnetState()
		state = state.WithPendingBurnGRC(req.Amount, req.Reason)
		intecon.SetMainnetState(state)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(state)
	})
}
