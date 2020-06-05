#!/bin/bash

mkdir -p dist

VARIANTS="aix/ppc64 darwin/386 darwin/amd64 dragonfly/amd64 freebsd/386 freebsd/amd64 freebsd/arm freebsd/arm64 illumos/amd64 linux/386 linux/amd64 linux/arm linux/arm64 linux/mips linux/mips64 linux/mips64le linux/mipsle linux/ppc64 linux/ppc64le linux/riscv64 linux/s390x netbsd/386 netbsd/amd64 netbsd/arm netbsd/arm64 openbsd/386 openbsd/amd64 openbsd/arm openbsd/arm64 plan9/386 plan9/amd64 plan9/arm solaris/amd64 windows/386 windows/amd64 windows/arm"

for VARIANT in $VARIANTS; do

  IFS=/ read PLAT ARCH <<< $VARIANT

  if [ "$PLAT" = "windows" ]; then
    EXT=".exe"
  fi

  echo "Compiling $VARIANT"
  CGO_ENABLED=0 GOOS=$PLAT GOARCH=$ARCH go build -ldflags="-s -w" -o dist/3gsocks_client_${PLAT}_${ARCH}${EXT} 3gsocks_client.go
  CGO_ENABLED=0 GOOS=$PLAT GOARCH=$ARCH go build -ldflags="-s -w" -o dist/3gsocks_server_${PLAT}_${ARCH}${EXT} 3gsocks_server.go


done



