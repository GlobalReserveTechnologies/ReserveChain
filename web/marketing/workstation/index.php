<?php ?>
<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <title>ReserveChain Workstation</title>
  <meta name="viewport" content="width=device-width,initial-scale=1" />
  <link rel="stylesheet" href="/assets/css/workstation.css">
  <script src="/assets/js/workstation_explorer.js"></script>
  <script src="/assets/js/wallet_chain.js"></script>
</head>
<body class="ws-body">

<button id="wallet-popup-toggle" class="wallet-floating-btn">Wallet</button>

<div class="ws-root">
  <!-- Sidebar -->
  <aside class="ws-sidebar">
    <div class="ws-sidebar-logo">
      <span class="logo-dot"></span>
      <span class="logo-text">ReserveChain<span class="logo-sub">Workstation</span></span>
    </div>

    <div class="ws-sidebar-section">
      <button class="ws-sidebar-group" data-group="vaults">
        <span class="chevron">▾</span>
        <span class="label">Vaults</span>
      </button>

      <div class="ws-sidebar-submenu" data-group="vaults">
        <button class="ws-sidebar-item is-active" data-panel="vault-dashboard">
          Vault Dashboard
        </button>
        <button class="ws-sidebar-item" data-panel="vault-stealth">
          Stealth Addresses
        </button>
        <button class="ws-sidebar-item" data-panel="vault-policies">
          Policies &amp; Approvals
        </button>
        <button class="ws-sidebar-item" data-panel="vault-roles">
          Roles &amp; Permissions
        </button>
        <button class="ws-sidebar-item" data-panel="vault-analytics">
          Analytics
        </button>
        <button class="ws-sidebar-item" data-panel="vault-audit">
          Audit &amp; Reports
        </button>
        <button class="ws-sidebar-item" data-panel="vault-settings">
          Settings
        </button>
      </div>
    </div>

    
<div class="ws-sidebar-section">
  <div class="ws-sidebar-heading">TRADING</div>
  <button class="ws-sidebar-item" data-panel="trading-swap">
    Swap
  </button>
  <button class="ws-sidebar-item" data-panel="trading-terminal">
    Trading Terminal
  </button>
</div>

<div class="ws-sidebar-section">
  <div class="ws-sidebar-heading">DECENTRALIZED FINANCE</div>
  <button class="ws-sidebar-item" data-panel="defi-lending">
    Lending
  </button>
  <button class="ws-sidebar-item" data-panel="defi-portfolio">
    Portfolio
  </button>
  <button class="ws-sidebar-item" data-panel="defi-wallet">
    Wallet
  </button>
  <button class="ws-sidebar-item" data-panel="defi-orders-fills">
    Orders &amp; Fills
  </button>
  <button class="ws-sidebar-item" data-panel="defi-transfers">
    Transfers
  </button>
  <button class="ws-sidebar-item" data-panel="defi-deposit-withdraw">
    Deposit / Withdraw
  </button>
  <button class="ws-sidebar-item" data-panel="defi-staking">
    Staking
  </button>
  <button class="ws-sidebar-item" data-panel="defi-liquidity">
    Liquidity
  </button>
  <button class="ws-sidebar-item" data-panel="defi-activity">
    Activity
  </button>
</div>

<div class="ws-sidebar-section">
  <div class="ws-sidebar-heading">PRIVACY</div>
  <button class="ws-sidebar-item" data-panel="vault-dashboard">
    Vault
  </button>
  <button class="ws-sidebar-item" data-panel="privacy-private-amm">
    Private AMM
  </button>
  <button class="ws-sidebar-item" data-panel="privacy-private-lending">
    Private Lending
  </button>
  <button class="ws-sidebar-item" data-panel="privacy-audit-proofs">
    Audit Proofs
  </button>
</div>

<div class="ws-sidebar-section">
  <div class="ws-sidebar-heading">NETWORK</div>
  <button class="ws-sidebar-item" data-panel="network-explorer">
    Block Explorer
  </button>
  <button class="ws-sidebar-item" data-panel="treasury-overview">
    Treasury
  </button>
  <button class="ws-sidebar-item" data-panel="network-epochs">
    Epochs
  </button>
  <button class="ws-sidebar-item" data-panel="network-governance">
    Governance
  </button>
  <button class="ws-sidebar-item" data-panel="network-analytics">
    Analytics
  </button>
  <button class="ws-sidebar-item" data-panel="network-risk-dashboard">
    Risk Dashboard
  </button>
  <button class="ws-sidebar-item" data-panel="network-benchmarks">
    Benchmarks
  </button>
  <button class="ws-sidebar-item" data-panel="network-health">
    Health
  </button>
</div>

<div class="ws-sidebar-section">
  <div class="ws-sidebar-heading">NODE OPERATOR</div>
  <button class="ws-sidebar-item" data-panel="node-governance">
    Governance
  </button>
</div>

