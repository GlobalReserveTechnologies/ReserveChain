package core

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reservechain/internal/store"
	"sync"
	"sync/atomic"
	"time"
)

// pendingTx holds a txType + body pair waiting to be mined by the Miner.
type pendingTx struct {
	Type string
	Body interface{}
}

// Store exposes the underlying in-memory AccountStore.
// This is primarily used by DevNet economics wiring (staking/PoP payouts).
func (c *Chain) Store() *AccountStore { return c.store }

// MuRLock exposes the read lock for external callers that need a safe snapshot.
func (c *Chain) MuRLock() { c.mu.RLock() }

// MuRUnlock releases the read lock acquired via MuRLock.
func (c *Chain) MuRUnlock() { c.mu.RUnlock() }

// PendingTxsSnapshot returns a shallow copy of the current mempool.
func (c *Chain) PendingTxsSnapshot() []pendingTx {
	out := make([]pendingTx, len(c.pendingTxs))
	copy(out, c.pendingTxs)
	return out
}

// AppendRemoteBlock appends a block that was mined by another node.
// This is used by simple follower nodes that synchronise via HTTP APIs
// rather than participating in PoW directly. We trust the provided
// header fields in DevNet.
func (c *Chain) AppendRemoteBlock(blk *Block) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.blocks = append(c.blocks, blk)
}

// Block represents a PoW‑secured L1 block for DevNet.
//
// Each block includes a basic proof‑of‑work over its header + transaction
// payload. While DevNet currently runs as a single node, this PoW is
// fully functional and can be used as the consensus backbone when peers
// are introduced later.
type Block struct {
	Height     uint64      `json:"height"`
	PrevHash   string      `json:"prev_hash"`
	Hash       string      `json:"hash"`
	Timestamp  time.Time   `json:"timestamp"`
	TxType     string      `json:"tx_type"`
	Tx         interface{} `json:"tx"`
	Nonce      uint64      `json:"nonce"`
	Difficulty uint32      `json:"difficulty"`
}

// Chain is an in-memory ledger + block log wrapped around the AccountStore.
//
// For DevNet we keep it simple: every call to Apply* creates a new block
// and applies the associated balance changes atomically. This is enough
// to make the rest of the system behave "as if" there is a working L1.
type Chain struct {
	mu         sync.RWMutex
	store      *AccountStore
	blocks     []*Block
	db         *store.DB
	pendingTxs []pendingTx
}

// allowedBackingAssets enumerates which assets can be used as backing for
// GRC mint/redeem operations in DevNet's crypto-only reserve mode.
var allowedBackingAssets = map[string]bool{
	"USDC": true,
	"USDT": true,
	"DAI":  true,
}

// isAllowedBackingAsset reports whether the given symbol is permitted as
// a backing asset for GRC mint/redeem. This keeps DevNet economic flows
// constrained to a small, well-understood set of stable-like assets.
func isAllowedBackingAsset(asset string) bool {
	return allowedBackingAssets[asset]
}

// // devnetPriceUSD provides a coarse USD price map for DevNet
// so that reserves held in multiple assets can be valued on a
// common USD-like basis. In production this would be replaced by
// oracle-driven pricing and the ReservePools machinery.
var devnetPriceUSD = map[string]float64{
	"USDC": 1.0,
	"USDT": 1.0,
	"DAI":  1.0,
	"ETH":  2000.0,
	"WBTC": 40000.0,
}

// computeDevnetNAVLocked computes a simple NAV estimate for GRC based on
// the current in-memory account store. It expects the Chain mutex to be
// held by the caller. For DevNet we treat treasury balances on the
// treasury account as the reserve backing GRC, valued using the
// devnetPriceUSD map.
func (c *Chain) computeDevnetNAVLocked() (nav float64) {
	// Sum reserve balances on the treasury account and value them
	// using the DevNet price map.
	snapAll := c.store.SnapshotAll()
	var reserve float64
	for _, acc := range snapAll {
		if acc.Address != "treasury" {
			continue
		}
		reserve += acc.Balances["USDC"] * devnetPriceUSD["USDC"]
		reserve += acc.Balances["USDT"] * devnetPriceUSD["USDT"]
		reserve += acc.Balances["DAI"] * devnetPriceUSD["DAI"]
		reserve += acc.Balances["ETH"] * devnetPriceUSD["ETH"]
		reserve += acc.Balances["WBTC"] * devnetPriceUSD["WBTC"]
	}

	// Sum total GRC supply across all accounts.
	var supply float64
	for _, acc := range snapAll {
		supply += acc.Balances["GRC"]
	}

	if supply <= 0 {
		// If there is no supply yet, default NAV to 1.0 so that the
		// first mint operation behaves like a 1:1 mapping.
		return 1.0
	}
	return reserve / supply
}

