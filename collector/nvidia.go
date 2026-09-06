package collector

import (
	"encoding/csv"
	"errors"
	"os/exec"
	"strconv"
	"strings"
)

type NvidiaCollector struct{ Command string }

func (n NvidiaCollector) Collect() (Snapshot, error) {
	cmd := n.Command
	if cmd == "" {
		cmd = "nvidia-smi"
	}
	out, err := exec.Command(cmd, "--query-gpu=utilization.gpu,memory.used,memory.total", "--format=csv,noheader,nounits").Output()
	if err != nil {
		return Snapshot{}, err
	}
	r := csv.NewReader(strings.NewReader(strings.TrimSpace(string(out))))
	row, err := r.Read()
	if err != nil || len(row) < 3 {
		return Snapshot{}, errors.New("invalid nvidia-smi output")
	}
	gpu, e1 := strconv.ParseFloat(strings.TrimSpace(row[0]), 64)
	used, e2 := strconv.ParseFloat(strings.TrimSpace(row[1]), 64)
	total, e3 := strconv.ParseFloat(strings.TrimSpace(row[2]), 64)
	if e1 != nil || e2 != nil || e3 != nil || total <= 0 {
		return Snapshot{}, errors.New("invalid nvidia metrics")
	}
	return Snapshot{GPU: ClampPercent(gpu), Memory: ClampPercent(used / total * 100)}, nil
}

func HasNvidia() bool { return exec.Command("nvidia-smi", "-L").Run() == nil }
