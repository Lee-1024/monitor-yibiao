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