// DevnetMonetarySnapshot returns a coarse monetary snapshot derived from
// in-memory account state. It computes the same reserve and supply
// quantities used by computeDevnetNAVLocked but is safe to call without
// holding the Chain mutex.
func (c *Chain) DevnetMonetarySnapshot() (reserve, supply, nav float64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	snapAll := c.store.SnapshotAll()
	for _, acc := range snapAll {
		if acc.Address == "treasury" {
			reserve += acc.Balances["USDC"] * devnetPriceUSD["USDC"]
			reserve += acc.Balances["USDT"] * devnetPriceUSD["USDT"]
			reserve += acc.Balances["DAI"] * devnetPriceUSD["DAI"]
			reserve += acc.Balances["ETH"] * devnetPriceUSD["ETH"]
			reserve += acc.Balances["WBTC"] * devnetPriceUSD["WBTC"]
		}
		supply += acc.Balances["GRC"]
	}

	if supply <= 0 {
		nav = 1.0
	} else {
		nav = reserve / supply
	}
	return
}

// NewChain creates a Chain. If a chain log already exists in the DB,
// it is replayed to reconstruct state; otherwise a fresh genesis block
// is appended and written out.
func NewChain(store *AccountStore, db *store.DB) *Chain {
	c := &Chain{
		store:      store,
		blocks:     make([]*Block, 0, 1024),
		db:         db,
		pendingTxs: make([]pendingTx, 0, 128),
	}

	ctx := context.Background()
	if db != nil {
		if _, txs, err := db.LoadAllBlocks(ctx); err == nil && len(txs) > 0 {
			// Rebuild in-memory state from the chain log.
			if err := c.replayStateFromTxRows(txs); err == nil {
				// Also reconstruct a minimal block header chain for explorer-style APIs.
				// We do not re-hash; we trust the DB contents for DevNet.
				blks, _, err2 := db.LoadAllBlocks(ctx)
				if err2 == nil {
					for _, b := range blks {
						blk := &Block{
							Height:     b.Height,
							PrevHash:   b.PrevHash,
							Hash:       b.Hash,
							Timestamp:  b.Timestamp,
							TxType:     b.TxType,
							Tx:         json.RawMessage(b.TxJSON),
							Nonce:      b.Nonce,
							Difficulty: b.Difficulty,
						}
						c.blocks = append(c.blocks, blk)
					}
				}
			}
		}
	}

	if len(c.blocks) == 0 {
		// No existing chain; start from a fresh genesis block.
		c.appendGenesisBlockLocked()
	}

	return c
}

// appendBlockLocked assumes c.mu is held.

// replayStateFromTxRows replays balance-affecting transactions from the
// persisted chain_tx rows. For DevNet we keep this intentionally simple
// and only handle core types (transfer, mint, redeem, tier renewals,
// and vault operations). Nonce checks are skipped in replay to avoid
// failures if historical nonces do not match the fresh in-memory state.

// ReplayFromTxRows replays a slice of chain transaction rows onto the in-memory
// state store. It is intended for follower nodes that pull new blocks from an
// upstream peer and need to incrementally apply their effects.
func (c *Chain) ReplayFromTxRows(txs []store.ChainTxRow) error {
	return c.replayStateFromTxRows(txs)
}

