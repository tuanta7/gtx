package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/tuanta7/gtx/internal/auth"
	"github.com/tuanta7/gtx/internal/config"
)

var ghClient *auth.GitHubClient

var rootCmd = &cobra.Command{
	Use:   "gtx",
	Short: "Git Extensions",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	getOrInitGitHubClient()

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

func getOrInitGitHubClient() *auth.GitHubClient {
	if ghClient == nil {
		ghClient = auth.NewGitHubClient(
			config.GitHubOAuthClientID,
			config.GitHubDeviceCodeURL,
			config.GitHubAccessTokenURL,
			config.GitHubProfileEndpoint,
		)
	}

	return ghClient
}
