<?php
require_once __DIR__ . '/lib/db.php';
require_once __DIR__ . '/lib/auth.php';
require_once __DIR__ . '/lib/tiers.php';
require_once __DIR__ . '/lib/earn.php';

$user = rc_require_user();
$input = json_decode(file_get_contents('php://input'), true);

$targetTier = $input['target_tier'] ?? null;
$billingCycle = $input['billing_cycle'] ?? 'yearly';
$paymentSource = $input['payment_source'] ?? 'vault';
$chainTx = $input['chain_tx'] ?? null;

if (!$targetTier || !isset(TIERS[$targetTier])) {
    rc_json_response(['success' => false, 'error' => 'invalid_tier'], 400);
}
if (!$chainTx || ($chainTx['type'] ?? '') !== 'TX_TIER_RENEW') {
    rc_json_response(['success' => false, 'error' => 'invalid_tx'], 400);
}

// Call Go devnet HTTP API
$nodeUrl = 'http://127.0.0.1:8080/api/tier/renew';

$ch = curl_init($nodeUrl);
curl_setopt_array($ch, [
    CURLOPT_POST => true,
    CURLOPT_RETURNTRANSFER => true,
    CURLOPT_HTTPHEADER => ['Content-Type: application/json'],
    CURLOPT_POSTFIELDS => json_encode($chainTx),
]);
$raw = curl_exec($ch);
if ($raw === false) {
    rc_json_response(['success' => false, 'error' => 'chain_unreachable'], 500);
}
$res = json_decode($raw, true);
if (empty($res['success'])) {
    rc_json_response(['success' => false, 'error' => 'chain_rejected_tx', 'chain_error' => $res], 400);
}
$chainHash = $res['tx_hash'] ?? null;

$currentTier = get_user_tier($user->id);
$earn = get_user_earn_balance($user->id);

$txBody = $chainTx['tx'];
$earnApplied = (float)($txBody['earn_applied_grc'] ?? 0);
$surplusGrc = (float)($txBody['surplus_to_time_grc'] ?? 0);
$stakeDisc = (float)($txBody['stake_discount_grc'] ?? 0);
$payAmount = (float)($txBody['payment']['amount_grc'] ?? 0);

// Update DB
rc_db()->beginTransaction();

rc_db_exec("UPDATE earn_credits SET balance_grc = balance_grc - ? WHERE user_id = ?", [
    $earnApplied + $surplusGrc,
    $user->id,
]);

$now = new DateTimeImmutable('now', new DateTimeZone('UTC'));
$renew = clone $now;
if ($billingCycle === 'yearly') {
    $renew = $renew->add(new DateInterval('P1Y'));
} elseif ($billingCycle === 'monthly') {
    $renew = $renew->add(new DateInterval('P1M'));
}

// Surplus -> extra days
$conf = TIERS[$targetTier];
$baseGrc = $billingCycle === 'yearly' ? $conf['price_yearly_grc'] : $conf['price_monthly_grc'];
$extraDays = 0;
if ($baseGrc > 0 && $surplusGrc > 0) {
    $extraMonths = $surplusGrc / $baseGrc;
    $extraDays = (int)round($extraMonths * 30.0);
    if ($extraDays > 0) {
        $renew = $renew->add(new DateInterval('P' . $extraDays . 'D'));
    }
}

// Upsert runtime row
rc_db_exec("INSERT INTO user_tier_runtime (user_id, collateral_mode, haircut_applied, margin_limit, pop_multiplier, staking_multiplier)
            VALUES (?,?,?,?,?,?)
            ON CONFLICT(user_id) DO UPDATE SET
                haircut_applied = excluded.haircut_applied,
                margin_limit = excluded.margin_limit,
                pop_multiplier = excluded.pop_multiplier,
                staking_multiplier = excluded.staking_multiplier", [
    $user->id,
    'protection',
    $conf['haircut'],
    $conf['margin'],
    $conf['pop_multiplier'],
    $conf['staking_multiplier'],
]);

// Upsert tier row
rc_db_exec("INSERT INTO user_tiers (user_id, tier, status, billing_cycle, current_renewal_tx, renew_expires_at, grace_ends_at)
            VALUES (?,?,?,?,?,?,NULL)
            ON CONFLICT(user_id) DO UPDATE SET
                tier = excluded.tier,
                status = 'active',
                billing_cycle = excluded.billing_cycle,
                current_renewal_tx = excluded.current_renewal_tx,
                renew_expires_at = excluded.renew_expires_at,
                grace_ends_at = NULL,
                updated_at = CURRENT_TIMESTAMP", [
    $user->id,
    $targetTier,
    'active',
    $billingCycle,
    $chainHash,
    $renew->format('Y-m-d H:i:s'),
]);

// Record payment
rc_db_exec("INSERT INTO tier_payments (user_id, tier_before, tier_after, billing_cycle,
            amount_grc, earn_applied_grc, stake_discount_grc, surplus_to_time_grc, tx_hash)
            VALUES (?,?,?,?,?,?,?,?,?)", [
    $user->id,
    $currentTier['tier'],
    $targetTier,
    $billingCycle,
    $payAmount,
    $earnApplied,
    $stakeDisc,
    $surplusGrc,
    $chainHash,
]);

rc_db()->commit();

rc_json_response([
    'success' => true,
    'tx_hash' => $chainHash,
    'new_tier' => $targetTier,
    'extra_days' => $extraDays,
]);
