<section id="governance" class="section section--governance">
  <div class="section-inner">
    <header class="section-header">
      <h2 class="section-kicker">Governance</h2>
      <h3 class="section-title">
        Policy that lives at the protocol layer, not in a Discord poll.
      </h3>
      <p class="section-lede">
        Governance in ReserveChain is designed for serious capital: reserves, corridors, and risk
        parameters are all governed by on-chain proposals and votes. The public site explains how
        it works; the workstation is where operators actually participate.
      </p>
    </header>

    <div class="gov-grid">
      <article class="gov-card">
        <h4>On-Chain Parameters</h4>
        <p>
          Issuance curves, corridor bounds, and vault eligibility rules can all be tuned via
          proposals. Changes are applied at epoch boundaries and are fully auditable.
        </p>
        <ul>
          <li>Corridor width and shape</li>
          <li>Reserve eligibility and weights</li>
          <li>PoP reward shares and criteria</li>
        </ul>
      </article>

      <article class="gov-card">
        <h4>Node-Local Voting</h4>
        <p>
          Governance signing never happens on the public website. Operators vote from their node’s
          local workstation, against their staked and delegated power.
        </p>
        <ul>
          <li>Local-only governance signing UI</li>
          <li>Delegation and quorum rules at the protocol layer</li>
          <li>Audit trail for each parameter change</li>
        </ul>
      </article>

      <article class="gov-card">
        <h4>Transparency Without Chaos</h4>
        <p>
          Proposals, vote tallies, and outcomes are public. The decision process is explicit,
          documented, and replayable from the chain history.
        </p>
        <ul>
          <li>Public proposal and outcome history</li>
          <li>Deterministic execution of accepted changes</li>
          <li>Clear separation of discussion vs. binding votes</li>
        </ul>
      </article>
    </div>
  </div>
</section>
