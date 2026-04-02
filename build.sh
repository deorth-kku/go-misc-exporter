#!/bin/bash
set -euo pipefail

args=("$@")
if [ "${#args[@]}" -eq 0 ]; then
    mapfile -t args < <(find exporter -mindepth 1 -maxdepth 1 -type d -printf '%f\n' | sort)
else
    mapfile -t args < <(printf '%s\n' "${args[@]}" | sort -u)
fi

joined=$(IFS=+; echo "${args[*]}")
tags=$(IFS=' '; echo "${args[*]}")
echo building "$joined"

if [[ " ${args[*]} " == *" ryzenadj "* ]]; then
    export CGO_ENABLED=1
else
    export CGO_ENABLED=0
fi
export GOAMD64=v3

out="/opt/gme/go-misc-exporter"
go build -tags "$tags" -o "$out" -trimpath -ldflags "-s -w" ./main
echo output file "$out"
