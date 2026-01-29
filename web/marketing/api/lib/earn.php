<?php
require_once __DIR__ . '/db.php';

function get_user_earn_balance($userId) {
    $stmt = rc_db_query("SELECT * FROM earn_credits WHERE user_id = ?", [$userId]);
    $row = $stmt->fetch(PDO::FETCH_ASSOC);
    if ($row) return $row;
    rc_db_exec("INSERT INTO earn_credits (user_id, balance_grc) VALUES (?,0)", [$userId]);
    return ['user_id' => $userId, 'balance_grc' => 0.0];
}

function add_earn_credit($userId, $amountGrc, $sourceType, $tradingEpoch = null, $treasuryEpoch = null, $stakingEpoch = null) {
    rc_db_exec("INSERT INTO earn_ledger (user_id, source_type, amount_grc, trading_epoch, treasury_epoch, staking_epoch)
                VALUES (?,?,?,?,?,?)", [
        $userId,
        $sourceType,
        $amountGrc,
        $tradingEpoch,
        $treasuryEpoch,
        $stakingEpoch,
    ]);
    rc_db_exec("INSERT INTO earn_credits (user_id, balance_grc)
                VALUES (?, ?)
                ON CONFLICT(user_id)
                DO UPDATE SET balance_grc = balance_grc + excluded.balance_grc,
                              last_update_at = CURRENT_TIMESTAMP", [
        $userId,
        $amountGrc,
    ]);
}

function get_earn_breakdown($userId) {
    $stmt = rc_db_query("SELECT source_type, SUM(amount_grc) as total FROM earn_ledger WHERE user_id = ? GROUP BY source_type", [$userId]);
    $breakdown = ['capital_grc' => 0, 'flow_grc' => 0, 'risk_grc' => 0, 'bonus_grc' => 0];
    while ($row = $stmt->fetch(PDO::FETCH_ASSOC)) {
        $key = $row['source_type'] . '_grc';
        if (!array_key_exists($key, $breakdown)) continue;
        $breakdown[$key] = (float)$row['total'];
    }
    return $breakdown;
}
