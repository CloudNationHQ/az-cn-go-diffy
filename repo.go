package diffy

import (
	"os"
	"strings"
)

// RepositoryInfoProvider provides information about the repository
type RepositoryInfoProvider interface {
	GetRepoInfo() (owner, name string)
}

// GitRepoInfo implements RepositoryInfoProvider for Git repositories
type GitRepoInfo struct {
	TerraformRoot string
}

// NewGitRepoInfo creates a new Git repository info provider
func NewGitRepoInfo(terraformRoot string) *GitRepoInfo {
	return &GitRepoInfo{
		TerraformRoot: terraformRoot,
	}
}

// GetRepoInfo extracts repository information from environment variables
func (g *GitRepoInfo) GetRepoInfo() (owner, repo string) {
	// Try to get repository info from GitHub environment variables
	owner = os.Getenv("GITHUB_REPOSITORY_OWNER")
	repo = os.Getenv("GITHUB_REPOSITORY_NAME")
	if owner != "" && repo != "" {
		return owner, repo
	}

	// Try to get from combined GITHUB_REPOSITORY
	if ghRepo := os.Getenv("GITHUB_REPOSITORY"); ghRepo != "" {
		parts := strings.SplitN(ghRepo, "/", 2)
		if len(parts) == 2 {
			owner, repo = parts[0], parts[1]
			return owner, repo
		}
	}

	// Could extend with git command to get remote info
	// when not running in GitHub Actions

	return "", ""
}

// BoolToStr converts a boolean to a string representation
func BoolToStr(cond bool, yes, no string) string {
	if cond {
		return yes
	}
	return no
}
