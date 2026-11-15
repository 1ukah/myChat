package main

import (
	"context"
	"errors"
	"log"
	"os"
	"strings"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"

	"mychat/internal/config"
	"mychat/internal/steam"
	"mychat/internal/steamlocal"
)

// Login sources.
const (
	SourceSteam = "steam" // the login saved by the Steam client on this PC
	SourceQR    = "qr"    // myChat's own session from a QR login
)

// EvLocalLogins reports the accounts saved by the Steam client.
const EvLocalLogins = "mychat:local-logins"

// LoginState drives the login screen.
type LoginState struct {
	Phase   string   `json:"phase"`  // idle | waiting (QR shown) | done | error
	Method  string   `json:"method"` // "steam" | "qr" | ""
	URL     string   `json:"url"`
	QR      []string `json:"qr"` // rows of '1' (dark) and '0' (light)
	Error   string   `json:"error"`
	Account string   `json:"account"`
}

// LocalLogins lists the accounts the Steam client on this PC knows.
type LocalLogins struct {
	SteamInstalled bool                 `json:"steamInstalled"`
	Accounts       []steamlocal.Account `json:"accounts"`
	// Watching is true after OpenSteamLogin while myChat waits for you to
	// sign in to Steam. The new login is used automatically.
	Watching bool   `json:"watching"`
	Error    string `json:"error"`
}

func init() {
	application.RegisterEvent[LocalLogins](EvLocalLogins)
}

func (a *App) localLogins() LocalLogins {
	a.mu.Lock()
	watching := a.watchCancel != nil
	a.mu.Unlock()
	l := LocalLogins{Watching: watching, Accounts: []steamlocal.Account{}}
	if _, err := steamlocal.SteamDir(); err != nil {
		l.Error = "Steam is not installed on this PC."
		return l
	}
	l.SteamInstalled = true
	accs, err := steamlocal.Accounts()
	if err != nil {
		l.Error = err.Error()
	}
	if accs != nil {
		l.Accounts = accs
	}
	return l
}

func (a *App) emitLocalLogins() { a.emit(EvLocalLogins, a.localLogins()) }

// autoLogin runs at startup. Nothing is used without a choice the user made
// earlier; on first run the settings window opens on the login screen.
func (a *App) autoLogin() {
	a.mu.Lock()
	src, acct := a.cfg.LoginSource, a.cfg.SteamAccount
	a.mu.Unlock()
	switch src {
	case SourceSteam:
		if err := a.useSteam(acct); err != nil {
			a.needsLogin(err.Error())
		}
		return
	case SourceQR:
		if sess, err := a.store.LoadSession(); err == nil {
			a.mu.Lock()
			a.session = sess
			a.mu.Unlock()
			a.setStatus(func(st *Status) { st.AccountName, st.Source = sess.AccountName, SourceQR })
			if err := a.steam.Start(sess.RefreshToken); err == nil {
				a.setLogin(LoginState{Phase: "done", Method: SourceQR, Account: sess.AccountName})
				return
			}
			a.store.ClearSession()
		}
	}
	a.needsLogin("")
}

// needsLogin shows the login screen with an optional reason.
func (a *App) needsLogin(reason string) {
	a.setStatus(func(st *Status) {
		*st = Status{State: string(steam.StateNeedsLogin), Detail: reason}
	})
	a.setLogin(LoginState{Phase: "idle", Error: reason})
	a.emitLocalLogins()
	if os.Getenv("MYCHAT_DEMO") == "" {
		a.showSettings()
	}
}

