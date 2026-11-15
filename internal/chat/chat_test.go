package chat

import (
	"context"
	"sync"
	"testing"
	"time"
)

type fakeBackend struct {
	mu      sync.Mutex
	friends []Friend
	self    Friend
	sent    []string
}

func (f *fakeBackend) SelfID() uint64 { return f.self.ID }
func (f *fakeBackend) Friends() []Friend {
	return f.friends
}
func (f *fakeBackend) Lookup(id uint64) (Friend, bool) {
	if id == f.self.ID {
		return f.self, true
	}
	for _, x := range f.friends {
		if x.ID == id {
			return x, true
		}
	}
	return Friend{}, false
}
func (f *fakeBackend) Send(_ context.Context, to uint64, text string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.sent = append(f.sent, idStr(to)+":"+text)
	return nil
}
func (f *fakeBackend) Ack(uint64, time.Time) {}

type nopNotifier struct{}

func (nopNotifier) EntryAdded(Entry)              {}
func (nopNotifier) EntryUpdated(Entry)            {}
func (nopNotifier) LogCleared()                   {}
func (nopNotifier) ComposerChanged(ComposerState) {}

func newTest() (*Engine, *fakeBackend) {
	b := &fakeBackend{
		self: Friend{ID: 1, Name: "me", Group: "g1"},
		friends: []Friend{
			{ID: 10, Name: "Leo DiCaprio", Online: true, Group: "g1"},
			{ID: 11, Name: "Leo Messi", Online: true},
			{ID: 12, Name: "Frisk", Online: false, Group: "g1"},
			{ID: 13, Name: "SANT1", Online: true},
		},
	}
	return New(b, nopNotifier{}), b
}

func typeText(e *Engine, s string) {
	for _, r := range s {
		e.HandleKey(string(r), "", false)
	}
}

func (f *fakeBackend) waitSent(t *testing.T, n int) []string {
	t.Helper()
	for i := 0; i < 100; i++ {
		f.mu.Lock()
		if len(f.sent) >= n {
			out := append([]string(nil), f.sent...)
			f.mu.Unlock()
			return out
		}
		f.mu.Unlock()
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("expected %d sends", n)
	return nil
}

func TestLiveWhisperSwitch(t *testing.T) {
	e, b := newTest()
	typeText(e, "/w fri ")
	c := e.Composer()
	if c.Channel.Kind != ChannelWhisper || c.Channel.TargetID != "12" || c.Text != "" {
		t.Fatalf("got %+v", c)
	}
	typeText(e, "hello")
	if act := e.HandleKey("", "Enter", false); act != ActSent {
		t.Fatalf("act %v", act)
	}
	if got := b.waitSent(t, 1); got[0] != "12:hello" {
		t.Fatalf("sent %v", got)
	}
}

func TestAmbiguousPrefixWaits(t *testing.T) {
	e, _ := newTest()
	typeText(e, "/w leo ")
	c := e.Composer()
	if c.Channel.Kind != ChannelNone || len(c.Suggestions) != 2 {
		t.Fatalf("got %+v", c)
	}
	e.HandleKey("", "Down", false)
	e.HandleKey("", "Tab", false)
	c = e.Composer()
	if c.Channel.TargetID != "11" {
		t.Fatalf("tab pick: %+v", c)
	}
}

func TestFullLineWhisperWithSpacedName(t *testing.T) {
	e, b := newTest()
	e.mu.Lock()
	e.text = []rune("/w leo messi gg wp")
	e.caret = len(e.text)
	e.mu.Unlock()
	e.HandleKey("", "Enter", false)
	if got := b.waitSent(t, 1); got[0] != "11:gg wp" {
		t.Fatalf("sent %v", got)
	}
}

func TestReplyAndTabCycle(t *testing.T) {
	e, b := newTest()
	typeText(e, "/r ")
	if e.Composer().Error == "" {
		t.Fatal("expected error with no whisper yet")
	}
	e.Reset()
	e.Incoming(13, "yo", time.Now(), false)
	e.Incoming(10, "hey", time.Now(), false)
	typeText(e, "/r ")
	if c := e.Composer(); c.Channel.TargetID != "10" {
		t.Fatalf("reply target %+v", c)
	}
	e.HandleKey("", "Tab", false)
	if c := e.Composer(); c.Channel.TargetID != "13" {
		t.Fatalf("tab cycle %+v", c)
	}
	typeText(e, "sup")
	e.HandleKey("", "Enter", false)
	if got := b.waitSent(t, 1); got[0] != "13:sup" {
		t.Fatalf("sent %v", got)
	}
}

func TestParty(t *testing.T) {
	e, b := newTest()
	typeText(e, "/p ")
	if c := e.Composer(); c.Channel.Kind != ChannelParty {
		t.Fatalf("got %+v", c)
	}
	typeText(e, "rush b")
	e.HandleKey("", "Enter", false)
	got := b.waitSent(t, 2)
	if len(got) != 2 {
		t.Fatalf("sent %v", got)
	}
}

func TestEchoDedupe(t *testing.T) {
	e, b := newTest()
	typeText(e, "/w sant1 ")
	typeText(e, "hi")
	e.HandleKey("", "Enter", false)
	b.waitSent(t, 1)
	if _, isNew := e.Incoming(13, "hi", time.Now(), true); isNew {
		t.Fatal("echo should not be new")
	}
	n := 0
	for _, l := range e.Log() {
		if l.Kind == KindOut {
			n++
		}
	}
	if n != 1 {
		t.Fatalf("expected 1 out entry, got %d", n)
	}
}

func TestEmptyEnterCloses(t *testing.T) {
	e, _ := newTest()
	if act := e.HandleKey("", "Enter", false); act != ActClose {
		t.Fatalf("act %v", act)
	}
}
