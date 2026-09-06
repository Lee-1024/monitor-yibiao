package collector

import (
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
	"runtime"
	"time"
)

type Snapshot struct{ CPU, Memory, GPU float64 }
type Collector interface{ Collect() (Snapshot, error) }

func ClampPercent(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 100 {
		return 100
	}
	return v
}

// PlatformPolicy defines platform-specific GPU source selection.
type PlatformPolicy struct{ OS, GPUPreference string }

func CurrentPolicy() PlatformPolicy {
	if runtime.GOOS == "darwin" {
		return PlatformPolicy{OS: "darwin", GPUPreference: "apple-silicon-unified"}
	}
	return PlatformPolicy{OS: runtime.GOOS, GPUPreference: "discrete-only-when-present"}
}

type SystemCollector struct{}

func (SystemCollector) Collect() (Snapshot, error) {
	p, e := cpu.Percent(100*time.Millisecond, false)
	if e != nil {
		return Snapshot{}, e
	}
	m, e := mem.VirtualMemory()
	if e != nil {
		return Snapshot{}, e
	}
	var c float64
	if len(p) > 0 {
		c = p[0]
	}
	return Snapshot{CPU: ClampPercent(c), Memory: ClampPercent(m.UsedPercent)}, nil
}

type CombinedCollector struct {
	System Collector
	GPU    Collector
}

func (c CombinedCollector) Collect() (Snapshot, error) {
	s, e := c.System.Collect()
	if e != nil {
		return Snapshot{}, e
	}
	g, e := c.GPU.Collect()
	if e != nil {
		return s, nil
	}
	s.GPU = g.GPU
	return s, nil
}
