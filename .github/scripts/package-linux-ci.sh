#!/usr/bin/env bash
# Run package-linux.sh inside manylinux2014 (glibc 2.17) so CGO binaries
# start on CentOS 7 / RHEL 7 and later. ubuntu-latest's glibc is too new.
# Env: same as package-linux.sh (VERSION, ORACLE, ORACLE_HOME, CGO_*).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT"

if [[ -z "${VERSION:-}" ]]; then
  echo "[ERROR] VERSION is required" >&2
  exit 1
fi

IMAGE="${LINUX_BUILD_IMAGE:-quay.io/pypa/manylinux2014_x86_64}"
GOROOT="$(go env GOROOT)"
GOMODCACHE="$(go env GOMODCACHE)"
if [[ -z "$GOROOT" || ! -x "$GOROOT/bin/go" ]]; then
  echo "[ERROR] GOROOT is missing; run setup-go on the host first" >&2
  exit 1
fi
mkdir -p "$GOMODCACHE"

ORACLE="${ORACLE:-0}"
# Paths inside the container. Host workspace is mounted at /src.
CONTAINER_ORACLE_HOME=""
if [[ "$ORACLE" == "1" ]]; then
  if [[ -z "${ORACLE_HOME:-}" ]]; then
    echo "[ERROR] ORACLE_HOME is required when ORACLE=1" >&2
    exit 1
  fi
  rel="${ORACLE_HOME#"$ROOT"/}"
  if [[ "$rel" == "$ORACLE_HOME" ]]; then
    echo "[ERROR] ORACLE_HOME must be inside the workspace: $ORACLE_HOME" >&2
    exit 1
  fi
  CONTAINER_ORACLE_HOME="/src/${rel}"
fi

echo "[INFO] Linux package image=${IMAGE} (glibc 2.17 ABI)"
echo "[INFO] GOROOT=${GOROOT}"
docker pull "$IMAGE"

DOCKER_ENV=(
  -e VERSION
  -e ORACLE
  -e GOTOOLCHAIN=local
)
if [[ "$ORACLE" == "1" ]]; then
  DOCKER_ENV+=(
    -e "ORACLE_HOME=${CONTAINER_ORACLE_HOME}"
    -e "CGO_CFLAGS=-I${CONTAINER_ORACLE_HOME}/sdk/include"
    -e "CGO_LDFLAGS=-L${CONTAINER_ORACLE_HOME} -lclntsh"
    -e "LD_LIBRARY_PATH=${CONTAINER_ORACLE_HOME}"
  )
fi

docker run --rm \
  "${DOCKER_ENV[@]}" \
  -e GOPROXY="${GOPROXY:-https://proxy.golang.org,direct}" \
  -e GOSUMDB="${GOSUMDB:-sum.golang.org}" \
  -e GOCACHE=/tmp/go-cache \
  -e GOMODCACHE=/go-mod \
  -e GIT_COMMIT="$(git rev-parse --short HEAD 2>/dev/null || echo unknown)" \
  -v "$ROOT:/src" \
  -v "$GOROOT:/usr/local/go:ro" \
  -v "$GOMODCACHE:/go-mod" \
  -w /src \
  "$IMAGE" \
  bash -lc '
    set -euo pipefail
    export PATH=/usr/local/go/bin:$PATH
    git config --global --add safe.directory /src
    if [[ "${ORACLE}" == "1" ]]; then
      yum install -y -q libaio-devel
    fi
    go version
    # Do not pipe ldd into head: SIGPIPE + pipefail => exit 141.
    ldd --version
    bash /src/.github/scripts/package-linux.sh
  '
