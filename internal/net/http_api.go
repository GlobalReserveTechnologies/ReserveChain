package net

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"reservechain/internal/analytics"
	"reservechain/internal/core"
	"reservechain/internal/econ"
	"reservechain/internal/store"
)

// -----------------------------------------------------------------------------
// DevNet Seed Registry (in-memory, TTL-based)
// -----------------------------------------------------------------------------

var (
	seedMode     bool
	seedRegistry = newSeedRegistry(5 * time.Minute)
)

// EnableSeedMode toggles whether this node behaves as a seed for simple
// HTTP-based peer discovery.
func EnableSeedMode(on bool) {
	seedMode = on
}

type SeedRegistry struct {
	mu    sync.Mutex
	peers map[string]time.Time
	ttl   time.Duration
}

func newSeedRegistry(ttl time.Duration) *SeedRegistry {
	return &SeedRegistry{
		peers: make(map[string]time.Time),
		ttl:   ttl,
	}
}

func (r *SeedRegistry) Add(addr string) {
	if addr == "" {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.peers[addr] = time.Now().UTC()
}

func (r *SeedRegistry) List() []string {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now().UTC()
	for addr, ts := range r.peers {
		if now.Sub(ts) > r.ttl {
			delete(r.peers, addr)
		}
	}

	out := make([]string, 0, len(r.peers))
	for addr := range r.peers {
		out = append(out, addr)
	}
	return out
}

// HTTPAPI bundles dependencies for HTTP handlers.
type HTTPAPI struct {
	Hub   *WSHub
	Store *core.AccountStore
	WMgr  *econ.WindowManager
	DB    *store.DB
	Chain *core.Chain
	Miner *core.Miner

	// Workstation portal static build (Vite dist)
	WorkstationDist string

	// Auth/session (in-memory for now; can be backed by DB later)
	authMu   sync.Mutex
	nonces   map[string]nonceEntry   // key: address
	sessions map[string]sessionEntry // key: session id
}

func defaultWorkstationDist() string {
	if v := os.Getenv("RESERVECHAIN_WORKSTATION_DIST"); v != "" {
		return v
	}
	// Default relative path for the built workstation portal (Vite).
	return "apps/workstation_portal/dist"
}

// NewHTTPAPI creates a new HTTPAPI instance.
func NewHTTPAPI(hub *WSHub, store *core.AccountStore, wm *econ.WindowManager, db *store.DB, chain *core.Chain, miner *core.Miner) *HTTPAPI {
	return &HTTPAPI{
		Hub:             hub,
		Store:           store,
		WMgr:            wm,
		DB:              db,
		Chain:           chain,
		Miner:           miner,
		WorkstationDist: defaultWorkstationDist(),
		nonces:          make(map[string]nonceEntry),
		sessions:        make(map[string]sessionEntry),
	}
}

// mempoolHandler exposes the current pending txs so the workstation can
// display them. For DevNet this is a simple best‑effort snapshot.
func (api *HTTPAPI) mempoolHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	type item struct {
		Type string          `json:"type"`
		Body json.RawMessage `json:"body"`
	}
	api.Chain.MuRLock()
	defer api.Chain.MuRUnlock()
	pending := api.Chain.PendingTxsSnapshot()
	out := make([]item, 0, len(pending))
	for _, pt := range pending {
		raw, _ := json.Marshal(pt.Body)
		out = append(out, item{Type: pt.Type, Body: raw})
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"pending": out,
	})
}

// balancesHandler returns a snapshot of all accounts.
func (api *HTTPAPI) balancesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	accounts := api.Store.SnapshotAll()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"accounts": accounts,
	})
}

// MintRequest describes a request to mint GRC against a backing asset (DevNet: USDC).
type MintRequest struct {
	Address string  `json:"address"`
	Asset   string  `json:"asset"`
	Amount  float64 `json:"amount"`
}

// mintHandler debits USD from the caller and credits GRC.

