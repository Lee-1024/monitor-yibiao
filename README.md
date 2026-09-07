# monitor-yibiao

PC 端 Go 采集发送程序：通过 UDP JSON 向 ESP32 发送 CPU、内存、显存使用率。

Windows 使用：

标准原理图接线图：[接线图.svg](接线图.svg)

```bash
将编译出的 `monitor-sender.exe` 双击运行。首次运行会在 exe 同目录生成 `config.json`，修改后再次双击即可。
```

配置示例：`{"esp32_address":"192.168.4.2:9000","interval_ms":200}`。CPU 和系统内存由 `gopsutil` 跨平台采集。Windows/Linux 启动时检测 `nvidia-smi`，发现 NVIDIA 独显后，第三路 `gpu` 发送该独显的显存占用率。Apple Silicon 没有独立显存，第三路按统一内存使用率显示。该方案不需要管理员权限。
