
// api_client.js
// Wrap fetch with timeout, JSON parsing, and consistent error shape.

(function () {
  const DEFAULT_TIMEOUT_MS = 10000;

  class APIError extends Error {
    constructor(message, opts) {
      super(message);
      this.name = 'APIError';
      this.status = opts?.status ?? null;
      this.code = opts?.code ?? null;
      this.details = opts?.details ?? null;
      this.url = opts?.url ?? null;
    }
  }

  function withTimeout(promise, ms, url) {
    let timeoutId;
    const t = new Promise((_, reject) => {
      timeoutId = setTimeout(
        () => reject(new APIError('Request timed out', { status: 0, code: 'TIMEOUT', url })),
        ms
      );
    });
    return Promise.race([
      promise.finally(() => clearTimeout(timeoutId)),
      t,
    ]);
  }

  async function safeFetchJSON(url, options = {}, timeoutMs = DEFAULT_TIMEOUT_MS) {
    const opts = {
      credentials: 'include',
      ...options,
      headers: {
        Accept: 'application/json',
        ...(options.headers || {}),
      },
    };

    try {
      const res = await withTimeout(fetch(url, opts), timeoutMs, url);

      let data = null;
      const text = await res.text();
      try {
        data = text ? JSON.parse(text) : null;
      } catch (e) {
        throw new APIError('Invalid JSON response from server', {
          status: res.status,
          code: 'BAD_JSON',
          url,
        });
      }

      if (!res.ok) {
        throw new APIError(data?.error || 'Request failed', {
          status: res.status,
          code: data?.code || 'HTTP_' + res.status,
          details: data,
          url,
        });
      }

      return data;
    } catch (err) {
      if (!(err instanceof APIError)) {
        throw new APIError(err.message || 'Network or CORS error', {
          status: 0,
          code: 'NETWORK',
          url,
        });
      }
      throw err;
    }
  }

  window.APIError = APIError;
  window.safeFetchJSON = safeFetchJSON;
})();
