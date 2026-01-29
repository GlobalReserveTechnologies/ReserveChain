import React, { useEffect, useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";

type KeystoreFile = {
  version: number;
  type: string;
  wallet: {
    id: string;
    label: string;
    created_at: string;
    alg: string;
    address: string;
    pub: {
      kty: string;
      crv: string;
      x: string;
      y: string;
      [k: string]: any;
    };
  };
  crypto: {
    kdf: string;
    hash: string;
    iterations: number;
    salt: string;
    cipher: string;
    iv: string;
    ciphertext: string;
  };
};

function b64ToBuf(b64: string): ArrayBuffer {
  const bin = atob(b64);
  const bytes = new Uint8Array(bin.length);
  for (let i = 0; i < bin.length; i++) bytes[i] = bin.charCodeAt(i);
  return bytes.buffer;
}

function bufToB64(buf: ArrayBuffer): string {
  const bytes = new Uint8Array(buf);
  let bin = "";
  for (let i = 0; i < bytes.length; i++) bin += String.fromCharCode(bytes[i]);
  return btoa(bin);
}

async function deriveKeyPBKDF2(password: string, salt: Uint8Array, iterations: number) {
  const baseKey = await crypto.subtle.importKey(
    "raw",
    new TextEncoder().encode(password),
    "PBKDF2",
    false,
    ["deriveKey"]
  );

  return crypto.subtle.deriveKey(
    {
      name: "PBKDF2",
      salt,
      iterations,
      hash: "SHA-256",
    },
    baseKey,
    { name: "AES-GCM", length: 256 },
    false,
    ["decrypt"]
  );
}

async function decryptJsonWithPassword(cryptoBlob: KeystoreFile["crypto"], password: string) {
  const salt = new Uint8Array(b64ToBuf(cryptoBlob.salt));
  const iv = new Uint8Array(b64ToBuf(cryptoBlob.iv));
  const ciphertext = b64ToBuf(cryptoBlob.ciphertext);

  const key = await deriveKeyPBKDF2(password, salt, cryptoBlob.iterations || 200000);

  const plaintext = await crypto.subtle.decrypt({ name: "AES-GCM", iv }, key, ciphertext);
  const decoded = new TextDecoder().decode(plaintext);
  return JSON.parse(decoded);
}

async function signChallengeWithKeystore(ks: KeystoreFile, password: string, challenge: string) {
  const decoded = await decryptJsonWithPassword(ks.crypto, password);
  const privJwk = decoded.priv;

  const privateKey = await crypto.subtle.importKey(
    "jwk",
    privJwk,
    { name: "ECDSA", namedCurve: "P-256" },
    false,
    ["sign"]
  );

  const data = new TextEncoder().encode(String(challenge));
  const sig = await crypto.subtle.sign({ name: "ECDSA", hash: "SHA-256" }, privateKey, data);
  return bufToB64(sig);
}

const ClientLogin: React.FC = () => {
  const navigate = useNavigate();

  const [fileName, setFileName] = useState<string>("");
  const [fileErr, setFileErr] = useState<string>("");
  const [ks, setKs] = useState<KeystoreFile | null>(null);

  const [password, setPassword] = useState<string>("");
  const [busy, setBusy] = useState<boolean>(false);
  const [status, setStatus] = useState<string>("");

  const hasEthereum = useMemo(() => typeof (window as any).ethereum !== "undefined", []);

  async function requestNonce(walletType: "rc" | "evm", address: string) {
    const resp = await fetch("/api/auth/nonce", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      credentials: "include",
      body: JSON.stringify({ wallet_type: walletType, address }),
    });
    const data = await resp.json().catch(() => ({}));
    if (!resp.ok || !data?.ok) {
      throw new Error(data?.error || "nonce_failed");
    }
    return data.challenge as string;
  }

  async function loginWithRCWallet(ksFile: KeystoreFile, pass: string) {
    const address = ksFile.wallet.address;
    const challenge = await requestNonce("rc", address);

    const { privKey, pubJwk } = await decryptKeystore(ksFile, pass);

    const sigAsn1 = await signChallengeP256(privKey, challenge);

    const resp = await fetch("/api/auth/wallet-login", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      credentials: "include",
      body: JSON.stringify({
        wallet_type: "rc",
        address,
        pub: pubJwk,
        challenge,
        signature_b64: btoa(String.fromCharCode(...new Uint8Array(sigAsn1))),
        role: "client",
      }),
    });

    const data = await resp.json().catch(() => ({}));
    if (!resp.ok || !data?.ok) {
      throw new Error(data?.error || "login_failed");
    }
  }

  async function loginWithMetaMask() {
    const eth = (window as any).ethereum;
    if (!eth) throw new Error("metamask_not_found");

    const accounts = (await eth.request({ method: "eth_requestAccounts" })) as string[];
    const address = accounts?.[0];
    if (!address) throw new Error("no_account");

    const challenge = await requestNonce("evm", address);

    // personal_sign expects params [message, address]
    const sig = (await eth.request({
      method: "personal_sign",
      params: [challenge, address],
    })) as string;

    const resp = await fetch("/api/auth/wallet-login", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      credentials: "include",
      body: JSON.stringify({
        wallet_type: "evm",
        address,
        challenge,
        signature_hex: sig,
        role: "client",
      }),
    });

    const data = await resp.json().catch(() => ({}));
    if (!resp.ok || !data?.ok) {
      throw new Error(data?.error || "login_failed");
    }
  }

  async function loginWithWalletConnect() {
    const projectId = (import.meta as any).env?.VITE_WC_PROJECT_ID;
    if (!projectId) throw new Error("walletconnect_missing_project_id");

    // Lazy import so builds don't crash if deps aren't installed yet.
    const mod = await import("@walletconnect/ethereum-provider");
    const EthereumProvider = (mod as any).default;

    const provider = await EthereumProvider.init({
      projectId,
      chains: [1],
      optionalChains: [1],
      showQrModal: true,
      methods: ["personal_sign", "eth_requestAccounts"],
      events: ["accountsChanged", "chainChanged", "disconnect"],
    });

    await provider.connect();
    const accounts = (provider.accounts || []) as string[];
    const address = accounts?.[0];
    if (!address) throw new Error("no_account");

    const challenge = await requestNonce("evm", address);

    const sig = (await provider.request({
      method: "personal_sign",
      params: [challenge, address],
    })) as string;

    const resp = await fetch("/api/auth/wallet-login", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      credentials: "include",
      body: JSON.stringify({
        wallet_type: "evm",
        address,
        challenge,
        signature_hex: sig,
        role: "client",
      }),
    });

    const data = await resp.json().catch(() => ({}));
    if (!resp.ok || !data?.ok) {
      throw new Error(data?.error || "login_failed");
    }

    try {
      await provider.disconnect();
    } catch {}
  }

  const canLoginRC = !!ks && password.length > 0 && !busy;

  async function onLoginRC() {
    if (!ks) return;
    setBusy(true);
    setStatus("Unlocking wallet…");
    setFileErr("");
    try {
      await loginWithRCWallet(ks, password);
      setStatus("Authenticated. Redirecting…");
      navigate("/client/dashboard");
    } catch (e: any) {
      setStatus("");
      setFileErr(e?.message || "Login failed");
    } finally {
      setBusy(false);
    }
  }

  async function onLoginMetaMask() {
    setBusy(true);
    setStatus("Connecting MetaMask…");
    setFileErr("");
    try {
      await loginWithMetaMask();
      setStatus("Authenticated. Redirecting…");
      navigate("/client/dashboard");
    } catch (e: any) {
      setStatus("");
      setFileErr(e?.message || "MetaMask login failed");
    } finally {
      setBusy(false);
    }
  }

  async function onLoginWC() {
    setBusy(true);
    setStatus("Opening WalletConnect…");
    setFileErr("");
    try {
      await loginWithWalletConnect();
      setStatus("Authenticated. Redirecting…");
      navigate("/client/dashboard");
    } catch (e: any) {
      setStatus("");
      setFileErr(e?.message || "WalletConnect login failed");
    } finally {
      setBusy(false);
    }
  }

  function onFileChange(ev: React.ChangeEvent<HTMLInputElement>) {
    setFileErr("");
    setStatus("");
    const f = ev.target.files?.[0];
    if (!f) {
      setKs(null);
      setFileName("");
      return;
    }
    setFileName(f.name);
    const reader = new FileReader();
    reader.onload = () => {
      try {
        const parsed = JSON.parse(String(reader.result || "{}"));
        setKs(parsed);
      } catch {
        setKs(null);
        setFileErr("Invalid wallet file");
      }
    };
    reader.readAsText(f);
  }

  return (
    <div>
      <h1 className="page-title">Client Portal</h1>
      <p className="page-subtitle">
        Sign in with an encrypted ReserveChain wallet file, MetaMask, or WalletConnect.
      </p>

      <div style={{ display: "grid", gap: 14, maxWidth: 560 }}>
        <div style={{ padding: 14, border: "1px solid rgba(255,255,255,0.06)", borderRadius: 12, background: "rgba(0,0,0,0.18)" }}>
          <div style={{ fontSize: 12, color: "var(--text-muted)", textTransform: "uppercase", letterSpacing: "0.12em" }}>
            ReserveChain Wallet File
          </div>

          <div style={{ marginTop: 10, display: "grid", gap: 10 }}>
            <input type="file" accept="application/json" onChange={onFileChange} />
            {fileName && (
              <div style={{ fontSize: 12, color: "var(--text-muted)" }}>
                Selected: <span style={{ color: "var(--text-main)" }}>{fileName}</span>
              </div>
            )}

            <input
              type="password"
              placeholder="Wallet password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              style={{ padding: "10px 12px", borderRadius: 10, border: "1px solid rgba(255,255,255,0.08)", background: "rgba(0,0,0,0.2)", color: "var(--text-main)" }}
            />

            <button
              disabled={!canLoginRC}
              onClick={onLoginRC}
              style={{
                padding: "10px 14px",
                borderRadius: 999,
                border: "1px solid rgba(77,163,255,0.8)",
                background: "var(--accent-soft)",
                color: "var(--accent)",
                cursor: canLoginRC ? "pointer" : "not-allowed",
                fontSize: "0.9rem",
                opacity: canLoginRC ? 1 : 0.5,
              }}
            >
              {busy ? "Working…" : "Unlock & Sign In"}
            </button>
          </div>
        </div>

        <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 12 }}>
          <button
            disabled={busy || !hasEthereum}
            onClick={onLoginMetaMask}
            style={{
              padding: "12px 14px",
              borderRadius: 12,
              border: "1px solid rgba(255,255,255,0.08)",
              background: "rgba(0,0,0,0.18)",
              color: "var(--text-main)",
              cursor: busy || !hasEthereum ? "not-allowed" : "pointer",
              opacity: busy || !hasEthereum ? 0.5 : 1,
            }}
          >
            🦊 Connect MetaMask
          </button>

          <button
            disabled={busy}
            onClick={onLoginWC}
            style={{
              padding: "12px 14px",
              borderRadius: 12,
              border: "1px solid rgba(255,255,255,0.08)",
              background: "rgba(0,0,0,0.18)",
              color: "var(--text-main)",
              cursor: busy ? "not-allowed" : "pointer",
              opacity: busy ? 0.5 : 1,
            }}
          >
            📱 WalletConnect
          </button>
        </div>

        {status && <div style={{ fontSize: 13, color: "var(--text-muted)" }}>{status}</div>}
        {fileErr && <div style={{ fontSize: 13, color: "#ff8b8b" }}>{fileErr}</div>}

        {!hasEthereum && (
          <div style={{ fontSize: 12, color: "var(--text-muted)" }}>
            MetaMask not detected. Install MetaMask to use extension-based sign-in.
          </div>
        )}

        <div style={{ fontSize: 12, color: "var(--text-muted)", lineHeight: 1.4 }}>
          WalletConnect requires a WalletConnect Cloud Project ID. Set{" "}
          <code>VITE_WC_PROJECT_ID</code> in your workstation portal environment before building.
        </div>
      </div>
    </div>
  );
};

export default ClientLogin;
