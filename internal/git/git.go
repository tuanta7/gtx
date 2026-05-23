package git

import (
	"errors"
	"fmt"
	"strings"
	"time"

	gogit "github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/config"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/object"
)

var (
	ErrRepositoryNotFound = errors.New("git repository not found")
	ErrRemoteURLNotFound  = errors.New("remote URL not found")
	ErrBranchNotFound     = errors.New("git branch not found")
)

type Repository struct {
	*gogit.Repository
	globalConfig *config.Config
}

func OpenRepository(path string) (*Repository, error) {
	repo, err := gogit.PlainOpen(path)
	if err != nil {
		if errors.Is(err, gogit.ErrRepositoryNotExists) {
			return nil, ErrRepositoryNotFound
		}
		return nil, err
	}

	globalConfig, _ := config.LoadConfig(config.GlobalScope)

	return &Repository{
		Repository:   repo,
		globalConfig: globalConfig,
	}, nil
}

func CloneRepository(path, remoteName, remoteURL string) (*Repository, error) {
	repo, err := gogit.PlainClone(path, &gogit.CloneOptions{
		URL:        remoteURL,
		RemoteName: remoteName,
	})
	if err != nil {
		return nil, err
	}

	globalConfig, _ := config.LoadConfig(config.GlobalScope)

	return &Repository{
		Repository:   repo,
		globalConfig: globalConfig,
	}, nil
}

func InitRepository(path, branchName string) (*Repository, error) {
	repo, err := gogit.PlainInit(path, false, gogit.WithDefaultBranch(plumbing.NewBranchReferenceName(branchName)))
	if err != nil {
		return nil, err
	}

	globalConfig, _ := config.LoadConfig(config.GlobalScope)

	return &Repository{
		Repository:   repo,
		globalConfig: globalConfig,
	}, nil
}

func (r *Repository) GetRemoteURL(name string) (string, error) {
	rm, err := r.Remote(name)
	if err != nil {
		return "", err
	}

	url := rm.Config().URLs[0]
	if url == "" {
		return "", ErrRemoteURLNotFound
	}

	return url, nil
}

func (r *Repository) CurrentBranch() (string, error) {
	head, err := r.Head()
	if err != nil {
		return "", err
	}

	if !head.Name().IsBranch() {
		return "", ErrBranchNotFound
	}

	return head.Name().Short(), nil
}

func (r *Repository) CheckoutBranch(branchName, remoteName string) error {
	branchRef := plumbing.NewBranchReferenceName(branchName)
	worktree, err := r.Worktree()
	if err != nil {
		return err
	}

	if _, err := r.Reference(branchRef, true); err == nil {
		return worktree.Checkout(&gogit.CheckoutOptions{Branch: branchRef})
	}

	remoteRefName := plumbing.ReferenceName(fmt.Sprintf("refs/remotes/%s/%s", remoteName, branchName))
	remoteRef, err := r.Reference(remoteRefName, true)
	if err != nil {
		if errors.Is(err, plumbing.ErrReferenceNotFound) {
			return ErrBranchNotFound
		}
		return err
	}

	return worktree.Checkout(&gogit.CheckoutOptions{
		Branch: branchRef,
		Create: true,
		Hash:   remoteRef.Hash(),
	})
}

func (r *Repository) commitSignature() (*object.Signature, error) {
	cfg, err := r.Config()
	if err == nil {
		name := coalesce(cfg.Author.Name, cfg.User.Name, cfg.Committer.Name)
		email := coalesce(cfg.Author.Email, cfg.User.Email, cfg.Committer.Email)
		if name != "" && email != "" {
			return &object.Signature{
				Name:  name,
				Email: email,
				When:  time.Now(),
			}, nil
		}

		if r.globalConfig != nil && r.globalConfig.User.Name != "" && r.globalConfig.User.Email != "" {
			return &object.Signature{
				Name:  r.globalConfig.User.Name,
				Email: r.globalConfig.User.Email,
				When:  time.Now(),
			}, nil
		}

		return nil, fmt.Errorf("author/committer name and email not found in git config")
	}

	return nil, fmt.Errorf("failed to determine commit signature: %w", err)
}

func coalesce(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}

	return ""
}
