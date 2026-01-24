package git

import (
	"fmt"
	"os"

	gogit "github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/config"
)

func (r *Repository) Prune(path, remote, commitMessage string) error {
	repo, err := OpenRepository(path)
	if err != nil {
		return err
	}

	upstreamURL, err := repo.GetRemoteURL(remote)
	if err != nil {
		return err
	}

	err = os.RemoveAll(fmt.Sprintf("%s/.git", path))
	if err != nil {
		return err
	}

	repo, err = InitRepository(path)
	if err != nil {
		return err
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return err
	}

	err = worktree.AddWithOptions(&gogit.AddOptions{All: true})
	if err != nil {
		return err
	}

	_, err = worktree.Commit(commitMessage, &gogit.CommitOptions{
		AllowEmptyCommits: false,
	})
	if err != nil {
		return err
	}

	upstreamRemote, err := repo.CreateRemote(&config.RemoteConfig{
		Name: remote,
		URLs: []string{upstreamURL},
	})
	if err != nil {
		return err
	}

	err = upstreamRemote.Push(&gogit.PushOptions{
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
