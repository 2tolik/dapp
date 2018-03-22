package git_artifact

import (
	"github.com/flant/dapp/pkg/git_repo"
)

type GitArtifact struct {
	GitRepo git_repo.GitRepo

	Name               string
	As                 string
	Branch             string
	Commit             string
	Cwd                string
	Owner              string
	Group              string
	IncludePaths       []string
	ExcludePaths       []string
	StagesDependencies map[string][]string
	Paramshash         string
}

func (ga *GitArtifact) LatestCommit() (string, error) {
	return ga.GitRepo.LatestCommit(ga.Branch)
}
