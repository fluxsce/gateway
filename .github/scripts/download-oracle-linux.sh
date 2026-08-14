#!/usr/bin/env bash
# Download Oracle Instant Client Basic + SDK for Linux amd64 (compile-time only).
# Does not bundle client libraries into release archives (OTN license).
set -euo pipefail

IC_VERSION="${ORACLE_IC_VERSION:-21.18.0.0.0dbru}"
IC_PATH="${ORACLE_IC_PATH:-2118000}"
DEST_ROOT="${ORACLE_DIR:-${GITHUB_WORKSPACE:-$PWD}/.oracle}"
EXTRACT_DIR="${DEST_ROOT}/instantclient_21_18"
BASE_URL="https://download.oracle.com/otn_software/linux/instantclient/${IC_PATH}"
BASIC_ZIP="instantclient-basic-linux.x64-${IC_VERSION}.zip"
SDK_ZIP="instantclient-sdk-linux.x64-${IC_VERSION}.zip"

mkdir -p "${DEST_ROOT}"

if [[ -f "${EXTRACT_DIR}/sdk/include/oci.h" ]] && [[ -e "${EXTRACT_DIR}/libclntsh.so" || -n "$(ls "${EXTRACT_DIR}"/libclntsh.so.* 2>/dev/null || true)" ]]; then
  echo "[INFO] Oracle Instant Client already present: ${EXTRACT_DIR}"
else
  echo "[INFO] Downloading Oracle Instant Client ${IC_VERSION} into ${DEST_ROOT}"
  tmpdir="$(mktemp -d)"
  trap 'rm -rf "${tmpdir}"' EXIT
  curl -fsSL --retry 5 --retry-delay 10 -o "${tmpdir}/${BASIC_ZIP}" "${BASE_URL}/${BASIC_ZIP}"
  curl -fsSL --retry 5 --retry-delay 10 -o "${tmpdir}/${SDK_ZIP}" "${BASE_URL}/${SDK_ZIP}"
  rm -rf "${EXTRACT_DIR}"
  unzip -qo "${tmpdir}/${BASIC_ZIP}" -d "${DEST_ROOT}"
  unzip -qo "${tmpdir}/${SDK_ZIP}" -d "${DEST_ROOT}"
  trap - EXIT
  rm -rf "${tmpdir}"
fi

if [[ ! -f "${EXTRACT_DIR}/sdk/include/oci.h" ]]; then
  echo "[ERROR] oci.h not found under ${EXTRACT_DIR}/sdk/include" >&2
  exit 1
fi

ORACLE_HOME="${EXTRACT_DIR}"
CGO_CFLAGS="-I${ORACLE_HOME}/sdk/include"
CGO_LDFLAGS="-L${ORACLE_HOME} -lclntsh"
LD_LIBRARY_PATH="${ORACLE_HOME}${LD_LIBRARY_PATH:+:$LD_LIBRARY_PATH}"

echo "[INFO] ORACLE_HOME=${ORACLE_HOME}"
echo "[INFO] CGO_CFLAGS=${CGO_CFLAGS}"
echo "[INFO] CGO_LDFLAGS=${CGO_LDFLAGS}"

if [[ -n "${GITHUB_ENV:-}" ]]; then
  {
    echo "ORACLE_HOME=${ORACLE_HOME}"
    echo "CGO_CFLAGS=${CGO_CFLAGS}"
    echo "CGO_LDFLAGS=${CGO_LDFLAGS}"
    echo "LD_LIBRARY_PATH=${LD_LIBRARY_PATH}"
  } >> "${GITHUB_ENV}"
fi
