package econ

import (
	"context"
	"log"
	"math"
	"time"

	"crypto/sha256"
	"encoding/json"
	"fmt"
	"reservechain/internal/core"
	"reservechain/internal/store"
	"sort"
)

// SettleEpochRewardsDevnet computes and applies (credits) the epoch's
// RSX staking rewards and PoP operator rewards into the in-memory chain
// store, and persists an auditable payout ledger into SQLite (if present).
//
// Budget units:
//   - stakeBudgetGRC and popBudgetGRC are denominated in GRC (issuance curve).
func SettleEpochRewardsDevnet(epochIndex uint64, stakeBudgetGRC, popBudgetGRC, treasuryBudgetGRC float64) {
	chain := runtimeChain()
	db := runtimeDB()
	if chain == nil {
		return
	}

	ctx := context.Background()

	// 1) RSX staking rewards (PoS-style): split stakeBudget across validators
	// by total delegated RSX, then apply validator commission and pay delegators.
	if stakeBudgetGRC > 0 {
		if err := applyStakeRewards(ctx, chain, db, int64(epochIndex), stakeBudgetGRC); err != nil {
			log.Printf("[econ] stake settle epoch=%d failed: %v", epochIndex, err)
		}
	}

	// 2) PoP rewards: compute node work score for the epoch and split popBudget across nodes.
	if popBudgetGRC > 0 {
		if err := applyPoPRewards(ctx, chain, db, int64(epochIndex), popBudgetGRC); err != nil {
			log.Printf("[econ] pop settle epoch=%d failed: %v", epochIndex, err)
		}
	}

	// 3) Treasury mint (simple): credit treasury bucket.
	if treasuryBudgetGRC > 0 {
		// Treasury address is hard-coded to "treasury" in devnet.
		chain.Store().Credit("treasury", "GRC", treasuryBudgetGRC)
		if db != nil {
			_ = db.InsertEpochPayout(ctx, store.EpochPayout{
				Epoch:     int64(epochIndex),
				Kind:      "treasury",
				Recipient: "treasury",
				AssetCode: "GRC",
				Amount:    treasuryBudgetGRC,
				Meta: map[string]any{
					"note": "devnet issuance treasury share",
				},
				CreatedAt: time.Now().UTC(),
			})

			// 4) Record an on-chain commitment to the payout ledger for auditability.
			if db != nil {
				if payoutHash, nPayouts, err := computeEpochPayoutCommit(ctx, db, int64(epochIndex)); err == nil && payoutHash != "" {
					author := "econ"
					nonce := chain.Store().GetNonce(author) + 1
					_, _, _ = chain.ApplyEpochPayoutCommit(core.EpochPayoutCommitTx{
						EpochIndex:        epochIndex,
						Author:            author,
						PayoutHashHex:     payoutHash,
						NumPayouts:        nPayouts,
						StakeBudgetGRC:    stakeBudgetGRC,
						PopBudgetGRC:      popBudgetGRC,
						TreasuryBudgetGRC: treasuryBudgetGRC,
						Nonce:             nonce,
					})
				}
			}
		}
	}
}

func applyStakeRewards(ctx context.Context, chain *core.Chain, db *store.DB, epoch int64, stakeBudgetGRC float64) error {
	if db == nil {
		return nil
	}

	validators, err := db.ListValidators(ctx)
	if err != nil {
		return err
	}
	stakes, err := db.ListStakes(ctx)
	if err != nil {
		return err
	}

	// Sum total stake and per-validator stake (ignore expired/unlocked positions for now).
	total := 0.0
	vTotal := make(map[string]float64)
	for _, s := range stakes {
		if s.AmountRSX <= 0 {
			continue
		}
		// If locked_until_epoch is set and this epoch is after lock, stake is still valid.
		// (Unlocking is explicit; lock is a minimum.)
		total += s.AmountRSX
		vTotal[s.ValidatorID] += s.AmountRSX
	}
	if total <= 0 {
		return nil
	}

	// Build a validator lookup for commission and operator wallet.
	vInfo := make(map[string]store.Validator)
	for _, v := range validators {
		vInfo[v.ValidatorID] = v
	}

	// For each validator, distribute budget.
	for vid, st := range vTotal {
		if st <= 0 {
			continue
		}
		share := st / total
		if share < 0 {
			share = 0
		}
		vBudget := stakeBudgetGRC * share
		if vBudget <= 0 {
			continue
		}

		v := vInfo[vid]
		commissionBps := v.CommissionBps
		if commissionBps < 0 {
			commissionBps = 0
		}
		if commissionBps > 5000 {
			commissionBps = 5000
		}

		commission := vBudget * (float64(commissionBps) / 10000.0)
		if commission < 0 {
			commission = 0
		}
		if commission > vBudget {
			commission = vBudget
		}
		remainder := vBudget - commission

		// Pay commission to operator wallet if present.
		if v.OperatorWallet != "" && commission > 0 {
			chain.Store().Credit(v.OperatorWallet, "GRC", commission)
			_ = db.InsertEpochPayout(ctx, store.EpochPayout{
				Epoch: epoch, Kind: "stake", Recipient: v.OperatorWallet, AssetCode: "GRC", Amount: commission,
				Meta:      map[string]any{"validator_id": vid, "role": "commission", "commission_bps": commissionBps},
				CreatedAt: time.Now().UTC(),
			})
		}

		// Pay delegators (including validator self, if they stake via same wallet) proportional to RSX.
		if remainder <= 0 {
			continue
		}
		// compute delegator total for this validator
		dTotal := 0.0
		for _, s := range stakes {
			if s.ValidatorID == vid && s.AmountRSX > 0 {
				dTotal += s.AmountRSX
			}
		}
		if dTotal <= 0 {
			continue
		}

		for _, s := range stakes {
			if s.ValidatorID != vid || s.AmountRSX <= 0 {
				continue
			}
			dShare := s.AmountRSX / dTotal
			amt := remainder * dShare
			// avoid dust blowups
			if math.Abs(amt) < 1e-12 {
				continue
			}
			chain.Store().Credit(s.StakerWallet, "GRC", amt)
			_ = db.InsertEpochPayout(ctx, store.EpochPayout{
				Epoch: epoch, Kind: "stake", Recipient: s.StakerWallet, AssetCode: "GRC", Amount: amt,
				Meta:      map[string]any{"validator_id": vid, "role": "delegator", "staked_rsx": s.AmountRSX},
				CreatedAt: time.Now().UTC(),
			})
		}
	}
	return nil
}

