package steam

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// Names, avatars and online state come from the web API
// (IPlayerService/GetPlayerLinkDetails, the call Steam's own web pages use),
// not from persona requests on the CM connection. A web request is answered
// only to myChat, so the user's Steam app never sees it.

const profileBatch = 100

var httpClient = &http.Client{Timeout: 15 * time.Second}

// RefreshProfiles asks the profile loop to fetch now. Results arrive as
// FriendsEvent.
func (c *Client) RefreshProfiles() {
	select {
	case c.refresh <- struct{}{}:
	default:
	}
}

func (c *Client) profileLoop(ctx context.Context, cn *conn) {
	tick := time.NewTicker(profileRefresh)
	defer tick.Stop()
	for {
		if err := c.fetchProfiles(ctx, cn); err != nil && ctx.Err() == nil {
			log.Printf("steam: profiles: %v", err)
		}
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
		case <-c.refresh:
		}
	}
}

// accessToken returns a web API access token, asking Steam for a new one
// when the old one is about to expire. The refresh token is never renewed.
func (c *Client) accessToken(ctx context.Context, cn *conn) (string, error) {
	c.mu.RLock()
	tok, exp, refresh, self := c.access, c.accessExpires, c.token, c.self
	c.mu.RUnlock()
	if tok != "" && time.Until(exp) > 5*time.Minute {
		return tok, nil
	}
	var req enc
	req.string(1, refresh)
	req.fixed64(2, self)
	req.varint(3, 0) // ETokenRenewalType None
	p, err := cn.call(ctx, emsgServiceMethodCallFromClient, "Authentication.GenerateAccessTokenForApp#1", req.b)
	if err != nil {
		return "", err
	}
	tok = ""
	_ = walk(p.body, func(f field) {
		if f.Num == 1 {
			tok = f.str()
		}
	})
	if tok == "" {
		return "", errors.New("no access token")
	}
	exp = time.Now().Add(time.Hour)
	if info, err := ParseToken(tok); err == nil && !info.Expires.IsZero() {
		exp = info.Expires
	}
	c.mu.Lock()
	c.access, c.accessExpires = tok, exp
	c.mu.Unlock()
	return tok, nil
}

type linkDetails struct {
	Response struct {
		Accounts []struct {
			Public struct {
				SteamID string `json:"steamid"`
				Name    string `json:"persona_name"`
				Avatar  string `json:"sha_digest_avatar"` // base64 SHA-1
			} `json:"public_data"`
			Private struct {
				State  uint32 `json:"persona_state"`
				GameID string `json:"game_id"`
			} `json:"private_data"`
		} `json:"accounts"`
	} `json:"response"`
}

func (c *Client) fetchProfiles(ctx context.Context, cn *conn) error {
	c.mu.RLock()
	ids := []uint64{c.self}
	for id, r := range c.relation {
		if r == 3 { // EFriendRelationship Friend
			ids = append(ids, id)
		}
	}
	c.mu.RUnlock()

	tok, err := c.accessToken(ctx, cn)
	if err != nil {
		return fmt.Errorf("access token: %w", err)
	}
	var all linkDetails
	for start := 0; start < len(ids); start += profileBatch {
		batch := ids[start:min(start+profileBatch, len(ids))]
		q := url.Values{"access_token": {tok}}
		for i, id := range batch {
			q.Set("steamids["+strconv.Itoa(i)+"]", strconv.FormatUint(id, 10))
		}
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet,
			"https://api.steampowered.com/IPlayerService/GetPlayerLinkDetails/v1/?"+q.Encode(), nil)
		resp, err := httpClient.Do(req)
		if err != nil {
			return err
		}
		var part linkDetails
		err = json.NewDecoder(resp.Body).Decode(&part)
		resp.Body.Close()
		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
			c.mu.Lock()
			c.access = "" // get a new token next time
			c.mu.Unlock()
		}
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("GetPlayerLinkDetails: HTTP %d", resp.StatusCode)
		}
		if err != nil {
			return err
		}
		all.Response.Accounts = append(all.Response.Accounts, part.Response.Accounts...)
	}

	c.mu.Lock()
	for _, a := range all.Response.Accounts {
		id, err := strconv.ParseUint(a.Public.SteamID, 10, 64)
		if err != nil || id == 0 {
			continue
		}
		fr := c.friends[id]
		if fr == nil {
			fr = &Friend{SteamID: id}
			c.friends[id] = fr
		}
		fr.Name = a.Public.Name
		if raw, err := base64.StdEncoding.DecodeString(a.Public.Avatar); err == nil && len(raw) > 0 {
			fr.AvatarHash = hex.EncodeToString(raw)
		}
		fr.PersonaState = a.Private.State
		fr.GameID, _ = strconv.ParseUint(a.Private.GameID, 10, 64)
	}
	c.mu.Unlock()
	c.emit(FriendsEvent{})
	return nil
}
