
// positions_panel.js
// Simple renderer that mirrors MarginContext.positions into the positions panel.

(function () {
  function renderPositions() {
    if (!window.MarginContext) return;
    const panel = document.getElementById('positions-panel');
    const body = document.getElementById('positions-body');
    if (!panel || !body) return;

    body.innerHTML = '';
    const positions = MarginContext.positions || [];
    if (!positions.length) {
      const empty = document.createElement('div');
      empty.className = 'positions-empty';
      empty.textContent = 'No open positions';
      body.appendChild(empty);
      return;
    }

    positions.forEach((p) => {
      const row = document.createElement('div');
      row.className = 'positions-row';
      row.innerHTML = `
        <div>${p.symbol}</div>
        <div>${p.side}</div>
        <div>${p.qty}</div>
        <div>${p.entryPrice}</div>
        <div>${p.markPrice}</div>
        <div class="${p.unrealizedPnl >= 0 ? 'pnl-pos' : 'pnl-neg'}">${p.unrealizedPnl.toFixed(4)}</div>
      `;
      body.appendChild(row);
    });
  }

  // Expose a hook MarginContext can call after snapshots if desired
  window.renderPositionsPanel = renderPositions;

  document.addEventListener('DOMContentLoaded', () => {
    // Poll every 1s as a simple starting point
    setInterval(renderPositions, 1000);
  });
})();
