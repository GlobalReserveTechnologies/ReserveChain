# ReserveChain Devnet — Changelog

## v1.0.0
- Initial devnet snapshot with marketing site, workstation, Go node, runtime, and database schema.
- This version matches the backup zip `ReserveChain-Devnet-v1.0.0.zip`.

## v1.0.1
- Added friendly `/site` and `/workstation` entrypoints that delegate to the existing `public` structure.
- Generated a high-level `FILE_INDEX.md` to document the repository layout.
- Prepared the tree for a React + TypeScript workstation SPA under `/apps/workstation` without changing current behavior.


## v2.0.0-pre1
- Scaffolded React + TypeScript SPA shell for the workstation under `apps/workstation/`.
- Sidebar navigation is generated from the existing PHP workstation sidebar, including Vaults, Trading, DeFi, Privacy, Network, Node Operator, and full Account (including Tier & Billing).
- Wired a TailwindCSS + Radix-ready shell with placeholder panels that mirror each current workstation panel.


## v2.0.0-pre2
- Reorganized workstation SPA sidebar into Finance, Network, Vaults, Privacy, and Account sections.
- Introduced Finance subgroups (Trading, Account, DeFi) with collapsible state persisted across sessions.
- Removed Operator section from the workstation; a dedicated `Reserve Operator` app will handle node/operator UX later.
- Applied an "extra dim" dark theme variant to the workstation SPA shell.


## v2.0.0-pre3
- Introduced initial Cyber-Sovereign workstation styling: panel chrome, extra-dim theme tokens, and semantic status pills.
- Updated SPA panel placeholder to use the new workstation panel shell layout.


## v2.0.0-pre4
- Added a floating wallet popup button and modal overlay for a MetaMask-style wallet experience inside the workstation.
- Introduced `wallet_popup.js` and a small `reserveWalletApi` wrapper around the existing browser keystore for use by the trading terminal and other modules.
- Upgraded the Network → Risk Dashboard panel with summary cards, placeholder charts, and vault / instrument risk tables.
- Added a simple `risk_summary.php` API stub so the Risk Dashboard has a backend entry point to evolve into the full risk engine.
- Extended `workstation.css` with wallet + risk UI styles while preserving the existing layout and behavior.


## v2.0.0-pre5
- Enhanced `public/api/risk_summary.php` to query the local Go node (`/api/chain/head` and `/api/chain/mempool`) and adjust risk level based on mempool load, with safe fallbacks when the node is offline.
- Updated `public/assets/js/workstation_risk_dashboard.js` to display chain height and mempool length in the Network Risk card and fixed the async fetch bug in `loadGlobal()`.

## v2.0.0-pre6
- Trading terminal now auto-connects to DevNet WebSocket feed for live NAV/windows/events.
- Workstation wallet popup can read live balances via new wallet_chain_client.js.
- Fixed syntax issue in workstation_risk_dashboard.js that could break the Risk Dashboard view.

## v2.0.0-pre7
- Added dedicated Windows launch scripts for Node 1/2/3, Website, and All services.
- Scripts now resolve Go and PHP from the runtime\ folder directly (no env var requirements).

## v2.0.0-pre8
- Restructured devnet configuration into a richer, annotated YAML schema (node, http, p2p, storage, windows, fx, yield, rewards, economics, privacy, risk, workstation, operator).
- Introduced forward-looking P2P seed/peer settings in config (mode, seed_nodes, max_peers, allow_external) for future multi-node discovery.
- Cleaned and consolidated launcher scripts into scripts/ with improved headers and error handling:
  - scripts/start_node.bat, scripts/start_website.bat, scripts/start_all.bat
  - scripts/start_node.sh, scripts/start_website.sh, scripts/start_all.sh


## v2.0.0-pre36
- Introduced a DevNet treasury balance sheet model under `internal/econ/treasury_balance.go` with mark-to-market valuation of the crypto reserve basket and coverage metrics for USDR and GRC.
- Wired `econ.InitStateForDevnet()` to also seed a conservative synthetic treasury state so Workstation / Operator Console have meaningful Treasury data from first tick.
- Exposed a new `/econ/treasury` HTTP endpoint via `rpc/econ/econ_handlers.go` that returns a JSON-encoded `TreasuryBalanceSheet` snapshot for future workstation dashboards.