// useSteam logs on with the login the Steam client saved for account. An
// empty account picks Steam's most recent account that has a saved login.
func (a *App) useSteam(account string) error {
	if account == "" {
		accs, err := steamlocal.Accounts()
		if err != nil {
			return err
		}
		for _, ac := range accs {
			if ac.HasLogin {
				account = ac.AccountName
				break
			}
		}
		if account == "" {
			return errors.New("Steam has no saved login on this PC. Sign in to Steam with \"Remember me\" on")
		}
	}
	tok, err := steamlocal.Token(account)
	if err != nil {
		return err
	}
	if err := a.steam.Start(tok); err != nil {
		return err
	}
	a.store.ClearSession() // a QR session is no longer needed
	a.mu.Lock()
	a.session = config.Session{}
	a.steamToken = tok
	a.mu.Unlock()
	a.setSource(SourceSteam, account)
	a.setStatus(func(st *Status) { st.AccountName, st.Source = account, SourceSteam })
	a.setLogin(LoginState{Phase: "done", Method: SourceSteam, Account: account})
	return nil
}

// setSource remembers the login choice for the next start.
func (a *App) setSource(src, account string) {
	a.mu.Lock()
	a.cfg.LoginSource, a.cfg.SteamAccount = src, account
	c := a.cfg
	view := a.configViewLocked()
	a.mu.Unlock()
	if err := a.store.Save(c); err != nil {
		log.Printf("save config: %v", err)
	}
	a.emit(EvConfig, view)
}

// onLoginExpired runs when Steam rejects the token. It is called from the
// event pump, so the work happens on another goroutine (Stop waits for the
// client loop, which may be sending to the pump).
func (a *App) onLoginExpired() {
	go func() {
		a.mu.Lock()
		src, acct, old := a.cfg.LoginSource, a.cfg.SteamAccount, a.steamToken
		a.mu.Unlock()
		if src == SourceSteam {
			// The Steam client may have renewed its saved login; use the new one.
			if tok, err := steamlocal.Token(acct); err == nil && tok != old {
				if a.useSteam(acct) == nil {
					return
				}
			}
			a.steam.Stop()
			a.engine.System("Steam's saved login expired. Sign in to Steam again, or open Settings.")
			a.needsLogin("Steam's saved login for " + acct + " has expired. Sign in to Steam again.")
			return
		}
		a.steam.Stop()
		a.store.ClearSession()
		a.engine.System("Steam login expired. Open Settings to log in again.")
		a.needsLogin("Your Steam login expired. Log in again.")
	}()
}

func (a *App) finishQR(token, account string) {
	sess := config.Session{RefreshToken: token, AccountName: account}
	if err := a.store.SaveSession(sess); err != nil {
		a.setLogin(LoginState{Phase: "error", Method: SourceQR, Error: err.Error()})
		return
	}
	if err := a.steam.Start(token); err != nil {
		a.setLogin(LoginState{Phase: "error", Method: SourceQR, Error: err.Error()})
		return
	}
	a.mu.Lock()
	a.session = sess
	a.steamToken = ""
	a.mu.Unlock()
	a.setSource(SourceQR, "")
	a.setStatus(func(st *Status) { st.AccountName, st.Source = account, SourceQR })
	a.setLogin(LoginState{Phase: "done", Method: SourceQR, Account: account})
}

func (a *App) stopWatchLocked() {
	if a.watchCancel != nil {
		a.watchCancel()
		a.watchCancel = nil
	}
}

func loginMessage(err error) string {
	var se *steam.Error
	if errors.As(err, &se) {
		return "Steam refused the login (" + err.Error() + ")."
	}
	return "Could not reach Steam: " + err.Error()
}

// ---- Service methods ----

// LocalLogins lists the accounts saved by the Steam client on this PC.
func (s *ChatService) LocalLogins() LocalLogins { return s.a.localLogins() }

// UseSteamAccount logs on with the login the Steam client saved for this
// account ("" = Steam's most recent account). The choice is remembered.
func (s *ChatService) UseSteamAccount(account string) error {
	a := s.a
	a.mu.Lock()
	if a.loginCancel != nil {
		a.loginCancel()
		a.loginCancel = nil
	}
	a.stopWatchLocked()
	a.mu.Unlock()
	err := a.useSteam(strings.TrimSpace(account))
	a.emitLocalLogins()
	if err != nil {
		a.setLogin(LoginState{Phase: "error", Method: SourceSteam, Account: account, Error: err.Error()})
	}
	return err
}

