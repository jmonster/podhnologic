#!/bin/sh
set -e

# Podhnologic Installer
# Usage: curl -fsSL https://raw.githubusercontent.com/jmonster/podhnologic/main/install.sh | sh
# Note: POSIX-compliant, works with sh, dash, bash, etc.

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration
REPO="jmonster/podhnologic"
BINARY_NAME="podhnologic"

# Set install directory based on OS (default: ~/.local/bin for Unix, no sudo required)
case "$(uname -s)" in
    Darwin*|Linux*)
        INSTALL_DIR="${HOME}/.local/bin"
        ;;
    MINGW*|MSYS*|CYGWIN*)
        INSTALL_DIR="$LOCALAPPDATA/Podhnologic"
        ;;
    *)
        INSTALL_DIR="${HOME}/.local/bin"
        ;;
esac

TMP_DIR=$(mktemp -d)

# Cleanup on exit
trap 'rm -rf "${TMP_DIR}"' EXIT

# Print banner
print_banner() {
    echo ""
    printf "${CYAN}╔═══════════════════════════════════════════════════════════╗${NC}\n"
    printf "${CYAN}║${NC}                                                           ${CYAN}║${NC}\n"
    printf "${CYAN}║${NC}              ${GREEN}PODHNOLOGIC INSTALLER${NC}                       ${CYAN}║${NC}\n"
    printf "${CYAN}║${NC}                                                           ${CYAN}║${NC}\n"
    printf "${CYAN}║${NC}     Convert your music collection to another format       ${CYAN}║${NC}\n"
    printf "${CYAN}║${NC}                                                           ${CYAN}║${NC}\n"
    printf "${CYAN}╚═══════════════════════════════════════════════════════════╝${NC}\n"
    echo ""
}

# Detect OS and architecture
detect_platform() {
    local os=""
    local arch=""

    # Detect OS
    case "$(uname -s)" in
        Darwin*)
            os="darwin"
            ;;
        Linux*)
            os="linux"
            ;;
        MINGW*|MSYS*|CYGWIN*)
            os="windows"
            ;;
        *)
            printf "${RED}Unsupported operating system: $(uname -s)${NC}\n"
            exit 1
            ;;
    esac

    # Detect architecture
    case "$(uname -m)" in
        x86_64|amd64)
            arch="amd64"
            ;;
        aarch64|arm64)
            arch="arm64"
            ;;
        *)
            printf "${RED}Unsupported architecture: $(uname -m)${NC}\n"
            exit 1
            ;;
    esac

    echo "${os}-${arch}"
}

# Download with progress
download_file() {
    local url="$1"
    local output="$2"

    printf "${BLUE}Downloading ${BINARY_NAME}...${NC}\n"

    if command -v curl >/dev/null 2>&1; then
        curl -fSL --progress-bar "$url" -o "$output"
    elif command -v wget >/dev/null 2>&1; then
        wget --show-progress -q "$url" -O "$output"
    else
        printf "${RED}Neither curl nor wget found. Please install one of them.${NC}\n"
        exit 1
    fi
}

# Add to PATH on Windows
add_to_path_windows() {
    local install_dir="$1"
    local win_path=$(cygpath -w "$install_dir" 2>/dev/null || echo "$install_dir")

    if powershell.exe -Command "[Environment]::GetEnvironmentVariable('Path', 'User')" 2>/dev/null | grep -qi "podhnologic"; then
        printf "${GREEN}Already in PATH${NC}\n"
        return 0
    fi

    printf "${BLUE}Adding to PATH...${NC}\n"
    powershell.exe -Command "[Environment]::SetEnvironmentVariable('Path', [Environment]::GetEnvironmentVariable('Path', 'User') + ';${win_path}', 'User')" 2>/dev/null
    printf "${GREEN}Added to PATH (restart terminal to use 'podhnologic' command)${NC}\n"
}

# Add to PATH on Unix (bash/zsh/fish)
add_to_path_unix() {
    local install_dir="$1"
    local path_added=false
    local rc_file=""

    # Check if already in PATH
    case ":$PATH:" in
        *":${install_dir}:"*)
            printf "${GREEN}Already in PATH${NC}\n"
            return 0
            ;;
    esac

    printf "${BLUE}Adding to PATH...${NC}\n"

    # Detect shell and add to appropriate rc file
    local shell_name=$(basename "${SHELL:-/bin/sh}")

    case "$shell_name" in
        bash)
            if [ "$(uname -s)" = "Darwin" ]; then
                rc_file="$HOME/.bash_profile"
            else
                rc_file="$HOME/.bashrc"
            fi
            ;;
        zsh)
            rc_file="$HOME/.zshrc"
            ;;
        fish)
            rc_file="$HOME/.config/fish/config.fish"
            mkdir -p "$(dirname "$rc_file")" 2>/dev/null
            ;;
        *)
            rc_file="$HOME/.profile"
            ;;
    esac

    # Add to rc file if not already present
    if [ -f "$rc_file" ]; then
        if ! grep -q "${install_dir}" "$rc_file" 2>/dev/null; then
            echo "" >> "$rc_file"
            echo "# Added by Podhnologic installer" >> "$rc_file"
            if [ "$shell_name" = "fish" ]; then
                echo "fish_add_path ${install_dir}" >> "$rc_file"
            else
                echo "export PATH=\"\$PATH:${install_dir}\"" >> "$rc_file"
            fi
            path_added=true
        fi
    else
        echo "# Added by Podhnologic installer" > "$rc_file"
        if [ "$shell_name" = "fish" ]; then
            echo "fish_add_path ${install_dir}" >> "$rc_file"
        else
            echo "export PATH=\"\$PATH:${install_dir}\"" >> "$rc_file"
        fi
        path_added=true
    fi

    if [ "$path_added" = true ]; then
        printf "${GREEN}Added to PATH${NC}\n"
        printf "${YELLOW}Restart your terminal or run: source ~/${rc_file##*/}${NC}\n"
    else
        printf "${GREEN}Already in PATH config${NC}\n"
    fi
}

