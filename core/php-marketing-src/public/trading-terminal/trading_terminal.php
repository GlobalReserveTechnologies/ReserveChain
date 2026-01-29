<?php ?>
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Workstation — Trading Terminal</title>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link rel="stylesheet" href="/trading-terminal/assets/css/app.css">
    <link rel="stylesheet" href="/trading-terminal/assets/css/terminal.css">
</head>
<body class="terminal-window">

<div id="rc-terminal-boot" class="rc-boot-overlay">
  <div class="rc-boot-panel">
    <header class="rc-boot-header">
      <div class="rc-boot-logo-mark">
        <span class="rc-boot-logo-emblem">RC</span>
      </div>
      <div class="rc-boot-header-text">
        <div class="rc-boot-title">ReserveChain Terminal</div>
        <div class="rc-boot-subtitle">Global Reserve Currency Workstation</div>
      </div>
    </header>
    <section class="rc-boot-body">
      <div class="rc-boot-steps"></div>
      <div class="rc-boot-status-tray">
        <div class="rc-boot-status-group">
          <div class="rc-boot-status-label">System Layer</div>
          <div class="rc-boot-status-row" data-status-key="ws">
            <span class="rc-boot-status-name">WebSocket Channels</span>
            <span class="rc-boot-status-metrics" data-status-metrics="ws">RTT: — / Jitter: —</span>
            <span class="rc-boot-status-indicator" data-status-indicator="ws"></span>
          </div>
          <div class="rc-boot-status-row" data-status-key="chain">
            <span class="rc-boot-status-name">Chain Sync</span>
            <span class="rc-boot-status-metrics" data-status-metrics="chain">Height: — / Slot: —</span>
            <span class="rc-boot-status-indicator" data-status-indicator="chain"></span>
          </div>
          <div class="rc-boot-status-row" data-status-key="time">
            <span class="rc-boot-status-name">Time Sync</span>
            <span class="rc-boot-status-metrics" data-status-metrics="time">Offset: —</span>
            <span class="rc-boot-status-indicator" data-status-indicator="time"></span>
          </div>
        </div>
        <div class="rc-boot-status-group">
          <div class="rc-boot-status-label">Financial Layer</div>
          <div class="rc-boot-status-row" data-status-key="feed">
            <span class="rc-boot-status-name">Price Feed</span>
            <span class="rc-boot-status-metrics" data-status-metrics="feed">Freshness: —</span>
            <span class="rc-boot-status-indicator" data-status-indicator="feed"></span>
          </div>
          <div class="rc-boot-status-row" data-status-key="exec">
            <span class="rc-boot-status-name">Execution Routing</span>
            <span class="rc-boot-status-metrics" data-status-metrics="exec">Last ACK: —</span>
            <span class="rc-boot-status-indicator" data-status-indicator="exec"></span>
          </div>
          <div class="rc-boot-status-row" data-status-key="margin">
            <span class="rc-boot-status-name">Margin Engine</span>
            <span class="rc-boot-status-metrics" data-status-metrics="margin">Mode: —</span>
            <span class="rc-boot-status-indicator" data-status-indicator="margin"></span>
          </div>
          <div class="rc-boot-status-row" data-status-key="reserve">
            <span class="rc-boot-status-name">Reserve Sync</span>
            <span class="rc-boot-status-metrics" data-status-metrics="reserve">Reserve: —</span>
            <span class="rc-boot-status-indicator" data-status-indicator="reserve"></span>
          </div>
        </div>
      </div>
      <div class="rc-boot-footer">
        <span class="rc-boot-footer-text">Bringing ReserveChain online…</span>
      </div>
    </section>
  </div>
</div>

