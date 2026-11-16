// Package chat is the chat box logic, modelled on World of Warcraft chat:
// a sticky channel, /r to reply, /w to whisper, Tab to cycle recent people,
// and a shared chat log.
package chat

import (
	"context"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Friend is the chat view of a Steam friend.
type Friend struct {
	ID      uint64
	Name    string // shown name: the nickname the user gave, else the Steam name
	AltName string // the Steam name when Name is a nickname
	Online  bool
	InGame  string
	Avatar  string // avatar image URL
}

// Backend is what the engine needs from the Steam session.
type Backend interface {
	SelfID() uint64
	Friends() []Friend
	Lookup(id uint64) (Friend, bool)
	Send(ctx context.Context, to uint64, text string) error
	Ack(partner uint64, t time.Time)
}

// Notifier receives state changes for the UI.
type Notifier interface {
	EntryAdded(e Entry)
	EntryUpdated(e Entry)
	LogCleared()
	ComposerChanged(s ComposerState)
}

// Entry kinds.
const (
	KindIn     = "in"     // whisper from a friend
	KindOut    = "out"    // whisper to a friend
	KindSystem = "system" // help and info lines
	KindError  = "error"
)

// Entry statuses.
const (
	StatusPending = "pending"
	StatusSent    = "sent"
	StatusFailed  = "failed"
)

// Entry is one line in the chat log.
type Entry struct {
	ID         int64  `json:"id"`
	Time       int64  `json:"time"` // Unix milliseconds
	Kind       string `json:"kind"`
	PeerID     string `json:"peerId"`
	PeerName   string `json:"peerName"`
	PeerAvatar string `json:"peerAvatar"`
	Text       string `json:"text"`
	Status     string `json:"status"`
	Detail     string `json:"detail"`
	// Media are the pictures in the message (at most one for now); the text
	// no longer holds the address of a picture shown here.
	Media []Media `json:"media"`
}

// Media is a picture in a message, served to the UI by myChat itself.
type Media struct {
	ID     string `json:"id"`
	URL    string `json:"url"`   // local address to load it from
	Width  int    `json:"width"` // as the sender says or as decoded; 0 when unknown
	Height int    `json:"height"`
	Kind   string `json:"kind"`  // "upload" | "gif" | "link"
	Title  string `json:"title"` // a GIF's title, if any
}

// Channel kinds.
const (
	ChannelNone    = "none"
	ChannelWhisper = "whisper"
)

// Channel is where plain text goes when Enter is pressed.
type Channel struct {
	Kind       string `json:"kind"`
	TargetID   string `json:"targetId"`
	TargetName string `json:"targetName"`
	Online     bool   `json:"online"`
	Avatar     string `json:"avatar"`
}

// Suggestion is a friend offered while typing /w.
type Suggestion struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	AltName string `json:"altName"` // Steam name when Name is a nickname
	Online  bool   `json:"online"`
	InGame  string `json:"inGame"`
	Avatar  string `json:"avatar"`
}

// ComposerState is everything the UI needs to draw the input line.
type ComposerState struct {
	Text        string       `json:"text"`
	Caret       int          `json:"caret"` // in characters (runes)
	Channel     Channel      `json:"channel"`
	Hint        string       `json:"hint"`
	Suggestions []Suggestion `json:"suggestions"`
	Selected    int          `json:"selected"`
	Error       string       `json:"error"`
	Scroll      int          `json:"scroll"` // log lines scrolled up from the bottom
}

// Action tells the caller what to do after a key.
type Action int

const (
	ActNone  Action = iota
	ActClose        // close the chat box
	ActSent         // a message was accepted for sending
)

const (
	maxLog     = 300
	maxHistory = 50
	maxText    = 1000
)

var (
	whisperCmds = map[string]bool{"/w": true, "/whisper": true, "/t": true, "/tell": true, "/msg": true, "/m": true}
	replyCmds   = map[string]bool{"/r": true, "/reply": true}
)