func (c *Chain) replayStateFromTxRows(txs []store.ChainTxRow) error {
	for _, row := range txs {
		switch row.TxType {
		case "TX_TRANSFER":
			var tx TransferTx
			if err := json.Unmarshal([]byte(row.BodyJSON), &tx); err != nil {
				continue
			}
			if tx.Asset == "" {
				tx.Asset = "GRC"
			}
			if tx.Amount <= 0 {
				continue
			}
			// Re-enforce nonce semantics during replay so nonces and balances
			// match a live chain execution.
			if err := c.store.ExpectAndIncrementNonce(tx.From, tx.Nonce); err != nil {
				continue
			}
			if err := c.store.Debit(tx.From, tx.Asset, tx.Amount); err != nil {
				continue
			}
			c.store.Credit(tx.To, tx.Asset, tx.Amount)

		case "TX_MINT":
			var tx MintTx
			if err := json.Unmarshal([]byte(row.BodyJSON), &tx); err != nil {
				continue
			}
			if tx.Asset == "" {
				tx.Asset = "GRC"
			}
			if tx.Amount <= 0 {
				continue
			}
			c.store.Credit(tx.To, tx.Asset, tx.Amount)

		case "TX_REDEEM":
			var tx RedeemTx
			if err := json.Unmarshal([]byte(row.BodyJSON), &tx); err != nil {
				continue
			}
			if tx.Asset == "" {
				tx.Asset = "GRC"
			}
			if tx.Amount <= 0 {
				continue
			}
			// For replay we simply burn from the address in question.
			_ = c.store.Debit(tx.Address, tx.Asset, tx.Amount)

		case "TX_TIER_RENEW":
			var tx TxTierRenew
			if err := json.Unmarshal([]byte(row.BodyJSON), &tx); err != nil {
				continue
			}
			payAmt := tx.Payment.AmountGRC
			if payAmt <= 0 || tx.Sender == "" {
				continue
			}
			// Enforce nonce like live ApplyTierRenew.
			if err := c.store.ExpectAndIncrementNonce(tx.Sender, tx.Nonce); err != nil {
				continue
			}
			if err := c.store.Debit(tx.Sender, "GRC", payAmt); err != nil {
				continue
			}
			c.store.Credit("treasury-tiers", "GRC", payAmt)

		case "TX_STAKE_LOCK":
	var tx StakeLockTx
	if err := json.Unmarshal([]byte(row.BodyJSON), &tx); err != nil {
		continue
	}
	if tx.StakerWallet == "" || tx.ValidatorID == "" || tx.AmountRSX <= 0 {
		continue
	}
	if err := c.store.ExpectAndIncrementNonce(tx.StakerWallet, tx.Nonce); err != nil {
		continue
	}
	if err := c.store.Debit(tx.StakerWallet, "RSX", tx.AmountRSX); err != nil {
		continue
	}
	c.store.Credit(stakeEscrowAddress, "RSX", tx.AmountRSX)
	if c.db != nil {
		_ = c.db.ApplyStakeDelta(context.Background(), tx.StakerWallet, tx.ValidatorID, +tx.AmountRSX, tx.LockUntilEpoch)
	}

case "TX_STAKE_UNLOCK":
	var tx StakeUnlockTx
	if err := json.Unmarshal([]byte(row.BodyJSON), &tx); err != nil {
		continue
	}
	if tx.StakerWallet == "" || tx.ValidatorID == "" || tx.AmountRSX <= 0 {
		continue
	}
	if err := c.store.ExpectAndIncrementNonce(tx.StakerWallet, tx.Nonce); err != nil {
		continue
	}
	// During replay we do not enforce lock expiry (epoch may differ); we only
	// ensure we don't unlock more than escrow has and stake position exists.
	if err := c.store.Debit(stakeEscrowAddress, "RSX", tx.AmountRSX); err != nil {
		continue
	}
	c.store.Credit(tx.StakerWallet, "RSX", tx.AmountRSX)
	if c.db != nil {
		_ = c.db.ApplyStakeDelta(context.Background(), tx.StakerWallet, tx.ValidatorID, -tx.AmountRSX, 0)
	}


case "TX_POP_REGISTER_NODE":
	var tx PoPRegisterNodeTx
	if err := json.Unmarshal([]byte(row.BodyJSON), &tx); err != nil {
		continue
	}
	if tx.OperatorWallet == "" || tx.NodeID == "" {
		continue
	}
	if err := c.store.ExpectAndIncrementNonce(tx.OperatorWallet, tx.Nonce); err != nil {
		continue
	}
	if c.db != nil {
		_ = c.db.UpsertPoPNodeWithTxHash(context.Background(), store.PoPNode{
			NodeID:         tx.NodeID,
			OperatorWallet: tx.OperatorWallet,
			Role:           tx.Role,
		}, row.TxHash)
	}

case "TX_POP_SET_CAPS":
	var tx PoPSetCapsTx
	if err := json.Unmarshal([]byte(row.BodyJSON), &tx); err != nil {
		continue
	}
	if tx.OperatorWallet == "" || tx.NodeID == "" {
		continue
	}
	if err := c.store.ExpectAndIncrementNonce(tx.OperatorWallet, tx.Nonce); err != nil {
		continue
	}
	if c.db != nil {
		_ = c.db.UpsertPoPCapabilityWithTxHash(context.Background(), store.PoPCapability{
			NodeID:         tx.NodeID,
			CPUScore:       tx.CPUScore,
			RAMScore:       tx.RAMScore,
			StorageScore:   tx.StorageScore,
			BandwidthScore: tx.BandwidthScore,
		}, row.TxHash)
	}
case "TX_POP_WORK_CLAIM":
	var tx PoPWorkClaimTx
	if err := json.Unmarshal([]byte(row.BodyJSON), &tx); err != nil {
		continue
	}
	if tx.OperatorWallet == "" || tx.NodeID == "" || tx.Epoch <= 0 {
		continue
	}
	// Enforce nonce.
	if err := c.store.ExpectAndIncrementNonce(tx.OperatorWallet, tx.Nonce); err != nil {
		continue
	}
	// Persist metrics idempotently (tx hash).
	if c.db != nil {
		_ = c.db.InsertPoPMetricsWithTxHash(context.Background(), store.PoPMetrics{
			Epoch:          tx.Epoch,
			NodeID:         tx.NodeID,
			UptimeScore:    tx.UptimeScore,
			RequestsServed: tx.RequestsServed,
			BlocksRelayed:  tx.BlocksRelayed,
			StorageIO:      tx.StorageIO,
			LatencyScore:   tx.LatencyScore,
		}, row.TxHash)
	}
case "TX_VAULT_CREATE":
			// Metadata-only at the chain layer for now; no balance effect.

case "TX_EPOCH_PAYOUT_COMMIT":
    var tx EpochPayoutCommitTx
    if err := json.Unmarshal([]byte(row.BodyJSON), &tx); err != nil {
        continue
    }
    if tx.Author == "" {
        tx.Author = epochPayoutAuthorDefault
    }
    if tx.PayoutHashHex == "" {
        continue
    }
    // Enforce nonce semantics for the author during replay.
    if err := c.store.ExpectAndIncrementNonce(tx.Author, tx.Nonce); err != nil {
        continue
    }
    // Best-effort persist commit row.
    if c.db != nil {
        _ = c.db.InsertEpochPayoutCommit(context.Background(), store.EpochPayoutCommit{
            Epoch:             int64(tx.EpochIndex),
            TxHash:            row.TxHash,
            Author:            tx.Author,
            PayoutHash:        tx.PayoutHashHex,
            NumPayouts:        tx.NumPayouts,
            StakeBudgetGRC:    tx.StakeBudgetGRC,
            PopBudgetGRC:      tx.PopBudgetGRC,
            TreasuryBudgetGRC: tx.TreasuryBudgetGRC,
            CreatedAt:         time.Now().UTC(),
        })
    }

		case "TX_VAULT_DEPOSIT":
			var tx TxVaultDeposit
			if err := json.Unmarshal([]byte(row.BodyJSON), &tx); err != nil {
				continue
			}
			if tx.Amount <= 0 || tx.VaultID == "" || tx.From == "" {
				continue
			}
			if tx.Asset == "" {
				tx.Asset = "GRC"
			}
			if err := c.store.ExpectAndIncrementNonce(tx.From, tx.Nonce); err != nil {
				continue
			}
			vaddr := vaultAddress(tx.VaultID)
			if err := c.store.Debit(tx.From, tx.Asset, tx.Amount); err != nil {
				continue
			}
			c.store.Credit(vaddr, tx.Asset, tx.Amount)

		case "TX_VAULT_WITHDRAW":
			var tx TxVaultWithdraw
			if err := json.Unmarshal([]byte(row.BodyJSON), &tx); err != nil {
				continue
			}
			if tx.Amount <= 0 || tx.VaultID == "" || tx.To == "" {
				continue
			}
			if tx.Asset == "" {
				tx.Asset = "GRC"
			}
			if err := c.store.ExpectAndIncrementNonce(tx.To, tx.Nonce); err != nil {
				continue
			}
			vaddr := vaultAddress(tx.VaultID)
			if err := c.store.Debit(vaddr, tx.Asset, tx.Amount); err != nil {
				continue
			}
			c.store.Credit(tx.To, tx.Asset, tx.Amount)

		case "TX_VAULT_TRANSFER":
			var tx TxVaultTransfer
			if err := json.Unmarshal([]byte(row.BodyJSON), &tx); err != nil {
				continue
			}
			if tx.Amount <= 0 || tx.FromVaultID == "" || tx.ToVaultID == "" {
				continue
			}
			if tx.Asset == "" {
				tx.Asset = "GRC"
			}
			fromAddr := vaultAddress(tx.FromVaultID)
			toAddr := vaultAddress(tx.ToVaultID)
			if err := c.store.Debit(fromAddr, tx.Asset, tx.Amount); err != nil {
				continue
			}
			c.store.Credit(toAddr, tx.Asset, tx.Amount)
		}
	}
	return nil
}

