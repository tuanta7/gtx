package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/tuanta7/tig/internal/config"
	"github.com/tuanta7/tig/internal/token"
)

var manager *token.Manager

var rootCmd = &cobra.Command{
	Use:   "tig",
	Short: "Delete old commits and rewrite repository history.",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	manager = token.NewManager()
	manager.Register(token.NewGitHubStrategy(
		config.GitHubOAuthClientID,
		config.GitHubDeviceCodeURL,
		config.GitHubAccessTokenURL,
		config.GitHubProfileEndpoint,
	))

	// Only support GitHub for now
	manager.Register(token.NewPATStrategy(
		token.GitHubProvider,
		config.GitHubTokensPage,
	))

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.tig.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