func (api *HTTPAPI) mintHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req MintRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if req.Address == "" {
		req.Address = "demo-user"
	}
	if req.Asset == "" {
		req.Asset = "USDC"
	}
	if req.Amount <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Delegate the balance changes + block creation to the Chain engine.
	blk, _, err := api.Chain.ApplyMint(req.Address, req.Asset, req.Amount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Emit existing mint event for frontends.
	ev := Event{
		ID:      "mint-" + time.Now().Format(time.RFC3339Nano),
		Type:    EventMint,
		Version: "v1",
		Payload: map[string]interface{}{
			"address": req.Address,
			"from":    req.Asset,
			"to":      "GRC",
			"amount":  req.Amount,
		},
		Timestamp: time.Now().UTC(),
	}
	api.Hub.Broadcast(ev)

	// Also broadcast a NewBlock event so explorers / dashboards can follow the chain.
	if blk != nil {
		bev := Event{
			ID:        fmt.Sprintf("block-%d", blk.Height),
			Type:      EventNewBlock,
			Version:   "v1",
			Payload:   blk,
			Timestamp: blk.Timestamp,
		}
		api.Hub.Broadcast(bev)
	}

	w.WriteHeader(http.StatusOK)
}

// RedeemRequest describes a request to redeem GRC back into a backing asset (DevNet: USDC).
type RedeemRequest struct {
	Address string  `json:"address"`
	Asset   string  `json:"asset"`
	Amount  float64 `json:"amount"`
}

// redeemHandler burns GRC and pays out the asset from the treasury via the Chain engine.
func (api *HTTPAPI) redeemHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req RedeemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if req.Address == "" {
		req.Address = "demo-user"
	}
	if req.Asset == "" {
		req.Asset = "USDC"
	}
	if req.Amount <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	blk, _, err := api.Chain.ApplyRedeem(req.Address, req.Asset, req.Amount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Record corridor volume in the window manager.
	api.WMgr.RecordVolume(req.Amount)

	ev := Event{
		ID:      "redeem-" + time.Now().Format(time.RFC3339Nano),
		Type:    EventRedeem,
		Version: "v1",
		Payload: map[string]interface{}{
			"address": req.Address,
			"from":    "GRC",
			"to":      req.Asset,
			"amount":  req.Amount,
		},
		Timestamp: time.Now().UTC(),
	}
	api.Hub.Broadcast(ev)

	if blk != nil {
		bev := Event{
			ID:        fmt.Sprintf("block-%d", blk.Height),
			Type:      EventNewBlock,
			Version:   "v1",
			Payload:   blk,
			Timestamp: blk.Timestamp,
		}
		api.Hub.Broadcast(bev)
	}

	w.WriteHeader(http.StatusOK)
}

// TransferRequest wraps a generic on-chain balance transfer.
type TransferRequest struct {
	Type string          `json:"type"`
	Tx   core.TransferTx `json:"tx"`
}

// transferHandler accepts TX_TRANSFER payloads and routes them into the Chain engine.
func (api *HTTPAPI) transferHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req TransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if req.Type != "TX_TRANSFER" {
		http.Error(w, "invalid type", http.StatusBadRequest)
		return
	}

	blk, hash, err := api.Chain.ApplyTransfer(req.Tx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Broadcast a Transfer event for frontends.
	ev := Event{
		ID:      "transfer-" + time.Now().Format(time.RFC3339Nano),
		Type:    EventType("Transfer"),
		Version: "v1",
		Payload: map[string]interface{}{
			"from":    req.Tx.From,
			"to":      req.Tx.To,
			"asset":   req.Tx.Asset,
			"amount":  req.Tx.Amount,
			"tx_hash": hash,
			"height":  blk.Height,
		},
		Timestamp: time.Now().UTC(),
	}
	api.Hub.Broadcast(ev)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"tx_hash": hash,
		"height":  blk.Height,
	})
}