// enqueueTx adds a transaction to the in-memory mempool. State mutations are
// performed by the Apply* functions, while the Miner is responsible for
// turning these into PoW-backed blocks and writing them to the chain log.
func (c *Chain) enqueueTx(txType string, body interface{}) {
	c.pendingTxs = append(c.pendingTxs, pendingTx{Type: txType, Body: body})
}

func (c *Chain) appendBlockLocked(txType string, txBody interface{}) *Block {
	height := uint64(len(c.blocks))
	prevHash := ""
	var prevDifficulty uint32 = 4
	var prevTimestamp time.Time

	if height > 0 {
		prev := c.blocks[height-1]
		prevHash = prev.Hash
		prevDifficulty = prev.Difficulty
		prevTimestamp = prev.Timestamp
	}

	payload, _ := json.Marshal(txBody)

	// Adaptive PoW difficulty: try to keep an approximate target block time
	// by nudging the difficulty up or down based on the observed inter‑block
	// interval. This is DevNet‑oriented and can be replaced with a more formal
	// retargeting rule later.
	const (
		targetBlockSeconds = 10.0
		minDifficulty      = 2
		maxDifficulty      = 8
	)

	difficulty := prevDifficulty
	if height == 0 {
		difficulty = 4
	} else {
		dt := time.Since(prevTimestamp).Seconds()
		if dt < targetBlockSeconds/2 && difficulty < maxDifficulty {
			difficulty++
		} else if dt > targetBlockSeconds*2 && difficulty > minDifficulty {
			difficulty--
		}
	}

	var nonce uint64
	var hashStr string

	for {
		header := fmt.Sprintf("%d:%s:%s:%s:%d", height, prevHash, txType, string(payload), nonce)
		sum := sha256.Sum256([]byte(header))
		hashStr = hex.EncodeToString(sum[:])
		ok := true
		for i := 0; i < int(difficulty); i++ {
			if hashStr[i] != '0' {
				ok = false
				break
			}
		}
		if ok {
			break
		}
		nonce++
	}

	blk := &Block{
		Height:     height,
		PrevHash:   prevHash,
		Hash:       hashStr,
		Timestamp:  time.Now().UTC(),
		TxType:     txType,
		Tx:         txBody,
		Nonce:      nonce,
		Difficulty: difficulty,
	}
	c.blocks = append(c.blocks, blk)

	// Persist to chain log if the DB handle is present. For DevNet we log
	// errors but do not abort the in‑memory chain.
	if c.db != nil {
		ctx := context.Background()
		if err := c.db.InsertBlockAndTx(ctx, blk.Hash, blk.PrevHash, blk.Height, blk.TxType, blk.Nonce, blk.Difficulty, blk.Tx); err != nil {
			_ = err
		}
	}

	return blk
}

