package main

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"math"
	"os"
)

func main() {
	sizes := []int{16, 20, 24, 32, 48, 256}
	images := make([][]byte, 0, len(sizes))
	for _, size := range sizes {
		images = append(images, encodeDIB(drawGauge(size)))
	}
	_ = os.WriteFile("cmd/monitor-tray/assets/tray.ico", encodeICO(sizes, images), 0644)
}

func encodeDIB(img *image.RGBA) []byte {
	w, h := img.Bounds().Dx(), img.Bounds().Dy()
	maskStride := ((w + 31) / 32) * 4
	var out bytes.Buffer
	for _, v := range []uint32{40, uint32(w), uint32(h * 2)} {
		_ = binary.Write(&out, binary.LittleEndian, v)
	}
	_ = binary.Write(&out, binary.LittleEndian, uint16(1))
	_ = binary.Write(&out, binary.LittleEndian, uint16(32))
	_ = binary.Write(&out, binary.LittleEndian, uint32(0))
	_ = binary.Write(&out, binary.LittleEndian, uint32(w*h*4))
	for i := 0; i < 4; i++ {
		_ = binary.Write(&out, binary.LittleEndian, uint32(0))
	}
	for y := h - 1; y >= 0; y-- {
		for x := 0; x < w; x++ {
			c := img.RGBAAt(x, y)
			out.Write([]byte{c.B, c.G, c.R, c.A})
		}
	}
	out.Write(make([]byte, maskStride*h))
	return out.Bytes()
}

func drawGauge(size int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	cx, cy := float64(size-1)/2, float64(size-1)/2
	r := float64(size) * 0.47
	fillCircle(img, cx, cy, r, color.RGBA{26, 32, 36, 255})
	green := color.RGBA{54, 211, 153, 255}
	white := color.RGBA{248, 250, 252, 255}
	for i := 0; i <= 8; i++ {
		a := math.Pi * (1.15 + 0.0875*float64(i))
		inner := r * 0.67
		outer := r * 0.88
		drawLine(img, cx+math.Cos(a)*inner, cy+math.Sin(a)*inner, cx+math.Cos(a)*outer, cy+math.Sin(a)*outer, max(1, size/16), green)
	}
	needle := math.Pi * 1.78
	drawLine(img, cx, cy, cx+math.Cos(needle)*r*0.68, cy+math.Sin(needle)*r*0.68, max(2, size/10), white)
	fillCircle(img, cx, cy, math.Max(1, float64(size)/10), green)
	return img
}

func fillCircle(img *image.RGBA, cx, cy, r float64, c color.RGBA) {
	for y := int(cy - r); y <= int(cy+r); y++ {
		for x := int(cx - r); x <= int(cx+r); x++ {
			if math.Hypot(float64(x)-cx, float64(y)-cy) <= r {
				img.SetRGBA(x, y, c)
			}
		}
	}
}

func drawLine(img *image.RGBA, x0, y0, x1, y1 float64, width int, c color.RGBA) {
	steps := int(math.Max(math.Abs(x1-x0), math.Abs(y1-y0))*2) + 1
	for i := 0; i <= steps; i++ {
		t := float64(i) / float64(steps)
		fillCircle(img, x0+(x1-x0)*t, y0+(y1-y0)*t, float64(width)/2, c)
	}
}

func encodeICO(sizes []int, images [][]byte) []byte {
	var out bytes.Buffer
	_ = binary.Write(&out, binary.LittleEndian, uint16(0))
	_ = binary.Write(&out, binary.LittleEndian, uint16(1))
	_ = binary.Write(&out, binary.LittleEndian, uint16(len(images)))
	offset := 6 + 16*len(images)
	for i, data := range images {
		s := sizes[i]
		if s == 256 {
			out.WriteByte(0)
			out.WriteByte(0)
		} else {
			out.WriteByte(byte(s))
			out.WriteByte(byte(s))
		}
		out.WriteByte(0)
		out.WriteByte(0)
		_ = binary.Write(&out, binary.LittleEndian, uint16(1))
		_ = binary.Write(&out, binary.LittleEndian, uint16(32))
		_ = binary.Write(&out, binary.LittleEndian, uint32(len(data)))
		_ = binary.Write(&out, binary.LittleEndian, uint32(offset))
		offset += len(data)
	}
	for _, data := range images {
		out.Write(data)
	}
	return out.Bytes()
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
