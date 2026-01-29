package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"
	"strings"
)

type Validator struct {
	ValidatorID    string `json:"validator_id"`
	OperatorWallet string `json:"operator_wallet"`
	CommissionBps  int    `json:"commission_bps"`
	Status         string `json:"status"`
}

type StakePosition struct {
	StakerWallet   string  `json:"staker_wallet"`
	ValidatorID    string  `json:"validator_id"`
	AmountRSX      float64 `json:"amount_rsx"`
	LockUntilEpoch int64   `json:"lock_until_epoch"`
}

type PoPNode struct {
	NodeID         string `json:"node_id"`
	OperatorWallet string `json:"operator_wallet"`
	Role           string `json:"role"`
}

type PoPCapability struct {
	NodeID         string  `json:"node_id"`
	CPUScore       float64 `json:"cpu_score"`
	RAMScore       float64 `json:"ram_score"`
	StorageScore   float64 `json:"storage_score"`
	BandwidthScore float64 `json:"bandwidth_score"`
}

type PoPMetrics struct {
	Epoch          int64   `json:"epoch"`
	NodeID         string  `json:"node_id"`
	UptimeScore    float64 `json:"uptime_score"`
	RequestsServed float64 `json:"requests_served"`
	BlocksRelayed  float64 `json:"blocks_relayed"`
	StorageIO      float64 `json:"storage_io"`
	LatencyScore   float64 `json:"latency_score"`
}

