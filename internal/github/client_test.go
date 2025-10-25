package github

import (
	"testing"
)

func TestParseRepositoryURL(t *testing.T) {
	tests := []struct {
		name        string
		repoURL     string
		wantOwner   string
		wantRepo    string
		wantError   bool
	}{
		{
			name:      "HTTPS URL",
			repoURL:   "https://github.com/owner/repo",
			wantOwner: "owner",
			wantRepo:  "repo",
			wantError: false,
		},
		{
			name:      "HTTPS URL with .git",
			repoURL:   "https://github.com/owner/repo.git",
			wantOwner: "owner",
			wantRepo:  "repo",
			wantError: false,
		},
		{
			name:      "SSH URL",
			repoURL:   "git@github.com:owner/repo.git",
			wantOwner: "owner",
			wantRepo:  "repo",
			wantError: false,
		},
		{
			name:      "Simple format",
			repoURL:   "owner/repo",
			wantOwner: "owner",
			wantRepo:  "repo",
			wantError: false,
		},
		{
			name:      "Empty URL",
			repoURL:   "",
			wantError: true,
		},
		{
			name:      "Invalid format",
			repoURL:   "invalid-url",
			wantError: true,
		},
		{
			name:      "Missing repo",
			repoURL:   "owner/",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			owner, repo, err := ParseRepositoryURL(tt.repoURL)
			
			if tt.wantError {
				if err == nil {
					t.Errorf("ParseRepositoryURL() expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("ParseRepositoryURL() unexpected error: %v", err)
				return
			}
			
			if owner != tt.wantOwner {
				t.Errorf("ParseRepositoryURL() owner = %v, want %v", owner, tt.wantOwner)
			}
			
			if repo != tt.wantRepo {
				t.Errorf("ParseRepositoryURL() repo = %v, want %v", repo, tt.wantRepo)
			}
		})
	}
}

func TestNewGitHubClient(t *testing.T) {
	token := "test-token"
	owner := "test-owner"
	repo := "test-repo"
	
	client := NewGitHubClient(token, owner, repo)
	
	if client == nil {
		t.Fatal("NewGitHubClient() returned nil")
	}
	
	if client.token != token {
		t.Errorf("NewGitHubClient() token = %v, want %v", client.token, token)
	}
	
	if client.owner != owner {
		t.Errorf("NewGitHubClient() owner = %v, want %v", client.owner, owner)
	}
	
	if client.repo != repo {
		t.Errorf("NewGitHubClient() repo = %v, want %v", client.repo, repo)
	}
	
	if client.client == nil {
		t.Error("NewGitHubClient() GitHub client is nil")
	}
}

func TestGitHubClientGetters(t *testing.T) {
	token := "test-token"
	owner := "test-owner"
	repo := "test-repo"
	
	client := NewGitHubClient(token, owner, repo)
	
	if client.GetToken() != token {
		t.Errorf("GetToken() = %v, want %v", client.GetToken(), token)
	}
	
	if client.GetOwner() != owner {
		t.Errorf("GetOwner() = %v, want %v", client.GetOwner(), owner)
	}
	
	if client.GetRepo() != repo {
		t.Errorf("GetRepo() = %v, want %v", client.GetRepo(), repo)
	}
}

func TestUpdateRepository(t *testing.T) {
	client := NewGitHubClient("token", "old-owner", "old-repo")
	
	newOwner := "new-owner"
	newRepo := "new-repo"
	
	client.UpdateRepository(newOwner, newRepo)
	
	if client.GetOwner() != newOwner {
		t.Errorf("UpdateRepository() owner = %v, want %v", client.GetOwner(), newOwner)
	}
	
	if client.GetRepo() != newRepo {
		t.Errorf("UpdateRepository() repo = %v, want %v", client.GetRepo(), newRepo)
	}
}