package collector

import "runtime"

func New() Collector {
	if runtime.GOOS == "windows" || runtime.GOOS == "linux" {
		if HasNvidia() {
			return NvidiaCollector{}
		}
	}
	return BasicCollector{}
}
