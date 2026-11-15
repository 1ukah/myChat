package steamlocal

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

const steamID64Base = 76561197960265728

// PersonaState returns the status the user picked in the Steam app for this
// account (EPersonaState: 0 offline, 1 online, 2 busy, 3 away, 4 snooze,
// 7 invisible). ok is false when Steam has not saved one.
func PersonaState(steamID uint64) (state uint32, ok bool) {
	if steamID <= steamID64Base {
		return 0, false
	}
	dir, err := SteamDir()
	if err != nil {
		return 0, false
	}
	acct := strconv.FormatUint(steamID-steamID64Base, 10)
	b, err := os.ReadFile(filepath.Join(dir, "userdata", acct, "config", "localconfig.vdf"))
	if err != nil {
		return 0, false
	}
	return parsePersonaState(string(b), acct)
}

func parsePersonaState(vdf, acct string) (uint32, bool) {
	re := regexp.MustCompile(`"FriendStoreLocalPrefs_` + regexp.QuoteMeta(acct) + `"\s+"((?:[^"\\]|\\.)*)"`)
	m := re.FindStringSubmatch(vdf)
	if m == nil {
		return 0, false
	}
	var prefs struct {
		EPersonaState *uint32 `json:"ePersonaState"`
	}
	if json.Unmarshal([]byte(strings.ReplaceAll(m[1], `\"`, `"`)), &prefs) != nil || prefs.EPersonaState == nil {
		return 0, false
	}
	return *prefs.EPersonaState, true
}
