package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"
)

type SlashingEvent struct {
	ID            int64           `json:"id"`
	Epoch         int64           `json:"epoch"`
	SubjectType   string          `json:"subject_type"`
	SubjectID     string          `json:"subject_id"`
	Severity      string          `json:"severity"`
	Score         float64         `json:"score"`
	PenaltyFactor float64         `json:"penalty_factor"`
	ReasonCode    string          `json:"reason_code"`
	ReasonDetail  string          `json:"reason_detail"`
	Evidence      json.RawMessage `json:"evidence_json"`
	Status        string          `json:"status"`
	CreatedAt     time.Time       `json:"created_at"`
	AppliedAt     *time.Time      `json:"applied_at,omitempty"`
}

func (db *DB) InsertSlashingEvent(ctx context.Context, e SlashingEvent) error {
	if db == nil || db.sql == nil {
		return nil
	}
	if e.SubjectType == "" || e.SubjectID == "" || e.ReasonCode == "" {
		return errors.New("subject_type, subject_id, reason_code required")
	}
	if e.Severity == "" {
		e.Severity = "warn"
	}
	if e.Status == "" {
		e.Status = "pending"
	}
	if len(e.Evidence) == 0 {
		e.Evidence = json.RawMessage(`{}`)
	}
	_, err := db.sql.ExecContext(ctx, `
        INSERT INTO slashing_events
            (epoch, subject_type, subject_id, severity, score, penalty_factor, reason_code, reason_detail, evidence_json, status)
        VALUES
            (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `, e.Epoch, e.SubjectType, e.SubjectID, e.Severity, e.Score, e.PenaltyFactor, e.ReasonCode, e.ReasonDetail, string(e.Evidence), e.Status)
	return err
}

func (db *DB) ListSlashingEvents(ctx context.Context, epoch *int64, subjectType, subjectID string, status string, limit int) ([]SlashingEvent, error) {
	if db == nil || db.sql == nil {
		return nil, nil
	}
	if limit <= 0 || limit > 1000 {
		limit = 200
	}
	q := `SELECT id, epoch, subject_type, subject_id, severity, score, penalty_factor, reason_code, reason_detail, evidence_json, status, created_at, applied_at
          FROM slashing_events WHERE 1=1`
	args := []any{}
	if epoch != nil {
		q += ` AND epoch=?`
		args = append(args, *epoch)
	}
	if subjectType != "" {
		q += ` AND subject_type=?`
		args = append(args, subjectType)
	}
	if subjectID != "" {
		q += ` AND subject_id=?`
		args = append(args, subjectID)
	}
	if status != "" {
		q += ` AND status=?`
		args = append(args, status)
	}
	q += ` ORDER BY id DESC LIMIT ?`
	args = append(args, limit)

	rows, err := db.sql.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []SlashingEvent{}
	for rows.Next() {
		var e SlashingEvent
		var evidence string
		var applied sql.NullTime
		if err := rows.Scan(&e.ID, &e.Epoch, &e.SubjectType, &e.SubjectID, &e.Severity, &e.Score, &e.PenaltyFactor, &e.ReasonCode, &e.ReasonDetail, &evidence, &e.Status, &e.CreatedAt, &applied); err != nil {
			continue
		}
		e.Evidence = json.RawMessage(evidence)
		if applied.Valid {
			t := applied.Time
			e.AppliedAt = &t
		}
		out = append(out, e)
	}
	return out, nil
}

func (db *DB) MarkSlashingEventApplied(ctx context.Context, id int64) error {
	if db == nil || db.sql == nil {
		return nil
	}
	_, err := db.sql.ExecContext(ctx, `UPDATE slashing_events SET status='applied', applied_at=CURRENT_TIMESTAMP WHERE id=?`, id)
	return err
}
