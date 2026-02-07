package token

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	PATProvider    = "pat"
	GitHubProvider = "github.com"
	GitLabProvider = "gitlab.com"
)

var AllProviders = []string{
	GitHubProvider,
}

type DeviceCodeResponse struct {
	DeviceCode      string
	UserCode        string
	VerificationURI string
	ExpiresIn       int
	Interval        int
}

type User struct {
	Login string `json:"login"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type AuthStrategy interface {
	Provider() string
	AuthorizeDevice() (*DeviceCodeResponse, error)
	PollAccessToken(deviceCode string, interval time.Duration) (string, error)
	FetchUser(token string) (*User, error)
	SaveToken(token string) error
}

type Manager struct {
	strategies map[string]AuthStrategy
}

func NewManager() *Manager {
	return &Manager{
		strategies: make(map[string]AuthStrategy),
	}
}

func (m *Manager) Register(strategy AuthStrategy) {
	if m.strategies == nil {
		m.strategies = make(map[string]AuthStrategy)
	}

	m.strategies[strategy.Provider()] = strategy
}

func (m *Manager) GetStrategy(provider string) (AuthStrategy, error) {
	if m.strategies == nil {
		return nil, fmt.Errorf("no strategies registered")
	}

	if strategy, ok := m.strategies[provider]; ok {
		return strategy, nil
	}

	return nil, fmt.Errorf("no strategy found for provider %s", provider)
}

func (m *Manager) LoadToken() (string, string, error) {
	configDir, err := os.UserHomeDir()
	if err != nil {
		return "", "", err
	}

	tokenFile := filepath.Join(configDir, ".netrc")
	data, err := os.ReadFile(tokenFile)
	if err != nil {
		return "", "", err
	}

	parts := strings.SplitN(string(data), ":", 2)
	if len(parts) != 2 || !includes(parts[0], AllProviders) {
		return "", "", fmt.Errorf("invalid token file format")
	}

	return parts[0], parts[1], nil
}
