// workstation_tiers.js
// DevNet helper to query tier status and perform TX_TIER_RENEW via popup flows.

document.addEventListener('DOMContentLoaded', function () {
  // Optionally auto-load tier status when a specific panel is opened later.
});

async function rcFetchJSON(url, opts) {
  const res = await fetch(url, opts || {});
  const data = await res.json();
  if (!data.success) {
    throw new Error(data.error || 'Request failed');
  }
  return data;
}

// Load current tier + pricing
async function rcLoadTierStatus() {
  return rcFetchJSON('/api/tier_status.php');
}

// Load current Earn summary
async function rcLoadEarnSummary() {
  return rcFetchJSON('/api/earn_summary.php');
}

// Request a checkout quote from PHP
async function rcTierCheckoutQuote(targetTier, billingCycle, paymentSource) {
  return rcFetchJSON('/api/tier_checkout_quote.php', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      target_tier: targetTier,
      billing_cycle: billingCycle || 'yearly',
      payment_source: paymentSource || 'vault'
    })
  });
}

// Commit a signed TX_TIER_RENEW via PHP -> Go -> chain.
async function rcTierCheckoutCommit(quote, accountId) {
  const wallet = window.ReserveWallet;
  if (!wallet) throw new Error('ReserveWallet not available');

  const accounts = wallet.listAccounts();
  const acct = accounts.find(a => a.id === accountId) || accounts[0];
  if (!acct) throw new Error('No wallet accounts configured');

  const txBody = {
    sender: acct.address,
    nonce: Date.now(), // simple monotonic nonce per browser; replace later
    tier: quote.target_tier,
    billing_cycle: quote.billing_cycle,
    payment: {
      source: quote.payment_source,
      amount_grc: quote.total_payable_grc
    },
    earn_applied_grc: quote.earn_applied_grc,
    stake_discount_grc: quote.stake_discount_grc,
    surplus_to_time_grc: quote.surplus_earn_grc,
    timestamp: Math.floor(Date.now() / 1000)
  };

  const sig = await wallet.signMessage(acct.id, txBody);

  const chainTx = {
    type: 'TX_TIER_RENEW',
    tx: Object.assign({}, txBody, {
      signature: sig.signature
    })
  };

  const commitRes = await fetch('/api/tier_checkout_commit.php', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      target_tier: quote.target_tier,
      billing_cycle: quote.billing_cycle,
      payment_source: quote.payment_source,
      chain_tx: chainTx
    })
  });
  const data = await commitRes.json();
  if (!data.success) {
    throw new Error(data.error || 'Tier commit failed');
  }
  return data;
}

// --- Tier & Billing panel wiring ---

