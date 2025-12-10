package service

import (
	"ys-cloud/pkg/git"
)

type GitService struct {
	*git.GitService
}

func NewGitService() *GitService {
	return &GitService{
		GitService: git.NewGitService(),
	}
}