{ pkgs ? import <nixpkgs> {} }:

let
  # Pull in nixpkgs standard library helpers
  lib = pkgs.lib;

  # Absolute path to the repository root.
  # Used for GOPATH setup inside shellHook.
  rootDir = builtins.toString ./.;

  # Detect macOS (Darwin) vs Linux.
  # IMPORTANT: this is used to *avoid evaluating* unsupported packages on macOS.
  isDarwin = pkgs.stdenv.isDarwin;

  # Path to Chromium binary from nixpkgs.
  #
  # CRITICAL DETAIL:
  #   We must NOT reference pkgs.chromium on Darwin at all.
  #   Even interpolating "${pkgs.chromium}" will fail evaluation on x86_64-darwin.
  #
  # Therefore, this is conditionally defined.
  chromiumPath =
    if isDarwin
    then ""
    else "${pkgs.chromium}/bin/chromium";
in
pkgs.mkShell {

  ############################################
  # Tools available in the development shell
  ############################################
  buildInputs = with pkgs; [
    terraform_1
    go_1_25
    git
    golangci-lint
    pre-commit
    nodejs_24
    nodePackages.mermaid-cli
    terraform-plugin-docs
    goreleaser
    gnupg
  ]
  # Chromium is ONLY added on non-Darwin systems.
  #
  # On macOS (x86_64-darwin), chromium is not supported in nixpkgs,
  # so attempting to include it would cause evaluation to fail.
  ++ lib.optionals (!isDarwin) [ pkgs.chromium ];

  ############################################
  # shellHook runs every time you enter nix-shell
  ############################################
  shellHook = ''
    set -e

    ############################################
    # Repository & Go environment setup
    ############################################

    # Absolute repo path injected from Nix
    REPO_DIR=${rootDir}

    # If GOPATH is not set, default to the repository root.
    # This matches the existing project workflow.
    if [ -z "$GOPATH" ]; then
      export GOPATH="$REPO_DIR"
    fi

    # Ensure the repo root is included in GOPATH.
    # Avoids clobbering an existing GOPATH if one is already defined.
    if [ -n "$GOPATH" ] && ! echo ":$GOPATH:" | grep -q ":$REPO_DIR:"; then
      export GOPATH="$REPO_DIR:$GOPATH"
    fi

    # Ensure Go-installed binaries (go install ...) are on PATH
    export PATH="$GOPATH/bin:$PATH"

    ############################################
    # Puppeteer / mermaid-cli configuration
    ############################################

    # Prevent Puppeteer from downloading its own Chromium binary.
    # We always provide a browser ourselves (system Chrome or nix Chromium).
    export PUPPETEER_SKIP_DOWNLOAD="1"

    # Puppeteer cache directory.
    # TMPDIR exists on macOS; on Linux it may not, so fall back to /tmp.
    export PUPPETEER_CACHE_DIR="${TMPDIR:-/tmp}/puppeteer"

    ############################################
    # OS-specific browser selection
    ############################################

    if [ "$(uname -s)" = "Darwin" ]; then
      # macOS:
      # Prefer system-installed browsers from /Applications.
      # This avoids nixpkgs Chromium issues on Darwin and matches user expectations.

      if [ -x "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome" ]; then
        export PUPPETEER_EXECUTABLE_PATH="/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"
      elif [ -x "/Applications/Chromium.app/Contents/MacOS/Chromium" ]; then
        export PUPPETEER_EXECUTABLE_PATH="/Applications/Chromium.app/Contents/MacOS/Chromium"
      else
        echo "⚠️  No Chrome/Chromium found in /Applications; mermaid-cli may fail."
      fi

    else
      # Linux:
      # Use Chromium provided by nixpkgs (guaranteed reproducible).
      export PUPPETEER_EXECUTABLE_PATH="${chromiumPath}"

      ############################################
      # Install required Go CLI tools if missing
      ############################################

      # Use the exact Go version provided by nixpkgs
      GO_BIN=${pkgs.go_1_25}/bin/go

      # godotenv is required by project scripts
      if ! command -v godotenv >/dev/null 2>&1; then
        echo "Installing godotenv..."
        "$GO_BIN" install github.com/joho/godotenv/cmd/godotenv@latest
      fi

      # task (go-task) is used for automation
      if ! command -v task >/dev/null 2>&1; then
        echo "Installing task (go-task/task)..."
        "$GO_BIN" install github.com/go-task/task/v3/cmd/task@latest
      fi
    fi
    set +e

    ############################################
    # Debug output on shell entry
    ############################################
    echo "PUPPETEER_EXECUTABLE_PATH: $PUPPETEER_EXECUTABLE_PATH"
  '';
}
