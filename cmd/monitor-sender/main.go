package main

import (
	"fmt"
	"log"
	"monitor-yibiao/collector"
	"monitor-yibiao/config"
	"monitor-yibiao/protocol"
	"monitor-yibiao/transport"
	"os"
	"path/filepath"
	"time"
)

func main() {
	exe, err := os.Executable()
	if err != nil {
		panic(err)
	}
	logFile, err := os.OpenFile(filepath.Join(filepath.Dir(exe), "monitor-sender.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err == nil {
		defer logFile.Close()
		log.SetOutput(logFile)
	}
	cfg, created, err := config.LoadOrCreate(filepath.Join(filepath.Dir(exe), "config.json"))
	if err != nil {
		panic(err)
	}
	if created {
		fmt.Println("已生成 config.json，请修改后重新运行")
		return
	}
	log.Printf("config loaded: esp32=%s interval=%s test_mode=%t test_cpu=%.1f test_memory=%.1f", cfg.ESP32Address, cfg.Interval(), cfg.TestMode, cfg.TestCPU, cfg.TestMemory)
	conn, err := transport.DialUDP(cfg.ESP32Address, 2*time.Second)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	c, backend := collector.NewWithBackend()
	log.Printf("GPU backend: %s", backend)
	var seq uint32
	for {
		var s collector.Snapshot
		var err error
		if cfg.TestMode {
			s = collector.Snapshot{CPU: collector.ClampPercent(cfg.TestCPU), Memory: collector.ClampPercent(cfg.TestMemory), GPU: collector.ClampPercent(cfg.TestGPU)}
		} else {
			s, err = c.Collect()
		}
		if err != nil {
			log.Println("采集失败:", err)
			time.Sleep(cfg.Interval())
			continue
		}
		seq++
		f := protocol.Frame{Version: 1, Sequence: seq, Timestamp: time.Now().Unix(), CPU: s.CPU, Memory: s.Memory, GPU: s.GPU}
		b, _ := protocol.Encode(f)
		if err := conn.Send(b); err != nil {
			log.Println(err)
		}
		time.Sleep(cfg.Interval())
	}
}
