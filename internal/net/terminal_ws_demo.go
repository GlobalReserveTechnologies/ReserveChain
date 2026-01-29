
package net

import (
    "encoding/json"
    "log"
    "math"
    "math/rand"
    "net/http"
    "sync"
    "time"

    "github.com/gorilla/websocket"
)

// TerminalMessage is the envelope sent to the trading terminal WS client.
// It matches the { channel, metrics, tick, event, fill, risk, chain } structure the JS expects.
type TerminalMessage struct {
    Channel string            `json:"channel"`
    Metrics *TerminalMetrics  `json:"metrics,omitempty"`
    Tick    *TerminalTick     `json:"tick,omitempty"`
    Event   *ReserveEventMsg  `json:"event,omitempty"`
    Fill    *ExecutionFill    `json:"fill,omitempty"`
    Risk    *MarginSnapshot   `json:"risk,omitempty"`
    Chain   *ChainSnapshot    `json:"chain,omitempty"`
}

// TerminalMetrics drives the Metrics HUD.
type TerminalMetrics struct {
    RTTMs      int64  `json:"rtt_ms,omitempty"`
    FeedMs     int64  `json:"feed_ms,omitempty"`
    AckMs      int64  `json:"ack_ms,omitempty"`
    Slot       uint64 `json:"slot,omitempty"`
    ReserveStr string `json:"reserve_str,omitempty"`
}

// TerminalTick drives the last-price behavior on the dev-chart.
type TerminalTick struct {
    Price float64 `json:"price"`
}

// ReserveEventMsg maps to the reserve overlay markers.
type ReserveEventMsg struct {
    Type            string             `json:"type"`
    CandleIndex     int                `json:"candleIndex,omitempty"`
    Amount          float64            `json:"amount,omitempty"`
    NAV             float64            `json:"nav,omitempty"`
    ReserveBefore   float64            `json:"reserve_before,omitempty"`
    ReserveAfter    float64            `json:"reserve_after,omitempty"`
    CompositionDelta map[string]float64 `json:"composition_delta,omitempty"`
}

// ExecutionFill is a synthetic fill event for DevNet.
type ExecutionFill struct {
    OrderID string  `json:"order_id"`
    Symbol  string  `json:"symbol"`
    Side    string  `json:"side"`
    Qty     float64 `json:"qty"`
    Price   float64 `json:"price"`
    Fee     float64 `json:"fee"`
    TS      int64   `json:"ts"`
}

// MarginSnapshot is a simple risk summary for future use.
type MarginSnapshot struct {
    Equity          float64 `json:"equity"`
    MarginUsed      float64 `json:"margin_used"`
    MarginFree      float64 `json:"margin_free"`
    ExposureNotional float64 `json:"exposure_notional"`
    Leverage        float64 `json:"leverage"`
}

// ChainSnapshot ties trading into chain telemetry.
type ChainSnapshot struct {
    Slot   uint64  `json:"slot"`
    Height uint64  `json:"height"`
    Epoch  uint64  `json:"epoch"`
    NAV    float64 `json:"nav"`
}

var (
    terminalUpgrader = websocket.Upgrader{
        CheckOrigin: func(r *http.Request) bool { return true },
    }

    terminalConnsMu sync.Mutex
    terminalConns   = make(map[*websocket.Conn]struct{})
)

func init() {
    // Register the terminal WS handler on the default mux.
    http.HandleFunc("/ws/terminal", HandleTerminalWS)
    // Seed RNG for demo streams.
    rand.Seed(time.Now().UnixNano())
    // Start a background demo loop so the terminal feels alive even before
    // the real engine wiring is finished.
    go startTerminalDemoLoop()
}

// HandleTerminalWS upgrades the HTTP connection to a WebSocket and
// tracks the terminal client. For now this is a pure broadcast-only
// channel; we ignore incoming messages.
func HandleTerminalWS(w http.ResponseWriter, r *http.Request) {
    conn, err := terminalUpgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Println("terminal ws upgrade:", err)
        return
    }

    terminalConnsMu.Lock()
    terminalConns[conn] = struct{}{}
    terminalConnsMu.Unlock()

    go func(c *websocket.Conn) {
        defer func() {
            terminalConnsMu.Lock()
            delete(terminalConns, c)
            terminalConnsMu.Unlock()
            _ = c.Close()
        }()

        for {
            // Drain any incoming messages; protocol is one-way for now.
            if _, _, err := c.ReadMessage(); err != nil {
                return
            }
        }
    }(conn)
}