## v2.0.0-pre37
- Wired the new DevNet treasury balance sheet model into the Workstation SPA via a dedicated Treasury / Reserves panel backed by `/econ/treasury`.
- Introduced an initial DevNet mint / redeem queue + epoch model in `internal/econ/redemptions.go`, including enqueue, snapshot, and epoch-advance helpers tied into the treasury pending redemptions + USDR supply.
- Extended the econ RPC surface with `/econ/redemptions` for future Operator Console / Workstation integrations.


## v2.0.0-pre38
- Extended the Workstation Network section with a Redemption Queue panel backed by `/econ/redemptions`, showing current epoch, aggregate pending USDR, and a per-request table view for DevNet redemptions.


## v2.0.0-pre39
- Introduced a DevNet mint queue in `internal/econ/mints.go` for USDR issuance, supporting both crypto-backed mints and free test mints, settled at epoch boundaries alongside redemptions.
- Updated `AdvanceDevnetEpoch` to settle mint requests before burning redemptions so mint/redeem flows share a single epoch-based settlement cycle.
- Extended the econ RPC adapter with `/econ/mints` and a DevNet-only `/econ/advance-epoch` endpoint.
- Enhanced the Workstation Network section by removing the unused Governance panel and adding an operator-facing “Force Epoch Advance” control on the Redemption Queue panel.


## v2.0.0-pre40
- Extended the Workstation Network section with a Mint Queue panel backed by `/econ/mints`, visualising total USDR to be minted, current epoch, and per-request crypto-backed + test mint entries.


## v2.0.0-pre41
- Introduced a DevNet GRC issuance signal model in `internal/econ/grc_issuance.go`, generating coverage- and equity-driven `GRCIssuanceSignal` recommendations from the live treasury snapshot.
- Extended the econ RPC surface with `/econ/grc-issuance` so Workstation / Operator Console UIs can later visualise and experiment with GRC issuance policy without actually changing on-ledger supply.


## v2.0.0-pre42
- Started the mainnet-first economics refactor by adding `ledger_mainnet.go` with a canonical `MainnetMonetaryState` representation (supply, reserves, liabilities, equity, pending actions) and a helper to derive the existing `TreasuryBalanceSheet` view from this state. New issuance / mint / redeem logic will build on this structure instead of the DevNet-only helpers.


## v2.0.0-pre43
- Added mainnet-style transition helpers on `MainnetMonetaryState` for registering pending USDR/GRC actions and a conservative `SettleEpoch` implementation that applies those actions and advances the epoch. This formalizes the epoch-final settlement semantics for mint/redeem/issuance logic going forward.


## v2.0.0-pre44
- Introduced `policy_mainnet.go` with a conservative USDR policy helper (`ApplyBasicUSDRPolicy`) and `BasicPolicyConfig`. This policy enforces a minimum reserve coverage threshold on USDR mints and caps redemptions to the available supply before calling `SettleEpoch`, formalizing basic solvency constraints for mainnet USDR flows.


## v2.0.0-pre45
- Introduced an in-memory mainnet monetary state singleton and wired it into the RPC layer via `/econ/mainnet-state` and `/econ/settle-mainnet-epoch`. These endpoints exercise the new mainnet ledger + policy engine, applying basic USDR coverage rules and epoch settlement through a simple operator-facing API.


## v2.0.0-pre46
- Extended the econ RPC adapter with `/econ/mint-usdr` and `/econ/redeem-usdr` endpoints. These endpoints register pending USDR mints/redemptions on the mainnet ledger and return the updated `MainnetMonetaryState`. Actual supply changes only occur when `/econ/settle-mainnet-epoch` is invoked, aligning the HTTP surface with the epoch-based mainnet monetary engine.