(function () {
  function selectChip(groupEl, valueAttr, value) {
    if (!groupEl) return;
    const chips = Array.from(groupEl.querySelectorAll('.chip'));
    chips.forEach(ch => {
      if (ch.dataset[valueAttr] === value) ch.classList.add('is-active');
      else ch.classList.remove('is-active');
    });
  }

  async function refreshTierPanel() {
    try {
      const [tierRes, earnRes] = await Promise.all([
        rcLoadTierStatus(),
        rcLoadEarnSummary()
      ]);

      const tier = tierRes.tier;
      const pricing = tierRes.pricing;
      const runtime = tierRes.runtime;
      const earn = earnRes;

      const titleEl = document.getElementById('tier-card-title');
      const subEl = document.getElementById('tier-card-subtitle');
      const statusEl = document.getElementById('tier-card-status');
      const billEl = document.getElementById('tier-pill-billing');
      const renewEl = document.getElementById('tier-pill-renewal');
      const marginEl = document.getElementById('tier-stat-margin');
      const popEl = document.getElementById('tier-stat-pop');
      const stakeEl = document.getElementById('tier-stat-stake');
      const hairEl = document.getElementById('tier-stat-haircut');
      const grcEl = document.getElementById('tier-pricing-grc');
      const usdEl = document.getElementById('tier-pricing-usd');

      if (titleEl) {
        const labelMap = {
          core: 'Core Reserve',
          elite: 'Elite Reserve',
          executive: 'Executive Reserve',
          express: 'Express Reserve'
        };
        titleEl.textContent = labelMap[tier.tier] || (tier.tier || '').toUpperCase();
      }
      if (subEl) {
        if (tier.tier === 'express') {
          subEl.textContent = 'Maximum coverage, speed, and benefits.';
        } else if (tier.tier === 'executive') {
          subEl.textContent = 'High throughput and enhanced benefits.';
        } else if (tier.tier === 'elite') {
          subEl.textContent = 'More coverage and higher limits.';
        } else {
          subEl.textContent = 'Baseline coverage and access.';
        }
      }
      if (statusEl) {
        statusEl.textContent = (tier.status || 'active').charAt(0).toUpperCase() + (tier.status || 'active').slice(1);
      }
      if (billEl) {
        billEl.textContent = (tier.billing_cycle === 'yearly' ? 'Yearly billing' : 'Monthly billing');
      }
      if (renewEl && tier.renew_expires_at) {
        try {
          const renewDate = new Date(tier.renew_expires_at);
          const now = new Date();
          const diffMs = renewDate.getTime() - now.getTime();
          const diffDays = Math.max(0, Math.round(diffMs / 86400000));
          renewEl.textContent = 'Renews in ' + diffDays + ' days';
        } catch (e) {
          renewEl.textContent = 'Renews at ' + tier.renew_expires_at;
        }
      }

      if (marginEl) marginEl.textContent = '×' + (runtime.margin_limit || 0);
      if (popEl) popEl.textContent = (runtime.pop_multiplier || 1).toFixed(1) + '×';
      if (stakeEl) stakeEl.textContent = (runtime.staking_multiplier || 1).toFixed(1) + '×';
      if (hairEl) hairEl.textContent = Math.round((runtime.haircut_applied || 0) * 100) + '%';

      if (grcEl) grcEl.textContent = (pricing.base_grc || 0) + ' GRC';
      if (usdEl) usdEl.textContent = '≈ $' + (pricing.base_usd || 0).toFixed(2);

      // Earn card
      const earnGrcEl = document.getElementById('earn-balance-grc');
      const earnUsdEl = document.getElementById('earn-balance-usd');
      const capEl = document.getElementById('earn-capital');
      const flowEl = document.getElementById('earn-flow');
      const riskEl = document.getElementById('earn-risk');
      const bonusEl = document.getElementById('earn-bonus');

      if (earnGrcEl) earnGrcEl.textContent = (earn.balance_grc || 0) + ' GRC';
      if (earnUsdEl) earnUsdEl.textContent = '≈ $' + (earn.balance_usd || 0).toFixed(2);

      if (earn.breakdown) {
        if (capEl) capEl.textContent = earn.breakdown.capital_grc || 0;
        if (flowEl) flowEl.textContent = earn.breakdown.flow_grc || 0;
        if (riskEl) riskEl.textContent = earn.breakdown.risk_grc || 0;
        if (bonusEl) bonusEl.textContent = earn.breakdown.bonus_grc || 0;
      }
    } catch (e) {
      console.error('Failed to refresh Tier & Billing panel', e);
    }
  }

  document.addEventListener('DOMContentLoaded', function () {
    const root = document.querySelector('[data-panel="account-tier-billing"]');
    if (!root) return;

    // Hook panel switching: when this panel is shown, refresh data.
    const observer = new MutationObserver(() => {
      const display = root.style.display || window.getComputedStyle(root).display;
      if (display !== 'none') {
        refreshTierPanel();
      }
    });
    observer.observe(root, { attributes: true, attributeFilter: ['style'] });

    // Buttons
    const btnUpgrade = document.getElementById('btn-tier-upgrade');
    const btnRefresh = document.getElementById('btn-tier-refresh');
    const modal = document.getElementById('tier-upgrade-modal');
    const btnClose = document.getElementById('tier-upgrade-close');
    const btnCancel = document.getElementById('tier-upgrade-cancel');
    const btnConfirm = document.getElementById('tier-upgrade-confirm');
    const quoteBody = document.getElementById('tier-quote-body');

    const planButtons = Array.from(document.querySelectorAll('.tier-plan-option'));
    const billingGroup = document.getElementById('tier-billing-group');
    const sourceGroup = document.getElementById('tier-source-group');

    let selectedTier = 'express';
    let selectedBilling = 'monthly';
    let selectedSource = 'vault';
    let lastQuote = null;

    function openModal() {
      if (modal) modal.style.display = 'flex';
      if (btnConfirm) {
        btnConfirm.disabled = true;
        btnConfirm.textContent = 'Sign & Upgrade';
      }
      if (quoteBody) {
        quoteBody.innerHTML = '<p class="tier-quote-placeholder">Select a plan and billing cycle to see a live quote.</p>';
      }
      lastQuote = null;
    }
    function closeModal() {
      if (modal) modal.style.display = 'none';
    }

    if (btnUpgrade && modal) {
      btnUpgrade.addEventListener('click', openModal);
    }
    if (btnRefresh) {
      btnRefresh.addEventListener('click', refreshTierPanel);
    }
    if (btnClose) btnClose.addEventListener('click', closeModal);
    if (btnCancel) btnCancel.addEventListener('click', closeModal);

    planButtons.forEach(btn => {
      btn.addEventListener('click', async function () {
        planButtons.forEach(b => b.classList.remove('is-active'));
        this.classList.add('is-active');
        selectedTier = this.dataset.tier;
        await updateQuote();
      });
    });

    if (billingGroup) {
      billingGroup.addEventListener('click', async function (e) {
        const target = e.target.closest('.chip');
        if (!target) return;
        selectedBilling = target.dataset.billing;
        selectChip(billingGroup, 'billing', selectedBilling);
        await updateQuote();
      });
    }
    if (sourceGroup) {
      sourceGroup.addEventListener('click', async function (e) {
        const target = e.target.closest('.chip');
        if (!target) return;
        selectedSource = target.dataset.source;
        selectChip(sourceGroup, 'source', selectedSource);
        await updateQuote();
      });
    }

    async function updateQuote() {
      try {
        if (!quoteBody) return;
        quoteBody.innerHTML = '<p class="tier-quote-placeholder">Fetching quote...</p>';
        const data = await rcTierCheckoutQuote(selectedTier, selectedBilling, selectedSource);
        const q = data.quote;
        lastQuote = q;
        const html = [
          '<div class="tier-quote-line"><span>Plan</span><span>' + q.target_tier.toUpperCase() + '</span></div>',
          '<div class="tier-quote-line"><span>Billing</span><span>' + q.billing_cycle + '</span></div>',
          '<div class="tier-quote-line"><span>Base cost</span><span>' + q.base_cost_grc + ' GRC</span></div>',
          '<div class="tier-quote-line"><span>Stake discount</span><span>- ' + q.stake_discount_grc + ' GRC</span></div>',
          '<div class="tier-quote-line"><span>Earn applied</span><span>- ' + q.earn_applied_grc + ' GRC</span></div>',
          '<div class="tier-quote-line"><span>Surplus Earn</span><span>' + q.surplus_earn_grc + ' GRC → time</span></div>',
          '<div class="tier-quote-line tier-quote-total"><span>Total today</span><span>' + q.total_payable_grc + ' GRC</span></div>'
        ].join('');
        quoteBody.innerHTML = '<div class="tier-quote-summary">' + html + '</div>';
        if (btnConfirm) btnConfirm.disabled = false;
      } catch (e) {
        console.error('Failed to update quote', e);
        if (quoteBody) {
          quoteBody.innerHTML = '<p class="tier-quote-error">Failed to fetch quote. Please try again.</p>';
        }
      }
    }

    if (btnConfirm) {
      btnConfirm.addEventListener('click', async function () {
        if (!lastQuote) return;
        try {
          btnConfirm.disabled = true;
          btnConfirm.textContent = 'Signing...';

          const accounts = window.ReserveWallet ? window.ReserveWallet.listAccounts() : [];
          const defaultAccountId = accounts[0] ? accounts[0].id : null;
          if (!defaultAccountId) throw new Error('No wallet accounts configured');

          const commit = await rcTierCheckoutCommit(lastQuote, defaultAccountId);
          console.log('Tier upgrade committed', commit);
          btnConfirm.textContent = 'Upgraded';
          await refreshTierPanel();
          setTimeout(closeModal, 600);
        } catch (e) {
          console.error('Tier upgrade failed', e);
          btnConfirm.disabled = false;
          btnConfirm.textContent = 'Sign & Upgrade';
          if (quoteBody) {
            quoteBody.innerHTML += '<p class="tier-quote-error">Upgrade failed: ' + (e.message || 'Unknown error') + '</p>';
          }
        }
      });
    }
  });
})();
