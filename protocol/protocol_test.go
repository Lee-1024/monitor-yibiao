package protocol

import (
	"encoding/json"
	"testing"
)

func TestFrameJSON(t *testing.T) {
	f := Frame{Version: 1, Sequence: 7, CPU: 25.5, Memory: 60, GPU: 10}
	b, err := json.Marshal(f)
	if err != nil {
		t.Fatal(err)
	}
	var got Frame
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}
	if got != f {
		t.Fatalf("round trip mismatch: %#v != %#v", got, f)
	}
}
