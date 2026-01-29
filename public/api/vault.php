<?php
// Simple file-backed vault + stealth API for the workstation prototype.

header('Content-Type: application/json');

$method = $_SERVER['REQUEST_METHOD'] ?? 'GET';
$action = $_GET['action'] ?? ($_POST['action'] ?? null);

$storageFile = __DIR__ . '/../../config/vault_state.json';

function load_state($file) {
    if (!file_exists($file)) {
        $now = date('c');
        $state = [
            'next_vault_id'   => 2,
            'next_stealth_id' => 1,
            'vaults' => [
                [
                    'vault_id' => 'v1',
                    'label'    => 'Primary Vault',
                    'type'     => 'single',
                    'visibility_mode'    => 'A',
                    'pnl_settlement_mode'=> 'source',
                    'yield_policy' => [
                        'sources' => ['external', 'internal'],
                        'duration_tier' => 'medium',
                    ],
                    'balance'  => ['GRC' => 0.0],
                    'created_at' => $now
                ]
            ],
            'stealth' => []
        ];
        return $state;
    }
    $raw = file_get_contents($file);
    if ($raw === false || $raw === '') {
        return [
            'next_vault_id'   => 1,
            'next_stealth_id' => 1,
            'vaults' => [],
            'stealth' => []
        ];
    }
    $decoded = json_decode($raw, true);

    // Backfill missing fields for older state versions
    if (!isset($decoded['vaults']) || !is_array($decoded['vaults'])) {
        $decoded['vaults'] = [];
    }
    foreach ($decoded['vaults'] as &$v) {
        if (!isset($v['template'])) {
            $v['template'] = 'custom';
        }
        if (!isset($v['visibility_mode'])) {
            $v['visibility_mode'] = 'A';
        }
        if (!isset($v['pnl_settlement_mode'])) {
            $v['pnl_settlement_mode'] = 'source';
        }
        if (!isset($v['yield_policy']) || !is_array($v['yield_policy'])) {
            $v['yield_policy'] = [
                'sources' => ['external', 'internal'],
                'duration_tier' => 'medium',
            ];
        } else {
            if (!isset($v['yield_policy']['sources']) || !is_array($v['yield_policy']['sources'])) {
                $v['yield_policy']['sources'] = ['external', 'internal'];
            }
            if (!isset($v['yield_policy']['duration_tier'])) {
                $v['yield_policy']['duration_tier'] = 'medium';
            }
        }
    }
    unset($v);

    if (!is_array($decoded)) {
        return [
            'next_vault_id'   => 1,
            'next_stealth_id' => 1,
            'vaults' => [],
            'stealth' => []
        ];
    }
    if (!isset($decoded['vaults'])) {
        $decoded['vaults'] = [];
    }
    if (!isset($decoded['stealth'])) {
        $decoded['stealth'] = [];
    }
    if (!isset($decoded['next_vault_id'])) {
        $decoded['next_vault_id'] = count($decoded['vaults']) + 1;
    }
    if (!isset($decoded['next_stealth_id'])) {
        $decoded['next_stealth_id'] = count($decoded['stealth']) + 1;
    }
    return $decoded;
}

function save_state($file, $state) {
    $dir = dirname($file);
    if (!is_dir($dir)) {
        @mkdir($dir, 0777, true);
    }
    $json = json_encode($state, JSON_PRETTY_PRINT);
    file_put_contents($file, $json);
}

function error_out($msg, $code = 400, $errCode = null, $details = []) {
    http_response_code($code);
    echo json_encode([
        'error'   => $msg,
        'code'    => $errCode ?: ('HTTP_' . $code),
        'details' => $details,
    ]);
    exit;
}

$state = load_state($storageFile);

if ($action === null) {
    error_out('Missing action', 400);
}

