package net

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"time"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
)

// nonceEntry tracks a login nonce issued for a given address.
type nonceEntry struct {
	Address   string
	Challenge string
	ExpiresAt time.Time
}

// sessionEntry tracks an authenticated session.
type sessionEntry struct {
	SessionID string
	Address   string
	Role      string
	CreatedAt time.Time
	ExpiresAt time.Time
}

// -------- helpers --------

func canonicalIdentity(walletType, address string) (wt string, addr string, ok bool) {
	return identity.Canonical(walletType, address)
}

func ethPersonalSignHash(message string) []byte {
	// Ethereum personal_sign uses:
	// keccak256("\x19Ethereum Signed Message:\n" + len(message) + message)
	msgBytes := []byte(message)
	prefix := fmt.Sprintf("\x19Ethereum Signed Message:\n%d", len(msgBytes))
	return ethcrypto.Keccak256([]byte(prefix), msgBytes)
}

func verifyEvmPersonalSign(address string, challenge string, sigHex string) bool {
	sigHex = strings.TrimSpace(sigHex)
	sigHex = strings.TrimPrefix(sigHex, "0x")
	sig, err := hex.DecodeString(sigHex)
	if err != nil || len(sig) != 65 {
		return false
	}
	// go-ethereum expects v as 0/1
	if sig[64] >= 27 {
		sig[64] -= 27
	}
	if sig[64] != 0 && sig[64] != 1 {
		return false
	}
	hash := ethPersonalSignHash(challenge)
	pub, err := ethcrypto.SigToPub(hash, sig)
	if err != nil {
		return false
	}
	recAddr := ethcrypto.PubkeyToAddress(*pub).Hex()
	return strings.EqualFold(recAddr, address)
}

func (api *HTTPAPI) pruneAuthLocked(now time.Time) {
	for addr, n := range api.nonces {
		if now.After(n.ExpiresAt) {
			delete(api.nonces, addr)
		}
	}
	for sid, s := range api.sessions {
		if now.After(s.ExpiresAt) {
			delete(api.sessions, sid)
		}
	}
}