<div id="rc-metrics-hud" class="rc-metrics rc-metrics-collapsed">
  <button type="button" class="rc-metrics-toggle" id="rc-metrics-toggle">
    Metrics ▸
  </button>
  <div class="rc-metrics-body" id="rc-metrics-body">
    <div class="rc-metrics-row">
      <span class="rc-metrics-label">RTT</span>
      <span class="rc-metrics-value" data-hud="rtt">—</span>
    </div>
    <div class="rc-metrics-row">
      <span class="rc-metrics-label">Feed</span>
      <span class="rc-metrics-value" data-hud="feed_fresh">—</span>
    </div>
    <div class="rc-metrics-row">
      <span class="rc-metrics-label">ACK</span>
      <span class="rc-metrics-value" data-hud="ack">—</span>
    </div>
    <div class="rc-metrics-row">
      <span class="rc-metrics-label">Slot</span>
      <span class="rc-metrics-value" data-hud="slot">—</span>
    </div>
    <div class="rc-metrics-row">
      <span class="rc-metrics-label">Reserve</span>
      <span class="rc-metrics-value" data-hud="reserve">—</span>
    </div>
  </div>
</div>

<div id="rc-reserve-tooltip" class="rc-reserve-tooltip"></div>

<div class="terminal-chrome">
    <div class="terminal-bar">
        <div class="terminal-title">Workstation — Trading Terminal</div>
        <div class="terminal-actions">
            <button class="terminal-btn" onclick="window.close()">Close</button>
        </div>
    </div>
    <div class="terminal-body">

  <div class="terminal-wrapper">
    <!-- TOP BAR -->
    <div class="top-bar">
      <div class="top-left">
        <div class="logo-wrap">
          <div class="logo"></div>
          <div class="logo-ticker">GRC</div>
        </div>
        <div>
          <div class="top-title">ReserveChain · Global Reserve Coin</div>
          <div class="top-subtitle top-market-label">GRC/USDT · ReserveNet Terminal</div>
        </div>
        <div class="top-stat">24h P&L <b>+2.48%</b></div>
        <div class="top-stat">Coverage <b>102.3%</b></div>
      </div>

      <div class="tab-strip">
        <div class="tab active" data-symbol="GRC/USDT">
          <span class="pair">GRC/USDT</span>
          <span class="change">+3.42%</span>
        </div>
        <div class="tab" data-symbol="GRC/RSV">
          <span class="pair">GRC/RSV</span>
          <span class="change">+0.88%</span>
        </div>
        <div class="tab" data-symbol="GRC/BTC">
          <span class="pair">GRC/BTC</span>
          <span class="change">+1.12%</span>
        </div>
        <div class="tab" data-symbol="GRC/ETH">
          <span class="pair">GRC/ETH</span>
          <span class="change">+0.39%</span>
        </div>
        <div class="tab" data-symbol="GRC/VLT">
          <span class="pair">GRC/VLT</span>
          <span class="change">+5.01%</span>
        </div>
      </div>

      <div class="top-right">
        <div class="balance-pill">
          <span>Portfolio Equity</span>
          <span>₲1,376,751.38 GRC</span>
        </div>
        <div class="icon-btn" title="Refresh">⟳</div>
        <div class="icon-btn" title="Settings">⚙</div>
        <div class="icon-btn" title="Layout">☰</div>
      </div>
    </div>

    <!-- MAIN AREA -->
    <div class="main-row">
      <!-- LEFT TOOLBAR -->
      <div class="left-toolbar">
        <div class="tool-icon active" title="Pointer">✚</div>
        <div class="tool-icon" title="Trendline">╱</div>
        <div class="tool-icon" title="Zone">▱</div>
        <div class="tool-icon" title="Text">T</div>
        <div class="tool-icon" title="Fibonacci">＋</div>
        <div class="tool-icon" title="Curve">∿</div>
        <div class="tool-icon" title="Measure %">%</div>
        <div class="tool-icon" title="Auto tools">⚡</div>
        <div class="left-toolbar-bottom">
          <div class="tool-icon" title="Preferences">⚙</div>
          <div class="tool-icon" title="Help">?</div>
        </div>
      </div>

      <!-- CENTER PANEL -->
      <div class="center-panel">
        <!-- PAIR BAR -->
        <div class="pair-bar">
          <div class="pair-main">
            <span class="pair-main-symbol">GRC/USDT · 1D · ReserveNet</span>
            <span>Global Reserve Corridor · Coverage SMA 9 · 8.18M GRC</span>
          </div>

          <div class="tf-buttons">
            <button class="tf-btn">1m</button>
            <button class="tf-btn">5m</button>
            <button class="tf-btn">15m</button>
            <button class="tf-btn">1h</button>
            <button class="tf-btn active">1D</button>
            <button class="tf-btn">1W</button>
          </div>

          <div class="pair-bar-section">
            <label>O</label><span>102.37</span>
            <label>H</label><span>108.42</span>
            <label>L</label><span>100.12</span>
            <label>C</label><span>107.88</span>
          </div>

          <div class="pair-bar-right">
            <div class="pair-bar-pill active">Indicators</div>
            <div class="pair-bar-pill">Alerts</div>
            <div class="pair-bar-pill">Orders</div>
            <div class="pair-bar-pill">Strategy</div>
            <span>GRC</span>
          </div>
        </div>

        <!-- CHART WRAPPER -->
        <div class="chart-wrapper">
          <!-- SMALL TOOLBAR ABOVE CHART -->
          <div class="chart-toolbar">
            <div class="chart-toolbar-group">
              <span class="label">Leverage</span>
              <span class="value chart-leverage">1.0× · Spot</span>
            </div>
            <div class="chart-toolbar-group">
              <span class="label">Corridor</span>
              <span class="value">82.00 → 134.00 GRC</span>
            </div>
            <div class="chart-toolbar-group">
              <div class="dot"></div>
              <span class="value">On-chain: synced</span>
            </div>
          </div>

          <!-- CHART MAIN AREA -->
          <div class="chart-main">
            <!-- price axis labels -->
            <div class="price-axis">
              <div class="price-label green">135.00</div>
              <div class="price-label">120.00</div>
              <div class="price-label yellow">Corridor 100.00</div>
              <div class="price-label">86.50</div>
              <div class="price-label">72.00</div>
            </div>
            <!-- fake volume band -->
            <div class="chart-volume"></div>
            <div class="toast" id="toast">Simulated action – no live orders placed.</div>
          </div>

          <!-- BOTTOM BAR -->
          <div class="chart-bottom-bar">
            <div class="chart-bottom-left">
              <button class="mini-btn active">5y</button>
              <button class="mini-btn">1y</button>
              <button class="mini-btn">6m</button>
              <button class="mini-btn">3m</button>
              <button class="mini-btn">1m</button>
              <button class="mini-btn">5d</button>
              <button class="mini-btn">1d</button>
            </div>
            <div class="chart-bottom-right">
              <span>log</span>
              <span>auto</span>
            </div>
          </div>
        </div>
      </div>

      <!-- RIGHT PANEL -->
      <div class="right-panel">
        <div class="right-top">
          <div class="right-top-line">
            <span>Today (realized)</span>
            <span id="rc-today-realized" class="value green">+25.5 GRC</span>
          </div>
          <div class="right-top-line">
            <span>Account Equity</span>
            <span id="rc-account-equity" class="value">₲1,376,751.38</span>
          </div>

          <div class="right-tabs">
            <div class="right-tab active" data-order-tab="limit">Limit</div>
            <div class="right-tab" data-order-tab="market">Market</div>
            <div class="right-tab" data-order-tab="stop">Stop</div>
            <div class="right-tab" data-order-tab="tpsl">TP/SL</div>
          </div>

          <div class="order-mode-tabs">
            <div class="order-mode-tab active">Cross</div>
            <div class="order-mode-tab">Isolated</div>
            <div class="order-mode-tab">Vault TP/SL</div>
          </div>
        </div>

        <div class="order-form">
          <div class="order-input">
            <input type="text" value="1,376,751.38" />
            <span class="unit">GRC</span>
          </div>

          <div class="order-label-row">
            <span class="price-label-text">Price</span>
            <span class="price-quote-text">Quote (USDT)</span>
          </div>
          <div class="order-input">
            <input type="text" value="107.88" />
            <span class="unit">USDT</span>
          </div>

          <div class="order-label-row">
            <span class="size-label-text">Size</span>
            <span class="size-quote-text">Base (GRC)</span>
          </div>
          <div class="order-input">
            <input type="text" value="0.000" />
            <span class="unit">GRC</span>
          </div>

          <div class="percent-row">
            <div class="percent-chip">25%</div>
            <div class="percent-chip">50%</div>
            <div class="percent-chip">75%</div>
            <div class="percent-chip">100%</div>
          </div>

          <div class="side-row">
            <button class="btn-main btn-sell">Sell</button>
            <button class="btn-main btn-buy">Buy</button>
          </div>
        </div>

        <div class="venue-switch">
          <div>Venue: ReserveNet · GRC Spot</div>
          <div class="venue-row">
            <div class="venue-pill active">USDT</div>
            <div class="venue-pill">RSV</div>
            <div class="venue-pill">BTC</div>
          </div>

          <div class="symbol-search">
            <input type="text" placeholder="Search GRC markets" />
            <select>
              <option>All</option>
              <option>Spot</option>
              <option>Perp</option>
            </select>
          </div>
        </div>

        <div class="market-table-wrapper">
          <div class="market-header">
            <span>Market</span>
            <span>Volume</span>
            <span>Change</span>
          </div>
          <div class="market-table">
            <div class="market-row active" data-symbol="GRC/USDT">
              <span>GRC/USDT</span>
              <span>315.6M</span>
              <span class="chg-pos">+3.42%</span>
            </div>
            <div class="market-row" data-symbol="GRC/RSV">
              <span>GRC/RSV</span>
              <span>73.8M</span>
              <span class="chg-pos">+0.88%</span>
            </div>
            <div class="market-row" data-symbol="GRC/BTC">
              <span>GRC/BTC</span>
              <span>45.0M</span>
              <span class="chg-pos">+1.12%</span>
            </div>
            <div class="market-row" data-symbol="GRC/ETH">
              <span>GRC/ETH</span>
              <span>51.2M</span>
              <span class="chg-neg">-0.37%</span>
            </div>
            <div class="market-row" data-symbol="GRC/VLT">
              <span>GRC/VLT</span>
              <span>20.9M</span>
              <span class="chg-pos">+5.01%</span>
            </div>
            <div class="market-row" data-symbol="GRC/DEF">
              <span>GRC/DEF</span>
              <span>12.4M</span>
              <span class="chg-pos">+1.44%</span>
            </div>
            <div class="market-row" data-symbol="GRC/EDGE">
              <span>GRC/EDGE</span>
              <span>9.3M</span>
              <span class="chg-neg">-0.82%</span>
            </div>
            <div class="market-row" data-symbol="GRC/USD (index)">
              <span>GRC/USD (index)</span>
              <span>502.6M</span>
              <span class="chg-pos">+2.09%</span>
            </div>
          </div>
        </div>
      </div>

    </div>
  </div>

  <script>
    document.addEventListener("DOMContentLoaded", function () {
      const tabs = document.querySelectorAll(".tab-strip .tab");
      const marketLabel = document.querySelector(".top-market-label");
      const pairMainSymbol = document.querySelector(".pair-main-symbol");
      const tfButtons = document.querySelectorAll(".tf-btn");
      const miniButtons = document.querySelectorAll(".mini-btn");
      const pairBarPills = document.querySelectorAll(".pair-bar-pill");
      const toolIcons = document.querySelectorAll(".left-toolbar .tool-icon");
      const rightTabs = document.querySelectorAll(".right-tab");
      const orderModeTabs = document.querySelectorAll(".order-mode-tab");

const layoutBtn = document.querySelector('.icon-btn[title="Layout"]');
if (layoutBtn) {
  layoutBtn.addEventListener('click', () => {
    window.open('/trading-terminal/multipanel.php', 'tradingTerminalMulti', 'width=1600,height=900,resizable=yes');
  });
}


      const percentChips = document.querySelectorAll(".percent-chip");
      const venuePills = document.querySelectorAll(".venue-pill");
      const marketRows = document.querySelectorAll(".market-row");
      const buyBtn = document.querySelector(".btn-buy");
      const sellBtn = document.querySelector(".btn-sell");
      const toast = document.getElementById("toast");

      const priceLabelText = document.querySelector(".price-label-text");
      const priceQuoteText = document.querySelector(".price-quote-text");
      const sizeLabelText = document.querySelector(".size-label-text");
      const sizeQuoteText = document.querySelector(".size-quote-text");

      function showToast(message) {
        if (!toast) return;
        toast.textContent = message;
        toast.classList.add("show");
        setTimeout(() => toast.classList.remove("show"), 1600);
      }

      // Top symbol tabs
      tabs.forEach((tab) => {
        tab.addEventListener("click", () => {
          tabs.forEach((t) => t.classList.remove("active"));
          tab.classList.add("active");
          const pairEl = tab.querySelector(".pair");
          const pair = pairEl ? pairEl.textContent.trim() : tab.getAttribute("data-symbol") || "GRC/USDT";
          marketLabel.textContent = pair + " · ReserveNet Terminal";
          const activeTf = document.querySelector(".tf-btn.active");
          const tfText = activeTf ? activeTf.textContent.trim() : "1D";
          pairMainSymbol.textContent = pair + " · " + tfText + " · ReserveNet";

          // also highlight in market list
          marketRows.forEach((row) => {
            const sym = row.getAttribute("data-symbol");
            row.classList.toggle("active", sym === pair);
          });
        });
      });

      // Timeframe buttons
      tfButtons.forEach((btn) => {
        btn.addEventListener("click", () => {
          tfButtons.forEach((b) => b.classList.remove("active"));
          btn.classList.add("active");
          const activeTab = document.querySelector(".tab-strip .tab.active .pair");
          const pair = activeTab ? activeTab.textContent.trim() : "GRC/USDT";
          const tfText = btn.textContent.trim();
          pairMainSymbol.textContent = pair + " · " + tfText + " · ReserveNet";
        });
      });

      // Bottom mini range buttons
      miniButtons.forEach((btn) => {
        btn.addEventListener("click", () => {
          miniButtons.forEach((b) => b.classList.remove("active"));
          btn.classList.add("active");
        });
      });

      // Pair bar pills (Indicators / Alerts / Orders / Strategy)
      pairBarPills.forEach((pill) => {
        pill.addEventListener("click", () => {
          pairBarPills.forEach((p) => p.classList.remove("active"));
          pill.classList.add("active");
        });
      });

      // Left toolbar tools
      toolIcons.forEach((icon) => {
        icon.addEventListener("click", () => {
          // Only one active at a time in the main set; bottom icons behave the same here for simplicity
          toolIcons.forEach((i) => i.classList.remove("active"));
          icon.classList.add("active");
        });
      });

      // Right-side order tabs (Limit / Market / Stop / TP/SL)
      rightTabs.forEach((tab) => {
        tab.addEventListener("click", () => {
          rightTabs.forEach((t) => t.classList.remove("active"));
          tab.classList.add("active");

          const mode = tab.getAttribute("data-order-tab");
          switch (mode) {
            case "market":
              priceLabelText.textContent = "Price";
              priceQuoteText.textContent = "Market";
              sizeLabelText.textContent = "Size";
              sizeQuoteText.textContent = "Base (GRC)";
              break;
            case "stop":
              priceLabelText.textContent = "Stop Price";
              priceQuoteText.textContent = "Trigger";
              sizeLabelText.textContent = "Size";
              sizeQuoteText.textContent = "Base (GRC)";
              break;
            case "tpsl":
              priceLabelText.textContent = "TP / SL Levels";
              priceQuoteText.textContent = "Targets";
              sizeLabelText.textContent = "Linked Size";
              sizeQuoteText.textContent = "Base (GRC)";
              break;
            default:
              priceLabelText.textContent = "Price";
              priceQuoteText.textContent = "Quote (USDT)";
              sizeLabelText.textContent = "Size";
              sizeQuoteText.textContent = "Base (GRC)";
          }
        });
      });

      // Order mode tabs (Cross / Isolated / Vault TP/SL)
      orderModeTabs.forEach((tab) => {
        tab.addEventListener("click", () => {
          orderModeTabs.forEach((t) => t.classList.remove("active"));
          tab.classList.add("active");
          showToast("Margin mode set to: " + tab.textContent.trim());
        });
      });

      // Percent chips
      percentChips.forEach((chip) => {
        chip.addEventListener("click", () => {
          percentChips.forEach((c) => c.classList.remove("active"));
          chip.classList.add("active");
        });
      });

      // Venue pills
      venuePills.forEach((pill) => {
        pill.addEventListener("click", () => {
          venuePills.forEach((p) => p.classList.remove("active"));
          pill.classList.add("active");
          showToast("Venue switched to: " + pill.textContent.trim());
        });
      });

      // Market rows (change active pair)
      marketRows.forEach((row) => {
        row.addEventListener("click", () => {
          marketRows.forEach((r) => r.classList.remove("active"));
          row.classList.add("active");
          const sym = row.getAttribute("data-symbol") || row.children[0].textContent.trim();
          // Update top tabs to match if exists, otherwise just update labels
          let matched = false;
          tabs.forEach((tab) => {
            const pairEl = tab.querySelector(".pair");
            const pair = pairEl ? pairEl.textContent.trim() : "";
            if (pair === sym) {
              tabs.forEach((t) => t.classList.remove("active"));
              tab.classList.add("active");
              matched = true;
            }
          });
          const activeTf = document.querySelector(".tf-btn.active");
          const tfText = activeTf ? activeTf.textContent.trim() : "1D";
          pairMainSymbol.textContent = sym + " · " + tfText + " · ReserveNet";
          marketLabel.textContent = sym + " · ReserveNet Terminal";
        });
      });

      // Buy / Sell buttons
      
      function handleOrderButton(btn, side) {
        btn.addEventListener("click", async () => {
          btn.classList.add("pressed");
          setTimeout(() => btn.classList.remove("pressed"), 160);

          try {
            if (!window.SessionManager) {
              const msg = "Session manager not loaded. Open Workstation and set collateral first.";
              window.UINotifier ? UINotifier.error(msg) : alert(msg);
              return;
            }
            const sess = SessionManager.get();
            if (!SessionManager.isValid(sess)) {
              const msg = "No valid trading session. Open Workstation → Vaults to set collateral.";
              window.UINotifier ? UINotifier.warn(msg) : alert(msg);
              return;
            }
            if (!window.MarginContext) {
              const msg = "Margin engine not initialized.";
              window.UINotifier ? UINotifier.error(msg) : alert(msg);
              return;
            }

            const symbolEl = document.querySelector(".pair-main-symbol");
            const symbol = symbolEl ? symbolEl.textContent.trim() : "GRC/USDT";

            const inputs = document.querySelectorAll(".order-panel .order-input input");
            const priceVal = inputs[0] ? inputs[0].value : "0";
            const qtyVal = inputs[1] ? inputs[1].value : "0";
            const price = parseFloat(priceVal.replace(/,/g, "")) || 0;
            const qty = parseFloat(qtyVal.replace(/,/g, "")) || 0;

            if (!qty || !price) {
              const msg = "Enter a valid price and size before submitting an order.";
              window.UINotifier ? UINotifier.warn(msg) : alert(msg);
              return;
            }

            const sideDir = side === "Buy" ? "long" : "short";

            // Ensure margin engine has equity loaded at least once
            if (MarginContext.equity === 0 && MarginContext.marginFree === 0) {
              await MarginContext.loadInitialEquity();
            }

            const pos = await MarginContext.openPosition({
              symbol,
              side: sideDir,
              qty,
              price,
            });

            const msg = `Opened ${side} ${qty} ${symbol} @ ${price.toFixed(2)} (pos: ${pos.id})`;
            window.UINotifier ? UINotifier.info(msg) : alert(msg);
          } catch (err) {
            const msg = "Order rejected: " + (err && err.message ? err.message : "Unknown error");
            console.error(msg, err);
            window.UINotifier ? UINotifier.error(msg, err) : alert(msg);
          }
        });
      }


      handleOrderButton(buyBtn, "Buy");
      handleOrderButton(sellBtn, "Sell");
    });
  </script>

    
    <div class="positions-panel" id="positions-panel">
      <div class="positions-header">
        <div>Symbol</div>
        <div>Side</div>
        <div>Qty</div>
        <div>Entry</div>
        <div>Mark</div>
        <div>Unrealized</div>
      </div>
      <div class="positions-body" id="positions-body">
        <!-- Filled by MarginContext renderer -->
      </div>
    </div>

