package steam

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"
)

func QRLogin(ctx context.Context, deviceName string, onChallenge func(url string)) (refreshToken, accountName string, err error) {
	for attempt := 0; attempt < 20; attempt++ {
		token, account, err := qrAttempt(ctx, deviceName, onChallenge)
		if err == nil {
			return token, account, nil
		}
		var se *Error
		if ctx.Err() != nil {
			return "", "", ctx.Err()
		}
		if errors.As(err, &se) && (se.EResult == EResultFileNotFound || se.EResult == EResultExpired) {
			continue
		}
		if errors.Is(err, errQRExpired) {
			continue
		}
		return "", "", err
	}
	return "", "", errors.New("steam: QR login timed out")
}

var errQRExpired = errors.New("steam: QR code expired")

func qrAttempt(ctx context.Context, deviceName string, onChallenge func(string)) (string, string, error) {
	c, err := dialAny(ctx)
	if err != nil {
		return "", "", err
	}
	defer c.close()

	var hello enc
	hello.varint(1, protocolVersion)
	if err := c.send(emsgClientHello, newHeader(), hello.b); err != nil {
		return "", "", err
	}

	var details enc
	details.string(1, deviceName)
	details.varint(2, 1)
	details.int32(3, 20)
	details.varint(4, 1)
	var req enc
	req.bytes(3, details.b)

	p, err := c.call(ctx, emsgServiceMethodCallFromClientNonAuthed, "Authentication.BeginAuthSessionViaQR#1", req.b)
	if err != nil {
		return "", "", err
	}
	var clientID uint64
	var requestID []byte
	var challenge string
	interval := float32(5)
	_ = walk(p.body, func(f field) {
		switch f.Num {
		case 1:
			clientID = f.Value
		case 2:
			challenge = f.str()
		case 3:
			requestID = append([]byte(nil), f.Bytes...)
		case 4:
			interval = f.f32()
		}
	})
	if challenge == "" || clientID == 0 {
		return "", "", errors.New("steam: empty QR challenge")
	}
	onChallenge(challenge)

	wait := time.Duration(float64(interval) * float64(time.Second))
	if wait < time.Second || wait > 10*time.Second {
		wait = 5 * time.Second
	}
	deadline := time.Now().Add(3 * time.Minute)
	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return "", "", ctx.Err()
		case <-c.done():
			return "", "", errClosed
		case <-time.After(wait):
		}
		r, err := pollStatus(ctx, c, clientID, requestID)
		if err != nil {
			return "", "", err
		}
		if r.clientID != 0 {
			clientID = r.clientID
		}
		if r.newURL != "" {
			onChallenge(r.newURL)
		}
		if r.refresh != "" {
			return r.refresh, r.account, nil
		}
	}
	return "", "", errQRExpired
}

type TokenInfo struct {
	SteamID  uint64
	Expires  time.Time
	Audience []string
}

func ParseToken(token string) (TokenInfo, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return TokenInfo{}, errors.New("steam: token is not a JWT")
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return TokenInfo{}, err
	}
	var claims struct {
		Sub string          `json:"sub"`
		Exp int64           `json:"exp"`
		Aud json.RawMessage `json:"aud"`
	}
	if err := json.Unmarshal(raw, &claims); err != nil {
		return TokenInfo{}, err
	}
	id, err := strconv.ParseUint(claims.Sub, 10, 64)
	if err != nil {
		return TokenInfo{}, errors.New("steam: token has no SteamID")
	}
	info := TokenInfo{SteamID: id, Expires: time.Unix(claims.Exp, 0)}
	var list []string
	if json.Unmarshal(claims.Aud, &list) != nil {
		var one string
		if json.Unmarshal(claims.Aud, &one) == nil {
			list = []string{one}
		}
	}
	info.Audience = list
	return info, nil
}

type pollResult struct {
	clientID  uint64
	newURL    string
	refresh   string
	account   string
	guardData string
}

func pollStatus(ctx context.Context, c *conn, clientID uint64, requestID []byte) (pollResult, error) {
	var poll enc
	poll.varint(1, clientID)
	poll.bytes(2, requestID)
	p, err := c.call(ctx, emsgServiceMethodCallFromClientNonAuthed, "Authentication.PollAuthSessionStatus#1", poll.b)
	if err != nil {
		return pollResult{}, err
	}
	var r pollResult
	_ = walk(p.body, func(f field) {
		switch f.Num {
		case 1:
			r.clientID = f.Value
		case 2:
			r.newURL = f.str()
		case 3:
			r.refresh = f.str()
		case 6:
			r.account = f.str()
		case 7:
			r.guardData = f.str()
		}
	})
	return r, nil
}