func randTokenB64URL(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// ReserveChain address derivation (must match wallet_keystore.js):
// address = "rc1" + base64url(sha256(x+"."+y)).slice(0,38).toLowerCase()
func deriveAddressFromJWKXY(x, y string) string {
	h := sha256.Sum256([]byte(x + "." + y))
	b64 := base64.StdEncoding.EncodeToString(h[:])
	b64 = strings.TrimRight(b64, "=")
	b64 = strings.ReplaceAll(b64, "+", "-")
	b64 = strings.ReplaceAll(b64, "/", "_")
	if len(b64) > 38 {
		b64 = b64[:38]
	}
	return "rc1" + strings.ToLower(b64)
}

func jwkToP256PublicKey(pub map[string]any) (*ecdsa.PublicKey, bool) {
	// Expect JWK fields: kty="EC", crv="P-256", x, y (base64url)
	kty, _ := pub["kty"].(string)
	crv, _ := pub["crv"].(string)
	xs, _ := pub["x"].(string)
	ys, _ := pub["y"].(string)
	if kty != "EC" || crv != "P-256" || xs == "" || ys == "" {
		return nil, false
	}

	xb, err := base64.RawURLEncoding.DecodeString(xs)
	if err != nil {
		return nil, false
	}
	yb, err := base64.RawURLEncoding.DecodeString(ys)
	if err != nil {
		return nil, false
	}

	x := new(big.Int).SetBytes(xb)
	y := new(big.Int).SetBytes(yb)

	curve := elliptic.P256()
	if !curve.IsOnCurve(x, y) {
		return nil, false
	}

	return &ecdsa.PublicKey{Curve: curve, X: x, Y: y}, true
}

// -------- handlers --------

// POST /api/auth/nonce
// Body: { "address": "rc1..." }
// Resp: { "challenge": "ReserveChain login: <nonce>", "expires_at": "...RFC3339..." }
func (api *HTTPAPI) authNonceHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		WalletType string `json:"wallet_type"`
		Address    string `json:"address"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": false, "error": "bad_request"})
		return
	}

	wt, addr, ok := canonicalIdentity(req.WalletType, req.Address)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": false, "error": "bad_address"})
		return
	}

	nonce, err := randTokenB64URL(24)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	challenge := "ReserveChain login: " + nonce
	exp := time.Now().UTC().Add(5 * time.Minute)

	id := wt + ":" + addr

	api.authMu.Lock()
	api.pruneAuthLocked(time.Now().UTC())
	api.nonces[id] = nonceEntry{
		Address:   id,
		Challenge: challenge,
		ExpiresAt: exp,
	}
	api.authMu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"ok":          true,
		"wallet_type": wt,
		"address":     addr,
		"identity":    id,
		"challenge":   challenge,
		"expires_at":  exp.Format(time.RFC3339),
	})
}

// POST /api/auth/wallet-login

// Body:
//
//	{
//	  "address": "rc1...",
//	  "pub": { "kty":"EC","crv":"P-256","x":"...","y":"..." },
//	  "challenge": "ReserveChain login: ...",
//	  "signature_b64": "....",
//	  "role": "client"
//	}
func (api *HTTPAPI) authWalletLoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		WalletType   string         `json:"wallet_type"`
		Address      string         `json:"address"`
		Pub          map[string]any `json:"pub,omitempty"`
		Challenge    string         `json:"challenge"`
		SignatureB64 string         `json:"signature_b64,omitempty"`
		SignatureHex string         `json:"signature_hex,omitempty"`
		Role         string         `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": false, "error": "bad_request"})
		return
	}
	if req.Role == "" {
		req.Role = "client"
	}
	if req.Role != "client" {
		w.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": false, "error": "role_not_allowed"})
		return
	}

	wt, addr, ok := canonicalIdentity(req.WalletType, req.Address)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": false, "error": "bad_address"})
		return
	}
	id := wt + ":" + addr

	// Validate nonce/challenge
	api.authMu.Lock()
	api.pruneAuthLocked(time.Now().UTC())
	ne, exists := api.nonces[id]
	api.authMu.Unlock()
	if !exists || time.Now().UTC().After(ne.ExpiresAt) {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": false, "error": "nonce_expired"})
		return
	}
	if req.Challenge == "" {
		req.Challenge = ne.Challenge
	}
	if req.Challenge != ne.Challenge {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": false, "error": "bad_challenge"})
		return
	}

	// Verify signature depending on wallet type
	switch wt {
	case "rc":
		pubKey, ok := jwkToP256PublicKey(req.Pub)
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]any{"ok": false, "error": "bad_pubkey"})
			return
		}
		xs, _ := req.Pub["x"].(string)
		ys, _ := req.Pub["y"].(string)
		derivedAddr := deriveAddressFromJWKXY(xs, ys)

		// For RC wallets, the provided addr must match the derived RC address.
		if derivedAddr != addr {
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]any{"ok": false, "error": "address_mismatch"})
			return
		}

		sig, err := base64.StdEncoding.DecodeString(req.SignatureB64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]any{"ok": false, "error": "bad_signature_encoding"})
			return
		}
		msgHash := sha256.Sum256([]byte(req.Challenge))
		if !ecdsa.VerifyASN1(pubKey, msgHash[:], sig) {
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]any{"ok": false, "error": "bad_signature"})
			return
		}
	case "evm":
		if !verifyEvmPersonalSign(addr, req.Challenge, req.SignatureHex) {
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]any{"ok": false, "error": "bad_signature"})
			return
		}
	default:
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": false, "error": "wallet_type_not_supported"})
		return
	}

	// Issue session
	sid, err := randTokenB64URL(32)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	now := time.Now().UTC()
	exp := now.Add(24 * time.Hour)

	api.authMu.Lock()
	api.sessions[sid] = sessionEntry{
		SessionID: sid,
		Address:   id, // store canonical identity
		Role:      req.Role,
		CreatedAt: now,
		ExpiresAt: exp,
	}
	// One-time nonce use
	delete(api.nonces, id)
	api.authMu.Unlock()

	http.SetCookie(w, &http.Cookie{
		Name:     "rc_session",
		Value:    sid,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   r.TLS != nil,
		Expires:  exp,
	})

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"ok":          true,
		"wallet_type": wt,
		"address":     addr,
		"identity":    id,
		"role":        req.Role,
		"expires_at":  exp.Format(time.RFC3339),
	})
}

// GET /api/session

func (api *HTTPAPI) sessionGetHandler(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("rc_session")
	if err != nil || c.Value == "" {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": false})
		return
	}

	api.authMu.Lock()
	api.pruneAuthLocked(time.Now().UTC())
	s, ok := api.sessions[c.Value]
	api.authMu.Unlock()

	if !ok || time.Now().UTC().After(s.ExpiresAt) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": false})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"ok":      true,
		"role":    s.Role,
		"address": s.Address,
		"expires": s.ExpiresAt.Format(time.RFC3339),
	})
}

// POST /api/auth/logout
func (api *HTTPAPI) authLogoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	c, err := r.Cookie("rc_session")
	if err == nil && c.Value != "" {
		api.authMu.Lock()
		delete(api.sessions, c.Value)
		api.authMu.Unlock()
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "rc_session",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   r.TLS != nil,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
	})

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
}
