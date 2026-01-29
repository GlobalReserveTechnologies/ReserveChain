(function () {
    const mainEl = document.getElementById('app-main');
    const navButtons = document.querySelectorAll('.app-nav button');
    const profileSpans = document.querySelectorAll('.profile-toggle span');
    const statusStrip = document.getElementById('status-strip');

    let currentView = 'dashboard';

    function fmt(n, decimals = 4) {
        if (n === null || n === undefined || isNaN(n)) return '--';
        return Number(n).toFixed(decimals);
    }

    function renderStatusStrip(state) {
        const latency = state.tickLatencyMs ? `${state.tickLatencyMs} ms` : '--';
        const connClass = `conn-${state.connection}`;
        const navVal = state.nav ? fmt(state.nav.grc, 6) : '--';
        const eurusd = state.fx ? fmt(state.fx.eur_usd, 5) : '--';

        statusStrip.className = `app-status-strip ${connClass} flash-update`;
        statusStrip.innerHTML = `
            <div class="status-item">
                <span class="label">Connection</span>
                <span class="value">${state.connection}</span>
            </div>
            <div class="status-item">
                <span class="label">Leader</span>
                <span class="value">${state.leaderId || '--'}</span>
            </div>
            <div class="status-item">
                <span class="label">Tick</span>
                <span class="value">#${state.lastTickId || '--'} (${latency})</span>
            </div>
            <div class="status-item">
                <span class="label">GRC NAV</span>
                <span class="value">${navVal}</span>
            </div>
            <div class="status-item">
                <span class="label">EUR/USD</span>
                <span class="value">${eurusd}</span>
            </div>
            <div class="status-item">
                <span class="label">Window</span>
                <span class="value">${state.windows ? state.windows.status : '--'} (#${state.windows ? state.windows.window_id : '--'})</span>
            </div>
            <div class="status-item">
                <span class="label">Profile</span>
                <span class="value">${state.profile}</span>
            </div>
        `;
        setTimeout(() => {
            statusStrip.classList.remove('flash-update');
        }, 200);
    }

    function sparkline(points, width = 120, height = 32, key = 'y') {
        if (!points || points.length === 0) return '';
        let xs = points.map((_, i) => i);
        let ys = points.map(p => p[key]);

        let minY = Math.min(...ys);
        let maxY = Math.max(...ys);
        if (maxY === minY) maxY = minY + 0.0001;

        const scaleX = (width - 4) / (xs.length - 1 || 1);
        const scaleY = (height - 4) / (maxY - minY);

        const path = xs.map((x, i) => {
            const px = 2 + x * scaleX;
            const py = height - 2 - (ys[i] - minY) * scaleY;
            return (i === 0 ? 'M' : 'L') + px + ' ' + py;
        }).join(' ');

        return `
            <svg class="sparkline" width="${width}" height="${height}">
                <path d="${path}" fill="none" stroke="currentColor" stroke-width="1.2" />
            </svg>
        `;
    }

    function windowProgressPercent(state) {
        const w = state.windows;
        if (!w) return 0;
        const now = Date.now() / 1000;
        const opened = w.opened_at_unix || w.OpenedAtUnix;
        const nextClose = w.next_close_unix || w.NextCloseUnix;
        if (!opened || !nextClose || nextClose <= opened) return 0;
        const pct = ((now - opened) / (nextClose - opened)) * 100;
        return Math.max(0, Math.min(100, pct));
    }

    function renderWindowStrip(state) {
        const history = state.windowHistory || [];
        if (!history.length) return '<div class="window-strip empty">No windows yet</div>';

        const tiles = history.map(w => {
            let cls = 'window-tile ';
            if (w.status === 'open') cls += 'open';
            else if (w.status === 'closing') cls += 'closing';
            else if (w.status === 'settled') cls += 'settled';

            const nav = w.nav ? Number(w.nav).toFixed(5) : '';
            return `
                <div class="${cls}">
                    <span class="id">#${w.window_id}</span>
                    <span class="status">${w.status}</span>
                    ${nav ? `<span class="nav">${nav}</span>` : ''}
                </div>
            `;
        }).join('');

        return `<div class="window-strip">${tiles}</div>`;
    }

    async function mintDemo(amount) {
        await fetch('/api/mint', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                address: window.ReserveStore.state.address,
                asset: 'USD',
                amount: parseFloat(amount)
            })
        });
        window.ReserveFetchBalances();
    }

    async function redeemDemo(amount) {
        await fetch('/api/redeem', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                address: window.ReserveStore.state.address,
                asset: 'USD',
                amount: parseFloat(amount)
            })
        });
        window.ReserveFetchBalances();
    }

    

    function renderPositions(state) {
        const b = (state.balances && state.balances.balances) || {};
        const totalGrc = b.GRC || 0;
        const totalUsd = b.USD || 0;
        const pending = state.pendingSettlement || 0;
        const redeemable = state.redeemableGrc || totalGrc;

        const navHist = state.navHistory || [];
        const supplyHist = state.supplyHistory || [];
        const treasHist = state.treasuryHistory || [];

        function miniSpark(id, data, label) {
            if (!data || !data.length) return `<div class="mini-sparkline empty">${label}: --</div>`;
            const last = data[data.length - 1];
            return `<div class="mini-sparkline">${label}: ${fmt(last.y || last.value || last, 4)}</div>`;
        }

        mainEl.innerHTML = `
            <section class="positions-top">
              <div class="metric-card">
                <div class="metric-label">Total Exposure (GRC)</div>
                <div class="metric-value">${fmt(totalGrc, 4)}</div>
                ${miniSpark('nav', navHist.map(p => ({ y: p.grc })), 'NAV (Net Asset Value)')}
              </div>
              <div class="metric-card">
                <div class="metric-label">Total Exposure (USD)</div>
                <div class="metric-value">${fmt(totalUsd, 2)}</div>
                ${miniSpark('supply', supplyHist, 'Supply')}
              </div>
              <div class="metric-card">
                <div class="metric-label">Pending Settlement</div>
                <div class="metric-value">${fmt(pending, 4)} GRC</div>
                ${miniSpark('treasury', treasHist, 'Treasury')}
              </div>
              <div class="metric-card">
                <div class="metric-label">Redeemable Now</div>
                <div class="metric-value">${fmt(redeemable, 4)} GRC</div>
              </div>
            </section>

            <section class="positions-section">
              <h2>Cash &amp; Balances</h2>
              <table class="rc-table">
                <thead>
                  <tr>
                    <th>Type</th>
                    <th>Amount</th>
                    <th>Status</th>
                  </tr>
                </thead>
                <tbody>
                  <tr>
                    <td>Wallet GRC Balance</td>
                    <td>${fmt(totalGrc, 4)} GRC</td>
                    <td>Liquid</td>
                  </tr>
                  <tr>
                    <td>Wallet USD Balance</td>
                    <td>${fmt(totalUsd, 2)} USD</td>
                    <td>Reference</td>
                  </tr>
                </tbody>
              </table>
            </section>

            <section class="positions-section">
              <h2>Issuance &amp; Redemption</h2>
              <div class="positions-subtext">
                View your mint / redeem activity and corridor participation.
              </div>
              <div id="positions-activity-empty">Activity history will appear here as DevNet events accumulate.</div>
            </section>

            <section class="positions-section">
              <h2>Settlement Exposure</h2>
              <div class="positions-subtext">
                Corridor windows group deposits and redemptions before settlement at NAV (Net Asset Value).
              </div>
              <div id="positions-corridor-empty">Per-window exposure will appear here once history is available.</div>
            </section>
        `;
    }

    function renderAnalytics(state) {
        const navHist = state.navHistory || [];
        const fxHist = state.fxHistory || [];
        const winHist = state.windowHistory || [];
        const supplyHist = state.supplyHistory || [];
        const treasHist = state.treasuryHistory || [];

        function buildSeries(data, key) {
            return (data || []).map((p, idx) => ({
                x: idx,
                y: key ? p[key] : (p.y || p.value || p)
            }));
        }

        const navSeries = buildSeries(navHist, 'grc');
        const fxSeries = buildSeries(fxHist, 'eur_usd');
        const winSeries = buildSeries(winHist, 'net_flow');
        const supplySeries = buildSeries(supplyHist);
        const treasSeries = buildSeries(treasHist);

        function chartBox(title, series) {
            const latest = series && series.length ? series[series.length - 1].y : null;
            return `
              <div class="analytics-chart">
                <div class="analytics-title">${title}</div>
                <div class="analytics-latest">${latest !== null ? fmt(latest, 6) : '--'}</div>
                <div class="analytics-sparkline-placeholder">
                  Sparkline (preview) — detailed charts will live in the Analytics Terminal.
                </div>
              </div>
            `;
        }

        mainEl.innerHTML = `
          <section class="analytics-grid">
            ${chartBox('NAV (Net Asset Value)', navSeries)}
            ${chartBox('FX (Foreign Exchange) EUR/USD', fxSeries)}
            ${chartBox('Corridor Net Flow per Window', winSeries)}
            ${chartBox('GRC Supply Over Time', supplySeries)}
            ${chartBox('Treasury Balance Over Time', treasSeries)}
          </section>

          <section class="positions-section">
            <h2>Corridor History (Preview)</h2>
            <p class="positions-subtext">
              Recent settlement windows, showing status and net flow. A dedicated Analytics Terminal will later
              host full interactive charts and detailed filters.
            </p>
            <table class="rc-table">
              <thead>
                <tr>
                  <th>Window</th>
                  <th>Status</th>
                  <th>Profile</th>
                  <th>Net Flow (GRC)</th>
                  <th>Last Settlement NAV</th>
                </tr>
              </thead>
              <tbody>
                ${
                  (winHist && winHist.length)
                    ? winHist.slice(-10).reverse().map(w => `
                      <tr>
                        <td>#${w.window_id}</td>
                        <td>${w.status || 'Open'}</td>
                        <td>${w.profile || state.profile}</td>
                        <td>${fmt(w.net_flow || 0, 4)} GRC</td>
                        <td>${fmt(w.settlement_nav || (state.nav && state.nav.grc) || 0, 6)}</td>
                      </tr>
                    `).join('')
                    : '<tr><td colspan="5">No corridor history yet on this DevNet session.</td></tr>'
                }
              </tbody>
            </table>
          </section>
        `;
    }

