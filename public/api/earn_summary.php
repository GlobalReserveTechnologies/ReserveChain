<?php
require_once __DIR__ . '/lib/db.php';
require_once __DIR__ . '/lib/auth.php';
require_once __DIR__ . '/lib/tiers.php';
require_once __DIR__ . '/lib/earn.php';

$user = rc_require_user();
$earn = get_user_earn_balance($user->id);
$breakdown = get_earn_breakdown($user->id);

rc_json_response([
    'success' => true,
    'balance_grc' => (float)$earn['balance_grc'],
    'balance_usd' => rc_grc_to_usd((float)$earn['balance_grc']),
    'breakdown' => $breakdown,
]);
