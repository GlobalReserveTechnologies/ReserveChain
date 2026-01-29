SQLite Runtime Binaries
-----------------------

This folder is reserved for platform-specific SQLite binaries used by the
ReserveChain DevNet runtime.

Because the build environment for this zip cannot download or ship real
executables for Windows / Linux / macOS, you should manually place the
appropriate `sqlite3` binaries here:

  - Windows:  sqlite3.exe
  - Linux:    sqlite3
  - macOS:    sqlite3

Once copied, the launch scripts and tooling can be configured to prefer
these binaries when present (before falling back to any system-installed
sqlite3 on PATH).