type EpochPayout struct {
	Epoch     int64          `json:"epoch"`
	Kind      string         `json:"kind"`
	Recipient string         `json:"recipient"`
	AssetCode string         `json:"asset_code"`
	Amount    float64        `json:"amount"`
	Meta      map[string]any `json:"meta,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
}

var ErrNotFound = errors.New("not found")

func (db *DB) exec(ctx context.Context, q string, args ...any) error {
	if db == nil || db.sql == nil {
		return nil
	}
	_, err := db.sql.ExecContext(ctx, q, args...)
	return err
}

// -------------------- Validators / staking --------------------

func (db *DB) UpsertValidator(ctx context.Context, v Validator) error {
	if db == nil || db.sql == nil {
		return nil
	}
	if v.ValidatorID == "" || v.OperatorWallet == "" {
		return errors.New("validator_id and operator_wallet required")
	}
	if v.CommissionBps < 0 {
		v.CommissionBps = 0
	}
	if v.CommissionBps > 5000 {
		v.CommissionBps = 5000
	} // 50% hard cap for safety
	if v.Status == "" {
		v.Status = "active"
	}

	_, err := db.sql.ExecContext(ctx, `
        INSERT INTO rsx_validators (validator_id, operator_wallet, commission_bps, status, updated_at)
        VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
        ON CONFLICT(validator_id) DO UPDATE SET
            operator_wallet=excluded.operator_wallet,
            commission_bps=excluded.commission_bps,
            status=excluded.status,
            updated_at=CURRENT_TIMESTAMP
    `, v.ValidatorID, v.OperatorWallet, v.CommissionBps, v.Status)
	return err
}

func (db *DB) ListValidators(ctx context.Context) ([]Validator, error) {
	if db == nil || db.sql == nil {
		return nil, nil
	}
	rows, err := db.sql.QueryContext(ctx, `SELECT validator_id, operator_wallet, commission_bps, status FROM rsx_validators`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []Validator{}
	for rows.Next() {
		var v Validator
		if err := rows.Scan(&v.ValidatorID, &v.OperatorWallet, &v.CommissionBps, &v.Status); err != nil {
			continue
		}
		out = append(out, v)
	}
	return out, nil
}

func (db *DB) UpsertStake(ctx context.Context, s StakePosition) error {
	if db == nil || db.sql == nil {
		return nil
	}
	if s.StakerWallet == "" || s.ValidatorID == "" {
		return errors.New("staker_wallet and validator_id required")
	}
	if s.AmountRSX < 0 {
		s.AmountRSX = 0
	}
	_, err := db.sql.ExecContext(ctx, `
        INSERT INTO rsx_stakes (staker_wallet, validator_id, amount_rsx, lock_until_epoch, updated_at)
        VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
        ON CONFLICT(staker_wallet, validator_id) DO UPDATE SET
            amount_rsx=excluded.amount_rsx,
            lock_until_epoch=excluded.lock_until_epoch,
            updated_at=CURRENT_TIMESTAMP
    `, s.StakerWallet, s.ValidatorID, s.AmountRSX, s.LockUntilEpoch)
	return err
}

// GetStakePosition returns the current stake position for staker+validator.
func (db *DB) GetStakePosition(ctx context.Context, stakerWallet, validatorID string) (StakePosition, error) {
	if db == nil || db.sql == nil {
		return StakePosition{}, ErrNotFound
	}
	row := db.sql.QueryRowContext(ctx, `
        SELECT staker_wallet, validator_id, amount_rsx, lock_until_epoch
        FROM rsx_stakes
        WHERE staker_wallet=? AND validator_id=?
    `, stakerWallet, validatorID)
	var p StakePosition
	if err := row.Scan(&p.StakerWallet, &p.ValidatorID, &p.AmountRSX, &p.LockUntilEpoch); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return StakePosition{}, ErrNotFound
		}
		return StakePosition{}, err
	}
	return p, nil
}

// ApplyStakeDelta adjusts stake amount by deltaRSX (positive to lock more, negative to unlock).
// If lockUntilEpoch is > 0, it will overwrite the lock_until_epoch field (used for locks).
func (db *DB) ApplyStakeDelta(ctx context.Context, stakerWallet, validatorID string, deltaRSX float64, lockUntilEpoch int64) error {
	if db == nil || db.sql == nil {
		return nil
	}
	if stakerWallet == "" || validatorID == "" {
		return errors.New("staker_wallet and validator_id required")
	}

	cur, err := db.GetStakePosition(ctx, stakerWallet, validatorID)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return err
	}
	newAmt := cur.AmountRSX + deltaRSX
	if newAmt < 0 {
		return errors.New("stake would go negative")
	}
	newLock := cur.LockUntilEpoch
	if lockUntilEpoch > 0 {
		newLock = lockUntilEpoch
	}
	return db.UpsertStake(ctx, StakePosition{
		StakerWallet:   stakerWallet,
		ValidatorID:    validatorID,
		AmountRSX:      newAmt,
		LockUntilEpoch: newLock,
	})
}

func (db *DB) GetStake(ctx context.Context, stakerWallet, validatorID string) (StakePosition, error) {
	var s StakePosition
	if db == nil || db.sql == nil {
		return s, ErrNotFound
	}
	row := db.sql.QueryRowContext(ctx, `SELECT staker_wallet, validator_id, amount_rsx, lock_until_epoch FROM rsx_stakes WHERE staker_wallet=? AND validator_id=?`, stakerWallet, validatorID)
	if err := row.Scan(&s.StakerWallet, &s.ValidatorID, &s.AmountRSX, &s.LockUntilEpoch); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return s, ErrNotFound
		}
		return s, err
	}
	return s, nil
}

func (db *DB) ListStakes(ctx context.Context) ([]StakePosition, error) {
	if db == nil || db.sql == nil {
		return nil, nil
	}
	rows, err := db.sql.QueryContext(ctx, `SELECT staker_wallet, validator_id, amount_rsx, lock_until_epoch FROM rsx_stakes`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []StakePosition{}
	for rows.Next() {
		var s StakePosition
		if err := rows.Scan(&s.StakerWallet, &s.ValidatorID, &s.AmountRSX, &s.LockUntilEpoch); err != nil {
			continue
		}
		out = append(out, s)
	}
	return out, nil
}

// -------------------- PoP nodes / metrics --------------------

func (db *DB) UpsertPoPNode(ctx context.Context, n PoPNode) error {
	if db == nil || db.sql == nil {
		return nil
	}
	if n.NodeID == "" || n.OperatorWallet == "" {
		return errors.New("node_id and operator_wallet required")
	}
	_, err := db.sql.ExecContext(ctx, `
        INSERT INTO pop_nodes (node_id, operator_wallet, role, updated_at)
        VALUES (?, ?, ?, CURRENT_TIMESTAMP)
        ON CONFLICT(node_id) DO UPDATE SET
            operator_wallet=excluded.operator_wallet,
            role=excluded.role,
            updated_at=CURRENT_TIMESTAMP
    `, n.NodeID, n.OperatorWallet, n.Role)
	return err
}

// UpsertPoPNodeWithTxHash upserts a node registration and (optionally) stores tx_hash for idempotent replay.
// If the schema lacks the tx_hash column (older DB), it falls back to UpsertPoPNode.
func (db *DB) UpsertPoPNodeWithTxHash(ctx context.Context, n PoPNode, txHash string) error {
    if db == nil || db.sql == nil {
        return nil
    }
    if n.NodeID == "" || n.OperatorWallet == "" {
        return errors.New("node_id and operator_wallet required")
    }
    if txHash != "" {
        // Try new schema first (includes tx_hash).
        _, err := db.sql.ExecContext(ctx, `
            INSERT INTO pop_nodes (node_id, operator_wallet, role, tx_hash, updated_at)
            VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
            ON CONFLICT(node_id) DO UPDATE SET
                operator_wallet=excluded.operator_wallet,
                role=excluded.role,
                tx_hash=excluded.tx_hash,
                updated_at=CURRENT_TIMESTAMP
        `, n.NodeID, n.OperatorWallet, n.Role, txHash)
        if err == nil {
            return nil
        }
        // Fall back if column missing.
        if strings.Contains(strings.ToLower(err.Error()), "no such column") {
            return db.UpsertPoPNode(ctx, n)
        }
        // If UNIQUE(tx_hash) blocks duplicate replay, treat as success.
        if strings.Contains(strings.ToLower(err.Error()), "unique") {
            return nil
        }
        return err
    }
    return db.UpsertPoPNode(ctx, n)
}


func (db *DB) UpsertPoPCapability(ctx context.Context, c PoPCapability) error {
	if db == nil || db.sql == nil {
		return nil
	}
	if c.NodeID == "" {
		return errors.New("node_id required")
	}
	_, err := db.sql.ExecContext(ctx, `
        INSERT INTO pop_node_caps (node_id, cpu_score, ram_score, storage_score, bandwidth_score, updated_at)
        VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
        ON CONFLICT(node_id) DO UPDATE SET
            cpu_score=excluded.cpu_score,
            ram_score=excluded.ram_score,
            storage_score=excluded.storage_score,
            bandwidth_score=excluded.bandwidth_score,
            updated_at=CURRENT_TIMESTAMP
    `, c.NodeID, c.CPUScore, c.RAMScore, c.StorageScore, c.BandwidthScore)
	return err
}

// UpsertPoPCapabilityWithTxHash upserts a node capability profile and (optionally) stores tx_hash for idempotent replay.
// If the schema lacks the tx_hash column (older DB), it falls back to UpsertPoPCapability.
func (db *DB) UpsertPoPCapabilityWithTxHash(ctx context.Context, c PoPCapability, txHash string) error {
    if db == nil || db.sql == nil {
        return nil
    }
    if c.NodeID == "" {
        return errors.New("node_id required")
    }
    if txHash != "" {
        _, err := db.sql.ExecContext(ctx, `
            INSERT INTO pop_node_caps (node_id, cpu_score, ram_score, storage_score, bandwidth_score, tx_hash, updated_at)
            VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
            ON CONFLICT(node_id) DO UPDATE SET
                cpu_score=excluded.cpu_score,
                ram_score=excluded.ram_score,
                storage_score=excluded.storage_score,
                bandwidth_score=excluded.bandwidth_score,
                tx_hash=excluded.tx_hash,
                updated_at=CURRENT_TIMESTAMP
        `, c.NodeID, c.CPUScore, c.RAMScore, c.StorageScore, c.BandwidthScore, txHash)
        if err == nil {
            return nil
        }
        if strings.Contains(strings.ToLower(err.Error()), "no such column") {
            return db.UpsertPoPCapability(ctx, c)
        }
        if strings.Contains(strings.ToLower(err.Error()), "unique") {
            return nil
        }
        return err
    }
    return db.UpsertPoPCapability(ctx, c)
}


func (db *DB) InsertPoPMetrics(ctx context.Context, m PoPMetrics) error {
	if db == nil || db.sql == nil {
		return nil
	}
	if m.NodeID == "" {
		return errors.New("node_id required")
	}
	_, err := db.sql.ExecContext(ctx, `
        INSERT INTO pop_epoch_metrics (epoch, node_id, uptime_score, requests_served, blocks_relayed, storage_io, latency_score, created_at)
        VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
    `, m.Epoch, m.NodeID, m.UptimeScore, m.RequestsServed, m.BlocksRelayed, m.StorageIO, m.LatencyScore)
	return err
}

func (db *DB) ListPoPMetricsForEpoch(ctx context.Context, epoch int64) ([]PoPMetrics, error) {
	if db == nil || db.sql == nil {
		return nil, nil
	}
	rows, err := db.sql.QueryContext(ctx, `SELECT epoch, node_id, uptime_score, requests_served, blocks_relayed, storage_io, latency_score FROM pop_epoch_metrics WHERE epoch=?`, epoch)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []PoPMetrics{}
	for rows.Next() {
		var m PoPMetrics
		if err := rows.Scan(&m.Epoch, &m.NodeID, &m.UptimeScore, &m.RequestsServed, &m.BlocksRelayed, &m.StorageIO, &m.LatencyScore); err != nil {
			continue
		}
		out = append(out, m)
	}
	return out, nil
}

func (db *DB) GetPoPCapability(ctx context.Context, nodeID string) (PoPCapability, error) {
	var c PoPCapability
	if db == nil || db.sql == nil {
		return c, ErrNotFound
	}
	row := db.sql.QueryRowContext(ctx, `SELECT node_id, cpu_score, ram_score, storage_score, bandwidth_score FROM pop_node_caps WHERE node_id=?`, nodeID)
	if err := row.Scan(&c.NodeID, &c.CPUScore, &c.RAMScore, &c.StorageScore, &c.BandwidthScore); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c, ErrNotFound
		}
		return c, err
	}
	return c, nil
}

func (db *DB) GetPoPNode(ctx context.Context, nodeID string) (PoPNode, error) {
	var n PoPNode
	if db == nil || db.sql == nil {
		return n, ErrNotFound
	}
	row := db.sql.QueryRowContext(ctx, `SELECT node_id, operator_wallet, role FROM pop_nodes WHERE node_id=?`, nodeID)
	if err := row.Scan(&n.NodeID, &n.OperatorWallet, &n.Role); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return n, ErrNotFound
		}
		return n, err
	}
	return n, nil
}

func (db *DB) ListPoPNodes(ctx context.Context) ([]PoPNode, error) {
	if db == nil || db.sql == nil {
		return nil, nil
	}
	rows, err := db.sql.QueryContext(ctx, `SELECT node_id, operator_wallet, role FROM pop_nodes`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []PoPNode{}
	for rows.Next() {
		var n PoPNode
		if err := rows.Scan(&n.NodeID, &n.OperatorWallet, &n.Role); err != nil {
			continue
		}
		out = append(out, n)
	}
	return out, nil
}

// -------------------- payout persistence --------------------

func (db *DB) InsertEpochPayout(ctx context.Context, p EpochPayout) error {
	if db == nil || db.sql == nil {
		return nil
	}
	metaJSON := ""
	if p.Meta != nil {
		if b, err := json.Marshal(p.Meta); err == nil {
			metaJSON = string(b)
		}
	}
	_, err := db.sql.ExecContext(ctx, `
        INSERT INTO epoch_payouts (epoch, kind, recipient, asset_code, amount, meta_json, created_at)
        VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
    `, p.Epoch, p.Kind, p.Recipient, p.AssetCode, p.Amount, metaJSON)
	return err
}

func (db *DB) ListEpochPayouts(ctx context.Context, epoch int64) ([]EpochPayout, error) {
	if db == nil || db.sql == nil {
		return nil, nil
	}
	rows, err := db.sql.QueryContext(ctx, `SELECT epoch, kind, recipient, asset_code, amount, meta_json, created_at FROM epoch_payouts WHERE epoch=? ORDER BY id ASC`, epoch)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []EpochPayout{}
	for rows.Next() {
		var p EpochPayout
		var meta string
		if err := rows.Scan(&p.Epoch, &p.Kind, &p.Recipient, &p.AssetCode, &p.Amount, &meta, &p.CreatedAt); err != nil {
			continue
		}
		if meta != "" {
			_ = json.Unmarshal([]byte(meta), &p.Meta)
		}
		out = append(out, p)
	}
	return out, nil
}

/*
Compatibility wrappers + on-chain PoP metric inserts
---------------------------------------------------

Historical handlers used UpsertPoPCap / UpsertPoPMetrics. The canonical
implementations are UpsertPoPCapability and InsertPoPMetrics. We provide
thin wrappers to keep HTTP handlers stable.

For on-chain PoP work claims, we store the chain tx hash into pop_epoch_metrics
when the column exists. Inserts use INSERT OR IGNORE to ensure idempotence
during chain replay.
*/

func (db *DB) UpsertPoPCap(ctx context.Context, c PoPCapability) error {
	return db.UpsertPoPCapability(ctx, c)
}

func (db *DB) UpsertPoPMetrics(ctx context.Context, m PoPMetrics) error {
	// allow multiple rows per node/epoch (aggregates handle it)
	return db.InsertPoPMetricsWithTxHash(ctx, m, "")
}

// InsertPoPMetricsWithTxHash inserts a PoP metric row. If txHash is provided and the
// pop_epoch_metrics table includes a tx_hash column, the insert is idempotent via UNIQUE(tx_hash).
// If the DB doesn't have that column (older schema), we fall back to the legacy insert.
func (db *DB) InsertPoPMetricsWithTxHash(ctx context.Context, m PoPMetrics, txHash string) error {
	if db == nil || db.sql == nil {
		return nil
	}
	if m.NodeID == "" {
		return errors.New("node_id required")
	}
	// Try new schema first.
	if txHash != "" {
		_, err := db.sql.ExecContext(ctx, `
            INSERT OR IGNORE INTO pop_epoch_metrics (epoch, node_id, uptime_score, requests_served, blocks_relayed, storage_io, latency_score, tx_hash, created_at)
            VALUES (?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
        `, m.Epoch, m.NodeID, m.UptimeScore, m.RequestsServed, m.BlocksRelayed, m.StorageIO, m.LatencyScore, txHash)
		if err == nil {
			return nil
		}
		// Fall through on errors (e.g. no such column).
	}
	// Legacy schema (no tx_hash).
	_, err := db.sql.ExecContext(ctx, `
        INSERT INTO pop_epoch_metrics (epoch, node_id, uptime_score, requests_served, blocks_relayed, storage_io, latency_score, created_at)
        VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
    `, m.Epoch, m.NodeID, m.UptimeScore, m.RequestsServed, m.BlocksRelayed, m.StorageIO, m.LatencyScore)
	return err
}


// EpochPayoutCommit captures an on-chain commitment to the payout ledger for an epoch.
type EpochPayoutCommit struct {
    Epoch             int64
    TxHash            string
    Author            string
    PayoutHash        string
    NumPayouts        int64
    StakeBudgetGRC    float64
    PopBudgetGRC      float64
    TreasuryBudgetGRC float64
    CreatedAt         time.Time
}

// InsertEpochPayoutCommit persists an epoch payout commitment if the schema supports it.
func (db *DB) InsertEpochPayoutCommit(ctx context.Context, c EpochPayoutCommit) error {
    if db == nil || db.sql == nil {
        return nil
    }
    // Best-effort: if table doesn't exist on older DBs, ignore.
    _, err := db.sql.ExecContext(ctx, `
        INSERT OR IGNORE INTO epoch_payout_commits
            (epoch, tx_hash, author, payout_hash, num_payouts, stake_budget_grc, pop_budget_grc, treasury_budget_grc, created_at)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
        c.Epoch,
        c.TxHash,
        c.Author,
        c.PayoutHash,
        c.NumPayouts,
        c.StakeBudgetGRC,
        c.PopBudgetGRC,
        c.TreasuryBudgetGRC,
        c.CreatedAt.UTC().Format(time.RFC3339),
    )
    if err != nil {
        // Older DB without the table: swallow error.
        return nil
    }
    return nil
}
