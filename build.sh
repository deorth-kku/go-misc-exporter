#!/bin/bash
args=("$@")
IFS=$'\n' sorted=($(sort <<<"${args[*]}"))
joined=$(IFS=+; echo "${sorted[*]}")
echo building "$joined"


if [ "$joined" == *hwmon* ] || [ "$joined" == *ryzenadj* ]; then
    export CGO_ENABLED=1
else
    export CGO_ENABLED=0
fi
export GOAMD64=v3

out="/opt/gme/go-misc-exporter"
go build -o "$out" -ldflags "-s -w" ./build/"$joined"
echo output file "$out"
