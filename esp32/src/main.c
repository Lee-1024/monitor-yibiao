#include <math.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/socket.h>
#include <netinet/in.h>
#include "freertos/FreeRTOS.h"
#include "freertos/task.h"
#include "esp_log.h"
#include "esp_wifi.h"
#include "esp_netif.h"
#include "esp_event.h"
#include "esp_netif_ip_addr.h"
#include "nvs_flash.h"
#include "driver/i2c_master.h"
#include "driver/gpio.h"
#include "esp_http_server.h"

#define UDP_PORT 9000
#define SDA_GPIO 21
#define SCL_GPIO 22
#define PCA_ADDR_DEFAULT 0x70
#define DAC_ADDR 0x60
#define DAC_MAX_CPU_CODE 652
#define DAC_MAX_MEMORY_CODE 635
#define DAC_MAX_GPU_CODE 635
#define STALE_MS 3000
#ifndef BYPASS_PCA
#define BYPASS_PCA 0
#endif
#define AP_SSID "Monitor-Setup"
#define AP_PASSWORD "12345678"
static const char *TAG = "monitor";
static TickType_t last_frame;
static uint8_t pca_addr = PCA_ADDR_DEFAULT;
static uint8_t active_channels;
static volatile bool data_online;
static SemaphoreHandle_t dac_mutex;
static i2c_master_bus_handle_t i2c_bus;
#if !BYPASS_PCA
static i2c_master_dev_handle_t pca_dev;
#endif
static i2c_master_dev_handle_t dac_dev;
static esp_netif_t *sta_netif;
static httpd_handle_t http_server;
static int json_number(const char *json, const char *key, double *value){
    char needle[32];

    snprintf(needle, sizeof(needle), "\"%s\"", key);
    const char *p = strstr(json, needle);

    if (!p)
        return 0;
    p = strchr(p, ':');
    if (!p)
        return 0;
    char *end;

    *value = strtod(p + 1, &end);
    return end != p + 1;
}
static esp_err_t wr(i2c_master_dev_handle_t dev, const uint8_t * d, size_t n){
    return i2c_master_transmit(dev, d, n, 100);
}
#if !BYPASS_PCA
static esp_err_t pca_select_checked(uint8_t ch) {
    uint8_t s = 1u << ch, r = 0;
    esp_err_t e = wr(pca_dev, &s, 1);

    if (e != ESP_OK)
        return e;
    vTaskDelay(pdMS_TO_TICKS(2));
    e = i2c_master_receive(pca_dev, &r, 1, 100);
    if (e == ESP_OK && r != s) {
        ESP_LOGW(TAG, "PCA control readback mismatch: requested 0x%02X got 0x%02X", s, r);
        return ESP_ERR_INVALID_RESPONSE;
    } return e;
}
#endif
static esp_err_t probe(uint8_t a) {
    return i2c_master_probe(i2c_bus, a, 50);
}
static void scan_i2c(void){
    int found = 0;

    ESP_LOGI(TAG, "I2C scan GPIO%d(SDA), GPIO%d(SCL), levels SDA=%d SCL=%d", SDA_GPIO, SCL_GPIO, gpio_get_level(SDA_GPIO), gpio_get_level(SCL_GPIO));
    for (uint8_t a = 1; a < 0x7f; a++) {
        if (probe(a) == ESP_OK) {
            ESP_LOGI(TAG, "I2C found 0x%02X", a);
            found++;
            if (a >= 0x70 && a <= 0x77) {
                pca_addr = a;
                ESP_LOGI(TAG, "PCA address selected: 0x%02X", pca_addr);
            }
        }
    } if (!found)
        ESP_LOGE(TAG, "I2C found no devices; check VIN/GND/SDA/SCL and PCA RST=3V3");
}
static esp_err_t dac(uint8_t ch, double p){
    esp_err_t e = ESP_OK;

    if (dac_mutex && xSemaphoreTake(dac_mutex, pdMS_TO_TICKS(200)) != pdTRUE)
        return ESP_ERR_TIMEOUT;
#if !BYPASS_PCA
    e = pca_select_checked(ch);
    if (e != ESP_OK) {
        ESP_LOGE(TAG, "PCA 0x%02X select CH%d: %s", pca_addr, ch, esp_err_to_name(e));
        if (dac_mutex)
            xSemaphoreGive(dac_mutex);
        return e;
    }
#else
    if (ch != 0) {
        if (dac_mutex)
            xSemaphoreGive(dac_mutex);
        return ESP_OK;
    }
#endif
    if (!isfinite(p) || p < 0) {
        p = 0;
    } if (p > 100) {
        p = 100;
    } uint16_t max_code = (ch == 0) ? DAC_MAX_CPU_CODE : ((ch == 1) ? DAC_MAX_MEMORY_CODE : DAC_MAX_GPU_CODE);
    uint16_t v = (uint16_t) lround(p * max_code / 100.0);

    if (v > 4095)
        v = 4095;
    uint8_t b[3] = {0x40, (uint8_t) (v >> 4), (uint8_t) ((v & 15) << 4)};

    for (int attempt = 0; attempt < 3; attempt++) {
        e = wr(dac_dev, b, 3);
        if (e == ESP_OK)
            break;
        vTaskDelay(pdMS_TO_TICKS(2));
    } if (e != ESP_OK)
        ESP_LOGE(TAG, "MCP 0x60 on CH%d: %s", ch, esp_err_to_name(e));
    if (dac_mutex)
        xSemaphoreGive(dac_mutex);
    return e;
}
static void configure_fixed_channels(void){
    active_channels = 0x07;
    ESP_LOGI(TAG, "fixed DAC channel mask: CH0 CPU, CH1 memory, CH2 GPU");
}
static void
zero(void)
{
#if BYPASS_PCA
    esp_err_t e = dac(0, 0);

    ESP_LOGI(TAG, "DIRECT MCP CH0 zero: %s", esp_err_to_name(e));
#else
    for (int i = 0; i < 3; i++) {
        if (active_channels & (1u << i)) {
            esp_err_t e = dac(i, 0);

            ESP_LOGI(TAG, "DAC CH%d zero: %s", i, esp_err_to_name(e));
        }
    }
#endif
}
static char wifi_ssid[33];
static char wifi_password[65];

