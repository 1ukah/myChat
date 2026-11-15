package steam

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"
)

// TestLiveQRChallenge talks to the real Steam network. Run with
// MYCHAT_LIVE=1 go test ./internal/steam -run Live -v
func TestLiveQRChallenge(t *testing.T) {
	if os.Getenv("MYCHAT_LIVE") == "" {
		t.Skip("set MYCHAT_LIVE=1 to run")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	got := make(chan string, 1)
	go func() {
		_, _, _ = QRLogin(ctx, "myChat-test", func(url string) {
			select {
			case got <- url:
			default:
			}
		})
	}()
	select {
	case url := <-got:
		t.Logf("challenge URL: %s", url)
	case <-ctx.Done():
		t.Fatal("no QR challenge from Steam")
	}
}

// TestLiveCredentialsWrongPassword checks the RSA key and login request
// encoding with an account that does not exist.
func TestLiveCredentialsWrongPassword(t *testing.T) {
	if os.Getenv("MYCHAT_LIVE") == "" {
		t.Skip("set MYCHAT_LIVE=1 to run")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	l, err := StartCredentials(ctx, "myChat-test", "mychat_nonexistent_test_user_7341", "definitely-wrong", "")
	if l != nil {
		l.Close()
		t.Fatalf("expected an error, got guards %+v", l.Guards)
	}
	var le *LoginError
	if !errors.As(err, &le) {
		t.Fatalf("expected LoginError, got %v", err)
	}
	t.Logf("Steam said: %s (%v)", le.Message, le.Err)
}
