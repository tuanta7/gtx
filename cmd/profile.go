package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// profileCmd represents the profile command
var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Show authentication and profile info",
	Long: `Display the current authentication status and token information.

Examples:
  # Show current token info
  tig profile`,
	RunE: func(cmd *cobra.Command, args []string) error {
		provider, token, err := manager.LoadToken()
		if err != nil {
			fmt.Println("Status: Not authenticated")
			fmt.Println("Run 'tig auth' to authenticate")
			return nil
		}

		maskedToken := maskToken(token)
		fmt.Println("Token:", maskedToken)

		strategy, err := manager.GetStrategy(provider)
		if err != nil {
			return fmt.Errorf("failed to get strategy: %w", err)
		}

		// Fetch user info from GitHub
		user, err := strategy.FetchUser(token)
		if err != nil {
			fmt.Println("Status: Token invalid or expired")
			fmt.Println("Run 'tig auth' to re-authenticate")
			return nil
		}

		fmt.Println("Status: Authenticated")
		fmt.Printf("Provider: %s\n", provider)
		fmt.Println("Username:", user.Login)
		if user.Name != "" {
			fmt.Println("Name:", user.Name)
		}
		if user.Email != "" {
			fmt.Println("Email:", user.Email)
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
