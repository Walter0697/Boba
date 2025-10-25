package github

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/google/go-github/v66/github"
	"golang.org/x/oauth2"
)

// GitHubClient handles GitHub API interactions
type GitHubClient struct {
	client *github.Client
	token  string
	owner  string
	repo   string
	ctx    context.Context
}

// AuthResult represents the result of GitHub authentication
type AuthResult struct {
	Success bool
	Error   error
	User    *github.User
}

// NewGitHubClient creates a new GitHub client instance
func NewGitHubClient(token, owner, repo string) *GitHubClient {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	return &GitHubClient{
		client: client,
		token:  token,
		owner:  owner,
		repo:   repo,
		ctx:    ctx,
	}
}

// ValidateToken validates the GitHub token and returns user information
func (gc *GitHubClient) ValidateToken() (*AuthResult, error) {
	if gc.token == "" {
		return &AuthResult{
			Success: false,
			Error:   fmt.Errorf("no GitHub token provided"),
		}, nil
	}

	// Test the token by getting the authenticated user
	user, _, err := gc.client.Users.Get(gc.ctx, "")
	if err != nil {
		return &AuthResult{
			Success: false,
			Error:   fmt.Errorf("invalid GitHub token: %w", err),
		}, nil
	}

	return &AuthResult{
		Success: true,
		Error:   nil,
		User:    user,
	}, nil
}

// ValidateRepositoryAccess checks if the client can access the specified repository
func (gc *GitHubClient) ValidateRepositoryAccess() error {
	if gc.owner == "" || gc.repo == "" {
		return fmt.Errorf("repository owner and name must be specified")
	}

	// Try to get repository information
	_, _, err := gc.client.Repositories.Get(gc.ctx, gc.owner, gc.repo)
	if err != nil {
		return fmt.Errorf("cannot access repository %s/%s: %w", gc.owner, gc.repo, err)
	}

	return nil
}

// ParseRepositoryURL extracts owner and repo from a GitHub URL
func ParseRepositoryURL(repoURL string) (owner, repo string, err error) {
	if repoURL == "" {
		return "", "", fmt.Errorf("repository URL cannot be empty")
	}

	// Handle different GitHub URL formats
	// https://github.com/owner/repo
	// https://github.com/owner/repo.git
	// git@github.com:owner/repo.git
	// owner/repo

	repoURL = strings.TrimSpace(repoURL)
	
	// Remove .git suffix if present
	repoURL = strings.TrimSuffix(repoURL, ".git")
	
	// Handle SSH format
	if strings.HasPrefix(repoURL, "git@github.com:") {
		repoURL = strings.TrimPrefix(repoURL, "git@github.com:")
	} else if strings.HasPrefix(repoURL, "https://github.com/") {
		repoURL = strings.TrimPrefix(repoURL, "https://github.com/")
	}
	
	// Split by /
	parts := strings.Split(repoURL, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid repository URL format: expected owner/repo")
	}
	
	owner = strings.TrimSpace(parts[0])
	repo = strings.TrimSpace(parts[1])
	
	if owner == "" || repo == "" {
		return "", "", fmt.Errorf("invalid repository URL: owner and repo cannot be empty")
	}
	
	return owner, repo, nil
}

// UpdateRepository updates the client's repository information
func (gc *GitHubClient) UpdateRepository(owner, repo string) {
	gc.owner = owner
	gc.repo = repo
}

// GetRepositoryContents fetches the contents of a file from the repository
func (gc *GitHubClient) GetRepositoryContents(path string) ([]byte, error) {
	if gc.owner == "" || gc.repo == "" {
		return nil, fmt.Errorf("repository owner and name must be specified")
	}

	fileContent, _, _, err := gc.client.Repositories.GetContents(gc.ctx, gc.owner, gc.repo, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get file %s: %w", path, err)
	}

	if fileContent == nil {
		return nil, fmt.Errorf("file %s not found", path)
	}

	content, err := fileContent.GetContent()
	if err != nil {
		return nil, fmt.Errorf("failed to decode file content: %w", err)
	}

	return []byte(content), nil
}

// GetDirectoryContents fetches the contents of a directory from the repository
func (gc *GitHubClient) GetDirectoryContents(path string) ([]string, error) {
	if gc.owner == "" || gc.repo == "" {
		return nil, fmt.Errorf("repository owner and name must be specified")
	}

	_, directoryContents, _, err := gc.client.Repositories.GetContents(gc.ctx, gc.owner, gc.repo, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get directory %s: %w", path, err)
	}

	var names []string
	for _, content := range directoryContents {
		if content.Name != nil {
			names = append(names, *content.Name)
		}
	}

	return names, nil
}

// TestConnection tests the GitHub connection and repository access
func (gc *GitHubClient) TestConnection() error {
	// First validate the token
	authResult, err := gc.ValidateToken()
	if err != nil {
		return fmt.Errorf("failed to validate token: %w", err)
	}

	if !authResult.Success {
		return authResult.Error
	}

	// Then validate repository access
	if err := gc.ValidateRepositoryAccess(); err != nil {
		return err
	}

	return nil
}

// GetToken returns the GitHub token
func (gc *GitHubClient) GetToken() string {
	return gc.token
}

// GetOwner returns the repository owner
func (gc *GitHubClient) GetOwner() string {
	return gc.owner
}

// GetRepo returns the repository name
func (gc *GitHubClient) GetRepo() string {
	return gc.repo
}

// GetFullRepoName returns the full repository name (owner/repo)
func (gc *GitHubClient) GetFullRepoName() string {
	return fmt.Sprintf("%s/%s", gc.owner, gc.repo)
}

// CloneRepository clones the repository to a local directory
func (gc *GitHubClient) CloneRepository(targetDir string) error {
	if gc.owner == "" || gc.repo == "" {
		return fmt.Errorf("repository owner and name must be specified")
	}

	// Check if git is available
	if _, err := exec.LookPath("git"); err != nil {
		return fmt.Errorf("git command not found - please install git: %w", err)
	}

	// Create target directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(targetDir), 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// Remove existing directory if it exists
	if _, err := os.Stat(targetDir); err == nil {
		if err := os.RemoveAll(targetDir); err != nil {
			return fmt.Errorf("failed to remove existing directory: %w", err)
		}
	}

	// Construct the clone URL using the token for authentication
	cloneURL := fmt.Sprintf("https://%s@github.com/%s/%s.git", gc.token, gc.owner, gc.repo)
	
	// Execute git clone command
	cmd := exec.Command("git", "clone", cloneURL, targetDir)
	
	// Capture output for debugging
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git clone failed for repository '%s/%s': %w\nOutput: %s", gc.owner, gc.repo, err, string(output))
	}

	return nil
}

// GetCloneTargetDir returns the default directory where the repository should be cloned
func (gc *GitHubClient) GetCloneTargetDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	
	return filepath.Join(homeDir, ".boba", "repos", gc.GetFullRepoName()), nil
}