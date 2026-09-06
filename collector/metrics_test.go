package collector

import "testing"

func TestClampPercent(t *testing.T) {
	for _, tc := range []struct{ in, want float64 }{{-1, 0}, {25.5, 25.5}, {101, 100}} {
		if got := ClampPercent(tc.in); got != tc.want {
			t.Fatalf("ClampPercent(%v)=%v, want %v", tc.in, got, tc.want)
		}
	}
}
