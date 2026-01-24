package git

import (
	"errors"

	gogit "github.com/go-git/go-git/v6"
)

var (
	ErrRepositoryNotFound = errors.New("git repository not found")
	ErrRemoteURLNotFound  = errors.New("remote URL not found")
)

type Repository struct {
	*gogit.Repository
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
		repo,
	}, nil
}

func InitRepository(path string) (*Repository, error) {
	repo, err := gogit.PlainInit(path, false)
	if err != nil {
		return nil, err
	}

	return &Repository{
		repo,
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
