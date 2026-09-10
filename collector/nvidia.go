package collector

import (
	"encoding/csv"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

type NvidiaCollector struct{ Command string }

func nvidiaCommand() string {
	if path, err := exec.LookPath("nvidia-smi"); err == nil {
		return path
	}
	for _, path := range []string{
		filepath.Join(os.Getenv("SystemRoot"), "System32", "nvidia-smi.exe"),
		filepath.Join(os.Getenv("ProgramW6432"), "NVIDIA Corporation", "NVSMI", "nvidia-smi.exe"),
		filepath.Join(os.Getenv("ProgramFiles"), "NVIDIA Corporation", "NVSMI", "nvidia-smi.exe"),
	} {
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			return path
		}
	}
	return "nvidia-smi"
}

func parseNvidiaRow(row []string) (float64, error) {
	if len(row) < 3 {
		return 0, errors.New("invalid nvidia-smi output")
	}
	used, e1 := strconv.ParseFloat(strings.TrimSpace(row[1]), 64)
	total, e2 := strconv.ParseFloat(strings.TrimSpace(row[2]), 64)
	if e1 != nil || e2 != nil || total <= 0 {
		return 0, errors.New("invalid nvidia metrics")
	}
	return ClampPercent(used / total * 100), nil
}

func (n NvidiaCollector) Collect() (Snapshot, error) {
	cmd := n.Command
	if cmd == "" {
		cmd = nvidiaCommand()
	}
	out, err := commandOutput(cmd, "--query-gpu=utilization.gpu,memory.used,memory.total", "--format=csv,noheader,nounits")
	if err != nil {
		return Snapshot{}, err
	}
	r := csv.NewReader(strings.NewReader(strings.TrimSpace(string(out))))
	row, err := r.Read()
	if err != nil || len(row) < 3 {
		return Snapshot{}, errors.New("invalid nvidia-smi output")
	}
	v, err := parseNvidiaRow(row)
	if err != nil {
		return Snapshot{}, err
	}
	return Snapshot{GPU: v}, nil
}

func HasNvidia() bool { return commandRun(nvidiaCommand(), "-L") == nil }
