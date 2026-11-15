package main

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"math"
)

// trayIcon draws a small white chat bubble with an accent dot, in the flat
// white style of the CS2 menu icons.
func trayIcon() []byte {
	const n = 64
	img := image.NewNRGBA(image.Rect(0, 0, n, n))
	// Rounded rectangle body.
	body := func(x, y float64) float64 {
		cx, cy, hw, hh, r := 32.0, 27.0, 26.0, 19.0, 9.0
		dx := math.Max(math.Abs(x-cx)-(hw-r), 0)
		dy := math.Max(math.Abs(y-cy)-(hh-r), 0)
		return math.Hypot(dx, dy) - r
	}
	// Tail triangle at the bottom left.
	tail := func(x, y float64) bool {
		return y >= 44 && y <= 58 && x >= 14 && x <= 30 && (x-14) <= (58-y)*1.15
	}
	dots := [][2]float64{{20, 27}, {32, 27}, {44, 27}}
	for y := 0; y < n; y++ {
		for x := 0; x < n; x++ {
			fx, fy := float64(x)+0.5, float64(y)+0.5
			a := math.Min(math.Max(0.5-body(fx, fy), 0), 1)
			if tail(fx, fy) {
				a = 1
			}
			if a == 0 {
				continue
			}
			c := color.NRGBA{235, 238, 242, uint8(a * 255)}
			for i, d := range dots {
				dd := math.Hypot(fx-d[0], fy-d[1]) - 4.2
				if dd < 0.5 {
					k := math.Min(math.Max(0.5-dd, 0), 1)
					dot := color.NRGBA{30, 34, 40, 255}
					if i == 2 {
						dot = color.NRGBA{94, 152, 217, 255}
					}
					c.R = uint8(float64(c.R)*(1-k) + float64(dot.R)*k)
					c.G = uint8(float64(c.G)*(1-k) + float64(dot.G)*k)
					c.B = uint8(float64(c.B)*(1-k) + float64(dot.B)*k)
				}
			}
			img.SetNRGBA(x, y, c)
		}
	}
	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	return buf.Bytes()
}