// vaultCreateHandler records a TxVaultCreate on-chain so that L1 history
// knows about vault creation events and multi-sig parameters.
func (api *HTTPAPI) vaultCreateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Type string             `json:"type"`
		Tx   core.TxVaultCreate `json:"tx"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if req.Type != "TX_VAULT_CREATE" {
		http.Error(w, "invalid type", http.StatusBadRequest)
		return
	}

	blk, hash, err := api.Chain.ApplyVaultCreate(req.Tx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Broadcast a lightweight event for UIs that care about vault creation.
	ev := Event{
		ID:      "vault-create-" + time.Now().Format(time.RFC3339Nano),
		Type:    EventType("VaultCreate"),
		Version: "v1",
		Payload: map[string]interface{}{
			"vault_id": req.Tx.VaultID,
			"owner":    req.Tx.Owner,
			"type":     req.Tx.Type,
		},
		Timestamp: time.Now().UTC(),
	}
	api.Hub.Broadcast(ev)

	if blk != nil {
		bev := Event{
			ID:        fmt.Sprintf("block-%d", blk.Height),
			Type:      EventNewBlock,
			Version:   "v1",
			Payload:   blk,
			Timestamp: blk.Timestamp,
		}
		api.Hub.Broadcast(bev)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"tx_hash": hash,
		"height":  blk.Height,
	})
}

// redeemHandler burns GRC and credits USD back to the user.
func (api *HTTPAPI) redeemHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var req RedeemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if req.Address == "" {
		req.Address = "demo-user"
	}
	if req.Asset == "" {
		req.Asset = "USDC"
	}
	if req.Amount <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Burn GRC from user, pay out USD from treasury.
	if err := api.Store.Debit(req.Address, "GRC", req.Amount); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := api.Store.Debit("treasury", req.Asset, req.Amount); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	api.Store.Credit(req.Address, req.Asset, req.Amount)

	// Record corridor volume in the window manager.
	api.WMgr.RecordVolume(req.Amount)

	ev := Event{
		ID:      "redeem-" + time.Now().Format(time.RFC3339Nano),
		Type:    EventRedeem,
		Version: "v1",
		Payload: map[string]interface{}{
			"address": req.Address,
			"from":    "GRC",
			"to":      req.Asset,
			"amount":  req.Amount,
		},
		Timestamp: time.Now().UTC(),
	}
	api.Hub.Broadcast(ev)
	w.WriteHeader(http.StatusOK)
}

// accountNonceHandler exposes the current nonce for a given L1 address so
// frontends can construct correctly ordered transactions.
func (api *HTTPAPI) accountNonceHandler(w http.ResponseWriter, r *http.Request) {
	addr := r.URL.Query().Get("address")
	if addr == "" {
		http.Error(w, "missing address", http.StatusBadRequest)
		return
	}
	nonce := api.Store.GetNonce(addr)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"address": addr,
		"nonce":   nonce,
	})
}

// chainHeadHandler exposes the current chain head block for explorer-style tooling.
func (api *HTTPAPI) chainHeadHandler(w http.ResponseWriter, r *http.Request) {
	head := api.Chain.Head()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"head": head,
	})
}

func (api *HTTPAPI) chainBlocksHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	fromHeight := uint64(0)
	limit := uint64(50)

	if fh := q.Get("from_height"); fh != "" {
		if v, err := strconv.ParseUint(fh, 10, 64); err == nil {
			fromHeight = v
		}
	}
	if lim := q.Get("limit"); lim != "" {
		if v, err := strconv.ParseUint(lim, 10, 64); err == nil && v > 0 && v <= 500 {
			limit = v
		}
	}

	blocks := api.Chain.Blocks()
	var out []*core.Block
	for _, blk := range blocks {
		if blk.Height < fromHeight {
			continue
		}
		if uint64(len(out)) >= limit {
			break
		}
		out = append(out, blk)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"blocks": out,
		"next_from_height": func() uint64 {
			if len(out) == 0 {
				return fromHeight
			}
			return out[len(out)-1].Height + 1
		}(),
	})
}

func (api *HTTPAPI) chainBlockByHeightHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	fh := q.Get("height")
	if fh == "" {
		http.Error(w, "missing height", http.StatusBadRequest)
		return
	}
	height, err := strconv.ParseUint(fh, 10, 64)
	if err != nil {
		http.Error(w, "invalid height", http.StatusBadRequest)
		return
	}

	blocks := api.Chain.Blocks()
	for _, blk := range blocks {
		if blk.Height == height {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"block": blk,
			})
			return
		}
	}
	http.Error(w, "not found", http.StatusNotFound)
}

// getProfileHandler exposes the active econ profile.
func (api *HTTPAPI) getProfileHandler(w http.ResponseWriter, r *http.Request) {
	prof := econ.GetProfile()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"profile": string(prof),
	})
}

// setProfileHandler changes the active econ profile (DevNet only).
func (api *HTTPAPI) setProfileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Profile string `json:"profile"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Profile == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	econ.SetProfile(econ.ProfileMode(body.Profile))

	ev := Event{
		ID:        "profile-change",
		Type:      EventType("ProfileChange"),
		Version:   "v1",
		Payload:   map[string]string{"profile": body.Profile},
		Timestamp: time.Now().UTC(),
	}
	api.Hub.Broadcast(ev)
	w.WriteHeader(http.StatusOK)
}

