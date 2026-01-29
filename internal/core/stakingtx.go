package core

import (
	"context"
	"fmt"

	"reservechain/internal/store"
)

// StakeLockTx locks RSX from a staker and delegates it to a validator.
// Funds are moved from the staker's spendable RSX balance to the
// global staking escrow address ("stake-escrow") to ensure locked
// stake cannot be spent.
type StakeLockTx struct {
	StakerWallet   string  `json:"staker_wallet"`
	ValidatorID    string  `json:"validator_id"`
	AmountRSX      float64 `json:"amount_rsx"`
	LockUntilEpoch int64   `json:"lock_until_epoch"`
	Nonce          uint64  `json:"nonce"`
}

// StakeUnlockTx unlocks previously locked RSX for a staker/validator pair.
// Funds are moved from the staking escrow address back to the staker.
type StakeUnlockTx struct {
	StakerWallet string  `json:"staker_wallet"`
	ValidatorID  string  `json:"validator_id"`
	AmountRSX    float64 `json:"amount_rsx"`
	Nonce        uint64  `json:"nonce"`
}

const stakeEscrowAddress = "stake-escrow"

// ApplyStakeLock locks RSX and updates the staking state. In DevNet this
// creates a new block immediately, similar to ApplyTransfer.
//
// IMPORTANT: the staking tables in SQLite are only mutated from this
// on-chain transaction path (not direct API writes), so stake state
// cannot be spoofed by callers.
func (c *Chain) ApplyStakeLock(tx StakeLockTx) (*Block, string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if tx.StakerWallet == "" || tx.ValidatorID == "" {
		return nil, "", fmt.Errorf("missing staker_wallet/validator_id")
	}
	if tx.AmountRSX <= 0 {
		return nil, "", fmt.Errorf("amount must be positive")
	}
	if tx.LockUntilEpoch < 0 {
		tx.LockUntilEpoch = 0
	}
	// Enforce per-address nonce for stake actions.
	if err := c.store.ExpectAndIncrementNonce(tx.StakerWallet, tx.Nonce); err != nil {
		return nil, "", err
	}

	// Move RSX into escrow.
	if err := c.store.Debit(tx.StakerWallet, "RSX", tx.AmountRSX); err != nil {
		return nil, "", err
	}
	c.store.Credit(stakeEscrowAddress, "RSX", tx.AmountRSX)

	// Persist stake state (only via on-chain tx).
	if c.db != nil {
		_ = c.db.ApplyStakeDelta(context.Background(), tx.StakerWallet, tx.ValidatorID, +tx.AmountRSX, tx.LockUntilEpoch)
	}

	blk := c.appendBlockLocked("TX_STAKE_LOCK", tx)
	return blk, blk.Hash, nil
}

// ApplyStakeUnlock unlocks RSX and updates staking state. Lock expiry is
// enforced at the API layer (which has access to the epoch scheduler/state).
func (c *Chain) ApplyStakeUnlock(tx StakeUnlockTx) (*Block, string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if tx.StakerWallet == "" || tx.ValidatorID == "" {
		return nil, "", fmt.Errorf("missing staker_wallet/validator_id")
	}
	if tx.AmountRSX <= 0 {
		return nil, "", fmt.Errorf("amount must be positive")
	}
	// Enforce nonce.
	if err := c.store.ExpectAndIncrementNonce(tx.StakerWallet, tx.Nonce); err != nil {
		return nil, "", err
	}

	// Ensure position exists and amount is available (DB is authoritative).
	pos := store.StakePosition{}
	if c.db != nil {
		p, err := c.db.GetStakePosition(context.Background(), tx.StakerWallet, tx.ValidatorID)
		if err == nil {
			pos = p
		}
	}
	if pos.AmountRSX <= 0 {
		return nil, "", fmt.Errorf("no stake position")
	}
	if tx.AmountRSX > pos.AmountRSX {
		return nil, "", fmt.Errorf("unlock amount exceeds staked amount")
	}

	// Move RSX from escrow back to staker.
	if err := c.store.Debit(stakeEscrowAddress, "RSX", tx.AmountRSX); err != nil {
		return nil, "", err
	}
	c.store.Credit(tx.StakerWallet, "RSX", tx.AmountRSX)

	if c.db != nil {
		_ = c.db.ApplyStakeDelta(context.Background(), tx.StakerWallet, tx.ValidatorID, -tx.AmountRSX, 0)
	}

	blk := c.appendBlockLocked("TX_STAKE_UNLOCK", tx)
	return blk, blk.Hash, nil
}
