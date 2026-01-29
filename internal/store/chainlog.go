package store

import (
    "context"
    "database/sql"
    "encoding/json"
    "log"
    "time"
)

// ChainBlockRow mirrors the chain_blocks table.
type ChainBlockRow struct {
    Height     uint64
    Hash       string
    PrevHash   *string
    Timestamp  time.Time
    TxType     string
    Nonce      uint64
    Difficulty uint32
}

// ChainTxRow mirrors the chain_tx table.
type ChainTxRow struct {
    ID          int64
    BlockHeight uint64
    TxHash      string
    TxType      string
    BodyJSON    string
}

// InsertBlockAndTx persists a single block + tx into the chain log tables.
// txBody is the Go struct which will be JSON-encoded into body_json.
func (db *DB) InsertBlockAndTx(ctx context.Context, blkHash string, prevHash string, height uint64, txType string, nonce uint64, difficulty uint32, txBody interface{}) error {
    if db == nil || db.sql == nil {
        return nil
    }

    payload, err := json.Marshal(txBody)
    if err != nil {
        return err
    }

    // Derive tx_hash as the same as block hash for the single-tx-per-block model.
    txHash := blkHash

    sqlTx, err := db.sql.BeginTx(ctx, &sql.TxOptions{})
    if err != nil {
        return err
    }
    defer func() {
        if err != nil {
            _ = sqlTx.Rollback()
        } else {
            _ = sqlTx.Commit()
        }
    }()

    var prevHashPtr *string
    if prevHash != "" {
        prevHashPtr = &prevHash
    }

    _, err = sqlTx.ExecContext(ctx,
        `INSERT OR REPLACE INTO chain_blocks (height, hash, prev_hash, timestamp, tx_type, nonce, difficulty)
         VALUES (?, ?, ?, ?, ?, ?, ?)`,
        height,
        blkHash,
        prevHashPtr,
        time.Now().UTC().Format(time.RFC3339),
        txType,
        nonce,
        difficulty,
    )
    if err != nil {
        log.Printf("[store] insert chain_blocks failed: %v", err)
        return err
    }

    res, err := sqlTx.ExecContext(ctx,
        `INSERT INTO chain_tx (block_height, tx_hash, tx_type, body_json)
         VALUES (?, ?, ?, ?)`,
        height,
        txHash,
        txType,
        string(payload),
    )
    if err != nil {
        log.Printf("[store] insert chain_tx failed: %v", err)
        return err
    }
    txID, _ := res.LastInsertId()

    // Optionally route into typed tables based on txType later.
    _ = txID

    return nil
}

// LoadAllBlocks loads all chain_blocks + chain_tx rows ordered by height and
// returns them as raw JSON payloads. The caller is responsible for decoding
// body_json into concrete tx structs and applying them to the in-memory state.
func (db *DB) LoadAllBlocks(ctx context.Context) ([]ChainBlockRow, []ChainTxRow, error) {
    if db == nil || db.sql == nil {
        return nil, nil, nil
    }

    rows, err := db.sql.QueryContext(ctx,
        `SELECT height, hash, prev_hash, timestamp, tx_type, nonce, difficulty
         FROM chain_blocks
         ORDER BY height ASC`)
    if err != nil {
        return nil, nil, err
    }
    defer rows.Close()

    blocks := make([]ChainBlockRow, 0, 1024)
    for rows.Next() {
        var r ChainBlockRow
        var ts string
        if err := rows.Scan(&r.Height, &r.Hash, &r.PrevHash, &ts, &r.TxType, &r.Nonce, &r.Difficulty); err != nil {
            return nil, nil, err
        }
        // Parse timestamp but don't fail hard if it is malformed.
        if t, perr := time.Parse(time.RFC3339, ts); perr == nil {
            r.Timestamp = t
        } else {
            r.Timestamp = time.Now().UTC()
        }
        blocks = append(blocks, r)
    }
    if err := rows.Err(); err != nil {
        return nil, nil, err
    }

    txRows, err := db.loadAllTx(ctx)
    if err != nil {
        return nil, nil, err
    }
    return blocks, txRows, nil
}

func (db *DB) loadAllTx(ctx context.Context) ([]ChainTxRow, error) {
    rows, err := db.sql.QueryContext(ctx,
        `SELECT id, block_height, tx_hash, tx_type, body_json
         FROM chain_tx
         ORDER BY block_height ASC, id ASC`)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    out := make([]ChainTxRow, 0, 4096)
    for rows.Next() {
        var r ChainTxRow
        if err := rows.Scan(&r.ID, &r.BlockHeight, &r.TxHash, &r.TxType, &r.BodyJSON); err != nil {
            return nil, err
        }
        out = append(out, r)
    }
    return out, rows.Err()
}