<div class="ws-sidebar-section">
  <div class="ws-sidebar-heading">ACCOUNT</div>
  <button class="ws-sidebar-item" data-panel="account-profile">
    Profile &amp; Settings
  </button>
  <button class="ws-sidebar-item" data-panel="account-api-keys">
    API Keys
  </button>
  <button class="ws-sidebar-item" data-panel="account-security">
    Security
  </button>
  <button class="ws-sidebar-item" data-panel="account-tier-billing">
    Tier &amp; Billing
  </button>
</div>
</div>
  </aside>

  <!-- Main content -->
  <main class="ws-main">
    <!-- Vault Dashboard panel -->
    <section class="ws-panel" data-panel="vault-dashboard">
      <header class="ws-panel-header">
        <div>
          <h1>Vault Dashboard</h1>
          <p class="ws-panel-subtitle">Private banking layer on top of the ReserveChain ledger.</p>
          <p>Overview of vault balances, flows, and approvals.</p>
        </div>
        <div class="ws-panel-actions">
          <button id="btn-vault-new" class="btn btn-primary">New Vault</button>
          <button id="btn-vault-transfer" class="btn btn-secondary">Transfer</button>
          <button id="btn-vault-deposit" class="btn btn-secondary">Deposit</button>
          <button id="btn-vault-withdraw" class="btn btn-secondary">Withdraw</button>
        </div>
      </header>

      <div class="vault-summary-grid">
        <div class="vault-card">
          <div class="vault-card-label">Total Vault Assets</div>
          <div class="vault-card-value" id="vault-total-assets">–</div>
        </div>
        <div class="vault-card">
          <div class="vault-card-label">Number of Vaults</div>
          <div class="vault-card-value" id="vault-count">0</div>
        </div>
        <div class="vault-card">
          <div class="vault-card-label">Pending Approvals</div>
          <div class="vault-card-value" id="vault-pending-approvals">0</div>
        </div>
        <div class="vault-card">
          <div class="vault-card-label">24h Net Flow</div>
          <div class="vault-card-value" id="vault-net-flow">–</div>
        </div>
      </div>

      <section class="vault-section">
        <div class="vault-section-header">
          <h2>Vaults</h2>
          <button class="link-button" id="btn-vault-view-analytics" data-panel-jump="vault-analytics">
            View analytics
          </button>
        </div>
        <table class="vault-table">
          <thead>
          <tr>
            <th>Vault</th>
            <th>Type</th>
            <th>Balance (USD-eq / GRC)</th>
            <th>PnL Mode</th>
            <th>Duration</th>
            <th>Visibility</th>
            <th>Actions</th>
          </tr>
          </thead>
          <tbody id="vault-table-body"></tbody>
        </table>
      </section>

      <div class="vault-two-col">
        <section class="vault-section">
          <div class="vault-section-header">
            <h2>Pending Approvals</h2>
            <button class="link-button" data-panel-jump="vault-policies">View all</button>
          </div>
          <div id="vault-approvals-list" class="vault-list-placeholder">
            No pending approvals.
          </div>
        </section>

        <section class="vault-section">
          <div class="vault-section-header">
            <h2>Recent Activity</h2>
            <button class="link-button" data-panel-jump="vault-audit">View ledger</button>
          </div>
          <div id="vault-activity-list" class="vault-list-placeholder">
            No recent activity.
          </div>
        </section>
      </div>
    </section>

    <!-- Stealth Addresses Panel -->
    <section class="ws-panel" data-panel="vault-stealth" style="display:none;">
      <header class="ws-panel-header">
        <div>
          <h1>Stealth Addresses</h1>
          <p>Generate and manage stealth receive addresses per vault.</p>
        </div>
      </header>

      <div class="vault-stealth-controls">
        <label class="field">
          <span>Vault</span>
          <select id="stealth-vault-select"></select>
        </label>

        <label class="field">
          <span>Label</span>
          <input id="stealth-label-input" type="text" placeholder="e.g. Client A, OTC Desk" />
        </label>

        <button id="stealth-generate-btn" class="btn btn-primary">Generate Stealth Address</button>
      </div>

      <section class="vault-section">
        <div class="vault-section-header">
          <h2>Addresses for Selected Vault</h2>
        </div>
        <table class="vault-table">
          <thead>
          <tr>
            <th>Label</th>
            <th>Stealth Address</th>
            <th>Ephemeral Pubkey</th>
            <th>Active</th>
            <th>Created</th>
            <th>Last Used</th>
            <th>Actions</th>
          </tr>
          </thead>
          <tbody id="stealth-table-body"></tbody>
        </table>
      </section>

      <section class="vault-section">
        <div class="vault-section-header">
          <h2>Last Generated</h2>
        </div>
        <div id="stealth-last-generated" class="vault-list-placeholder">
          No stealth address generated yet.
        </div>
      </section>
    </section>

    <!-- Stub panels for now -->
    <section class="ws-panel" data-panel="vault-policies" style="display:none;">
      <header class="ws-panel-header">
        <div>
          <h1>Policies &amp; Approvals</h1>
          <p>Configure vault policies and review pending approvals.</p>
        </div>
      </header>
      <div id="vault-policies-root" class="vault-list-placeholder">
        TODO: hook to policy and approvals APIs.
      </div>
    </section>

    <section class="ws-panel" data-panel="vault-roles" style="display:none;">
      <header class="ws-panel-header">
        <div>
          <h1>Roles &amp; Permissions</h1>
          <p>Manage who can trade, withdraw, approve, or audit per vault.</p>
        </div>
      </header>
      <div id="vault-roles-root" class="vault-list-placeholder">
        TODO: RBAC matrix UI &amp; API endpoints.
      </div>
    </section>

    <section class="ws-panel" data-panel="vault-analytics" style="display:none;">
      <header class="ws-panel-header">
        <div>
          <h1>Vault Analytics</h1>
          <p>PNL, flows, allocations and exposures across all vaults.</p>
        </div>
      </header>
      <div id="vault-analytics-root" class="vault-list-placeholder">
        TODO: charts (PNL over time, allocations) using aggregated vault data.
      </div>
    </section>

    <section class="ws-panel" data-panel="vault-audit" style="display:none;">
      <header class="ws-panel-header">
        <div>
          <h1>Audit &amp; Reports</h1>
          <p>Export vault ledgers and generate audit views.</p>
        </div>
      </header>
      <div id="vault-audit-root" class="vault-list-placeholder">
        TODO: export buttons + audit link generator.
      </div>
    </section>

    <section class="ws-panel" data-panel="vault-settings" style="display:none;">
      <header class="ws-panel-header">
        <div>
          <h1>Vault Settings</h1>
          <p>Configure labels, visibility modes, and PnL settlement modes.</p>
        </div>
      </header>
      <div id="vault-settings-root" class="vault-list-placeholder">
        <!-- Settings UI rendered by workstation_vault_settings.js -->
      </div>
    </section>

    <section class="ws-panel" data-panel="treasury-overview" style="display:none;">
      <header class="ws-panel-header">
        <div>
          <h1>Treasury &amp; Reserves</h1>
          <p>Overview of reserve backing, coverage ratios, and duration profiles.</p>
        </div>
      </header>
      <div class="vault-summary-grid">
        <div class="vault-card">
          <div class="vault-card-label">Total Reserve (simulated)</div>
          <div class="vault-card-value" id="treasury-total-reserve">–</div>
        </div>
        <div class="vault-card">
          <div class="vault-card-label">Coverage Ratio</div>
          <div class="vault-card-value" id="treasury-coverage-ratio">–</div>
        </div>
        <div class="vault-card">
          <div class="vault-card-label">Duration Split</div>
          <div class="vault-card-value" id="treasury-duration-split">–</div>
        </div>
        <div class="vault-card">
          <div class="vault-card-label">Yield Sources</div>
          <div class="vault-card-value" id="treasury-yield-sources">–</div>
        </div>
      </div>
      <div class="vault-two-col">
        <section class="vault-section">
          <div class="vault-section-header">
            <h2>Reserve Notes</h2>
          </div>
          <div id="treasury-notes" class="vault-list-placeholder">
            Treasury and reserve wiring will be connected to the Go engine in the next phase. For now, this view summarizes simulated values based on vault balances and duration tiers.
          </div>
        </section>
        <section class="vault-section">
          <div class="vault-section-header">
            <h2>Upcoming</h2>
          </div>
          <div class="vault-list-placeholder">
            Planned: reserve tier breakdown, external vs internal yield composition, and redemption queue visibility.
          </div>
        </section>
      </div>
    </section>
  
    <section class="ws-panel" data-panel="account-tier-billing" style="display:none;">
      <header class="ws-panel-header">
        <div>
          <h1>Tier &amp; Billing</h1>
          <p>Manage your Reserve account tier, Earn credits, and renewal preferences.</p>
        </div>
      </header>

      <div class="ws-grid ws-grid-2">
        <!-- Tier card -->
        <div class="card tier-card" id="tier-card-root">
          <div class="card-header">
            <div>
              <h2 id="tier-card-title">Core Reserve</h2>
              <p id="tier-card-subtitle">Baseline coverage and access.</p>
            </div>
            <div class="badge badge-soft" id="tier-card-status">Active</div>
          </div>

          <div class="tier-card-pill-row">
            <span class="pill" id="tier-pill-billing">Yearly billing</span>
            <span class="pill" id="tier-pill-renewal">Renews in — days</span>
          </div>

          <div class="tier-card-grid">
            <div class="tier-stat">
              <div class="tier-stat-label">Margin Limit</div>
              <div class="tier-stat-value" id="tier-stat-margin">×2</div>
            </div>
            <div class="tier-stat">
              <div class="tier-stat-label">PoP Multiplier</div>
              <div class="tier-stat-value" id="tier-stat-pop">1.0×</div>
            </div>
            <div class="tier-stat">
              <div class="tier-stat-label">Staking Multiplier</div>
              <div class="tier-stat-value" id="tier-stat-stake">1.0×</div>
            </div>
            <div class="tier-stat">
              <div class="tier-stat-label">Haircut</div>
              <div class="tier-stat-value" id="tier-stat-haircut">80%</div>
            </div>
          </div>

          <div class="tier-pricing">
            <div>
              <div class="tier-pricing-label">Current plan</div>
              <div class="tier-pricing-main">
                <span id="tier-pricing-grc">0 GRC</span>
                <span class="tier-pricing-usd" id="tier-pricing-usd">≈ $0.00</span>
              </div>
              <div class="tier-pricing-sub" id="tier-pricing-sub">Includes reserved Earn benefits and corridor coverage.</div>
            </div>
          </div>

          <div class="tier-actions">
            <button class="button primary" id="btn-tier-upgrade">
              Upgrade account
            </button>
            <button class="button ghost" id="btn-tier-refresh">
              Refresh
            </button>
          </div>
        </div>

        <!-- Earn card -->
        <div class="card tier-card" id="earn-card-root">
          <div class="card-header">
            <div>
              <h2>Earn Credits</h2>
              <p>Surplus value you can apply toward renewals or additional time.</p>
            </div>
            <div class="badge badge-soft" id="earn-card-badge">DevNet</div>
          </div>

          <div class="tier-card-main-balance">
            <div class="tier-earn-balance">
              <div class="tier-earn-label">Balance</div>
              <div class="tier-earn-value">
                <span id="earn-balance-grc">0 GRC</span>
                <span class="tier-earn-usd" id="earn-balance-usd">≈ $0.00</span>
              </div>
            </div>
          </div>

          <div class="tier-card-grid">
            <div class="tier-stat">
              <div class="tier-stat-label">Capital Earn</div>
              <div class="tier-stat-value" id="earn-capital">0</div>
            </div>
            <div class="tier-stat">
              <div class="tier-stat-label">Flow Earn</div>
              <div class="tier-stat-value" id="earn-flow">0</div>
            </div>
            <div class="tier-stat">
              <div class="tier-stat-label">Risk Earn</div>
              <div class="tier-stat-value" id="earn-risk">0</div>
            </div>
            <div class="tier-stat">
              <div class="tier-stat-label">Bonus</div>
              <div class="tier-stat-value" id="earn-bonus">0</div>
            </div>
          </div>

          <div class="tier-pricing">
            <div class="tier-pricing-sub">
              Surplus Earn can be converted to additional renewal time (S3) once your base plan cost is covered.
            </div>
          </div>
        </div>
      </div>

      <!-- Upgrade modal -->
      <div class="ws-modal-backdrop" id="tier-upgrade-modal" style="display:none;">
        <div class="ws-modal">
          <div class="ws-modal-header">
            <h2>Upgrade account</h2>
            <button class="ws-modal-close" id="tier-upgrade-close">&times;</button>
          </div>
          <div class="ws-modal-body">
            <div class="tier-modal-grid">
              <div class="tier-modal-column">
                <h3>Choose plan</h3>
                <div class="tier-plan-list">
                  <button class="tier-plan-option" data-tier="core">
                    <span class="tier-plan-name">Core Reserve</span>
                    <span class="tier-plan-tag">Included</span>
                  </button>
                  <button class="tier-plan-option" data-tier="elite">
                    <span class="tier-plan-name">Elite Reserve</span>
                    <span class="tier-plan-tag">More coverage</span>
                  </button>
                  <button class="tier-plan-option" data-tier="executive">
                    <span class="tier-plan-name">Executive Reserve</span>
                    <span class="tier-plan-tag">High throughput</span>
                  </button>
                  <button class="tier-plan-option tier-plan-featured" data-tier="express">
                    <span class="tier-plan-name">Express Reserve</span>
                    <span class="tier-plan-tag">Maximum benefits</span>
                  </button>
                </div>

                <div class="tier-modal-field-group">
                  <label>Billing cycle</label>
                  <div class="chip-group" id="tier-billing-group">
                    <button class="chip is-active" data-billing="monthly">Monthly</button>
                    <button class="chip" data-billing="yearly">Yearly</button>
                  </div>
                </div>

                <div class="tier-modal-field-group">
                  <label>Pay from</label>
                  <div class="chip-group" id="tier-source-group">
                    <button class="chip is-active" data-source="vault">Vault</button>
                    <button class="chip" data-source="hot">Hot Wallet</button>
                    <button class="chip" data-source="stake">Stake</button>
                  </div>
                </div>
              </div>

              <div class="tier-modal-column">
                <h3>Quote</h3>
                <div id="tier-quote-body" class="tier-quote-body">
                  <p class="tier-quote-placeholder">
                    Select a plan and billing cycle to see a live quote.
                  </p>
                </div>
              </div>
            </div>
          </div>
          <div class="ws-modal-footer">
            <button class="button ghost" id="tier-upgrade-cancel">Cancel</button>
            <button class="button primary" id="tier-upgrade-confirm" disabled>Sign &amp; Upgrade</button>
          </div>
        </div>
      </div>
    </section>

  </main>
