#!/usr/bin/env bash
rm -rf builds/copa-agent-mssql-go_* || true
for GOOS in "windows" "linux" "darwin"; do
  for GOARCH in "386" "amd64" "arm64"; do
    if [ $GOOS == "windows" ]; then EXT=".exe"; else EXT=""; fi
    env GOOS=$GOOS GOARCH=$GOARCH go build -o ./builds/copa-agent-mssql-go_$GOOS"_"$GOARCH$EXT main.go
  done
done