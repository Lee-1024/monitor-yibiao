package collector

import "runtime"

func New() Collector {
	c, _ := NewWithBackend()
	return c
}

func NewWithBackend() (Collector, string) {
	if runtime.GOOS == "darwin" {
		return CombinedCollector{System: SystemCollector{}, GPU: NewPlatformGPU()}, "apple-powermetrics"
	}
	if runtime.GOOS == "windows" || runtime.GOOS == "linux" {
		if HasNvidia() {
			return CombinedCollector{System: SystemCollector{}, GPU: NvidiaCollector{}}, "nvidia-smi"
		}
	}
	return CombinedCollector{System: SystemCollector{}, GPU: NewPlatformGPU()}, "platform-gpu"
}

type AppleSiliconCollector struct{ System Collector }

func (c AppleSiliconCollector) Collect() (Snapshot, error) {
	s, e := c.System.Collect()
	if e == nil {
		s.GPU = s.Memory
	}
	return s, e
}
