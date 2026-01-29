
// ReserveChain Trading Terminal JS
// Custom lightweight chart engine shell + on-chart order overlays.

(function () {
    console.log('Trading Terminal booted.');

    const root = document.querySelector('.terminal-body') || document.body;
    const canvas = root.querySelector('canvas#rsc-chart') || createChartCanvas(root);
    const ctx = canvas.getContext('2d');

    const todayRealizedEl = document.getElementById('rc-today-realized');
    const accountEquityEl = document.getElementById('rc-account-equity');

    function formatGRC(value) {
        if (value == null || !isFinite(value)) return '—';
        const abs = Math.abs(value);
        let scaled = value;
        let suffix = '';
        if (abs >= 1_000_000_000) {
            scaled = value / 1_000_000_000;
            suffix = 'B';
        } else if (abs >= 1_000_000) {
            scaled = value / 1_000_000;
            suffix = 'M';
        } else if (abs >= 1_000) {
            scaled = value / 1_000;
            suffix = 'K';
        }
        const s = scaled.toLocaleString(undefined, { maximumFractionDigits: 2 });
        return s + suffix + ' GRC';
    }

    const state = {
        candles: [],
        orders: [],
        positions: [],
        overlays: {
            entryLines: [],
            liquidationLines: [],
            breakevenLines: [],
            trendlines: [],
            rectangles: []
        }
    };

    function createChartCanvas(rootEl) {
        const c = document.createElement('canvas');
        c.id = 'rsc-chart';
        c.width = rootEl.clientWidth || window.innerWidth;
        c.height = Math.round((rootEl.clientHeight || window.innerHeight) * 0.55);
        c.style.display = 'block';
        c.style.width = '100%';
        c.style.height = '55vh';
        c.style.background = '#050915';
        c.style.borderBottom = '1px solid rgba(88,131,255,0.4)';
        const mount = rootEl.querySelector('.chart-host') || rootEl;
        mount.insertBefore(c, mount.firstChild);
        window.addEventListener('resize', () => {
            c.width = mount.clientWidth;
            c.height = Math.round(window.innerHeight * 0.55);
            draw();
        });
        return c;
    }

    
    // === Reserve event overlays & tooltips ===
    const reserveEvents = [];

    function addReserveEvent(evt) {
        // normalize payload
        const e = Object.assign({}, evt);
        reserveEvents.push(e);
        return e;
    }

    function seedDemoReserveEvents() {
        if (!state.candles.length) return;
        const n = state.candles.length;
        const baseReserve = 34500000;

        addReserveEvent({
            type: 'issuance',
            candleIndex: Math.round(n * 0.2),
            amount: 400000,
            nav: 1.0,
            reserve_before: baseReserve,
            reserve_after: baseReserve + 400000,
            composition_delta: { USDC: 0.02, USDT: -0.02 }
        });

        addReserveEvent({
            type: 'redemption',
            candleIndex: Math.round(n * 0.5),
            amount: 700000,
            nav: 1.0,
            reserve_before: baseReserve + 400000,
            reserve_after: baseReserve - 300000,
            composition_delta: { USDC: -0.03, DAI: 0.01 }
        });

        addReserveEvent({
            type: 'rebalance',
            candleIndex: Math.round(n * 0.75),
            amount: 2000000,
            nav: 1.0,
            reserve_before: baseReserve - 300000,
            reserve_after: baseReserve - 300000,
            composition_delta: { USDC: -0.06, USDT: 0.06 }
        });

        // expose simple API for future WS integration
        window.ReserveEvents = {
            add: addReserveEvent,
            all: function() { return reserveEvents.slice(); }
        };
    }

    function drawReserveEvents() {
        if (!reserveEvents.length || !state.candles.length) return;
        const w = canvas.width;
        const h = canvas.height;
        const candles = state.candles;
        const cw = w / candles.length;

        reserveEvents.forEach((e) => {
            if (typeof e.candleIndex !== 'number') return;
            const idx = Math.min(candles.length - 1, Math.max(0, e.candleIndex));
            const x = idx * cw + cw / 2;
            const y = h * 0.18; // keep markers in upper band of chart

            e._x = x;
            e._y = y;

            let color;
            switch (e.type) {
                case 'issuance': color = '#3F9BFF'; break;     // blue
                case 'redemption': color = '#E74B4B'; break;  // red
                case 'rebalance': color = '#E6C24A'; break;   // yellow
                default: color = '#22C55E'; break;            // green / other
            }

            const radius = 4;
            const outline = 'rgba(15,23,42,0.95)';

            ctx.save();
            ctx.lineWidth = 1;
            ctx.strokeStyle = outline;
            ctx.fillStyle = color;
            ctx.beginPath();
            ctx.arc(x, y, radius, 0, Math.PI * 2);
            ctx.fill();
            ctx.stroke();
            ctx.restore();
        });
    }

        // reserve overlays
        drawReserveEvents();
    }

    const reserveTooltip = document.getElementById('rc-reserve-tooltip');

    function formatReserveTooltip(e) {
        const delta = (e.reserve_after != null && e.reserve_before != null)
            ? (e.reserve_after - e.reserve_before)
            : (e.amount || 0);
        const dir = delta >= 0 ? '+' : '';
        const amountStr = (delta || 0).toLocaleString();

        let compLines = '—';
        if (e.composition_delta) {
            compLines = Object.entries(e.composition_delta)
                .map(([k, v]) => `${k}: ${v > 0 ? '+' : ''}${(v * 100).toFixed(1)}%`)
                .join('<br>');
        }

        let title;
        switch (e.type) {
            case 'issuance': title = 'Issuance Event'; break;
            case 'redemption': title = 'Redemption Event'; break;
            case 'rebalance': title = 'Rebalance Event'; break;
            default: title = 'Reserve Event'; break;
        }

        const nav = (e.nav != null && e.nav.toFixed) ? e.nav.toFixed(4) : (e.nav || '1.0000');

        return `
            <div class="rc-reserve-tip-title">${title}</div>
            <div class="rc-reserve-tip-line">Δ Reserve: ${dir}${amountStr} GRC</div>
            <div class="rc-reserve-tip-line">NAV: ${nav}</div>
            <div class="rc-reserve-tip-line">Before: ${e.reserve_before ? e.reserve_before.toLocaleString() : '—'} GRC</div>
            <div class="rc-reserve-tip-line">After: ${e.reserve_after ? e.reserve_after.toLocaleString() : '—'} GRC</div>
            <div class="rc-reserve-tip-line rc-reserve-tip-sub">Composition:</div>
            <div class="rc-reserve-tip-line rc-reserve-tip-sub">${compLines}</div>
        `;
    }

    function handleReserveHover(e) {
        if (!reserveTooltip || !reserveEvents.length) return;
        const rect = canvas.getBoundingClientRect();
        const mx = e.clientX - rect.left;
        const my = e.clientY - rect.top;

        let best = null;
        let bestDist = Infinity;
        reserveEvents.forEach((ev) => {
            if (typeof ev._x !== 'number' || typeof ev._y !== 'number') return;
            const dx = ev._x - mx;
            const dy = ev._y - my;
            const d = Math.sqrt(dx * dx + dy * dy);
            if (d < bestDist && d <= 14) {
                bestDist = d;
                best = ev;
            }
        });

        if (!best) {
            reserveTooltip.style.opacity = 0;
            return;
        }

        reserveTooltip.innerHTML = formatReserveTooltip(best);
        reserveTooltip.style.opacity = 1;
        reserveTooltip.style.transform = `translate(${e.clientX + 12}px, ${e.clientY + 12}px)`;
    }

function draw() {
        if (!ctx) return;
        const w = canvas.width;
        const h = canvas.height;
        ctx.clearRect(0, 0, w, h);

        // grid
        ctx.strokeStyle = '#151b2c';
        ctx.lineWidth = 1;
        const rows = 6;
        const cols = 12;
        for (let i = 1; i < rows; i++) {
            const y = (h / rows) * i;
            ctx.beginPath();
            ctx.moveTo(0, y);
            ctx.lineTo(w, y);
            ctx.stroke();
        }
        for (let j = 1; j < cols; j++) {
            const x = (w / cols) * j;
            ctx.beginPath();
            ctx.moveTo(x, 0);
            ctx.lineTo(x, h);
            ctx.stroke();
        }

        // simple candle rendering (placeholder, to be wired to devnet feed)
        const candles = state.candles;
        if (candles.length) {
            const max = Math.max(...candles.map(c => c.high));
            const min = Math.min(...candles.map(c => c.low));
            const range = max - min || 1;
            const cw = w / candles.length;

            candles.forEach((candle, i) => {
                const x = i * cw + cw / 2;
                const openY = h - ((candle.open - min) / range) * h;
                const closeY = h - ((candle.close - min) / range) * h;
                const highY = h - ((candle.high - min) / range) * h;
                const lowY = h - ((candle.low - min) / range) * h;
                const up = candle.close >= candle.open;
                ctx.strokeStyle = up ? '#3dffb8' : '#ff4d7a';
                ctx.fillStyle = up ? '#3dffb8' : '#ff4d7a';

                // wick
                ctx.beginPath();
                ctx.moveTo(x, highY);
                ctx.lineTo(x, lowY);
                ctx.stroke();

                // body
                const bodyTop = Math.min(openY, closeY);
                const bodyBottom = Math.max(openY, closeY);
                const bodyHeight = Math.max(2, bodyBottom - bodyTop);
                ctx.fillRect(x - cw * 0.25, bodyTop, cw * 0.5, bodyHeight);
            });
        }

        // overlays: entry / liquidation / breakeven lines
        drawHorizontalLines(state.overlays.entryLines, '#4da6ff');
        drawHorizontalLines(state.overlays.liquidationLines, '#ff4d7a');
        drawHorizontalLines(state.overlays.breakevenLines, '#ffd452');

        // trendlines
        ctx.strokeStyle = '#f6f6ff';
        ctx.lineWidth = 1.5;
        state.overlays.trendlines.forEach(tl => {
            ctx.beginPath();
            ctx.moveTo(tl.x1, tl.y1);
            ctx.lineTo(tl.x2, tl.y2);
            ctx.stroke();
        });

        // rectangles (zones)
        state.overlays.rectangles.forEach(r => {
            ctx.save();
            ctx.fillStyle = 'rgba(77, 255, 163, 0.07)';
            ctx.strokeStyle = 'rgba(77, 255, 163, 0.5)';
            ctx.lineWidth = 1;
            ctx.beginPath();
            ctx.rect(r.x, r.y, r.w, r.h);
            ctx.fill();
            ctx.stroke();
            ctx.restore();
        });
    }

    function drawHorizontalLines(lines, color) {
        if (!lines || !lines.length) return;
        const w = canvas.width;
        ctx.save();
        ctx.strokeStyle = color;
        ctx.lineWidth = 1;
        ctx.setLineDash([4, 4]);
        lines.forEach(line => {
            ctx.beginPath();
            ctx.moveTo(0, line.y);
            ctx.lineTo(w, line.y);
            ctx.stroke();
        });
        ctx.restore();
        ctx.setLineDash([]);
    }

    // Simple demo data so the chart isn't blank
    function seedDemoCandles() {
        const base = 100;
        let last = base;
        for (let i = 0; i < 80; i++) {
            const delta = (Math.random() - 0.5) * 4;
            const open = last;
            const close = last + delta;
            const high = Math.max(open, close) + Math.random() * 2;
            const low = Math.min(open, close) - Math.random() * 2;
            state.candles.push({ open, high, low, close });
            last = close;
        }
        const h = canvas.height;
        state.overlays.entryLines = [{ y: h * 0.4 }];
        state.overlays.liquidationLines = [{ y: h * 0.8 }];
        state.overlays.breakevenLines = [{ y: h * 0.5 }];
        seedDemoReserveEvents();
        draw();
    }

    // Basic mouse interaction for trendlines / rectangles hooks (placeholder)
    let drawingTrend = false;
    let trendStart = null;

    canvas.addEventListener('mousedown', (e) => {
        const rect = canvas.getBoundingClientRect();
        const x = e.clientX - rect.left;
        const y = e.clientY - rect.top;
        drawingTrend = true;
        trendStart = { x, y };
    });

    canvas.addEventListener('mouseup', (e) => {
        if (!drawingTrend || !trendStart) return;
        const rect = canvas.getBoundingClientRect();
        const x2 = e.clientX - rect.left;
        const y2 = e.clientY - rect.top;
        state.overlays.trendlines.push({ x1: trendStart.x, y1: trendStart.y, x2, y2 });
        drawingTrend = false;
        trendStart = null;
        draw();
    });


    canvas.addEventListener('mousemove', (e) => {
        handleReserveHover(e);
    });

    canvas.addEventListener('mouseleave', () => {
        if (reserveTooltip) {
            reserveTooltip.style.opacity = 0;
        }
    });



    // === WebSocket bootstrap (Hybrid topology: user terminal WS) ===
    const WS_ENDPOINT = (window.RC_WS_URL || (window.location.protocol === 'https:' ? 'wss://' : 'ws://') + window.location.host + '/ws/terminal');

    let ws = null;
    let wsConnected = false;

    const channelHandlers = {
        price: [],
        reserve: [],
        metrics: [],
        execution: [],
        margin: [],
        chain: []
    };

    function registerHandler(channel, fn) {
        if (!channelHandlers[channel]) channelHandlers[channel] = [];
        channelHandlers[channel].push(fn);
    }

    function routeMessage(msg) {
        if (!msg || !msg.channel) return;
        const handlers = channelHandlers[msg.channel] || [];
        handlers.forEach(fn => {
            try { fn(msg); } catch (e) { console.error('handler error', e); }
        });
    }

    function connectWS() {
        try {
            ws = new WebSocket(WS_ENDPOINT);
        } catch (e) {
            console.warn('WS connect failed', e);
            return;
        }

        ws.onopen = () => {
            wsConnected = true;
            console.log('Terminal WS connected.');
        };

        ws.onclose = () => {
            wsConnected = false;
            console.warn('Terminal WS closed, retrying…');
            setTimeout(connectWS, 3000);
        };

        ws.onerror = (err) => {
            console.error('Terminal WS error', err);
        };

        ws.onmessage = (ev) => {
            let data;
            try {
                data = JSON.parse(ev.data);
            } catch (e) {
                console.warn('Bad WS payload', ev.data);
                return;
            }
            routeMessage(data);
        };
    }

    // === Channel handlers ===

    // Reserve events → overlay markers
    registerHandler('reserve', (msg) => {
        if (!msg || !msg.event) return;
        // expected shape:
        // msg.event = { type, candleIndex or ts, amount, nav, reserve_before, reserve_after, composition_delta }
        addReserveEvent(msg.event);
        draw();
    });

    // Metrics → HUD + (optionally) boot tray in future
    registerHandler('metrics', (msg) => {
        if (!msg || !msg.metrics) return;
        const m = msg.metrics;
        if (window.RC_MetricsHUD && window.RC_MetricsHUD.set) {
            if (m.rtt_ms != null) window.RC_MetricsHUD.set('rtt', m.rtt_ms + 'ms');
            if (m.feed_ms != null) window.RC_MetricsHUD.set('feed_fresh', m.feed_ms + 'ms');
            if (m.ack_ms != null) window.RC_MetricsHUD.set('ack', m.ack_ms + 'ms');
            if (m.slot != null) window.RC_MetricsHUD.set('slot', String(m.slot));
            if (m.reserve_str != null) window.RC_MetricsHUD.set('reserve', m.reserve_str);
        }
    });

    // Price ticks → chart + MarginContext mark price
    registerHandler('price', (msg) => {
        if (!msg || !msg.tick) return;
        const t = msg.tick;
        if (!state.candles.length) return;
        const last = state.candles[state.candles.length - 1];
        if (t.price != null) {
            last.close = t.price;
            last.high = Math.max(last.high, t.price);
            last.low = Math.min(last.low, t.price);
            draw();
            if (window.MarginContext && typeof MarginContext.updateMark === 'function') {
                MarginContext.updateMark(t.price);
            }
        }
    });

    // Execution / fills, margin, chain hooks can be wired here later:
    registerHandler('execution', (msg) => {
        if (!msg || !msg.fill) return;
        if (window.MarginContext && typeof MarginContext.applyFill === 'function') {
            MarginContext.applyFill(msg.fill);
        }
    });

    registerHandler('margin', (msg) => {
        if (!msg || !msg.risk) return;
        const r = msg.risk;
        if (accountEquityEl) {
            accountEquityEl.textContent = formatGRC(r.equity);
        }
        if (todayRealizedEl && typeof r.equity === 'number') {
            // For DevNet, approximate "today realized" as equity delta vs base 100k.
            const base = 100000;
            const delta = r.equity - base;
            const sign = delta >= 0 ? '+' : '';
            todayRealizedEl.textContent = sign + delta.toFixed(2) + ' GRC';
            if (todayRealizedEl.classList) {
                todayRealizedEl.classList.remove('green', 'red');
                if (delta > 0) todayRealizedEl.classList.add('green');
                else if (delta < 0) todayRealizedEl.classList.add('red');
            }
        }
    });

    registerHandler('chain', (msg) => {
        // TODO: slot/height/NAV updates if needed
        // console.log('chain', msg);
    });

    connectWS();


    seedDemoCandles();
})();