# Install binary
install_binary() {
    local binary_path="$1"
    local binary_name="$BINARY_NAME"

    # Add .exe extension on Windows
    case "$PLATFORM" in
        windows-*) binary_name="${BINARY_NAME}.exe" ;;
    esac

    # Create install directory if needed
    if [ ! -d "$INSTALL_DIR" ]; then
        mkdir -p "$INSTALL_DIR"
    fi

    local install_path="${INSTALL_DIR}/${binary_name}"

    # Install based on platform
    case "$PLATFORM" in
        windows-*)
            cp "$binary_path" "$install_path"
            add_to_path_windows "$INSTALL_DIR"
            ;;
        *)
            install -m 755 "$binary_path" "$install_path"
            add_to_path_unix "$INSTALL_DIR"
            ;;
    esac

    printf "${GREEN}Installed to ${install_path}${NC}\n"
}

# Check if already installed
check_existing_installation() {
    if command -v "$BINARY_NAME" >/dev/null 2>&1; then
        existing_path=$(command -v "$BINARY_NAME")
        printf "${YELLOW}Podhnologic is already installed at: ${existing_path}${NC}\n"
        echo ""
        printf "Update to the latest version? (y/N) "
        read -r REPLY
        echo ""

        case "$REPLY" in
            [Yy]|[Yy][Ee][Ss]) ;;
            *)
                printf "${YELLOW}Installation cancelled.${NC}\n"
                exit 0
                ;;
        esac

        return 0
    fi
    return 1
}

# Show next steps
show_next_steps() {
    echo ""
    printf "${CYAN}═══════════════════════════════════════════════════════════${NC}\n"
    printf "${BLUE}Getting Started${NC}\n"
    printf "${CYAN}═══════════════════════════════════════════════════════════${NC}\n"
    echo ""

    case "$PLATFORM" in
        windows-*)
            printf "${YELLOW}Restart your terminal, then run:${NC}\n"
            printf "  ${GREEN}podhnologic.exe${NC}\n"
            ;;
        *)
            printf "${GREEN}Run podhnologic:${NC}\n"
            printf "  ${YELLOW}podhnologic${NC}\n"
            ;;
    esac
    echo ""

    printf "${GREEN}Features:${NC}\n"
    printf "  - Interactive terminal UI with keyboard shortcuts\n"
    printf "  - FFmpeg embedded - no installation needed\n"
    printf "  - Multi-threaded conversion using all CPU cores\n"
    printf "  - iPod optimized conversions\n"
    echo ""

    printf "${GREEN}Learn more:${NC}\n"
    printf "  ${CYAN}https://github.com/${REPO}${NC}\n"
    echo ""
}

# Main installation
main() {
    print_banner

    # Check for existing installation
    local is_update=false
    if check_existing_installation; then
        is_update=true
    fi

    # Detect platform
    printf "${BLUE}Detecting platform...${NC}\n"
    PLATFORM=$(detect_platform)
    printf "${GREEN}Platform: ${PLATFORM}${NC}\n"
    echo ""

    # Build download URL
    local binary_suffix=""
    case "$PLATFORM" in
        windows-*) binary_suffix=".exe" ;;
    esac

    DOWNLOAD_URL="https://github.com/${REPO}/releases/latest/download/${BINARY_NAME}-${PLATFORM}${binary_suffix}"

    # Download binary
    BINARY_PATH="${TMP_DIR}/${BINARY_NAME}${binary_suffix}"
    download_file "$DOWNLOAD_URL" "$BINARY_PATH"

    # Check if download was successful
    if [ ! -f "$BINARY_PATH" ] || [ ! -s "$BINARY_PATH" ]; then
        echo ""
        printf "${RED}Download failed or file is empty${NC}\n"
        printf "${YELLOW}Please check:${NC}\n"
        printf "  1. Your internet connection\n"
        printf "  2. That releases are available for your platform (${PLATFORM})\n"
        printf "  3. The download URL: ${DOWNLOAD_URL}\n"
        exit 1
    fi
    echo ""

    # Install binary
    printf "${BLUE}Installing...${NC}\n"
    install_binary "$BINARY_PATH"
    echo ""

    # Show next steps (only for new installations)
    if [ "$is_update" = false ]; then
        show_next_steps
    fi

    # Success message
    echo ""
    printf "${GREEN}╔═══════════════════════════════════════════════════════════╗${NC}\n"
    printf "${GREEN}║                                                           ║${NC}\n"
    printf "${GREEN}║              Installation Complete!                       ║${NC}\n"
    printf "${GREEN}║                                                           ║${NC}\n"
    printf "${GREEN}╚═══════════════════════════════════════════════════════════╝${NC}\n"
    echo ""

    if [ "$is_update" = true ]; then
        printf "${CYAN}Updated to latest version!${NC}\n"
        echo ""
    fi
}

# Run main
main
