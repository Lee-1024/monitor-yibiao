package collector

import (
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type CPUReader struct {
	prevIdle, prevTotal uint64
	initialized         bool
}

func (r *CPUReader) Read() float64 {
	if runtime.GOOS == "linux" {
		b, e := os.ReadFile("/proc/stat")
		if e == nil {
			for _, l := range strings.Split(string(b), "\n") {
				if strings.HasPrefix(l, "cpu ") {
					f := strings.Fields(l)
					var total, idle uint64
					for i := 1; i < len(f); i++ {
						v, _ := strconv.ParseUint(f[i], 10, 64)
						total += v
						if i == 4 {
							idle = v
						}
					}
					if r.initialized {
						dt := total - r.prevTotal
						di := idle - r.prevIdle
						r.prevTotal, total = r.prevTotal+dt, total
						r.prevIdle = idle
						if dt > 0 {
							return ClampPercent(float64(dt-di) / float64(dt) * 100)
						}
					}
					r.prevTotal, r.prevIdle = total, idle
					r.initialized = true
					return 0
				}
			}
		}
	}
	if runtime.GOOS == "windows" {
		out, e := exec.Command("powershell", "-NoProfile", "-Command", "(Get-Counter '\\Processor(_Total)\\% Processor Time').CounterSamples[0].CookedValue").Output()
		if e == nil {
			v, _ := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
			return ClampPercent(v)
		}
	}
	return 0
}

type SystemCollector struct{ cpu CPUReader }

func (c *SystemCollector) Collect() (Snapshot, error) {
	s := BasicCollector{}
	out, e := s.Collect()
	out.CPU = c.cpu.Read()
	if runtime.GOOS == "linux" {
		time.Sleep(10 * time.Millisecond)
		out.CPU = c.cpu.Read()
	}
	return out, e
}
