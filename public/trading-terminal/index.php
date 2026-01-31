<?php
header('Location: /trading-terminal/trading_terminal.php', true, 302);
?>
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>ReserveChain Trading Terminal</title>
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <link rel="stylesheet" href="/trading-terminal/assets/css/terminal.css">
</head>
<body class="terminal-window">
  <div class="terminal-chrome">
    <div class="terminal-bar">
      <div class="terminal-title">ReserveChain Trading Terminal</div>
    </div>
    <div class="terminal-body">
      <div class="terminal-wrapper">
        <div class="top-bar">
          <div class="top-left">
            <div class="logo-wrap">
              <div class="logo"></div>
              <div class="logo-ticker">GRC</div>
            </div>
            <div>
              <div class="top-title">Redirecting to the Trading Terminal…</div>
              <div class="top-subtitle">If you are not redirected, use the link below.</div>
            </div>
          </div>
        </div>
        <div class="chart-wrapper">
          <div class="chart-main" style="display:flex;align-items:center;justify-content:center;text-align:center;">
            <div>
              <p>Continue to the live terminal:</p>
              <p><a href="/trading-terminal/trading_terminal.php">Open Trading Terminal</a></p>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</body>
</html>
