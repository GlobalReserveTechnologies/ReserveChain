
// Workstation trading links for popup windows
document.addEventListener('DOMContentLoaded', function () {
  const openTerminal = document.getElementById('btn-open-terminal-popup');
  const openMulti    = document.getElementById('btn-open-terminal-multi');
  if (openTerminal) {
    openTerminal.addEventListener('click', function () {
      window.open('/trading-terminal/trading_terminal.php', 'tradingTerminal', 'width=1400,height=820,resizable=yes');
    });
  }
  if (openMulti) {
    openMulti.addEventListener('click', function () {
      window.open('/trading-terminal/multipanel.php', 'tradingTerminalMulti', 'width=1600,height=900,resizable=yes');
    });
  }
});
