package collector

import "runtime"

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

// BasicCollector is portable; platform collectors supply CPU/GPU metrics.
type BasicCollector struct{}

func (BasicCollector) Collect() (Snapshot, error) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return Snapshot{Memory: ClampPercent(float64(m.Alloc) / float64(max(m.Sys, 1)) * 100)}, nil
}
