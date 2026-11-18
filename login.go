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

const (
	SourceSteam    = "steam"
	SourceQR       = "qr"
	SourcePassword = "password"
)

const EvLocalLogins = "mychat:local-logins"

type LoginState struct {
	Phase     string   `json:"phase"`
	Method    string   `json:"method"`
	URL       string   `json:"url"`
	QR        []string `json:"qr"`
	Error     string   `json:"error"`
	Account   string   `json:"account"`
	Guards    []string `json:"guards"`
	EmailHint string   `json:"emailHint"`
	CodeError string   `json:"codeError"`
	Checking  bool     `json:"checking"`
}

type LocalLogins struct {
	SteamInstalled bool                 `json:"steamInstalled"`
	Accounts       []steamlocal.Account `json:"accounts"`
	Watching       bool                 `json:"watching"`
	Error          string               `json:"error"`
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

func (a *App) autoLogin() {
	a.mu.Lock()
	src, acct := a.cfg.LoginSource, a.cfg.SteamAccount
	a.mu.Unlock()
	if src == SourceSteam {
		if err := a.useSteam(acct); err != nil {
			a.needsLogin(err.Error())
		}
		return
	}
	sess, err := a.store.LoadSession()
	if err != nil {
		a.needsLogin("")
		return
	}
	if sess.Method == "" {
		sess.Method = SourceQR
	}
	a.mu.Lock()
	a.session = sess
	a.mu.Unlock()
	a.setStatus(func(st *Status) { st.AccountName, st.Source = sess.AccountName, sess.Method })
	if err := a.steam.Start(sess.RefreshToken); err != nil {
		a.store.ClearSession()
		a.needsLogin("")
		return
	}
	a.setLogin(LoginState{Phase: "done", Method: sess.Method, Account: sess.AccountName})
}

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
	a.cancelLogins()
	if err := a.steam.Start(tok); err != nil {
		return err
	}
	a.store.ClearSession()
	a.mu.Lock()
	a.session = config.Session{}
	a.steamToken = tok
	a.mu.Unlock()
	a.setSource(SourceSteam, account)
	a.setStatus(func(st *Status) { st.AccountName, st.Source = account, SourceSteam })
	a.setLogin(LoginState{Phase: "done", Method: SourceSteam, Account: account})
	return nil
}

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

func (a *App) onLoginExpired() {
	go func() {
		a.mu.Lock()
		src, acct, old := a.cfg.LoginSource, a.cfg.SteamAccount, a.steamToken
		a.mu.Unlock()
		if src == SourceSteam {
			if tok, err := steamlocal.Token(acct); err == nil && tok != old {
				if a.useSteam(acct) == nil {
					return
				}
			}
			a.steam.Stop()
			a.engine.System("Steam login expired. Sign in again in Settings.")
			a.needsLogin("The saved login for " + acct + " expired. Sign in to Steam again.")
			return
		}
		a.steam.Stop()
		a.store.ClearSession()
		a.mu.Lock()
		a.session = config.Session{}
		a.mu.Unlock()
		a.engine.System("Steam login expired. Sign in again in Settings.")
		a.needsLogin("Your Steam login expired. Sign in again.")
	}()
}

func (a *App) finishLogin(method, token, account, guardData string) {
	a.store.SaveGuardData(account, guardData)
	sess := config.Session{RefreshToken: token, AccountName: account, Method: method}
	if err := a.store.SaveSession(sess); err != nil {
		a.setLogin(LoginState{Phase: "error", Method: method, Error: err.Error()})
		return
	}
	if err := a.steam.Start(token); err != nil {
		a.store.ClearSession()
		a.setLogin(LoginState{Phase: "error", Method: method, Error: err.Error()})
		return
	}
	a.mu.Lock()
	a.session = sess
	a.steamToken = ""
	a.mu.Unlock()
	a.setSource(method, "")
	a.setStatus(func(st *Status) { st.AccountName, st.Source = account, method })
	a.setLogin(LoginState{Phase: "done", Method: method, Account: account})
}

func (a *App) cancelLogins() {
	a.mu.Lock()
	if a.loginCancel != nil {
		a.loginCancel()
		a.loginCancel = nil
	}
	a.credLogin = nil
	a.stopWatchLocked()
	a.mu.Unlock()
}

func (a *App) beginLogin() (context.Context, context.CancelFunc) {
	a.cancelLogins()
	a.emitLocalLogins()
	a.mu.Lock()
	defer a.mu.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	a.loginCancel = cancel
	return ctx, cancel
}

func (a *App) stopWatchLocked() {
	if a.watchCancel != nil {
		a.watchCancel()
		a.watchCancel = nil
	}
}

func loginMessage(err error) string {
	var le *steam.LoginError
	if errors.As(err, &le) {
		return le.Message
	}
	var se *steam.Error
	if errors.As(err, &se) {
		return "Steam refused the login (" + err.Error() + ")."
	}
	return "Could not reach Steam: " + err.Error()
}

func guardNames(gs []steam.Guard) (names []string, emailHint string) {
	for _, g := range gs {
		switch g.Type {
		case steam.GuardDeviceCode:
			names = append(names, "device_code")
		case steam.GuardEmailCode:
			names = append(names, "email_code")
			emailHint = g.Message
		case steam.GuardDeviceConfirmation:
			names = append(names, "device_confirmation")
		case steam.GuardEmailConfirmation:
			names = append(names, "email_confirmation")
		}
	}
	return names, emailHint
}

func (s *ChatService) LocalLogins() LocalLogins { return s.a.localLogins() }

