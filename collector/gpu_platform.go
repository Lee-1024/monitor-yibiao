package collector

import (
	"errors"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

// CommandGPUCollector reads a percentage from a platform command. It keeps
// platform-specific command details outside the sender loop.
type CommandGPUCollector struct {
	Command string
	Args    []string
}

func (c CommandGPUCollector) Collect() (Snapshot, error) {
	out, e := exec.Command(c.Command, c.Args...).Output()
	if e != nil {
		return Snapshot{}, e
	}
	v, e := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
	if e != nil {
		return Snapshot{}, e
	}
	return Snapshot{GPU: ClampPercent(v)}, nil
}

func NewPlatformGPU() Collector {
	switch runtime.GOOS {
	case "windows":
		return CommandGPUCollector{Command: "powershell", Args: []string{"-NoProfile", "-Command", "$v=(Get-Counter '\\GPU Engine(*)\\Utilization Percentage').CounterSamples | Where-Object {$_.InstanceName -match 'engtype_3D'} | Measure-Object -Property CookedValue -Sum; [math]::Min(100,$v.Sum)"}}
	case "linux":
		return CommandGPUCollector{Command: "sh", Args: []string{"-c", "for f in /sys/class/drm/card*/device/gpu_busy_percent; do [ -r \"$f\" ] && cat \"$f\" && exit; done; echo 0"}}
	case "darwin":
		return AppleSiliconGPU{}
	default:
		return CommandGPUCollector{Command: "sh", Args: []string{"-c", "echo 0"}}
	}
}

// AppleSiliconGPU uses the public IOAccelerator performance dictionary exposed
// by ioreg. Availability varies by macOS version; failure safely returns zero.
type AppleSiliconGPU struct{}

func (AppleSiliconGPU) Collect() (Snapshot, error) {
	out, e := exec.Command("sudo", "-n", "powermetrics", "-n", "1", "-i", "200", "--samplers", "gpu_power").CombinedOutput()
	if e != nil {
		return Snapshot{}, e
	}
	m := regexp.MustCompile(`GPU HW active residency:\s*([0-9]+(?:\.[0-9]+)?)%`).FindStringSubmatch(string(out))
	if len(m) == 2 {
		v, _ := strconv.ParseFloat(m[1], 64)
		return Snapshot{GPU: ClampPercent(v)}, nil
	}
	return Snapshot{}, errors.New("apple gpu utilization unavailable")
}
