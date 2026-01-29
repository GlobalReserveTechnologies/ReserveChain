package econ

import (
	"context"
	"encoding/json"
	"math"
	"sort"

	"reservechain/internal/store"
)

// SlashingConfig is intentionally conservative.
// It aims to minimize false positives by requiring:
//   - multi-signal anomaly agreement, and/or
//   - persistence across epochs, and/or
//   - provable contradictions (mismatch / impossible values).
type SlashingConfig struct {
	// Work cap multiplier: if claimed work is above cap * multiplier for >= 2 epochs, apply a penalty.
	CapMultiplier float64

	// Minimum epochs of corroboration required for "soft slashing" based on anomalies.
	CorroborationEpochs int

	// Penalty factors (0..1). These are applied to rewards for the epoch (soft slashing).
	PenaltySuspect float64
	PenaltySevere  float64
}

func DefaultSlashingConfig() SlashingConfig {
	return SlashingConfig{
		CapMultiplier:       5.0,  // very high (conservative)
		CorroborationEpochs: 2,    // require persistence
		PenaltySuspect:      0.15, // mild reward haircut
		PenaltySevere:       1.00, // zero rewards for provable fraud
	}
}

type PoPAnomalyResult struct {
	PenaltyFactor float64
	Score         float64
	ReasonCode    string
	ReasonDetail  string
	Evidence      map[string]any
}

