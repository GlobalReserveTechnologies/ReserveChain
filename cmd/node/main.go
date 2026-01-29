package main

import (
	"log"
	"net/http"
	"strings"
	"time"

	"reservechain/internal/config"
	"reservechain/internal/core"
	"reservechain/internal/econ"
	"reservechain/internal/net"
	storepkg "reservechain/internal/store"
)

func main() {
	cfg, err := config.Load("config/devnet.yaml")
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	// Enable seed mode for this node if configured.
	if cfg.P2P.Mode == "seed" {
		net.EnableSeedMode(true)
	}

	nodeID := cfg.Node.ID
	allNodes := []string{nodeID}

	// Core services
	leaderSel := net.NewLeaderSelector(allNodes)
	wsHub := net.NewWSHub()
	store := core.NewAccountStore()
	core.SeedDemoBalances(store)

	// DevNet chain engine with DB-backed block log
	var chain *core.Chain

	wm := econ.NewWindowManager(cfg)
	econ.InitStateForDevnet()

	// Open DevNet SQLite database (optional; logs if missing schema).
	dbPath := cfg.Node.DB.SQLitePath
	if dbPath == "" {
		dbPath = "runtime/chain_devnet.sqlite"
	}
	sqldb, dberr := storepkg.OpenSQLite(dbPath)
	if dberr == nil && sqldb != nil {
		if err := storepkg.EnsureSchemaFromFile(sqldb, "database/schema.sql"); err != nil {
			log.Printf("[node] warning: could not apply schema.sql automatically: %v", err)
		}
	}
	// Construct chain engine once DB is available so it can replay or persist.
	chain = core.NewChain(store, sqldb)
	// Wire chain + DB into econ so DevNet epoch settlement can credit payouts.
	econ.SetRuntime(chain, sqldb)

	// Heartbeat miner (produces EMPTY blocks when enabled)
	miner := core.NewMiner(chain, 5*time.Second)
	if dberr != nil {
		log.Printf("[node] warning: could not open SQLite DB: %v", dberr)
	}
	defer func() {
		if sqldb != nil {
			_ = sqldb.Close()
		}
	}()

	listenAddr := cfg.Node.RPC.HTTPListen
	if listenAddr == "" {
		listenAddr = ":8080"
	}

	httpServer := net.NewHTTPServer(listenAddr, wsHub, store, wm, sqldb, chain, miner)

	// Start HTTP + WS server
	go wsHub.Run()
	go func() {
		log.Printf("HTTP/WS server on %s", listenAddr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	// Optional follower sync loop: if this node is configured with an
	// upstream URL, it will act as a HTTP follower and mirror the block
	// log from that peer. This is a DevNet-friendly way to run multiple
	// nodes without full P2P wiring yet.
	if cfg.Node.FollowUpstreamURL != "" {
		log.Printf("[node] starting follower loop against %s", cfg.Node.FollowUpstreamURL)
		follower := net.NewChainFollower(chain, sqldb, cfg.Node.FollowUpstreamURL, 3*time.Second)
		go follower.Run()
	}

	// Cluster v2 P2P-style peer sync:
	// - If this node is in peer mode, it will contact configured seed_nodes,
	//   register itself, and pull a live peer list from /api/p2p/peers.
	// - Manual cfg.Node.Peers are still honoured and merged.
	// - If seeds are unavailable, we gracefully fall back to manual peers only.
	var discoveredPeers []string
	if cfg.P2P.Mode == "peer" && len(cfg.P2P.SeedNodes) > 0 {
		// Derive an advertised base URL for this node's HTTP API.
		selfBase := listenAddr
		if strings.HasPrefix(selfBase, ":") {
			selfBase = "http://127.0.0.1" + selfBase
		} else if strings.HasPrefix(selfBase, "http://") || strings.HasPrefix(selfBase, "https://") {
			// already a full URL
		} else {
			selfBase = "http://" + selfBase
		}
		discoveredPeers = net.DiscoverPeersFromSeeds(cfg.P2P.SeedNodes, selfBase, cfg.P2P.MaxPeers)
		if len(discoveredPeers) == 0 {
			log.Printf("[node] seed discovery returned no peers, falling back to static peers only")
		} else {
			log.Printf("[node] discovered %d peer(s) from seed registry", len(discoveredPeers))
		}
	}

	// Merge static peers (if any) with discovered peers, de-duplicating.
	allPeers := make([]string, 0, len(cfg.Node.Peers)+len(discoveredPeers))
	seen := make(map[string]struct{})
	for _, p := range cfg.Node.Peers {
		if p == "" {
			continue
		}
		if !strings.HasPrefix(p, "http://") && !strings.HasPrefix(p, "https://") {
			p = "http://" + p
		}
		if _, ok := seen[p]; !ok {
			seen[p] = struct{}{}
			allPeers = append(allPeers, p)
		}
	}
	for _, p := range discoveredPeers {
		if p == "" {
			continue
		}
		if _, ok := seen[p]; !ok {
			seen[p] = struct{}{}
			allPeers = append(allPeers, p)
		}
	}

	if len(allPeers) > 0 {
		log.Printf("[node] starting peer sync against %d peer(s)", len(allPeers))
		ps := net.NewPeerSync(chain, sqldb, allPeers, 5*time.Second)
		go ps.Run()
	}

	// Periodic valuation ticks
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	var tickID uint64
	for range ticker.C {
		tickID++
		leaderID := leaderSel.LeaderForTick(tickID)
		if leaderID != nodeID {
			continue
		}
		val := econ.ComputeDevnetTick(tickID, leaderID, wm, time.Now().UTC())
		ev := net.Event{
			ID:        makeTickEventID(tickID),
			Type:      net.EventValuationTick,
			Version:   "v1",
			Payload:   val,
			Timestamp: time.Now().UTC(),
		}
		wsHub.Broadcast(ev)
	}
}

func makeTickEventID(tickID uint64) string {
	return "tick-" + time.Now().Format("20060102150405")
}
