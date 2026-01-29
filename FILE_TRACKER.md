# ReserveChain Devnet — File Tracker (v2.0.0-pre5)

This tracker gives a high-level map of the most important files and directories
in this build. It is not an exhaustive listing, but a curated overview of where
things live.

## Top-level

- `README.md` — project overview and run instructions
- `FILE_TRACKER.md` — this file
- `go.mod` — Go module definition
- `cmd/node/main.go` — node entrypoint
- `config/devnet.yaml` — main devnet configuration
- `database/schema.sql` — SQLite schema for chain + wallet + operator scaffolding
- `runtime/` — devnet databases and runtime artifacts
- `scripts/` — helper scripts for starting/resetting devnet

## Backend (Go)

- `internal/core/chain.go` — chain struct, block list, basic transition logic
- `internal/core/work.go` — node work scoring model (consensus/network/storage/service)
- `internal/net/server.go` (or similar) — HTTP/WS server wiring
- `internal/net/follower.go` — single-upstream follower sync loop
- `internal/net/peersync.go` — multi-peer HTTP-based sync loop (P2P v1)
- `internal/store/db.go` — SQLite connection + helpers
- `rpc/` — PHP/HTTP layer that talks to the Go node APIs

## Web Frontend

- `public/index.php` — shell for marketing + workstation entry
- `public/sections/overview...php` — main overview content
- `public/sections/governance.php` — governance section
- `public/sections/security.php` — security section
- `public/sections/docs.php` — docs/documentation section
- `public/sections/faq.php` — FAQ section

### Workstation

- `public/workstation/index.php` — workstation shell / main layout
- `public/assets/css/workstation.css` — workstation styling
- `public/assets/js/workstation_*.js` — workstation modules:
  - wallet & balances
  - vaults (including duration and PoP-ready scaffolding)
  - treasury & reserves
  - explorer
  - risk dashboard
  - tiers & billing scaffolding

### Trading Terminal

- `public/trading-terminal/index.php` — popup shell for the trading terminal
- `public/assets/js/trading_terminal*.js` — JS hooks for the terminal UI (to be evolved)

## Scripts & Runtime

- `scripts/` — platform scripts for starting node, resetting DB, etc.
- `runtime/` — created at runtime; stores SQLite DB and node artifacts

Future builds will expand this tracker with more detail as additional features
are implemented (consensus phases, PoP flows, trading terminal wiring, etc.).


## Workstation wallet & risk additions

- `public/workstation/index.php` — now includes a wallet popup modal (`#wallet-popup-modal`) and an enhanced Network → Risk Dashboard panel.
- `public/assets/js/wallet_popup.js` — browser-keystore-backed wallet popup + `reserveWalletApi` wrapper for workstation and trading modules.
- `public/assets/js/workstation_risk_dashboard.js` — frontend wiring for the Network → Risk Dashboard panel (summary cards + tables using `risk_summary.php`).
- `public/assets/css/workstation.css` — extended with wallet popup and risk layout styles.
- `public/api/risk_summary.php` — PHP stub endpoint returning a synthetic global risk snapshot for DevNet.


### v2.0.0-pre5 delta
- `public/api/risk_summary.php` — now queries the local Go node for chain head and mempool data when available.
- `public/assets/js/workstation_risk_dashboard.js` — async bug fix in `loadGlobal()` and displays chain height + mempool length in the Network Risk summary.
v2.0.0-pre6: Enabled WebSocket-driven feed in trading terminal and wired workstation wallet popup to live balances via wallet_chain_client.js. Fixed risk dashboard JS typo.
v2.0.0-pre7: Added Windows helper scripts start_node1/2/3, start_website, start_all using runtime\go and runtime\php without env vars.
- v2.0.0-pre8: Refined devnet.yaml with structured sections and moved/updated all launcher scripts under scripts/ with improved UX and error handling.

- `internal/econ/reserve_pools_crypto.go` — crypto-only reserve pool types + NAV helpers for GRC backing

- `internal/econ/valuation.go` — FX/yield/NAV types plus corridor helper (ComputeCorridorBounds) for GRC band logic

- `internal/core/chain.go`, `internal/net/http_api.go` — DevNet Mint/Redeem now treat USDC as the default backing asset instead of USD (crypto-only reserve mode).

