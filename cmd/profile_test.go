package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tuanta7/gtx/internal/token"
)

func TestProfileRequiresAuthentication(t *testing.T) {
	originalManager := manager
	t.Cleanup(func() {
		manager = originalManager
	})

	manager = token.NewManager()
	t.Setenv("HOME", t.TempDir())

	var output bytes.Buffer
	profileCmd.SetOut(&output)
	profileCmd.SetErr(&output)

	err := profileCmd.RunE(profileCmd, nil)
	require.ErrorIs(t, err, token.ErrAuthRequired)
	require.Contains(t, output.String(), "Authentication required. Run 'gtx auth'.")
}

func TestProfileUsesStoredToken(t *testing.T) {
	originalManager := manager
	t.Cleanup(func() {
		manager = originalManager
	})

	strategy := &fakeAuthStrategy{
		user: &token.User{Login: "octocat"},
	}

	manager = token.NewManager()
	manager.Register(strategy)
	home := t.TempDir()
	t.Setenv("HOME", home)
	require.NoError(t, os.WriteFile(filepath.Join(home, ".netrc"), []byte("machine github.com\nlogin oauth2\npassword stored-token\n"), 0o600))

	var output bytes.Buffer
	profileCmd.SetOut(&output)
	profileCmd.SetErr(&output)

	err := profileCmd.RunE(profileCmd, nil)
	require.NoError(t, err)
	require.Contains(t, output.String(), "Status: Authenticated")
	require.Contains(t, output.String(), "Username: octocat")
}

func TestProfileReportsInvalidToken(t *testing.T) {
	originalManager := manager
	t.Cleanup(func() {
		manager = originalManager
	})

	strategy := &fakeAuthStrategy{
		fetchErr: fmt.Errorf("unauthorized"),
	}

	manager = token.NewManager()
	manager.Register(strategy)
	home := t.TempDir()
	t.Setenv("HOME", home)
	require.NoError(t, os.WriteFile(filepath.Join(home, ".netrc"), []byte("machine github.com\nlogin oauth2\npassword stored-token\n"), 0o600))

	var output bytes.Buffer
	profileCmd.SetOut(&output)
	profileCmd.SetErr(&output)

	err := profileCmd.RunE(profileCmd, nil)
	require.NoError(t, err)
	require.Contains(t, output.String(), "Status: Token invalid or expired")
}
