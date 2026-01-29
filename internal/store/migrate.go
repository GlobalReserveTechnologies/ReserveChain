package store

import (
	"bufio"
	"database/sql"
	"errors"
	"io"
	"os"
	"strings"
)

// EnsureSchemaFromFile applies the SQL schema file if it looks like the DB is uninitialized.
// This is a DevNet convenience to reduce manual setup friction.
//
// It checks for existence of a core table ("wallets"). If missing, it will execute the schema.
func EnsureSchemaFromFile(db *DB, schemaPath string) error {
	if db == nil || db.sql == nil {
		return nil
	}

	has, err := hasTable(db.sql, "wallets")
	if err != nil {
		return err
	}
	if has {
		return nil
	}

	f, err := os.Open(schemaPath)
	if err != nil {
		return err
	}
	defer f.Close()

	stmts, err := splitSQLStatements(f)
	if err != nil {
		return err
	}
	if len(stmts) == 0 {
		return errors.New("schema file contained no statements")
	}

	tx, err := db.sql.Begin()
	if err != nil {
		return err
	}
	for _, s := range stmts {
		if strings.TrimSpace(s) == "" {
			continue
		}
		if _, err := tx.Exec(s); err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func hasTable(db *sql.DB, table string) (bool, error) {
	row := db.QueryRow(`SELECT 1 FROM sqlite_master WHERE type='table' AND name=? LIMIT 1`, table)
	var one int
	err := row.Scan(&one)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// splitSQLStatements is a small helper that:
// - strips line comments starting with "--"
// - splits on ';' boundaries
func splitSQLStatements(r io.Reader) ([]string, error) {
	scanner := bufio.NewScanner(r)
	var b strings.Builder
	for scanner.Scan() {
		line := scanner.Text()
		if idx := strings.Index(line, "--"); idx >= 0 {
			line = line[:idx]
		}
		b.WriteString(line)
		b.WriteByte('\n')
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	raw := b.String()
	parts := strings.Split(raw, ";")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		s := strings.TrimSpace(p)
		if s == "" {
			continue
		}
		out = append(out, s)
	}
	return out, nil
}
