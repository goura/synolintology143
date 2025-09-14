#!/bin/sh
set -eu

# Usage (from GoReleaser build hook):
#   bash scripts/build-ipk.sh "<binary_path>" "<version>" "<target>" "<name>"
# Example args:
#   BIN=/home/runner/work/repo/dist/synolintology143_linux_amd64/synolintology143
#   VERSION=v0.1.0 (from GoReleaser); we strip the leading 'v' for opkg
#   TARGET=linux_amd64 or linux_arm64
#   NAME=synolintology143

BIN_ABS=${1:?binary path required}
VERSION_RAW=${2:?version required}
TARGET=${3:?target required}
NAME=${4:?package name required}

# Strip a single leading 'v' if present (opkg prefers plain semver)
VERSION=${VERSION_RAW#v}

case "$TARGET" in
  linux_arm64)
    ENTWARE_ARCH="aarch64-3.10"
    ;;
  linux_amd64)
    ENTWARE_ARCH="x64-3.2"
    ;;
  *)
    echo "unsupported target: $TARGET" >&2
    exit 1
    ;;
esac

WORK_DIR="dist/ipk-${TARGET}"
CTRL_DIR="$WORK_DIR/CONTROL"
DEST_DIR="$WORK_DIR/opt/bin"

rm -rf "$WORK_DIR"
mkdir -p "$CTRL_DIR" "$DEST_DIR"

# Install binary into /opt/bin
cp "$BIN_ABS" "$DEST_DIR/$NAME"
chmod 0755 "$DEST_DIR/$NAME"

# Control metadata
cat >"$CTRL_DIR/control" <<EOF
Package: $NAME
Version: $VERSION
Architecture: $ENTWARE_ARCH
Section: utils
Priority: optional
Maintainer: Goura <goura@example.com>
Homepage: https://github.com/goura/synolintology143
Description: Find filenames exceeding eCryptfs 143-byte limit.
EOF

# Build .ipk into dist/ (use absolute dest since opkg-build chdirs)
OUT_DIR="$(pwd)/dist"
mkdir -p "$OUT_DIR"
opkg-build -Z gzip "$WORK_DIR" "$OUT_DIR"

IPK_PATH="$OUT_DIR/${NAME}_${VERSION}_${ENTWARE_ARCH}.ipk"
if [ -f "$IPK_PATH" ]; then
  echo "Built IPK: $IPK_PATH"
else
  echo "IPK not found at expected path: $IPK_PATH" >&2
  echo "dist/ contents:" >&2
  ls -la dist >&2 || true
  exit 1
fi
