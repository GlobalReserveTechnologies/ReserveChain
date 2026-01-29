package identity

import (
	"strings"
)

// Canonical normalizes a (walletType,address) pair into a canonical
// wallet type and address string for internal accounting.
//
// Canonical wallet types:
//   - "rc"  : ReserveChain native wallet IDs (e.g. rc1...)
//   - "evm" : EVM addresses (0x...)
func Canonical(walletType, address string) (wt string, addr string, ok bool) {
	wt = strings.ToLower(strings.TrimSpace(walletType))
	a := strings.TrimSpace(address)

	switch wt {
	case "rc":
		if a == "" {
			return "", "", false
		}
		// Keep rc addresses as-is but trimmed.
		return "rc", a, true
	case "evm":
		if a == "" {
			return "", "", false
		}
		// EVM addresses are case-insensitive; store lowercase.
		a = strings.ToLower(a)
		if !strings.HasPrefix(a, "0x") || len(a) < 10 {
			// basic sanity; exact checksum validation can be added later
			return "", "", false
		}
		return "evm", a, true
	default:
		return "", "", false
	}
}

// CanonicalID returns an internal identity key (e.g. "rc:rc1..." or "evm:0x...")
func CanonicalID(walletType, address string) (id string, ok bool) {
	wt, addr, ok := Canonical(walletType, address)
	if !ok {
		return "", false
	}
	return wt + ":" + addr, true
}
