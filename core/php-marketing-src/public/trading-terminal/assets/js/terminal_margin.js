
// terminal_margin.js
// Margin / PnL / liquidation model with hybrid (C) rules (devnet shell).

(function () {
  const DEFAULT_BASE_EQUITY = 100000; // demo starting equity
  const DEFAULT_LEVERAGE = 5;

  function safeNumber(x, fallback = 0) {
    const n = Number(x);
    return Number.isFinite(n) ? n : fallback;
  }

  const MarginContext = {
    baseEquity: DEFAULT_BASE_EQUITY,
    equity: DEFAULT_BASE_EQUITY,
    marginUsed: 0,
    marginFree: DEFAULT_BASE_EQUITY,
    exposureNotional: 0,
    totalFees: 0,
    lastPrice: null,
    positions: [],
    lastSnapshot: null,

    // Apply a single fill coming from WS 'execution' channel.
    // fill = { order_id, symbol, side, qty, price, fee, ts }
    applyFill(fill) {
      if (!fill) return;

      const symbol = fill.symbol || 'GRC-USD';
      const side = (fill.side || 'buy').toLowerCase();
      const qty = safeNumber(fill.qty, 0);
      const px = safeNumber(fill.price, this.lastPrice || 1.0);
      const fee = safeNumber(fill.fee, 0);

      if (!qty || !px) return;

      // Track cumulative fees (reduce equity)
      this.totalFees += fee;

      let pos = this.positions.find((p) => p.symbol === symbol);
      const dir = side === 'buy' ? 1 : -1;

      if (!pos) {
        // New position
        pos = {
          symbol,
          side: dir > 0 ? 'long' : 'short',
          qty,
          entryPrice: px,
          markPrice: px,
          unrealizedPnl: 0,
        };
        this.positions.push(pos);
      } else {
        // Netting logic
        const existingDir = pos.side === 'long' ? 1 : -1;
        const existingQty = pos.qty;

        if (existingDir === dir) {
          // Same direction → size up & blend entry price
          const newQty = existingQty + qty;
          const weightedEntry =
            (pos.entryPrice * existingQty + px * qty) / newQty;
          pos.qty = newQty;
          pos.entryPrice = weightedEntry;
          pos.side = existingDir > 0 ? 'long' : 'short';
        } else {
          // Opposite direction → close or flip
          if (qty < existingQty) {
            // Partial close
            const remaining = existingQty - qty;
            pos.qty = remaining;
            // entryPrice unchanged
          } else if (qty === existingQty) {
            // Fully closed
            this.positions = this.positions.filter((p) => p !== pos);
            pos = null;
          } else {
            // Flip through zero
            const leftover = qty - existingQty;
            pos.side = dir > 0 ? 'long' : 'short';
            pos.qty = leftover;
            pos.entryPrice = px;
          }
        }
      }

      if (this.lastPrice != null) {
        this.updateMark(this.lastPrice);
      } else {
        this.updateMark(px);
      }

      this.snapshot('fill');
    },

    // Called whenever we receive a new mark price.
    updateMark(price) {
      const px = safeNumber(price, this.lastPrice || 1.0);
      this.lastPrice = px;

      let totalUnreal = 0;
      let totalNotional = 0;

      this.positions.forEach((p) => {
        p.markPrice = px;
        const notional = p.qty * px;
        totalNotional += notional;

        let pnl = 0;
        if (p.side === 'long') {
          pnl = (px - p.entryPrice) * p.qty;
        } else {
          pnl = (p.entryPrice - px) * p.qty;
        }
        p.unrealizedPnl = pnl;
        totalUnreal += pnl;
      });

      this.exposureNotional = totalNotional;
      this.equity = this.baseEquity + totalUnreal - this.totalFees;

      const marginPerPos = this.positions.reduce((acc, p) => {
        const n = p.qty * this.lastPrice;
        return acc + n / DEFAULT_LEVERAGE;
      }, 0);

      this.marginUsed = marginPerPos;
      this.marginFree = this.equity - this.marginUsed;

      this.snapshot('price');
    },

    snapshot(eventName) {
      this.lastSnapshot = {
        event: eventName,
        ts: Date.now(),
        equity: this.equity,
        marginUsed: this.marginUsed,
        marginFree: this.marginFree,
        exposureNotional: this.exposureNotional,
        positions: this.positions.map((p) => ({ ...p })),
      };

      if (typeof window.renderPositionsPanel === 'function') {
        try {
          window.renderPositionsPanel();
        } catch (e) {
          console.error('renderPositionsPanel error', e);
        }
      }

      if (window.console && console.groupCollapsed) {
        console.groupCollapsed('[MarginContext] snapshot:', eventName);
        console.log(this.lastSnapshot);
        console.groupEnd();
      }
    },
  };

  window.MarginContext = MarginContext;
})();
