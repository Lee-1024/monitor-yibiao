#include <math.h>
#include <string.h>
#include <sys/socket.h>
#include <netinet/in.h>
#include "freertos/FreeRTOS.h"
#include "freertos/task.h"
#include "esp_log.h"
#include "esp_wifi.h"
#include "esp_netif.h"
#include "esp_event.h"
#include "nvs_flash.h"
#include "driver/i2c.h"
#include "cJSON.h"

#define WIFI_SSID "CHANGE_ME"
#define WIFI_PASSWORD "CHANGE_ME"
#define UDP_PORT 9000
#define SDA_GPIO 21
#define SCL_GPIO 22
#define PCA_ADDR 0x70
#define DAC_ADDR 0x60
#define DAC_MAX_CODE 186
#define STALE_MS 3000
static const char *TAG="monitor"; static TickType_t last_frame;
static esp_err_t wr(uint8_t a,const uint8_t*d,size_t n){i2c_cmd_handle_t c=i2c_cmd_link_create();i2c_master_start(c);i2c_master_write_byte(c,(a<<1)|I2C_MASTER_WRITE,true);i2c_master_write(c,(uint8_t*)d,n,true);i2c_master_stop(c);esp_err_t e=i2c_master_cmd_begin(I2C_NUM_0,c,pdMS_TO_TICKS(100));i2c_cmd_link_delete(c);return e;}
static esp_err_t dac(uint8_t ch,double p){uint8_t s=1u<<ch;if(wr(PCA_ADDR,&s,1)!=ESP_OK)return ESP_FAIL;if(!isfinite(p)||p<0)p=0;if(p>100)p=100;uint16_t v=(uint16_t)lround(p*DAC_MAX_CODE/100.0);uint8_t b[2]={v>>4,(uint8_t)((v&15)<<4)};return wr(DAC_ADDR,b,2);}
static void zero(void){for(int i=0;i<3;i++)dac(i,0);}
static void on_wifi(void*a,esp_event_base_t b,int32_t id,void*d){if(b==WIFI_EVENT&&id==WIFI_EVENT_STA_START)esp_wifi_connect();if(b==WIFI_EVENT&&id==WIFI_EVENT_STA_DISCONNECTED){zero();esp_wifi_connect();}}
static void wifi_start(void){esp_netif_init();esp_event_loop_create_default();esp_netif_create_default_wifi_sta();wifi_init_config_t x=WIFI_INIT_CONFIG_DEFAULT();esp_wifi_init(&x);esp_event_handler_register(WIFI_EVENT,ESP_EVENT_ANY_ID,on_wifi,0);wifi_config_t c={0};strncpy((char*)c.sta.ssid,WIFI_SSID,sizeof(c.sta.ssid));strncpy((char*)c.sta.password,WIFI_PASSWORD,sizeof(c.sta.password));esp_wifi_set_mode(WIFI_MODE_STA);esp_wifi_set_config(WIFI_IF_STA,&c);esp_wifi_start();}
static void udp_task(void*a){struct sockaddr_in x={.sin_family=AF_INET,.sin_port=htons(UDP_PORT),.sin_addr.s_addr=htonl(INADDR_ANY)};int s=socket(AF_INET,SOCK_DGRAM,IPPROTO_IP);bind(s,(struct sockaddr*)&x,sizeof(x));char b[512];for(;;){int n=recv(s,b,sizeof(b)-1,0);if(n<=0)continue;b[n]=0;cJSON*r=cJSON_Parse(b);if(r){cJSON*c=cJSON_GetObjectItem(r,"cpu"),*m=cJSON_GetObjectItem(r,"mem"),*g=cJSON_GetObjectItem(r,"gpu");if(cJSON_IsNumber(c)&&cJSON_IsNumber(m)&&cJSON_IsNumber(g)&&dac(0,c->valuedouble)==ESP_OK&&dac(1,m->valuedouble)==ESP_OK&&dac(2,g->valuedouble)==ESP_OK)last_frame=xTaskGetTickCount();cJSON_Delete(r);}}}
static void safety_task(void*a){for(;;){if(last_frame&&xTaskGetTickCount()-last_frame>pdMS_TO_TICKS(STALE_MS))zero();vTaskDelay(pdMS_TO_TICKS(250));}}
void app_main(void){nvs_flash_init();i2c_config_t i={.mode=I2C_MODE_MASTER,.sda_io_num=SDA_GPIO,.scl_io_num=SCL_GPIO,.sda_pullup_en=GPIO_PULLUP_ENABLE,.scl_pullup_en=GPIO_PULLUP_ENABLE,.master.clk_speed=100000};i2c_param_config(I2C_NUM_0,&i);i2c_driver_install(I2C_NUM_0,i.mode,0,0,0);zero();wifi_start();xTaskCreate(udp_task,"udp",4096,0,5,0);xTaskCreate(safety_task,"safety",2048,0,6,0);ESP_LOGI(TAG,"started");}
