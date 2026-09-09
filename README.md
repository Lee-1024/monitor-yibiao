# monitor-yibiao

PC 端 Go 采集程序通过 Wi-Fi/UDP 向 ESP32 发送 CPU、内存和 GPU 指标，ESP32 通过 PCA9548A 和三个 MCP4725 驱动三只 100uA 指针表。

## 目录

```text
cmd/monitor-sender/   普通发送程序
cmd/monitor-tray/     Windows 托盘发送程序
collector/            CPU、内存、GPU 采集后端
config/               配置加载和校验
protocol/             UDP JSON 数据帧
transport/            UDP 发送封装
esp32/                PlatformIO + ESP-IDF 固件
```

## Go 配置

程序读取可执行文件同目录的 `config.json`。首次运行会生成配置文件并退出，填写后再次运行。

```json
{
  "esp32_address": "192.168.21.17:9000",
  "interval_ms": 200,
  "test_mode": false,
  "test_cpu": 50,
  "test_memory": 50,
  "test_gpu": 50
}
```

`esp32_address` 必须是 `IP:端口`；`interval_ms` 范围为 10 到 60000；`test_mode=true` 时使用三个测试值，否则采集真实数据。日志写入同目录的 `monitor-sender.log`。

## macOS

```bash
chmod +x build-macos.sh
./build-macos.sh
cd dist/monitor-sender-macos
cp config.example.json config.json
./monitor-sender
```

Apple Silicon 的 GPU 指标来自 IOAccelerator；统一内存没有独立 VRAM。

## Windows

在 PowerShell 中执行：

```powershell
.\build-windows.ps1
```

生成 `dist/monitor-sender-windows-amd64.exe` 和 `dist/config.example.json`。将配置文件复制为 `config.json` 放在 exe 同目录，双击 exe 后程序运行在系统托盘，右键托盘图标退出。NVIDIA 优先使用 `nvidia-smi`，否则读取 Windows GPU Engine 性能计数器。

## Linux

```bash
go build -trimpath -ldflags "-s -w" -o monitor-sender ./cmd/monitor-sender
cp config.example.json config.json
./monitor-sender
```

Linux NVIDIA 优先使用 `nvidia-smi`，AMD/Intel 使用 DRM/sysfs GPU 忙碌率；没有可用接口时 `gpu` 为 0。

## ESP32

PlatformIO 工程目标为 `esp32dev`。VS Code 扩展安装后，`pio` 可能不在系统 PATH，可使用绝对路径：

```bash
PIO="$HOME/.platformio/penv/bin/pio"
cd esp32
"$PIO" run -e esp32dev
"$PIO" run -e esp32dev -t upload --upload-port /dev/cu.usbserial-230
"$PIO" device monitor --port /dev/cu.usbserial-230 --baud 115200
```

首次配网连接 `Monitor-Setup`，密码 `12345678`，访问 `http://192.168.4.1`，保存目标 Wi-Fi 后页面显示 STA IP。ESP32 监听 UDP `9000`，数据格式为：

```json
{"v":1,"seq":1,"ts":1720000000,"cpu":25.5,"mem":60,"gpu":10}
```

## 验证

```bash
GOCACHE=/tmp/monitor-go-cache go test ./...
go vet ./...
```

测试仪表时将 `test_mode` 设为 `true`，修改 `test_cpu`、`test_memory` 或 `test_gpu` 后重启发送程序。
