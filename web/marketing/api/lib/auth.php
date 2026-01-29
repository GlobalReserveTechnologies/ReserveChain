<?php
require_once __DIR__ . '/db.php';

// For devnet we just ensure there is a demo user and return it.
function rc_require_user() {
    $stmt = rc_db_query("SELECT * FROM users WHERE id = 1");
    $user = $stmt->fetch(PDO::FETCH_OBJ);
    if (!$user) {
        rc_db_exec("INSERT INTO users (external_id, email, display_name, status) VALUES (?,?,?,?)", [
            'devnet-demo',
            'demo@reservechain.local',
            'DevNet Demo',
            'active',
        ]);
        $id = rc_db()->lastInsertId();
        $stmt = rc_db_query("SELECT * FROM users WHERE id = ?", [$id]);
        $user = $stmt->fetch(PDO::FETCH_OBJ);
    }
    return $user;
}
