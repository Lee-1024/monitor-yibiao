package collector

import "testing"

func TestParseNvidiaRow(t *testing.T) {
	v, e := parseNvidiaRow([]string{"40", "2048", "8192"})
	if e != nil || v != 40 {
		t.Fatalf("got %v %v", v, e)
	}
	if _, e = parseNvidiaRow([]string{"x"}); e == nil {
		t.Fatal("expected error")
	}
}
