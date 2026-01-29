package net

import (
    "encoding/json"
    "net/http"
    "strconv"

    "reservechain/internal/store"
)

// GET /api/econ/epoch-commit?epoch=N
// Returns the on-chain payout commitment for an epoch.
func (api *HTTPAPI) econEpochCommitHandler(w http.ResponseWriter, r *http.Request) {
    epochStr := r.URL.Query().Get("epoch")
    if epochStr == "" {
        http.Error(w, "missing epoch", http.StatusBadRequest)
        return
    }
    epoch, err := strconv.ParseInt(epochStr, 10, 64)
    if err != nil {
        http.Error(w, "invalid epoch", http.StatusBadRequest)
        return
    }

    commit, err := store.GetEpochPayoutCommit(api.Store.DB, epoch)
    if err != nil {
        http.Error(w, err.Error(), http.StatusNotFound)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(commit)
}
