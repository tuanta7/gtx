package manager

import (
	"errors"
	"fmt"
	"os"

	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/config"
)

var (
	ErrGitRepositoryNotFound = errors.New("git repository not found")
)

type Git struct {
}

func NewGit() *Git {
	return &Git{}
}

func (g *Git) OpenRepository(path string) (*git.Repository, error) {
	repo, err := git.PlainOpen(path)
	if err != nil {
		if errors.Is(err, git.ErrRepositoryNotExists) {
			return nil, ErrGitRepositoryNotFound
		}
		return nil, err
	}

	return repo, nil
}

func (g *Git) PruneHistory(path, remote, commitMessage string) error {
	repo, err := g.OpenRepository(path)
	if err != nil {
		return err
	}

	upstreamRemote, err := repo.Remote(remote)
	if err != nil {
		return err
	}

	err = os.RemoveAll(fmt.Sprintf("%s/.git", path))
	if err != nil {
		return err
	}

	upstreamURL := upstreamRemote.Config().URLs[0]
	repo, err = git.PlainInit(path, false)
	if err != nil {
		return err
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return err
	}

	err = worktree.AddWithOptions(&git.AddOptions{All: true})
	if err != nil {
		return err
	}

	_, err = worktree.Commit(commitMessage, &git.CommitOptions{
		AllowEmptyCommits: false,
	})
	if err != nil {
		return err
	}

	upstreamRemote, err = repo.CreateRemote(&config.RemoteConfig{
		Name: remote,
		URLs: []string{upstreamURL},
	})
	if err != nil {
		return err
	}

	err = upstreamRemote.Push(&git.PushOptions{
		RemoteName: remote,
		RemoteURL:  upstreamURL,
		Force:      true,
		Progress:   os.Stdout,
	})
	if err != nil {
		return err
	}

	return nil
}

func (g *Git) Init(path string) error {
	_, err := git.PlainInit(path, false)
	if err != nil {
		return err
	}

	return nil
}
