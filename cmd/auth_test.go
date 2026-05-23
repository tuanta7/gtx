package cmd

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tuanta7/gtx/internal/token"
)

type fakeAuthStrategy struct {
	deviceCode   *token.DeviceCodeResponse
	pollToken    string
	pollErr      error
	user         *token.User
	fetchErr     error
	savedToken   string
	authorizeHit bool
}

func (f *fakeAuthStrategy) Provider() string { return token.GitHubProvider }

func (f *fakeAuthStrategy) AuthorizeDevice() (*token.DeviceCodeResponse, error) {
	f.authorizeHit = true
	return f.deviceCode, nil
}

func (f *fakeAuthStrategy) PollAccessToken(deviceCode string, interval time.Duration) (string, error) {
	return f.pollToken, f.pollErr
}

func (f *fakeAuthStrategy) FetchUser(accessToken string) (*token.User, error) {
	if f.fetchErr != nil {
		return nil, f.fetchErr
	}
	return f.user, nil
}

func (f *fakeAuthStrategy) SaveToken(accessToken string) error {
	f.savedToken = accessToken
	return nil
}

func TestAuthCommandUsesDeviceFlowByDefault(t *testing.T) {
	originalManager := manager
	originalOpenBrowser := openBrowser
	originalTokenFlag := tokenFlag
	t.Cleanup(func() {
		manager = originalManager
		openBrowser = originalOpenBrowser
		tokenFlag = originalTokenFlag
	})

	strategy := &fakeAuthStrategy{
		deviceCode: &token.DeviceCodeResponse{
			DeviceCode:      "device-code",
			UserCode:        "ABCD-EFGH",
			VerificationURI: "https://github.com/login/device",
			Interval:        1,
		},
		pollToken: "device-token",
		user:      &token.User{Login: "octocat"},
	}

	manager = token.NewManager()
	manager.Register(strategy)
	openBrowser = func(string) bool { return true }
	tokenFlag = false

	var output bytes.Buffer
	authCmd.SetOut(&output)
	authCmd.SetErr(&output)
	authCmd.SetIn(bytes.NewBuffer(nil))

	err := authCmd.RunE(authCmd, nil)
	require.NoError(t, err)
	require.True(t, strategy.authorizeHit)
	require.Equal(t, "device-token", strategy.savedToken)
	require.Contains(t, output.String(), "Authenticated with GitHub as octocat")
}

func TestAuthCommandTokenFallbackValidatesBeforeSaving(t *testing.T) {
	originalManager := manager
	originalOpenBrowser := openBrowser
	originalTokenFlag := tokenFlag
	t.Cleanup(func() {
		manager = originalManager
		openBrowser = originalOpenBrowser
		tokenFlag = originalTokenFlag
	})

	strategy := &fakeAuthStrategy{
		user: &token.User{Login: "octocat"},
	}

	manager = token.NewManager()
	manager.Register(strategy)
	openBrowser = func(string) bool { return false }
	tokenFlag = true

	var output bytes.Buffer
	authCmd.SetOut(&output)
	authCmd.SetErr(&output)
	authCmd.SetIn(bytes.NewBufferString("manual-token\n"))

	err := authCmd.RunE(authCmd, nil)
	require.NoError(t, err)
	require.Equal(t, "manual-token", strategy.savedToken)
	require.Contains(t, output.String(), "GitHub token settings could not be opened automatically.")
}

func TestAuthCommandRejectsInvalidManualToken(t *testing.T) {
	originalManager := manager
	originalOpenBrowser := openBrowser
	originalTokenFlag := tokenFlag
	t.Cleanup(func() {
		manager = originalManager
		openBrowser = originalOpenBrowser
		tokenFlag = originalTokenFlag
	})

	strategy := &fakeAuthStrategy{
		fetchErr: fmt.Errorf("unauthorized"),
	}

	manager = token.NewManager()
	manager.Register(strategy)
	openBrowser = func(string) bool { return true }
	tokenFlag = true

	var output bytes.Buffer
	authCmd.SetOut(&output)
	authCmd.SetErr(&output)
	authCmd.SetIn(bytes.NewBufferString("bad-token\n"))

	err := authCmd.RunE(authCmd, nil)
	require.ErrorContains(t, err, "failed to validate GitHub token")
	require.Empty(t, strategy.savedToken)
}
