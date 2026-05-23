package token

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"

	pkgnetrc "github.com/tuanta7/gtx/pkg/netrc"
)

const netrcLogin = "oauth2"

var userHomeDir = os.UserHomeDir

func saveToken(token string) error {
	homeDir, err := userHomeDir()
	if err != nil {
		return err
	}

	tokenFile := filepath.Join(homeDir, ".netrc")
	adapter, err := pkgnetrc.NewAdapter(tokenFile)
	if err != nil {
		return err
	}

	return adapter.Set(&pkgnetrc.Machine{
		Name:     GitHubProvider,
		Login:    netrcLogin,
		Password: token,
	})
}

func LoadToken() (string, string, error) {
	homeDir, err := userHomeDir()
	if err != nil {
		return "", "", err
	}

	tokenFile := filepath.Join(homeDir, ".netrc")
	data, err := os.ReadFile(tokenFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", "", ErrAuthRequired
		}
		return "", "", err
	}

	if provider, token, ok := parseLegacyToken(data); ok {
		if err := saveToken(token); err != nil {
			return "", "", err
		}
		return provider, token, nil
	}

	machines, err := pkgnetrc.Parse(bytes.NewReader(data))
	if err != nil {
		return "", "", err
	}

	machine, ok := machines[GitHubProvider]
	if !ok || machine.Password == "" {
		return "", "", ErrAuthRequired
	}

	return GitHubProvider, machine.Password, nil
}

func parseLegacyToken(data []byte) (string, string, bool) {
	parts := strings.SplitN(strings.TrimSpace(string(data)), ":", 2)
	if len(parts) != 2 || parts[0] != GitHubProvider || parts[1] == "" {
		return "", "", false
	}

	return parts[0], parts[1], true
}