static void load_wifi_config(void)
{
    nvs_handle_t h;
    if (nvs_open("wifi", NVS_READONLY, &h) == ESP_OK) {
        size_t n = sizeof(wifi_ssid);
        nvs_get_str(h, "ssid", wifi_ssid, &n);
        n = sizeof(wifi_password);
        nvs_get_str(h, "password", wifi_password, &n);
        nvs_close(h);
    }
}

static void on_wifi(void *a, esp_event_base_t b, int32_t id, void *d){
    if (b == WIFI_EVENT && id == WIFI_EVENT_STA_START) {
        if (wifi_ssid[0]) { ESP_LOGI(TAG, "connecting WiFi: %s", wifi_ssid); esp_wifi_connect(); }
    } if (b == WIFI_EVENT && id == WIFI_EVENT_STA_DISCONNECTED) {
        ESP_LOGW(TAG, "WiFi disconnected, retrying");
        zero();
        if (wifi_ssid[0]) esp_wifi_connect();
    } if (b == IP_EVENT && id == IP_EVENT_STA_GOT_IP) {
        ip_event_got_ip_t *e = (ip_event_got_ip_t *) d;

        ESP_LOGI(TAG, "IP: " IPSTR, IP2STR(&e->ip_info.ip));
    }
}
static esp_err_t portal_get(httpd_req_t *req){char ip[16]="未连接";if(sta_netif){esp_netif_ip_info_t info;if(esp_netif_get_ip_info(sta_netif,&info)==ESP_OK&&info.ip.addr)snprintf(ip,sizeof(ip),IPSTR,IP2STR(&info.ip));}char body[1000];int n=snprintf(body,sizeof(body),"<html><meta charset='utf-8'><h2>ESP32 Monitor</h2><p>配置热点: %s</p><p>当前内网 IP: <b>%s</b></p><p>Go 地址: <b>%s:9000</b></p><form action='/save' method='get'>WiFi名称: <input name='ssid' required><br>WiFi密码: <input name='password' type='password'><br><button type='submit'>保存并重启</button></form></html>",AP_SSID,ip,ip);httpd_resp_set_type(req,"text/html; charset=utf-8");return httpd_resp_send(req,body,n);}
static esp_err_t portal_save(httpd_req_t *req){char q[256]={0},ssid[33]={0},password[65]={0};if(httpd_req_get_url_query_str(req,q,sizeof(q))!=ESP_OK)return httpd_resp_send_err(req,HTTPD_400_BAD_REQUEST,"missing query");if(httpd_query_key_value(q,"ssid",ssid,sizeof(ssid))!=ESP_OK||!ssid[0])return httpd_resp_send_err(req,HTTPD_400_BAD_REQUEST,"ssid required");httpd_query_key_value(q,"password",password,sizeof(password));nvs_handle_t h;if(nvs_open("wifi",NVS_READWRITE,&h)!=ESP_OK)return httpd_resp_send_err(req,HTTPD_500_INTERNAL_SERVER_ERROR,"nvs open failed");esp_err_t e=nvs_set_str(h,"ssid",ssid);if(e==ESP_OK)e=nvs_set_str(h,"password",password);if(e==ESP_OK)e=nvs_commit(h);nvs_close(h);if(e!=ESP_OK)return httpd_resp_send_err(req,HTTPD_500_INTERNAL_SERVER_ERROR,"nvs save failed");httpd_resp_sendstr(req,"已保存 WiFi 配置，ESP32 即将重启");vTaskDelay(pdMS_TO_TICKS(500));esp_restart();return ESP_OK;}
static void portal_start(void){httpd_config_t cfg=HTTPD_DEFAULT_CONFIG();if(httpd_start(&http_server,&cfg)==ESP_OK){httpd_uri_t uri={.uri="/",.method=HTTP_GET,.handler=portal_get};httpd_uri_t save={.uri="/save",.method=HTTP_GET,.handler=portal_save};httpd_register_uri_handler(http_server,&uri);httpd_register_uri_handler(http_server,&save);}}
static void wifi_start(void){
    load_wifi_config();
    esp_netif_init();
    esp_event_loop_create_default();
    sta_netif=esp_netif_create_default_wifi_sta();esp_netif_create_default_wifi_ap();
    wifi_init_config_t x = WIFI_INIT_CONFIG_DEFAULT();

    esp_wifi_init(&x);
    esp_event_handler_register(WIFI_EVENT, ESP_EVENT_ANY_ID, on_wifi, 0);
    esp_event_handler_register(IP_EVENT, IP_EVENT_STA_GOT_IP, on_wifi, 0);
    wifi_config_t c = {0};

    strncpy((char *)c.sta.ssid, wifi_ssid, sizeof(c.sta.ssid));
    strncpy((char *)c.sta.password, wifi_password, sizeof(c.sta.password));
    wifi_config_t ap={0};strncpy((char*)ap.ap.ssid,AP_SSID,sizeof(ap.ap.ssid));strncpy((char*)ap.ap.password,AP_PASSWORD,sizeof(ap.ap.password));ap.ap.ssid_len=strlen(AP_SSID);ap.ap.authmode=WIFI_AUTH_WPA2_PSK;ap.ap.max_connection=4;
    esp_wifi_set_mode(WIFI_MODE_APSTA);
    esp_wifi_set_config(WIFI_IF_STA, &c);
    esp_wifi_set_config(WIFI_IF_AP, &ap);
    esp_wifi_start();
    portal_start();ESP_LOGI(TAG,"setup AP: %s password: %s portal: http://192.168.4.1",AP_SSID,AP_PASSWORD);
}
static void udp_task(void *a){
    struct sockaddr_in x = {.sin_family = AF_INET,.sin_port = htons(UDP_PORT),.sin_addr.s_addr = htonl(INADDR_ANY)};
    int s = socket(AF_INET, SOCK_DGRAM, IPPROTO_IP);

    bind(s, (struct sockaddr *)&x, sizeof(x));
    char b[512];
    unsigned frames = 0;

    for (;;) {
        int n = recv(s, b, sizeof(b) - 1, 0);

        if (n <= 0)
            continue;
        b[n] = 0;
        double values[3];

        if (json_number(b, "cpu", &values[0]) && json_number(b, "mem", &values[1]) && json_number(b, "gpu", &values[2])) {
            bool any_ok = false;

            for (uint8_t ch = 0; ch < 3; ch++) {
                if ((active_channels & (1u << ch)) && dac(ch, values[ch]) == ESP_OK)
                    any_ok = true;
            } if (any_ok) {
                last_frame = xTaskGetTickCount();
                if (!data_online) {
                    data_online = true;
                    ESP_LOGI(TAG, "DATA ONLINE cpu=%.1f mem=%.1f gpu=%.1f", values[0], values[1], values[2]);
                } if (++frames % 25 == 0)
                    ESP_LOGI(TAG, "RX cpu=%.1f mem=%.1f gpu=%.1f", values[0], values[1], values[2]);
            }
        }
    }
}
static void safety_task(void *a){
    for (;;) {
        if (data_online && xTaskGetTickCount() - last_frame > pdMS_TO_TICKS(STALE_MS)) {
            ESP_LOGW(TAG, "DATA STALE: zeroing active outputs");
            zero();
            data_online = false;
        } vTaskDelay(pdMS_TO_TICKS(250));
    }
}
void app_main(void){
    nvs_flash_init();
    dac_mutex = xSemaphoreCreateMutex();
    i2c_master_bus_config_t b = {.i2c_port = I2C_NUM_0,.sda_io_num = SDA_GPIO,.scl_io_num = SCL_GPIO,.clk_source = I2C_CLK_SRC_DEFAULT,.glitch_ignore_cnt = 7,.flags.enable_internal_pullup = true};

    ESP_ERROR_CHECK(i2c_new_master_bus(&b, &i2c_bus));
    scan_i2c();
#if !BYPASS_PCA
    i2c_device_config_t p = {.dev_addr_length = I2C_ADDR_BIT_LEN_7,.device_address = pca_addr,.scl_speed_hz = 100000};

    ESP_ERROR_CHECK(i2c_master_bus_add_device(i2c_bus, &p, &pca_dev));
#else
    ESP_LOGW(TAG, "DIRECT MCP DIAGNOSTIC MODE: PCA bypassed, CPU channel only");
#endif
    i2c_device_config_t d = {.dev_addr_length = I2C_ADDR_BIT_LEN_7,.device_address = DAC_ADDR,.scl_speed_hz = 100000};

    ESP_ERROR_CHECK(i2c_master_bus_add_device(i2c_bus, &d, &dac_dev));
    configure_fixed_channels();
    zero();
    wifi_start();
    xTaskCreate(udp_task, "udp", 4096, 0, 5, 0);
    xTaskCreate(safety_task, "safety", 2048, 0, 6, 0);
    ESP_LOGI(TAG, "started");
}
