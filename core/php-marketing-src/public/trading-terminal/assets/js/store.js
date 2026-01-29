
const Store = {
    state: {
        profile: 'stable',
        fx: null,
        nav: null,
        yield: null,
        windows: null,
        lastTickId: null,
        balances: null,
        address: 'demo-user',
        leaderId: null,
        tickLatencyMs: null,
        connection: 'disconnected',
        fxHistory: [],
        navHistory: [],
        lastTickTimestamp: null,
        windowHistory: [],
        eventLog: [] // Mint/Redeem/Window events for Explorer
    },
    listeners: [],
    setProfile(profile) {
        this.state.profile = profile;
        fetch('/api/profile/set', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ profile })
        }).catch(console.error);
        this.notify();
    },
    updateWindowFromTick(snapshot) {
        if (!snapshot) return;
        this.state.windows = snapshot;
        const existing = this.state.windowHistory.find(w => w.window_id === snapshot.window_id);
        if (!existing) {
            this.state.windowHistory.push(snapshot);
            if (this.state.windowHistory.length > 20) {
                this.state.windowHistory.shift();
            }
        } else {
            Object.assign(existing, snapshot);
        }
    },
    appendEvent(ev) {
        // ev = full WS event
        const entry = {
            id: ev.id || '',
            type: ev.type,
            ts: ev.timestamp || null,
            payload: ev.payload || {}
        };
        this.state.eventLog.unshift(entry);
        if (this.state.eventLog.length > 100) {
            this.state.eventLog.pop();
        }
        this.notify();
    },
    updateFromValuationTick(payload) {
        const now = Date.now();
        if (this.state.lastTickTimestamp) {
            this.state.tickLatencyMs = now - this.state.lastTickTimestamp;
        }
        this.state.lastTickTimestamp = now;

        this.state.fx = payload.fx;
        this.state.nav = payload.nav;
        this.state.yield = payload.yield;
        this.state.leaderId = payload.leader_id;
        this.state.lastTickId = payload.tick_id;
        this.state.profile = payload.profile || this.state.profile;
        this.updateWindowFromTick(payload.windows);

        if (payload.fx && typeof payload.fx.eur_usd === 'number') {
            this.state.fxHistory.push({ t: now, eur_usd: payload.fx.eur_usd });
        }
        if (payload.nav && typeof payload.nav.grc === 'number') {
            this.state.navHistory.push({ t: now, grc: payload.nav.grc });
        }
        if (this.state.fxHistory.length > 120) this.state.fxHistory.shift();
        if (this.state.navHistory.length > 120) this.state.navHistory.shift();

        this.notify();
    },
    updateBalances(balances) {
        this.state.balances = balances;
        this.notify();
    },
    setConnectionStatus(status) {
        this.state.connection = status;
        this.notify();
    },
    subscribe(fn) {
        this.listeners.push(fn);
        fn(this.state);
    },
    notify() {
        this.listeners.forEach(fn => fn(this.state));
    }
};

async function fetchBalances() {
    try {
        const res = await fetch('/api/balances?address=' + encodeURIComponent(Store.state.address));
        if (!res.ok) return;
        const data = await res.json();
        Store.updateBalances(data);
    } catch (e) {
        console.error("balances error", e);
    }
}

function connectWebSocket() {
    const ws = new WebSocket("ws://127.0.0.1:8080/ws");
    Store.setConnectionStatus('connecting');

    ws.onopen = () => {
        Store.setConnectionStatus('connected');
        fetchBalances();
        fetch('/api/profile/get')
            .then(r => r.json())
            .then(d => {
                if (d.profile) {
                    Store.state.profile = d.profile;
                    Store.notify();
                }
            }).catch(console.error);
    };

    ws.onmessage = (msg) => {
        try {
            const ev = JSON.parse(msg.data);
            if (ev.type === "ValuationTick") {
                Store.updateFromValuationTick(ev.payload);
            } else if (ev.type === "Mint" || ev.type === "Redeem" || ev.type === "Transfer") {
                Store.appendEvent(ev);
                fetchBalances();
            } else if (ev.type === "ProfileChange") {
                if (ev.payload.profile) {
                    Store.state.profile = ev.payload.profile;
                    Store.notify();
                }
            } else if (ev.type === "WindowUpdate") {
                Store.appendEvent(ev);
            }
        } catch (e) {
            console.error("WS parse error", e);
        }
    };

    ws.onclose = () => {
        Store.setConnectionStatus('disconnected');
        setTimeout(connectWebSocket, 2000);
    };
}

// Live data feed is disabled by default for now.
// To enable WebSocket-driven updates, call connectWebSocket() from your app bootstrap.

// Expose helpers for other UIs (workstation launcher / terminal popup)
window.ReserveConnectWS = connectWebSocket;