</div>


<!-- SPA Stub Panels for new navigation (to be fully wired later) -->
<section class="ws-panel" data-panel="trading-swap" style="display:none;">
  <header class="ws-panel-header">
    <div>
      <h1>Swap</h1>
      <p>Hybrid swap UI will route through your corridor model and vault routing.</p>
    </div>
  </header>
  <div class="vault-list-placeholder">
    Swap terminal wiring pending engine RPC integration.
  </div>
</section>

<section class="ws-panel" data-panel="trading-terminal" style="display:none;">
  <header class="ws-panel-header">
    <div>
      <h1>Trading Terminal</h1>
      <p>Launch the full-screen trading terminal in a popup window.</p>
    </div>
    <div class="ws-panel-actions">
      <button class="btn btn-primary" id="btn-open-terminal-popup">Open Trading Terminal</button>
      <button class="btn btn-secondary" id="btn-open-terminal-multi">Open Multi-Panel Layout</button>
    </div>
  </header>
  <div class="vault-list-placeholder">
    Use the buttons above to open the dedicated trading windows while keeping this workstation as your control plane.
  </div>
</section>

<section class="ws-panel" data-panel="defi-lending" style="display:none;">
  <header class="ws-panel-header">
    <h1>Lending</h1>
  </header>
  <div class="vault-list-placeholder">
    DeFi lending flows will plug into ReserveNet vaults and on-chain positions.
  </div>
