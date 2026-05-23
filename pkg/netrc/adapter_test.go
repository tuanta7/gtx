package netrc

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAdapterSetUpsertsMachineWithoutDroppingOthers(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), ".netrc")
	require.NoError(t, os.WriteFile(path, []byte("machine example.com\nlogin user\npassword old\n"), 0o600))

	adapter, err := NewAdapter(path)
	require.NoError(t, err)

	require.NoError(t, adapter.Set(&Machine{
		Name:     "github.com",
		Login:    "oauth2",
		Password: "token",
	}))

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Contains(t, string(data), "machine example.com\nlogin user\npassword old\n")
	require.Contains(t, string(data), "machine github.com\nlogin oauth2\npassword token\n")
}

func TestParseReadsMachines(t *testing.T) {
	t.Parallel()

	machines, err := Parse(strings.NewReader("machine github.com\nlogin oauth2\npassword token\n"))
	require.NoError(t, err)
	require.Equal(t, "oauth2", machines["github.com"].Login)
	require.Equal(t, "token", machines["github.com"].Password)
}