switch ($action) {
    case 'list':
        // Return vaults plus a NAV snapshot so UIs can display client-facing
        // balances in USD-equivalent terms while the underlying ledger remains GRC.
        $vaults = $state['vaults'];

        $nav = 1.0;
        try {
            $valJson = @file_get_contents('http://127.0.0.1:8080/api/valuation/latest');
            if ($valJson !== false) {
                $val = json_decode($valJson, true);
                if (is_array($val) && isset($val['nav']) && is_numeric($val['nav'])) {
                    $nav = (float)$val['nav'];
                }
            }
        } catch (Throwable $e) {
            error_log('vault list: failed to load NAV, defaulting to 1.0: ' . $e->getMessage());
        }

        foreach ($vaults as &$v) {
            $balGrc = 0.0;
            if (isset($v['balance']) && is_array($v['balance']) && isset($v['balance']['GRC'])) {
                $balGrc = (float)$v['balance']['GRC'];
            }
            $v['balance_usd'] = $balGrc * $nav;
        }
        unset($v);

        echo json_encode([
            'vaults' => $vaults,
            'nav'    => $nav,
        ]);
        break;

    case 'create':
        if ($method !== 'POST') error_out('POST required', 405);
        $body = json_decode(file_get_contents('php://input'), true) ?? [];
        $label = trim($body['label'] ?? '');
        if ($label === '') $label = 'Vault ' . $state['next_vault_id'];

        $type = $body['type'] ?? 'single';
        $visibility = $body['visibility_mode'] ?? 'A';
        $pnl = $body['pnl_settlement_mode'] ?? 'source';
        $template = $body['template'] ?? 'custom';

        $vid = 'v' . $state['next_vault_id'];
        $state['next_vault_id']++;

                $vault = [
            'vault_id' => $vid,
            'label'    => $label,
            'type'     => $type,
            'template' => $template,
            'visibility_mode'     => $visibility,
            'pnl_settlement_mode' => $pnl,
            'balance'  => ['GRC' => 0.0],
            'created_at' => date('c')
        ];
        $state['vaults'][] = $vault;
        save_state($storageFile, $state);

        // Also record this vault creation on-chain via the devnet node.
        $tx = [
            'type' => 'TX_VAULT_CREATE',
            'tx'   => [
                'vault_id'        => $vid,
                'owner'           => 'local-user', // placeholder; wire real owner later
                'label'           => $label,
                'type'            => $type,
                'threshold'       => ($type === 'multi') ? 2 : 1,
                'signers'         => [],
                'visibility_mode' => $visibility,
                'duration_tier'   => 'medium',
                'timestamp'       => time(),
            ],
        ];
        // Fire-and-forget; if the node is down this will be logged but not fatal.
        try {
            $ctx = stream_context_create([
                'http' => [
                    'method'  => 'POST',
                    'header'  => "Content-Type: application/json\r\n",
                    'content' => json_encode($tx),
                    'timeout' => 0.5,
                ]
            ]);
            @file_get_contents('http://127.0.0.1:8080/api/tx/vault_create', false, $ctx);
        } catch (Throwable $e) {
            error_log('vault_create on-chain failed: ' . $e->getMessage());
        }

        echo json_encode($vault);
        break;

    case 'update_settings':
        if ($method !== 'POST') error_out('POST required', 405);
        $body = json_decode(file_get_contents('php://input'), true) ?? [];
        $vid = $body['vault_id'] ?? null;
        if (!$vid) error_out('vault_id required');

        $updated = null;
        foreach ($state['vaults'] as &$v) {
            if ($v['vault_id'] === $vid) {
                if (isset($body['visibility_mode'])) {
                    $v['visibility_mode'] = $body['visibility_mode'];
                }
                if (isset($body['pnl_settlement_mode'])) {
                    $v['pnl_settlement_mode'] = $body['pnl_settlement_mode'];
                }
                if (isset($body['duration_tier'])) {
                    $allowedDurations = ['short', 'medium', 'long'];
                    if (in_array($body['duration_tier'], $allowedDurations, true)) {
                        if (!isset($v['yield_policy']) || !is_array($v['yield_policy'])) {
                            $v['yield_policy'] = [
                                'sources' => ['external', 'internal'],
                                'duration_tier' => $body['duration_tier'],
                            ];
                        } else {
                            $v['yield_policy']['duration_tier'] = $body['duration_tier'];
                            if (!isset($v['yield_policy']['sources']) || !is_array($v['yield_policy']['sources'])) {
                                $v['yield_policy']['sources'] = ['external', 'internal'];
                            }
                        }
                    }
                }
                if (isset($body['label']) && trim($body['label']) !== '') {
                    $v['label'] = trim($body['label']);
                }
                $updated = $v;
                break;
            }
        }
        unset($v);
        if ($updated === null) error_out('Vault not found', 404);
        save_state($storageFile, $state);
        echo json_encode($updated);
        break;

    case 'transfer_internal':
        if ($method !== 'POST') error_out('POST required', 405);
        $body = json_decode(file_get_contents('php://input'), true) ?? [];
        $fromId = $body['from_vault_id'] ?? null;
        $toId   = $body['to_vault_id'] ?? null;
        $amount = floatval($body['amount'] ?? 0);
        if (!$fromId || !$toId) error_out('from_vault_id and to_vault_id required');
        if ($fromId === $toId) error_out('from and to cannot be same');
        if ($amount <= 0) error_out('amount must be > 0');

        $fromIdx = null;
        $toIdx   = null;
        foreach ($state['vaults'] as $idx => $v) {
            if ($v['vault_id'] === $fromId) $fromIdx = $idx;
            if ($v['vault_id'] === $toId)   $toIdx   = $idx;
        }
        if ($fromIdx === null || $toIdx === null) error_out('Vault not found', 404);

        $fromBal = floatval($state['vaults'][$fromIdx]['balance']['GRC'] ?? 0);
        if ($fromBal < $amount) error_out('Insufficient balance', 400);

        $state['vaults'][$fromIdx]['balance']['GRC'] = $fromBal - $amount;
        $state['vaults'][$toIdx]['balance']['GRC']   = floatval($state['vaults'][$toIdx]['balance']['GRC'] ?? 0) + $amount;

        save_state($storageFile, $state);
        echo json_encode(['ok' => true]);
        break;

    case 'stealth_list':
        $vid = $_GET['vault_id'] ?? null;
        if (!$vid) error_out('vault_id required');
        $rows = [];
        foreach ($state['stealth'] as $s) {
            if ($s['vault_id'] === $vid) $rows[] = $s;
        }
        echo json_encode($rows);
        break;

    case 'stealth_create':
        if ($method !== 'POST') error_out('POST required', 405);
        $body = json_decode(file_get_contents('php://input'), true) ?? [];
        $vid = $body['vault_id'] ?? null;
        if (!$vid) error_out('vault_id required');

        $found = false;
        foreach ($state['vaults'] as $v) {
            if ($v['vault_id'] === $vid) { $found = true; break; }
        }
        if (!$found) error_out('Vault not found', 404);

        $sid = 's' . $state['next_stealth_id'];
        $state['next_stealth_id']++;

        $row = [
            'id'              => $sid,
            'vault_id'        => $vid,
            'label'           => $body['label'] ?? null,
            'stealth_address' => $body['stealth_address'] ?? null,
            'ephemeral_pubkey'=> $body['ephemeral_pubkey'] ?? null,
            'is_active'       => true,
            'created_at'      => date('c'),
            'last_used_at'    => null
        ];
        $state['stealth'][] = $row;
        save_state($storageFile, $state);
        echo json_encode($row);
        break;


    case 'apply_pnl':
        if ($method !== 'POST') error_out('POST required', 405, 'METHOD_NOT_ALLOWED');
        $body = json_decode(file_get_contents('php://input'), true) ?? [];
        $vid  = $body['vault_id'] ?? null;
        $pnl  = isset($body['pnl']) ? (float)$body['pnl'] : null;
        $mode = $body['mode'] ?? 'source';

        if (!$vid) {
            error_out('vault_id required', 400, 'VALIDATION');
        }
        if ($pnl === null || !is_numeric($body['pnl'])) {
            error_out('pnl must be numeric', 400, 'VALIDATION');
        }

        $targetIndex = null;
        foreach ($state['vaults'] as $idx => $v) {
            if ($v['vault_id'] === $vid) {
                $targetIndex = $idx;
                break;
            }
        }
        if ($targetIndex === null) {
            error_out('Vault not found', 404, 'VAULT_NOT_FOUND');
        }

        if (!isset($state['vaults'][$targetIndex]['balance']['GRC'])) {
            $state['vaults'][$targetIndex]['balance']['GRC'] = 0.0;
        }
        $current = (float)$state['vaults'][$targetIndex]['balance']['GRC'];
        $newBal  = $current + (float)$pnl;

        if ($newBal < 0) {
            $details = ['requested' => $pnl, 'previous_balance' => $current, 'clamped_to' => 0.0];
            $newBal = 0.0;
        } else {
            $details = ['requested' => $pnl, 'previous_balance' => $current, 'new_balance' => $newBal];
        }

        $state['vaults'][$targetIndex]['balance']['GRC'] = $newBal;
        save_state($storageFile, $state);

        echo json_encode([
            'ok' => true,
            'vault_id' => $vid,
            'mode' => $mode,
            'balance' => $newBal,
            'details' => $details,
        ]);
        break;

    default:
        error_out('Unknown action: ' . $action, 400, 'UNKNOWN_ACTION');
}
