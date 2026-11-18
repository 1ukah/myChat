package steam

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"errors"
	"math/big"
	"sync"
	"time"
)

type GuardType int32

const (
	GuardNone               GuardType = 1
	GuardEmailCode          GuardType = 2
	GuardDeviceCode         GuardType = 3
	GuardDeviceConfirmation GuardType = 4
	GuardEmailConfirmation  GuardType = 5
	GuardMachineToken       GuardType = 6
)

type Guard struct {
	Type    GuardType
	Message string
}

const (
	eresultDuplicateRequest  = 29
	eresultInvalidAuthCode   = 65
	eresultTwoFactorMismatch = 88
)

type LoginError struct {
	Message string
	Err     error
}

func (e *LoginError) Error() string { return e.Message }
func (e *LoginError) Unwrap() error { return e.Err }

func friendly(err error) error {
	var se *Error
	if !errors.As(err, &se) {
		return err
	}
	switch se.EResult {
	case EResultInvalidPassword:
		return &LoginError{"Wrong account name or password.", err}
	case EResultRateLimitExceeded:
		return &LoginError{"Too many login attempts. Wait a while and try again.", err}
	case eresultInvalidAuthCode, eresultTwoFactorMismatch:
		return &LoginError{"Wrong Steam Guard code.", err}
	case EResultExpired, EResultFileNotFound:
		return &LoginError{"The login took too long. Start again.", err}
	}
	return err
}

type CredentialLogin struct {
	Account string
	Guards  []Guard

	c         *conn
	clientID  uint64
	requestID []byte
	steamID   uint64
	interval  time.Duration
	closeOnce sync.Once
}

func StartCredentials(ctx context.Context, deviceName, account, password, guardData string) (*CredentialLogin, error) {
	c, err := dialAny(ctx)
	if err != nil {
		return nil, err
	}
	l := &CredentialLogin{Account: account, c: c}
	if err := l.begin(ctx, deviceName, account, password, guardData); err != nil {
		c.close()
		return nil, friendly(err)
	}
	return l, nil
}

func (l *CredentialLogin) begin(ctx context.Context, deviceName, account, password, guardData string) error {
	var hello enc
	hello.varint(1, protocolVersion)
	if err := l.c.send(emsgClientHello, newHeader(), hello.b); err != nil {
		return err
	}

	var kreq enc
	kreq.string(1, account)
	p, err := l.c.call(ctx, emsgServiceMethodCallFromClientNonAuthed, "Authentication.GetPasswordRSAPublicKey#1", kreq.b)
	if err != nil {
		return err
	}
	var modHex, expHex string
	var ts uint64
	_ = walk(p.body, func(f field) {
		switch f.Num {
		case 1:
			modHex = f.str()
		case 2:
			expHex = f.str()
		case 3:
			ts = f.Value
		}
	})
	mod, ok1 := new(big.Int).SetString(modHex, 16)
	exp, ok2 := new(big.Int).SetString(expHex, 16)
	if !ok1 || !ok2 || !exp.IsInt64() {
		return errors.New("steam: bad RSA key")
	}
	pub := &rsa.PublicKey{N: mod, E: int(exp.Int64())}
	ct, err := rsa.EncryptPKCS1v15(rand.Reader, pub, []byte(password))
	if err != nil {
		return err
	}

	var details enc
	details.string(1, deviceName)
	details.varint(2, 1)
	details.int32(3, 20)
	details.varint(4, 1)
	var req enc
	req.string(1, deviceName)
	req.string(2, account)
	req.string(3, base64.StdEncoding.EncodeToString(ct))
	req.varint(4, ts)
	req.bool(5, true)
	req.varint(6, 1)
	req.varint(7, 1)
	req.string(8, "Client")
	req.bytes(9, details.b)
	if guardData != "" {
		req.string(10, guardData)
	}
	p, err = l.c.call(ctx, emsgServiceMethodCallFromClientNonAuthed, "Authentication.BeginAuthSessionViaCredentials#1", req.b)
	if err != nil {
		return err
	}
	interval := float32(5)
	_ = walk(p.body, func(f field) {
		switch f.Num {
		case 1:
			l.clientID = f.Value
		case 2:
			l.requestID = append([]byte(nil), f.Bytes...)
		case 3:
			interval = f.f32()
		case 4:
			var g Guard
			_ = walk(f.Bytes, func(h field) {
				switch h.Num {
				case 1:
					g.Type = GuardType(h.i32())
				case 2:
					g.Message = h.str()
				}
			})
			l.Guards = append(l.Guards, g)
		case 5:
			l.steamID = f.Value
		}
	})
	if l.clientID == 0 {
		return errors.New("steam: login was not started")
	}
	l.interval = time.Duration(float64(interval) * float64(time.Second))
	if l.interval < time.Second || l.interval > 10*time.Second {
		l.interval = 5 * time.Second
	}
	return nil
}

func (l *CredentialLogin) Needs(t GuardType) bool {
	for _, g := range l.Guards {
		if g.Type == t {
			return true
		}
	}
	return false
}

func (l *CredentialLogin) SubmitCode(ctx context.Context, t GuardType, code string) error {
	var req enc
	req.varint(1, l.clientID)
	req.fixed64(2, l.steamID)
	req.string(3, code)
	req.int32(4, int32(t))
	_, err := l.c.call(ctx, emsgServiceMethodCallFromClientNonAuthed, "Authentication.UpdateAuthSessionWithSteamGuardCode#1", req.b)
	var se *Error
	if errors.As(err, &se) && se.EResult == eresultDuplicateRequest {
		return nil
	}
	return friendly(err)
}

func (l *CredentialLogin) Wait(ctx context.Context) (refreshToken, account, guardData string, err error) {
	deadline := time.Now().Add(5 * time.Minute)
	first := true
	for time.Now().Before(deadline) {
		if !first {
			select {
			case <-ctx.Done():
				return "", "", "", ctx.Err()
			case <-l.c.done():
				return "", "", "", errClosed
			case <-time.After(l.interval):
			}
		}
		first = false
		r, err := pollStatus(ctx, l.c, l.clientID, l.requestID)
		if err != nil {
			return "", "", "", friendly(err)
		}
		if r.clientID != 0 {
			l.clientID = r.clientID
		}
		if r.refresh != "" {
			if r.account == "" {
				r.account = l.Account
			}
			return r.refresh, r.account, r.guardData, nil
		}
	}
	return "", "", "", &LoginError{"The login took too long. Start again.", nil}
}

func (l *CredentialLogin) Close() { l.closeOnce.Do(l.c.close) }
