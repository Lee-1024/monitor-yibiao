package config

import (
	"encoding/json"
	"os"
	"time"
)

type Config struct {
	ESP32Address string `json:"esp32_address"`
	IntervalMS   int    `json:"interval_ms"`
}

func Defaults() Config { return Config{ESP32Address: "192.168.4.2:9000", IntervalMS: 200} }
func LoadOrCreate(path string) (Config, bool, error) {
	b, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		c := Defaults()
		out, _ := json.MarshalIndent(c, "", "  ")
		out = append(out, '\n')
		if err = os.WriteFile(path, out, 0644); err != nil {
			return Config{}, false, err
		}
		return c, true, nil
	}
	if err != nil {
		return Config{}, false, err
	}
	c := Defaults()
	if err = json.Unmarshal(b, &c); err != nil {
		return Config{}, false, err
	}
	if c.ESP32Address == "" {
		c.ESP32Address = Defaults().ESP32Address
	}
	if c.IntervalMS < 10 {
		c.IntervalMS = 200
	}
	return c, false, nil
}
func (c Config) Interval() time.Duration { return time.Duration(c.IntervalMS) * time.Millisecond }