- `internal/core/chain.go` — DevNet Mint/Redeem now enforce allowed backing assets and use a simple NAV-based conversion (USDC-only redemption, R3).

- `internal/core/chain.go`, `internal/net/http_api.go` — added DevNet monetary snapshot helper and `/api/valuation/latest` for NAV/reserve/corridor status.

- `internal/core/chain.go` — DevNet NAV now values multi-asset reserves (USDC/USDT/DAI/ETH/WBTC) and Mint/Redeem are corridor-gated around 1.0000 (±10bps).

- `public/sections/workstation.php`, `public/assets/js/workstation_reserve_monitor.js`, `public/assets/css/site.css` — Workstation Reserve Monitor card wired to `/api/valuation/latest` for live NAV/corridor display.

- `internal/core/chain.go`, `internal/net/http_api.go` — Arbitrage-friendly corridor policy: Mint allowed up to upper band, Redeem allowed from lower band, `/api/valuation/latest` now reports a `mode` hint (MINT_ONLY / REDEEM_ONLY / NEUTRAL).

- `public/index.php`, `public/sections/privacy.php` — Top nav now links to “Privacy & Vaults” and privacy section explains Private Vaults as a private banking layer.
- `public/workstation/index.php`, `public/assets/js/workstation_vault_dashboard.js` — Vault Dashboard labeled as the private banking layer; copy updated to reflect client/strategy segmentation.

- `public/api/vault.php`, `public/assets/js/vault_client.js`, `public/assets/js/workstation_vault_dashboard.js`, `public/assets/css/site.css` — Vault balances now treated as client currency: server returns NAV + USD-equivalent per vault, dashboard shows USD-eq with underlying GRC.

- `public/api/vault.php`, `public/assets/js/vault_client.js`, `public/assets/js/workstation_vault_dashboard.js`, `public/workstation/index.php`, `public/assets/css/site.css` — Vault templates added (client/strategy/treasury); create flow picks a template and stores it, dashboard shows template labels and USD-eq / GRC balances.


- `internal/econ/treasury_balance.go` — New DevNet-only treasury balance sheet + coverage model. Tracks synthetic crypto reserve pools, USDR/GRC supplies, pending USDR redemptions, and computes mark-to-market assets, equity, and coverage ratios. Called from `internal/econ/devnet_wrap.go` and exposed over RPC via `rpc/econ/econ_handlers.go:/treasury`.


- `internal/econ/redemptions.go` — DevNet mint / redeem queue + epoch scaffolding. Tracks in-memory USDR redemption requests, a simple epoch counter, and exposes helpers to enqueue requests, snapshot the queue, and advance epochs while updating the treasury-facing pending redemption total and USDR supply.
- `apps/workstation/src/App.tsx` — Wired a dedicated `TreasuryOverviewPanel` to the `treasury-overview` route, fetching `/econ/treasury` and rendering reserve assets, USDR/GRC supplies, coverage ratios, and a basic redemption queue summary.
- `apps/workstation/src/workstation.css` — Added `rc-ws-card*` utility classes for Treasury panel cards.
- `rpc/econ/econ_handlers.go` — Extended econ RPC surface with `/econ/redemptions` for DevNet redemption queue snapshots alongside `/econ/treasury`.


- `apps/workstation/src/App.tsx` — Added a dedicated `RedemptionQueuePanel` wired to the `treasury-redemptions` route, reading `/econ/redemptions` and rendering a live epoch + per-request redemption queue table.
- `apps/workstation/src/workstation.css` — Added basic `rc-ws-table*` styles for small data tables used by the Redemption Queue panel.


- `internal/econ/mints.go` — DevNet mint queue + epoch scaffolding. Tracks USDR mint requests (crypto-backed and free test mints), exposes enqueue + snapshot helpers, and settles mints at epoch boundaries while updating treasury reserve pools and USDR supply.
- `internal/econ/redemptions.go` — `AdvanceDevnetEpoch` now settles DevNet mints first (via `settleDevnetMintsForEpoch`) before processing the redemption queue, so mint + redeem flows share the same epoch clock.
- `rpc/econ/econ_handlers.go` — Added `/econ/mints` (mint queue snapshot) and `/econ/advance-epoch` (DevNet-only epoch advance hook) to the econ RPC surface.
- `apps/workstation/src/App.tsx` — Removed the unused `network-governance` panel from the Network section and extended `RedemptionQueuePanel` with a DevNet-only “Force Epoch Advance” operator button that POSTs to `/econ/advance-epoch`.


