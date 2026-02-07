package token

import (
	"fmt"
	"os"
	"path/filepath"
)

func saveToken(provider, token string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	tokenFile := filepath.Join(homeDir, ".netrc")
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
