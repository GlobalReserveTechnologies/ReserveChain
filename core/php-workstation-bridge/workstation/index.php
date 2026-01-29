<?php
// ReserveChain Workstation entrypoint
// This file is a friendly alias for the workstation UI.
// During development you can hit:
//   http://localhost:8080/workstation/
//
// Internally this delegates to the legacy public/workstation/index.php
// so existing behavior and paths remain intact.
require __DIR__ . '/../public/workstation/index.php';
