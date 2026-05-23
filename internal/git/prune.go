package git

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
	"github.com/go-git/go-git/v5/plumbing/transport"
)

type PruneOptions struct {
	Path          string
	RemoteName    string
	RemoteURL     string
	BranchName    string
	CommitMessage string
	ForcePush     bool
	Auth          transport.AuthMethod
}

func (r *Repository) Prune(options PruneOptions) error {
	signature, err := r.commitSignature()
	if err != nil {
		// failed quickly
		return err
	}

	liveGitPath := filepath.Join(options.Path, ".git")
	backupGitPath := filepath.Join(options.Path, fmt.Sprintf(".git.prune-backup-%d", time.Now().UnixNano()))

	if err := os.Rename(liveGitPath, backupGitPath); err != nil {
		return fmt.Errorf("failed to move current git directory aside: %w", err)
	}

	restore := func(cause error, message string) error {
		_ = os.RemoveAll(liveGitPath)
		restoreErr := os.Rename(backupGitPath, liveGitPath)
		if restoreErr != nil {
			return fmt.Errorf("%s: %w; restore failed: %v", message, cause, restoreErr)
		}

		return fmt.Errorf("%s: %w", message, cause)
	}

	repo, err := InitRepository(options.Path, options.BranchName)
	if err != nil {
		return restore(err, "failed to initialize replacement repository")
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return restore(err, "failed to open replacement worktree")
	}

	worktree.Excludes = append(worktree.Excludes, gitignore.ParsePattern(filepath.Base(backupGitPath)+"/", nil))

	if err := worktree.AddWithOptions(&gogit.AddOptions{All: true}); err != nil {
		return restore(err, "failed to add files to replacement repository")
	}

	if _, err := worktree.Commit(options.CommitMessage, &gogit.CommitOptions{
		Author:            signature,
		Committer:         signature,
		AllowEmptyCommits: false,
	}); err != nil {
		return restore(err, "failed to create replacement commit")
	}

	if _, err := repo.CreateRemote(&config.RemoteConfig{
		Name: options.RemoteName,
		URLs: []string{options.RemoteURL},
	}); err != nil {
		return restore(err, "failed to configure replacement remote")
	}

	if err := os.RemoveAll(backupGitPath); err != nil {
		return fmt.Errorf("prune succeeded but failed to remove backup git directory %q: %w", backupGitPath, err)
	}

	if !options.ForcePush {
		return nil
	}

	return r.forcePush(options)
}

func (r *Repository) forcePush(options PruneOptions) error {
	liveRepo, err := OpenRepository(options.Path)
	if err != nil {
		return fmt.Errorf("failed to reopen repository: %w", err)
	}

	err = liveRepo.Push(&gogit.PushOptions{
		RemoteName: options.RemoteName,
		RemoteURL:  options.RemoteURL,
		Auth:       options.Auth,
		RefSpecs: []config.RefSpec{
			config.RefSpec(fmt.Sprintf("+refs/heads/%[1]s:refs/heads/%[1]s", options.BranchName)),
		},
		Force: true,
	})
	if err != nil {
		return fmt.Errorf("local rewrite succeeded but force push failed: %w", err)
	}

	return nil
}
