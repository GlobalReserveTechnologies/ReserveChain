package store

import (
	"database/sql"
	"log"
	"time"

	_ "modernc.org/sqlite"
)

// TreasurySnapshot is a store-level representation of the reserve basket snapshot
// persisted for operator analytics. This type intentionally mirrors the fields
// produced by internal/analytics.BuildTreasurySnapshot, but lives in the store
// package to avoid import cycles.

type TreasurySnapshot struct {
	TotalUSD float64        `json:"total_usd"`
	Tier1    Tier1Liquidity `json:"tier1"`
	Tier2    []Tier2Bucket  `json:"tier2"`
}

type Tier1Liquidity struct {
	CashUSD  float64 `json:"cash_usd"`
	BillsUSD float64 `json:"bills_usd,omitempty"`
}

type Tier2Bucket struct {
	Type        string  `json:"type"`
	NotionalUSD float64 `json:"notional_usd"`
	Rate        float64 `json:"rate"`
	DurationD   int     `json:"duration_days"`
}

// DB wraps an underlying *sql.DB. For DevNet, we keep it very simple:
// a single SQLite database file, with hybrid write semantics (M3).
type DB struct {
	sql *sql.DB
}

// OpenSQLite opens (or creates) a SQLite database at the given path.
func OpenSQLite(path string) (*DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	// For dev, keep timeouts lenient but not infinite.
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	return &DB{sql: db}, nil
}

// Close closes the underlying connection.
func (db *DB) Close() error {
	if db == nil || db.sql == nil {
		return nil
	}
	return db.sql.Close()
}

// InsertTreasurySnapshot writes a snapshot into reserve_snapshots.
//
// This is treated as a "structural" event in the hybrid model (M3),
// so we perform it synchronously. If the schema is missing (e.g. the
// user has not run database/schema.sql yet), we log the error and
// continue without failing the HTTP request.
func (db *DB) InsertTreasurySnapshot(ts TreasurySnapshot, at time.Time) {
	if db == nil || db.sql == nil {
		return
	}

	tx, err := db.sql.Begin()
	if err != nil {
		log.Printf("[store] begin tx for treasury snapshot failed: %v", err)
		return
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			_ = tx.Commit()
		}
	}()

	// Tier1 aggregate row
	tier1Notional := ts.Tier1.CashUSD + ts.Tier1.BillsUSD
	_, err = tx.Exec(
		`INSERT INTO reserve_snapshots (snapshot_time, total_usd, tier_code, notional_usd, notes)
         VALUES (?, ?, ?, ?, ?)`,
		at.UTC().Format(time.RFC3339),
		ts.TotalUSD,
		"T1",
		tier1Notional,
		"Tier1 high-liquidity synthetic basket",
	)
	if err != nil {
		log.Printf("[store] insert Tier1 reserve_snapshot failed: %v", err)
		return
	}

	// Tier2 buckets
	for _, b := range ts.Tier2 {
		_, err = tx.Exec(
			`INSERT INTO reserve_snapshots (snapshot_time, total_usd, tier_code, notional_usd, notes)
             VALUES (?, ?, ?, ?, ?)`,
			at.UTC().Format(time.RFC3339),
			ts.TotalUSD,
			"T2",
			b.NotionalUSD,
			b.Type,
		)
		if err != nil {
			log.Printf("[store] insert Tier2 reserve_snapshot failed: %v", err)
			return
		}
	}
}
