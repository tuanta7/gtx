package token

import (
	"fmt"
	"os"
	"path/filepath"
)

func saveToken(provider, token string) error {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return err
	}

	tigDir := filepath.Join(configDir, "tig")
	if err := os.MkdirAll(tigDir, 0700); err != nil {
		return err
	}

	tokenFile := filepath.Join(tigDir, "token")
	contents := fmt.Sprintf("%s:%s", provider, token)
	return os.WriteFile(tokenFile, []byte(contents), 0600)
}

func includes(s string, list []string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}

	return false
}
