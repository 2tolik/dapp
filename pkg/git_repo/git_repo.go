package git_repo

type GitRepo struct {
}

func (gr *GitRepo) LatestCommit(branch string) (string, error) {
	return "deadbeef", nil
}
