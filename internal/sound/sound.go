//go:build windows

// Package sound plays myChat's new message sound: a soft, short falling
// two-note chime (D5 to G4), quiet and round, built in memory as a WAV and
// played through the Windows sound API, so it follows myChat's volume in
// the Windows volume mixer.
package sound

import (
	"bytes"
	"encoding/binary"
	"math"
	"sync"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

var pPlaySoundW = windows.NewLazySystemDLL("winmm.dll").NewProc("PlaySoundW")

const (
	sndAsync     = 0x0001
	sndNoDefault = 0x0002
	sndMemory    = 0x0004

	rate = 44100
	// Messages that arrive closer together than this play one sound.
	minGap = 250 * time.Millisecond
)

var (
	once sync.Once
	wav  []byte // kept for the life of the app: Windows reads it while playing

	mu   sync.Mutex
	last time.Time
)

// Play plays the message sound and returns at once.
func Play() {
	mu.Lock()
	if time.Since(last) < minGap {
		mu.Unlock()
		return
	}
	last = time.Now()
	mu.Unlock()
	once.Do(func() { wav = build() })
	if pPlaySoundW.Find() != nil {
		return
	}
	pPlaySoundW.Call(uintptr(unsafe.Pointer(&wav[0])), 0, sndMemory|sndAsync|sndNoDefault)
}

// note is one soft mallet-like tone: a sine with a little of its octave,
// a short rounded attack and a smooth exponential decay.
type note struct {
	freq  float64 // Hz
	start float64 // s
	tau   float64 // decay time constant, s
	gain  float64
}

func build() []byte {
	notes := []note{
		{freq: 587.33, start: 0, tau: 0.075, gain: 0.16},
		{freq: 392.00, start: 0.085, tau: 0.14, gain: 0.2},
	}
	const length = 0.7
	const attack = 0.008
	const tail = 0.03
	n := int(length * rate)
	pcm := make([]int16, n)
	for i := range pcm {
		t := float64(i) / rate
		var v float64
		for _, nt := range notes {
			dt := t - nt.start
			if dt < 0 {
				continue
			}
			env := math.Exp(-dt / nt.tau)
			if dt < attack {
				env *= 0.5 - 0.5*math.Cos(math.Pi*dt/attack)
			}
			w := 2 * math.Pi * nt.freq * dt
			v += nt.gain * env * (math.Sin(w) + 0.12*math.Exp(-dt/nt.tau)*math.Sin(2*w))
		}
		// The end fades to silence, so nothing clicks.
		if rest := length - t; rest < tail {
			v *= rest / tail
		}
		pcm[i] = int16(math.Max(-1, math.Min(1, v)) * 32767)
	}

	var b bytes.Buffer
	data := uint32(n * 2)
	b.WriteString("RIFF")
	binary.Write(&b, binary.LittleEndian, 36+data)
	b.WriteString("WAVEfmt ")
	binary.Write(&b, binary.LittleEndian, struct {
		Size             uint32
		Format, Channels uint16
		Rate, ByteRate   uint32
		Align, Bits      uint16
	}{16, 1, 1, rate, rate * 2, 2, 16}) // PCM, mono, 16 bit
	b.WriteString("data")
	binary.Write(&b, binary.LittleEndian, data)
	binary.Write(&b, binary.LittleEndian, pcm)
	return b.Bytes()
}
