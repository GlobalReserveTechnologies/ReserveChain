<?php
require_once __DIR__ . '/lib/db.php';
require_once __DIR__ . '/lib/auth.php';
require_once __DIR__ . '/lib/tiers.php';
require_once __DIR__ . '/lib/earn.php';

$user = rc_require_user();
$raw = file_get_contents('php://input');
$input = json_decode($raw, true);

$targetTier = $input['target_tier'] ?? null;
$billingCycle = $input['billing_cycle'] ?? 'yearly';
$paymentSource = $input['payment_source'] ?? 'vault';

if (!$targetTier || !isset(TIERS[$targetTier])) {
    rc_json_response(['success' => false, 'error' => 'invalid_tier'], 400);
}

$currentTier = get_user_tier($user->id);
$earn = get_user_earn_balance($user->id);

$tmp = $currentTier;
$tmp['tier'] = $targetTier;
$tmp['billing_cycle'] = $billingCycle;
$pricing = calculate_tier_pricing($tmp, $earn);

$totalPayableGrc = $pricing['estimated_payable_grc'];
$quoteId = bin2hex(random_bytes(8));

rc_json_response([
    'success' => true,
    'quote' => [
        'quote_id' => $quoteId,
        'user_id' => $user->id,
        'target_tier' => $targetTier,
        'billing_cycle' => $billingCycle,
        'payment_source' => $paymentSource,
        'base_cost_grc' => $pricing['base_grc'],
        'stake_discount_grc' => $pricing['stake_discount_grc'],
        'earn_applied_grc' => $pricing['earn_applied_grc'],
        'surplus_earn_grc' => max(0, $earn['balance_grc'] - $pricing['earn_applied_grc']),
        'total_payable_grc' => $totalPayableGrc,
        'base_cost_usd' => $pricing['base_usd'],
        'total_payable_usd' => $pricing['estimated_payable_usd'],
    ],
]);
