#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."
export PATH="$HOME/.local/bin:${ASDF_DATA_DIR:-$HOME/.asdf}/bin:${ASDF_DATA_DIR:-$HOME/.asdf}/shims:$PATH"

missing=0
for dependency in make git curl tar gzip awk grep python3 docker; do
    if ! command -v "$dependency" >/dev/null 2>&1; then
        echo "Error: install $dependency using your system package manager, then run make setup again." >&2
        missing=1
    else
        echo "OK: $dependency available."
    fi
done

if (( BASH_VERSINFO[0] < 4 || (BASH_VERSINFO[0] == 4 && BASH_VERSINFO[1] < 3) )); then
    echo "Error: Bash 4.3 or newer is required by make dev (wait -n)." >&2
    missing=1
fi
if command -v make >/dev/null 2>&1 && ! make --version | grep -q 'GNU Make'; then
    echo "Error: GNU Make is required." >&2
    missing=1
fi
if command -v python3 >/dev/null 2>&1 && ! python3 -c 'import sys; sys.exit(sys.version_info.major != 3)'; then
    echo "Error: python3 must run Python 3." >&2
    missing=1
fi
if command -v docker >/dev/null 2>&1; then
    if ! docker compose version >/dev/null 2>&1; then
        echo "Error: install the Docker Compose plugin (docker compose); standalone docker-compose is not used." >&2
        missing=1
    else
        echo "OK: Docker Compose plugin available."
    fi
    if ! docker info >/dev/null 2>&1; then
        echo "Error: Docker daemon unavailable. Start Docker and ensure your user can access it (docker info)." >&2
        missing=1
    else
        echo "OK: Docker daemon accessible."
    fi
fi
if [ "$missing" -ne 0 ]; then
    echo "Fix the prerequisites above and run make setup again. No toolchains were installed." >&2
    exit 1
fi

if ! command -v asdf >/dev/null 2>&1; then
    case "$(uname -s)" in
        Linux) platform=linux ;;
        Darwin) platform=darwin ;;
        *) echo "Error: automatic asdf installation supports Linux and macOS." >&2; exit 1 ;;
    esac
    case "$(uname -m)" in
        x86_64) architecture=amd64 ;;
        aarch64|arm64) architecture=arm64 ;;
        *) echo "Error: unsupported architecture for automatic asdf installation." >&2; exit 1 ;;
    esac
    echo "Installing asdf v0.20.2..."
    archive=$(mktemp)
    trap 'rm -f "$archive"' EXIT
    curl -fL "https://github.com/asdf-vm/asdf/releases/download/v0.20.2/asdf-v0.20.2-$platform-$architecture.tar.gz" -o "$archive"
    mkdir -p "$HOME/.local/bin"
    tar -xzf "$archive" -C "$HOME/.local/bin" asdf
else
    echo "OK: asdf already installed ($(asdf --version))."
fi

for tool in nodejs golang; do
    version=$(awk -v tool="$tool" '$1 == tool { print $2 }' .tool-versions)
    if [ -z "$version" ]; then
        echo "Error: $tool version missing from .tool-versions." >&2
        exit 1
    fi
    if ! asdf plugin list | grep -qx "$tool"; then
        case "$tool" in
            nodejs) repository=https://github.com/asdf-vm/asdf-nodejs.git ;;
            golang) repository=https://github.com/asdf-community/asdf-golang.git ;;
        esac
        echo "Installing asdf plugin: $tool..."
        asdf plugin add "$tool" "$repository"
    fi
    if asdf where "$tool" "$version" >/dev/null 2>&1; then
        echo "OK: $tool $version already installed."
    else
        echo "Installing $tool $version..."
        asdf install "$tool" "$version"
    fi
    asdf reshim "$tool" "$version"
done

# Validate the actual executables, including npm bundled with Node.js.
asdf exec node --version
asdf exec npm --version
asdf exec go version
echo 'OK: Node.js, npm and Go are ready. Run make dev to start the application.'
