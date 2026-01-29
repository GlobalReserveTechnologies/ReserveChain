package net

import (
    "encoding/json"
    "net/http"
    "time"

    "reservechain/internal/core"
)

// tierRenewHandler accepts TX_TIER_RENEW from PHP, routes it into the Chain
// engine, and returns the resulting tx_hash. The rich tier business logic
// (Earn application, grace periods, multipliers) lives in PHP; here we
// ensure the payment leg is represented on-chain.
func (api *HTTPAPI) tierRenewHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }

    var body struct {
        Type string        `json:"type"`
        Tx   core.TxTierRenew `json:"tx"`
    }
    if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    if body.Type != "TX_TIER_RENEW" {
        http.Error(w, "invalid type", http.StatusBadRequest)
        return
    }

    blk, txHash, err := api.Chain.ApplyTierRenew(body.Tx)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Broadcast a synthetic event so UIs can react to the tier renewal if needed.
    ev := Event{
        ID:      "tier-renew-" + time.Now().Format(time.RFC3339Nano),
        Type:    EventType("TierRenew"),
        Version: "v1",
        Payload: map[string]interface{}{
            "sender": body.Tx.Sender,
            "tier":   body.Tx.Tier,
            "amount": body.Tx.Payment.AmountGRC,
        },
        Timestamp: time.Now().UTC(),
    }
    api.Hub.Broadcast(ev)

    // Also broadcast a NewBlock event for explorers / dashboards.
    if blk != nil {
        bev := Event{
            ID:        "block-" + time.Now().Format(time.RFC3339Nano),
            Type:      EventNewBlock,
            Version:   "v1",
            Payload:   blk,
            Timestamp: blk.Timestamp,
        }
        api.Hub.Broadcast(bev)
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "success": true,
        "tx_hash": txHash,
    })
}
