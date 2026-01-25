package cmd

import (
	"fmt"
	"os/exec"
	"time"

	"github.com/spf13/cobra"
	"github.com/tuanta7/tig/internal/config"
	"github.com/tuanta7/tig/internal/token"
)

var (
	tokenFlag    bool
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
  tig auth --token`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if providerFlag != token.GitHubProvider {
			return fmt.Errorf("unsupported provider: %s (only 'github' is currently supported)", providerFlag)
		}

		var err error
		var accessToken string
		var strategy token.AuthStrategy

		if tokenFlag {
			// Only support GitHub for now
			strategy, err = manager.GetStrategy(token.PATProvider)
			if err != nil {
				return fmt.Errorf("failed to get strategy: %w", err)
			}

			tryOpenBrowser(config.GitHubTokensPage)
			fmt.Print("Enter your personal access token: ")
			_, err := fmt.Scanln(&accessToken)
			if err != nil {
				return fmt.Errorf("failed to read token: %w", err)
			}
		} else {
			strategy, err = manager.GetStrategy(providerFlag)
			if err != nil {
				return fmt.Errorf("failed to get strategy: %w", err)
			}

			deviceCode, err := strategy.AuthorizeDevice()
			if err != nil {
				return fmt.Errorf("device authorization failed: %w", err)
			}

			fmt.Println("Copy your one-time code:", deviceCode.UserCode)
			fmt.Printf("Click to open in your browser\n%s", deviceCode.VerificationURI)
			_, _ = fmt.Scanln()

			tryOpenBrowser(deviceCode.VerificationURI)

			accessToken, err = strategy.PollAccessToken(deviceCode.DeviceCode, time.Duration(deviceCode.Interval)*time.Second)
			if err != nil {
				return fmt.Errorf("failed to poll access token: %w", err)
			}
		}

		if err := strategy.SaveToken(accessToken); err != nil {
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
	authCmd.Flags().BoolVar(&tokenFlag, "token", false, "Your personal access token")
	authCmd.Flags().StringVar(&providerFlag, "provider", token.GitHubProvider, "OAuth provider (currently only 'github' is supported)")
}
