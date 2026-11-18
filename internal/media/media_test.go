package media

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

const (
	giphyBB    = `[img src="https://giphy.com/gifs/ai-cat-meme-774uK11mxE3W0cjE8L" thumbnail_src="https://media2.giphy.com/media/v1.Y2lk/774uK11mxE3W0cjE8L/giphy.gif" height="480" width="480" giphy_search="cat" title="High Five Ai GIF"][url=https://media2.giphy.com/media/v1.Y2lk/774uK11mxE3W0cjE8L/giphy.gif]https://media2.giphy.com/media/v1.Y2lk/774uK11mxE3W0cjE8L/giphy.gif[/url][/img]`
	giphyPlain = `https://media2.giphy.com/media/v1.Y2lk/774uK11mxE3W0cjE8L/giphy.gif`

	uploadBB    = `[img src=https://images.steamusercontent.com/ugc/14305680125121194903/2AF8350DCEE66C620D025AB0A00F30C4576AE987/ thumbnail_src=https://images.steamusercontent.com/ugc/14305680125121194903/2AF8350DCEE66C620D025AB0A00F30C4576AE987/?imw=512&&ima=fit&impolicy=Letterbox&imcolor=%23000000&letterbox=false srcset="https://images.steamusercontent.com/ugc/1/A/?imw=1024 1024w, https://images.steamusercontent.com/ugc/1/A/?imw=1920 1920w" width=3840 height=2160][url=https://images.steamusercontent.com/ugc/14305680125121194903/2AF8350DCEE66C620D025AB0A00F30C4576AE987/]https://images.steamusercontent.com/ugc/14305680125121194903/2AF8350DCEE66C620D025AB0A00F30C4576AE987/[/url][/img]`
	uploadPlain = `https://images.steamusercontent.com/ugc/14305680125121194903/2AF8350DCEE66C620D025AB0A00F30C4576AE987/`
)

func TestFindGiphy(t *testing.T) {
	f, ok := Find(giphyBB, giphyPlain)
	if !ok {
		t.Fatal("no picture found")
	}
	if f.Kind != KindGIF || !f.Trusted || f.Width != 480 || f.Height != 480 || f.Title != "High Five Ai GIF" {
		t.Fatalf("got %+v", f)
	}
	if f.URL != giphyPlain {
		t.Fatalf("loads %q, want the GIF itself", f.URL)
	}
	if f.Full != "" {
		t.Fatalf("original %q, want none (src is a Giphy page)", f.Full)
	}
	if got := StripText(giphyPlain, f); got != "" {
		t.Fatalf("text left: %q", got)
	}
}

func TestFindUpload(t *testing.T) {
	f, ok := Find(uploadBB, uploadPlain)
	if !ok {
		t.Fatal("no picture found")
	}
	if f.Kind != KindUpload || !f.Trusted || f.Width != 3840 || f.Height != 2160 {
		t.Fatalf("got %+v", f)
	}
	if !strings.Contains(f.URL, "imw=512") {
		t.Fatalf("loads %q, want the 512 px thumbnail", f.URL)
	}
	if f.Full != uploadPlain {
		t.Fatalf("original %q, want %q", f.Full, uploadPlain)
	}
	if got := StripText(uploadPlain, f); got != "" {
		t.Fatalf("text left: %q", got)
	}
}

func TestFindLinks(t *testing.T) {
	cases := []struct {
		plain   string
		want    bool
		trusted bool
		left    string
	}{
		{"https://i.imgur.com/abc.png", true, false, ""},
		{"look https://example.com/cat.JPG!", true, false, "look https://example.com/cat.JPG!"},
		{"https://cdn.discordapp.com/attachments/1/2/a.webp?ex=1", true, false, ""},
		{"https://media1.giphy.com/media/x/giphy.gif", true, true, ""},
		{"https://example.com/page", false, false, ""},
		{"ftp://example.com/a.png", false, false, ""},
		{"no links here", false, false, ""},
	}
	for _, c := range cases {
		f, ok := Find(c.plain, c.plain)
		if ok != c.want {
			t.Errorf("%q: found=%v, want %v", c.plain, ok, c.want)
			continue
		}
		if !ok {
			continue
		}
		if f.Kind != KindLink || f.Trusted != c.trusted {
			t.Errorf("%q: got %+v", c.plain, f)
		}
		if got := StripText(c.plain, f); got != c.left {
			t.Errorf("%q: text left %q, want %q", c.plain, got, c.left)
		}
	}
}

