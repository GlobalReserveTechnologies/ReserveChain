
package econ

// Mainnet-oriented policy helpers.
// ---------------------------------------------------------
// This file introduces a first-pass, conservative policy layer
// that can be used to filter / reshape pending monetary actions
// (mints, redemptions, issuance) before they are applied by
// MainnetMonetaryState.SettleEpoch.
//
// The intent is that more sophisticated corridor / FX / volatility
// handling can evolve here over time, but the basic safety rules
// around USDR coverage are enforced from the start.

import "math"

// BasicPolicyConfig holds coarse global thresholds for mainnet.
// In a real deployment these would be configurable via governance
// or operator policies, but here they are hard-coded defaults.
type BasicPolicyConfig struct {
    MinUSDRCoverage float64 // minimum allowed reserve coverage for USDR (e.g. 1.0 = 100%)
}

// DefaultBasicPolicyConfig returns conservative, hard-coded
// thresholds suitable for early mainnet iterations.
func DefaultBasicPolicyConfig() BasicPolicyConfig {
    return BasicPolicyConfig{
        MinUSDRCoverage: 1.0,
    }
}

// ApplyBasicUSDRPolicy inspects the pending entries on the given
// monetary state and rejects any USDR mints that would cause the
// reserve coverage to fall below the configured minimum.
// Redemptions are always allowed up to the available supply.
func ApplyBasicUSDRPolicy(
    s MainnetMonetaryState,
    prices map[CryptoAssetKind]float64,
    cfg BasicPolicyConfig,
) MainnetMonetaryState {
    // Compute current reserve USD value using the same mapping
    // logic as ToTreasuryBalanceSheet.
    var reservesUSD float64
    for _, pool := range s.Reserves {
        for _, b := range pool.Balances {
            var kind CryptoAssetKind
            switch string(b.Asset) {
            case string(AssetUSDC):
                kind = AssetUSDC
            case string(AssetUSDT):
                kind = AssetUSDT
            case string(AssetDAI):
                kind = AssetDAI
            case string(AssetETH):
                kind = AssetETH
            case string(AssetWBTC):
                kind = AssetWBTC
            default:
                kind = ""
            }
            if kind == "" {
                continue
            }
            px, ok := prices[kind]
            if !ok {
                continue
            }
            reservesUSD += b.Amount * px
        }
    }

    effectiveSupply := s.Supply.USDR
    if effectiveSupply < 0 {
        effectiveSupply = 0
    }

    // Walk pending entries and build a filtered list. We do this
    // in a dry-run manner, adjusting a working copy of the supply
    // so that subsequent entries see the effect of earlier ones.
    filtered := make([]MainnetPendingEntry, 0, len(s.Pending))

    for _, p := range s.Pending {
        switch p.Kind {
        case PendingMintUSDR:
            // Projected coverage after this mint would be:
            //   coverage' = reservesUSD / (supply + amount)
            // If this drops below MinUSDRCoverage, we reject the
            // mint entirely for now (no partial fills).
            if p.Amount <= 0 {
                continue
            }
            projectedSupply := effectiveSupply + p.Amount
            if projectedSupply <= 0 {
                continue
            }
            coverage := reservesUSD / projectedSupply
            if coverage+1e-9 < cfg.MinUSDRCoverage {
                // Reject: do not include in filtered list.
                continue
            }
            // Accept and update working supply.
            effectiveSupply = projectedSupply
            filtered = append(filtered, p)

        case PendingRedeemUSDR:
            // Allow redemptions, but cap the amount to what is
            // actually outstanding so we do not drive supply
            // negative (even though SettleEpoch also floors).
            if p.Amount <= 0 || effectiveSupply <= 0 {
                continue
            }
            amt := p.Amount
            if amt > effectiveSupply {
                amt = effectiveSupply
            }
            if amt <= 0 {
                continue
            }
            // Mutate a copy of the entry with the capped amount.
            q := p
            q.Amount = amt
            effectiveSupply -= amt
            filtered = append(filtered, q)

        default:
            // For now, non-USDR entries are passed through
            // unchanged. GRC issuance policy will be layered
            // on top of this in a later iteration.
            filtered = append(filtered, p)
        }
    }

    // Replace the state's pending set with the filtered slice.
    s.Pending = filtered
    // Clamp supply to non-negative to avoid drift due to float math.
    s.Supply.USDR = math.Max(0, effectiveSupply)

    return s
}
