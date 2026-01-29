# WalletConnect Configuration

This project uses WalletConnect v2 for external wallet support.

The WalletConnect Project ID is stored in:

    apps/workstation_portal/.env

Environment variable:

    VITE_WC_PROJECT_ID

This value is public and safe to include in builds.
If you change the Project ID, rebuild the workstation portal:

    ./build.sh