func TestFindPage(t *testing.T) {
	cases := []struct {
		plain string
		ok    bool
		site  string
		id    string
		strip string
	}{
		{"https://www.youtube.com/watch?v=dQw4w9WgXcQ&t=10", true, SiteYouTube, "dQw4w9WgXcQ", "https://www.youtube.com/watch?v=dQw4w9WgXcQ&t=10"},
		{"watch https://youtu.be/dQw4w9WgXcQ!", true, SiteYouTube, "dQw4w9WgXcQ", ""},
		{"https://youtube.com/shorts/dQw4w9WgXcQ", true, SiteYouTube, "dQw4w9WgXcQ", "https://youtube.com/shorts/dQw4w9WgXcQ"},
		{"https://x.com/jack/status/20", true, SiteX, "20", "https://x.com/jack/status/20"},
		{"lol https://vxtwitter.com/jack/status/20/photo/1", true, SiteX, "20", ""},
		{"https://twitter.com/i/status/123", true, SiteX, "123", "https://twitter.com/i/status/123"},
		{"https://x.com/jack", true, SiteWeb, "", "https://x.com/jack"},
		{"https://www.reddit.com/r/computadores/comments/1wis7x7/esse_pc_consegue_rodar_gta_6/", true, SiteReddit, "1wis7x7", "https://www.reddit.com/r/computadores/comments/1wis7x7/esse_pc_consegue_rodar_gta_6/"},
		{"see https://old.reddit.com/comments/1wis7x7", true, SiteReddit, "1wis7x7", ""},
		{"https://redd.it/1wis7x7", true, SiteReddit, "1wis7x7", "https://redd.it/1wis7x7"},
		{"https://www.reddit.com/r/cs2/s/AbCdEf123", true, SiteWeb, "", "https://www.reddit.com/r/cs2/s/AbCdEf123"},
		{"https://example.com/page", true, SiteWeb, "", "https://example.com/page"},
		{"https://i.imgur.com/abc.png", false, "", "", ""},
		{"no links here", false, "", "", ""},
	}
	for _, c := range cases {
		p, ok := FindPage(c.plain)
		if ok != c.ok || p.Site != c.site || p.ID != c.id || p.Strip != c.strip {
			t.Errorf("%q: got %+v, %v", c.plain, p, ok)
		}
	}
}

func TestPlainHTTPIsNotTrusted(t *testing.T) {
	f, ok := Find("http://images.steamusercontent.com/ugc/1/a.png", "http://images.steamusercontent.com/ugc/1/a.png")
	if !ok || f.Trusted {
		t.Fatalf("got %+v, %v", f, ok)
	}
}

func pngBytes(t *testing.T, w, h int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	img.Set(0, 0, color.White)
	var b bytes.Buffer
	if err := png.Encode(&b, img); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}

func TestStoreFetchesAndServes(t *testing.T) {
	pic := pngBytes(t, 3, 2)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/a.png":
			w.Header().Set("Content-Type", "text/plain")
			_, _ = w.Write(pic)
		case "/fake.png":
			_, _ = w.Write([]byte("<html>not a picture</html>"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	s := NewStore()
	s.client = srv.Client()

	id := s.Start(srv.URL + "/a.png")
	if again := s.Start(srv.URL + "/a.png"); again != id {
		t.Fatal("same address got a new id")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	got, err := s.Wait(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if got.Type != "image/png" || got.Width != 3 || got.Height != 2 {
		t.Fatalf("got %s %dx%d", got.Type, got.Width, got.Height)
	}

	if _, err := s.Wait(ctx, s.Start(srv.URL+"/fake.png")); err == nil {
		t.Fatal("a web page passed as a picture")
	}

	h := s.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(299) }))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", Prefix+id, nil))
	if rec.Code != 200 || rec.Header().Get("Content-Type") != "image/png" || !bytes.Equal(rec.Body.Bytes(), pic) {
		t.Fatalf("served %d %q", rec.Code, rec.Header().Get("Content-Type"))
	}
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", Prefix+"nope", nil))
	if rec.Code != 404 {
		t.Fatalf("unknown id: %d", rec.Code)
	}
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/index.html", nil))
	if rec.Code != 299 {
		t.Fatal("other paths are not passed on")
	}
}

func TestStoreLaterFetchesWhenAsked(t *testing.T) {
	var hits atomic.Int32
	pic := pngBytes(t, 4, 4)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		_, _ = w.Write(pic)
	}))
	defer srv.Close()

	s := NewStore()
	s.client = srv.Client()
	id := s.Later(srv.URL + "/full.png")
	time.Sleep(50 * time.Millisecond)
	if hits.Load() != 0 {
		t.Fatal("fetched before anyone asked")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	got, err := s.Wait(ctx, id)
	if err != nil || got.Width != 4 || hits.Load() != 1 {
		t.Fatalf("got %dx%d, %v, %d fetches", got.Width, got.Height, err, hits.Load())
	}
}

func TestStoreRefusesLocalAddresses(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(pngBytes(t, 1, 1))
	}))
	defer srv.Close()
	s := NewStore()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := s.Wait(ctx, s.Start(srv.URL+"/a.png")); err == nil {
		t.Fatal("fetched from 127.0.0.1")
	}
}
