package collector

import "runtime"

func New() Collector {
	if runtime.GOOS == "darwin" {
		return AppleSiliconCollector{System: SystemCollector{}}
	}
	if runtime.GOOS == "windows" || runtime.GOOS == "linux" {
		if HasNvidia() {
			return CombinedCollector{System: SystemCollector{}, GPU: NvidiaCollector{}}
		}
	}
	return SystemCollector{}
}

type AppleSiliconCollector struct{ System Collector }

func (c AppleSiliconCollector) Collect() (Snapshot, error) {
	s, e := c.System.Collect()
	if e == nil {
		s.GPU = s.Memory
	}
	return s, e
}
