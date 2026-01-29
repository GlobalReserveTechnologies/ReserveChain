<?php
declare(strict_types=1);

/**
 * Simple proxy for the DevNet Go treasury analytics endpoint.
 *
 * This lets the workstation call /api/treasury.php from the same origin
 * without running into CORS issues when the Go engine is on :8080.
 */

header('Content-Type: application/json');

if ($_SERVER['REQUEST_METHOD'] !== 'GET') {
    http_response_code(405);
    echo json_encode(['error' => 'GET only']);
    exit;
}

// Allow override via env if you ever want to point at a remote DevNet.
$base = getenv('RESERVECHAIN_DEVNET_BASE');
if (!$base) {
    $base = 'http://127.0.0.1:8080';
}

$url = rtrim($base, '/') . '/api/analytics/treasury';

$ctx = stream_context_create([
    'http' => [
        'method'  => 'GET',
        'timeout' => 2,
        'header'  => "Accept: application/json\r\n",
    ],
]);

$raw = @file_get_contents($url, false, $ctx);
if ($raw === false) {
    http_response_code(502);
    echo json_encode(['error' => 'Failed to reach DevNet treasury endpoint']);
    exit;
}

$decoded = json_decode($raw, true);
if (!is_array($decoded)) {
    http_response_code(502);
    echo json_encode(['error' => 'Invalid response from DevNet treasury endpoint']);
    exit;
}

echo json_encode($decoded);
