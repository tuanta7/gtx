package token

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type GitHubStrategy struct {
	clientID       string
	deviceCodeURL  string
	accessTokenURL string
	userProfileURL string
}

func NewGitHubStrategy(clientID, deviceCodeURL, accessTokenURL, userProfileURL string) *GitHubStrategy {
	return &GitHubStrategy{
		clientID:       clientID,
		deviceCodeURL:  deviceCodeURL,
		accessTokenURL: accessTokenURL,
		userProfileURL: userProfileURL,
	}
}

func (g *GitHubStrategy) Provider() string {
	return GitHubProvider
}

func (g *GitHubStrategy) AuthorizeDevice() (*DeviceCodeResponse, error) {
	bodyForm := &url.Values{}
	bodyForm.Set("client_id", g.clientID)
	bodyForm.Set("scope", "repo,read:org")
	body := bytes.NewBufferString(bodyForm.Encode())

	req, err := http.NewRequest(http.MethodPost, g.deviceCodeURL, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub device code request returned status %d", resp.StatusCode)
	}

	var deviceCode struct {
		DeviceCode      string `json:"device_code"`
		UserCode        string `json:"user_code"`
		VerificationURI string `json:"verification_uri"`
		ExpiresIn       int    `json:"expires_in"`
		Interval        int    `json:"interval"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&deviceCode); err != nil {
		return nil, err
	}

	return &DeviceCodeResponse{
		DeviceCode:      deviceCode.DeviceCode,
		UserCode:        deviceCode.UserCode,
		VerificationURI: deviceCode.VerificationURI,
		ExpiresIn:       deviceCode.ExpiresIn,
		Interval:        deviceCode.Interval,
	}, nil
}

func (g *GitHubStrategy) PollAccessToken(deviceCode string, interval time.Duration) (string, error) {
	for {
		time.Sleep(interval)

		token, err := g.pollAccessToken(deviceCode)
		if err != nil {
			return "", err
		}

		if token != "" {
			return token, nil
		}
	}
}

func (g *GitHubStrategy) pollAccessToken(deviceCode string) (string, error) {
	bodyForm := url.Values{}
	bodyForm.Set("client_id", g.clientID)
	bodyForm.Set("device_code", deviceCode)
	bodyForm.Set("grant_type", "urn:ietf:params:oauth:grant-type:device_code")
	body := bytes.NewBufferString(bodyForm.Encode())

	req, err := http.NewRequest(http.MethodPost, g.accessTokenURL, body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub token request returned status %d", resp.StatusCode)
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		Scope       string `json:"scope"`
		Error       string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", err
	}

	switch tokenResp.Error {
	case "":
		return tokenResp.AccessToken, nil
	case "authorization_pending":
		return "", nil
	case "slow_down":
		time.Sleep(5 * time.Second)
		return "", nil
	case "expired_token":
		return "", fmt.Errorf("device code expired, please try again")
	case "access_denied":
		return "", fmt.Errorf("access denied by user")
	default:
		return "", fmt.Errorf("authentication error: %s", tokenResp.Error)
	}
}

func (g *GitHubStrategy) SaveToken(token string) error {
	return saveToken(token)
}

func (g *GitHubStrategy) FetchUser(token string) (*User, error) {
	req, err := http.NewRequest(http.MethodGet, g.userProfileURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var user struct {
		Login string `json:"login"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}

	return &User{
		Login: user.Login,
		Name:  user.Name,
		Email: user.Email,
	}, nil
}
