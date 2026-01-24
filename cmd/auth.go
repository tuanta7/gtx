package cmd

import (
	"fmt"
	"os/exec"
	"time"

	"github.com/spf13/cobra"
	"github.com/tuanta7/tig/internal/config"
	"github.com/tuanta7/tig/internal/oauth"
)

var (
	tokenFlag    string
	providerFlag string
)

// authCmd represents the auth command
var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authenticate with GitHub",
	Long: `Authenticate with GitHub using a personal access token or device authorization flow.

Examples:
  # Authenticate using device authorization flow (interactive)
  tig auth
	
  # Specify provider (currently only github is supported)
  tig auth --provider github

  # Authenticate using a personal access token
  tig auth --token ghp_xxxxxxxxxxxx`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if providerFlag != oauth.GitHubProvider {
			return fmt.Errorf("unsupported provider: %s (only 'github' is currently supported)", providerFlag)
		}

		var token string

		om := oauth.NewManager()
		om.Register(oauth.NewGitHubStrategy(
			config.GitHubOAuthClientID,
			config.GitHubDeviceCodeURL,
			config.GitHubAccessTokenURL,
		))

		if tokenFlag != "" {
			token = tokenFlag
		} else {
			s := om.GetStrategy(providerFlag)
			deviceCode, err := s.AuthorizeDevice()
			if err != nil {
				return fmt.Errorf("device authorization failed: %w", err)
			}

			fmt.Println("Copy your one-time code:", deviceCode.UserCode)
			fmt.Printf("Click to open in your browser\n%s", deviceCode.VerificationURI)
			_, _ = fmt.Scanln()

			tryOpenBrowser(deviceCode.VerificationURI)

			token, err = s.PollAccessToken(deviceCode.DeviceCode, time.Duration(deviceCode.Interval)*time.Second)
			if err != nil {
				return fmt.Errorf("failed to poll access token: %w", err)
			}
		}

		if err := om.SaveToken(token); err != nil {
			return fmt.Errorf("failed to save token: %w", err)
		}

		return nil
	},
}

func tryOpenBrowser(url string) {
	commands := []string{"xdg-open", "open", "sensible-browser"}
	for _, cmd := range commands {
		if err := exec.Command(cmd, url).Start(); err == nil {
			return
		}
	}
}

func init() {
	rootCmd.AddCommand(authCmd)
	authCmd.Flags().StringVar(&tokenFlag, "token", "", "Your personal access token")
	authCmd.Flags().StringVar(&providerFlag, "provider", oauth.GitHubProvider, "OAuth provider (currently only 'github' is supported)")
}
