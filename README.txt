# ReserveChain Base Website (Template Integration)

This folder contains the ReserveChain base marketing / explorer website,
integrated with your Go devnet node via a PHP RPC proxy.

Structure:

- web/public/index.php      -> Landing page (ReserveChain overview)
- web/public/protocol.php   -> Protocol overview
- web/public/vault.php      -> Vault & privacy layer
- web/public/explorer.php   -> Explorer stub (height & supply wired to RPC)
- web/public/docs.php       -> Documentation stub
- web/public/css/site.css   -> Styles
- web/public/js/site.js     -> Scroll animations + RPC wiring
- web/public/api/config.php -> Node RPC URL configuration
- web/public/api/rpc.php    -> PHP -> Go JSON-RPC proxy

To run the site locally (PHP dev server):

    cd web/public
    php -S 127.0.0.1:8080

Make sure your Go node is running and exposing HTTP RPC on 127.0.0.1:8545/rpc.
