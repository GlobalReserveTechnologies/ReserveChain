
// session_manager.js
// Session tokens for the trading terminal (H2).

(function () {
  const STORAGE_KEY = 'rc_terminal_session_v1';
  const DEFAULT_TTL_MS = 4 * 60 * 60 * 1000; // 4 hours

  function nowMs() {
    return Date.now();
  }

  function randomHex(len) {
    const bytes = new Uint8Array(len);
    crypto.getRandomValues(bytes);
    return Array.from(bytes)
      .map((b) => b.toString(16).padStart(2, '0'))
      .join('');
  }

  function loadRaw() {
    try {
      const raw = localStorage.getItem(STORAGE_KEY);
      if (!raw) return null;
      return JSON.parse(raw);
    } catch (e) {
      console.error('SessionManager: failed to parse session', e);
      return null;
    }
  }

  function saveRaw(obj) {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(obj));
  }

  const SessionManager = {
    get() {
      return loadRaw();
    },

    clear() {
      localStorage.removeItem(STORAGE_KEY);
    },

    isValid(session) {
      if (!session) return false;
      if (!session.token || !session.scopes || !session.expires_at) return false;
      if (nowMs() > session.expires_at) return false;
      return true;
    },

    /**
     * Create a new trading session.
     *
     * options:
     *  - scopes: ['trade'] etc.
     *  - collateral_source: 'hotwallet' | 'vault'
     *  - vault_id: string | null
     */
    async create(options) {
      const scopes = options.scopes || ['trade'];
      const collateralSource = options.collateral_source || 'hotwallet';
      const vaultId = options.vault_id || null;

      if (!window.ReserveWallet || !ReserveWallet.isUnlocked || !ReserveWallet.isUnlocked()) {
        throw new Error('Wallet must be unlocked to create a session');
      }

      const walletAddress =
        (ReserveWallet.getRootAddress && ReserveWallet.getRootAddress()) || 'unknown';
      const walletPubkey =
        (ReserveWallet.getPublicKeyBase64 && ReserveWallet.getPublicKeyBase64()) || null;

      const createdAt = nowMs();
      const expiresAt = createdAt + DEFAULT_TTL_MS;
      const nonce = randomHex(16);

      const payload = {
        type: 'terminal_session',
        scopes,
        collateral_source: collateralSource,
        vault_id: vaultId,
        wallet_address: walletAddress,
        wallet_pubkey: walletPubkey,
        created_at: createdAt,
        expires_at: expiresAt,
        nonce,
      };

      let token = null;
      let signature = null;

      if (ReserveWallet.signPayload) {
        const signed = await ReserveWallet.signPayload(payload);
        token = nonce + ':' + signed.signature;
        signature = signed.signature;
      } else {
        token = 'dev-' + randomHex(32);
        signature = null;
      }

      const session = {
        token,
        signature,
        payload,
        scopes,
        collateral_source: collateralSource,
        vault_id: vaultId,
        wallet_address: walletAddress,
        wallet_pubkey: walletPubkey,
        created_at: createdAt,
        expires_at: expiresAt,
      };

      saveRaw(session);
      return session;
    },
  };

  window.SessionManager = SessionManager;
})();
