package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/tuanta7/gtx/internal/auth"
	"github.com/tuanta7/gtx/internal/config"
)

var tokenFlag bool

var openBrowser = tryOpenBrowser

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authenticate with GitHub",
	Long: `Authenticate with GitHub using GitHub Device Authorization by default.
Use --token only when you need to enter a personal access token manually.`,
	Example: `  # Authenticate with GitHub device authorization
  gtx auth

  # Authenticate using a personal access token
  gtx auth --token`,
	RunE: func(cmd *cobra.Command, args []string) error {
		github := getOrInitGitHubClient()

		var accessToken string
		var err error
		if tokenFlag {
			accessToken, err = authenticateWithToken(cmd.OutOrStdout(), cmd.InOrStdin())
			if err != nil {
				return err
			}
		} else {
			accessToken, err = authenticateWithDeviceFlow(cmd.OutOrStdout(), github)
			if err != nil {
				return err
			}
		}

		user, err := github.FetchUser(accessToken)
		if err != nil {
			return fmt.Errorf("failed to validate GitHub token: %w", err)
		}

		if err := auth.SaveToken(accessToken); err != nil {
			return fmt.Errorf("failed to save GitHub token: %w", err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Authenticated with GitHub as %s\n", user.Login)
		return nil
	},
}

func authenticateWithDeviceFlow(out io.Writer, github *auth.GitHubClient) (string, error) {
	deviceCode, err := github.AuthorizeDevice()
	if err != nil {
		return "", fmt.Errorf("device authorization failed: %w", err)
	}

	fmt.Fprintln(out, "GitHub device authorization")
	fmt.Fprintf(out, "Open: %s\n", deviceCode.VerificationURI)
	fmt.Fprintf(out, "Code: %s\n", deviceCode.UserCode)
	fmt.Fprintln(out, "Waiting for GitHub authorization...")

	if !openBrowser(deviceCode.VerificationURI) {
		fmt.Fprintln(out, "Browser could not be opened automatically. Continue in your browser manually.")
	}

	accessToken, err := github.PollAccessToken(deviceCode.DeviceCode, time.Duration(deviceCode.Interval)*time.Second)
	if err != nil {
		return "", fmt.Errorf("failed to poll access token: %w", err)
	}

	return accessToken, nil
}

func authenticateWithToken(out io.Writer, in io.Reader) (string, error) {
	if !openBrowser(config.GitHubTokensPage) {
		fmt.Fprintln(out, "GitHub token settings could not be opened automatically.")
	}

	fmt.Fprint(out, "Enter your GitHub personal access token: ")
	reader := bufio.NewReader(in)
	accessToken, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("failed to read token: %w", err)
	}

	accessToken = strings.TrimSpace(accessToken)
	if accessToken == "" {
		return "", fmt.Errorf("token is required")
	}

	return accessToken, nil
}

func tryOpenBrowser(url string) bool {
	commands := []string{"xdg-open", "open", "sensible-browser"}
	for _, cmd := range commands {
		if err := exec.Command(cmd, url).Start(); err == nil {
			return true
		}
	}

	return false
}

func init() {
	rootCmd.AddCommand(authCmd)
	authCmd.Flags().BoolVar(&tokenFlag, "token", false, "Enter a GitHub personal access token manually")
}
