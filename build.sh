#!/bin/bash
SRC="main.go"
NAME="stats"
OUTDIR="./build"
LDFLAGS='ldflags="-s -w"'

TARGETS=(
  "darwin-amd64"
  "freebsd-386"
  "linux-386"
  "linux-amd64"
  "netbsd-386"
  "solaris-amd64"
  "windows-amd64-.exe"
)

mkdir -p "${OUTDIR}"

HAS_UPX=$(command -v upx >/dev/null 2>&1 && echo 1)

if [ -z "$HAS_UPX" ]; then
  echo "Warning: upx binary not found. Skipping compression."
fi

for TARGET in "${TARGETS[@]}"; do
  T=(${TARGET//-/ })
  OS="${T[0]}"
  ARCH="${T[1]}"
  EXT="${T[2]}"
  OUTFILE="${OUTDIR}/${NAME}-${OS}-${ARCH}${EXT}"

  echo "Building OS ${OS} arch ${ARCH}..."
  GOOS=${OS} GOARCH=${ARCH} go build \
    -ldflags="${LDFLAGS}" \
    -o ${OUTFILE} \
    ${SRC}

  if [ -n "$HAS_UPX" ]; then
    echo "Compressing..."
    upx --brute ${OUTFILE}
  fi
done
