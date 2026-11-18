package media

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"mime"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"
)

const (
	SiteYouTube = "youtube"
	SiteX       = "x"
	SiteReddit  = "reddit"
	SiteWeb     = "web"
)

const (
	agentCrawler = "Mozilla/5.0 (compatible; Discordbot/2.0; +https://discordapp.com)"
	agentBrowser = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0 Safari/537.36"
)

const (
	maxJSON  = 1 << 20
	maxPage  = 768 << 10
	maxTitle = 200
	maxText  = 280
)

var errNoPreview = errors.New("media: no preview")

type Page struct {
	URL   string
	Site  string
	ID    string
	Strip string
}

type Preview struct {
	URL      string
	Site     string
	SiteName string
	Title    string
	Author   string
	Text     string
	Video    bool
	Thumb    string
	Width    int
	Height   int
}

func Links(plain string) []string {
	var out []string
	for _, raw := range linkRE.FindAllString(plain, -1) {
		raw = strings.TrimRight(raw, ".,!?)")
		if _, ok := imageURL(raw); ok {
			out = append(out, raw)
		}
	}
	return out
}

func FindPage(plain string) (Page, bool) {
	for _, raw := range Links(plain) {
		if _, pic := findLink(raw); pic {
			continue
		}
		u, _ := url.Parse(raw)
		p := Page{URL: raw, Site: SiteWeb}
		if id := youTubeID(u); id != "" {
			p.Site, p.ID = SiteYouTube, id
		} else if id := xPostID(u); id != "" {
			p.Site, p.ID = SiteX, id
		} else if id := redditPostID(u); id != "" {
			p.Site, p.ID = SiteReddit, id
		}
		if strings.TrimSpace(plain) == raw {
			p.Strip = raw
		}
		return p, true
	}
	return Page{}, false
}

var youTubeIDRE = regexp.MustCompile(`^[A-Za-z0-9_-]{11}$`)

func youTubeID(u *url.URL) string {
	host := strings.TrimPrefix(strings.ToLower(u.Hostname()), "www.")
	var id string
	switch host {
	case "youtu.be":
		id = strings.Trim(u.Path, "/")
	case "youtube.com", "m.youtube.com", "music.youtube.com":
		if u.Path == "/watch" {
			id = u.Query().Get("v")
		} else {
			parts := strings.Split(strings.Trim(u.Path, "/"), "/")
			if len(parts) == 2 && (parts[0] == "shorts" || parts[0] == "live" || parts[0] == "embed") {
				id = parts[1]
			}
		}
	}
	if youTubeIDRE.MatchString(id) {
		return id
	}
	return ""
}

var xHosts = map[string]bool{
	"x.com": true, "twitter.com": true, "mobile.twitter.com": true, "mobile.x.com": true,
	"vxtwitter.com": true, "fxtwitter.com": true, "fixupx.com": true, "fixvx.com": true,
	"twittpr.com": true, "girlcockx.com": true, "stupidpenisx.com": true,
}

var digitsRE = regexp.MustCompile(`^\d{1,25}$`)

func xPostID(u *url.URL) string {
	host := strings.TrimPrefix(strings.ToLower(u.Hostname()), "www.")
	if !xHosts[host] {
		return ""
	}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) >= 3 && (parts[1] == "status" || parts[1] == "statuses") && digitsRE.MatchString(parts[2]) {
		return parts[2]
	}
	return ""
}

var redditIDRE = regexp.MustCompile(`^[a-z0-9]{4,12}$`)

func redditPostID(u *url.URL) string {
	host := strings.ToLower(u.Hostname())
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	var id string
	switch {
	case host == "redd.it":
		id = parts[0]
	case host == "reddit.com" || strings.HasSuffix(host, ".reddit.com"):
		for i := 0; i+1 < len(parts); i++ {
			if parts[i] == "comments" {
				id = parts[i+1]
				break
			}
		}
	}
	id = strings.ToLower(id)
	if redditIDRE.MatchString(id) {
		return id
	}
	return ""
}

func (s *Store) Preview(ctx context.Context, p Page) (Preview, error) {
	var pv Preview
	var err error
	switch p.Site {
	case SiteYouTube:
		pv, err = s.youTube(ctx, p)
	case SiteX:
		pv, err = s.xPost(ctx, p)
	case SiteReddit:
		pv, err = s.redditPost(ctx, p)
	default:
		return s.webPage(ctx, p)
	}
	if err != nil && ctx.Err() == nil {
		web, werr := s.webPage(ctx, p)
		if werr == nil {
			web.Site = p.Site
			return web, nil
		}
	}
	return pv, err
}

