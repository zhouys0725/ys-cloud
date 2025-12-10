package git

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/sirupsen/logrus"
)

type GitService struct {
	tempDir string
	logger  *logrus.Logger
}

type GitRepo struct {
	URL      string
	Branch   string
	Tag      string
	Commit   string
	RepoPath string
}

type GitWebhookPayload struct {
	Event      string                 `json:"event"`
	Repository GitWebhookRepository  `json:"repository"`
	Ref        string                `json:"ref"`
	Commit     GitWebhookCommit      `json:"commit"`
	Headers    map[string]string     `json:"headers"`
}

type GitWebhookRepository struct {
	Name        string `json:"name"`
	FullName    string `json:"full_name"`
	CloneURL    string `json:"clone_url"`
	DefaultBranch string `json:"default_branch"`
}

type GitWebhookCommit struct {
	ID      string `json:"id"`
	Message string `json:"message"`
	URL     string `json:"url"`
}

func NewGitService() *GitService {
	tempDir := filepath.Join(os.TempDir(), "ys-cloud-repos")
	os.MkdirAll(tempDir, 0755)

	return &GitService{
		tempDir: tempDir,
		logger:  logrus.New(),
	}
}

func (s *GitService) Clone(repoURL, branch, tag, username, password string) (*GitRepo, error) {
	repoName := strings.TrimSuffix(filepath.Base(repoURL), ".git")
	repoPath := filepath.Join(s.tempDir, fmt.Sprintf("%s-%d", repoName, os.Getpid()))

	auth := &http.BasicAuth{
		Username: username,
		Password: password,
	}

	var ref plumbing.ReferenceName
	if tag != "" {
		ref = plumbing.NewTagReferenceName(tag)
	} else if branch != "" {
		ref = plumbing.NewBranchReferenceName(branch)
	} else {
		ref = plumbing.HEAD
	}

	repo, err := git.PlainClone(repoPath, false, &git.CloneOptions{
		URL:           repoURL,
		Auth:          auth,
		ReferenceName: ref,
		SingleBranch:  true,
		Depth:         1,
	})

	if err != nil {
		os.RemoveAll(repoPath)
		return nil, fmt.Errorf("failed to clone repository: %w", err)
	}

	commitHash, err := s.getCommitHash(repo, ref)
	if err != nil {
		os.RemoveAll(repoPath)
		return nil, fmt.Errorf("failed to get commit hash: %w", err)
	}

	gitRepo := &GitRepo{
		URL:      repoURL,
		Branch:   branch,
		Tag:      tag,
		Commit:   commitHash,
		RepoPath: repoPath,
	}

	s.logger.WithFields(logrus.Fields{
		"repo_url": repoURL,
		"branch":   branch,
		"tag":      tag,
		"commit":   commitHash,
		"repo_path": repoPath,
	}).Info("Repository cloned successfully")

	return gitRepo, nil
}

func (s *GitService) getCommitHash(repo *git.Repository, ref plumbing.ReferenceName) (string, error) {
	if ref == plumbing.HEAD {
		head, err := repo.Head()
		if err != nil {
			return "", err
		}
		return head.Hash().String(), nil
	}

	hash, err := repo.ResolveRevision(plumbing.Revision(ref))
	if err != nil {
		return "", err
	}

	return hash.String(), nil
}

func (s *GitService) Checkout(repoPath, branchOrTag string) error {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	// Try to checkout as branch first
	if err := worktree.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(branchOrTag),
	}); err == nil {
		return nil
	}

	// If not a branch, try as tag
	return worktree.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewTagReferenceName(branchOrTag),
	})
}

func (s *GitService) Pull(repoPath, username, password string) error {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	auth := &http.BasicAuth{
		Username: username,
		Password: password,
	}

	err = worktree.Pull(&git.PullOptions{
		Auth:     auth,
		Force:    true,
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf("failed to pull changes: %w", err)
	}

	return nil
}

func (s *GitService) Cleanup(repoPath string) {
	if repoPath != "" {
		if err := os.RemoveAll(repoPath); err != nil {
			s.logger.WithError(err).WithField("repo_path", repoPath).Error("Failed to cleanup repository")
		} else {
			s.logger.WithField("repo_path", repoPath).Info("Repository cleaned up successfully")
		}
	}
}

func (s *GitService) ParseGitHubWebhook(payload []byte, headers map[string]string) (*GitWebhookPayload, error) {
	// TODO: Implement GitHub webhook parsing
	// This is a placeholder implementation
	return &GitWebhookPayload{
		Event: "push",
		Ref:   "refs/heads/main",
	}, nil
}

func (s *GitService) ParseGitLabWebhook(payload []byte, headers map[string]string) (*GitWebhookPayload, error) {
	// TODO: Implement GitLab webhook parsing
	// This is a placeholder implementation
	return &GitWebhookPayload{
		Event: "push",
		Ref:   "refs/heads/main",
	}, nil
}

func (s *GitService) ParseGiteeWebhook(payload []byte, headers map[string]string) (*GitWebhookPayload, error) {
	// TODO: Implement Gitee webhook parsing
	// This is a placeholder implementation
	return &GitWebhookPayload{
		Event: "push",
		Ref:   "refs/heads/main",
	}, nil
}