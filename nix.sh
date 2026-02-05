#!/usr/bin/env bash
set -euo pipefail

# Block macOS Apple Silicon (M1/M2/M3)
if [[ "$(uname -s)" == "Darwin" && "$(uname -m)" == "arm64" ]]; then
  cat <<'EOF'
❌ Nix dev shell is NOT supported on macOS Apple Silicon (M1/M2/M3).

Supported platforms:
  • Linux
  • macOS Intel (x86_64)

Please use system-installed tools on Apple Silicon, such as Go and Terraform installed via Homebrew.
EOF
  exit 1
fi

# Otherwise run nix-shell as before
exec env NIXPKGS_ALLOW_UNFREE=1 nix-shell shell.nix
