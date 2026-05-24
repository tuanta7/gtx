package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"

	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/spf13/cobra"
	"github.com/tuanta7/gtx/internal/auth"
	internalgit "github.com/tuanta7/gtx/internal/git"
)

var (
	prunePath      string
	pruneRemote    string
	pruneRemoteURL string
	pruneBranch    string
	pruneMessage   string
	pruneYes       bool
)

type promptSession struct {
	in  *bufio.Reader
	out io.Writer
	yes bool
}

type pruneInput struct {
	Path          string
	RemoteName    string
	RemoteURL     string
	BranchName    string
	CommitMessage string
	ForcePush     bool
}

var pruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "Reset a branch history to a single fresh commit",
	Long: `Prune rewrites the selected branch history into a single fresh commit.
If the target path is not already a git repository, it can clone the remote
first and then rewrite the selected branch.

WARNING: This operation is destructive and can force-push rewritten history.`,
	Example: `  # chore: reinit project in the current repository
  gtx prune

  # chore: reinit project in a specific path
  gtx prune --path /path/to/repo

  # Skip prompts and use defaults
  gtx prune -y`,
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := resolvePrunePath(prunePath)
		if err != nil {
			return err
		}

		prompt := promptSession{
			in:  bufio.NewReader(cmd.InOrStdin()),
			out: cmd.OutOrStdout(),
			yes: pruneYes,
		}

		input, repo, err := buildPruneInput(path, pruneRemote, pruneRemoteURL, pruneBranch, pruneMessage, &prompt)
		if err != nil {
			return err
		}

		fmt.Fprintf(prompt.out, "Repository path: %s\n", input.Path)
		fmt.Fprintf(prompt.out, "Branch name: %s\n", input.BranchName)
		fmt.Fprintf(prompt.out, "Commit message: %s\n", input.CommitMessage)
		fmt.Fprintf(prompt.out, "Remote name: %s\n", input.RemoteName)
		fmt.Fprintf(prompt.out, "Origin URL: %s\n", input.RemoteURL)
		if input.ForcePush {
			fmt.Fprintln(prompt.out, "Force push: enabled")
		} else {
			fmt.Fprintln(prompt.out, "Force push: disabled")
		}

		var pushAuth *githttp.BasicAuth
		if input.ForcePush {
			pushAuth, err = pushAuthForRemote(input.RemoteURL, prompt.out)
			if err != nil {
				return err
			}
		}

		if err := repo.Prune(internalgit.PruneOptions{
			Path:          input.Path,
			RemoteName:    input.RemoteName,
			RemoteURL:     input.RemoteURL,
			BranchName:    input.BranchName,
			CommitMessage: input.CommitMessage,
			ForcePush:     input.ForcePush,
			Auth:          pushAuth,
		}); err != nil {
			return fmt.Errorf("failed to prune history: %w", err)
		}

		fmt.Fprintln(prompt.out, "Successfully rewrote branch history.")
		return nil
	},
}

func pushAuthForRemote(remoteURL string, out io.Writer) (*githttp.BasicAuth, error) {
	parsedURL, err := url.Parse(remoteURL)
	if err != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
		return nil, nil
	}

	if !strings.EqualFold(parsedURL.Hostname(), "github.com") {
		return nil, nil
	}

	_, tokenValue, err := auth.LoadToken()
	if err != nil {
		if errors.Is(err, auth.ErrAuthRequired) {
			fmt.Fprintln(out, "Authentication required. Run 'gtx auth'.")
			return nil, err
		}
		return nil, fmt.Errorf("failed to load authentication token: %w", err)
	}

	return &githttp.BasicAuth{
		Username: "oauth2",
		Password: tokenValue,
	}, nil
}

