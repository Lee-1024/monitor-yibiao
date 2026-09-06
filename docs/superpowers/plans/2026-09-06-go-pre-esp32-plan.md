# Go Pre-ESP32 Integration Implementation Plan

> **For agentic workers:** Execute this plan task-by-task with tests and commits.

**Goal:** Finish the PC Go sender before ESP32 integration, with real cross-platform metrics, reliable configuration, transport, tests, and Windows packaging.

**Architecture:** Keep a small `Collector` interface producing CPU, RAM, and GPU-memory percentages. Use `gopsutil` for CPU/RAM, `nvidia-smi` for NVIDIA discrete GPU memory on Windows/Linux, and unified-memory percentage on Apple Silicon. A configuration-driven UDP sender validates frames, handles transient errors, and emits structured logs.

**Tech Stack:** Go 1.24, gopsutil/v4, standard library UDP/JSON, `nvidia-smi` runtime interface, Go tests, Windows amd64 build.

**Spec:** `需求文档.md`, `设计文档.md`

## Global Constraints

- The executable must run by double-clicking and read `config.json` beside itself.
- ESP32 receives JSON over UDP; default endpoint is `192.168.4.2:9000`.
- `cpu`, `mem`, and `gpu` are percentages in `[0,100]`; `gpu` means discrete VRAM percentage on NVIDIA and unified-memory percentage on Apple Silicon.
- No administrator/root permission may be required for normal operation.
- No custom hardware or ESP32 changes are part of this plan.

### Task 1: Replace platform sampling with tested gopsutil collector

**Files:** Modify `collector/metrics.go`; Create `collector/system_test.go`.

- [ ] Test `SystemCollector.Collect` returns bounded CPU and memory values and propagates dependency errors through an injectable function.
- [ ] Keep `Collector` and `Snapshot` stable; use `cpu.Percent` and `mem.VirtualMemory` with a 100ms sample interval.
- [ ] Run `GOCACHE=/tmp/monitor-go-cache go test ./collector -v`.
- [ ] Commit `feat: use gopsutil for system metrics`.

### Task 2: Harden NVIDIA selection and parsing

**Files:** Modify `collector/nvidia.go`, `collector/select.go`; Create `collector/nvidia_test.go`.

- [ ] Extract CSV parsing into a pure function and test valid rows, malformed rows, multiple-GPU first-row behavior, and out-of-range values.
- [ ] Preserve system RAM while assigning NVIDIA VRAM percentage only to `Snapshot.GPU`.
- [ ] Make `HasNvidia` detection non-blocking and ensure missing `nvidia-smi` cleanly falls back.
- [ ] Run `go test ./collector -v` and cross-build Windows/Linux.
- [ ] Commit `feat: harden nvidia collector selection`.

### Task 3: Configuration validation and runtime logging

**Files:** Modify `config/config.go`, `cmd/monitor-sender/main.go`; Create `config/config_validation_test.go`.

- [ ] Validate endpoint host/port, interval range `10ms..60s`, and reject malformed JSON with a clear error.
- [ ] Keep first-run `config.json` creation behavior; add `log_level` and optional `protocol` (`udp` only for this milestone).
- [ ] Write `monitor-sender.log` beside the executable using standard `log` package and continue after transient send/collect errors.
- [ ] Run config tests and verify first-run behavior in a temporary directory.
- [ ] Commit `feat: validate config and add runtime logging`.

### Task 4: Transport and protocol reliability

**Files:** Create `transport/udp.go`, `transport/udp_test.go`; Modify `protocol/protocol.go`, `protocol/protocol_test.go`.

- [ ] Add a `Sender` interface and UDP implementation with write deadlines and injectable `net.Conn` for tests.
- [ ] Validate frame version, finite numeric values, and percentage bounds before encoding.
- [ ] Test loopback UDP delivery, invalid frames, and write errors without external hardware.
- [ ] Commit `feat: add tested udp transport`.

### Task 5: Windows packaging and operator documentation

**Files:** Modify `README.md`; Create `build-windows.ps1`, `config.example.json`.

- [ ] Provide a PowerShell build script producing `dist/monitor-sender-windows-amd64.exe` and copying `config.example.json`.
- [ ] Document double-click startup, first-run config creation, NVIDIA driver prerequisite, log location, and platform GPU semantics.
- [ ] Run the script or equivalent cross-build and verify the output exists.
- [ ] Commit `docs: document windows packaging and operations`.

### Task 6: Final pre-integration verification

- [ ] Run `gofmt`, `go vet ./...`, `GOCACHE=/tmp/monitor-go-cache go test ./...`.
- [ ] Cross-build Windows amd64 and Linux amd64 binaries.
- [ ] Run a local UDP receiver test against the sender for at least five frames.
- [ ] Check `git status` is clean and record the final commit.