// startTerminalDemoLoop emits synthetic metrics, price ticks, reserve events,
// execution fills, margin snapshots, and chain snapshots to any connected
// terminal clients. This is purely for DevNet and can later be replaced
// by real engine-driven streams.
func startTerminalDemoLoop() {
    basePrice := 1.0000
    reserve := 34_500_000.0
    slot := uint64(41)
    height := uint64(1000)
    epoch := uint64(1)

    // Very simple internal position model to keep fills/margin coherent.
    var posSide int // +1 long, -1 short, 0 flat
    var posQty float64
    const baseEquity = 100000.0
    totalFees := 0.0

    metricsTicker := time.NewTicker(1 * time.Second)
    priceTicker := time.NewTicker(250 * time.Millisecond)
    reserveTicker := time.NewTicker(12 * time.Second)
    execTicker := time.NewTicker(7 * time.Second)

    defer metricsTicker.Stop()
    defer priceTicker.Stop()
    defer reserveTicker.Stop()
    defer execTicker.Stop()

    for {
        select {
        case <-metricsTicker.C:
            slot++
            if slot%50 == 0 {
                height++
            }
            if slot%200 == 0 {
                epoch++
            }

            nav := 1.0 + (rand.Float64()-0.5)*0.003

            m := TerminalMessage{
                Channel: "metrics",
                Metrics: &TerminalMetrics{
                    RTTMs:      40 + int64(rand.Intn(12)),
                    FeedMs:     15 + int64(rand.Intn(10)),
                    AckMs:      40 + int64(rand.Intn(20)),
                    Slot:       slot,
                    ReserveStr: formatReserveString(reserve),
                },
            }
            broadcastTerminalMessage(m)

            cmsg := TerminalMessage{
                Channel: "chain",
                Chain: &ChainSnapshot{
                    Slot:   slot,
                    Height: height,
                    Epoch:  epoch,
                    NAV:    nav,
                },
            }
            broadcastTerminalMessage(cmsg)

            // Also emit a margin snapshot derived from our internal position.
            equity, marginUsed, marginFree, exposureNotional, lev := computeMarginSnapshot(baseEquity, basePrice, posSide, posQty, totalFees)
            risk := &MarginSnapshot{
                Equity:          equity,
                MarginUsed:      marginUsed,
                MarginFree:      marginFree,
                ExposureNotional: exposureNotional,
                Leverage:        lev,
            }
            rmsg := TerminalMessage{
                Channel: "margin",
                Risk:    risk,
            }
            broadcastTerminalMessage(rmsg)

        case <-priceTicker.C:
            delta := (rand.Float64() - 0.5) * 0.0008
            basePrice += delta
            if basePrice <= 0 {
                basePrice = 1.0000
            }
            pmsg := TerminalMessage{
                Channel: "price",
                Tick: &TerminalTick{
                    Price: basePrice,
                },
            }
            broadcastTerminalMessage(pmsg)

        case <-reserveTicker.C:
            eventType := []string{"issuance", "redemption", "rebalance"}[rand.Intn(3)]
            candleIdx := 20 + rand.Intn(80)
            amount := 400_000.0 + rand.Float64()*300_000.0
            before := reserve

            after := reserve
            compDelta := map[string]float64{
                "USDC": 0.0,
                "USDT": 0.0,
                "DAI":  0.0,
            }

            switch eventType {
            case "issuance":
                after = reserve + amount
                compDelta["USDC"] = 0.02
                compDelta["USDT"] = -0.02
            case "redemption":
                after = reserve - amount
                compDelta["USDC"] = -0.03
                compDelta["DAI"] = 0.01
            case "rebalance":
                after = reserve
                compDelta["USDC"] = -0.06
                compDelta["USDT"] = 0.06
            }

            reserve = after

            ev := &ReserveEventMsg{
                Type:            eventType,
                CandleIndex:     candleIdx,
                Amount:          amount,
                NAV:             1.0000,
                ReserveBefore:   before,
                ReserveAfter:    after,
                CompositionDelta: compDelta,
            }

            rmsg2 := TerminalMessage{
                Channel: "reserve",
                Event:   ev,
            }
            broadcastTerminalMessage(rmsg2)

        case <-execTicker.C:
            // Generate a synthetic fill consistent with current price.
            side := "buy"
            dir := 1
            if rand.Intn(2) == 0 {
                side = "sell"
                dir = -1
            }
            qty := 10.0 + rand.Float64()*15.0
            price := basePrice + (rand.Float64()-0.5)*0.002
            fee := math.Abs(qty*price) * 0.0002
            ts := time.Now().Unix()

            // Update internal position state
            if posSide == 0 {
                posSide = dir
                posQty = qty
            } else if posSide == dir {
                posQty += qty
            } else {
                if qty < posQty {
                    posQty -= qty
                } else if qty == posQty {
                    posSide = 0
                    posQty = 0
                } else {
                    leftover := qty - posQty
                    posSide = dir
                    posQty = leftover
                }
            }
            totalFees += fee

            fill := &ExecutionFill{
                OrderID: randomID(),
                Symbol:  "GRC-USD",
                Side:    side,
                Qty:     qty,
                Price:   price,
                Fee:     fee,
                TS:      ts,
            }
            emsg := TerminalMessage{
                Channel: "execution",
                Fill:    fill,
            }
            broadcastTerminalMessage(emsg)
        }
    }
}

