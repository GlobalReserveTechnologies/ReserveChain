
package net

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"reservechain/internal/core"
	"reservechain/internal/store"
)

// ChainFollower is a very small HTTP-based follower that can be used by
// secondary nodes to mirror the block log from a primary node. For now it
// focuses on bringing the block headers + DB log up to date; full state
// replay continues to be driven by core.NewChain on startup.
type ChainFollower struct {
	Chain        *core.Chain
	DB           *store.DB
	Client       *http.Client
	BaseURL      string
	PollInterval time.Duration
}

func NewChainFollower(chain *core.Chain, db *store.DB, baseURL string, poll time.Duration) *ChainFollower {
	return &ChainFollower{
		Chain:        chain,
		DB:           db,
		Client:       &http.Client{Timeout: 5 * time.Second},
		BaseURL:      baseURL,
		PollInterval: poll,
	}
}

func (f *ChainFollower) Run() {
	ticker := time.NewTicker(f.PollInterval)
	defer ticker.Stop()

	for range ticker.C {
		if err := f.syncOnce(context.Background()); err != nil {
			log.Printf("[follower] sync error: %v", err)
		}
	}
}

func (f *ChainFollower) syncOnce(ctx context.Context) error {
	// Determine local height.
	head := f.Chain.Head()
	var localHeight uint64
	if head != nil {
		localHeight = head.Height
	}

	// Ask upstream for its head.
	headResp, err := f.Client.Get(f.BaseURL + "/api/chain/head")
	if err != nil {
		return err
	}
	defer headResp.Body.Close()
	if headResp.StatusCode != http.StatusOK {
		return nil
	}
	var headPayload struct {
		Head *core.Block `json:"head"`
	}
	if err := json.NewDecoder(headResp.Body).Decode(&headPayload); err != nil {
		return err
	}
	if headPayload.Head == nil {
		return nil
	}
	if headPayload.Head.Height <= localHeight {
		return nil
	}

	// Pull missing blocks in batches.
	next := localHeight + 1
	for next <= headPayload.Head.Height {
		limit := uint64(50)
		url := f.BaseURL + "/api/chain/blocks?from_height=" + strconv.FormatUint(next, 10) + "&limit=" + strconv.FormatUint(limit, 10)
		resp, err := f.Client.Get(url)
		if err != nil {
			return err
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			break
		}
		var blkPayload struct {
			Blocks         []*core.Block `json:"blocks"`
			NextFromHeight uint64        `json:"next_from_height"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&blkPayload); err != nil {
			resp.Body.Close()
			return err
		}
		resp.Body.Close()

		if len(blkPayload.Blocks) == 0 {
			break
		}
		for _, blk := range blkPayload.Blocks {
			// Persist to local DB if available.
			if f.DB != nil {
				if err := f.DB.InsertBlockAndTx(ctx, blk.Hash, blk.PrevHash, blk.Height, blk.TxType, blk.Nonce, blk.Difficulty, blk.Tx); err != nil {
					log.Printf("[follower] insert block %d failed: %v", blk.Height, err)
				}
			}

			// Incrementally replay this block's transaction into local state.
			if blk.Tx != nil {
				bodyJSON, err := json.Marshal(blk.Tx)
				if err == nil {
					row := store.ChainTxRow{
						Height:    blk.Height,
						Hash:      blk.Hash,
						BlockHash: blk.Hash,
						TxType:    blk.TxType,
						BodyJSON:  string(bodyJSON),
						Timestamp: blk.Timestamp,
					}
					if err := f.Chain.ReplayFromTxRows([]store.ChainTxRow{row}); err != nil {
						log.Printf("[follower] replay block %d failed: %v", blk.Height, err)
					}
				}
			}

			// Append to in-memory chain header list.
			f.Chain.AppendRemoteBlock(blk)
			next = blk.Height + 1
		}

		if blkPayload.NextFromHeight <= next {
			break
		}
		next = blkPayload.NextFromHeight
	}
	return nil
}