</section>

<section class="ws-panel" data-panel="defi-portfolio" style="display:none;">
  <header class="ws-panel-header">
    <div>
      <h1>Portfolio</h1>
      <p class="ws-panel-subtitle">Live balances, vault collateral, and open positions across ReserveNet.</p>
    </div>
    <div class="ws-panel-actions">
      <button class="btn btn-secondary" id="btn-portfolio-refresh">Refresh</button>
    </div>
  </header>

  <div class="ws-grid ws-grid-3">
    <div class="card metric-card">
      <div class="metric-label">Net Portfolio Value</div>
      <div class="metric-value" id="pf-net-value">—</div>
      <div class="metric-sub">Across wallets, vaults, and margin positions</div>
    </div>

    <div class="card metric-card">
      <div class="metric-label">On-Chain Balance</div>
      <div class="metric-value" id="pf-onchain-balance">—</div>
      <div class="metric-sub">Direct L1 wallet &amp; vault balances (GRC)</div>
    </div>

    <div class="card metric-card">
      <div class="metric-label">Open PnL</div>
      <div class="metric-value" id="pf-open-pnl">—</div>
      <div class="metric-sub" id="pf-open-pnl-sub">Unrealized PnL across terminal positions</div>
    </div>
  </div>

  <div class="ws-grid ws-grid-2">
    <div class="card">
      <div class="card-header">
        <h2>Positions</h2>
        <span class="badge" id="pf-positions-count">0 open</span>
      </div>
      <table class="table table-compact">
        <thead>
          <tr>
            <th>Symbol</th>
            <th>Side</th>
            <th>Size</th>
            <th>Entry</th>
            <th>Mark</th>
            <th>LiQ</th>
            <th>PnL</th>
          </tr>
        </thead>
        <tbody id="pf-positions-body">
          <tr><td colspan="7">No open positions.</td></tr>
        </tbody>
      </table>
    </div>

    <div class="card">
      <div class="card-header">
        <h2>Balances</h2>
      </div>
      <table class="table table-compact">
        <thead>
          <tr>
            <th>Source</th>
            <th>Asset</th>
            <th>Amount</th>
          </tr>
        </thead>
        <tbody id="pf-balances-body">
          <tr><td colspan="3">Loading…</td></tr>
        </tbody>
      </table>
    </div>
  </div>

  <div class="card">
    <div class="card-header">
      <h2>Recent Activity</h2>
    </div>
    <table class="table table-compact">
      <thead>
        <tr>
          <th>Time</th>
          <th>Type</th>
          <th>Details</th>
          <th>Delta</th>
        </tr>
      </thead>
      <tbody id="pf-activity-body">
        <tr><td colspan="4">No recent activity.</td></tr>
      </tbody>
    </table>
  </div>