type echoKey struct {
	to   uint64
	text string
	at   time.Time
}

// Engine holds chat state. All methods are safe for concurrent use.
type Engine struct {
	mu sync.Mutex
	b  Backend
	n  Notifier

	log    []Entry
	nextID int64

	text  []rune
	caret int

	channel  Channel
	chanID   uint64
	lastFrom uint64
	mru      []uint64

	history []string
	histPos int
	draft   string

	sugg    []Friend
	suggSel int

	errMsg string
	scroll int

	echoes  []echoKey
	unacked map[uint64]time.Time

	q chan func() // notifications, delivered in order by one goroutine
}

// New creates an engine.
func New(b Backend, n Notifier) *Engine {
	e := &Engine{b: b, n: n, channel: Channel{Kind: ChannelNone}, unacked: map[uint64]time.Time{}, q: make(chan func(), 1024)}
	go func() {
		for f := range e.q {
			f()
		}
	}()
	return e
}

func (e *Engine) notify(f func()) { e.q <- f }

// Log returns a copy of the chat log.
func (e *Engine) Log() []Entry {
	e.mu.Lock()
	defer e.mu.Unlock()
	return append([]Entry(nil), e.log...)
}

// Composer returns the current input line state.
func (e *Engine) Composer() ComposerState {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.composerLocked()
}

func (e *Engine) composerLocked() ComposerState {
	s := ComposerState{
		Text:     string(e.text),
		Caret:    e.caret,
		Channel:  e.channel,
		Selected: e.suggSel,
		Error:    e.errMsg,
		Scroll:   e.scroll,
	}
	if e.channel.Kind == ChannelWhisper {
		if f, ok := e.b.Lookup(e.chanID); ok {
			s.Channel.TargetName = f.Name
			s.Channel.Online = f.Online
			s.Channel.Avatar = f.Avatar
		}
	}
	for _, f := range e.sugg {
		s.Suggestions = append(s.Suggestions, Suggestion{ID: idStr(f.ID), Name: f.Name, AltName: f.AltName, Online: f.Online, InGame: f.InGame, Avatar: f.Avatar})
	}
	switch {
	case len(e.sugg) > 0:
		s.Hint = "Tab or Enter to pick"
	case e.channel.Kind == ChannelNone && e.lastFrom != 0:
		s.Hint = "/r to reply · /w name"
	case e.channel.Kind == ChannelNone:
		s.Hint = "/w name to whisper · /help"
	}
	return s
}

func (e *Engine) changed() {
	s := e.composerLocked()
	e.notify(func() { e.n.ComposerChanged(s) })
}

// Reset clears the input line (not the channel), as when the box closes.
func (e *Engine) Reset() {
	e.mu.Lock()
	e.text, e.caret = nil, 0
	e.sugg, e.suggSel = nil, 0
	e.errMsg = ""
	e.scroll = 0
	e.histPos = len(e.history)
	e.changed()
	e.mu.Unlock()
}

// Refresh re-sends the composer state (for example after friend data changed).
func (e *Engine) Refresh() {
	e.mu.Lock()
	e.changed()
	e.mu.Unlock()
}

// AckAll marks every conversation with unread messages as read.
func (e *Engine) AckAll() {
	e.mu.Lock()
	pending := e.unacked
	e.unacked = map[uint64]time.Time{}
	e.mu.Unlock()
	for id, t := range pending {
		e.b.Ack(id, t)
	}
}

