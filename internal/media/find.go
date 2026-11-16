// Package media finds pictures in chat messages and fetches them for the
// overlay, so the WebView never loads an address taken from a message.
package media

import (
	"net/url"
	"path"
	"regexp"
	"strconv"
	"strings"
)

// Kinds of picture.
const (
	KindUpload = "upload" // sent through Steam chat
	KindGIF    = "gif"    // sent with /giphy
	KindLink   = "link"   // an image address typed into the message
)

// Found is the picture a message refers to.
type Found struct {
	URL    string // the picture to load
	Width  int    // as the message says; 0 when unknown
	Height int
	Kind   string
	Title  string
	// Trusted pictures are hosted by Steam or Giphy, so loading them tells
	// no stranger the user's address.
	Trusted bool
	// Strip is the text the picture stands for in Steam's plain version of
	// the message; "" when the text should stay as it is.
	Strip string
}

// Find returns the first picture in a message: an [img] tag in its BBCode,
// or else an image address in its plain text.
func Find(bbcode, plain string) (Found, bool) {
	if f, ok := findImg(bbcode); ok {
		return f, true
	}
	return findLink(plain)
}

// StripText removes the text a picture stands for from the plain message.
func StripText(plain string, f Found) string {
	if f.Strip == "" {
		return plain
	}
	return strings.TrimSpace(strings.Replace(plain, f.Strip, "", 1))
}

// findImg reads Steam's image tag. Uploads look like
//
//	[img src=https://images.steamusercontent.com/ugc/…/ thumbnail_src=…?imw=512… width=3840 height=2160][url=…]…[/url][/img]
//
// and /giphy like
//
//	[img src="https://giphy.com/gifs/…" thumbnail_src="https://media2.giphy.com/…/giphy.gif" height="480" width="480" giphy_search="cat" title="…"][url=…]…[/url][/img]
//
// Values may be quoted or bare.
func findImg(s string) (Found, bool) {
	lower := strings.ToLower(s)
	for from := 0; ; {
		i := strings.Index(lower[from:], "[img")
		if i < 0 {
			return Found{}, false
		}
		i += from
		from = i + 4
		if from >= len(s) || (s[from] != ' ' && s[from] != ']') {
			continue
		}
		attrs, end, ok := readAttrs(s, from)
		if !ok {
			return Found{}, false
		}
		inner := s[end:]
		if j := strings.Index(strings.ToLower(inner), "[/img]"); j >= 0 {
			inner = inner[:j]
		}
		src := attrs["thumbnail_src"]
		if src == "" {
			src = attrs["src"]
		}
		u, ok := imageURL(src)
		if !ok {
			continue
		}
		f := Found{URL: src, Title: attrs["title"], Trusted: trusted(u), Kind: KindUpload}
		f.Width, _ = strconv.Atoi(attrs["width"])
		f.Height, _ = strconv.Atoi(attrs["height"])
		if f.Width <= 0 || f.Height <= 0 {
			f.Width, f.Height = 0, 0
		}
		if _, gif := attrs["giphy_search"]; gif || strings.HasSuffix(u.Hostname(), "giphy.com") {
			f.Kind = KindGIF
		}
		// Steam's plain text for an image is the text inside the tag (the
		// address in its [url]).
		f.Strip = strings.TrimSpace(tagRE.ReplaceAllString(inner, ""))
		return f, true
	}
}

var tagRE = regexp.MustCompile(`\[[^\]]*\]`)

// readAttrs reads `name=value` pairs from s[i:] up to the closing "]" and
// returns them with the index after it.
func readAttrs(s string, i int) (map[string]string, int, bool) {
	attrs := map[string]string{}
	for i < len(s) {
		for i < len(s) && s[i] == ' ' {
			i++
		}
		if i >= len(s) {
			break
		}
		if s[i] == ']' {
			return attrs, i + 1, true
		}
		start := i
		for i < len(s) && s[i] != '=' && s[i] != ' ' && s[i] != ']' {
			i++
		}
		name := strings.ToLower(s[start:i])
		if i >= len(s) || s[i] != '=' {
			attrs[name] = ""
			continue
		}
		i++
		var val string
		if i < len(s) && s[i] == '"' {
			j := strings.IndexByte(s[i+1:], '"')
			if j < 0 {
				return nil, 0, false
			}
			val = s[i+1 : i+1+j]
			i += j + 2
		} else {
			start := i
			for i < len(s) && s[i] != ' ' && s[i] != ']' {
				i++
			}
			val = s[start:i]
		}
		attrs[name] = val
	}
	return nil, 0, false
}

var linkRE = regexp.MustCompile(`https?://[^\s\[\]<>"']+`)

// Hosts that serve pictures at addresses without a file extension.
var imageHosts = map[string]bool{
	"i.imgur.com":     true,
	"i.redd.it":       true,
	"pbs.twimg.com":   true,
	"media.tenor.com": true,
}

var imageExts = map[string]bool{".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".webp": true}

func findLink(plain string) (Found, bool) {
	for _, raw := range linkRE.FindAllString(plain, -1) {
		raw = strings.TrimRight(raw, ".,!?)")
		u, ok := imageURL(raw)
		if !ok {
			continue
		}
		host := u.Hostname()
		if !imageExts[strings.ToLower(path.Ext(u.Path))] && !imageHosts[host] && !trusted(u) {
			continue
		}
		f := Found{URL: raw, Kind: KindLink, Trusted: trusted(u)}
		if strings.TrimSpace(plain) == raw {
			f.Strip = raw
		}
		return f, true
	}
	return Found{}, false
}

func imageURL(raw string) (*url.URL, bool) {
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Hostname() == "" {
		return nil, false
	}
	return u, true
}

// trusted reports whether a picture is served by Steam or Giphy.
func trusted(u *url.URL) bool {
	if u.Scheme != "https" {
		return false
	}
	return trustedHost(u.Hostname())
}

func trustedHost(h string) bool {
	h = strings.ToLower(h)
	switch {
	case h == "images.steamusercontent.com", h == "steamuserimages-a.akamaihd.net",
		h == "giphy.com", strings.HasSuffix(h, ".giphy.com"),
		strings.HasSuffix(h, ".steamstatic.com"):
		return true
	}
	return false
}