function renderDashboard(state) {
        const navHist = state.navHistory.map(p => ({ y: p.grc }));
        const fxHist = state.fxHistory.map(p => ({ y: p.eur_usd }));
        const b = (state.balances && state.balances.balances) || {};
        const progress = windowProgressPercent(state);

        mainEl.innerHTML = `
            <section class="view-dashboard">
                <div class="panel-row">
                    <div class="panel card-large">
                        <div class="panel-header">
                            <h2>GRC NAV</h2>
                            <span class="pill">${state.profile}</span>
                        </div>
                        <div class="panel-body">
                            <div class="nav-value">${state.nav ? fmt(state.nav.grc, 6) : '--'}</div>
                            <div class="nav-sub">T-Bill Yield: ${state.yield ? fmt(state.yield.t_bill_yield * 100, 2) + '%' : '--'}</div>
                            <div class="spark-container">
                                ${sparkline(navHist, 200, 40, 'y')}
                            </div>
                        </div>
                    </div>
                    <div class="panel card-medium">
                        <div class="panel-header">
                            <h2>FX Monitor</h2>
                        </div>
                        <div class="panel-body">
                            <div class="fx-row">
                                <span>EUR/USD</span>
                                <span>${state.fx ? fmt(state.fx.eur_usd, 5) : '--'}</span>
                            </div>
                            <div class="spark-container">
                                ${sparkline(fxHist, 200, 40, 'y')}
                            </div>
                        </div>
                    </div>
                    <div class="panel card-medium">
                        <div class="panel-header">
                            <h2>Window & Settlement</h2>
                        </div>
                        <div class="panel-body">
                            <p>Status: <strong>${state.windows ? state.windows.status : '--'}</strong></p>
                            <p>Window ID: #${state.windows ? state.windows.window_id : '--'}</p>
                            <p>Next Close: ${state.windows ? new Date((state.windows.next_close_unix || 0) * 1000).toLocaleTimeString() : '--'}</p>
                        </div>
                    </div>
                </div>

                <div class="panel-row">
                    <div class="panel card-large">
                        <div class="panel-header">
                            <h2>Corridor Windows</h2>
                            <span class="pill">${state.windows ? state.windows.mode : ''}</span>
                        </div>
                        <div class="panel-body">
                            <div class="window-progress">
                                <div class="bar" style="width:${progress}%;"></div>
                            </div>
                            ${renderWindowStrip(state)}
                        </div>
                    </div>
                    <div class="panel card-medium">
                        <div class="panel-header">
                            <h2>Wallet Snapshot</h2>
                        </div>
                        <div class="panel-body">
                            <div>USD: ${b.USD ?? 0}</div>
                            <div>GRC: ${b.GRC ?? 0}</div>
                            <div>EUR: ${b.EUR ?? 0}</div>
                            <div>USDC: ${b.USDC ?? 0}</div>
                        </div>
                    </div>
                </div>
            </section>
        `;
    }

    function renderWallet(state) {
        const b = (state.balances && state.balances.balances) || {};
        mainEl.innerHTML = `
            <section class="view-wallet">
                <h1>Wallet</h1>
                <p>Address: <code>${state.address}</code></p>
                <div class="wallet-balances">
                    <div>USD: ${b.USD ?? 0}</div>
                    <div>GRC: ${b.GRC ?? 0}</div>
                    <div>EUR: ${b.EUR ?? 0}</div>
                    <div>USDC: ${b.USDC ?? 0}</div>
                </div>
            </section>
        `;
    }

    function renderTreasury(state) {
        const b = (state.balances && state.balances.balances) || {};
        mainEl.innerHTML = `
            <section class="view-treasury">
                <h1>Mint / Redeem GRC</h1>
                <p>USD Balance: ${b.USD ?? 0}</p>
                <p>GRC Balance: ${b.GRC ?? 0}</p>
                <div class="treasury-actions">
                    <div>
                        <h2>Mint GRC from USD</h2>
                        <input id="mint-amount" type="number" step="0.01" placeholder="Amount in USD">
                        <button id="mint-btn">Mint</button>
                    </div>
                    <div>
                        <h2>Redeem GRC to USD</h2>
                        <input id="redeem-amount" type="number" step="0.01" placeholder="Amount in GRC">
                        <button id="redeem-btn">Redeem</button>
                    </div>
                </div>
            </section>
        `;

        document.getElementById('mint-btn').addEventListener('click', () => {
            const v = parseFloat(document.getElementById('mint-amount').value);
            if (!isNaN(v) && v > 0) mintDemo(v);
        });

        document.getElementById('redeem-btn').addEventListener('click', () => {
            const v = parseFloat(document.getElementById('redeem-amount').value);
            if (!isNaN(v) && v > 0) redeemDemo(v);
        });
    }

    function renderExplorer(state) {
        const windows = (state.windowHistory || []).slice().reverse();
        const events = state.eventLog || [];

        const windowRows = windows.map(w => `
            <tr data-window-id="${w.window_id}">
                <td>#${w.window_id}</td>
                <td>${w.status}</td>
                <td>${w.mode || ''}</td>
                <td>${w.pending_volume_usd != null ? fmt(w.pending_volume_usd, 2) : '--'}</td>
                <td>${w.last_settle_unix ? new Date(w.last_settle_unix * 1000).toLocaleTimeString() : '--'}</td>
            </tr>
        `).join('');

        const eventRows = events.map(e => {
            const ts = e.ts ? new Date(e.ts).toLocaleTimeString() : '';
            let detail = '';
            if (e.type === 'Mint' || e.type === 'Redeem') {
                const p = e.payload;
                detail = `${p.amount ?? ''} ${p.from ?? ''} -> ${p.to ?? ''} (${p.address ?? ''})`;
            }
            return `
                <tr>
                    <td>${ts}</td>
                    <td>${e.type}</td>
                    <td>${detail}</td>
                </tr>
            `;
        }).join('');

        mainEl.innerHTML = `
            <section class="view-explorer">
                <h1>Explorer</h1>

                <div class="panel-row">
                    <div class="panel card-large">
                        <div class="panel-header">
                            <h2>Windows</h2>
                        </div>
                        <div class="panel-body">
                            <table class="explorer-table" id="windows-table">
                                <thead>
                                    <tr>
                                        <th>ID</th>
                                        <th>Status</th>
                                        <th>Mode</th>
                                        <th>Pending Vol (USD)</th>
                                        <th>Last Settle</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    ${windowRows || '<tr><td colspan="5">No windows yet</td></tr>'}
                                </tbody>
                            </table>
                        </div>
                    </div>
                    <div class="panel card-large">
                        <div class="panel-header">
                            <h2>Events</h2>
                        </div>
                        <div class="panel-body">
                            <table class="explorer-table">
                                <thead>
                                    <tr>
                                        <th>Time</th>
                                        <th>Type</th>
                                        <th>Details</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    ${eventRows || '<tr><td colspan="3">No events yet</td></tr>'}
                                </tbody>
                            </table>
                        </div>
                    </div>
                </div>

                <div class="panel-row">
                    <div class="panel card-large">
                        <div class="panel-header">
                            <h2>Settlement Ticket</h2>
                        </div>
                        <div class="panel-body" id="settlement-ticket">
                            <p>Select a window to view details.</p>
                        </div>
                    </div>
                </div>
            </section>
        `;

        const table = document.getElementById('windows-table');
        const ticketEl = document.getElementById('settlement-ticket');

        if (table) {
            table.querySelectorAll('tbody tr[data-window-id]').forEach(row => {
                row.addEventListener('click', () => {
                    const id = parseInt(row.getAttribute('data-window-id'), 10);
                    const w = (window.ReserveStore.state.windowHistory || []).find(x => x.window_id === id);
                    if (!w) return;
                    const opened = w.opened_at_unix ? new Date(w.opened_at_unix * 1000).toLocaleString() : '--';
                    const nextClose = w.next_close_unix ? new Date(w.next_close_unix * 1000).toLocaleString() : '--';
                    const lastSettle = w.last_settle_unix ? new Date(w.last_settle_unix * 1000).toLocaleString() : '--';

                    ticketEl.innerHTML = `
                        <div class="ticket">
                            <div class="ticket-row">
                                <span class="label">Window ID</span>
                                <span class="value">#${w.window_id}</span>
                            </div>
                            <div class="ticket-row">
                                <span class="label">Status</span>
                                <span class="value">${w.status}</span>
                            </div>
                            <div class="ticket-row">
                                <span class="label">Mode</span>
                                <span class="value">${w.mode || ''}</span>
                            </div>
                            <div class="ticket-row">
                                <span class="label">Profile</span>
                                <span class="value">${w.profile || window.ReserveStore.state.profile}</span>
                            </div>
                            <div class="ticket-row">
                                <span class="label">Opened</span>
                                <span class="value">${opened}</span>
                            </div>
                            <div class="ticket-row">
                                <span class="label">Next Close (planned)</span>
                                <span class="value">${nextClose}</span>
                            </div>
                            <div class="ticket-row">
                                <span class="label">Last Settled</span>
                                <span class="value">${lastSettle}</span>
                            </div>
                            <div class="ticket-row">
                                <span class="label">Pending Volume (USD)</span>
                                <span class="value">${w.pending_volume_usd != null ? fmt(w.pending_volume_usd, 2) : '--'}</span>
                            </div>
                            <p class="ticket-note">
                                DevNet: this ticket summarizes the current window state. 
                                MainNet can extend this with per-window deposits, redeems, and price improvement.
                            </p>
                        </div>
                    `;
                });
            });
        }
    }

    function render(state) {
        renderStatusStrip(state);
        if (currentView === 'dashboard') {
            renderDashboard(state);
        } else if (currentView === 'wallet') {
            renderWallet(state);
        } else if (currentView === 'treasury') {
            renderTreasury(state);
        } else if (currentView === 'positions') {
            renderPositions(state);
        } else if (currentView === 'analytics') {
            renderAnalytics(state);
        } else if (currentView === 'trading-terminal') {
            renderTradingTerminalBoot(state);
        } else if (currentView === 'explorer') {
            renderExplorer(state);
        }
    }

    navButtons.forEach(btn => {
        btn.addEventListener('click', () => {
            currentView = btn.dataset.view;
            render(window.ReserveStore.state);
        });
    });

    profileSpans.forEach(span => {
        span.addEventListener('click', () => {
            const profile = span.dataset.profile;
            window.ReserveStore.setProfile(profile);
        });
    });



    function renderTradingTerminalBoot(state) {
        mainEl.innerHTML = `
            <section class="view-terminal-boot">
                <h1>Launching Trading Terminal…</h1>
                <p class="boot-sub">
                    Initializing chart engine, attaching feeds, and preparing on-chart order controls.
                </p>
                <div class="boot-steps">
                    <div class="boot-step">✓ Connected to devnet node ${state.node_id || ''}</div>
                    <div class="boot-step">• Syncing NAV &amp; corridor windows…</div>
                    <div class="boot-step">• Attaching workstation routing layer…</div>
                    <div class="boot-step">• Preparing Trading Terminal UI…</div>
                </div>
                <div class="boot-progress">
                    <div class="boot-progress-bar"></div>
                </div>
            </section>
        `;

        // after short delay, open popup and then mark as running
        setTimeout(() => {
            openTradingTerminalPopup(state);
        }, 700);
    }

    function renderTradingTerminalRunning(state) {
        mainEl.innerHTML = `
            <section class="view-terminal-running">
                <h1>Trading Terminal Running</h1>
                <p>The Trading Terminal is open in a dedicated window.</p>
                <div class="terminal-status-grid">
                    <div class="status-card">
                        <span class="label">Environment</span>
                        <span class="value">${state.network || 'DevNet'}</span>
                    </div>
                    <div class="status-card">
                        <span class="label">Connected Node</span>
                        <span class="value">${state.node_id || '--'}</span>
                    </div>
                    <div class="status-card">
                        <span class="label">Ping</span>
                        <span class="value">${state.tickLatencyMs ? state.tickLatencyMs + ' ms' : '--'}</span>
                    </div>
                </div>
                <button class="btn-primary" id="btn-reopen-terminal">Re-open Trading Terminal</button>
            </section>
        `;
        const btn = document.getElementById('btn-reopen-terminal');
        if (btn) {
            btn.addEventListener('click', () => {
                openTradingTerminalPopup(window.ReserveStore.state);
            });
        }
    }

    function openTradingTerminalPopup(state) {
        const w = window.screen.availWidth || 1280;
        const h = window.screen.availHeight || 720;
        const features = [
            'toolbar=no',
            'menubar=no',
            'location=no',
            'status=no',
            'resizable=yes',
            'scrollbars=yes',
            `width=${w}`,
            `height=${h}`,
            'top=0',
            'left=0'
        ].join(',');

        const win = window.open('/trading-terminal/trading_terminal.php', 'reservechain_trading_terminal', features);
        if (win && win.focus) {
            win.focus();
        }

        // Mark as running in main workstation view
        setTimeout(() => {
            renderTradingTerminalRunning(state);
        }, 600);
    }


    window.ReserveStore.subscribe(render);
})();