// Incoming records a message from Steam. It returns the entry and whether it
// is new (a toast is worth showing).
func (e *Engine) Incoming(partner uint64, text string, media []Media, at time.Time, echo bool) (Entry, bool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if echo {
		for i, k := range e.echoes {
			if k.to == partner && k.text == text && time.Since(k.at) < time.Minute {
				e.echoes = append(e.echoes[:i], e.echoes[i+1:]...)
				return Entry{}, false
			}
		}
		ent := e.addLocked(Entry{Kind: KindOut, PeerID: idStr(partner), PeerName: e.nameLocked(partner), PeerAvatar: e.avatarLocked(partner), Text: text, Media: media, Status: StatusSent, Time: at.UnixMilli()})
		e.touchLocked(partner)
		return ent, false
	}
	ent := e.addLocked(Entry{Kind: KindIn, PeerID: idStr(partner), PeerName: e.nameLocked(partner), PeerAvatar: e.avatarLocked(partner), Text: text, Media: media, Time: at.UnixMilli()})
	e.lastFrom = partner
	e.touchLocked(partner)
	e.unacked[partner] = at
	e.changed()
	return ent, true
}

// System adds an info line to the log.
func (e *Engine) System(text string) {
	e.mu.Lock()
	e.addLocked(Entry{Kind: KindSystem, Text: text})
	e.mu.Unlock()
}

func (e *Engine) addLocked(ent Entry) Entry {
	e.nextID++
	ent.ID = e.nextID
	if ent.Time == 0 {
		ent.Time = time.Now().UnixMilli()
	}
	e.log = append(e.log, ent)
	if len(e.log) > maxLog {
		e.log = append([]Entry(nil), e.log[len(e.log)-maxLog:]...)
	}
	e.notify(func() { e.n.EntryAdded(ent) })
	return ent
}

// SetMedia gives a logged entry its pictures (once a linked picture has
// loaded) and the text that goes with them. It reports false when the entry
// has already left the log.
func (e *Engine) SetMedia(id int64, text string, media []Media) (Entry, bool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	for i := range e.log {
		if e.log[i].ID == id {
			e.log[i].Text = text
			e.log[i].Media = media
			out := e.log[i]
			e.notify(func() { e.n.EntryUpdated(out) })
			return out, true
		}
	}
	return Entry{}, false
}

func (e *Engine) updateEntry(id int64, status, detail string) {
	e.mu.Lock()
	var out Entry
	found := false
	for i := range e.log {
		if e.log[i].ID == id {
			e.log[i].Status = status
			e.log[i].Detail = detail
			out, found = e.log[i], true
			break
		}
	}
	if found {
		e.notify(func() { e.n.EntryUpdated(out) })
	}
	e.mu.Unlock()
}

func (e *Engine) nameLocked(id uint64) string {
	if f, ok := e.b.Lookup(id); ok && f.Name != "" {
		return f.Name
	}
	return "Unknown"
}

func (e *Engine) avatarLocked(id uint64) string {
	f, _ := e.b.Lookup(id)
	return f.Avatar
}

func (e *Engine) touchLocked(id uint64) {
	for i, v := range e.mru {
		if v == id {
			e.mru = append(e.mru[:i], e.mru[i+1:]...)
			break
		}
	}
	e.mru = append([]uint64{id}, e.mru...)
	if len(e.mru) > 20 {
		e.mru = e.mru[:20]
	}
}

func (e *Engine) setWhisperLocked(id uint64) {
	e.chanID = id
	e.channel = Channel{Kind: ChannelWhisper, TargetID: idStr(id), TargetName: e.nameLocked(id)}
}

// Insert types text at the caret (used for paste).
func (e *Engine) Insert(s string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	s = strings.NewReplacer("\r\n", " ", "\n", " ", "\r", " ", "\t", " ").Replace(s)
	e.insertLocked(s)
	e.afterEditLocked()
	e.changed()
}

func (e *Engine) insertLocked(s string) {
	r := []rune(s)
	if room := maxText - len(e.text); len(r) > room {
		r = r[:max(room, 0)]
	}
	if len(r) == 0 {
		return
	}
	nt := make([]rune, 0, len(e.text)+len(r))
	nt = append(nt, e.text[:e.caret]...)
	nt = append(nt, r...)
	nt = append(nt, e.text[e.caret:]...)
	e.text = nt
	e.caret += len(r)
}

