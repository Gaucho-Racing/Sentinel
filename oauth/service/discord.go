package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/gaucho-racing/sentinel/oauth/config"
	"github.com/gaucho-racing/sentinel/oauth/pkg/logger"
)

// Subset of Discord's OAuth2 token-exchange response. We only consume
// access_token to call /users/@me; the rest is parsed for forward-compat /
// debug logging.
type DiscordAccessTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
}

// Subset of Discord's user object — only the fields we either persist or
// surface back to the client. `id` is the snowflake we key external auth on;
// `email` is present because we request the `email` scope at authorize time.
type DiscordUser struct {
	ID            string `json:"id"`
	Username      string `json:"username"`
	GlobalName    string `json:"global_name"`
	Avatar        string `json:"avatar"`
	Email         string `json:"email"`
	Verified      bool   `json:"verified"`
}

// ExchangeDiscordCode trades an authorization code for a Discord access
// token. The redirect_uri must byte-match the one the web client used at
// authorize time — Discord rejects the exchange otherwise.
func ExchangeDiscordCode(code string) (*DiscordAccessTokenResponse, error) {
	form := url.Values{}
	form.Set("client_id", config.DiscordClientID)
	form.Set("client_secret", config.DiscordClientSecret)
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("redirect_uri", config.DiscordRedirectURI)

	resp, err := http.PostForm("https://discord.com/api/oauth2/token", form)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		logger.SugarLogger.Errorf("discord: token exchange returned %d: %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("discord token exchange failed")
	}

	var out DiscordAccessTokenResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetDiscordUser fetches the authenticated Discord user via /users/@me.
func GetDiscordUser(accessToken string) (*DiscordUser, error) {
	req, err := http.NewRequest(http.MethodGet, "https://discord.com/api/users/@me", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		logger.SugarLogger.Errorf("discord: /users/@me returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
		return nil, fmt.Errorf("discord user lookup failed")
	}

	var u DiscordUser
	if err := json.Unmarshal(body, &u); err != nil {
		return nil, err
	}
	return &u, nil
}