</section>


<section class="ws-panel" data-panel="defi-wallet" style="display:none;">
  <header class="ws-panel-header">
    <h1>Wallet</h1>
  </header>
  <div class="vault-list-placeholder">
    Wallet overview and quick actions (send, receive, sign) will be wired here.
  </div>
</section>

<section class="ws-panel" data-panel="defi-orders-fills" style="display:none;">
  <header class="ws-panel-header">
    <h1>Orders &amp; Fills</h1>
  </header>
  <div class="vault-list-placeholder">
    Order and fill history across venues will be surfaced here.
  </div>
</section>

<section class="ws-panel" data-panel="defi-transfers" style="display:none;">
  <header class="ws-panel-header">
    <h1>Transfers</h1>
  </header>
  <div class="vault-list-placeholder">
    Internal vault-to-vault transfers and on-chain sends will be managed here.
  </div>
</section>

<section class="ws-panel" data-panel="defi-deposit-withdraw" style="display:none;">
  <header class="ws-panel-header">
    <h1>Deposit / Withdraw</h1>
  </header>
  <div class="vault-list-placeholder">
    Fiat onramps / offramps and custody bridges (if any) will be wired here.
  </div>
</section>

<section class="ws-panel" data-panel="defi-staking" style="display:none;">
  <header class="ws-panel-header">
    <h1>Staking</h1>
  </header>
  <div class="vault-list-placeholder">
    Staking flows and governance staking will live here.
  </div>