// analyticsNAVHandler exposes current NAV plus a small synthetic series.
func (api *HTTPAPI) analyticsNAVHandler(w http.ResponseWriter, r *http.Request) {
	nav := econ.GetLastNAV()
	series := analytics.BuildNAVSeries(nav, 30)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"nav":        nav,
		"nav_series": series,
	})
}

// valuationLatestHandler exposes a coarse monetary snapshot for DevNet,
// derived from in-memory account state. It reports total reserve backing
// held on the treasury account (USDC/USDT/DAI), total GRC supply, the
// implied NAV, and a trivial corridor classification around 1.0000 using
// a ±10bps band.

func (api *HTTPAPI) valuationLatestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	reserve, supply, nav := api.Chain.DevnetMonetarySnapshot()
	target := 1.0
	lower, upper := econ.ComputeCorridorBounds(target, 10)

	pegStatus := "INSIDE"
	mode := "NEUTRAL"
	if nav < lower {
		pegStatus = "BELOW"
		mode = "MINT_ONLY"
	} else if nav > upper {
		pegStatus = "ABOVE"
		mode = "REDEEM_ONLY"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"reserve_usd":    reserve,
		"supply_grc":     supply,
		"nav":            nav,
		"target":         target,
		"corridor_lower": lower,
		"corridor_upper": upper,
		"peg_status":     pegStatus,
		"mode":           mode,
		"decimals":       4,
		"band_bps":       10,
		"timestamp_unix": time.Now().Unix(),
	})
}

// analyticsWindowsHandler exposes the current settlement window snapshot.
func (api *HTTPAPI) analyticsWindowsHandler(w http.ResponseWriter, r *http.Request) {
	ws := api.WMgr.Snapshot()
	flows := analytics.BuildWindowFlows(ws)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"window": ws,
		"flows":  flows,
	})
}

// analyticsTreasuryHandler exposes a tiered MMF-style treasury snapshot.
func (api *HTTPAPI) analyticsTreasuryHandler(w http.ResponseWriter, r *http.Request) {
	nav := econ.GetLastNAV()
	y := econ.GetLastYield()
	tsnap := analytics.BuildTreasurySnapshot(nav, y)

	// Hybrid (M3): structural event written synchronously if DB present.
	if api.DB != nil {
		// Convert to store-level type to avoid import cycles.
		ss := store.TreasurySnapshot{
			TotalUSD: tsnap.TotalUSD,
			Tier1:    store.Tier1Liquidity{CashUSD: tsnap.Tier1.CashUSD},
			Tier2: func() []store.Tier2Bucket {
				out := make([]store.Tier2Bucket, 0, len(tsnap.Tier2))
				for _, b := range tsnap.Tier2 {
					out = append(out, store.Tier2Bucket{Type: b.Type, NotionalUSD: b.NotionalUSD, Rate: b.Rate, DurationD: b.DurationD})
				}
				return out
			}(),
		}
		api.DB.InsertTreasurySnapshot(ss, time.Now().UTC())
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tsnap)
}

