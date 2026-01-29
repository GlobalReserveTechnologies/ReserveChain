package config

import (
    "os"

    "gopkg.in/yaml.v3"
)

// WindowMode controls how settlement / issuance windows behave.
type WindowMode string

const (
    WindowFixed     WindowMode = "fixed"
    WindowTriggered WindowMode = "triggered"
    WindowBoth      WindowMode = "both"
)

// RPCSettings describes how the HTTP/WS API should bind.
type RPCSettings struct {
    HTTPListen string `yaml:"http_listen"`
    WSPath     string `yaml:"ws_path"`
}

// DBSettings controls persistence for the chain log and related state.
type DBSettings struct {
    Backend     string `yaml:"backend"`
    SQLitePath  string `yaml:"sqlite_path"`
    AutoGenesis bool   `yaml:"auto_genesis"`
}

// RuntimePaths are convenience pointers for bundled runtimes in DevNet.
type RuntimePaths struct {
    Go     string `yaml:"go"`
    PHP    string `yaml:"php"`
    SQLite string `yaml:"sqlite"`
}


// P2PSettings controls basic DevNet peer-to-peer behaviour.
// For early DevNet this maps onto simple HTTP-based peer sync and
// seed discovery rather than full gossip.
type P2PSettings struct {
    Port         int      `yaml:"port"`
    Mode         string   `yaml:"mode"`        // "seed" or "peer"
    SeedNodes    []string `yaml:"seed_nodes"`  // seed endpoints this node should contact
    MaxPeers     int      `yaml:"max_peers"`
    AllowExternal bool    `yaml:"allow_external"`
}

// NodeSettings configures the behaviour of a single DevNet node.
type NodeSettings struct {
    ID                  string        `yaml:"id"`
    HostPublicUI        bool          `yaml:"host_web_interface"`
    HostFinancePlatform bool          `yaml:"host_finance_platform"`
    FollowUpstreamURL   string        `yaml:"follow_upstream_url"`
    Peers               []string      `yaml:"peers"`

    RPC          RPCSettings   `yaml:"rpc"`
    DB           DBSettings    `yaml:"db"`
    RuntimePaths RuntimePaths  `yaml:"runtime_paths"`
}

// WindowSettings controls the behaviour of corridor settlement windows.
type WindowSettings struct {
    Mode               WindowMode `yaml:"mode"`
    FixedLengthSeconds int        `yaml:"fixed_length_seconds"`
    MinVolumeUSD       float64    `yaml:"min_volume_usd"`
    MinWindowSeconds   int        `yaml:"min_window_seconds"`
    MaxWindowSeconds   int        `yaml:"max_window_seconds"`
}

// FXSettings and YieldSettings control whether the DevNet simulator
// should generate synthetic FX / yield data. For now they are simple
// on/off flags.
type FXSettings struct {
    Simulate bool `yaml:"simulate"`
}

type YieldSettings struct {
    Simulate bool `yaml:"simulate"`
}

// WorkWeightsConfig mirrors the four-component work weighting used to
// compute node operator work scores (consensus, network, storage, service).
type WorkWeightsConfig struct {
    Consensus float64 `yaml:"consensus"`
    Network   float64 `yaml:"network"`
    Storage   float64 `yaml:"storage"`
    Service   float64 `yaml:"service"`
}

// RewardsSettings controls DevNet reward pool sizing and the relative
// weights applied when splitting the variable operator reward pool
// based on measured work.
type RewardsSettings struct {
    // VariablePoolGRC is the notional per-epoch pool (in GRC) reserved
    // for work-based operator rewards, before USD floor top-ups.
    VariablePoolGRC float64          `yaml:"variable_pool_grc"`
    WorkWeights     WorkWeightsConfig `yaml:"work_weights"`
}

// NodeConfig is the top-level configuration loaded from YAML.
type NodeConfig struct {
    Node     NodeSettings    `yaml:"node"`
    Windows  WindowSettings  `yaml:"windows"`
    FX       FXSettings      `yaml:"fx"`
    Yield    YieldSettings   `yaml:"yield"`
    Rewards  RewardsSettings `yaml:"rewards"`
    P2P      P2PSettings     `yaml:"p2p"`
}

// Load reads a YAML configuration file and unmarshals it into NodeConfig.
func Load(path string) (*NodeConfig, error) {
    cfg := &NodeConfig{}
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }
    if err := yaml.Unmarshal(data, cfg); err != nil {
        return nil, err
    }
    return cfg, nil
}