// DetectPoPAnomalies computes a conservative penalty factor for a PoP node for an epoch.
// It never burns stake; it only reduces rewards for the epoch when evidence is strong.
func DetectPoPAnomalies(ctx context.Context, db *store.DB, cfg SlashingConfig, epoch int64, nodeID string) PoPAnomalyResult {
	res := PoPAnomalyResult{PenaltyFactor: 0, Score: 0, ReasonCode: "", ReasonDetail: "", Evidence: map[string]any{}}
	if db == nil {
		return res
	}

	// Gather this epoch's metrics for the node (may be multiple claims).
	metrics, _ := db.ListPoPMetricsForEpoch(ctx, epoch)
	nodeRows := []store.PoPMetrics{}
	for _, m := range metrics {
		if m.NodeID == nodeID {
			nodeRows = append(nodeRows, m)
		}
	}
	if len(nodeRows) == 0 {
		return res
	}

	// Aggregate metrics conservatively: use medians to avoid a single outlier dominating.
	sort.Slice(nodeRows, func(i, j int) bool { return nodeRows[i].RequestsServed < nodeRows[j].RequestsServed })
	median := func(get func(store.PoPMetrics) float64) float64 {
		vals := make([]float64, 0, len(nodeRows))
		for _, r := range nodeRows {
			v := get(r)
			if math.IsNaN(v) || math.IsInf(v, 0) {
				continue
			}
			vals = append(vals, v)
		}
		if len(vals) == 0 {
			return 0
		}
		sort.Float64s(vals)
		mid := len(vals) / 2
		if len(vals)%2 == 1 {
			return vals[mid]
		}
		return (vals[mid-1] + vals[mid]) / 2
	}

	uptime := median(func(m store.PoPMetrics) float64 { return m.UptimeScore })
	latency := median(func(m store.PoPMetrics) float64 { return m.LatencyScore })
	req := median(func(m store.PoPMetrics) float64 { return float64(m.RequestsServed) })
	relayed := median(func(m store.PoPMetrics) float64 { return float64(m.BlocksRelayed) })
	io := median(func(m store.PoPMetrics) float64 { return float64(m.StorageIO) })

	// Provable invalid values -> severe penalty (zero payout) but do not touch stake.
	if uptime < 0 || latency < 0 || req < 0 || relayed < 0 || io < 0 {
		res.PenaltyFactor = cfg.PenaltySevere
		res.Score = 1.0
		res.ReasonCode = "POP_INVALID_VALUES"
		res.ReasonDetail = "metrics contained negative values"
		res.Evidence["uptime"] = uptime
		res.Evidence["latency"] = latency
		res.Evidence["requests_served"] = req
		res.Evidence["blocks_relayed"] = relayed
		res.Evidence["storage_io"] = io
		return res
	}

	// Cap-based anomaly (requires corroboration across epochs).
	cap, err := db.GetPoPCapability(ctx, nodeID)
	if err == nil {
		capTotal := cap.CPUScore + cap.RAMScore + cap.StorageScore + cap.BandwidthScore
		// Heuristic maximum "work" budget this epoch.
		// This is intentionally generous to reduce false positives.
		maxWork := math.Max(1.0, capTotal) * 1_000_000.0
		work := req + (relayed * 50.0) + (io * 5.0)

		res.Evidence["cap_total"] = capTotal
		res.Evidence["work"] = work
		res.Evidence["max_work"] = maxWork

		if work > (maxWork * cfg.CapMultiplier) {
			// Check previous epochs for corroboration.
			// We treat "excess work" as suspicious only if persistent.
			priorCount := 0
			for e := epoch - 1; e >= 0 && e >= epoch-int64(cfg.CorroborationEpochs); e-- {
				prevMetrics, _ := db.ListPoPMetricsForEpoch(ctx, e)
				prevRows := []store.PoPMetrics{}
				for _, pm := range prevMetrics {
					if pm.NodeID == nodeID {
						prevRows = append(prevRows, pm)
					}
				}
				if len(prevRows) == 0 {
					continue
				}
				// simple median again
				sort.Slice(prevRows, func(i, j int) bool { return prevRows[i].RequestsServed < prevRows[j].RequestsServed })
				preq := float64(prevRows[len(prevRows)/2].RequestsServed)
				prelay := float64(prevRows[len(prevRows)/2].BlocksRelayed)
				pio := float64(prevRows[len(prevRows)/2].StorageIO)
				pwork := preq + (prelay * 50.0) + (pio * 5.0)
				if pwork > (maxWork * cfg.CapMultiplier) {
					priorCount++
				}
			}
			if priorCount >= cfg.CorroborationEpochs-1 {
				res.PenaltyFactor = cfg.PenaltySuspect
				res.Score = 0.7
				res.ReasonCode = "POP_WORK_EXCEEDS_CAP"
				res.ReasonDetail = "work exceeded conservative cap multiple across epochs"
				res.Evidence["corroboration_epochs"] = priorCount + 1
				res.Evidence["uptime"] = uptime
				res.Evidence["latency"] = latency
				return res
			}
			// Not corroborated: only warn (no penalty).
			res.PenaltyFactor = 0
			res.Score = 0.35
			res.ReasonCode = "POP_WORK_SPIKE_UNCONFIRMED"
			res.ReasonDetail = "work exceeded cap multiple but was not corroborated across epochs"
			res.Evidence["corroboration_epochs"] = priorCount + 1
			res.Evidence["uptime"] = uptime
			res.Evidence["latency"] = latency
			return res
		}
	}

	// Multi-signal anomaly: very low uptime + extremely high work (fabrication risk).
	if uptime < 0.20 && req > 2_000_000 {
		res.PenaltyFactor = cfg.PenaltySuspect
		res.Score = 0.6
		res.ReasonCode = "POP_UPTIME_WORK_CONTRADICTION"
		res.ReasonDetail = "high claimed work with very low uptime"
		res.Evidence["uptime"] = uptime
		res.Evidence["requests_served"] = req
		return res
	}

	return res
}

func RecordPoPSlashingEvent(ctx context.Context, db *store.DB, epoch int64, nodeID string, r PoPAnomalyResult) {
	if db == nil || r.ReasonCode == "" {
		return
	}
	evJSON, _ := json.Marshal(r.Evidence)
	sev := "warn"
	status := "pending"
	if r.PenaltyFactor > 0 {
		sev = "penalty"
		status = "applied"
	}
	if r.PenaltyFactor >= 1.0 {
		sev = "critical"
		status = "applied"
	}
	_ = db.InsertSlashingEvent(ctx, store.SlashingEvent{
		Epoch:         epoch,
		SubjectType:   "pop_node",
		SubjectID:     nodeID,
		Severity:      sev,
		Score:         r.Score,
		PenaltyFactor: r.PenaltyFactor,
		ReasonCode:    r.ReasonCode,
		ReasonDetail:  r.ReasonDetail,
		Evidence:      evJSON,
		Status:        status,
	})
}
