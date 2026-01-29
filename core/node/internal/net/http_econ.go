package net

import (
    "encoding/json"
    "math"
    "net/http"
    "time"

    "github.com/gorilla/websocket"

    "reservechain/internal/econ"
)

// econUpgrader is a dedicated WebSocket upgrader for the econ live stream.
var econUpgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        // In a production deployment you would want to tighten this,
        // but for DevNet / internal workstation this is acceptable.
        return true
    },
}

// econLiveHandler exposes a compact coverage / equity snapshot over WebSocket
// so the workstation Reserve System panel can display a live view.
func (api *HTTPAPI) econLiveHandler(w http.ResponseWriter, r *http.Request) {
    conn, err := econUpgrader.Upgrade(w, r, nil)
    if err != nil {
        return
    }
    defer conn.Close()

    ticker := time.NewTicker(2 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-r.Context().Done():
            return
        case <-ticker.C:
            snap := econ.SnapshotCurrentCoverage()
            payload := map[string]interface{}{
                "epoch":    snap.Epoch,
                "coverage": snap.Coverage,
                "reserves": snap.EffReservesUSD,
                "liabs":    snap.USDRSupply,
                "equity":   snap.EffReservesUSD - snap.USDRSupply,
            }
            if err := conn.WriteJSON(payload); err != nil {
                return
            }
        }
    }
}

// econSimHandler runs a workstation-facing what-if simulation for the Reserve System panel.
//
// NOTE: This is a conservative bridge implementation. It provides stable time series output and
// the dual-mode + merge-mode scaffolding needed for the UI. Governance proposal ingestion can
// be wired later once proposal types are implemented on-chain.
func (api *HTTPAPI) econSimHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }

    type overrides struct {
        Alpha            *float64 `json:"alpha,omitempty"`
        CorridorFloor    *float64 `json:"corridor_floor,omitempty"`
        CorridorTarget   *float64 `json:"corridor_target,omitempty"`
        CorridorCeiling  *float64 `json:"corridor_ceiling,omitempty"`
        IssuanceHalfLife *float64 `json:"issuance_half_life,omitempty"`
        TreasurySmoothing *float64 `json:"treasury_smoothing,omitempty"`
        PopShare         *float64 `json:"pop_share,omitempty"`
    }

    type simReq struct {
        NumEpochs int `json:"num_epochs"`
        Mode string `json:"mode"`
        PreferredMerge string `json:"preferred_merge"`
        Overrides overrides `json:"overrides"`
    }

    var req simReq
    _ = json.NewDecoder(r.Body).Decode(&req)
    if req.NumEpochs <= 0 {
        req.NumEpochs = 500
    }
    if req.Mode == "" {
        req.Mode = "current_plus_proposals"
    }
    if req.PreferredMerge == "" {
        req.PreferredMerge = "weighted"
    }

    // Start from live state snapshot.
    snap := econ.SnapshotCurrentCoverage()
    startEpoch := snap.Epoch
    startReserves := snap.EffReservesUSD
    startLiabs := snap.USDRSupply
    if startLiabs <= 0 {
        startLiabs = math.Max(1.0, startReserves*0.75)
    }

    cfg := econ.DefaultRewardEconomicsConfig()

    // Apply overrides.
    if req.Overrides.Alpha != nil {
        cfg.StakeVsPoPAlpha = econClamp(*req.Overrides.Alpha, 0.0, 1.0)
    }
    if req.Overrides.PopShare != nil {
        // PopShare is convenience; if set, map to alpha.
        ps := econClamp(*req.Overrides.PopShare, 0.0, 1.0)
        cfg.StakeVsPoPAlpha = 1.0 - ps
    }
    if req.Overrides.CorridorFloor != nil {
        cfg.CorridorFloor = econClamp(*req.Overrides.CorridorFloor, 0.5, 2.0)
    }
    if req.Overrides.CorridorTarget != nil {
        cfg.CorridorTarget = econClamp(*req.Overrides.CorridorTarget, 0.5, 2.0)
    }
    if req.Overrides.CorridorCeiling != nil {
        cfg.CorridorCeiling = econClamp(*req.Overrides.CorridorCeiling, 0.5, 4.0)
    }
    if req.Overrides.IssuanceHalfLife != nil {
        cfg.IssuanceHalfLife = econClamp(*req.Overrides.IssuanceHalfLife, 500.0, 500000.0)
    }
    if req.Overrides.TreasurySmoothing != nil {
        cfg.TreasurySmoothing = econClamp(*req.Overrides.TreasurySmoothing, 0.0, 1.0)
    }

    // Baseline series (current policy; proposals are not yet wired).
    baseline := econ.RunPolicySimulation(startEpoch, startReserves, startLiabs, req.NumEpochs, cfg)

    // Merged series for the 3 merge modes (max/median/weighted).
    // Since proposals are not yet wired, these currently match baseline; the UI contract is ready.
    merged := []map[string]interface{}{
        {
            "mode": "max_impact",
            "epochs": baseline.Epochs,
            "coverage": baseline.Coverage,
            "equity": baseline.Equity,
        },
        {
            "mode": "median",
            "epochs": baseline.Epochs,
            "coverage": baseline.Coverage,
            "equity": baseline.Equity,
        },
        {
            "mode": "weighted",
            "epochs": baseline.Epochs,
            "coverage": baseline.Coverage,
            "equity": baseline.Equity,
        },
    }

    resp := map[string]interface{}{
        "mode": req.Mode,
        "merge_pref": req.PreferredMerge,
        "baseline": map[string]interface{}{
            "epochs": baseline.Epochs,
            "coverage": baseline.Coverage,
            "equity": baseline.Equity,
            "reserves": baseline.Reserves,
            "liabs": baseline.Liabs,
            "reward_pool_total": baseline.RewardPoolTotal,
            "reward_rsx": baseline.RewardRSX,
            "reward_pop": baseline.RewardPoP,
            "issuance": baseline.Issuance,
            "treasury": baseline.Treasury,
            "corridor": map[string]interface{}{
                "floor": baseline.CorridorFloor,
                "target": baseline.CorridorTarget,
                "ceiling": baseline.CorridorCeiling,
            },
        },
        "proposals": []interface{}{},
        "merged": merged,
        "resolved_config": cfg,
    }

    w.Header().Set("Content-Type", "application/json")
    enc := json.NewEncoder(w)
    enc.SetIndent("", "  ")
    _ = enc.Encode(resp)
}

func econClamp(x, lo, hi float64) float64 {
    if x < lo { return lo }
    if x > hi { return hi }
    return x
}