func applyPoPRewards(ctx context.Context, chain *core.Chain, db *store.DB, epoch int64, popBudgetGRC float64) error {
	if db == nil {
		return nil
	}
	// Pull all metrics for epoch.
	metrics, err := db.ListPoPMetricsForEpoch(ctx, epoch)
	if err != nil {
		return err
	}
	if len(metrics) == 0 {
		return nil
	}

	// Build per-node aggregates for the epoch. If multiple rows exist for the node,
	// we sum the additive metrics and average the 0..1 scores.
	type agg struct {
		count   int
		uptime  float64
		latency float64
		req     float64
		relayed float64
		storage float64
	}
	aggs := map[string]*agg{}
	for _, m := range metrics {
		a := aggs[m.NodeID]
		if a == nil {
			a = &agg{}
			aggs[m.NodeID] = a
		}
		a.count++
		a.uptime += m.UptimeScore
		a.latency += m.LatencyScore
		a.req += m.RequestsServed
		a.relayed += m.BlocksRelayed
		a.storage += m.StorageIO
	}

	// Find maxima for normalization of additive counters.
	maxReq, maxRelayed, maxStorage := 0.0, 0.0, 0.0
	for _, a := range aggs {
		if a.req > maxReq {
			maxReq = a.req
		}
		if a.relayed > maxRelayed {
			maxRelayed = a.relayed
		}
		if a.storage > maxStorage {
			maxStorage = a.storage
		}
	}
	if maxReq <= 0 {
		maxReq = 1
	}
	if maxRelayed <= 0 {
		maxRelayed = 1
	}
	if maxStorage <= 0 {
		maxStorage = 1
	}

	// Build core snapshots.
	snaps := make([]core.NodeWorkSnapshot, 0, len(aggs))
	for nodeID, a := range aggs {
		// cap derived from capability profile
		cap := 1.0
		if c, err := db.GetPoPCapability(ctx, nodeID); err == nil {
			cap = (clamp01(c.CPUScore) + clamp01(c.RAMScore) + clamp01(c.StorageScore) + clamp01(c.BandwidthScore)) / 4.0
			if cap <= 0 {
				cap = 0.25
			}
		}

		uptime := clamp01(a.uptime / float64(max(1, a.count)))
		latency := clamp01(a.latency / float64(max(1, a.count)))

		// Map to work components.
		consensus := uptime
		network := clamp01(0.5*(a.relayed/maxRelayed) + 0.5*latency)
		storage := clamp01(a.storage / maxStorage)
		service := clamp01(a.req / maxReq)

		snaps = append(snaps, core.NodeWorkSnapshot{
			NodeID:      nodeID,
			EpochStart:  0,
			EpochEnd:    0,
			Consensus:   consensus,
			Network:     network,
			Storage:     storage,
			Service:     service,
			HardwareCap: cap,
		})
	}

	// Use default weights for now (can be governed later).
	weights := core.WorkWeights{Consensus: 0.45, Network: 0.25, Storage: 0.15, Service: 0.15}
	res := core.ComputeOperatorPayouts(weights, snaps, popBudgetGRC)

	// Apply payouts: credit operator wallet (node owner).
	for _, n := range res.Nodes {
		if n.RewardGRC <= 0 {
			continue
		}
		node, err := db.GetPoPNode(ctx, n.NodeID)
		if err != nil || node.OperatorWallet == "" {
			continue
		}
		// Conservative slashing/anomaly detection (soft slashing via reward haircut).
		cfg := DefaultSlashingConfig()
		an := DetectPoPAnomalies(ctx, db, cfg, epoch, n.NodeID)
		if an.ReasonCode != "" {
			RecordPoPSlashingEvent(ctx, db, epoch, n.NodeID, an)
		}
		penalty := clamp01(an.PenaltyFactor)
		reward := n.RewardGRC * (1.0 - penalty)
		slashed := n.RewardGRC - reward
		if reward <= 0 {
			reward = 0
		}
		if reward > 0 {
			chain.Store().Credit(node.OperatorWallet, "GRC", reward)
		}
		if slashed > 0 {
			chain.Store().Credit("treasury", "GRC", slashed)
		}
		_ = db.InsertEpochPayout(ctx, store.EpochPayout{
			Epoch: epoch, Kind: "pop", Recipient: node.OperatorWallet, AssetCode: "GRC", Amount: reward,
			Meta: map[string]any{
				"node_id":           n.NodeID,
				"work_raw":          n.WorkRaw,
				"work_final":        n.WorkFinal,
				"cap":               n.HardwareCap,
				"penalty":           penalty,
				"slash_to_treasury": slashed,
				"anomaly_code":      an.ReasonCode,
			},
			CreatedAt: time.Now().UTC(),
		})
	}
	return nil
}
func clamp01(x float64) float64 {
	if x < 0 {
		return 0
	}
	if x > 1 {
		return 1
	}
	return x
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
