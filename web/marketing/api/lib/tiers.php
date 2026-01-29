<?php
require_once __DIR__ . '/db.php';
require_once __DIR__ . '/earn.php';

const TIERS = [
    'core' => [
        'label' => 'Core Reserve',
        'margin' => 2,
        'pop_multiplier' => 1.0,
        'staking_multiplier' => 1.0,
        'haircut' => 0.80,
        'price_monthly_grc' => 0,
        'price_yearly_grc'  => 0,
    ],
    'elite' => [
        'label' => 'Elite Reserve',
        'margin' => 4,
        'pop_multiplier' => 1.2,
        'staking_multiplier' => 1.1,
        'haircut' => 0.80,
        'price_monthly_grc' => 350,
        'price_yearly_grc'  => 3150,
    ],
    'executive' => [
        'label' => 'Executive Reserve',
        'margin' => 7,
        'pop_multiplier' => 1.6,
        'staking_multiplier' => 1.5,
        'haircut' => 0.90,
        'price_monthly_grc' => 650,
        'price_yearly_grc'  => 5850,
    ],
    'express' => [
        'label' => 'Express Reserve',
        'margin' => 20,
        'pop_multiplier' => 3.0,
        'staking_multiplier' => 3.0,
        'haircut' => 0.95,
        'price_monthly_grc' => 850,
        'price_yearly_grc'  => 7650,
    ],
];

function rc_grc_to_usd($grc) {
    // Devnet stub: 1 GRC = 0.26 USD
    return round($grc * 0.26, 2);
}

function get_user_tier($userId) {
    $stmt = rc_db_query("SELECT * FROM user_tiers WHERE user_id = ?", [$userId]);
    $row = $stmt->fetch(PDO::FETCH_ASSOC);
    if (!$row) {
        // default core
        $now = (new DateTimeImmutable('now', new DateTimeZone('UTC')));
        $renew = $now->add(new DateInterval('P30D'));
        rc_db_exec("INSERT INTO user_tiers (user_id, tier, status, billing_cycle, renew_expires_at) VALUES (?,?,?,?,?)", [
            $userId,
            'core',
            'active',
            'monthly',
            $renew->format('Y-m-d H:i:s'),
        ]);
        return [
            'user_id' => $userId,
            'tier' => 'core',
            'status' => 'active',
            'billing_cycle' => 'monthly',
            'renew_expires_at' => $renew->format(DATE_ATOM),
            'grace_ends_at' => null,
        ];
    }
    $row['renew_expires_at'] = $row['renew_expires_at'] ? (new DateTimeImmutable($row['renew_expires_at'], new DateTimeZone('UTC')))->format(DATE_ATOM) : null;
    $row['grace_ends_at'] = $row['grace_ends_at'] ? (new DateTimeImmutable($row['grace_ends_at'], new DateTimeZone('UTC')))->format(DATE_ATOM) : null;
    return $row;
}

function get_user_tier_runtime($userId) {
    $stmt = rc_db_query("SELECT * FROM user_tier_runtime WHERE user_id = ?", [$userId]);
    $row = $stmt->fetch(PDO::FETCH_ASSOC);
    if ($row) return $row;

    $tier = get_user_tier($userId);
    $tconf = TIERS[$tier['tier']];
    rc_db_exec("INSERT INTO user_tier_runtime (user_id, collateral_mode, haircut_applied, margin_limit, pop_multiplier, staking_multiplier)
                VALUES (?,?,?,?,?,?)", [
        $userId,
        'protection',
        $tconf['haircut'],
        $tconf['margin'],
        $tconf['pop_multiplier'],
        $tconf['staking_multiplier'],
    ]);
    return [
        'user_id' => $userId,
        'collateral_mode' => 'protection',
        'haircut_applied' => $tconf['haircut'],
        'margin_limit' => $tconf['margin'],
        'pop_multiplier' => $tconf['pop_multiplier'],
        'staking_multiplier' => $tconf['staking_multiplier'],
    ];
}

function calculate_tier_pricing(array $tierRow, array $earnRow) {
    $tierName = $tierRow['tier'];
    $conf = TIERS[$tierName] ?? TIERS['core'];
    $cycle = $tierRow['billing_cycle'] ?? 'monthly';

    $baseGrc = $cycle === 'yearly' ? $conf['price_yearly_grc'] : $conf['price_monthly_grc'];
    $stakeDiscountGrc = 0; // hook in staking later
    $earnBalanceGrc = $earnRow['balance_grc'] ?? 0;

    // Apply earn credits but let surplus be computed by caller
    $applyEarn = min($baseGrc - $stakeDiscountGrc, $earnBalanceGrc);
    if ($applyEarn < 0) $applyEarn = 0;

    $payableGrc = max(0, $baseGrc - $stakeDiscountGrc - $applyEarn);

    $equivMonthlyGrc = $cycle === 'yearly' && $baseGrc > 0 ? $baseGrc / 12.0 : $baseGrc;

    return [
        'tier' => $tierName,
        'billing_cycle' => $cycle,
        'base_grc' => $baseGrc,
        'base_usd' => rc_grc_to_usd($baseGrc),
        'stake_discount_grc' => $stakeDiscountGrc,
        'stake_discount_usd' => rc_grc_to_usd($stakeDiscountGrc),
        'earn_balance_grc' => $earnBalanceGrc,
        'earn_balance_usd' => rc_grc_to_usd($earnBalanceGrc),
        'earn_applied_grc' => $applyEarn,
        'earn_applied_usd' => rc_grc_to_usd($applyEarn),
        'estimated_payable_grc' => $payableGrc,
        'estimated_payable_usd' => rc_grc_to_usd($payableGrc),
        'equivalent_monthly_grc' => $equivMonthlyGrc,
        'equivalent_monthly_usd' => rc_grc_to_usd($equivMonthlyGrc),
    ];
}
