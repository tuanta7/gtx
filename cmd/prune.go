package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tuanta7/tig/internal/git"
)

var (
	prunePath    string
	pruneRemote  string
	pruneMessage string
)

// pruneCmd represents the prune command
var pruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "Delete all git history and create a fresh commit",
	Long: `Prune removes all git history from a repository and creates a single
fresh commit with all current files. This is useful when you want to
start fresh without any historical commits.

WARNING: This operation is destructive and will force push to the remote,
overwriting all existing history.

Examples:
  # Prune history in current directory
  tig prune

  # Prune history in a specific path
  tig prune --path /path/to/repo

  # Prune with custom remote and commit message
  tig prune --remote upstream --message "Fresh start"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		repo, err := git.OpenRepository(prunePath)
		if err != nil {
			return fmt.Errorf("failed to open repository: %w", err)
		}

		if prunePath == "" {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get current directory: %w", err)
			}
			prunePath = cwd
		}

		fmt.Printf("Pruning git history in: %s\n", prunePath)
		fmt.Printf("Remote: %s\n", pruneRemote)
		fmt.Printf("Commit message: %s\n", pruneMessage)

		err = repo.Prune(prunePath, pruneRemote, pruneMessage)
		if err != nil {
			return fmt.Errorf("failed to prune history: %w", err)
		}

		fmt.Println("Successfully pruned git history!")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(pruneCmd)

	pruneCmd.Flags().StringVarP(&prunePath, "path", "p", "", "Path to the git repository (default: current directory)")
	pruneCmd.Flags().StringVarP(&pruneRemote, "remote", "r", "origin", "Remote name to push to")
	pruneCmd.Flags().StringVarP(&pruneMessage, "message", "m", "rewrite history", "Commit message for the new initial commit")
}
