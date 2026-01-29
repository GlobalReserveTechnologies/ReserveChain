// Block Explorer wiring for NETWORK → Block Explorer panel.

(function () {
  async function fetchJSON(url) {
    const res = await fetch(url, { credentials: "same-origin" });
    if (!res.ok) throw new Error("HTTP " + res.status);
    return res.json();
  }

  function formatTs(ts) {
    if (!ts) return "-";
    try {
      const d = new Date(ts);
      return d.toLocaleString();
    } catch (e) {
      return String(ts);
    }
  }

  function shortHash(h) {
    if (!h) return "-";
    if (h.length <= 12) return h;
    return h.slice(0, 10) + "…" + h.slice(-6);
  }

  async function refreshExplorer() {
    const headEl = document.getElementById("explorer-head-height");
    const hashEl = document.getElementById("explorer-head-hash");
    const typeEl = document.getElementById("explorer-head-txtype");
    const tsEl = document.getElementById("explorer-head-ts");
    const tbody = document.getElementById("explorer-blocks-body");
    if (!tbody) return;

    try {
      const headResp = await fetchJSON("/api/chain/head");
      const head = headResp.head;

      if (head) {
        headEl.textContent = head.height;
        hashEl.textContent = shortHash(head.hash);
        typeEl.textContent = head.tx_type;
        tsEl.textContent = formatTs(head.timestamp);
      } else {
        headEl.textContent = "0";
        hashEl.textContent = "-";
        typeEl.textContent = "-";
        tsEl.textContent = "-";
      }

      const blocksResp = await fetchJSON("/api/chain/blocks?from_height=0&limit=50");
      const blocks = blocksResp.blocks || [];

      let html = "";
      if (!blocks.length) {
        html = '<tr><td colspan="4" class="explorer-empty">No blocks yet.</td></tr>';
      } else {
        for (let i = blocks.length - 1; i >= 0; i--) {
          const b = blocks[i];
          html += "<tr>" +
            "<td>" + b.height + "</td>" +
            "<td class=\"explorer-hash\">" + shortHash(b.hash) + "</td>" +
            "<td><span class=\"explorer-txtype explorer-txtype-" + String(b.tx_type || "").toLowerCase() + "\">" + b.tx_type + "</span></td>" +
            "<td>" + formatTs(b.timestamp) + "</td>" +
            "</tr>";
        }
      }
      tbody.innerHTML = html;
    } catch (e) {
      console.error("Failed to refresh explorer", e);
      tbody.innerHTML = '<tr><td colspan="4" class="explorer-empty">Error loading blocks.</td></tr>';
    }
  }

  document.addEventListener("DOMContentLoaded", function () {
    const panel = document.querySelector('[data-panel="network-explorer"]');
    if (!panel) return;

    const refreshBtn = document.getElementById("explorer-refresh");
    if (refreshBtn) {
      refreshBtn.addEventListener("click", function () {
        refreshExplorer();
      });
    }

    // Basic hook: when the panel becomes visible, refresh.
    const observer = new MutationObserver(() => {
      const display = panel.style.display || window.getComputedStyle(panel).display;
      if (display !== "none") {
        refreshExplorer();
      }
    });
    observer.observe(panel, { attributes: true, attributeFilter: ["style"] });

    // Initial idle refresh in case the panel starts visible.
    setTimeout(refreshExplorer, 500);
  });
})();
