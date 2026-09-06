package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadOrCreate(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "config.json")
	c, created, err := LoadOrCreate(p)
	if err != nil || !created || c.ESP32Address == "" {
		t.Fatalf("first load: %#v %v %v", c, created, err)
	}
	c, created, err = LoadOrCreate(p)
	if err != nil || created {
		t.Fatalf("second load: %#v %v", c, created)
	}
	_ = os.Remove(p)
}
