package config

import "testing"

func TestInvalidConfig(t *testing.T) {
	if err := validate(Config{ESP32Address: "bad", IntervalMS: 200}); err == nil {
		t.Fatal("expected invalid address")
	}
}
