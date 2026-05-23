package git

import (
	"fmt"
	"path/filepath"
	"testing"

	gogit "github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/stretchr/testify/require"
)

func TestPruneRewritesHistoryAndPushes(t *testing.T) {
	t.Parallel()

	remotePath := filepath.Join(t.TempDir(), "remote.git")
	sourcePath := filepath.Join(t.TempDir(), "source")

	createRemoteRepo(t, remotePath)
	sourceRepo := createLocalRepo(t, sourcePath, "main")
	commitFile(t, sourceRepo, sourcePath, "file.txt", "before", "before rewrite")
	createAndPushBranch(t, sourceRepo, "origin", remotePath, "main")
	commitFile(t, sourceRepo, sourcePath, "file.txt", "after", "latest state")

	repo, err := OpenRepository(sourcePath)
	require.NoError(t, err)

	err = repo.Prune(PruneOptions{
		Path:          sourcePath,
		RemoteName:    "origin",
		RemoteURL:     remotePath,
		BranchName:    "main",
		CommitMessage: "chore: reinit project",
		ForcePush:     true,
	})
	require.NoError(t, err)

	prunedRepo, err := OpenRepository(sourcePath)
	require.NoError(t, err)

	head, err := prunedRepo.Head()
	require.NoError(t, err)

	commit, err := prunedRepo.CommitObject(head.Hash())
	require.NoError(t, err)
	require.Equal(t, "chore: reinit project", commit.Message)
	require.Len(t, commit.ParentHashes, 0)

	remoteRepo, err := gogit.PlainOpen(remotePath)
	require.NoError(t, err)

	remoteRef, err := remoteRepo.Reference(plumbing.NewBranchReferenceName("main"), true)
	require.NoError(t, err)
	require.Equal(t, head.Hash(), remoteRef.Hash())
}

func TestPrunePushFailureKeepsLocalRewrite(t *testing.T) {
	t.Parallel()

	sourcePath := filepath.Join(t.TempDir(), "source")
	sourceRepo := createLocalRepo(t, sourcePath, "main")
	commitFile(t, sourceRepo, sourcePath, "file.txt", "before", "before rewrite")

	repo, err := OpenRepository(sourcePath)
	require.NoError(t, err)

	err = repo.Prune(PruneOptions{
		Path:          sourcePath,
		RemoteName:    "origin",
		RemoteURL:     filepath.Join(t.TempDir(), "missing-remote.git"),
		BranchName:    "main",
		CommitMessage: "chore: reinit project",
		ForcePush:     true,
	})
	require.ErrorContains(t, err, "local rewrite succeeded but force push failed")

	rewrittenRepo, openErr := OpenRepository(sourcePath)
	require.NoError(t, openErr)

	head, headErr := rewrittenRepo.Head()
	require.NoError(t, headErr)
	commit, commitErr := rewrittenRepo.CommitObject(head.Hash())
	require.NoError(t, commitErr)
	require.Equal(t, "chore: reinit project", commit.Message)
	require.Len(t, commit.ParentHashes, 0)
}

func TestPruneRestoresOriginalGitDirWhenReplacementInitFails(t *testing.T) {
	sourcePath := filepath.Join(t.TempDir(), "source")
	sourceRepo := createLocalRepo(t, sourcePath, "main")
	commitFile(t, sourceRepo, sourcePath, "file.txt", "before", "before rewrite")

	repo, err := OpenRepository(sourcePath)
	require.NoError(t, err)

	originalInitRepository := initRepository
	initRepository = func(path, branchName string) (*Repository, error) {
		return nil, fmt.Errorf("boom")
	}
	t.Cleanup(func() {
		initRepository = originalInitRepository
	})

	err = repo.Prune(PruneOptions{
		Path:          sourcePath,
		RemoteName:    "origin",
		RemoteURL:     "unused",
		BranchName:    "main",
		CommitMessage: "chore: reinit project",
		ForcePush:     false,
	})
	require.ErrorContains(t, err, "failed to initialize replacement repository")

	restoredRepo, openErr := OpenRepository(sourcePath)
	require.NoError(t, openErr)

	head, headErr := restoredRepo.Head()
	require.NoError(t, headErr)
	commit, commitErr := restoredRepo.CommitObject(head.Hash())
	require.NoError(t, commitErr)
	require.Equal(t, "before rewrite", commit.Message)
}
