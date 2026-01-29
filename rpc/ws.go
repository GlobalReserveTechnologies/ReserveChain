package rpc

import (
    "log"
    "net"

    "reservechain/config"
    "reservechain/core"
)

type WSServer struct {
    cfg    *config.Config
    engine *core.Engine
}

func NewWSServer(cfg *config.Config, e *core.Engine) *WSServer {
    return &WSServer{cfg, e}
}

func (s *WSServer) Start() error {
    ln, err := net.Listen("tcp", s.cfg.RPC.WS)
    if err != nil {
        return err
    }
    log.Printf("WS placeholder listening on %s", s.cfg.RPC.WS)
    for {
        _, _ = ln.Accept()
    }
}
