#!/bin/bash
set -euo pipefail
mkdir -p dist/monitor-sender-macos
go build -trimpath -ldflags "-s -w" -o dist/monitor-sender-macos/monitor-sender ./cmd/monitor-sender
cp config.example.json dist/monitor-sender-macos/config.example.json
echo "已生成 dist/monitor-sender-macos/monitor-sender"
