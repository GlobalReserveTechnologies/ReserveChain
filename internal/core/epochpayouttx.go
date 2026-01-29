package core

import (
    "context"
    "fmt"
    "time"

    "reservechain/internal/store"
)

// EpochPayoutCommitTx commits to the epoch's payout ledger (staking + PoP + treasury).
// This does not move balances itself; it is an auditable commitment recorded on-chain.
type EpochPayoutCommitTx struct {
    EpochIndex         uint64 `json:"epoch_index"`
    Author             string `json:"author"`
    PayoutHashHex      string `json:"payout_hash_hex"`
    NumPayouts         int64  `json:"num_payouts"`
    StakeBudgetGRC     float64 `json:"stake_budget_grc"`
    PopBudgetGRC       float64 `json:"pop_budget_grc"`
    TreasuryBudgetGRC  float64 `json:"treasury_budget_grc"`
    Nonce              uint64 `json:"nonce"`
}

const epochPayoutAuthorDefault = "econ"

// ApplyEpochPayoutCommit records a payout commitment as an on-chain tx.
func (c *Chain) ApplyEpochPayoutCommit(tx EpochPayoutCommitTx) (*Block, string, error) {
    c.mu.Lock()
    defer c.mu.Unlock()

    if tx.Author == "" {
        tx.Author = epochPayoutAuthorDefault
    }
    if tx.PayoutHashHex == "" {
        return nil, "", fmt.Errorf("missing payout_hash_hex")
    }
    if err := c.store.ExpectAndIncrementNonce(tx.Author, tx.Nonce); err != nil {
        return nil, "", err
    }

    blk := c.appendBlockLocked("TX_EPOCH_PAYOUT_COMMIT", tx)

    // Best-effort persistence for fast queries.
    if c.db != nil {
        _ = c.db.InsertEpochPayoutCommit(context.Background(), store.EpochPayoutCommit{
            Epoch:             int64(tx.EpochIndex),
            TxHash:            blk.Hash,
            Author:            tx.Author,
            PayoutHash:        tx.PayoutHashHex,
            NumPayouts:        tx.NumPayouts,
            StakeBudgetGRC:    tx.StakeBudgetGRC,
            PopBudgetGRC:      tx.PopBudgetGRC,
            TreasuryBudgetGRC: tx.TreasuryBudgetGRC,
            CreatedAt:         time.Now().UTC(),
        })
    }

    return blk, blk.Hash, nil
}