// MintTx captures the parameters for a mint operation.
type MintTx struct {
	Address string  `json:"address"`
	Asset   string  `json:"asset"`
	Amount  float64 `json:"amount"`
}

// RedeemTx captures the parameters for a redeem operation.
type RedeemTx struct {
	Address string  `json:"address"`
	Asset   string  `json:"asset"`
	Amount  float64 `json:"amount"`
}

// ApplyMint debits the user's asset (e.g. USDC), credits treasury with that
// asset, and mints GRC to the user. It then appends a TX_MINT block.
func (c *Chain) ApplyMint(addr, asset string, amount float64) (*Block, string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if amount <= 0 {
		return nil, "", fmt.Errorf("amount must be positive")
	}
	if asset == "" {
		asset = "USDC"
	}
	if addr == "" {
		addr = "demo-user"
	}
	if !isAllowedBackingAsset(asset) {
		return nil, "", fmt.Errorf("unsupported backing asset: %s", asset)
	}

	// Compute a simple NAV from current reserves (treasury) and total GRC
	// supply, then determine how many GRC units to mint for the given
	// deposit amount. In DevNet we treat USDC/USDT/DAI as USD-like.
	nav := c.computeDevnetNAVLocked()
	if nav <= 0 {
		return nil, "", fmt.Errorf("NAV is non-positive, cannot mint")
	}
	lower, upper := econ.ComputeCorridorBounds(1.0, 10)
	// Arbitrage-friendly: allow mint when NAV is at or below the upper
	// corridor bound. When NAV is above the corridor we block mint so
	// that supply expansion does not further weaken the peg.
	if nav > upper {
		return nil, "", fmt.Errorf("mint disabled: NAV above corridor (nav=%.6f, upper=%.6f)", nav, upper)
	}
	deposit := amount
	minted := deposit / nav

	// Move backing asset from user to treasury, mint GRC at NAV.
	if err := c.store.Debit(addr, asset, deposit); err != nil {
		return nil, "", err
	}
	c.store.Credit("treasury", asset, deposit)
	c.store.Credit(addr, "GRC", minted)

	tx := MintTx{
		Address: addr,
		Asset:   asset,
		Amount:  minted,
	}
	blk := c.appendBlockLocked("TX_MINT", tx)
	return blk, blk.Hash, nil
}

