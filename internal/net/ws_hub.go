package net

import (
    "encoding/json"
    "log"
    "net/http"

    "github.com/gorilla/websocket"
)

type WSHub struct {
    clients    map[*websocket.Conn]bool
    broadcast  chan Event
    register   chan *websocket.Conn
    unregister chan *websocket.Conn
    upgrader   websocket.Upgrader
}

func NewWSHub() *WSHub {
    return &WSHub{
        clients:    make(map[*websocket.Conn]bool),
        broadcast:  make(chan Event, 128),
        register:   make(chan *websocket.Conn),
        unregister: make(chan *websocket.Conn),
        upgrader: websocket.Upgrader{
            CheckOrigin: func(r *http.Request) bool { return true },
        },
    }
}

func (h *WSHub) Run() {
    for {
        select {
        case conn := <-h.register:
            h.clients[conn] = true
        case conn := <-h.unregister:
            if _, ok := h.clients[conn]; ok {
                delete(h.clients, conn)
                conn.Close()
            }
        case ev := <-h.broadcast:
            data, err := json.Marshal(ev)
            if err != nil {
                log.Println("marshal event:", err)
                continue
            }
            for conn := range h.clients {
                if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
                    log.Println("write ws:", err)
                    h.unregister <- conn
                }
            }
        }
    }
}

func (h *WSHub) Broadcast(ev Event) {
    h.broadcast <- ev
}

func (h *WSHub) HandleWS(w http.ResponseWriter, r *http.Request) {
    conn, err := h.upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Println("upgrade:", err)
        return
    }
    h.register <- conn
}
