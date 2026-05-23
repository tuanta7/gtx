package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tuanta7/gtx/internal/token"
)

// profileCmd represents the profile command
var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Show authentication and profile info",
	Long: `Display the current authentication status and token information.

Examples:
  # Show current token info
  gtx profile`,
	RunE: func(cmd *cobra.Command, args []string) error {
		provider, tokenValue, err := token.LoadToken()
		if err != nil {
			if errors.Is(err, token.ErrAuthRequired) {
				fmt.Fprintln(cmd.OutOrStdout(), "Authentication required. Run 'gtx auth'.")
				return err
			}
			return fmt.Errorf("failed to load authentication token: %w", err)
		}

		maskedToken := maskToken(tokenValue)
		fmt.Fprintln(cmd.OutOrStdout(), "Token:", maskedToken)

		strategy, err := manager.GetStrategy(provider)
		if err != nil {
			return fmt.Errorf("failed to get strategy: %w", err)
		}

		// Fetch user info from GitHub
		user, err := strategy.FetchUser(tokenValue)
		if err != nil {
			fmt.Fprintln(cmd.OutOrStdout(), "Status: Token invalid or expired")
			fmt.Fprintln(cmd.OutOrStdout(), "Run 'gtx auth' to re-authenticate")
			return nil
		}

		fmt.Fprintln(cmd.OutOrStdout(), "Status: Authenticated")
		fmt.Fprintf(cmd.OutOrStdout(), "Provider: %s\n", provider)
		fmt.Fprintln(cmd.OutOrStdout(), "Username:", user.Login)
		if user.Name != "" {
			fmt.Fprintln(cmd.OutOrStdout(), "Name:", user.Name)
		}
		if user.Email != "" {
			fmt.Fprintln(cmd.OutOrStdout(), "Email:", user.Email)
		}

		return nil
	},
}

func maskToken(token string) string {
	if len(token) <= 8 {
		return "****"
	}
	return token[:4] + strings.Repeat("*", len(token)-8) + token[len(token)-4:]
}

func init() {
	rootCmd.AddCommand(profileCmd)
}
