<?php
require_once __DIR__ . '/lib/db.php';
require_once __DIR__ . '/lib/auth.php';
require_once __DIR__ . '/lib/tiers.php';
require_once __DIR__ . '/lib/earn.php';

$user = rc_require_user();
$tier = get_user_tier($user->id);
$runtime = get_user_tier_runtime($user->id);
$earn = get_user_earn_balance($user->id);
$pricing = calculate_tier_pricing($tier, $earn);

rc_json_response([
    'success' => true,
    'tier' => $tier,
    'pricing' => $pricing,
    'runtime' => $runtime,
]);
