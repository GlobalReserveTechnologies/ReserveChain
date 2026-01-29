
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

// PeerSync is a simple multi-peer synchroniser that allows any node to
// pull new blocks from a set of HTTP peers. It complements the single-
// upstream ChainFollower and is a stepping stone toward full P2P gossip.
type PeerSync struct {
    Chain  *core.Chain
    DB     *store.DB
    Client *http.Client
    Peers  []string

    PollInterval time.Duration
}

func NewPeerSync(chain *core.Chain, db *store.DB, peers []string, poll time.Duration) *PeerSync {
    return &PeerSync{
        Chain:        chain,
        DB:           db,
        Client:       &http.Client{Timeout: 5 * time.Second},
        Peers:        peers,
        PollInterval: poll,
    }
}

func (p *PeerSync) Run() {
    ticker := time.NewTicker(p.PollInterval)
    defer ticker.Stop()

    for range ticker.C {
        for _, base := range p.Peers {
            if base == "" {
                continue
            }
            if err := p.syncPeer(context.Background(), base); err != nil {
                log.Printf("[peersync] sync from %s error: %v", base, err)
            }
        }
    }
}

func (p *PeerSync) syncPeer(ctx context.Context, baseURL string) error {
    // Determine local height.
    head := p.Chain.Head()
    var localHeight uint64
    if head != nil {
        localHeight = head.Height
    }

    // Ask peer for its head.
    headResp, err := p.Client.Get(baseURL + "/api/chain/head")
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

    // Pull missing blocks in batches, just like the follower does.
    next := localHeight + 1
    for next <= headPayload.Head.Height {
        limit := uint64(50)
        url := baseURL + "/api/chain/blocks?from_height=" + strconv.FormatUint(next, 10) + "&limit=" + strconv.FormatUint(limit, 10)
        resp, err := p.Client.Get(url)
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
            if p.DB != nil {
                if err := p.DB.InsertBlockAndTx(ctx, blk.Hash, blk.PrevHash, blk.Height, blk.TxType, blk.Nonce, blk.Difficulty, blk.Tx); err != nil {
                    log.Printf("[peersync] insert block %d failed: %v", blk.Height, err)
                }
            }

            // Replay into local state.
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
                    if err := p.Chain.ReplayFromTxRows([]store.ChainTxRow{row}); err != nil {
                        log.Printf("[peersync] replay block %d failed: %v", blk.Height, err)
                    }
                }
            }

            p.Chain.AppendRemoteBlock(blk)
            next = blk.Height + 1
        }

        if blkPayload.NextFromHeight <= next {
            break
        }
        next = blkPayload.NextFromHeight
    }

    return nil
}
