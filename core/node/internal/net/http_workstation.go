package net

import (
    "net/http"
    "os"
    "path"
    "path/filepath"
    "strings"
    "time"
)

// workstationHandler serves the built workstation portal (Vite dist) at /workstation/.
// It supports SPA routing by falling back to index.html for unknown paths.
func (api *HTTPAPI) workstationHandler(w http.ResponseWriter, r *http.Request) {
    // Strip the /workstation prefix.
    reqPath := strings.TrimPrefix(r.URL.Path, "/workstation")
    if reqPath == "" || reqPath == "/" {
        reqPath = "/index.html"
    }

    // Resolve filesystem path safely.
    clean := path.Clean("/" + reqPath) // force leading slash
    if strings.Contains(clean, "..") {
        http.Error(w, "invalid path", http.StatusBadRequest)
        return
    }

    dist := api.WorkstationDist
    // If dist doesn't exist, give a helpful error.
    if st, err := os.Stat(dist); err != nil || !st.IsDir() {
        http.Error(w, "workstation portal not built. Run: cd apps/workstation_portal && npm install && npm run build", http.StatusServiceUnavailable)
        return
    }

    fsPath := filepath.Join(dist, filepath.FromSlash(clean))
    // If requested file doesn't exist, serve index.html (SPA fallback)
    if _, err := os.Stat(fsPath); err != nil {
        fsPath = filepath.Join(dist, "index.html")
    }

    // Cache policy:
    // - index.html: no-cache (so new deploy loads immediately)
    // - assets: long cache (hashed filenames)
    if strings.HasSuffix(fsPath, "index.html") {
        w.Header().Set("Cache-Control", "no-store")
    } else if strings.Contains(fsPath, string(filepath.Separator)+"assets"+string(filepath.Separator)) {
        w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
    } else {
        w.Header().Set("Cache-Control", "public, max-age=600")
    }

    // Basic security headers
    w.Header().Set("X-Content-Type-Options", "nosniff")
    w.Header().Set("Referrer-Policy", "no-referrer")
    w.Header().Set("X-Frame-Options", "DENY")

    // Serve file
    http.ServeFile(w, r, fsPath)

    _ = time.Now() // keep imports stable if you later add access logging
}
