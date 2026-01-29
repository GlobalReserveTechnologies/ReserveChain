<section id="workstation" class="section section--workstation">
  <div class="section-inner">
    <header class="section-header">
      <h2 class="section-kicker">Workstation</h2>
      <h3 class="section-title">
        A private finance workstation built directly on the reserve engine.
      </h3>
      <p class="section-lede">
        The ReserveChain Workstation is where reserve economics, trading, lending, and governance
        come together. It behaves like a professional trading desk, but every position, ticket, and
        payout is aware of the underlying reserve model.
      </p>
    </header>

    <div class="workstation-grid">
      <article class="workstation-card">
        <h4>Trading Terminal</h4>
        <p>
          A full trading terminal with custom charts, on-chart order lines, liquidation and
          breakeven overlays, and position panels designed for intraday use.
        </p>
        <ul>
          <li>Place and adjust orders directly from the chart</li>
          <li>See liquidation and breakeven lines for each position</li>
          <li>Reserve-aware margin and collateral views</li>
        </ul>
      </article>

      <article class="workstation-card">
        <h4>Lending &amp; Credit</h4>
        <p>
          Issue and manage credit against reserve-aware collateral. Lending policy is bound to the
          same corridor and coverage rules that govern issuance.
        </p>
        <ul>
          <li>Collateral ratios tied to corridor bands</li>
          <li>Transparent liquidation and recovery flows</li>
          <li>Vault-aware risk metrics per borrower</li>
        </ul>
      </article>

      <article class="workstation-card">
        <h4>Analytics &amp; Governance</h4>
        <p>
          Node operators and power users can see corridor windows, coverage, issuance, and PoP
          rewards in the same place they vote on changes.
        </p>
        <ul>
          <li>Live coverage and NAV views</li>
          <li>Corridor window and settlement ticket dashboards</li>
          <li>Node-local governance controls (no public signing UI)</li>
        </ul>
      </article>
    </div>

    <div class="workstation-reserve-monitor">
      <div class="info-card info-card--wide">
        <div class="info-card__tag">Reserve Monitor</div>
        <div class="info-card__title">
          Live NAV and corridor status from your connected node.
        </div>
        <div class="info-card__body">
          <div class="reserve-metrics-row">
            <div class="reserve-metric">
              <div class="reserve-metric__label">Total Reserves</div>
              <div class="reserve-metric__value" id="rm-total-reserve">–</div>
              <div class="reserve-metric__hint" id="rm-reserve-detail">Waiting for node…</div>
            </div>
            <div class="reserve-metric">
              <div class="reserve-metric__label">GRC Supply</div>
              <div class="reserve-metric__value" id="rm-supply-grc">–</div>
              <div class="reserve-metric__hint">Outstanding DetNet-issued GRC.</div>
            </div>
            <div class="reserve-metric">
              <div class="reserve-metric__label">NAV</div>
              <div class="reserve-metric__value" id="rm-nav">–</div>
              <div class="reserve-metric__hint" id="rm-corridor">–</div>
            </div>
            <div class="reserve-metric reserve-metric--status">
              <div class="reserve-metric__label">Peg Status</div>
              <div class="reserve-metric__value" id="rm-peg-status">–</div>
              <div class="reserve-metric__hint">
                Computed around 1.0000 with a ±10bps corridor.
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>


    <div class="workstation-cta">
      <p>Run the full workstation against your own node or a trusted gateway.</p>
      <button class="btn-primary" id="btn-launch-workstation-secondary">
        Launch Workstation
      </button>
    </div>
  </div>
</section>