func (s *Store) redditPost(ctx context.Context, p Page) (Preview, error) {
	type image struct {
		URL    string `json:"url"`
		Width  int    `json:"width"`
		Height int    `json:"height"`
	}
	var o struct {
		Data struct {
			Children []struct {
				Data struct {
					Sub     string `json:"subreddit_name_prefixed"`
					Title   string `json:"title"`
					Text    string `json:"selftext"`
					Author  string `json:"author"`
					NSFW    bool   `json:"over_18"`
					Spoiler bool   `json:"spoiler"`
					Video   bool   `json:"is_video"`
					Hint    string `json:"post_hint"`
					Thumb   string `json:"thumbnail"`
					ThumbW  int    `json:"thumbnail_width"`
					ThumbH  int    `json:"thumbnail_height"`
					Preview struct {
						Images []struct {
							Source      image   `json:"source"`
							Resolutions []image `json:"resolutions"`
						} `json:"images"`
					} `json:"preview"`
				} `json:"data"`
			} `json:"children"`
		} `json:"data"`
	}
	if err := s.getJSON(ctx, "https://api.reddit.com/api/info/?raw_json=1&id=t3_"+p.ID, &o); err != nil {
		return Preview{}, err
	}
	if len(o.Data.Children) == 0 {
		return Preview{}, errNoPreview
	}
	d := o.Data.Children[0].Data
	if d.Title == "" {
		return Preview{}, errNoPreview
	}
	pv := Preview{
		URL: p.URL, Site: SiteReddit, SiteName: d.Sub, Title: clean(d.Title, maxTitle),
		Text: clean(d.Text, maxText), Video: d.Video || strings.Contains(d.Hint, "video"),
	}
	if d.Author != "" && d.Author != "[deleted]" {
		pv.Author = "u/" + d.Author
	}
	if pv.SiteName == "" {
		pv.SiteName = "Reddit"
	}
	if d.NSFW || d.Spoiler {
		return pv, nil
	}
	if imgs := d.Preview.Images; len(imgs) > 0 {
		pick := imgs[0].Source
		for _, r := range imgs[0].Resolutions {
			if r.Width >= 320 {
				pick = r
				break
			}
		}
		if _, ok := imageURL(pick.URL); ok {
			pv.Thumb, pv.Width, pv.Height = pick.URL, pick.Width, pick.Height
		}
	}
	if pv.Thumb == "" {
		if _, ok := imageURL(d.Thumb); ok {
			pv.Thumb, pv.Width, pv.Height = d.Thumb, d.ThumbW, d.ThumbH
		}
	}
	return pv, nil
}

func (s *Store) youTube(ctx context.Context, p Page) (Preview, error) {
	watch := "https://www.youtube.com/watch?v=" + p.ID
	var o struct {
		Title  string `json:"title"`
		Author string `json:"author_name"`
	}
	if err := s.getJSON(ctx, "https://www.youtube.com/oembed?format=json&url="+url.QueryEscape(watch), &o); err != nil {
		return Preview{}, err
	}
	return Preview{
		URL: p.URL, Site: SiteYouTube, SiteName: "YouTube", Video: true,
		Title: clean(o.Title, maxTitle), Author: clean(o.Author, maxTitle),
		Thumb: "https://i.ytimg.com/vi/" + p.ID + "/mqdefault.jpg", Width: 320, Height: 180,
	}, nil
}

func (s *Store) xPost(ctx context.Context, p Page) (Preview, error) {
	var o struct {
		Code  int `json:"code"`
		Tweet struct {
			Text   string `json:"text"`
			Author struct {
				Name   string `json:"name"`
				Handle string `json:"screen_name"`
			} `json:"author"`
			Media struct {
				All []struct {
					Type   string `json:"type"`
					URL    string `json:"url"`
					Thumb  string `json:"thumbnail_url"`
					Width  int    `json:"width"`
					Height int    `json:"height"`
				} `json:"all"`
			} `json:"media"`
		} `json:"tweet"`
	}
	if err := s.getJSON(ctx, "https://api.fxtwitter.com/status/"+p.ID, &o); err != nil {
		return Preview{}, err
	}
	if o.Code != 0 && o.Code != http.StatusOK {
		return Preview{}, errNoPreview
	}
	t := o.Tweet
	pv := Preview{URL: p.URL, Site: SiteX, SiteName: "X", Text: clean(t.Text, maxText), Title: clean(t.Author.Name, maxTitle)}
	if t.Author.Handle != "" {
		pv.Author = "@" + t.Author.Handle
	}
	if pv.Title == "" && pv.Text == "" {
		return Preview{}, errNoPreview
	}
	for _, m := range t.Media.All {
		pic := m.URL
		if m.Type != "photo" {
			pic, pv.Video = m.Thumb, true
		}
		if _, ok := imageURL(pic); ok {
			pv.Thumb, pv.Width, pv.Height = pic, m.Width, m.Height
			break
		}
	}
	return pv, nil
}

