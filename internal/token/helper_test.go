package token

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSaveAndLoadTokenWithNetrc(t *testing.T) {
	home := t.TempDir()
	originalHome := userHomeDir
	userHomeDir = func() (string, error) { return home, nil }
	t.Cleanup(func() { userHomeDir = originalHome })

	require.NoError(t, saveToken("secret-token"))

	provider, tokenValue, err := LoadToken()
	require.NoError(t, err)
	require.Equal(t, GitHubProvider, provider)
	require.Equal(t, "secret-token", tokenValue)

	data, err := os.ReadFile(filepath.Join(home, ".netrc"))
	require.NoError(t, err)
	require.Contains(t, string(data), "machine github.com")
	require.Contains(t, string(data), "login oauth2")
	require.Contains(t, string(data), "password secret-token")
}

func TestLoadTokenMigratesLegacyFormat(t *testing.T) {
	home := t.TempDir()
	originalHome := userHomeDir
	userHomeDir = func() (string, error) { return home, nil }
	t.Cleanup(func() { userHomeDir = originalHome })

	require.NoError(t, os.WriteFile(filepath.Join(home, ".netrc"), []byte("github.com:legacy-token"), 0o600))

	provider, tokenValue, err := LoadToken()
	require.NoError(t, err)
	require.Equal(t, GitHubProvider, provider)
	require.Equal(t, "legacy-token", tokenValue)

	data, err := os.ReadFile(filepath.Join(home, ".netrc"))
	require.NoError(t, err)
	require.Contains(t, string(data), "machine github.com")
	require.Contains(t, string(data), "password legacy-token")
}

func TestLoadTokenRequiresAuthWhenMissing(t *testing.T) {
	home := t.TempDir()
	originalHome := userHomeDir
	userHomeDir = func() (string, error) { return home, nil }
	t.Cleanup(func() { userHomeDir = originalHome })

	_, _, err := LoadToken()
	require.ErrorIs(t, err, ErrAuthRequired)
}
