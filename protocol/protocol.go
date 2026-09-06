package protocol

import "encoding/json"

type Frame struct {
	Version   uint8   `json:"v"`
	Sequence  uint32  `json:"seq"`
	Timestamp int64   `json:"ts,omitempty"`
	CPU       float64 `json:"cpu"`
	Memory    float64 `json:"mem"`
	GPU       float64 `json:"gpu"`
}

func (f Frame) Valid() bool {
	return f.Version == 1 && f.CPU >= 0 && f.CPU <= 100 && f.Memory >= 0 && f.Memory <= 100 && f.GPU >= 0 && f.GPU <= 100
}
func Encode(f Frame) ([]byte, error) { return json.Marshal(f) }
