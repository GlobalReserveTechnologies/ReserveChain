<section id="architecture" class="section section--architecture">
  <div class="section-inner">
    <header class="section-header">
      <h2 class="section-kicker">Architecture</h2>
      <h3 class="section-title">
        A layered reserve network: <span>engine, vault, privacy, and workstation.</span>
      </h3>
      <p class="section-lede">
        ReserveChain is not just a chain with a UI. It is a layered system: the consensus engine
        enforces deterministic issuance, the vault layer tracks real coverage, the privacy layer
        protects flows, and the workstation exposes everything as a professional finance platform.
      </p>
    </header>

    <div class="arch-diagram">
      <div class="arch-layer arch-layer--top">
        <h4>Workstation &amp; Applications</h4>
        <p>
          Trading terminal, lending, analytics, and governance dashboards. Runs as a dedicated
          workstation on top of the node stack, speaking to the engine via REST and WebSocket RPC.
        </p>
        <ul>
          <li>Trading terminal with custom chart &amp; order overlays</li>
          <li>Positions, PnL, liquidation &amp; collateral views</li>
          <li>Node operator console &amp; governance controls (local only)</li>
        </ul>
      </div>

      <div class="arch-layer arch-layer--middle">
        <div class="arch-column">
          <h4>Vault &amp; Reserve Engine</h4>
          <p>
            Tracks basket-style reserves, coverage ratios, and corridor bands. Converts utilization
            signals into issuance and redemption decisions instead of chasing reflexive APY.
          </p>
          <ul>
            <li>Multi-asset reserve vaults (stablecoins, majors, yield assets)</li>
            <li>Coverage ratio &amp; NAV computation per epoch</li>
            <li>Settlement tickets and redemption queues</li>
          </ul>
        </div>
        <div class="arch-column">
          <h4>Privacy &amp; Stealth Routing</h4>
          <p>
            A dedicated privacy layer for capital flows: stealth addresses, vault-shielded balances,
            and routing that preserves privacy while keeping the reserve engine auditable.
          </p>
          <ul>
            <li>Per-user stealth address trees for private deposits</li>
            <li>Vault-level aggregation: public coverage, private balances</li>
            <li>Optional compliance views with controlled revelation</li>
          </ul>
        </div>
      </div>

      <div class="arch-layer arch-layer--bottom">
        <div class="arch-column">
          <h4>Consensus &amp; Ledger</h4>
          <p>
            A hybrid consensus flow: Proof-of-Work for bootstrapping, Proof-of-Stake for steady-state
            security, and Proof-of-Participation to reward real usage and routing.
          </p>
          <ul>
            <li>PoW for initial distribution and censorship resistance</li>
            <li>PoS for energy-efficient finality and stake-based security</li>
            <li>PoP for rewarding nodes that actually route and settle value</li>
          </ul>
        </div>
        <div class="arch-column">
          <h4>RPC, Indexing &amp; Node Roles</h4>
          <p>
            Nodes focus on consensus, vault accounting, and data. A dedicated main UI server hosts
            the public website and workstation entry point, while nodes sit behind load-balancing
            gateways.
          </p>
          <ul>
            <li>Standardized REST + WebSocket RPC for wallets &amp; apps</li>
            <li>Gateway pool in front of the node cluster</li>
            <li>Indexer views for explorers and analytics</li>
          </ul>
        </div>
      </div>
    </div>
  </div>
</section>
