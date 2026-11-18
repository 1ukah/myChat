package media

import (
	"net/url"
	"path"
	"regexp"
	"strconv"
	"strings"
)

const (
	KindUpload = "upload"
	KindGIF    = "gif"
	KindLink   = "link"
)

type Found struct {
	URL     string
	Full    string
	Width   int
	Height  int
	Kind    string
	Title   string
	Trusted bool
	Strip   string
}

func Find(bbcode, plain string) (Found, bool) {
	if f, ok := findImg(bbcode); ok {
		return f, true
	}
	return findLink(plain)
}

func StripText(plain string, f Found) string {
	if f.Strip == "" {
		return plain
	}
	return strings.TrimSpace(strings.Replace(plain, f.Strip, "", 1))
}

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
		if full := attrs["src"]; f.Kind == KindUpload && full != "" && full != src {
			if fu, ok := imageURL(full); ok && (trusted(fu) || !f.Trusted) {
				f.Full = full
			}
		}
		f.Strip = strings.TrimSpace(tagRE.ReplaceAllString(inner, ""))
		return f, true
	}
}

var tagRE = regexp.MustCompile(`\[[^\]]*\]`)

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
