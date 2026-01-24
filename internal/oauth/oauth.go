package oauth

import (
	"os"
	"path/filepath"
	"time"
)

const (
	GitHubProvider = "github"
)

type DeviceCodeResponse struct {
	DeviceCode      string
	UserCode        string
	VerificationURI string
	ExpiresIn       int
	Interval        int
}

type Strategy interface {
	Provider() string
	AuthorizeDevice() (*DeviceCodeResponse, error)
	PollAccessToken(deviceCode string, interval time.Duration) (string, error)
}

type Manager struct {
	strategies map[string]Strategy
}

func NewManager() *Manager {
	return &Manager{
		strategies: make(map[string]Strategy),
	}
}

func (m *Manager) Register(strategy Strategy) {
	if m.strategies == nil {
		m.strategies = make(map[string]Strategy)
	}

	m.strategies[strategy.Provider()] = strategy
}

func (m *Manager) SaveToken(token string) error {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return err
	}

	tigDir := filepath.Join(configDir, "tig")
	if err := os.MkdirAll(tigDir, 0700); err != nil {
		return err
	}

	tokenFile := filepath.Join(tigDir, "token")
	return os.WriteFile(tokenFile, []byte(token), 0600)
}

func (m *Manager) GetStrategy(provider string) Strategy {
	return m.strategies[provider]
}
