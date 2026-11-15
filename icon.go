package main

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"math"
)

// trayIcon is the tray icon: a white chat bubble.
func trayIcon() []byte { return drawIcon(64, false) }

// drawIcon draws the app mark: a white chat bubble with two dark dots and one
// light-blue dot, in the flat white style of the CS2 menu icons. With tile set
// it sits on a dark square tile (app icon).
func drawIcon(n int, tile bool) []byte {
	img := image.NewNRGBA(image.Rect(0, 0, n, n))
	s := float64(n) / 64
	clamp := func(v float64) float64 { return math.Min(math.Max(v, 0), 1) }
	rrect := func(x, y, cx, cy, hw, hh, r float64) float64 {
		dx := math.Max(math.Abs(x-cx)-(hw-r), 0)
		dy := math.Max(math.Abs(y-cy)-(hh-r), 0)
		return math.Hypot(dx, dy) - r
	}
	inset, scale := 0.0, 1.0
	if tile {
		inset, scale = 12, 0.62
	}
	tf := func(v float64) float64 { return (v-inset*s)/scale - 0 }
	dots := [][2]float64{{20, 27}, {32, 27}, {44, 27}}
	for y := 0; y < n; y++ {
		for x := 0; x < n; x++ {
			fx, fy := float64(x)+0.5, float64(y)+0.5
			var c color.NRGBA
			if tile {
				d := rrect(fx, fy, float64(n)/2, float64(n)/2, float64(n)/2-0.5, float64(n)/2-0.5, 6*s)
				if a := clamp(0.5 - d); a > 0 {
					t := fy / float64(n)
					c = color.NRGBA{uint8(34 - 12*t), uint8(38 - 12*t), uint8(44 - 12*t), uint8(a * 255)}
				}
			}
			// Bubble coordinates in the 64-unit design space.
			bx, by := tf(fx)/s, tf(fy)/s
			px := 1 / (s * scale)
			a := clamp(0.5 - rrect(bx, by, 32, 27, 26, 19, 9)/px)
			if by >= 40 {
				hyp := ((bx - 14) - (58-by)*1.15) / math.Sqrt(1+1.15*1.15)
				d := math.Max(math.Max(hyp, 14-bx), by-58)
				a = math.Max(a, clamp(0.5-d/px))
			}
			if a > 0 {
				fg := color.NRGBA{235, 238, 242, 255}
				for i, d := range dots {
					k := clamp(0.5 - (math.Hypot(bx-d[0], by-d[1])-4.2)/px)
					if k > 0 {
						dot := color.NRGBA{30, 34, 40, 255}
						if i == 2 {
							dot = color.NRGBA{94, 152, 217, 255}
						}
						fg.R = uint8(float64(fg.R)*(1-k) + float64(dot.R)*k)
						fg.G = uint8(float64(fg.G)*(1-k) + float64(dot.G)*k)
						fg.B = uint8(float64(fg.B)*(1-k) + float64(dot.B)*k)
					}
				}
				// Blend the bubble over the tile.
				ba := float64(c.A) / 255
				oa := a + ba*(1-a)
				mix := func(f, b uint8) uint8 {
					if oa == 0 {
						return 0
					}
					return uint8((float64(f)*a + float64(b)*ba*(1-a)) / oa)
				}
				c = color.NRGBA{mix(fg.R, c.R), mix(fg.G, c.G), mix(fg.B, c.B), uint8(oa * 255)}
			}
			img.SetNRGBA(x, y, c)
		}
	}
	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	return buf.Bytes()
}
