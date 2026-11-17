//go:build windows

package sound

import (
	"encoding/binary"
	"testing"
)

func TestBuild(t *testing.T) {
	b := build()
	if string(b[0:4]) != "RIFF" || string(b[8:16]) != "WAVEfmt " || string(b[36:40]) != "data" {
		t.Fatal("not a WAV header")
	}
	if got := binary.LittleEndian.Uint32(b[4:8]); int(got) != len(b)-8 {
		t.Fatalf("RIFF size %d, want %d", got, len(b)-8)
	}
	pcm := b[44:]
	if got := binary.LittleEndian.Uint32(b[40:44]); int(got) != len(pcm) {
		t.Fatalf("data size %d, want %d", got, len(pcm))
	}
	peak := 0
	for i := 0; i < len(pcm); i += 2 {
		v := int(int16(binary.LittleEndian.Uint16(pcm[i:])))
		if v < 0 {
			v = -v
		}
		peak = max(peak, v)
	}
	// Soft: well under full scale, but clearly audible.
	if peak > 32767/2 || peak < 32767/10 {
		t.Fatalf("peak %d out of range", peak)
	}
	first := int16(binary.LittleEndian.Uint16(pcm[0:]))
	last := int16(binary.LittleEndian.Uint16(pcm[len(pcm)-2:]))
	if first != 0 || last != 0 {
		t.Fatalf("starts at %d and ends at %d, want silence (no click)", first, last)
	}
}
