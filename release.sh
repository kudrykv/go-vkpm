#!/usr/bin/env bash

set -e

VERSION=${VERSION:-undefined}

gooses=(darwin linux windows)
goarches=(amd64)

for goos in "${gooses[@]}"; do
  for goarch in "${goarches[@]}"; do
    filename=vkpm
    if [[ $goos = "windows" ]]; then
      filename=vkpm.exe
    fi

    GOOS=${goos} GOARCH=${goarch} CGO_ENABLED=0 go build -o ${filename}

    gzip ${filename}
    mv ${filename}.gz vkpm_"${goos}"_"${goarch}"_"${VERSION}".gz
  done
done

