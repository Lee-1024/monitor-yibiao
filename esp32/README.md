# ESP32 固件

本目录是 PlatformIO + ESP-IDF 工程，目标板为 `esp32dev`。

## 使用前修改

编辑 `src/main.c` 顶部：

```c
#define WIFI_SSID "你的WiFi名称"
#define WIFI_PASSWORD "你的WiFi密码"
```

## 固定接线

```text
GPIO21 -> PCA9548A SDA
GPIO22 -> PCA9548A SCL
3V3    -> PCA9548A VIN
GND    -> PCA9548A GND
PCA A0/A1/A2 -> 公共 GND（地址 0x70）
PCA SD0/SC0 -> CPU MCP4725 SDA/SCL
PCA SD1/SC1 -> 内存 MCP4725 SDA/SCL
PCA SD2/SC2 -> 显存 MCP4725 SDA/SCL
```

三个 MCP4725 地址均为 `0x60`，通过 PCA9548A 通道隔离。固件将 `OUT` 限制到约 0.15V 对应的 DAC 码值 186。

## 接收协议

UDP 端口 `9000`，JSON 示例：

```json
{"v":1,"seq":1,"ts":1720000000,"cpu":25.5,"mem":60,"gpu":10}
```

启动时三个输出为零；WiFi 断开或连续 3 秒没有合法数据时，三个输出自动清零。

CPU 万用表测试：Go 端配置 `test_mode: true`、`test_cpu: 50` 时，CPU 对应 MCP4725 `OUT` 理论约 0.075V。依次设置 0、25、50、75、100，可测得约 0、0.0375、0.075、0.1125、0.15V。
