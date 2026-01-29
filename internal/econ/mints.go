
package econ

// DevNet mint queue + epoch scaffolding.
//
// This file introduces a simple, in-memory model for USDR mint requests
// and epoch-based settlement. There are two DevNet mint types:
//   - crypto-backed USDR mints (asset deposit -> USDR liability)
//   - free test USDR mints (no backing, used to stress coverage)
// GRC issuance will be layered on top of this in a later iteration.

import (
    "sync"
    "time"
)

type MintType string

const (
    MintTypeUSDRFromCrypto MintType = "usdr_from_crypto"
    MintTypeUSDRTest       MintType = "usdr_test"
    // MintTypeGRC can be added later for base money issuance tied to
    // the issuance curve and policy engine.
)

type MintStatus string

const (
    MintPending  MintStatus = "pending"
    MintSettled  MintStatus = "settled"
    MintRejected MintStatus = "rejected"
)

// DevnetMintRequest represents a single DevNet mint operation that will
// be considered at the next epoch boundary. For crypto-backed USDR mints
// both the asset quantity and the USDR equivalent are tracked so that
// treasury reserve pools can be updated consistently.
type DevnetMintRequest struct {
    ID          string     `json:"id"`
    CreatedAt   time.Time  `json:"created_at"`
    Epoch       int64      `json:"epoch"`
    AccountRef  string     `json:"account_ref"`
    MintType    MintType   `json:"mint_type"`
    Asset       string     `json:"asset"`         // e.g. "USDC", "ETH"
    AmountAsset float64    `json:"amount_asset"`  // asset units for crypto-backed mints
    AmountUSDR  float64    `json:"amount_usdr"`   // USDR units to be minted
    Status      MintStatus `json:"status"`
}

// DevnetMintSnapshot is a read-only view suitable for RPC / UI.
type DevnetMintSnapshot struct {
    CurrentEpoch int64              `json:"current_epoch"`
    Pending      []DevnetMintRequest `json:"pending"`
    TotalUSDR    float64            `json:"total_usdr"`
}

var (
    mintMu          sync.RWMutex
    devnetMintsPend = make([]DevnetMintRequest, 0)
)

// EnqueueDevnetCryptoMintUSDR registers a crypto-backed USDR mint
// request. The USDR equivalent is computed using the DevNet price map
// at enqueue time for simplicity.
func EnqueueDevnetCryptoMintUSDR(accountRef string, asset CryptoAssetKind, amountAsset float64) DevnetMintRequest {
    mintMu.Lock()
    defer mintMu.Unlock()

    if amountAsset <= 0 {
        amountAsset = 0
    }

    prices := getDevnetPriceMap()
    px, ok := prices[asset]
    if !ok {
        px = 0
    }
    usdr := amountAsset * px

    req := DevnetMintRequest{
        ID:          makeSimpleID("mnt"),
        CreatedAt:   time.Now().UTC(),
        Epoch:       devnetCurrentEpoch,
        AccountRef:  accountRef,
        MintType:    MintTypeUSDRFromCrypto,
        Asset:       string(asset),
        AmountAsset: amountAsset,
        AmountUSDR:  usdr,
        Status:      MintPending,
    }

    devnetMintsPend = append(devnetMintsPend, req)
    return req
}

// EnqueueDevnetTestMintUSDR registers a free, unbacked USDR mint
// request. This is DevNet-only and is useful to stress coverage and
// observe the system's response to over-issuance.
func EnqueueDevnetTestMintUSDR(accountRef string, amountUSDR float64) DevnetMintRequest {
    mintMu.Lock()
    defer mintMu.Unlock()

    if amountUSDR <= 0 {
        amountUSDR = 0
    }

    req := DevnetMintRequest{
        ID:          makeSimpleID("mnt"),
        CreatedAt:   time.Now().UTC(),
        Epoch:       devnetCurrentEpoch,
        AccountRef:  accountRef,
        MintType:    MintTypeUSDRTest,
        Asset:       "",
        AmountAsset: 0,
        AmountUSDR:  amountUSDR,
        Status:      MintPending,
    }

    devnetMintsPend = append(devnetMintsPend, req)
    return req
}

// SnapshotDevnetMints returns a read-only view of the current mint
// queue state.
func SnapshotDevnetMints() DevnetMintSnapshot {
    mintMu.RLock()
    defer mintMu.RUnlock()

    pendCopy := make([]DevnetMintRequest, len(devnetMintsPend))
    copy(pendCopy, devnetMintsPend)

    var total float64
    for _, m := range devnetMintsPend {
        total += m.AmountUSDR
    }

    return DevnetMintSnapshot{
        CurrentEpoch: devnetCurrentEpoch,
        Pending:      pendCopy,
        TotalUSDR:    total,
    }
}

// settleDevnetMintsForEpoch is called from AdvanceDevnetEpoch to apply
// all pending mint requests for the current epoch to the treasury
// balance sheet. For now this is intentionally conservative and does
// not perform corridor or coverage gating; that logic will be added in
// a later iteration.
func settleDevnetMintsForEpoch(currentEpoch int64) {
    mintMu.Lock()
    defer mintMu.Unlock()

    if len(devnetMintsPend) == 0 {
        return
    }

    prices := getDevnetPriceMap()

    treasuryMu.Lock()
    defer treasuryMu.Unlock()

    // Ensure there is at least one synthetic pool we can credit.
    if len(treasuryPools) == 0 {
        treasuryPools = []ReservePoolSnapshot{
            {
                PoolID: ReservePoolID("devnet-main"),
                At:     time.Now().UTC(),
                Balances: []ReservePoolBalance{},
            },
        }
    }

    // Helper to credit an asset into the first pool.
    ensureBalance := func(asset CryptoAssetKind, delta float64) {
        if delta == 0 {
            return
        }
        pool := &treasuryPools[0]
        found := false
        for i := range pool.Balances {
            if pool.Balances[i].Asset == asset {
                pool.Balances[i].Amount += delta
                if pool.Balances[i].Amount < 0 {
                    pool.Balances[i].Amount = 0
                }
                found = true
                break
            }
        }
        if !found {
            pool.Balances = append(pool.Balances, ReservePoolBalance{
                Asset:  asset,
                Amount: delta,
            })
        }
        pool.At = time.Now().UTC()
    }

    // Process mints.
    next := devnetMintsPend[:0]
    for _, m := range devnetMintsPend {
        if m.Status != MintPending || m.Epoch != currentEpoch {
            next = append(next, m)
            continue
        }

        switch m.MintType {
        case MintTypeUSDRFromCrypto:
            // Credit the backing asset into the pool and increase USDR supply.
            var assetKind CryptoAssetKind
            switch m.Asset {
            case string(AssetUSDC):
                assetKind = AssetUSDC
            case string(AssetUSDT):
                assetKind = AssetUSDT
            case string(AssetDAI):
                assetKind = AssetDAI
            case string(AssetETH):
                assetKind = AssetETH
            case string(AssetWBTC):
                assetKind = AssetWBTC
            default:
                assetKind = ""
            }

            if assetKind != "" && m.AmountAsset > 0 {
                ensureBalance(assetKind, m.AmountAsset)
            }

            // Increase USDR supply by the requested amount.
            treasuryUSDRSupply += m.AmountUSDR

            m.Status = MintSettled

        case MintTypeUSDRTest:
            // DevNet test-only mint: no backing, just increase supply.
            treasuryUSDRSupply += m.AmountUSDR
            m.Status = MintSettled

        default:
            m.Status = MintRejected
        }

        // Settled / rejected entries are dropped from the pending slice.
    }

    devnetMintsPend = next
}
