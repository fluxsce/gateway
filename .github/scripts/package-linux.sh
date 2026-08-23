#!/usr/bin/env bash
# CI-owned Linux packaging. Not used by local scripts/build.
# Release CI invokes this via package-linux-ci.sh (manylinux2014 / glibc 2.17).
# Env:
#   VERSION   required, e.g. 3.2.9
#   GOARCH    amd64 (default) or arm64
#   ORACLE    0 (MySQL/SQLite/ClickHouse) or 1 (also Oracle; amd64 only)
#   ORACLE_HOME required when ORACLE=1
set -euo pipefail

if [[ -z "${VERSION:-}" ]]; then
  echo "[ERROR] VERSION is required" >&2
  exit 1
fi

ORACLE="${ORACLE:-0}"
GOARCH="${GOARCH:-amd64}"
case "$GOARCH" in
  amd64|arm64) ;;
  *)
    echo "[ERROR] unsupported GOARCH=${GOARCH} (want amd64 or arm64)" >&2
    exit 1
    ;;
esac

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT"

if [[ ! -f go.mod ]]; then
  echo "[ERROR] go.mod not found in $ROOT" >&2
  exit 1
fi

if [[ ! -d web/frontend/dist ]]; then
  echo "[ERROR] web/frontend/dist missing; frontend job must run first" >&2
  exit 1
fi

if [[ "$ORACLE" == "1" ]]; then
  if [[ "$GOARCH" != "amd64" ]]; then
    echo "[ERROR] Oracle Instant Client 21.18 is linux x64 only; arm64 packages use no_oracle" >&2
    exit 1
  fi
  if [[ -z "${ORACLE_HOME:-}" ]]; then
    echo "[ERROR] ORACLE_HOME is required when ORACLE=1" >&2
    exit 1
  fi
  if [[ ! -f "${ORACLE_HOME}/sdk/include/oci.h" ]]; then
    echo "[ERROR] Oracle SDK header not found: ${ORACLE_HOME}/sdk/include/oci.h" >&2
    exit 1
  fi
  BUILD_TAGS="netgo,osusergo"
  VARIANT="linux-${GOARCH}-oracle"
  export CGO_CFLAGS="${CGO_CFLAGS:--I${ORACLE_HOME}/sdk/include}"
  export CGO_LDFLAGS="${CGO_LDFLAGS:--L${ORACLE_HOME} -lclntsh}"
  export LD_LIBRARY_PATH="${ORACLE_HOME}${LD_LIBRARY_PATH:+:$LD_LIBRARY_PATH}"
  echo "[INFO] Building with Oracle support"
else
  BUILD_TAGS="netgo,osusergo,no_oracle"
  VARIANT="linux-${GOARCH}"
  echo "[INFO] Building without Oracle (MySQL/SQLite/ClickHouse)"
fi

export CGO_ENABLED=1
export GOOS=linux
export GOARCH
echo "[INFO] Target GOOS=${GOOS} GOARCH=${GOARCH}"

GIT_COMMIT="${GIT_COMMIT:-$(git rev-parse --short HEAD 2>/dev/null || echo unknown)}"
BUILD_TIME="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
LDFLAGS="-s -w -X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME} -X main.GitCommit=${GIT_COMMIT}"

PACKAGE_DIR="dist/gateway"
ARCHIVE_NAME="gateway-${VARIANT}-${VERSION}.tar.gz"
rm -rf dist
mkdir -p "$PACKAGE_DIR"

echo "[INFO] go build gateway tags=${BUILD_TAGS}"
go build -tags "$BUILD_TAGS" -ldflags "$LDFLAGS" -o "$PACKAGE_DIR/gateway" cmd/app/main.go
chmod +x "$PACKAGE_DIR/gateway"

echo "[INFO] go build password_plugin"
go build -tags "$BUILD_TAGS" -ldflags "-s -w" -o "$PACKAGE_DIR/password_plugin" cmd/plugins/password_plugin/main.go
chmod +x "$PACKAGE_DIR/password_plugin"

copy_tree() {
  local src="$1" dest="$2"
  mkdir -p "$dest"
  if [[ -d "$src" ]]; then
    cp -a "$src"/. "$dest"/
  fi
}

copy_tree configs "$PACKAGE_DIR/configs"
copy_tree web/static "$PACKAGE_DIR/web/static"
copy_tree web/frontend/dist "$PACKAGE_DIR/web/frontend/dist"
copy_tree scripts/db "$PACKAGE_DIR/scripts/db"
copy_tree scripts/deploy "$PACKAGE_DIR/scripts/deploy"
copy_tree scripts/k8s "$PACKAGE_DIR/scripts/k8s"
copy_tree scripts/test "$PACKAGE_DIR/scripts/test"
mkdir -p "$PACKAGE_DIR/logs" "$PACKAGE_DIR/backup" "$PACKAGE_DIR/scripts/data" "$PACKAGE_DIR/pprof_analysis"

if [[ -d scripts/docker ]]; then
  copy_tree scripts/docker "$PACKAGE_DIR/scripts/docker"
  # Keep credentials out of release archives.
  rm -f "$PACKAGE_DIR/scripts/docker/push.sh"
fi

if [[ "$ORACLE" == "1" ]]; then
  echo "[INFO] Oracle Instant Client is not bundled (OTN license). Install it on the target host."
fi

# Release archive convention (Reproducible Builds / GoReleaser):
# root:root, dirs 0755, data 0644, binaries and shell scripts 0755.
find "$PACKAGE_DIR" -type d -exec chmod 755 {} +
find "$PACKAGE_DIR" -type f -exec chmod 644 {} +
chmod 755 "$PACKAGE_DIR/gateway" "$PACKAGE_DIR/password_plugin"
find "$PACKAGE_DIR" -type f -name '*.sh' -exec chmod 755 {} +

TAR_OWNER_OPTS=(--owner=0 --group=0 --numeric-owner)
if tar --help 2>&1 | grep -- '--sort' >/dev/null; then
  TAR_OWNER_OPTS+=(--sort=name)
fi
tar "${TAR_OWNER_OPTS[@]}" -czf "dist/${ARCHIVE_NAME}" -C dist gateway

if command -v objdump >/dev/null 2>&1; then
  echo "[INFO] highest GLIBC symbols in gateway:"
  objdump -T "$PACKAGE_DIR/gateway" 2>/dev/null | grep -oE 'GLIBC_[0-9.]+' | sort -uV | tail -5 || true
fi

echo "[OK] ${ARCHIVE_NAME} ($(du -h "dist/${ARCHIVE_NAME}" | cut -f1))"
ls -lh "$PACKAGE_DIR/gateway" "dist/${ARCHIVE_NAME}"
