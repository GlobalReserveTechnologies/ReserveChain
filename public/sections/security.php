<section id="security" class="section section--security">
  <div class="section-inner">
    <header class="section-header">
      <h2 class="section-kicker">Security</h2>
      <h3 class="section-title">
        Protocol, vault, and workstation security designed like a private bank.
      </h3>
      <p class="section-lede">
        ReserveChain assumes that serious capital requires serious hardening. The protocol, node
        cluster, and workstation are all designed to survive abuse: web attacks, traffic floods,
        and economic stress.
      </p>
    </header>

    <div class="security-grid">
      <article class="security-card">
        <h4>Protocol &amp; Economic Security</h4>
        <p>
          Hybrid PoW/PoS/PoP reduces single-mode failure. Deterministic issuance and corridor bands
          avoid the runaway feedback loops that broke earlier “reserve” experiments.
        </p>
        <ul>
          <li>Hybrid consensus (PoW boot, PoS steady-state, PoP rewards)</li>
          <li>Deterministic issuance tied to real reserves and corridors</li>
          <li>No reflexive “infinite APY” games</li>
        </ul>
      </article>

      <article class="security-card">
        <h4>Gateway, WAF &amp; Traffic Defense</h4>
        <p>
          The main UI server sits behind a web application firewall and load balancers. Gateways
          can throttle, filter, and route traffic before it ever touches the node cluster.
        </p>
        <ul>
          <li>Web application firewall in front of the UI origin</li>
          <li>L4/L7 load balancing across node gateways</li>
          <li>DoS and flood detection with rate limiting</li>
        </ul>
      </article>

      <article class="security-card">
        <h4>Workstation &amp; Node Security</h4>
        <p>
          The workstation talks to nodes via defined RPC, not raw database access. Governance
          controls are node-local, and trading sessions can be hardened with additional auth in
          production deployments.
        </p>
        <ul>
          <li>Strict RPC schema between workstation and nodes</li>
          <li>Node-local governance only (no public signing UI)</li>
          <li>Optional authentication and mTLS for production setups</li>
        </ul>
      </article>
    </div>
  </div>
</section>
