$ErrorActionPreference = "Stop"
New-Item -ItemType Directory -Force dist | Out-Null
$env:GOOS = "windows"
$env:GOARCH = "amd64"
go build -trimpath -ldflags "-s -w" -o dist/monitor-sender-windows-amd64.exe ./cmd/monitor-sender
Copy-Item config.example.json dist/config.example.json -Force
Write-Host "Built dist/monitor-sender-windows-amd64.exe"
