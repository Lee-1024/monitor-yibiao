# monitor-yibiao

PC 端 Go 采集发送程序：通过 UDP JSON 向 ESP32 发送 CPU、内存、显存使用率。

Windows 使用：

```bash
将编译出的 `monitor-sender.exe` 双击运行。首次运行会在 exe 同目录生成 `config.json`，修改后再次双击即可。
```

配置示例：`{"esp32_address":"192.168.4.2:9000","interval_ms":200}`。Windows/Linux 启动时会检测 `nvidia-smi`，发现 NVIDIA 独显后读取 GPU 利用率和独显显存，并优先使用该数据；未发现时回退到基础采集器。Apple Silicon 和非 NVIDIA 平台采集器接口已预留，后续可接入系统专用 API。
