// ReserveChain Browser Wallet Keystore (v2)
// - Generates a local ECDSA P-256 keypair using WebCrypto
// - Encrypts private key material into a downloadable JSON keystore file
// - Stores only encrypted keystore blobs in localStorage (no plaintext keys)

window.ReserveWallet = (function() {
  const STORAGE_KEY = 'reservechain_keystore_v2';

  function loadState() {
    try {
      const raw = window.localStorage.getItem(STORAGE_KEY);
      if (!raw) return { wallets: [] };
      return JSON.parse(raw);
    } catch (e) {
      console.error('Failed to load keystore', e);
      return { wallets: [] };
    }
  }

  function saveState(state) {
    try {
      window.localStorage.setItem(STORAGE_KEY, JSON.stringify(state));
    } catch (e) {
      console.error('Failed to save keystore', e);
    }
  }

  function makeId() {
    return 'w_' + Math.random().toString(36).slice(2) + Date.now().toString(36);
  }

  function bufToB64(buf) {
    const bytes = new Uint8Array(buf);
    let bin = '';
    for (let i = 0; i < bytes.length; i++) bin += String.fromCharCode(bytes[i]);
    return btoa(bin);
  }

  function b64ToBuf(b64) {
    const bin = atob(b64);
    const bytes = new Uint8Array(bin.length);
    for (let i = 0; i < bin.length; i++) bytes[i] = bin.charCodeAt(i);
    return bytes.buffer;
  }

  async function sha256Base32(input) {
    const enc = new TextEncoder().encode(input);
    const hash = await crypto.subtle.digest('SHA-256', enc);
    // base32-ish (no padding) using base64url as a simple readable address for now
    const b64 = bufToB64(hash).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/g, '');
    return 'rc1' + b64.slice(0, 38).toLowerCase();
  }

  async function deriveKeyPBKDF2(password, saltBytes, iterations) {
    const enc = new TextEncoder();
    const keyMaterial = await crypto.subtle.importKey(
      'raw',
      enc.encode(password),
      { name: 'PBKDF2' },
      false,
      ['deriveKey']
    );

    return crypto.subtle.deriveKey(
      {
        name: 'PBKDF2',
        salt: saltBytes,
        iterations: iterations,
        hash: 'SHA-256'
      },
      keyMaterial,
      { name: 'AES-GCM', length: 256 },
      false,
      ['encrypt', 'decrypt']
    );
  }

  async function encryptJsonWithPassword(obj, password) {
    const salt = crypto.getRandomValues(new Uint8Array(16));
    const iv = crypto.getRandomValues(new Uint8Array(12));
    const iterations = 200000;

    const key = await deriveKeyPBKDF2(password, salt, iterations);

    const plaintext = new TextEncoder().encode(JSON.stringify(obj));
    const ciphertext = await crypto.subtle.encrypt(
      { name: 'AES-GCM', iv },
      key,
      plaintext
    );

    return {
      kdf: 'pbkdf2',
      hash: 'sha256',
      iterations,
      salt: bufToB64(salt.buffer),
      cipher: 'aes-256-gcm',
      iv: bufToB64(iv.buffer),
      ciphertext: bufToB64(ciphertext)
    };
  }

  async function decryptJsonWithPassword(keystoreCrypto, password) {
    const salt = new Uint8Array(b64ToBuf(keystoreCrypto.salt));
    const iv = new Uint8Array(b64ToBuf(keystoreCrypto.iv));
    const ciphertext = b64ToBuf(keystoreCrypto.ciphertext);

    const key = await deriveKeyPBKDF2(password, salt, keystoreCrypto.iterations || 200000);

    const plaintext = await crypto.subtle.decrypt(
      { name: 'AES-GCM', iv },
      key,
      ciphertext
    );

    const decoded = new TextDecoder().decode(plaintext);
    return JSON.parse(decoded);
  }

  async function generateWallet(label, password) {
    if (!crypto || !crypto.subtle) throw new Error('WebCrypto not available in this browser');
    if (!password || password.length < 8) throw new Error('Password must be at least 8 characters');

    const id = makeId();
    const keyPair = await crypto.subtle.generateKey(
      { name: 'ECDSA', namedCurve: 'P-256' },
      true,
      ['sign', 'verify']
    );

    const pubJwk = await crypto.subtle.exportKey('jwk', keyPair.publicKey);
    const privJwk = await crypto.subtle.exportKey('jwk', keyPair.privateKey);

    const address = await sha256Base32(pubJwk.x + '.' + pubJwk.y);

    const walletCore = {
      id,
      label: label || 'Wallet ' + id.slice(-6),
      created_at: new Date().toISOString(),
      alg: 'ECDSA_P256',
      address,
      pub: pubJwk
    };

    const encrypted = await encryptJsonWithPassword({ priv: privJwk }, password);

    const keystoreFile = {
      version: 1,
      type: 'reservechain-keystore',
      wallet: walletCore,
      crypto: encrypted
    };

    // Store only safe, non-secret metadata + encrypted blob
    const state = loadState();
    state.wallets = state.wallets || [];
    state.wallets.push({
      id: walletCore.id,
      label: walletCore.label,
      address: walletCore.address,
      created_at: walletCore.created_at,
      alg: walletCore.alg,
      pub: walletCore.pub,
      crypto: encrypted
    });
    saveState(state);

    return { wallet: walletCore, keystore: keystoreFile };
  }

  function listWallets() {
    const state = loadState();
    return state.wallets || [];
  }

  function getWalletMeta(id) {
    const state = loadState();
    return (state.wallets || []).find(w => w.id === id) || null;
  }

  async function signMessage(walletId, password, message) {
    const meta = getWalletMeta(walletId);
    if (!meta) throw new Error('Wallet not found');

    // Decrypt private key JWK
    const decoded = await decryptJsonWithPassword(meta.crypto, password);
    const privJwk = decoded.priv;

    const privateKey = await crypto.subtle.importKey(
      'jwk',
      privJwk,
      { name: 'ECDSA', namedCurve: 'P-256' },
      false,
      ['sign']
    );

    const data = new TextEncoder().encode(String(message));
    const sig = await crypto.subtle.sign(
      { name: 'ECDSA', hash: 'SHA-256' },
      privateKey,
      data
    );

    return {
      wallet_id: walletId,
      address: meta.address,
      signature_b64: bufToB64(sig)
    };
  }

  return {
    generateWallet,
    listWallets,
    getWalletMeta,
    signMessage
  };
})();
