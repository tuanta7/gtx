package token

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadTokenReturnsAuthRequired(t *testing.T) {
	home := t.TempDir()
	originalHome := userHomeDir
	userHomeDir = func() (string, error) { return home, nil }
	t.Cleanup(func() { userHomeDir = originalHome })

	_, _, err := LoadToken()
	require.ErrorIs(t, err, ErrAuthRequired)
}