</section>

<section class="ws-panel" data-panel="defi-liquidity" style="display:none;">
  <header class="ws-panel-header">
    <h1>Liquidity</h1>
  </header>
  <div class="vault-list-placeholder">
    AMM and orderbook liquidity positions will be surfaced here.
  </div>
</section>

<section class="ws-panel" data-panel="defi-activity" style="display:none;">
  <header class="ws-panel-header">
    <h1>Activity</h1>
  </header>
  <div class="vault-list-placeholder">
    Unified activity stream across wallet, vault, and trading actions.
  </div>
</section>

<section class="ws-panel" data-panel="privacy-private-amm" style="display:none;">
  <header class="ws-panel-header">
    <h1>Private AMM</h1>
  </header>
  <div class="vault-list-placeholder">
    Stealth pool routing and shielded swaps will be designed here.
  </div>
</section>

<section class="ws-panel" data-panel="privacy-private-lending" style="display:none;">
  <header class="ws-panel-header">
    <h1>Private Lending</h1>
  </header>
  <div class="vault-list-placeholder">
    Vault-backed anonymous credit flows will sit here once wired.
  </div>
</section>

<section class="ws-panel" data-panel="privacy-audit-proofs" style="display:none;">
  <header class="ws-panel-header">
    <h1>Audit Proofs</h1>
  </header>
  <div class="vault-list-placeholder">
    View proof trees and solvency attestations for private flows.
  </div>
</section>

<section class="ws-panel" data-panel="network-explorer" style="display:none;">
  <header class="ws-panel-header">
    <h1>Block Explorer</h1>
  </header>
  <section class="explorer-head">
    <div class="explorer-head-card">
      <div class="explorer-head-label">Chain Height</div>
      <div class="explorer-head-value" id="explorer-head-height">-</div>
      <div class="explorer-head-sub" id="explorer-head-hash">-</div>
    </div>
    <div class="explorer-head-card">
      <div class="explorer-head-label">Last Block</div>
      <div class="explorer-head-value" id="explorer-head-txtype">-</div>
      <div class="explorer-head-sub" id="explorer-head-ts">-</div>
    </div>
  </section>
  <section class="explorer-table-section">
    <header class="explorer-table-header">
      <h2>Recent Blocks</h2>
      <button class="btn btn-ghost" id="explorer-refresh">
        Refresh
      </button>
    </header>
    <div class="explorer-table-wrapper">
      <table class="explorer-table">
        <thead>
          <tr>
            <th>Height</th>
            <th>Hash</th>
            <th>Tx Type</th>
            <th>Timestamp</th>
          </tr>
        </thead>
        <tbody id="explorer-blocks-body">
          <tr>
            <td colspan="4" class="explorer-empty">No blocks loaded yet.</td>
          </tr>
        </tbody>
      </table>
    </div>
  </section>
</section>

<section class="ws-panel" data-panel="network-epochs" style="display:none;">
  <header class="ws-panel-header">
    <h1>Epochs</h1>
  </header>
  <div class="vault-list-placeholder">
    Epoch history and validator assignments will be listed here.
  </div>
</section>

<section class="ws-panel" data-panel="network-governance" style="display:none;">
  <header class="ws-panel-header">
    <h1>Network Governance</h1>
  </header>
  <div class="vault-list-placeholder">
    Governance proposals, votes, and timelocks surface here.
  </div>
</section>

<section class="ws-panel" data-panel="network-analytics" style="display:none;">
  <header class="ws-panel-header">
    <h1>Network Analytics</h1>
  </header>
  <div class="vault-list-placeholder">
    TPS, gas usage, and chain health metrics will be charted here.
  </div>
</section>

