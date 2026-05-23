package git

import (
	"os"
	"path/filepath"
	"testing"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/require"
)

func TestCloneRepositoryAndCheckoutBranch(t *testing.T) {
	t.Parallel()

	remotePath := filepath.Join(t.TempDir(), "remote.git")
	sourcePath := filepath.Join(t.TempDir(), "source")

	createRemoteRepo(t, remotePath)
	sourceRepo := createLocalRepo(t, sourcePath, "main")
	commitFile(t, sourceRepo, sourcePath, "main.txt", "main", "main commit")
	createAndPushBranch(t, sourceRepo, "origin", remotePath, "main")
	createBranchCommitAndPush(t, sourceRepo, "origin", "feature", sourcePath, "feature.txt", "feature", remotePath)

	clonePath := filepath.Join(t.TempDir(), "clone")
	clonedRepo, err := CloneRepository(clonePath, "origin", remotePath)
	require.NoError(t, err)

	currentBranch, err := clonedRepo.CurrentBranch()
	require.NoError(t, err)
	require.Equal(t, "main", currentBranch)

	require.NoError(t, clonedRepo.CheckoutBranch("feature", "origin"))

	currentBranch, err = clonedRepo.CurrentBranch()
	require.NoError(t, err)
	require.Equal(t, "feature", currentBranch)

	url, err := clonedRepo.GetRemoteURL("origin")
	require.NoError(t, err)
	require.Equal(t, remotePath, url)
}

func TestPruneExcludesBackupGitDirectoryFromCommit(t *testing.T) {
	t.Parallel()

	sourcePath := filepath.Join(t.TempDir(), "source")
	sourceRepo := createLocalRepo(t, sourcePath, "main")
	commitFile(t, sourceRepo, sourcePath, "file.txt", "before", "before rewrite")

	repo, err := OpenRepository(sourcePath)
	require.NoError(t, err)

	err = repo.Prune(PruneOptions{
		Path:          sourcePath,
		RemoteName:    "origin",
		RemoteURL:     "https://example.invalid/repo.git",
		BranchName:    "main",
		CommitMessage: "chore: reinit project",
		ForcePush:     false,
	})
	require.NoError(t, err)

	prunedRepo, err := OpenRepository(sourcePath)
	require.NoError(t, err)

	head, err := prunedRepo.Head()
	require.NoError(t, err)

	commit, err := prunedRepo.CommitObject(head.Hash())
	require.NoError(t, err)

	tree, err := commit.Tree()
	require.NoError(t, err)

	err = tree.Files().ForEach(func(file *object.File) error {
		require.NotContains(t, file.Name, ".git.prune-backup-")
		return nil
	})
	require.NoError(t, err)
}

func createRemoteRepo(t *testing.T, path string) {
	t.Helper()

	_, err := gogit.PlainInit(path, true, gogit.WithDefaultBranch(plumbing.NewBranchReferenceName("main")))
	require.NoError(t, err)
}

func createLocalRepo(t *testing.T, path, branch string) *gogit.Repository {
	t.Helper()

	repo, err := gogit.PlainInit(path, false, gogit.WithDefaultBranch(plumbing.NewBranchReferenceName(branch)))
	require.NoError(t, err)

	cfg, err := repo.Config()
	require.NoError(t, err)
	cfg.User.Name = "test"
	cfg.User.Email = "test@example.com"
	require.NoError(t, repo.SetConfig(cfg))

	return repo
}

func commitFile(t *testing.T, repo *gogit.Repository, repoPath, name, contents, message string) plumbing.Hash {
	t.Helper()

	require.NoError(t, osWriteFile(filepath.Join(repoPath, name), []byte(contents)))

	worktree, err := repo.Worktree()
	require.NoError(t, err)

	_, err = worktree.Add(name)
	require.NoError(t, err)

	hash, err := worktree.Commit(message, &gogit.CommitOptions{
		Author: &object.Signature{
			Name:  "test",
			Email: "test@example.com",
		},
	})
	require.NoError(t, err)

	return hash
}

func createAndPushBranch(t *testing.T, repo *gogit.Repository, remoteName, remotePath, branch string) {
	t.Helper()

	if _, err := repo.Remote(remoteName); err != nil {
		_, err = repo.CreateRemote(&config.RemoteConfig{
			Name: remoteName,
			URLs: []string{remotePath},
		})
		require.NoError(t, err)
	}

	err := repo.Push(&gogit.PushOptions{
		RemoteName: remoteName,
		RemoteURL:  remotePath,
		RefSpecs: []config.RefSpec{
			config.RefSpec("+" + plumbing.NewBranchReferenceName(branch).String() + ":" + plumbing.NewBranchReferenceName(branch).String()),
		},
		Force: true,
	})
	require.NoError(t, err)
}

func createBranchCommitAndPush(t *testing.T, repo *gogit.Repository, remoteName, branch, repoPath, fileName, contents, remotePath string) {
	t.Helper()

	worktree, err := repo.Worktree()
	require.NoError(t, err)

	require.NoError(t, worktree.Checkout(&gogit.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(branch),
		Create: true,
	}))

	commitFile(t, repo, repoPath, fileName, contents, branch+" commit")
	createAndPushBranch(t, repo, remoteName, remotePath, branch)
}

func osWriteFile(path string, content []byte) error {
	return os.WriteFile(path, content, 0o644)
}
