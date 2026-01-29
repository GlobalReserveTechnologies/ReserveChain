package rpc

import (
    "encoding/json"
    "net/http"

    "reservechain/config"
    "reservechain/core"
    "reservechain/state"
)

type HTTPServer struct {
    cfg    *config.Config
    engine *core.Engine
    state  *state.State
}

func NewHTTPServer(cfg *config.Config, e *core.Engine, s *state.State) *HTTPServer {
    return &HTTPServer{cfg, e, s}
}

func (s *HTTPServer) Start() error {
    http.HandleFunc("/rpc", s.handle)
    return http.ListenAndServe(s.cfg.RPC.HTTP, nil)
}

type rpcRequest struct {
    Method string            `json:"method"`
    Params []json.RawMessage `json:"params"`
}

type rpcResponse struct {
    Result interface{} `json:"result,omitempty"`
    Error  string      `json:"error,omitempty"`
}

func (s *HTTPServer) handle(w http.ResponseWriter, r *http.Request) {
    defer r.Body.Close()

    var req rpcRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        json.NewEncoder(w).Encode(rpcResponse{Error: "invalid_request"})
        return
    }

    switch req.Method {
    case "getHeight":
        json.NewEncoder(w).Encode(rpcResponse{Result: s.state.Height})

    case "getBalance":
        if len(req.Params) < 1 {
            json.NewEncoder(w).Encode(rpcResponse{Error: "missing_address"})
            return
        }
        var addr string
        if err := json.Unmarshal(req.Params[0], &addr); err != nil {
            json.NewEncoder(w).Encode(rpcResponse{Error: "invalid_address"})
            return
        }
        json.NewEncoder(w).Encode(rpcResponse{Result: s.state.GetBalance(addr)})

    case "getSupply":
        json.NewEncoder(w).Encode(rpcResponse{Result: s.state.GetTotalSupply()})

    case "submitTx":
        if len(req.Params) < 1 {
            json.NewEncoder(w).Encode(rpcResponse{Error: "missing_tx"})
            return
        }
        var tx core.Tx
        if err := json.Unmarshal(req.Params[0], &tx); err != nil {
            json.NewEncoder(w).Encode(rpcResponse{Error: "invalid_tx"})
            return
        }
        if err := s.engine.SubmitTx(tx); err != nil {
            json.NewEncoder(w).Encode(rpcResponse{Error: err.Error()})
            return
        }
        json.NewEncoder(w).Encode(rpcResponse{Result: "ok"})

    case "faucet":
        if len(req.Params) < 2 {
            json.NewEncoder(w).Encode(rpcResponse{Error: "missing_params"})
            return
        }
        var addr string
        var amt uint64
        if err := json.Unmarshal(req.Params[0], &addr); err != nil {
            json.NewEncoder(w).Encode(rpcResponse{Error: "invalid_address"})
            return
        }
        if err := json.Unmarshal(req.Params[1], &amt); err != nil {
            json.NewEncoder(w).Encode(rpcResponse{Error: "invalid_amount"})
            return
        }
        s.engine.Faucet(addr, amt)
        json.NewEncoder(w).Encode(rpcResponse{Result: "ok"})

    case "getRecentBlocks":
        limit := 10
        if len(req.Params) >= 1 {
            _ = json.Unmarshal(req.Params[0], &limit)
        }
        blocks := s.engine.GetRecentBlocks(limit)
        json.NewEncoder(w).Encode(rpcResponse{Result: blocks})

    default:
        json.NewEncoder(w).Encode(rpcResponse{Error: "unknown_method"})
    }
}
