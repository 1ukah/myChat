//go:build windows

// Package steamlocal reads the logins that the Steam client on this PC has
// saved ("Remember me"). Steam keeps a refresh token per account in
// %LOCALAPPDATA%\Steam\local.vdf, encrypted with Windows DPAPI for the
// current Windows user. Nothing is written to Steam's files and nothing
// leaves this PC; the token is only used to log on to Steam.
package steamlocal

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"hash/crc32"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

// Account is one account listed by the Steam client.
type Account struct {
	AccountName string `json:"accountName"`
	PersonaName string `json:"personaName"`
	SteamID     string `json:"steamId"`
	MostRecent  bool   `json:"mostRecent"`
	HasLogin    bool   `json:"hasLogin"` // a saved login token exists
}

// SteamDir returns the Steam install folder.
func SteamDir() (string, error) {
	k, err := registry.OpenKey(registry.CURRENT_USER, `Software\Valve\Steam`, registry.QUERY_VALUE)
	if err == nil {
		defer k.Close()
		if p, _, err := k.GetStringValue("SteamPath"); err == nil && p != "" {
			return filepath.Clean(p), nil
		}
	}
	for _, p := range []string{`C:\Program Files (x86)\Steam`, `C:\Program Files\Steam`} {
		if _, err := os.Stat(filepath.Join(p, "steam.exe")); err == nil {
			return p, nil
		}
	}
	return "", errors.New("Steam is not installed")
}

func localVDF() string {
	return filepath.Join(os.Getenv("LOCALAPPDATA"), "Steam", "local.vdf")
}

// Accounts lists the accounts known to the Steam client, most recent first.
func Accounts() ([]Account, error) {
	dir, err := SteamDir()
	if err != nil {
		return nil, err
	}
	b, err := os.ReadFile(filepath.Join(dir, "config", "loginusers.vdf"))
	if err != nil {
		return nil, fmt.Errorf("no Steam accounts on this PC")
	}
	root, err := parseVDF(string(b))
	if err != nil {
		return nil, err
	}
	users := root.child("users")
	if users == nil {
		return nil, nil
	}
	cache := connectCache()
	auto := autoLoginUser()
	var out []Account
	ts := map[string]int64{}
	for _, u := range users.kids {
		name := u.str("AccountName")
		if name == "" {
			continue
		}
		a := Account{
			AccountName: name,
			PersonaName: u.str("PersonaName"),
			SteamID:     u.key,
			MostRecent:  u.str("MostRecent") == "1" || u.str("mostrecent") == "1",
		}
		_, a.HasLogin = cache.blob(name)
		ts[name], _ = strconv.ParseInt(u.str("Timestamp"), 10, 64)
		out = append(out, a)
	}
	// Steam's auto-login account first, then the newest sign-in.
	sort.SliceStable(out, func(i, j int) bool { return ts[out[i].AccountName] > ts[out[j].AccountName] })
	pick := auto
	for _, a := range out {
		if pick == "" && a.MostRecent {
			pick = a.AccountName
		}
	}
	if pick == "" && len(out) > 0 {
		pick = out[0].AccountName
	}
	for i := range out {
		out[i].MostRecent = false
	}
	for i := range out {
		if strings.EqualFold(out[i].AccountName, pick) {
			out[i].MostRecent = true
			out[0], out[i] = out[i], out[0]
			break
		}
	}
	return out, nil
}

func autoLoginUser() string {
	k, err := registry.OpenKey(registry.CURRENT_USER, `Software\Valve\Steam`, registry.QUERY_VALUE)
	if err != nil {
		return ""
	}
	defer k.Close()
	v, _, _ := k.GetStringValue("AutoLoginUser")
	return v
}

// Token returns the saved refresh token for an account, or an error if Steam
// has no saved login for it (for example "Remember me" was off).
func Token(account string) (string, error) {
	blob, ok := connectCache().blob(account)
	if !ok {
		return "", fmt.Errorf("Steam has no saved login for %s. Sign in to Steam with \"Remember me\" on", account)
	}
	var lastErr error
	for _, ent := range []string{strings.ToLower(account), account} {
		plain, err := unprotect(blob, []byte(ent))
		if err == nil {
			tok := strings.TrimRight(string(plain), "\x00")
			if err := checkJWT(tok); err != nil {
				return "", err
			}
			return tok, nil
		}
		lastErr = err
	}
	return "", fmt.Errorf("could not read the Steam login: %w", lastErr)
}

// Signature returns a value that changes when Steam's saved logins change.
func Signature() string {
	st, err := os.Stat(localVDF())
	if err != nil {
		return ""
	}
	return strconv.FormatInt(st.ModTime().UnixNano(), 10) + ":" + strconv.FormatInt(st.Size(), 10)
}

type cacheMap map[string]string // lower-case key -> hex blob

func connectCache() cacheMap {
	b, err := os.ReadFile(localVDF())
	if err != nil {
		return nil
	}
	root, err := parseVDF(string(b))
	if err != nil {
		return nil
	}
	n := root
	for _, k := range []string{"MachineUserConfigStore", "Software", "Valve", "Steam", "ConnectCache"} {
		if n = n.child(k); n == nil {
			return nil
		}
	}
	m := cacheMap{}
	for _, kv := range n.vals {
		m[strings.ToLower(kv[0])] = kv[1]
	}
	return m
}

// blob finds an account's entry. Keys are the CRC32 of the account name in
// hex, followed by a digit.
func (m cacheMap) blob(account string) ([]byte, bool) {
	for _, name := range []string{strings.ToLower(account), account} {
		prefix := fmt.Sprintf("%08x", crc32.ChecksumIEEE([]byte(name)))
		for k, v := range m {
			if strings.HasPrefix(k, prefix) && v != "" {
				if b, err := hex.DecodeString(v); err == nil && len(b) > 0 {
					return b, true
				}
			}
		}
	}
	return nil, false
}

func checkJWT(tok string) error {
	parts := strings.Split(tok, ".")
	if len(parts) != 3 {
		return errors.New("Steam's saved login has an unknown format")
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return errors.New("Steam's saved login has an unknown format")
	}
	var c struct {
		Aud json.RawMessage `json:"aud"`
	}
	_ = json.Unmarshal(raw, &c)
	if !strings.Contains(string(c.Aud), "client") {
		return errors.New("Steam's saved login cannot be used for chat")
	}
	return nil
}

func unprotect(enc, entropy []byte) ([]byte, error) {
	in := windows.DataBlob{Size: uint32(len(enc)), Data: &enc[0]}
	ent := windows.DataBlob{Size: uint32(len(entropy)), Data: &entropy[0]}
	var out windows.DataBlob
	if err := windows.CryptUnprotectData(&in, nil, &ent, 0, nil, windows.CRYPTPROTECT_UI_FORBIDDEN, &out); err != nil {
		return nil, err
	}
	defer windows.LocalFree(windows.Handle(unsafe.Pointer(out.Data)))
	return append([]byte(nil), unsafe.Slice(out.Data, out.Size)...), nil
}

// OpenSteam shows the Steam client. If no account is signed in, Steam shows
// its own login window.
func OpenSteam() error {
	verb, _ := windows.UTF16PtrFromString("open")
	url, _ := windows.UTF16PtrFromString("steam://open/main")
	return windows.ShellExecute(0, verb, url, nil, nil, windows.SW_SHOWNORMAL)
}