func broadcastTerminalMessage(msg TerminalMessage) {
    payload, err := json.Marshal(msg)
    if err != nil {
        log.Println("terminal ws marshal:", err)
        return
    }

    terminalConnsMu.Lock()
    defer terminalConnsMu.Unlock()

    for c := range terminalConns {
        if err := c.WriteMessage(websocket.TextMessage, payload); err != nil {
            log.Println("terminal ws write:", err)
            _ = c.Close()
            delete(terminalConns, c)
        }
    }
}

func computeMarginSnapshot(baseEquity, price float64, posSide int, posQty, totalFees float64) (equity, marginUsed, marginFree, exposureNotional, leverage float64) {
    exposureNotional = 0
    unreal := 0.0

    if posSide != 0 && posQty > 0 && price > 0 {
        exposureNotional = posQty * price
        entry := price // for demo; we are not tracking a separate entry here
        if posSide > 0 {
            unreal = (price - entry) * posQty
        } else {
            unreal = (entry - price) * posQty
        }
    }

    equity = baseEquity + unreal - totalFees
    if exposureNotional > 0 && equity > 0 {
        leverage = exposureNotional / equity
    } else {
        leverage = 0
    }

    marginUsed = exposureNotional / 5.0
    marginFree = equity - marginUsed
    return
}

func formatReserveString(v float64) string {
    millions := v / 1_000_000.0
    return sprintfFloat(millions) + "M GRC"
}

// sprintfFloat formats a float with 1 decimal place without pulling fmt in.
func sprintfFloat(x float64) string {
    neg := x < 0
    if neg {
        x = -x
    }
    scaled := int64(x*10 + 0.5)
    intPart := scaled / 10
    fracPart := scaled % 10

    buf := make([]byte, 0, 24)
    if neg {
        buf = append(buf, '-')
    }
    buf = appendInt(buf, intPart)
    buf = append(buf, '.')
    buf = append(buf, byte('0'+fracPart))
    return string(buf)
}

func appendInt(buf []byte, v int64) []byte {
    if v == 0 {
        return append(buf, '0')
    }
    var tmp [20]byte
    i := len(tmp)
    for v > 0 {
        i--
        tmp[i] = byte('0' + v%10)
        v /= 10
    }
    return append(buf, tmp[i:]...)
}

func randomID() string {
    const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
    b := make([]byte, 10)
    for i := range b {
        b[i] = letters[rand.Intn(len(letters))]
    }
    return string(b)
}
