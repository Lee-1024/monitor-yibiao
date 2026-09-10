package main

import (
	"github.com/getlantern/systray"
	"log"
	"monitor-yibiao/collector"
	"monitor-yibiao/config"
	"monitor-yibiao/protocol"
	"monitor-yibiao/transport"
	"os"
	"path/filepath"
	"time"
)

func main() { systray.Run(onReady, onExit) }
func onReady() {
	systray.SetTitle("Monitor")
	systray.SetTooltip("CPU / Memory / GPU monitor")
	quit := systray.AddMenuItem("退出", "Exit")
	go runSender()
	go func() { <-quit.ClickedCh; systray.Quit() }()
}
func onExit() {}
func runSender() {
	exe, e := os.Executable()
	if e != nil {
		return
	}
	dir := filepath.Dir(exe)
	f, e := os.OpenFile(filepath.Join(dir, "monitor-sender.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if e == nil {
		defer f.Close()
		log.SetOutput(f)
	}
	cfg, created, e := config.LoadOrCreate(filepath.Join(dir, "config.json"))
	if e != nil || created {
		return
	}
	conn, e := transport.DialUDP(cfg.ESP32Address, 2*time.Second)
	if e != nil {
		return
	}
	defer conn.Close()
	c, backend := collector.NewWithBackend()
	log.Printf("GPU backend: %s", backend)
	var seq uint32
	for {
		var s collector.Snapshot
		if cfg.TestMode {
			s = collector.Snapshot{CPU: collector.ClampPercent(cfg.TestCPU), Memory: collector.ClampPercent(cfg.TestMemory), GPU: collector.ClampPercent(cfg.TestGPU)}
		} else {
			s, e = c.Collect()
			if e != nil {
				log.Println(e)
				time.Sleep(cfg.Interval())
				continue
			}
		}
		seq++
		b, _ := protocol.Encode(protocol.Frame{Version: 1, Sequence: seq, Timestamp: time.Now().Unix(), CPU: s.CPU, Memory: s.Memory, GPU: s.GPU})
		if e = conn.Send(b); e != nil {
			log.Println(e)
		}
		time.Sleep(cfg.Interval())
	}
}
