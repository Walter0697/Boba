package installer

// GitHubClientInterface defines the interface for GitHub operations needed by the installer
type GitHubClientInterface interface {
	GetRepositoryContents(path string) ([]byte, error)
}