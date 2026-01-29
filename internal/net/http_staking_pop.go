package net

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"reservechain/internal/core"
	"reservechain/internal/econ"
	"reservechain/internal/store"
)

// /api/staking/validators
//
//	GET  -> list validators
//	POST -> upsert validator
func (api *HTTPAPI) stakingValidatorsHandler(w http.ResponseWriter, r *http.Request) {
	if api.DB == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "DB not configured"})
		return
	}
	ctx := r.Context()
	switch r.Method {
	case http.MethodGet:
		vals, err := api.DB.ListValidators(ctx)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]any{"error": err.Error()})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(vals)
		return
	case http.MethodPost:
		var req store.Validator
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]any{"error": "invalid JSON body"})
			return
		}
		if err := api.DB.UpsertValidator(ctx, req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]any{"error": err.Error()})
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}

// /api/staking/lock (POST)
// Body is core.StakeLockTx.
// This now creates an on-chain transaction (block) and is the only path
// that mutates staking state; direct DB writes are no longer accepted.
func (api *HTTPAPI) stakingLockHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if api.Chain == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "chain not configured"})
		return
	}

	var tx core.StakeLockTx
	if err := json.NewDecoder(r.Body).Decode(&tx); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "invalid JSON body"})
		return
	}

	blk, hash, err := api.Chain.ApplyStakeLock(tx)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": err.Error()})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"tx_hash": hash,
		"height":  blk.Height,
	})
}

// /api/staking/unlock (POST)
// Body is core.StakeUnlockTx.
func (api *HTTPAPI) stakingUnlockHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if api.Chain == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "chain not configured"})
		return
	}

	var tx core.StakeUnlockTx
	if err := json.NewDecoder(r.Body).Decode(&tx); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "invalid JSON body"})
		return
	}

	// Enforce lock expiry at API layer (epoch state lives in econ).
if api.DB != nil {
	pos, err := api.DB.GetStakePosition(r.Context(), tx.StakerWallet, tx.ValidatorID)
	if err == nil {
		curEpoch := econ.CurrentDevnetEpoch()
		if pos.LockUntilEpoch > 0 && curEpoch < pos.LockUntilEpoch {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]any{"error": "stake locked until epoch " + strconv.FormatInt(pos.LockUntilEpoch, 10)})
			return
		}
	}
}

blk, hash, err := api.Chain.ApplyStakeUnlock(tx)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": err.Error()})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"tx_hash": hash,
		"height":  blk.Height,
	})
}

// /api/staking/state (GET)
func (api *HTTPAPI) stakingStateHandler(w http.ResponseWriter, r *http.Request) {
	if api.DB == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "DB not configured"})
		return
	}
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	ctx := r.Context()
	vals, _ := api.DB.ListValidators(ctx)
	stakes, _ := api.DB.ListStakes(ctx)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"validators": vals,
		"stakes":     stakes,
		"epoch":      econ.CurrentDevnetEpoch(),
	})
}

// /api/pop/register-node (POST)
func (api *HTTPAPI) popRegisterNodeHandler(w http.ResponseWriter, r *http.Request) {
    if api.Chain == nil || api.Store == nil {
        w.WriteHeader(http.StatusServiceUnavailable)
        _ = json.NewEncoder(w).Encode(map[string]any{"error": "chain/store not configured"})
        return
    }
    if r.Method != http.MethodPost {
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }

    // Accept both the on-chain tx format and the legacy {node_id, operator_wallet, role} payload.
    var raw map[string]any
    if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
        w.WriteHeader(http.StatusBadRequest)
        _ = json.NewEncoder(w).Encode(map[string]any{"error": "invalid JSON body"})
        return
    }

    tx := core.PoPRegisterNodeTx{}
    if v, ok := raw["operator_wallet"].(string); ok {
        tx.OperatorWallet = v
    }
    if v, ok := raw["node_id"].(string); ok {
        tx.NodeID = v
    }
    if v, ok := raw["role"].(string); ok {
        tx.Role = v
    }
    // nonce optional: if omitted/0, server will use next nonce.
    if v, ok := raw["nonce"].(float64); ok {
        tx.Nonce = uint64(v)
    }

    if tx.OperatorWallet == "" || tx.NodeID == "" {
        w.WriteHeader(http.StatusBadRequest)
        _ = json.NewEncoder(w).Encode(map[string]any{"error": "operator_wallet and node_id required"})
        return
    }
    if tx.Nonce == 0 {
        tx.Nonce = api.Store.GetNonce(tx.OperatorWallet) + 1
    }

    blk, txh, err := api.Chain.ApplyPoPRegisterNode(tx)
    if err != nil {
        w.WriteHeader(http.StatusBadRequest)
        _ = json.NewEncoder(w).Encode(map[string]any{"error": err.Error()})
        return
    }

    _ = json.NewEncoder(w).Encode(map[string]any{
        "ok":      true,
        "tx_hash": txh,
        "block":   blk,
    })
}

