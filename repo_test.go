package diffy

import (
	"os"
	"testing"
)

func TestNewGitRepoInfo(t *testing.T) {
	terraformRoot := "/test/terraform/root"
	repoInfo := NewGitRepoInfo(terraformRoot)

	if repoInfo == nil {
		t.Fatal("NewGitRepoInfo() should not return nil")
	}

	if repoInfo.TerraformRoot != terraformRoot {
		t.Errorf("TerraformRoot = %s, want %s", repoInfo.TerraformRoot, terraformRoot)
	}
}

func TestGitRepoInfo_GetRepoInfo_FromEnv(t *testing.T) {
	tests := []struct {
		name        string
		envOwner    string
		envRepo     string
		envCombined string
		wantOwner   string
		wantRepo    string
	}{
		{
			name:      "from separate env vars",
			envOwner:  "testowner",
			envRepo:   "testrepo",
			wantOwner: "testowner",
			wantRepo:  "testrepo",
		},
		{
			name:        "from combined GITHUB_REPOSITORY",
			envCombined: "testowner/testrepo",
			wantOwner:   "testowner",
			wantRepo:    "testrepo",
		},
		{
			name:      "no env vars set",
			wantOwner: "",
			wantRepo:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear env vars
			os.Unsetenv("GITHUB_REPOSITORY_OWNER")
			os.Unsetenv("GITHUB_REPOSITORY_NAME")
			os.Unsetenv("GITHUB_REPOSITORY")

			// Set test env vars
			if tt.envOwner != "" {
				os.Setenv("GITHUB_REPOSITORY_OWNER", tt.envOwner)
			}
			if tt.envRepo != "" {
				os.Setenv("GITHUB_REPOSITORY_NAME", tt.envRepo)
			}
			if tt.envCombined != "" {
				os.Setenv("GITHUB_REPOSITORY", tt.envCombined)
			}

			repoInfo := NewGitRepoInfo("/test")
			owner, repo := repoInfo.GetRepoInfo()

			if owner != tt.wantOwner {
				t.Errorf("owner = %q, want %q", owner, tt.wantOwner)
			}

			if repo != tt.wantRepo {
				t.Errorf("repo = %q, want %q", repo, tt.wantRepo)
			}

			// Clean up
			os.Unsetenv("GITHUB_REPOSITORY_OWNER")
			os.Unsetenv("GITHUB_REPOSITORY_NAME")
			os.Unsetenv("GITHUB_REPOSITORY")
		})
	}
}

func TestRepositoryInfoProvider_Interface(t *testing.T) {
	// Test that GitRepoInfo implements RepositoryInfoProvider interface
	var _ RepositoryInfoProvider = (*GitRepoInfo)(nil)

	repoInfo := NewGitRepoInfo("/test")

	// Should not panic
	owner, repo := repoInfo.GetRepoInfo()

	// Just verify it returns something (even if empty)
	t.Logf("GetRepoInfo() returned owner=%q, repo=%q", owner, repo)
}

func TestGitRepoInfo_MultipleCallsConsistency(t *testing.T) {
	// Set consistent env vars
	os.Setenv("GITHUB_REPOSITORY_OWNER", "testowner")
	os.Setenv("GITHUB_REPOSITORY_NAME", "testrepo")
	defer func() {
		os.Unsetenv("GITHUB_REPOSITORY_OWNER")
		os.Unsetenv("GITHUB_REPOSITORY_NAME")
	}()

	repoInfo := NewGitRepoInfo("/test")

	// Test multiple calls return consistent results
	owner1, repo1 := repoInfo.GetRepoInfo()
	owner2, repo2 := repoInfo.GetRepoInfo()

	if owner1 != owner2 {
		t.Errorf("Multiple calls should return same owner: %q vs %q", owner1, owner2)
	}

	if repo1 != repo2 {
		t.Errorf("Multiple calls should return same repo: %q vs %q", repo1, repo2)
	}
}

func TestGitRepoInfo_InvalidGitHubRepository(t *testing.T) {
	// Set invalid GITHUB_REPOSITORY value
	os.Setenv("GITHUB_REPOSITORY", "invalid-no-slash")
	defer os.Unsetenv("GITHUB_REPOSITORY")

	repoInfo := NewGitRepoInfo("/test")
	owner, repo := repoInfo.GetRepoInfo()

	// Should return empty strings for invalid format
	if owner != "" || repo != "" {
		t.Errorf("Invalid GITHUB_REPOSITORY should return empty, got owner=%q, repo=%q", owner, repo)
	}
}