// TransferTx moves balances between addresses on-chain.
type TransferTx struct {
	From   string  `json:"from"`
	To     string  `json:"to"`
	Asset  string  `json:"asset"`
	Amount float64 `json:"amount"`
	Nonce  uint64  `json:"nonce"`
	Memo   string  `json:"memo,omitempty"`
}

// ApplyTransfer debits the sender and credits the receiver, enforcing
// per-address nonces and recording a TX_TRANSFER block.
func (c *Chain) ApplyTransfer(tx TransferTx) (*Block, string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if tx.From == "" || tx.To == "" {
		return nil, "", fmt.Errorf("missing from/to")
	}
	if tx.Asset == "" {
		tx.Asset = "GRC"
	}
	if tx.Amount <= 0 {
		return nil, "", fmt.Errorf("amount must be positive")
	}

	// Enforce per-address nonce for externally signed transfers.
	if err := c.store.ExpectAndIncrementNonce(tx.From, tx.Nonce); err != nil {
		return nil, "", err
	}

	if err := c.store.Debit(tx.From, tx.Asset, tx.Amount); err != nil {
		return nil, "", err
	}
	c.store.Credit(tx.To, tx.Asset, tx.Amount)

	blk := c.appendBlockLocked("TX_TRANSFER", tx)
	return blk, blk.Hash, nil
}

// ApplyRedeem burns GRC from the user, moves asset from treasury to user,
// and records a TX_REDEEM block.
func (c *Chain) ApplyRedeem(addr, asset string, amount float64) (*Block, string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if amount <= 0 {
		return nil, "", fmt.Errorf("amount must be positive")
	}
	// DevNet: always redeem into USDC (R3) regardless of requested asset.
	asset = "USDC"
	if addr == "" {
		addr = "demo-user"
	}

	nav := c.computeDevnetNAVLocked()
	if nav <= 0 {
		return nil, "", fmt.Errorf("NAV is non-positive, cannot redeem")
	}
	lower, upper := econ.ComputeCorridorBounds(1.0, 10)
	// Arbitrage-friendly: allow redeem when NAV is at or above the lower
	// corridor bound. When NAV is below the corridor we block redeem so
	// that redemptions do not drain reserves while GRC is trading rich.
	if nav < lower {
		return nil, "", fmt.Errorf("redeem disabled: NAV below corridor (nav=%.6f, lower=%.6f)", nav, lower)
	}
	burnGRC := amount
	payout := burnGRC * nav

	// Burn GRC from user, pay out USDC from treasury at NAV.
	if err := c.store.Debit(addr, "GRC", burnGRC); err != nil {
		return nil, "", err
	}
	if err := c.store.Debit("treasury", asset, payout); err != nil {
		return nil, "", err
	}
	c.store.Credit(addr, asset, payout)

	tx := RedeemTx{
		Address: addr,
		Asset:   asset,
		Amount:  payout,
	}
	blk := c.appendBlockLocked("TX_REDEEM", tx)
	return blk, blk.Hash, nil
}