<section class="ws-panel" data-panel="network-risk-dashboard" style="display:none;">
  <header class="ws-panel-header">
    <div>
      <h1>Risk Dashboard</h1>
      <p>Network-wide risk posture, issuance corridor utilization, and your personal exposure.</p>
    </div>
  </header>

  <!-- Top summary cards -->
  <div class="risk-grid risk-grid-top">
    <div class="risk-card">
      <div class="risk-card-label">Network Risk Level</div>
      <div class="risk-card-value" id="risk-network-level">–</div>
      <div class="risk-card-meta" id="risk-network-meta">Awaiting data…</div>
    </div>
    <div class="risk-card">
      <div class="risk-card-label">Your Risk (This Wallet)</div>
      <div class="risk-card-value" id="risk-wallet-level">–</div>
      <div class="risk-card-meta" id="risk-wallet-meta">Connect a wallet to see account risk.</div>
    </div>
    <div class="risk-card">
      <div class="risk-card-label">Vault Exposure</div>
      <div class="risk-card-value" id="risk-vault-exposure">–</div>
      <div class="risk-card-meta" id="risk-vault-meta">Vault balances backing margin and flows.</div>
    </div>
    <div class="risk-card">
      <div class="risk-card-label">Insurance &amp; Stabilizer</div>
      <div class="risk-card-value" id="risk-insurance-level">–</div>
      <div class="risk-card-meta" id="risk-insurance-meta">Health of the insurance + stabilizer pools.</div>
    </div>
  </div>

  <!-- Middle charts row -->
  <div class="risk-grid risk-grid-middle">
    <div class="risk-panel">
      <div class="risk-panel-header-row">
        <h2>Issuance Corridor &amp; Price</h2>
        <span class="risk-badge" id="risk-corridor-badge">Corridor</span>
      </div>
      <div class="risk-chart-placeholder" id="risk-corridor-chart">
        <span>Corridor utilization &amp; price history will appear here.</span>
      </div>
    </div>

    <div class="risk-panel">
      <div class="risk-panel-header-row">
        <h2>Leverage &amp; Liquidations</h2>
        <span class="risk-badge" id="risk-leverage-badge">Leverage</span>
      </div>
      <div class="risk-chart-placeholder" id="risk-leverage-chart">
        <span>Average leverage and liquidation events will appear here.</span>
      </div>
    </div>
  </div>

  <!-- Bottom tables row -->
  <div class="risk-grid risk-grid-bottom">
    <div class="risk-table-panel">
      <div class="risk-panel-header-row">
        <h3>Per-Vault Risk (This Wallet)</h3>
      </div>
      <table class="risk-table">
        <thead>
          <tr>
            <th>Vault</th>
            <th>Tier</th>
            <th>Balance</th>
            <th>Linked Margin</th>
            <th>Status</th>
            <th>Risk</th>
          </tr>
        </thead>
        <tbody id="risk-vault-table-body">
          <tr>
            <td colspan="6" class="risk-table-empty">Connect a wallet to see vault risk.</td>
          </tr>
        </tbody>
      </table>
    </div>

    <div class="risk-table-panel">
      <div class="risk-panel-header-row">
        <h3>Per-Instrument Risk (Global)</h3>
      </div>
      <table class="risk-table">
        <thead>
          <tr>
            <th>Instrument</th>
            <th>Open Interest</th>
            <th>Avg Lev</th>
            <th>Vol Regime</th>
            <th>Corridor Util</th>
            <th>Risk</th>
          </tr>
        </thead>
        <tbody id="risk-instrument-table-body">
          <tr>
            <td colspan="6" class="risk-table-empty">Risk engine will populate instruments when DevNet is running.</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</section>

<section class="ws-panel" data-panel="network-benchmarks" style="display:none;">
  <header class="ws-panel-header">
    <h1>Benchmarks</h1>
  </header>
  <div class="vault-list-placeholder">
    Performance benchmarks vs corridor limits and targets.
  </div>
</section>

<section class="ws-panel" data-panel="network-health" style="display:none;">
  <header class="ws-panel-header">
    <h1>Network Health</h1>
  </header>
  <div class="vault-list-placeholder">
    Node status, latency, and partition risk visualizations.
  </div>
</section>

<section class="ws-panel" data-panel="node-governance" style="display:none;">
  <header class="ws-panel-header">
    <h1>Node Operator Governance</h1>
  </header>
  <div class="vault-list-placeholder">
    Operator-focused governance tools will be added here.
  </div>
</section>

<section class="ws-panel" data-panel="account-profile" style="display:none;">
  <header class="ws-panel-header">
    <h1>Profile &amp; Settings</h1>
  </header>
  <div class="vault-list-placeholder">
    Basic account profile and preferences will live here.
  </div>
</section>

<section class="ws-panel" data-panel="account-api-keys" style="display:none;">
  <header class="ws-panel-header">
    <h1>API Keys</h1>
  </header>
  <div class="vault-list-placeholder">
    API key management for programmatic access.
  </div>
</section>

<section class="ws-panel" data-panel="account-security" style="display:none;">
  <header class="ws-panel-header">
    <h1>Security</h1>
  </header>
  <div class="vault-list-placeholder">
    2FA, device fingerprints, and login history will surface here.
  </div>
</section>

<!-- Vault Transfer Modal -->
<div id="vault-transfer-modal" class="modal-backdrop" style="display:none;">
  <div class="modal">
    <div class="modal-header">
      <h2>Vault Transfer</h2>
      <button class="modal-close" data-close-modal>&times;</button>
    </div>
    <div class="modal-body">
      <div class="field">
        <span>From Vault</span>
        <select id="vault-transfer-from"></select>
      </div>
      <div class="field">
        <span>To Vault</span>
        <select id="vault-transfer-to"></select>
      </div>
      <div class="field">
        <span>Amount (GRC)</span>
        <input id="vault-transfer-amount" type="number" min="0" step="0.0001" />
      </div>
      <div class="modal-note">
        Internal transfers move balances ledger-side only (no on-chain transaction).
      </div>
    </div>
    <div class="modal-footer">
      <button class="btn btn-secondary" data-close-modal>Cancel</button>
      <button class="btn btn-primary" id="vault-transfer-submit">Transfer</button>
    </div>
  </div>
</div>


