package cmd

import (
	"fmt"
	"os"

	gogit "github.com/go-git/go-git/v5"
	"github.com/spf13/cobra"
	"github.com/tuanta7/gtx/internal/git"
)

// backCmd represents the back command
var backCmd = &cobra.Command{
	Use:   "back",
	Short: "Undo the last commit (soft reset to HEAD~1)",
	Long: `Undo the last commit by performing a soft reset to HEAD~1.

Run this command from inside a git repository. It will fail if the current HEAD has no parent 
(e.g. the repository has only a single commit).

Examples:
  # Soft-reset the last commit, keep changes staged/unstaged
  gtx back

  # After running ` + "`gtx back`" + ` you can adjust files, re-add, and recommit:
  git add .
  git commit -m "new message"
  git push -f
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}

		repo, err := git.OpenRepository(cwd)
		if err != nil {
			return err
		}

		worktree, err := repo.Worktree()
		if err != nil {
			return err
		}

		head, err := repo.Head()
		if err != nil {
			return err
		}

		headCommit, err := repo.CommitObject(head.Hash())
		if err != nil {
			return err
		}

		if len(headCommit.ParentHashes) == 0 {
			return fmt.Errorf("current HEAD has no parent commit")
		}

		if err := worktree.Reset(&gogit.ResetOptions{
			Commit: headCommit.ParentHashes[0],
			Mode:   gogit.SoftReset,
		}); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(backCmd)
}
