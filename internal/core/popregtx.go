package core

import (
    "context"
    "fmt"

    "reservechain/internal/store"
)

// PoPRegisterNodeTx registers (or updates) a PoP node to an operator identity.
// This is recorded as an on-chain transaction so the node registry is auditable.
type PoPRegisterNodeTx struct {
    OperatorWallet string `json:"operator_wallet"`
    NodeID         string `json:"node_id"`
    Role           string `json:"role"`
    Nonce          uint64 `json:"nonce"`
}

// PoPSetCapsTx sets (or updates) a node's capability ceilings used for PoP scoring.
type PoPSetCapsTx struct {
    OperatorWallet  string  `json:"operator_wallet"`
    NodeID          string  `json:"node_id"`
    CPUScore        float64 `json:"cpu_score"`
    RAMScore        float64 `json:"ram_score"`
    StorageScore    float64 `json:"storage_score"`
    BandwidthScore  float64 `json:"bandwidth_score"`
    Nonce           uint64  `json:"nonce"`
}

func (c *Chain) ApplyPoPRegisterNode(tx PoPRegisterNodeTx) (*Block, string, error) {
    c.mu.Lock()
    defer c.mu.Unlock()

    if tx.OperatorWallet == "" || tx.NodeID == "" {
        return nil, "", fmt.Errorf("missing operator_wallet/node_id")
    }
    if err := c.store.ExpectAndIncrementNonce(tx.OperatorWallet, tx.Nonce); err != nil {
        return nil, "", err
    }

    blk := c.appendBlockLocked("TX_POP_REGISTER_NODE", tx)

    if c.db != nil {
        _ = c.db.UpsertPoPNodeWithTxHash(context.Background(), store.PoPNode{
            NodeID:         tx.NodeID,
            OperatorWallet: tx.OperatorWallet,
            Role:           tx.Role,
        }, blk.Hash)
    }

    return blk, blk.Hash, nil
}

func (c *Chain) ApplyPoPSetCaps(tx PoPSetCapsTx) (*Block, string, error) {
    c.mu.Lock()
    defer c.mu.Unlock()

    if tx.OperatorWallet == "" || tx.NodeID == "" {
        return nil, "", fmt.Errorf("missing operator_wallet/node_id")
    }
    if err := c.store.ExpectAndIncrementNonce(tx.OperatorWallet, tx.Nonce); err != nil {
        return nil, "", err
    }

    blk := c.appendBlockLocked("TX_POP_SET_CAPS", tx)

    if c.db != nil {
        _ = c.db.UpsertPoPCapabilityWithTxHash(context.Background(), store.PoPCapability{
            NodeID:         tx.NodeID,
            CPUScore:       tx.CPUScore,
            RAMScore:       tx.RAMScore,
            StorageScore:   tx.StorageScore,
            BandwidthScore: tx.BandwidthScore,
        }, blk.Hash)
    }

    return blk, blk.Hash, nil
}
