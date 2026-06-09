package git

import (
	"errors"
	"fmt"
	"strings"
	"time"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

var (
	ErrRepositoryNotFound = errors.New("git repository not found")
	ErrRemoteURLNotFound  = errors.New("remote URL not found")
	ErrBranchNotFound     = errors.New("git branch not found")
)

type Author struct {
	Name  string
	Email string
}

type Repository struct {
	*gogit.Repository
	author Author
}

func OpenRepository(path string) (*Repository, error) {
	repo, err := gogit.PlainOpen(path)
	if err != nil {
		if errors.Is(err, gogit.ErrRepositoryNotExists) {
			return nil, ErrRepositoryNotFound
		}
		return nil, err
	}

	return &Repository{
		Repository: repo,
		author:     getAuthor(),
	}, nil
}

func CloneRepository(path, remoteName, remoteURL string) (*Repository, error) {
	repo, err := gogit.PlainClone(path, false, &gogit.CloneOptions{
		URL:        remoteURL,
		RemoteName: remoteName,
	})
	if err != nil {
		return nil, err
	}

	return &Repository{
		Repository: repo,
		author:     getAuthor(),
	}, nil
}

func getAuthor() Author {
	localConfig, _ := config.LoadConfig(config.LocalScope)
	if author := getAuthorFromConfig(localConfig); author.Name != "" && author.Email != "" {
		return author
	}

	globalConfig, _ := config.LoadConfig(config.GlobalScope)
	return getAuthorFromConfig(globalConfig)
}

func InitRepository(path, branchName string) (*Repository, error) {
	repo, err := gogit.PlainInitWithOptions(path, &gogit.PlainInitOptions{
		InitOptions: gogit.InitOptions{
			DefaultBranch: plumbing.NewBranchReferenceName(branchName),
		},
		Bare: false,
	})
	if err != nil {
		return nil, err
	}

	return &Repository{
		Repository: repo,
		author:     getAuthor(),
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

func (r *Repository) Commit(message string, signature *object.Signature) (plumbing.Hash, error) {
	worktree, err := r.Worktree()
	if err != nil {
		return plumbing.ZeroHash, err
	}

	if signature == nil {
		signature, err = r.CommitSignature()
		if err != nil {
			return plumbing.ZeroHash, err
		}
	}

	return worktree.Commit(message, &gogit.CommitOptions{
		Author:    signature,
		Committer: signature,
	})
}

func (r *Repository) CommitSignature() (*object.Signature, error) {
	if r.author.Name != "" && r.author.Email != "" {
		return &object.Signature{
			Name:  r.author.Name,
			Email: r.author.Email,
			When:  time.Now(),
		}, nil
	}

	return nil, fmt.Errorf("author/committer name and email not found in git config")
}

func getAuthorFromConfig(cfg *config.Config) Author {
	if cfg == nil {
		return Author{}
	}

	return Author{
		Name:  coalesce(cfg.Author.Name, cfg.User.Name, cfg.Committer.Name),
		Email: coalesce(cfg.Author.Email, cfg.User.Email, cfg.Committer.Email),
	}
}

func coalesce(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}

	return ""
}
