package cmd

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	gogit "github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/config"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/object"
	githttp "github.com/go-git/go-git/v6/plumbing/transport/http"
	"github.com/stretchr/testify/require"
	"github.com/tuanta7/gtx/internal/token"
)

func TestBuildPruneInputUsesDefaultsForExistingRepo(t *testing.T) {
	t.Parallel()

	remotePath := filepath.Join(t.TempDir(), "remote.git")
	repoPath := filepath.Join(t.TempDir(), "repo")

	createTestRemote(t, remotePath)
	repo := createTestRepo(t, repoPath, "main")
	testCommitFile(t, repo, repoPath, "file.txt", "content", "initial")
	testPushBranch(t, repo, "origin", remotePath, "main")

	inputStream := strings.NewReader("\n\n\n\ny\n")
	output := &bytes.Buffer{}
	prompt := &promptSession{
		in:  bufio.NewReader(inputStream),
		out: output,
	}

	input, openedRepo, err := buildPruneInput(repoPath, "origin", "", "", "", prompt)
	require.NoError(t, err)
	require.NotNil(t, openedRepo)
	require.Equal(t, repoPath, input.Path)
	require.Equal(t, "main", input.BranchName)
	require.Equal(t, "chore: reinit project", input.CommitMessage)
	require.Equal(t, remotePath, input.RemoteURL)
	require.True(t, input.ForcePush)
}

func TestBuildPruneInputRequiresRemoteURLInYesModeWhenNotRepo(t *testing.T) {
	t.Parallel()

	output := &bytes.Buffer{}
	prompt := &promptSession{
		in:  bufio.NewReader(strings.NewReader("")),
		out: output,
		yes: true,
	}

	_, _, err := buildPruneInput(filepath.Join(t.TempDir(), "missing"), "origin", "", "", "", prompt)
	require.EqualError(t, err, "remote url to clone is required")
}

func TestBuildPruneInputClonesAndPromptsForBranch(t *testing.T) {
	t.Parallel()

	sourcePath := filepath.Join(t.TempDir(), "source")
	clonePath := filepath.Join(t.TempDir(), "clone")

	sourceRepo := createTestRepo(t, sourcePath, "main")
	testCommitFile(t, sourceRepo, sourcePath, "file.txt", "main", "main")
	createFeatureBranch(t, sourceRepo, sourcePath)

	inputStream := strings.NewReader("\n" + sourcePath + "\nfeature\nchore: reinit project\n\nY\n")
	output := &bytes.Buffer{}
	prompt := &promptSession{
		in:  bufio.NewReader(inputStream),
		out: output,
	}

	input, repo, err := buildPruneInput(clonePath, "origin", "", "", "", prompt)
	require.NoError(t, err)
	require.NotNil(t, repo)
	require.Equal(t, "feature", input.BranchName)
	require.Equal(t, sourcePath, input.RemoteURL)

	currentBranch, err := repo.CurrentBranch()
	require.NoError(t, err)
	require.Equal(t, "feature", currentBranch)
}

func TestPushAuthForRemoteRequiresStoredAuthForGitHubHTTPS(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	output := &bytes.Buffer{}

	auth, err := pushAuthForRemote("https://github.com/openai/example.git", output)
	require.Nil(t, auth)
	require.ErrorIs(t, err, token.ErrAuthRequired)
	require.Contains(t, output.String(), "Authentication required. Run 'gtx auth'.")
}

func TestPushAuthForRemoteUsesStoredTokenForGitHubHTTPS(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	require.NoError(t, os.WriteFile(filepath.Join(home, ".netrc"), []byte("machine github.com\nlogin oauth2\npassword stored-token\n"), 0o600))

	auth, err := pushAuthForRemote("https://github.com/openai/example.git", &bytes.Buffer{})
	require.NoError(t, err)
	require.Equal(t, &githttp.BasicAuth{
		Username: "oauth2",
		Password: "stored-token",
	}, auth)
}

func TestPushAuthForRemoteSkipsNonGitHubOrNonHTTPRemotes(t *testing.T) {
	auth, err := pushAuthForRemote("git@github.com:openai/example.git", &bytes.Buffer{})
	require.NoError(t, err)
	require.Nil(t, auth)

	auth, err = pushAuthForRemote("/tmp/example.git", &bytes.Buffer{})
	require.NoError(t, err)
	require.Nil(t, auth)
}

func createTestRemote(t *testing.T, path string) {
	t.Helper()

	_, err := gogit.PlainInit(path, true, gogit.WithDefaultBranch(plumbing.NewBranchReferenceName("main")))
	require.NoError(t, err)
}

func createTestRepo(t *testing.T, path, branch string) *gogit.Repository {
	t.Helper()

	repo, err := gogit.PlainInit(path, false, gogit.WithDefaultBranch(plumbing.NewBranchReferenceName(branch)))
	require.NoError(t, err)

	return repo
}

func testCommitFile(t *testing.T, repo *gogit.Repository, repoPath, name, contents, message string) {
	t.Helper()

	require.NoError(t, os.WriteFile(filepath.Join(repoPath, name), []byte(contents), 0o644))

	worktree, err := repo.Worktree()
	require.NoError(t, err)

	_, err = worktree.Add(name)
	require.NoError(t, err)

	_, err = worktree.Commit(message, &gogit.CommitOptions{
		Author: &object.Signature{
			Name:  "test",
			Email: "test@example.com",
		},
	})
	require.NoError(t, err)
}

func testPushBranch(t *testing.T, repo *gogit.Repository, remoteName, remotePath, branch string) {
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

func createFeatureBranch(t *testing.T, repo *gogit.Repository, repoPath string) {
	t.Helper()

	worktree, err := repo.Worktree()
	require.NoError(t, err)

	require.NoError(t, worktree.Checkout(&gogit.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName("feature"),
		Create: true,
	}))

	testCommitFile(t, repo, repoPath, "feature.txt", "feature", "feature")
	require.NoError(t, worktree.Checkout(&gogit.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName("main"),
	}))
}
