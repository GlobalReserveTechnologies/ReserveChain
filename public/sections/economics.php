<section id="economics" class="section section--economics">
  <div class="section-inner">
    <header class="section-header">
      <h2 class="section-kicker">Economics</h2>
      <h3 class="section-title">
        Time Wonderland, <span>fixed and made auditable.</span>
      </h3>
      <p class="section-lede">
        ReserveChain keeps the idea of a reserve-backed, yield-aware system, but replaces reflexive
        “APY theater” with corridor bands, deterministic issuance, and vault-based coverage. The
        result: a reserve asset whose behavior can be explained, simulated, and audited.
      </p>
    </header>

    <div class="econ-grid">
      <article class="econ-card">
        <h4>Deterministic Issuance</h4>
        <p>
          New supply is minted on a fixed schedule with per-epoch rules. Utilization and coverage
          can modulate issuance inside tight bounds, but never explode it into runaway inflation.
        </p>
        <ul>
          <li>Epoch-based issuance schedule (no “random” APY spikes)</li>
          <li>Utilization bands adjust within a capped envelope</li>
          <li>Mint &amp; burn events tied to real reserve changes</li>
        </ul>
      </article>

      <article class="econ-card">
        <h4>Corridor Bands</h4>
        <p>
          Instead of targeting a single “peg” with infinite ammo, ReserveChain defines a corridor:
          a band within which the asset is allowed to float while reserves and policy respond.
        </p>
        <ul>
          <li>Upper band: dampen speculation and slow issuance</li>
          <li>Lower band: enable redemptions and incentivize coverage</li>
          <li>Transparent policy curves instead of mysterious “interventions”</li>
        </ul>
      </article>

      <article class="econ-card">
        <h4>Vault-Backed Coverage</h4>
        <p>
          The reserve engine tracks NAV, coverage, and stress. The workstation exposes this as a
          first-class metric, not a hidden spreadsheet.
        </p>
        <ul>
          <li>Multi-asset vaults with per-asset risk weights</li>
          <li>Coverage ratio directly visible in the workstation</li>
          <li>Stress testing across corridors and epochs</li>
        </ul>
      </article>
    </div>

    <div class="econ-charts">
      <div class="econ-chart econ-chart--corridor">
        <h4>Corridor Band Visualization</h4>
        <p>
          A stylized view of the target corridor: price can move, but policy only reacts when it
          presses against the bands. This keeps ReserveChain responsive without being manic.
        </p>
        <div class="econ-chart__graphic econ-chart__graphic--corridor">
          <div class="corridor-band corridor-band--lower">Lower band</div>
          <div class="corridor-band corridor-band--target">Target region</div>
          <div class="corridor-band corridor-band--upper">Upper band</div>
          <div class="corridor-marker corridor-marker--price">Current indicative level</div>
        </div>
      </div>

      <div class="econ-chart econ-chart--issuance">
        <h4>Epoch Issuance Profile</h4>
        <p>
          Issuance is front-loaded when the system is under-utilized and grows carefully with the
          network. Utilization and coverage can tighten the curve, but never blow it up.
        </p>
        <div class="econ-chart__graphic econ-chart__graphic--issuance">
          <div class="issuance-bar issuance-bar--early"></div>
          <div class="issuance-bar issuance-bar--mid"></div>
          <div class="issuance-bar issuance-bar--late"></div>
        </div>
        <small class="econ-chart__note">
          Visualization is conceptual. The workstation shows live numbers from the devnet
          economics engine.
        </small>
      </div>
    </div>
  </div>
</section>
