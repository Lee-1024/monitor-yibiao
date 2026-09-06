package main

import (
	"fmt"
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
	cfg, created, err := config.LoadOrCreate(filepath.Join(filepath.Dir(exe), "config.json"))
	if err != nil {
		panic(err)
	}
	if created {
		fmt.Println("已生成 config.json，请修改后重新运行")
		return
	}
	conn, err := transport.DialUDP(cfg.ESP32Address, 2*time.Second)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	c := collector.New()
	var seq uint32
	for {
		s, err := c.Collect()
		if err != nil {
			fmt.Println("采集失败:", err)
			time.Sleep(cfg.Interval())
			continue
		}
		seq++
		f := protocol.Frame{Version: 1, Sequence: seq, Timestamp: time.Now().Unix(), CPU: s.CPU, Memory: s.Memory, GPU: s.GPU}
		b, _ := protocol.Encode(f)
		if err := conn.Send(b); err != nil {
			fmt.Println(err)
		}
		time.Sleep(cfg.Interval())
	}
}