// HandleKey processes one key from the input hook. key is empty for text.
func (e *Engine) HandleKey(text, key string, shift bool) Action {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.errMsg = ""
	act := ActNone
	if text != "" {
		e.insertLocked(text)
		e.afterEditLocked()
		e.changed()
		return ActNone
	}
	switch key {
	case "Enter":
		act = e.enterLocked()
	case "Escape":
		e.text, e.caret, e.sugg = nil, 0, nil
		act = ActClose
	case "Backspace":
		if e.caret > 0 {
			e.text = append(e.text[:e.caret-1], e.text[e.caret:]...)
			e.caret--
		}
		e.afterEditLocked()
	case "WordBackspace":
		start := wordLeft(e.text, e.caret)
		e.text = append(e.text[:start], e.text[e.caret:]...)
		e.caret = start
		e.afterEditLocked()
	case "Delete":
		if e.caret < len(e.text) {
			e.text = append(e.text[:e.caret], e.text[e.caret+1:]...)
		}
		e.afterEditLocked()
	case "Left":
		e.caret = max(e.caret-1, 0)
	case "Right":
		e.caret = min(e.caret+1, len(e.text))
	case "WordLeft":
		e.caret = wordLeft(e.text, e.caret)
	case "WordRight":
		e.caret = wordRight(e.text, e.caret)
	case "Home":
		e.caret = 0
	case "End":
		e.caret = len(e.text)
	case "Up":
		if len(e.sugg) > 0 {
			e.suggSel = (e.suggSel - 1 + len(e.sugg)) % len(e.sugg)
		} else {
			e.historyLocked(-1)
		}
	case "Down":
		if len(e.sugg) > 0 {
			e.suggSel = (e.suggSel + 1) % len(e.sugg)
		} else {
			e.historyLocked(1)
		}
	case "PageUp":
		e.scroll = min(e.scroll+5, max(len(e.log)-1, 0))
	case "PageDown":
		e.scroll = max(e.scroll-5, 0)
	case "Tab":
		e.tabLocked(shift)
	}
	e.changed()
	return act
}

func (e *Engine) historyLocked(dir int) {
	if len(e.history) == 0 {
		return
	}
	if e.histPos == len(e.history) {
		e.draft = string(e.text)
	}
	e.histPos = min(max(e.histPos+dir, 0), len(e.history))
	if e.histPos == len(e.history) {
		e.text = []rune(e.draft)
	} else {
		e.text = []rune(e.history[e.histPos])
	}
	e.caret = len(e.text)
}

func (e *Engine) tabLocked(shift bool) {
	if len(e.sugg) > 0 {
		e.pickSuggestionLocked()
		return
	}
	if len(e.mru) == 0 {
		e.errMsg = "No recent conversations"
		return
	}
	idx := -1
	if e.channel.Kind == ChannelWhisper {
		for i, v := range e.mru {
			if v == e.chanID {
				idx = i
				break
			}
		}
	}
	if shift {
		idx--
		if idx < 0 {
			idx = len(e.mru) - 1
		}
	} else {
		idx = (idx + 1) % len(e.mru)
	}
	e.setWhisperLocked(e.mru[idx])
}

func (e *Engine) pickSuggestionLocked() {
	if e.suggSel >= len(e.sugg) {
		e.suggSel = 0
	}
	f := e.sugg[e.suggSel]
	e.setWhisperLocked(f.ID)
	e.text, e.caret = nil, 0
	e.sugg, e.suggSel = nil, 0
}

