#!/usr/bin/env bash
#
# build_release.sh - build the hybrid (GUI+CLI) binary and package a release.
#
# Produces:
#   dist/sssd-inspector-<version>-linux-amd64/
#     sssd-inspector        stripped hybrid binary (wails build)
#     kb_articles/          SUSE knowledge base, loaded next to the executable
#     LICENSE               GPL-3.0-or-later notice for this project
#     LICENCE.md            full GNU GPLv3 text
#     README.md             project documentation
#     config.yaml           default configuration (optional at runtime)
#     build_instructions.txt
#   dist/sssd-inspector-<version>-linux-amd64.tar.gz (+ .sha256)
#
# Usage:
#   ./build_release.sh                build, verify and package
#   ./build_release.sh --skip-build   reuse an existing build/bin/sssd-inspector
#   ./build_release.sh --no-checksum  do not write the .sha256 file
#   ./build_release.sh -h             this help
set -euo pipefail

# Always work from the repository root, whoever invokes the script.
cd "$(dirname "$0")"

APP_NAME="sssd-inspector"
PLATFORM="linux/amd64"
ARCH_SUFFIX="linux-amd64"
BUILD_TAGS="webkit2_41"   # webkit2gtk-4.1 (Ubuntu 24.04 / SLES 15 SP7+)
LDFLAGS="-w -s"
KB_DIR="kb_articles"
DIST_DIR="dist"
BIN_PATH="build/bin/${APP_NAME}"

SKIP_BUILD=0
WRITE_CHECKSUM=1

usage() { sed -n '2,22p' "$0"; }

for arg in "$@"; do
    case "$arg" in
        --skip-build)  SKIP_BUILD=1 ;;
        --no-checksum) WRITE_CHECKSUM=0 ;;
        -h|--help)     usage; exit 0 ;;
        *) printf 'Unknown option: %s (try -h)\n' "$arg" >&2; exit 2 ;;
    esac
done

die() { printf 'ERROR: %s\n' "$*" >&2; exit 1; }

# --- Version: the single source of truth is constants.AppVersion --------------
VERSION="$(sed -n 's/^[[:space:]]*AppVersion[[:space:]]*=[[:space:]]*"\([^"]*\)".*/\1/p' \
    constants/constants.go | head -n 1)"
[ -n "$VERSION" ] || die "cannot extract AppVersion from constants/constants.go"

STAGE_NAME="${APP_NAME}-${VERSION}-${ARCH_SUFFIX}"
STAGE_DIR="${DIST_DIR}/${STAGE_NAME}"
TARBALL="${DIST_DIR}/${STAGE_NAME}.tar.gz"
BUILD_LOG="${DIST_DIR}/build-${VERSION}.log"

printf '==> %s v%s - %s (tags: %s)\n' "$APP_NAME" "$VERSION" "$PLATFORM" "$BUILD_TAGS"

# --- Preflight -----------------------------------------------------------------
command -v wails  >/dev/null 2>&1 || die "wails CLI not found in PATH"
command -v go     >/dev/null 2>&1 || die "go not found in PATH"
command -v tar    >/dev/null 2>&1 || die "tar not found in PATH"
pkg-config --exists webkit2gtk-4.1 \
    || die "webkit2gtk-4.1 development files missing (required by -tags ${BUILD_TAGS})"
[ -d "$KB_DIR" ] || die "${KB_DIR}/ not found"
KB_COUNT="$(find "$KB_DIR" -maxdepth 1 -name '*.json' | wc -l)"
[ "$KB_COUNT" -gt 0 ] || die "${KB_DIR}/ contains no .json articles"
for f in LICENSE LICENCE.md README.md config.yaml build_instructions.txt; do
    [ -f "$f" ] || die "missing $f"
done

mkdir -p "$DIST_DIR"

# --- Build ---------------------------------------------------------------------
if [ "$SKIP_BUILD" -eq 0 ]; then
    # The canonical build command (see build_instructions.txt). Output is
    # captured in a log; pipefail makes a failed wails abort the script.
    if ! wails build -platform "$PLATFORM" -tags "$BUILD_TAGS" \
            -ldflags "$LDFLAGS" -clean 2>&1 | tee "$BUILD_LOG"; then
        die "wails build failed (see ${BUILD_LOG})"
    fi
fi

[ -x "$BIN_PATH" ] || die "expected binary ${BIN_PATH} not found"

# The binary prints its version before touching any display (runHybridCLI
# returns before launchGUI), so this smoke test is safe on headless machines
# and it catches a binary built with a stale version.
"$BIN_PATH" -v | tee "${DIST_DIR}/version.txt"
grep -q "version ${VERSION} " "${DIST_DIR}/version.txt" \
    || die "binary reports a different version than constants.AppVersion (${VERSION})"

# --- Stage ----------------------------------------------------------------------
rm -rf "$STAGE_DIR"
mkdir -p "$STAGE_DIR"
install -m 0755 "$BIN_PATH" "${STAGE_DIR}/${APP_NAME}"
cp -a "$KB_DIR" "${STAGE_DIR}/kb_articles"
cp -a LICENSE LICENCE.md README.md config.yaml build_instructions.txt "$STAGE_DIR/"

# --- Package --------------------------------------------------------------------
tar -C "$DIST_DIR" -czf "$TARBALL" "$STAGE_NAME"
if [ "$WRITE_CHECKSUM" -eq 1 ]; then
    ( cd "$DIST_DIR" && sha256sum "$(basename "$TARBALL")" > "$(basename "$TARBALL").sha256" )
fi

printf '\n==> Done\n'
printf '    binary : %s (%s)\n' "${STAGE_DIR}/${APP_NAME}" "$(du -h "${STAGE_DIR}/${APP_NAME}" | cut -f1)"
printf '    kb     : %s articles\n' "$KB_COUNT"
printf '    tarball: %s (%s)\n' "$TARBALL" "$(du -h "$TARBALL" | cut -f1)"
if [ "$WRITE_CHECKSUM" -eq 1 ]; then
    printf '    sha256 : %s\n' "$(cut -d' ' -f1 "${TARBALL}.sha256")"
fi
printf '    content:\n'
tar -tzf "$TARBALL" | sed 's/^/      /'
