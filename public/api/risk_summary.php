<?php
// Risk summary endpoint for DevNet.
// For now this is a lightweight aggregator that prefers live data from the Go node
// (chain head + mempool) and falls back to safe defaults if the node is offline.

header('Content-Type: application/json');

$resp = [
  'network_risk_level'        => 'GREEN',
  'global_leverage_factor'    => 1.0,
  'corridor_utilization_pct'  => 0.0,
  'vault_exposure_total'      => 0.0,
  'insurance_fund'            => 0.0,
  'stabilizer_mode'           => 'NORMAL',
  'chain_height'              => null,
  'chain_supply'              => null,
  'mempool_len'               => null,
];

$nodeBase = 'http://127.0.0.1:8080';

// Helper: fetch JSON with short timeout, swallow errors.
function rc_fetch_json_or_null($url) {
  $opts = [
    'http' => [
      'method'  => 'GET',
      'timeout' => 0.8,
    ]
  ];
  $ctx = stream_context_create($opts);
  $raw = @file_get_contents($url, false, $ctx);
  if ($raw === false) {
    return null;
  }
  $data = json_decode($raw, true);
  if (!is_array($data)) {
    return null;
  }
  return $data;
}

// Try to get chain head stats.
$head = rc_fetch_json_or_null($nodeBase . '/api/chain/head');
if (is_array($head)) {
  if (isset($head['height'])) {
    $resp['chain_height'] = (int)$head['height'];
  }
  if (isset($head['supply'])) {
    $resp['chain_supply'] = (float)$head['supply'];
  }
}

// Try to get mempool snapshot.
$mp = rc_fetch_json_or_null($nodeBase . '/api/chain/mempool');
if (is_array($mp)) {
  // mempoolHandler currently returns a JSON array of tx summaries.
  $resp['mempool_len'] = count($mp);
  // Bump risk a bit if mempool is very full in devnet.
  if ($resp['mempool_len'] > 50) {
    $resp['network_risk_level'] = 'AMBER';
  }
  if ($resp['mempool_len'] > 200) {
    $resp['network_risk_level'] = 'RED';
  }
}

echo json_encode($resp);