// OpenSteamLogin opens the Steam client. If no account is signed in there,
// Steam shows its own sign-in window (password, Steam Guard, QR). myChat
// watches for the new saved login for 10 minutes and uses it automatically.
func (s *ChatService) OpenSteamLogin() error {
	a := s.a
	if err := steamlocal.OpenSteam(); err != nil {
		return errors.New("could not open Steam")
	}
	before := map[string]bool{}
	if accs, err := steamlocal.Accounts(); err == nil {
		for _, ac := range accs {
			if ac.HasLogin {
				before[strings.ToLower(ac.AccountName)] = true
			}
		}
	}
	sig := steamlocal.Signature()

	a.mu.Lock()
	a.stopWatchLocked()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	a.watchCancel = cancel
	a.mu.Unlock()
	a.emitLocalLogins()

	go func() {
		defer func() {
			a.mu.Lock()
			if ctx.Err() == context.DeadlineExceeded {
				a.watchCancel = nil
			}
			a.mu.Unlock()
			cancel()
			a.emitLocalLogins()
		}()
		t := time.NewTicker(2 * time.Second)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
			}
			if cur := steamlocal.Signature(); cur == sig {
				continue
			} else {
				sig = cur
			}
			accs, err := steamlocal.Accounts()
			if err != nil {
				continue
			}
			a.emitLocalLogins()
			for _, ac := range accs {
				if ac.HasLogin && !before[strings.ToLower(ac.AccountName)] {
					a.mu.Lock()
					a.watchCancel = nil
					a.mu.Unlock()
					if err := a.useSteam(ac.AccountName); err != nil {
						a.setLogin(LoginState{Phase: "error", Method: SourceSteam, Account: ac.AccountName, Error: err.Error()})
					}
					return
				}
			}
		}
	}()
	return nil
}

// StartLogin begins a QR login. Progress arrives as mychat:login events.
func (s *ChatService) StartLogin() {
	a := s.a
	a.mu.Lock()
	if a.loginCancel != nil {
		a.loginCancel()
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	a.loginCancel = cancel
	a.mu.Unlock()

	a.setLogin(LoginState{Phase: "waiting", Method: SourceQR})
	go func() {
		defer cancel()
		token, account, err := steam.QRLogin(ctx, "myChat", func(url string) {
			a.setLogin(LoginState{Phase: "waiting", Method: SourceQR, URL: url, QR: qrRows(url)})
		})
		if err != nil {
			if ctx.Err() != nil {
				return // cancelled or replaced by another login
			}
			a.setLogin(LoginState{Phase: "error", Method: SourceQR, Error: loginMessage(err)})
			return
		}
		a.finishQR(token, account)
	}()
}

// CancelLogin stops a QR login and the wait for a Steam sign-in.
func (s *ChatService) CancelLogin() {
	a := s.a
	a.mu.Lock()
	if a.loginCancel != nil {
		a.loginCancel()
		a.loginCancel = nil
	}
	a.stopWatchLocked()
	a.mu.Unlock()
	a.setLogin(LoginState{Phase: "idle"})
	a.emitLocalLogins()
}

// Logout ends myChat's Steam session and forgets the login choice. With the
// Steam source this does not sign the Steam client out.
func (s *ChatService) Logout() {
	a := s.a
	a.steam.Stop()
	a.store.ClearSession()
	a.mu.Lock()
	a.session = config.Session{}
	a.steamToken = ""
	a.stopWatchLocked()
	a.mu.Unlock()
	a.setSource("", "")
	a.setStatus(func(st *Status) {
		*st = Status{State: string(steam.StateNeedsLogin)}
	})
	a.setLogin(LoginState{Phase: "idle"})
	a.emitLocalLogins()
}