func buildPruneInput(path, remoteName, remoteURL, branchName, commitMessage string, prompt *promptSession) (pruneInput, *internalgit.Repository, error) {
	pathValue, err := prompt.promptText("Repository path", path)
	if err != nil {
		return pruneInput{}, nil, err
	}

	repo, err := internalgit.OpenRepository(pathValue)
	if err != nil && !errors.Is(err, internalgit.ErrRepositoryNotFound) {
		return pruneInput{}, nil, err
	}

	if errors.Is(err, internalgit.ErrRepositoryNotFound) {
		cloneURL, promptErr := prompt.promptTextRequired("Remote URL to clone", remoteURL)
		if promptErr != nil {
			return pruneInput{}, nil, promptErr
		}

		repo, err = internalgit.CloneRepository(pathValue, remoteName, cloneURL)
		if err != nil {
			return pruneInput{}, nil, fmt.Errorf("failed to clone repository: %w", err)
		}

		remoteURL = cloneURL
	}

	defaultBranch := branchName
	if defaultBranch == "" {
		defaultBranch, err = repo.CurrentBranch()
		if err != nil {
			return pruneInput{}, nil, fmt.Errorf("failed to detect current branch: %w", err)
		}
	}

	branchValue, err := prompt.promptTextRequired("Branch name", defaultBranch)
	if err != nil {
		return pruneInput{}, nil, err
	}

	if err := repo.CheckoutBranch(branchValue, remoteName); err != nil {
		return pruneInput{}, nil, fmt.Errorf("failed to checkout branch %q: %w", branchValue, err)
	}

	commitValue, err := prompt.promptTextRequired("Commit message", firstNonEmpty(commitMessage, "chore: reinit project"))
	if err != nil {
		return pruneInput{}, nil, err
	}

	defaultRemoteURL := remoteURL
	if defaultRemoteURL == "" {
		defaultRemoteURL, err = repo.GetRemoteURL(remoteName)
		if err != nil {
			return pruneInput{}, nil, fmt.Errorf("failed to detect remote URL: %w", err)
		}
	}

	originURL, err := prompt.promptTextRequired("Origin URL", defaultRemoteURL)
	if err != nil {
		return pruneInput{}, nil, err
	}

	forcePush, err := prompt.confirm("Force push rewritten branch", true)
	if err != nil {
		return pruneInput{}, nil, err
	}

	return pruneInput{
		Path:          pathValue,
		RemoteName:    remoteName,
		RemoteURL:     originURL,
		BranchName:    branchValue,
		CommitMessage: commitValue,
		ForcePush:     forcePush,
	}, repo, nil
}

func resolvePrunePath(flagValue string) (string, error) {
	if strings.TrimSpace(flagValue) != "" {
		return flagValue, nil
	}

	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	return cwd, nil
}

func (p *promptSession) promptText(label, defaultValue string) (string, error) {
	if p.yes {
		return defaultValue, nil
	}

	if defaultValue == "" {
		fmt.Fprintf(p.out, "%s: ", label)
	} else {
		fmt.Fprintf(p.out, "%s [%s]: ", label, defaultValue)
	}

	value, err := p.readLine()
	if err != nil {
		return "", err
	}

	if value == "" {
		return defaultValue, nil
	}

	return value, nil
}

func (p *promptSession) promptTextRequired(label, defaultValue string) (string, error) {
	value, err := p.promptText(label, defaultValue)
	if err != nil {
		return "", err
	}

	if strings.TrimSpace(value) == "" {
		return "", fmt.Errorf("%s is required", strings.ToLower(label))
	}

	return value, nil
}

func (p *promptSession) confirm(label string, defaultYes bool) (bool, error) {
	if p.yes {
		return defaultYes, nil
	}

	suffix := "[y/N]"
	if defaultYes {
		suffix = "[Y/n]"
	}

	for {
		fmt.Fprintf(p.out, "%s %s: ", label, suffix)

		value, err := p.readLine()
		if err != nil {
			return false, err
		}

		switch strings.ToLower(value) {
		case "":
			return defaultYes, nil
		case "y", "yes":
			return true, nil
		case "n", "no":
			return false, nil
		default:
			fmt.Fprintln(p.out, "Please answer y or n.")
		}
	}
}

func (p *promptSession) readLine() (string, error) {
	value, err := p.in.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}

	return strings.TrimSpace(value), nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}

	return ""
}

func init() {
	rootCmd.AddCommand(pruneCmd)

	pruneCmd.Flags().StringVarP(&prunePath, "path", "p", "", "Path to the git repository (default: current directory)")
	pruneCmd.Flags().StringVarP(&pruneRemote, "remote", "r", "origin", "Remote name to push to")
	pruneCmd.Flags().StringVar(&pruneRemoteURL, "origin-url", "", "Origin URL to clone from or reconfigure before pushing")
	pruneCmd.Flags().StringVarP(&pruneBranch, "branch", "b", "", "Branch name to rewrite (default: current or remote HEAD branch)")
	pruneCmd.Flags().StringVarP(&pruneMessage, "message", "m", "chore: reinit project", "Commit message for the new initial commit")
	pruneCmd.Flags().BoolVarP(&pruneYes, "yes", "y", false, "Use the default answer for every prompt")
}
