# monitor-yibiao

PC 端 Go 采集发送程序：通过 UDP JSON 向 ESP32 发送 CPU、内存、显存使用率。

Windows 使用：

```bash
将编译出的 `monitor-sender.exe` 双击运行。首次运行会在 exe 同目录生成 `config.json`，修改后再次双击即可。
```

配置示例：`{"esp32_address":"192.168.4.2:9000","interval_ms":200}`。采集器按平台区分：Apple Silicon Mac 使用统一内存和 GPU 活跃度；Windows/Linux 检测到独立显卡时只采集独立显卡，未检测到独显时使用集成显卡。当前 BasicCollector 仅提供可移植的内存占用估算，平台采集器需继续接入。