// afterEditLocked implements live channel switching, as in WoW: typing
// "/r " or "/w name " turns the line into that channel at once.
func (e *Engine) afterEditLocked() {
	e.histPos = len(e.history)
	t := string(e.text)
	e.sugg = nil
	if !strings.HasPrefix(t, "/") {
		return
	}
	sp := strings.IndexByte(t, ' ')
	if sp < 0 {
		return
	}
	cmd := strings.ToLower(t[:sp])
	arg := t[sp+1:]
	switch {
	case replyCmds[cmd] && arg == "":
		if e.lastFrom == 0 {
			e.errMsg = "No one has whispered you yet"
			return
		}
		e.setWhisperLocked(e.lastFrom)
		e.text, e.caret = nil, 0
	case whisperCmds[cmd]:
		q := strings.TrimSpace(arg)
		if q == "" {
			e.sugg = e.recentFriendsLocked(6)
			e.suggSel = 0
			return
		}
		matches, exact := e.matchLocked(q)
		if strings.HasSuffix(arg, " ") && (len(exact) == 1 || len(matches) == 1) {
			target := matches[0]
			if len(exact) == 1 {
				target = exact[0]
			}
			e.setWhisperLocked(target.ID)
			e.text, e.caret = nil, 0
			return
		}
		if len(matches) > 6 {
			matches = matches[:6]
		}
		e.sugg = matches
		if e.suggSel >= len(matches) {
			e.suggSel = 0
		}
	}
}

func (e *Engine) recentFriendsLocked(n int) []Friend {
	var out []Friend
	for _, id := range e.mru {
		if f, ok := e.b.Lookup(id); ok {
			out = append(out, f)
		}
		if len(out) == n {
			break
		}
	}
	return out
}

// matchLocked finds friends by nickname or Steam name: exact, then prefix,
// then word prefix, then substring. Online friends come first inside each
// group.
func (e *Engine) matchLocked(q string) (all, exact []Friend) {
	q = strings.ToLower(q)
	var prefix, word, sub []Friend
	for _, f := range e.b.Friends() {
		switch min(matchRank(f.Name, q), matchRank(f.AltName, q)) {
		case 0:
			exact = append(exact, f)
		case 1:
			prefix = append(prefix, f)
		case 2:
			word = append(word, f)
		case 3:
			sub = append(sub, f)
		}
	}
	for _, g := range [][]Friend{exact, prefix, word, sub} {
		all = append(all, onlineFirst(g)...)
	}
	return all, exact
}

// matchRank is 0 for an exact match, 1 prefix, 2 word prefix, 3 substring,
// 4 no match. q is lower case.
func matchRank(name, q string) int {
	n := strings.ToLower(name)
	switch {
	case n == "":
		return 4
	case n == q:
		return 0
	case strings.HasPrefix(n, q):
		return 1
	case strings.Contains(" "+n, " "+q):
		return 2
	case strings.Contains(n, q):
		return 3
	}
	return 4
}

func onlineFirst(in []Friend) []Friend {
	out := make([]Friend, 0, len(in))
	for _, f := range in {
		if f.Online {
			out = append(out, f)
		}
	}
	for _, f := range in {
		if !f.Online {
			out = append(out, f)
		}
	}
	return out
}

func (e *Engine) enterLocked() Action {
	t := strings.TrimSpace(string(e.text))
	if len(e.sugg) > 0 && isWhisperPrefixOnly(t) {
		e.pickSuggestionLocked()
		return ActNone
	}
	if t == "" {
		return ActClose
	}
	e.history = append(e.history, t)
	if len(e.history) > maxHistory {
		e.history = e.history[len(e.history)-maxHistory:]
	}
	e.histPos = len(e.history)

	if strings.HasPrefix(t, "/") {
		return e.commandLocked(t)
	}
	if e.channel.Kind == ChannelWhisper {
		return e.sendWhisperLocked(e.chanID, t)
	}
	e.errMsg = "Pick who to message: /w name or /r"
	return ActNone
}

func isWhisperPrefixOnly(t string) bool {
	f := strings.Fields(t)
	return len(f) >= 1 && whisperCmds[strings.ToLower(f[0])]
}

