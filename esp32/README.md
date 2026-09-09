# ESP32 固件

本目录是 PlatformIO + ESP-IDF 工程，目标板为 `esp32dev`。

## 使用前配置

首次启动连接 `Monitor-Setup` 热点，在 `http://192.168.4.1` 页面填写目标 WiFi；配置保存到 NVS，无需修改源码。

## 固定接线

```text
GPIO21 -> PCA9548A SDA
GPIO22 -> PCA9548A SCL
3V3    -> PCA9548A VIN
GND    -> PCA9548A GND
PCA A0/A1/A2 -> 公共 GND（地址 0x70）
PCA RST -> 不接线（当前模块板载上拉，悬空实测 3.3V）
PCA SD0/SC0 -> CPU MCP4725 SDA/SCL
PCA SD1/SC1 -> 内存 MCP4725 SDA/SCL
PCA SD2/SC2 -> 显存 MCP4725 SDA/SCL
```

三个 MCP4725 地址均为 `0x60`，通过 PCA9548A 通道隔离。每只表独立校准：CPU 通道码值 390，内存通道码值 650；显存通道待单独测量后设置。

## AP 配网

ESP32 会开启配置热点 `Monitor-Setup`，密码 `12345678`。连接该热点后访问 `http://192.168.4.1`，填写目标 WiFi 名称和密码，点击“保存并重启”。配置写入 NVS，重启后自动使用新 WiFi；配置页面会显示当前 STA 内网 IP，Go 端填入该 IP 的 `:9000`。首次启动没有保存配置时只开启配置热点，不包含默认 WiFi 密码。

## 接收协议

UDP 端口 `9000`，JSON 示例：

```json
{"v":1,"seq":1,"ts":1720000000,"cpu":25.5,"mem":60,"gpu":10}
```

启动时三个输出为零；WiFi 断开或连续 3 秒没有合法数据时，三个输出自动清零。

CPU 万用表测试：Go 端配置 `test_mode: true`。依次设置 0、25、50、75、90、100，预计指针接近对应刻度；最终以表头位置而不是线圈电阻测量值校准。
