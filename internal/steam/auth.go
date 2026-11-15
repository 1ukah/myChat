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

// QRLogin runs Steam's QR code login. onChallenge is called with each QR URL
// (the URL changes when an old code expires). The user scans the code with the
// Steam mobile app. The returned refresh token can log on to the Steam client
// network, so the password is never seen by this app.
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
		// An expired or unknown session gets a fresh QR code.
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
	details.varint(2, 1) // EAuthTokenPlatformType SteamClient
	details.int32(3, 20) // EOSType Win11
	details.varint(4, 1) // gaming device type: desktop PC
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

// TokenInfo is the readable part of a Steam refresh token (a JWT).
type TokenInfo struct {
	SteamID  uint64
	Expires  time.Time
	Audience []string
}

// ParseToken reads the SteamID, expiry and audience from a refresh token.
// It does not verify the signature; Steam does that on logon.
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