- `apps/workstation/src/App.tsx` — Added a Mint Queue panel (`treasury-mints`) that consumes `/econ/mints` and renders a DevNet mint queue summary + per-request table alongside the existing Treasury and Redemption Queue views.


- `internal/econ/grc_issuance.go` — DevNet GRC issuance signal engine. Produces a coverage- and equity-aware `GRCIssuanceSignal` based on the current `SnapshotTreasury` state without actually minting or burning GRC.
- `rpc/econ/econ_handlers.go` — Added `/econ/grc-issuance` endpoint to expose the current GRC issuance recommendation over RPC for future Operator Console / analytics panels.


- `internal/econ/ledger_mainnet.go` — Introduced mainnet-oriented monetary ledger primitives (`MainnetMonetaryState`, supply/reserve/liability/equity structs, and a bridge helper to render a `TreasuryBalanceSheet` view from mainnet state). This is the canonical structure future issuance / mint / redeem policy will target, rather than the ad-hoc DevNet helpers.


- `internal/econ/ledger_mainnet.go` — Extended the mainnet ledger with functional helpers to enqueue pending USDR/GRC actions (`WithPendingMintUSDR`, `WithPendingRedeemUSDR`, `WithPendingIssueGRC`, `WithPendingBurnGRC`) and a `SettleEpoch` method that applies pending entries, advances the epoch counter, and clears the pending set. This will be the core primitive that future mint/redeem/issuance policy uses for final settlement.


- `internal/econ/policy_mainnet.go` — Added a first-pass mainnet policy layer with `BasicPolicyConfig` and `ApplyBasicUSDRPolicy`, which filters pending USDR mints/redemptions based on reserve coverage and available supply before epoch settlement. This is the initial safety net for mainnet-style USDR behavior.


- `internal/econ/mainnet_state.go` — Added an in-memory singleton `MainnetMonetaryState` holder with helpers to snapshot and advance the mainnet monetary state (`SnapshotMainnetState`, `SetMainnetState`, `SettleMainnetEpochBasic`).
- `rpc/econ/econ_handlers.go` — Extended the econ RPC surface with `/econ/mainnet-state` (read-only view of the mainnet monetary state) and `/econ/settle-mainnet-epoch` (DevNet/Testnet-only endpoint that applies basic USDR policy and settles a mainnet epoch).


- `rpc/econ/econ_handlers.go` — Added `/econ/mint-usdr` and `/econ/redeem-usdr` endpoints that enqueue pending USDR mints/redemptions onto the mainnet monetary state (`MainnetMonetaryState`) instead of directly manipulating supply. These are the first mainnet-style mint/redeem RPCs and are finalized only when `/econ/settle-mainnet-epoch` is called.


- `rpc/econ/econ_handlers.go` — Extended the econ RPC surface with `/econ/issue-grc` and `/econ/burn-grc` endpoints that enqueue pending GRC issuance/burn entries on the mainnet ledger. Like the USDR endpoints, these do not immediately change supply and rely on `/econ/settle-mainnet-epoch` for finalization.


- `internal/econ/grc_issuance.go` — Extended with `GRCPolicyConfig`, a hybrid mainnet GRC issuance model (`ComputeMainnetGRCIssuanceSignal`, `ApplyMainnetGRCPolicy`), and a convenience wrapper `ComputeMainnetGRCIssuanceSignalAuto` that operates on the in-memory mainnet state.
- `internal/econ/mainnet_state.go` — Updated `SettleMainnetEpochBasic` to apply both USDR coverage policy and the new hybrid GRC policy before calling `SettleEpoch`, so each mainnet epoch now drives deterministic GRC issuance/burn based on corridor, equity, and demand impulses.
- `rpc/econ/econ_handlers.go` — Switched `/econ/grc-issuance` to use the mainnet signal (`ComputeMainnetGRCIssuanceSignalAuto`) instead of the DevNet-only helper, aligning the operator-facing view with the new mainnet policy engine.