// /api/pop/submit-caps (POST)
func (api *HTTPAPI) popSubmitCapsHandler(w http.ResponseWriter, r *http.Request) {
    if api.Chain == nil || api.Store == nil {
        w.WriteHeader(http.StatusServiceUnavailable)
        _ = json.NewEncoder(w).Encode(map[string]any{"error": "chain/store not configured"})
        return
    }
    if r.Method != http.MethodPost {
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }

    var raw map[string]any
    if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
        w.WriteHeader(http.StatusBadRequest)
        _ = json.NewEncoder(w).Encode(map[string]any{"error": "invalid JSON body"})
        return
    }

    tx := core.PoPSetCapsTx{}
    if v, ok := raw["operator_wallet"].(string); ok {
        tx.OperatorWallet = v
    }
    if v, ok := raw["node_id"].(string); ok {
        tx.NodeID = v
    }
    if v, ok := raw["cpu_score"].(float64); ok {
        tx.CPUScore = v
    }
    if v, ok := raw["ram_score"].(float64); ok {
        tx.RAMScore = v
    }
    if v, ok := raw["storage_score"].(float64); ok {
        tx.StorageScore = v
    }
    if v, ok := raw["bandwidth_score"].(float64); ok {
        tx.BandwidthScore = v
    }
    if v, ok := raw["nonce"].(float64); ok {
        tx.Nonce = uint64(v)
    }

    if tx.OperatorWallet == "" || tx.NodeID == "" {
        w.WriteHeader(http.StatusBadRequest)
        _ = json.NewEncoder(w).Encode(map[string]any{"error": "operator_wallet and node_id required"})
        return
    }
    if tx.CPUScore == 0 { tx.CPUScore = 1 }
    if tx.RAMScore == 0 { tx.RAMScore = 1 }
    if tx.StorageScore == 0 { tx.StorageScore = 1 }
    if tx.BandwidthScore == 0 { tx.BandwidthScore = 1 }

    if tx.Nonce == 0 {
        tx.Nonce = api.Store.GetNonce(tx.OperatorWallet) + 1
    }

    blk, txh, err := api.Chain.ApplyPoPSetCaps(tx)
    if err != nil {
        w.WriteHeader(http.StatusBadRequest)
        _ = json.NewEncoder(w).Encode(map[string]any{"error": err.Error()})
        return
    }

    _ = json.NewEncoder(w).Encode(map[string]any{
        "ok":      true,
        "tx_hash": txh,
        "block":   blk,
    })
}

// /api/pop/submit-metrics (POST)
func (api *HTTPAPI) popClaimWorkHandler(w http.ResponseWriter, r *http.Request) {
	if api.DB == nil || api.Chain == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "node not configured"})
		return
	}
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Work claim is an on-chain transaction.
	// We accept the same body shape as PoPMetrics, plus operator_wallet + nonce.
	var body struct {
		OperatorWallet string  `json:"operator_wallet"`
		Nonce          uint64  `json:"nonce"`
		Epoch          int64   `json:"epoch"`
		NodeID         string  `json:"node_id"`
		UptimeScore    float64 `json:"uptime_score"`
		RequestsServed float64 `json:"requests_served"`
		BlocksRelayed  float64 `json:"blocks_relayed"`
		StorageIO      float64 `json:"storage_io"`
		LatencyScore   float64 `json:"latency_score"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "invalid JSON body"})
		return
	}
	if body.NodeID == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "node_id required"})
		return
	}
	if body.OperatorWallet == "" {
		// If operator_wallet omitted, attempt to derive from node registry.
		if n, err := api.DB.GetPoPNode(r.Context(), body.NodeID); err == nil {
			body.OperatorWallet = n.OperatorWallet
		}
	}
	if body.OperatorWallet == "" {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "operator_wallet required"})
		return
	}
	if body.Epoch <= 0 {
		body.Epoch = econ.CurrentDevnetEpoch()
	}

	blk, txh, err := api.Chain.ApplyPoPWorkClaim(core.PoPWorkClaimTx{
		OperatorWallet: body.OperatorWallet,
		NodeID:         body.NodeID,
		Epoch:          body.Epoch,
		UptimeScore:    body.UptimeScore,
		RequestsServed: body.RequestsServed,
		BlocksRelayed:  body.BlocksRelayed,
		StorageIO:      body.StorageIO,
		LatencyScore:   body.LatencyScore,
		Nonce:          body.Nonce,
	})
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": err.Error()})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"ok":      true,
		"tx_hash": txh,
		"height":  blk.Height,
		"epoch":   body.Epoch,
		"ts":      time.Now().UTC(),
	})
}

// /api/pop/submit-metrics (POST)
// Deprecated: kept for backward compatibility. Use /api/pop/claim-work.
func (api *HTTPAPI) popSubmitMetricsHandler(w http.ResponseWriter, r *http.Request) {
	api.popClaimWorkHandler(w, r)
}
}

// /api/pop/payouts?epoch=N (GET)
func (api *HTTPAPI) popPayoutsHandler(w http.ResponseWriter, r *http.Request) {
	if api.DB == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "DB not configured"})
		return
	}
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	epochStr := r.URL.Query().Get("epoch")
	epoch := econ.CurrentDevnetEpoch()
	if epochStr != "" {
		if v, err := strconv.ParseInt(epochStr, 10, 64); err == nil {
			epoch = v
		}
	}
	rows, err := api.DB.ListEpochPayouts(r.Context(), epoch)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": err.Error()})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"epoch":   epoch,
		"payouts": rows,
		"ts":      time.Now().UTC(),
	})
}