func (e *Engine) commandLocked(t string) Action {
	fields := strings.Fields(t)
	cmd := strings.ToLower(fields[0])
	rest := strings.TrimSpace(strings.TrimPrefix(t, fields[0]))
	switch {
	case replyCmds[cmd]:
		if e.lastFrom == 0 {
			e.errMsg = "No one has whispered you yet"
			return ActNone
		}
		e.setWhisperLocked(e.lastFrom)
		if rest == "" {
			e.text, e.caret = nil, 0
			return ActNone
		}
		return e.sendWhisperLocked(e.lastFrom, rest)
	case whisperCmds[cmd]:
		words := fields[1:]
		if len(words) == 0 {
			e.errMsg = "Usage: /w name message"
			return ActNone
		}
		// Try the longest run of words that is exactly one friend's name.
		for k := len(words); k >= 1; k-- {
			_, exact := e.matchLocked(strings.Join(words[:k], " "))
			if len(exact) == 1 {
				e.setWhisperLocked(exact[0].ID)
				msg := strings.Join(words[k:], " ")
				if msg == "" {
					e.text, e.caret = nil, 0
					return ActNone
				}
				return e.sendWhisperLocked(exact[0].ID, msg)
			}
		}
		all, _ := e.matchLocked(words[0])
		if len(all) != 1 {
			if len(all) == 0 {
				e.errMsg = "No friend named \"" + words[0] + "\""
			} else {
				e.errMsg = "More than one friend matches \"" + words[0] + "\""
			}
			return ActNone
		}
		e.setWhisperLocked(all[0].ID)
		msg := strings.Join(words[1:], " ")
		if msg == "" {
			e.text, e.caret = nil, 0
			return ActNone
		}
		return e.sendWhisperLocked(all[0].ID, msg)
	case cmd == "/help" || cmd == "/?":
		for _, l := range []string{
			"/r [msg] – reply to the last whisper",
			"/w name [msg] – whisper a friend",
			"Tab – cycle recent people · ↑↓ – sent history",
			"PgUp/PgDn – scroll · Esc – close",
		} {
			e.addLocked(Entry{Kind: KindSystem, Text: l})
		}
		e.text, e.caret = nil, 0
		return ActNone
	case cmd == "/clear":
		e.log = nil
		e.notify(e.n.LogCleared)
		e.text, e.caret = nil, 0
		return ActNone
	}
	e.errMsg = "Unknown command " + cmd + " (try /help)"
	return ActNone
}

func (e *Engine) sendWhisperLocked(to uint64, msg string) Action {
	ent := e.addLocked(Entry{Kind: KindOut, PeerID: idStr(to), PeerName: e.nameLocked(to), PeerAvatar: e.avatarLocked(to), Text: msg, Status: StatusPending})
	e.echoes = append(e.echoes, echoKey{to: to, text: msg, at: time.Now()})
	if len(e.echoes) > 50 {
		e.echoes = e.echoes[len(e.echoes)-50:]
	}
	e.touchLocked(to)
	delete(e.unacked, to)
	e.text, e.caret = nil, 0
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		if err := e.b.Send(ctx, to, msg); err != nil {
			e.updateEntry(ent.ID, StatusFailed, err.Error())
			return
		}
		e.updateEntry(ent.ID, StatusSent, "")
		e.b.Ack(to, time.Now())
	}()
	return ActSent
}

func wordLeft(t []rune, i int) int {
	for i > 0 && t[i-1] == ' ' {
		i--
	}
	for i > 0 && t[i-1] != ' ' {
		i--
	}
	return i
}

func wordRight(t []rune, i int) int {
	for i < len(t) && t[i] == ' ' {
		i++
	}
	for i < len(t) && t[i] != ' ' {
		i++
	}
	return i
}

func idStr(id uint64) string { return strconv.FormatUint(id, 10) }

// ParseID converts a string ID from the UI.
func ParseID(s string) uint64 {
	v, _ := strconv.ParseUint(s, 10, 64)
	return v
}
