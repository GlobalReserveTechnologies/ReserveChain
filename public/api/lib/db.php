<?php
// Shared DB helper for ReserveChain PHP APIs (DevNet, SQLite app.db)

function rc_db() {
    static $pdo = null;
    if ($pdo === null) {
        $root = realpath(__DIR__ . '/../../..');
        $dbPath = $root . DIRECTORY_SEPARATOR . 'database' . DIRECTORY_SEPARATOR . 'app.db';
        $dsn = 'sqlite:' . $dbPath;
        $pdo = new PDO($dsn);
        $pdo->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);
    }
    return $pdo;
}

function rc_db_query($sql, $params = []) {
    $stmt = rc_db()->prepare($sql);
    $stmt->execute($params);
    return $stmt;
}

function rc_db_exec($sql, $params = []) {
    $stmt = rc_db()->prepare($sql);
    return $stmt->execute($params);
}

function rc_json_response($payload, $code = 200) {
    http_response_code($code);
    header('Content-Type: application/json');
    echo json_encode($payload);
    exit;
}