<div id="wallet-popup-modal" class="modal-backdrop" style="display:none;">
  <div class="modal wallet-modal">
    <div class="modal-header">
      <div class="wallet-modal-title-row">
        <h2>Wallet</h2>
        <span id="wallet-popup-address" class="wallet-address-chip">Not connected</span>
      </div>
      <button class="modal-close" data-wallet-close>&times;</button>
    </div>
    <div class="modal-body wallet-modal-body">
      <div class="wallet-modal-tabs">
        <button class="wallet-tab is-active" data-wallet-tab="overview">Overview</button>
        <button class="wallet-tab" data-wallet-tab="vaults">Vaults</button>
        <button class="wallet-tab" data-wallet-tab="stealth">Stealth</button>
        <button class="wallet-tab" data-wallet-tab="activity">Activity</button>
        <button class="wallet-tab" data-wallet-tab="settings">Settings</button>
      </div>

      <div class="wallet-tab-panel" data-wallet-tab-panel="overview">
        <div class="wallet-summary-row">
          <div>
            <div class="wallet-summary-label">Total Balance</div>
            <div class="wallet-summary-value" id="wallet-total-balance">–</div>
          </div>
          <div>
            <div class="wallet-summary-label">Available</div>
            <div class="wallet-summary-value" id="wallet-available-balance">–</div>
          </div>
        </div>
        <div class="wallet-assets-list" id="wallet-assets-list">
          <div class="wallet-assets-empty">Connect or create a wallet to view balances.</div>
        </div>
      </div>

      <div class="wallet-tab-panel" data-wallet-tab-panel="vaults" style="display:none;">
        <div class="wallet-tab-panel-header">
          <h3>Vaults</h3>
          <p>Vault balances and links to margin accounts.</p>
        </div>
        <div id="wallet-vaults-list" class="wallet-vaults-list">
          <div class="wallet-assets-empty">Vault data will appear here when available.</div>
        </div>
      </div>

      <div class="wallet-tab-panel" data-wallet-tab-panel="stealth" style="display:none;">
        <div class="wallet-tab-panel-header">
          <h3>Stealth Addresses</h3>
          <p>Generate and monitor stealth addresses tied to this wallet.</p>
        </div>
        <div id="wallet-stealth-list" class="wallet-stealth-list">
          <div class="wallet-assets-empty">No stealth addresses yet.</div>
        </div>
      </div>

      <div class="wallet-tab-panel" data-wallet-tab-panel="activity" style="display:none;">
        <div class="wallet-tab-panel-header">
          <h3>Recent Activity</h3>
          <p>Transfers, vault moves, tier renewals and trading-related actions.</p>
        </div>
        <div id="wallet-activity-list" class="wallet-activity-list">
          <div class="wallet-assets-empty">No recent activity recorded in this session.</div>
        </div>
      </div>

      <div class="wallet-tab-panel" data-wallet-tab-panel="settings" style="display:none;">
        <div class="wallet-tab-panel-header">
          <h3>Wallet Settings</h3>
          <p>Label, limits and security preferences for this browser-based dev wallet.</p>
        </div>
        <div class="wallet-settings-grid">
          <div class="field">
            <span>Label</span>
            <input type="text" id="wallet-settings-label" placeholder="My Workstation Wallet" />
          </div>
          <div class="field">
            <span>Max per-tx (dev only)</span>
            <input type="number" id="wallet-settings-max-tx" placeholder="2000" />
          </div>
          <div class="field">
            <span>Max per-day (dev only)</span>
            <input type="number" id="wallet-settings-max-day" placeholder="10000" />
          </div>
          <div class="field checkbox-field">
            <label>
              <input type="checkbox" id="wallet-settings-require-confirm" checked />
              Always show confirmation for trading and vault actions
            </label>
          </div>
          <div class="wallet-settings-actions">
            <button class="btn btn-secondary" id="wallet-settings-reset">Reset Dev Wallet</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</div>



<script src="/assets/js/ui_notifier.js"></script>
<script src="/assets/js/api_client.js"></script>
<script src="/assets/js/wallet_chain_client.js"></script>
<script src="/assets/js/wallet_keystore.js"></script>
<script src="/assets/js/session_manager.js"></script>

<script src="/assets/js/vault_client.js"></script>
<script src="/assets/js/workstation_nav_vaults.js"></script>
<script src="/assets/js/workstation_vault_dashboard.js"></script>
<script src="/assets/js/workstation_vault_stealth.js"></script>
<script src="/assets/js/workstation_vault_transfer_modal.js"></script>
<script src="/assets/js/workstation_vault_policies.js"></script>
<script src="/assets/js/workstation_vault_roles.js"></script>
<script src="/assets/js/workstation_vault_analytics.js"></script>
<script src="/assets/js/workstation_vault_audit.js"></script>
<script src="/assets/js/workstation_vault_settings.js"></script>
<script src="/assets/js/workstation_treasury_overview.js"></script>


<script src="/assets/js/workstation_trading_links.js"></script>
  <script src="/assets/js/workstation_portfolio.js"></script>

<script src="/assets/js/wallet_keystore.js"></script>
<script src="/assets/js/workstation_tiers.js"></script>
<script src="/assets/js/wallet_popup.js"></script>
<script src="/assets/js/workstation_risk_dashboard.js"></script>
</body>
</html>
