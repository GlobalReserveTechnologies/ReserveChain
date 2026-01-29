<?php ?>
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>ReserveChain — Network for Private Reserve Finance</title>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link rel="stylesheet" href="/assets/css/site.css">
</head>
<body>
<header class="top-nav">
    <div class="top-nav__left">
      <div class="logo-chip"><span>R</span></div>
      <div class="top-nav__brand">
        <div class="top-nav__brand-title">ReserveChain</div>
        <div class="top-nav__brand-sub">Private reserve routing network</div>
      </div>
    </div>
    <nav class="top-nav__center">
      <button class="nav-pill nav-pill--active" data-section="overview">Overview</button>
      <button class="nav-pill" data-section="architecture">Architecture</button>
      <button class="nav-pill" data-section="economics">Economics</button>
      <button class="nav-pill" data-section="privacy">Privacy &amp; Vaults</button>
      <button class="nav-pill" data-section="workstation">Workstation</button>
      <button class="nav-pill" data-section="governance">Governance</button>
      <button class="nav-pill" data-section="docs">Docs</button>
      <button class="nav-pill" data-section="faq">FAQ</button>
    </nav>
    <div class="top-nav__right">
      <button class="btn-ghost" id="node-status-pill">Status: <span class="status-text">Detecting node…</span></button>
      <button class="btn-ghost" id="btn-create-account">Create Account</button>
      <button class="btn-primary" id="btn-launch-workstation">Launch Workstation</button>
    </div>
  </header>

  <main>
    <section class="hero">
      <div class="hero-inner">
        <div>
          <div class="hero-badge-row">
            <div class="hero-badge">
              <span style="width:7px;height:7px;border-radius:999px;background:#4dffa3;box-shadow:0 0 10px rgba(77,255,163,0.8);"></span>
              Private reserve routing network
            </div>
            <div class="hero-badge">
              Deterministic issuance • Privacy-aligned
            </div>
          </div>
          <h1 class="hero-title">
            A network for <span>private wealth protection.</span>
          </h1>
          <p class="hero-subtitle">
            ReserveChain is a privacy-aligned reserve network with utilization-aware issuance, vault-backed reserves, and a hybrid DeFi/workstation stack. Wallet, workstation, explorer, and governance all share a single reserve engine.
          </p>
          <div class="hero-cta-row">
            <button class="btn-primary">Launch Workstation</button>
            <button class="btn-ghost">View Explorer</button>
            <button class="btn-ghost">Read Protocol Spec</button>
          </div>
          <div class="hero-tagline">
            Reserve routing · deterministic issuance
          </div>
          <div class="hero-metrics">
            <div class="metric-card">
              <div class="metric-label">Issuance Policy</div>
              <div class="metric-value">Adaptive · EMA-smoothed</div>
            </div>
            <div class="metric-card">
              <div class="metric-label">Trading Engine</div>
              <div class="metric-value">Hybrid CLOB</div>
            </div>
            <div class="metric-card">
              <div class="metric-label">Privacy</div>
              <div class="metric-value">Stealth + Pools</div>
            </div>
          </div>
        </div>
        <div class="hero-visual">
          <div class="hero-visual-orbit"></div>
          <div class="hero-node hero-node--1"></div>
          <div class="hero-node hero-node--2"></div>
          <div class="hero-node hero-node--3"></div>

          <div class="hero-console">
            <div class="hero-console__header">
              <div class="hero-console__title">rsc-node01 · reserve console</div>
              <div class="hero-chip-row">
                <span class="chip">PoAR + PoP</span>
                <span class="chip">Hybrid CLOB</span>
                <span class="chip">Stealth Wallets</span>
              </div>
            </div>
            <div class="console-line">
              <span class="console-prefix">[GUARD]</span>
              <span>Ledger integrity verified (18 classes)</span>
            </div>
            <div class="console-line">
              <span class="console-prefix">[ISSUANCE]</span>
              <span class="console-text-muted">Epoch 142 · multiplier 0.97 → 1.00 (Δ≤3%)</span>
            </div>
            <div class="console-line">
              <span class="console-prefix">[TRADECHAIN]</span>
              <span>orderbook synced · markets=3 · 142 open orders</span>
            </div>
            <div class="console-line">
              <span class="console-prefix">[DEFI]</span>
              <span>liquidity synced · pools=5 · TVL 1.23M WEB</span>
            </div>
            <div class="console-line">
              <span class="console-prefix">[WALLET]</span>
              <span class="console-text-muted">Stealth graph: 0 metadata leaks · 3 active sessions</span>
            </div>
          </div>
        </div>
      </div>
    </section>
    <?php include __DIR__ . '/sections/overview.php'; ?>
    <?php include __DIR__ . '/sections/architecture.php'; ?>
    <?php include __DIR__ . '/sections/network.php'; ?>
    <?php include __DIR__ . '/sections/economics.php'; ?>
    <?php include __DIR__ . '/sections/privacy.php'; ?>
    <?php include __DIR__ . '/sections/workstation.php'; ?>
    <?php include __DIR__ . '/sections/governance.php'; ?>
    <?php include __DIR__ . '/sections/security.php'; ?>
    <?php include __DIR__ . '/sections/docs.php'; ?>
    <?php include __DIR__ . '/sections/faq.php'; ?>
<div id="workstation-boot-overlay" class="boot-overlay boot-overlay--hidden">
    <div class="boot-overlay__backdrop"></div>
    <div class="boot-overlay__panel">
      <div class="boot-overlay__header">
        <div class="boot-overlay__title">Launching ReserveChain Workstation</div>
        <div class="boot-overlay__subtitle">Preparing a private finance session…</div>
      </div>
      <div class="boot-overlay__steps">
        <div class="boot-step" data-step="1">
          <span class="boot-step__label">Connecting to ReserveChain network</span>
          <span class="boot-step__status">Pending</span>
        </div>
        <div class="boot-step" data-step="2">
          <span class="boot-step__label">Discovering available nodes</span>
          <span class="boot-step__status">Pending</span>
        </div>
        <div class="boot-step" data-step="3">
          <span class="boot-step__label">Selecting optimal node from pool</span>
          <span class="boot-step__status">Pending</span>
        </div>
        <div class="boot-step" data-step="4">
          <span class="boot-step__label">Checking RPC &amp; ledger sync</span>
          <span class="boot-step__status">Pending</span>
        </div>
        <div class="boot-step" data-step="5">
          <span class="boot-step__label">Finalizing secure workstation session</span>
          <span class="boot-step__status">Pending</span>
        </div>
      </div>
      <div class="boot-overlay__footer">
        <div class="boot-overlay__hint">
          Nodes are auto-selected from your configured ReserveChain pool.
        </div>
        <button type="button" class="boot-overlay__cancel" id="boot-cancel-btn">Cancel</button>
      </div>
    </div>
  </div>

  <script src="/assets/js/site.js"></script>
  <script src="/assets/js/workstation_reserve_monitor.js"></script>
<script src="/assets/js/wallet_keystore.js"></script>
<script src="/assets/js/wallet_onboarding.js"></script>
</body>
</html>
