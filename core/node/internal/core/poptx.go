package core

import (
	"context"
	"fmt"

	"reservechain/internal/store"
)

// PoPWorkClaimTx records a node's proof-of-participation work metrics for a given epoch.
// The metrics are submitted as an on-chain transaction so they can be audited via the
// chain log. The reward settlement step later (econ) reads the persisted metrics and
// distributes PoP budget accordingly.
type PoPWorkClaimTx struct {
	OperatorWallet string  `json:"operator_wallet"`
	NodeID         string  `json:"node_id"`
	Epoch          int64   `json:"epoch"`
	UptimeScore    float64 `json:"uptime_score"`
	RequestsServed float64 `json:"requests_served"`
	BlocksRelayed  float64 `json:"blocks_relayed"`
	StorageIO      float64 `json:"storage_io"`
	LatencyScore   float64 `json:"latency_score"`
	Nonce          uint64  `json:"nonce"`
}

// ApplyPoPWorkClaim appends an on-chain work-claim tx and persists the metrics into SQLite.
// IMPORTANT: pop_epoch_metrics is only mutated via this on-chain path, so callers cannot
// spoof PoP metrics by writing directly to the DB through an API endpoint.
func (c *Chain) ApplyPoPWorkClaim(tx PoPWorkClaimTx) (*Block, string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if tx.OperatorWallet == "" || tx.NodeID == "" {
		return nil, "", fmt.Errorf("missing operator_wallet/node_id")
	}
	if tx.Epoch <= 0 {
		return nil, "", fmt.Errorf("missing epoch")
	}
	// Enforce per-operator nonce.
	if err := c.store.ExpectAndIncrementNonce(tx.OperatorWallet, tx.Nonce); err != nil {
		return nil, "", err
	}

	// Verify the node registry (if DB available).
	if c.db != nil {
		n, err := c.db.GetPoPNode(context.Background(), tx.NodeID)
		if err != nil {
			return nil, "", fmt.Errorf("unknown node_id")
		}
		if n.OperatorWallet != tx.OperatorWallet {
			return nil, "", fmt.Errorf("operator_wallet does not match node registry")
		}
	}

	// Append block first so we can use the block hash as tx_hash for idempotent metric inserts.
	blk := c.appendBlockLocked("TX_POP_WORK_CLAIM", tx)

	// Persist metrics (auditable) keyed by tx hash to avoid replay duplicates.
	if c.db != nil {
		_ = c.db.InsertPoPMetricsWithTxHash(context.Background(), store.PoPMetrics{
			Epoch:          tx.Epoch,
			NodeID:         tx.NodeID,
			UptimeScore:    tx.UptimeScore,
			RequestsServed: tx.RequestsServed,
			BlocksRelayed:  tx.BlocksRelayed,
			StorageIO:      tx.StorageIO,
			LatencyScore:   tx.LatencyScore,
		}, blk.Hash)
	}

	return blk, blk.Hash, nil
}
