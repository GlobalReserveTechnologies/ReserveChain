<section id="network" class="section section--network">
  <div class="section-inner">
    <header class="section-header">
      <h2 class="section-kicker">Network Layout</h2>
      <h3 class="section-title">
        One main UI origin, <span>a pool of ReserveChain nodes behind it.</span>
      </h3>
      <p class="section-lede">
        ReserveChain separates the public-facing experience from the node cluster. A single
        canonical UI host (plus CDN edge cache) serves the website and workstation shell, while
        a pool of nodes handle consensus, vaults, and settlement behind load-balancing gateways.
      </p>
    </header>

    <div class="network-diagram">
      <div class="network-row">
        <div class="network-box network-box--user">
          <h4>User Devices</h4>
          <p>Browsers and workstations connect to a single trusted UI origin.</p>
          <ul>
            <li>Institutional desks</li>
            <li>Individual users</li>
            <li>Node operators</li>
          </ul>
        </div>
        <div class="network-arrow">UI + TLS</div>
        <div class="network-box network-box--ui">
          <h4>Main UI Server</h4>
          <p>
            Hosts the marketing site and workstation shell. Protected by TLS, WAF,
            and rate limiting. Static assets can be cached at the edge via CDN.
          </p>
          <ul>
            <li>Canonical HTTPS origin</li>
            <li>Web application firewall in front</li>
            <li>CDN cache for static assets</li>
          </ul>
        </div>
      </div>

      <div class="network-row network-row--gateways">
        <div class="network-arrow">RPC, WS, analytics</div>
        <div class="network-box network-box--gateway">
          <h4>Gateway &amp; Load Balancing Layer</h4>
          <p>
            Routes workstation and API traffic to the node pool. Can apply traffic shaping, DoS
            detection, and health-based node selection.
          </p>
          <ul>
            <li>L4/L7 load balancers</li>
            <li>Node health and latency probing</li>
            <li>Flood and DoS mitigation</li>
          </ul>
        </div>
      </div>

      <div class="network-row network-row--nodes">
        <div class="network-box network-box--nodes">
          <h4>ReserveChain Nodes</h4>
          <p>
            Nodes process blocks, track vaults, and serve RPC. They do not serve the public UI
            directly; instead they focus on security, performance, and correctness.
          </p>
          <ul>
            <li>Consensus and ledger replication</li>
            <li>Vault and coverage computation</li>
            <li>Workstation RPC and analytics feeds</li>
          </ul>
        </div>
      </div>
    </div>
  </div>
</section>