var (
	metaRE  = regexp.MustCompile(`(?is)<meta\s[^>]*>`)
	attrRE  = regexp.MustCompile(`(?is)([a-z:_-]+)\s*=\s*("[^"]*"|'[^']*'|[^\s>]+)`)
	titleRE = regexp.MustCompile(`(?is)<title[^>]*>(.*?)</title>`)
)

func (s *Store) webPage(ctx context.Context, p Page) (Preview, error) {
	pv, tagged, err := s.webPageAs(ctx, p, agentCrawler)
	if (err == nil && tagged) || ctx.Err() != nil {
		return pv, err
	}
	again, tagged2, err2 := s.webPageAs(ctx, p, agentBrowser)
	if err2 == nil && (tagged2 || err != nil) {
		return again, nil
	}
	return pv, err
}

func (s *Store) webPageAs(ctx context.Context, p Page, agent string) (Preview, bool, error) {
	body, typ, err := s.get(ctx, p.URL, maxPage, "text/html,application/xhtml+xml;q=0.9,*/*;q=0.5", agent, true)
	if err != nil {
		return Preview{}, false, err
	}
	if mt, _, _ := mime.ParseMediaType(typ); mt != "text/html" && mt != "application/xhtml+xml" {
		return Preview{}, false, errNoPreview
	}
	page := string(body)
	if i := strings.Index(strings.ToLower(page), "</head>"); i >= 0 {
		page = page[:i]
	}
	meta := map[string]string{}
	for _, tag := range metaRE.FindAllString(page, -1) {
		var key, content string
		for _, a := range attrRE.FindAllStringSubmatch(tag, -1) {
			v := strings.Trim(a[2], `"'`)
			switch strings.ToLower(a[1]) {
			case "property", "name":
				key = strings.ToLower(v)
			case "content":
				content = v
			}
		}
		if key != "" && content != "" {
			if _, seen := meta[key]; !seen {
				meta[key] = html.UnescapeString(content)
			}
		}
	}
	first := func(keys ...string) string {
		for _, k := range keys {
			if v := strings.TrimSpace(meta[k]); v != "" {
				return v
			}
		}
		return ""
	}
	u, _ := url.Parse(p.URL)
	pv := Preview{
		URL: p.URL, Site: SiteWeb,
		SiteName: clean(first("og:site_name"), maxTitle),
		Title:    clean(first("og:title", "twitter:title"), maxTitle),
		Text:     clean(first("og:description", "twitter:description", "description"), maxText),
		Video:    strings.HasPrefix(first("og:type"), "video"),
	}
	tagged := pv.Title != "" || pv.Text != ""
	if pv.Title == "" {
		if m := titleRE.FindStringSubmatch(page); m != nil {
			pv.Title = clean(html.UnescapeString(m[1]), maxTitle)
		}
	}
	if pv.SiteName == "" {
		pv.SiteName = strings.TrimPrefix(u.Hostname(), "www.")
	}
	if pv.Title == "" && pv.Text == "" {
		return Preview{}, false, errNoPreview
	}
	if img := first("og:image:secure_url", "og:image", "og:image:url", "twitter:image", "twitter:image:src"); img != "" {
		if ref, err := u.Parse(img); err == nil {
			if _, ok := imageURL(ref.String()); ok {
				pv.Thumb = ref.String()
				fmt.Sscan(meta["og:image:width"], &pv.Width)
				fmt.Sscan(meta["og:image:height"], &pv.Height)
			}
		}
	}
	return pv, tagged, nil
}

func (s *Store) getJSON(ctx context.Context, rawURL string, v any) error {
	body, _, err := s.get(ctx, rawURL, maxJSON, "application/json", agentCrawler, false)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, v)
}

func (s *Store) get(ctx context.Context, rawURL string, limit int64, accept, agent string, cut bool) ([]byte, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("User-Agent", agent)
	req.Header.Set("Accept", accept)
	req.Header.Set("Accept-Language", "en")
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("media: HTTP %d", resp.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return nil, "", err
	}
	if int64(len(data)) > limit {
		if !cut {
			return nil, "", errTooBig
		}
		data = data[:limit]
	}
	return data, resp.Header.Get("Content-Type"), nil
}

func clean(s string, n int) string {
	s = strings.Join(strings.Fields(s), " ")
	if !utf8.ValidString(s) {
		s = strings.ToValidUTF8(s, "")
	}
	if r := []rune(s); len(r) > n {
		s = strings.TrimSpace(string(r[:n-1])) + "…"
	}
	return s
}
