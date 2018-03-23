package git_repo

import (
	"fmt"
	"gopkg.in/src-d/go-git.v4"
)

type GitRepo interface {
	LatestCommit(branch string) (string, error)
}

type BaseGitRepo struct {
	Name string
}

type LocalGitRepo struct {
	BaseGitRepo
	Path string
}

func (gr *LocalGitRepo) LatestCommit(branch string) (string, error) {
	repository, err := git.PlainOpen(gr.Path)
	if err != nil {
		return "", err
	}

	_ = repository

	return "", fmt.Errorf("not imlemented")
}

type RemoteGitRepo struct {
	BaseGitRepo
	Url string
}

func (gr *RemoteGitRepo) LatestCommit(branch string) (string, error) {
	return "", fmt.Errorf("not imlemented")
}