func (s *ChatService) UseSteamAccount(account string) error {
	a := s.a
	err := a.useSteam(strings.TrimSpace(account))
	a.emitLocalLogins()
	if err != nil {
		a.setLogin(LoginState{Phase: "error", Method: SourceSteam, Account: account, Error: err.Error()})
	}
	return err
}

func (s *ChatService) OpenSteamLogin() error {
	a := s.a
	if err := steamlocal.OpenSteam(); err != nil {
		return errors.New("could not open Steam")
	}
	if accs, err := steamlocal.Accounts(); err == nil {
		for _, ac := range accs {
			if ac.MostRecent && ac.HasLogin {
				if _, err := steamlocal.Token(ac.AccountName); err == nil {
					return s.UseSteamAccount(ac.AccountName)
				}
			}
		}
	}
	before := map[string]string{}
	if accs, err := steamlocal.Accounts(); err == nil {
		for _, ac := range accs {
			if ac.HasLogin {
				tok, _ := steamlocal.Token(ac.AccountName)
				before[strings.ToLower(ac.AccountName)] = tok
			}
		}
	}
	sig := steamlocal.Signature()

	a.cancelLogins()
	a.mu.Lock()
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
				if !ac.HasLogin {
					continue
				}
				old, had := before[strings.ToLower(ac.AccountName)]
				if tok, err := steamlocal.Token(ac.AccountName); err != nil || (had && tok == old) {
					continue
				}
				a.mu.Lock()
				a.watchCancel = nil
				a.mu.Unlock()
				if err := a.useSteam(ac.AccountName); err != nil {
					a.setLogin(LoginState{Phase: "error", Method: SourceSteam, Account: ac.AccountName, Error: err.Error()})
				}
				return
			}
		}
	}()
	return nil
}

func (s *ChatService) StartLogin() {
	a := s.a
	ctx, cancel := a.beginLogin()
	a.setLogin(LoginState{Phase: "waiting", Method: SourceQR})
	go func() {
		defer cancel()
		token, account, err := steam.QRLogin(ctx, "myChat", func(url string) {
			a.setLogin(LoginState{Phase: "waiting", Method: SourceQR, URL: url, QR: qrRows(url)})
		})
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			a.setLogin(LoginState{Phase: "error", Method: SourceQR, Error: loginMessage(err)})
			return
		}
		a.finishLogin(SourceQR, token, account, "")
	}()
}

func (s *ChatService) LoginWithPassword(account, password string) {
	a := s.a
	account = strings.TrimSpace(account)
	ctx, cancel := a.beginLogin()
	if account == "" || password == "" {
		cancel()
		a.setLogin(LoginState{Phase: "error", Method: SourcePassword, Account: account, Error: "Enter your account name and password."})
		return
	}
	a.setLogin(LoginState{Phase: "submitting", Method: SourcePassword, Account: account})
	go func() {
		defer cancel()
		l, err := steam.StartCredentials(ctx, "myChat", account, password, a.store.LoadGuardData(account))
		if err != nil {
			if ctx.Err() == nil {
				a.setLogin(LoginState{Phase: "error", Method: SourcePassword, Account: account, Error: loginMessage(err)})
			}
			return
		}
		defer l.Close()
		a.mu.Lock()
		if ctx.Err() != nil {
			a.mu.Unlock()
			return
		}
		a.credLogin = l
		a.mu.Unlock()

		if names, hint := guardNames(l.Guards); len(names) > 0 {
			a.setLogin(LoginState{Phase: "guard", Method: SourcePassword, Account: account, Guards: names, EmailHint: hint})
		}
		token, acct, guard, err := l.Wait(ctx)
		a.mu.Lock()
		if a.credLogin == l {
			a.credLogin = nil
		}
		a.mu.Unlock()
		if err != nil {
			if ctx.Err() == nil {
				a.setLogin(LoginState{Phase: "error", Method: SourcePassword, Account: account, Error: loginMessage(err)})
			}
			return
		}
		a.finishLogin(SourcePassword, token, acct, guard)
	}()
}

func (s *ChatService) SubmitGuardCode(code string) {
	a := s.a
	code = strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(code), " ", ""))
	a.mu.Lock()
	l := a.credLogin
	st := a.login
	a.mu.Unlock()
	if l == nil || st.Phase != "guard" || code == "" {
		return
	}
	st.Checking, st.CodeError = true, ""
	a.setLogin(st)
	go func() {
		kind := steam.GuardEmailCode
		if l.Needs(steam.GuardDeviceCode) {
			kind = steam.GuardDeviceCode
		}
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		err := l.SubmitCode(ctx, kind, code)
		a.mu.Lock()
		cur := a.login
		a.mu.Unlock()
		if cur.Phase != "guard" {
			return
		}
		cur.Checking = false
		if err != nil {
			cur.CodeError = loginMessage(err)
		}
		a.setLogin(cur)
	}()
}

func (s *ChatService) CancelLogin() {
	a := s.a
	a.cancelLogins()
	a.setLogin(LoginState{Phase: "idle"})
	a.emitLocalLogins()
}

func (s *ChatService) Logout() {
	a := s.a
	a.steam.Stop()
	a.store.ClearSession()
	a.cancelLogins()
	a.mu.Lock()
	a.session = config.Session{}
	a.steamToken = ""
	a.mu.Unlock()
	a.setSource("", "")
	a.setStatus(func(st *Status) {
		*st = Status{State: string(steam.StateNeedsLogin)}
	})
	a.setLogin(LoginState{Phase: "idle"})
	a.emitLocalLogins()
}
