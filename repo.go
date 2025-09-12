package diffy

import (
	"os"
	"strings"
)

type GitRepoInfo struct {
	TerraformRoot string
}

func NewGitRepoInfo(terraformRoot string) *GitRepoInfo {
	return &GitRepoInfo{
		TerraformRoot: terraformRoot,
	}
}

func (g *GitRepoInfo) GetRepoInfo() (owner, repo string) {
	owner = os.Getenv("GITHUB_REPOSITORY_OWNER")
	repo = os.Getenv("GITHUB_REPOSITORY_NAME")
	if owner != "" && repo != "" {
		return owner, repo
	}

	if ghRepo := os.Getenv("GITHUB_REPOSITORY"); ghRepo != "" {
		parts := strings.SplitN(ghRepo, "/", 2)
		if len(parts) == 2 {
			owner, repo = parts[0], parts[1]
			return owner, repo
		}
	}
	return "", ""
}