// NewHTTPServer wires all HTTP routes for DevNet.
func NewHTTPServer(listenAddr string, hub *WSHub, store *core.AccountStore, wm *econ.WindowManager, db *store.DB, chain *core.Chain, miner *core.Miner) *http.Server {
	api := NewHTTPAPI(hub, store, wm, db, chain, miner)
	mux := http.NewServeMux()

	mux.HandleFunc("/ws", hub.HandleWS)
	mux.HandleFunc("/api/balances", api.balancesHandler)
	mux.HandleFunc("/api/mint", api.mintHandler)
	mux.HandleFunc("/api/redeem", api.redeemHandler)
	// Auth (client wallet) + session
	mux.HandleFunc("/api/auth/nonce", api.authNonceHandler)
	mux.HandleFunc("/api/auth/wallet-login", api.authWalletLoginHandler)
	mux.HandleFunc("/api/session", api.sessionGetHandler)
	mux.HandleFunc("/api/auth/logout", api.authLogoutHandler)

	mux.HandleFunc("/api/profile/get", api.getProfileHandler)
	mux.HandleFunc("/api/profile/set", api.setProfileHandler)
	mux.HandleFunc("/api/analytics/nav", api.analyticsNAVHandler)
	mux.HandleFunc("/api/valuation/latest", api.valuationLatestHandler)
	mux.HandleFunc("/api/analytics/windows", api.analyticsWindowsHandler)
	mux.HandleFunc("/api/account/nonce", api.accountNonceHandler)
	mux.HandleFunc("/api/analytics/treasury", api.analyticsTreasuryHandler)
	mux.HandleFunc("/api/chain/head", api.chainHeadHandler)
	mux.HandleFunc("/api/chain/mining/status", api.miningStatusHandler)
	mux.HandleFunc("/api/chain/mining/start", api.miningStartHandler)
	mux.HandleFunc("/api/chain/mining/stop", api.miningStopHandler)
	mux.HandleFunc("/api/chain/blocks", api.chainBlocksHandler)
	mux.HandleFunc("/api/chain/block", api.chainBlockByHeightHandler)
	mux.HandleFunc("/api/chain/mempool", api.mempoolHandler)
	mux.HandleFunc("/api/tx/transfer", api.transferHandler)
	mux.HandleFunc("/api/tx/vault_create", api.vaultCreateHandler)
	mux.HandleFunc("/api/tier/renew", api.tierRenewHandler)
	mux.HandleFunc("/api/p2p/register", api.p2pRegisterHandler)
	mux.HandleFunc("/api/p2p/peers", api.p2pPeersHandler)

	mux.HandleFunc("/workstation/", api.workstationHandler)
	mux.HandleFunc("/workstation", api.workstationHandler)

	mux.HandleFunc("/econ/live", api.econLiveHandler)
	mux.HandleFunc("/api/sim", api.econSimHandler)

	mux.HandleFunc("/api/econ/epoch-commit", api.econEpochCommitHandler)
	mux.HandleFunc("/api/slashing/events", api.slashingEventsHandler)
	// RSX staking + PoP wiring (state + payouts)
	mux.HandleFunc("/api/staking/validators", api.stakingValidatorsHandler)
	mux.HandleFunc("/api/staking/lock", api.stakingLockHandler)
	mux.HandleFunc("/api/staking/unlock", api.stakingUnlockHandler)
	// backwards-compatible alias
	mux.HandleFunc("/api/staking/stake", api.stakingLockHandler)
	mux.HandleFunc("/api/staking/state", api.stakingStateHandler)
	mux.HandleFunc("/api/pop/register-node", api.popRegisterNodeHandler)
	mux.HandleFunc("/api/pop/submit-caps", api.popSubmitCapsHandler)
	mux.HandleFunc("/api/pop/claim-work", api.popClaimWorkHandler)
	mux.HandleFunc("/api/pop/submit-metrics", api.popSubmitMetricsHandler)
	mux.HandleFunc("/api/pop/payouts", api.popPayoutsHandler)

	if listenAddr == "" {
		listenAddr = ":8080"
	}
	return &http.Server{
		Addr:    listenAddr,
		Handler: mux,
	}
}

// p2pRegisterHandler allows a peer to register its address with this node
// when running in seed mode. It expects a JSON body like:
//
//	{ "addr": "127.0.0.1:8080" }
func (api *HTTPAPI) p2pRegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if !seedMode {
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "seed_mode_disabled",
		})
		return
	}
	type payload struct {
		Addr string `json:"addr"`
	}
	var p payload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "invalid_payload",
		})
		return
	}
	if p.Addr == "" {
		// Fallback to remote address if caller did not send one.
		p.Addr = r.RemoteAddr
	}
	seedRegistry.Add(p.Addr)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}

// p2pPeersHandler returns the current set of known peers from the in-memory
// seed registry. Entries are TTL-aged each time this handler is invoked.
func (api *HTTPAPI) p2pPeersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if !seedMode {
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": "seed_mode_disabled",
		})
		return
	}
	peers := seedRegistry.List()
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"peers": peers,
	})
}
