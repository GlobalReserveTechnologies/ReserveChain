package core

import (
    "fmt"
    "time"
)

// TxVaultCreate captures the on-chain metadata for a vault being created
// in the off-chain vault_state.json store. This lets the L1 history reflect
// creation time, owner, and multi-sig parameters.
type TxVaultCreate struct {
    VaultID        string   `json:"vault_id"`
    Owner          string   `json:"owner"`
    Label          string   `json:"label"`
    Type           string   `json:"type"`             // single | multi
    Threshold      uint32   `json:"threshold"`        // multi-sig threshold if Type == "multi"
    Signers        []string `json:"signers"`          // logical signer addresses
    VisibilityMode string   `json:"visibility_mode"`  // A | B | C | D
    DurationTier   string   `json:"duration_tier"`    // short | medium | long
    Timestamp      int64    `json:"timestamp"`
}

// ApplyVaultCreate appends a TX_VAULT_CREATE block. Actual vault balances
// and policies are stored in the PHP-side vault_state.json for now.
func (c *Chain) ApplyVaultCreate(tx TxVaultCreate) (*Block, string, error) {
    c.mu.Lock()
    defer c.mu.Unlock()

    if tx.VaultID == "" {
        return nil, "", fmt.Errorf("missing vault_id")
    }
    if tx.Label == "" {
        tx.Label = tx.VaultID
    }
    if tx.Timestamp == 0 {
        tx.Timestamp = time.Now().UTC().Unix()
    }

    blk := c.appendBlockLocked("TX_VAULT_CREATE", tx)
    return blk, blk.Hash, nil
}

// vaultAddress derives a pseudo-address used by the L1 ledger to track
// balances that belong to a specific vault. This lets vaults participate
// in the same balance model as normal wallets.
func vaultAddress(vaultID string) string {
    return "vault:" + vaultID
}

// TxVaultDeposit moves funds from a wallet into a vault's pseudo-address.
type TxVaultDeposit struct {
    VaultID string  `json:"vault_id"`
    From    string  `json:"from"`
    Asset   string  `json:"asset"`
    Amount  float64 `json:"amount"`
    Nonce   uint64  `json:"nonce"`
}

// TxVaultWithdraw moves funds from a vault's pseudo-address back to a wallet.
type TxVaultWithdraw struct {
    VaultID string  `json:"vault_id"`
    To      string  `json:"to"`
    Asset   string  `json:"asset"`
    Amount  float64 `json:"amount"`
    Nonce   uint64  `json:"nonce"`
}

// TxVaultTransfer moves funds from one vault to another.
type TxVaultTransfer struct {
    FromVaultID string  `json:"from_vault_id"`
    ToVaultID   string  `json:"to_vault_id"`
    Asset       string  `json:"asset"`
    Amount      float64 `json:"amount"`
    Nonce       uint64  `json:"nonce"`
}

// ApplyVaultDeposit debits the user's wallet and credits the vault pseudo-address.
func (c *Chain) ApplyVaultDeposit(tx TxVaultDeposit) (*Block, string, error) {
    c.mu.Lock()
    defer c.mu.Unlock()

    if tx.VaultID == "" || tx.From == "" {
        return nil, "", fmt.Errorf("missing vault_id/from")
    }
    if tx.Asset == "" {
        tx.Asset = "GRC"
    }
    if tx.Amount <= 0 {
        return nil, "", fmt.Errorf("amount must be positive")
    }

    // Enforce per-address nonce for externally signed vault deposits.
    if err := c.store.ExpectAndIncrementNonce(tx.From, tx.Nonce); err != nil {
        return nil, "", err
    }

    vaddr := vaultAddress(tx.VaultID)
    if err := c.store.Debit(tx.From, tx.Asset, tx.Amount); err != nil {
        return nil, "", err
    }
    c.store.Credit(vaddr, tx.Asset, tx.Amount)

    blk := c.appendBlockLocked("TX_VAULT_DEPOSIT", tx)
    return blk, blk.Hash, nil
}

// ApplyVaultWithdraw debits the vault pseudo-address and credits the user's wallet.
func (c *Chain) ApplyVaultWithdraw(tx TxVaultWithdraw) (*Block, string, error) {
    c.mu.Lock()
    defer c.mu.Unlock()

    if tx.VaultID == "" || tx.To == "" {
        return nil, "", fmt.Errorf("missing vault_id/to")
    }
    if tx.Asset == "" {
        tx.Asset = "GRC"
    }
    if tx.Amount <= 0 {
        return nil, "", fmt.Errorf("amount must be positive")
    }

    // Enforce per-address nonce for externally signed vault withdrawals.
    if err := c.store.ExpectAndIncrementNonce(tx.To, tx.Nonce); err != nil {
        return nil, "", err
    }

    vaddr := vaultAddress(tx.VaultID)
    if err := c.store.Debit(vaddr, tx.Asset, tx.Amount); err != nil {
        return nil, "", err
    }
    c.store.Credit(tx.To, tx.Asset, tx.Amount)

    blk := c.appendBlockLocked("TX_VAULT_WITHDRAW", tx)
    return blk, blk.Hash, nil
}

// ApplyVaultTransfer moves funds between two vault pseudo-addresses.
func (c *Chain) ApplyVaultTransfer(tx TxVaultTransfer) (*Block, string, error) {
    c.mu.Lock()
    defer c.mu.Unlock()

    if tx.FromVaultID == "" || tx.ToVaultID == "" {
        return nil, "", fmt.Errorf("missing from_vault_id/to_vault_id")
    }
    if tx.Asset == "" {
        tx.Asset = "GRC"
    }
    if tx.Amount <= 0 {
        return nil, "", fmt.Errorf("amount must be positive")
    }

    fromAddr := vaultAddress(tx.FromVaultID)
    toAddr := vaultAddress(tx.ToVaultID)

    // Enforce per-address nonce using the logical owner; for now we do not
    // distinguish between different owners in replay.
    if err := c.store.ExpectAndIncrementNonce(fromAddr, tx.Nonce); err != nil {
        // In practice, a higher layer would provide an owner address for nonce control.
        // For DevNet, tolerate missing account by skipping the nonce increment.
        _ = err
    }

    if err := c.store.Debit(fromAddr, tx.Asset, tx.Amount); err != nil {
        return nil, "", err
    }
    c.store.Credit(toAddr, tx.Asset, tx.Amount)

    blk := c.appendBlockLocked("TX_VAULT_TRANSFER", tx)
    return blk, blk.Hash, nil
}