// ApplyTierRenew debits the user's GRC (via Payment.AmountGRC) and credits
// a dedicated on-chain account to track tier revenue, then records a
// TX_TIER_RENEW block.
//
// The higher-level PHP layer handles the rich business logic (Earn usage,
// grace periods, runtime multipliers, etc.). Here we ensure the payment
// side is represented at L1.

func (c *Chain) ApplyTierRenew(tx TxTierRenew) (*Block, string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if tx.Sender == "" {
		return nil, "", fmt.Errorf("missing sender")
	}
	payAmt := tx.Payment.AmountGRC
	if payAmt <= 0 {
		return nil, "", fmt.Errorf("payment amount must be positive")
	}

	// Enforce per-address nonce for externally signed tier renewals.
	if err := c.store.ExpectAndIncrementNonce(tx.Sender, tx.Nonce); err != nil {
		return nil, "", err
	}

	// Debit from sender, credit to tier-revenue bucket.
	if err := c.store.Debit(tx.Sender, "GRC", payAmt); err != nil {
		return nil, "", err
	}
	// DevNet convention: tier revenue bucket.
	c.store.Credit("treasury-tiers", "GRC", payAmt)

	blk := c.appendBlockLocked("TX_TIER_RENEW", tx)
	return blk, blk.Hash, nil
}

// Miner produces empty heartbeat blocks when enabled. It does not currently
// own transaction selection; Apply* calls still mine blocks synchronously,
// but this gives the chain a live tip and can be extended into a full
// mempool‑driven miner later.
type Miner struct {
	chain    *Chain
	quit     chan struct{}
	interval time.Duration
	running  int32
}

func NewMiner(chain *Chain, interval time.Duration) *Miner {
	return &Miner{
		chain:    chain,
		quit:     make(chan struct{}),
		interval: interval,
	}
}

func (m *Miner) Start() {
	if !atomic.CompareAndSwapInt32(&m.running, 0, 1) {
		return
	}
	go m.loop()
}

func (m *Miner) Stop() {
	if !atomic.CompareAndSwapInt32(&m.running, 1, 0) {
		return
	}
	close(m.quit)
}

func (m *Miner) IsRunning() bool {
	return atomic.LoadInt32(&m.running) == 1
}

// txPriority returns a simple priority score for pending txs.
// Later this can incorporate explicit fee fields; for now it prefers
// user transfers and vault withdrawals over other traffic.
func txPriority(pt pendingTx) int {
	switch pt.Type {
	case "TX_TRANSFER":
		return 3
	case "TX_VAULT_WITHDRAW":
		return 3
	case "TX_VAULT_DEPOSIT", "TX_TIER_RENEW":
		return 2
	default:
		return 1
	}
}

func (m *Miner) loop() {
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			m.chain.mu.Lock()
			// If there are pending txs, pick the highest‑priority one.
			if len(m.chain.pendingTxs) > 0 {
				bestIdx := 0
				bestScore := txPriority(m.chain.pendingTxs[0])
				for i := 1; i < len(m.chain.pendingTxs); i++ {
					if score := txPriority(m.chain.pendingTxs[i]); score > bestScore {
						bestScore = score
						bestIdx = i
					}
				}
				ptx := m.chain.pendingTxs[bestIdx]
				// Remove chosen tx from slice.
				m.chain.pendingTxs = append(m.chain.pendingTxs[:bestIdx], m.chain.pendingTxs[bestIdx+1:]...)
				m.chain.appendBlockLocked(ptx.Type, ptx.Body)
				m.chain.mu.Unlock()
				continue
			}
			// Otherwise, produce an EMPTY heartbeat block to keep the tip moving.
			m.chain.appendBlockLocked("EMPTY", map[string]string{
				"note": "heartbeat",
			})
			m.chain.mu.Unlock()
		case <-m.quit:
			return
		}
	}
}