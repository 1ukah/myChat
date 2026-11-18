//go:build windows

package steamlocal

import (
	"os"
	"strings"
	"testing"
)

func TestParseVDF(t *testing.T) {
	root, err := parseVDF(`"users"
{
	// comment
	"7656" { "AccountName" "bob" "MostRecent" "1" "Persona" "B \"x\"" }
}`)
	if err != nil {
		t.Fatal(err)
	}
	u := root.child("users").child("7656")
	if u.str("accountname") != "bob" || u.str("Persona") != `B "x"` {
		t.Fatalf("got %+v", u)
	}
}

func TestLocalAccounts(t *testing.T) {
	if os.Getenv("MYCHAT_LOCAL") == "" {
		t.Skip("set MYCHAT_LOCAL=1")
	}
	accs, err := Accounts()
	if err != nil {
		t.Fatal(err)
	}
	for _, a := range accs {
		res := "no saved login"
		if a.HasLogin {
			if tok, err := Token(a.AccountName); err != nil {
				res = "saved login unreadable: " + err.Error()
			} else {
				res = "usable login (JWT, " + itoa(strings.Count(tok, ".")+1) + " parts)"
			}
		}
		t.Logf("%s (recent=%v): %s", a.AccountName, a.MostRecent, res)
	}
}

func itoa(n int) string { return string(rune('0' + n)) }