- `apps/workstation/src/App.tsx` — Added a dedicated `GRCPolicyPanel` that visualizes the mainnet GRC issuance signal (`/econ/grc-issuance`) and wired it into the Network section as “GRC Issuance Policy”. The panel shows mode (issue/hold/constrict), recommended ΔGRC, target supply, and a small treasury snapshot.


- `apps/workstation/src/App.tsx` — Added an `EpochControlPanel` under Network → “Epoch Control” that lets operators manually settle epochs via `/econ/settle-mainnet-epoch`, inspect supply/equity snapshots, and (DevNet-only) enable timed auto-settlement. The panel also shows a compact Δ summary for the last epoch along with an expandable verbose before/after JSON view.


- `internal/econ/mainnet_state.go` — Introduced `EpochHistoryEntry`, a rolling in-memory history buffer (`mainnetHistory`) capped at 5,000 epochs, and wired `SettleMainnetEpochBasic` to append compact per-epoch entries (supply, equity, reserves, GRC policy output).
- `rpc/econ/econ_handlers.go` — Added `/econ/history` HTTP endpoint (with optional `?limit=`) to expose the epoch history window for Workstation and analytics clients.
- `apps/workstation/src/App.tsx` — Added `SupplyHistoryPanel`, `EquityHistoryPanel`, and `GRCPolicyHistoryPanel` under Network, each using `/econ/history` and a lightweight inline SVG `HistoryChart` to visualise supply, equity/coverage, and GRC policy deltas over time.


- `internal/econ/assets_extra.go` — Introduced `AssetSTETH` as an additional `CryptoAssetKind` for LST-backed reserves.
- `internal/econ/reserve_mainnet_basket.go` — Added `MainnetReserveBasketConfig` and `MainnetReserveAssetConfig` plus `DefaultMainnetReserveBasketConfig()` encoding the Phase A basket (USDC, ETH, WBTC, STETH) and a helper `ComputeHaircuttedReserveUSDFromMainnetState` to compute raw vs haircutted reserve USD value from `MainnetMonetaryState`.


- `internal/econ/coverage_mainnet.go` — Added `MainnetCoverageSnapshot`, `AssetCoverageBreakdown`, and helpers to compute a coverage snapshot (raw vs effective reserves, coverage multiple, stable share, per-asset breakdown) from `MainnetMonetaryState` using the configured crypto-only reserve basket.
- `rpc/econ/econ_handlers.go` — Introduced `/econ/coverage` HTTP endpoint exposing the coverage snapshot for Workstation / operator use.
- `apps/workstation/src/App.tsx` — Added a new “Reserve Coverage” panel under Network, backed by `/econ/coverage`, showing USDR supply, raw/effective reserves, coverage multiple, stable share, and a per-asset breakdown table.


- `internal/econ/staking_mainnet.go` — Introduces PoS+PoP scaffolding for mainnet: consensus mode enum, node capability / participation / cost profiles, epoch reward split helper, and PoP reward distribution helper.
- `docs/ECON_PoP_Operator_Rewards.md` — High‑level design doc describing the PoP operator reward model, cost coverage targets, and the PoW → PoS+PoP transition framing for RSX, USDR, and GRC.


- Updated `internal/econ/mainnet_state.go` to record per-epoch reward splits for RSX stakers vs PoP operators, including the PoP stable vs volatile (USDR vs GRC) composition used by the operator reward model.
- Extended Workstation PoP rewards panel to surface the mixed USDR/GRC operator payout composition using fields from `/econ/mainnet-history`.


- `internal/econ/mainnet_state.go` — Extended EpochHistoryEntry with PoP payout composition (stable vs volatile) and wired a default 70/30 USD split for operator rewards into the epoch settlement path.
- `apps/workstation/src/App.tsx` — Enhanced the Network → Staking (RSX) and Operator PoP panels to surface epoch-level staking vs PoP reward shares, and PoP stable (USDR) vs volatile (GRC) reward portions.
- `docs/ECON_PoP_Operator_Rewards.md` — Documented the mixed USDR+GRC payout design for PoP rewards and how these show up in the epoch history metrics.
