package net

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"
)

// DiscoverPeersFromSeeds contacts the configured seed HTTP endpoints, registers
// this node's advertised address, and pulls the current peer list. It returns
// a de-duplicated set of HTTP base URLs suitable for PeerSync.
//
// seeds should be host:port, "http://host:port", etc. selfAddr should be the
// HTTP base URL for this node's /api endpoints (e.g. "http://127.0.0.1:8080").
func DiscoverPeersFromSeeds(seeds []string, selfAddr string, maxPeers int) []string {
	client := &http.Client{Timeout: 3 * time.Second}
	seen := make(map[string]struct{})
	peers := make([]string, 0, maxPeers)

	for _, raw := range seeds {
		if raw == "" {
			continue
		}
		base := normaliseSeedBase(raw)

		// Best-effort registration; failure here is non-fatal for discovery.
		if selfAddr != "" {
			payload := map[string]string{"addr": selfAddr}
			buf, _ := json.Marshal(payload)
			req, err := http.NewRequest(http.MethodPost, base+"/api/p2p/register", bytes.NewReader(buf))
			if err == nil {
				req.Header.Set("Content-Type", "application/json")
				resp, err := client.Do(req)
				if err == nil {
					_ = resp.Body.Close()
				}
			}
		}

		// Pull peer list from this seed.
		resp, err := client.Get(base + "/api/p2p/peers")
		if err != nil {
			log.Printf("[p2p] seed %s peers error: %v", base, err)
			continue
		}
		if resp.StatusCode != http.StatusOK {
			_ = resp.Body.Close()
			log.Printf("[p2p] seed %s peers status: %s", base, resp.Status)
			continue
		}
		var payload struct {
			Peers []string `json:"peers"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
			_ = resp.Body.Close()
			log.Printf("[p2p] seed %s peers decode error: %v", base, err)
			continue
		}
		_ = resp.Body.Close()

		for _, p := range payload.Peers {
			if p == "" {
				continue
			}
			if !strings.HasPrefix(p, "http://") && !strings.HasPrefix(p, "https://") {
				p = "http://" + p
			}
			if selfAddr != "" && p == selfAddr {
				continue
			}
			if _, ok := seen[p]; ok {
				continue
			}
			seen[p] = struct{}{}
			peers = append(peers, p)
			if maxPeers > 0 && len(peers) >= maxPeers {
				return peers
			}
		}
	}

	return peers
}

// normaliseSeedBase ensures we always have an HTTP base URL for contacting
// seed discovery endpoints.
func normaliseSeedBase(raw string) string {
	if strings.HasPrefix(raw, "http://") || strings.HasPrefix(raw, "https://") {
		return raw
	}
	// If the seed is given as ":port" or "host:port", assume HTTP.
	if strings.HasPrefix(raw, ":") {
		return "http://127.0.0.1" + raw
	}
	return "http://" + raw
}