## v2.0.0-pre47
- Added mainnet-style GRC issuance and burn endpoints (`/econ/issue-grc`, `/econ/burn-grc`) that register pending GRC supply adjustments on `MainnetMonetaryState`. Combined with `/econ/settle-mainnet-epoch`, this completes the basic mainnet monetary control surface for both USDR and GRC.


## v2.0.0-pre48
- Implemented a hybrid mainnet GRC issuance policy that combines USDR coverage, treasury equity, and net USDR demand into a single issuance signal. The policy enqueues GRC issuance/burn actions before epoch settlement, and `/econ/grc-issuance` now reports this mainnet-oriented signal instead of the legacy DevNet heuristic.


## v2.0.0-pre49
- Extended the Workstation UI with a GRC Issuance Policy panel under Network → “GRC Issuance Policy”. This panel reads from `/econ/grc-issuance` and surfaces the hybrid mainnet GRC policy (mode, delta, target supply, and treasury snapshot) so operators can see what the economics brain is recommending each epoch.


## v2.0.0-pre50
- Extended the Workstation with an Epoch Control console (Network → “Epoch Control”). Operators can now manually trigger mainnet epoch settlement, view a compact summary of the last epoch’s ΔUSDR/ΔGRC/ΔEquity, and optionally enable a DevNet-only auto-settle loop. A verbose mode exposes before/after JSON snapshots for tuning and auditing the economic policy.


## v2.0.0-pre51
- Added an in-memory mainnet epoch history buffer (up to 5,000 epochs) and `/econ/history` RPC to expose compact per-epoch snapshots (supply, reserves, equity, and GRC policy output).
- Extended the Workstation Network section with three analytics panels—“Supply Over Time”, “Equity Over Time”, and “GRC Policy Output”—each reading from `/econ/history` and rendering simple inline SVG charts suitable for DevNet/Testnet policy tuning.


## v2.0.0-pre52
- Defined a first-pass crypto-only reserve basket for mainnet (USDC, ETH, WBTC, STETH) via `MainnetReserveBasketConfig`, including simple haircuts and soft concentration caps per asset role.
- Added `ComputeHaircuttedReserveUSDFromMainnetState` to translate the in-memory mainnet reserve pools into raw vs haircutted USD value using the basket config, ready for integration into USDR coverage checks, corridor logic, and analytics.


## v2.0.0-pre53
- Implemented a crypto-only reserve coverage snapshot helper and exposed it via `/econ/coverage`, including raw vs haircutted reserves, USDR supply, coverage multiple, stable share, and per-asset composition derived from the mainnet monetary state and basket config.
- Extended the Workstation Network section with a dedicated “Reserve Coverage” panel that consumes `/econ/coverage` and surfaces coverage health plus composition for operator and analytics workflows.


## v2.0.0-pre54
- Added initial PoS+PoP economics scaffolding for mainnet, including data structures for node capability, participation, and cost profiles plus helpers to split epoch rewards between RSX stakers and PoP operator work.
- Documented the PoP operator reward model (work‑based payouts, node cost coverage targets, and the PoW to PoS+PoP transition) in `docs/ECON_PoP_Operator_Rewards.md`.


## v2.0.0-pre56
- Wired the PoP operator reward currency mix into the mainnet epoch engine by recording a 70/30 USDR/GRC split (by USD value) for the PoP reward portion in `EpochHistoryEntry`.
- Updated the PoP operator rewards workstation panel to display the stable (USDR) and volatile (GRC) portions of the latest epoch's PoP pool, while keeping per-node allocation wiring as a later step.


## v2.0.0-pre56
- Wired epoch settlement into the new PoS+PoP economics layer by splitting each epoch's aggregate reward pool into RSX staking and PoP operator portions and recording the split in mainnet history.
- Introduced a policy-level 70/30 PoP payout mix between stable (USDR) and volatile (GRC) operator rewards, with the composition exposed via EpochHistoryEntry fields and the workstation PoP panel.
- Extended the workstation's Network section with dedicated Staking (RSX) and Operator PoP panels that read from `/econ/mainnet-history` and display the latest epoch reward splits.