</div>
</div>

<script>
(function() {
  const bootOverlay = document.getElementById('rc-terminal-boot');
  const stepsContainer = bootOverlay ? bootOverlay.querySelector('.rc-boot-steps') : null;
  const mainTerminal = document.querySelector('.terminal-chrome');

  if (!bootOverlay || !stepsContainer || !mainTerminal) return;

  const BOOT_MODES = {
    FULL: 'full',
    FAST: 'fast',
    RESUME: 'resume',
  };

  function getBootMode() {
    try {
      const stored = localStorage.getItem('rc_terminal_boot_stage');
      if (!stored) return BOOT_MODES.FULL;
      if (stored === BOOT_MODES.FULL) return BOOT_MODES.FAST;
      if (stored === BOOT_MODES.FAST) return BOOT_MODES.RESUME;
      return BOOT_MODES.RESUME;
    } catch (e) {
      return BOOT_MODES.FULL;
    }
  }

  function setBootMode(next) {
    try {
      localStorage.setItem('rc_terminal_boot_stage', next);
    } catch (e) {}
  }

  const fullSteps = [
    "Loading trading terminal…..",
    "Initializing WebSocket channels…..",
    "Negotiating encryption keys…..",
    "Bootstrapping instrument catalog…..",
    "Syncing order book snapshots…..",
    "Calibrating margin engine…..",
    "Allocating risk buffers…..",
    "Registering execution handlers…..",
    "Mounting UI compositor…..",
    "Finalizing launch sequence…..",
    "Activating trading terminal….."
  ];

  const fastSteps = [
    "Resuming trading session…..",
    "Restoring data channels…..",
    "Terminal ready….."
  ];

  function createSteps(lines) {
    stepsContainer.innerHTML = '';
    return lines.map((text, idx) => {
      const row = document.createElement('div');
      row.className = 'rc-boot-step';
      const label = document.createElement('div');
      label.className = 'rc-boot-step-label';
      label.textContent = text;
      const icon = document.createElement('div');
      icon.className = 'rc-boot-step-icon';
      row.appendChild(label);
      row.appendChild(icon);
      stepsContainer.appendChild(row);
      return { row, label, icon, text, index: idx };
    });
  }

  function markStepActive(step) {
    step.row.classList.add('rc-boot-step-active');
    step.icon.classList.add('rc-boot-step-spinner');
  }

  function markStepDone(step) {
    step.row.classList.remove('rc-boot-step-active');
    step.icon.classList.remove('rc-boot-step-spinner');
    step.row.classList.add('rc-boot-step-done');
    step.icon.classList.add('rc-boot-step-done');
    step.icon.textContent = '✓';
  }

  function setStatusIndicator(key, state, text) {
    if (!bootOverlay) return;
    const metricEl = bootOverlay.querySelector('[data-status-metrics="' + key + '"]');
    const indicator = bootOverlay.querySelector('[data-status-indicator="' + key + '"]');
    if (metricEl && typeof text === 'string') {
      metricEl.textContent = text;
    }
    if (indicator) {
      indicator.classList.remove('rc-status-ok', 'rc-status-error', 'rc-status-pending');
      if (state === 'ok') indicator.classList.add('rc-status-ok');
      else if (state === 'error') indicator.classList.add('rc-status-error');
      else if (state === 'pending') indicator.classList.add('rc-status-pending');
    }
  }

  function fakeMetricsTick() {
    setStatusIndicator('ws', 'ok', 'RTT: 42ms / Jitter: 3ms');
    setStatusIndicator('chain', 'ok', 'Height: 96421 / Slot: 41');
    setStatusIndicator('time', 'ok', 'Offset: +3ms');
    setStatusIndicator('feed', 'ok', 'Freshness: 21ms');
    setStatusIndicator('exec', 'ok', 'Last ACK: 58ms');
    setStatusIndicator('margin', 'ok', 'Mode: Real-time');
    setStatusIndicator('reserve', 'ok', 'Reserve: 34.5M GRC ($34.5M)');
  }

  function revealTerminalAndEnd() {
    bootOverlay.classList.add('rc-boot-hidden');
    mainTerminal.classList.add('rc-main-visible');
    mainTerminal.classList.remove('rc-main-hidden');
  }

  async function runBoot(mode) {
    if (mode === BOOT_MODES.RESUME) {
      bootOverlay.classList.add('rc-boot-hidden');
      mainTerminal.classList.add('rc-main-visible');
      mainTerminal.classList.remove('rc-main-hidden');
      return;
    }

    const steps = createSteps(mode === BOOT_MODES.FULL ? fullSteps : fastSteps);
    fakeMetricsTick();

    const baseDelay = (mode === BOOT_MODES.FULL) ? 520 : 380;

    for (let i = 0; i < steps.length; i++) {
      const step = steps[i];
      markStepActive(step);
      await new Promise(r => setTimeout(r, baseDelay));
      markStepDone(step);
    }

    setTimeout(() => {
      revealTerminalAndEnd();
      if (mode === BOOT_MODES.FULL) setBootMode(BOOT_MODES.FAST);
      else if (mode === BOOT_MODES.FAST) setBootMode(BOOT_MODES.RESUME);
    }, 350);
  }

  document.addEventListener('DOMContentLoaded', () => {
    mainTerminal.classList.add('rc-main-hidden');
    const mode = getBootMode();
    runBoot(mode);
  });
})();
</script>
<script>
(function() {
  const hud = document.getElementById('rc-metrics-hud');
  const toggle = document.getElementById('rc-metrics-toggle');
  const body = document.getElementById('rc-metrics-body');

  if (!hud || !toggle || !body) return;

  function setHudValue(key, value) {
    const el = hud.querySelector('[data-hud="' + key + '"]');
    if (el) el.textContent = value;
  }

  toggle.addEventListener('click', () => {
    const collapsed = hud.classList.toggle('rc-metrics-collapsed');
    toggle.textContent = collapsed ? 'Metrics ▸' : 'Metrics ▾';
  });

  function fakeHudTick() {
    setHudValue('rtt', '43ms');
    setHudValue('feed_fresh', '18ms');
    setHudValue('ack', '54ms');
    setHudValue('slot', '41');
    setHudValue('reserve', '34.5M GRC ($34.5M)');
  }

  document.addEventListener('DOMContentLoaded', fakeHudTick);

  window.RC_MetricsHUD = { set: setHudValue };
})();
</script>

<script src="/assets/js/ui_notifier.js"></script>
<script src="/assets/js/api_client.js"></script>
<script src="/assets/js/vault_client.js"></script>
<script src="/assets/js/wallet_keystore.js"></script>
<script src="/assets/js/session_manager.js"></script>
<script src="/trading-terminal/assets/js/terminal_margin.js"></script>
<script src="/trading-terminal/assets/js/store.js"></script>
<script src="/trading-terminal/assets/js/app.js"></script>
<script src="/trading-terminal/assets/js/terminal.js"></script>
<script src="/trading-terminal/assets/js/positions_panel.js"></script>
<script>
  if (window.ReserveConnectWS) {
    window.ReserveConnectWS();
  }
</script>
</body>
</html>

