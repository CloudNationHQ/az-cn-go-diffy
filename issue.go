package diffy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// GitHubIssueManager implements IssueManager for GitHub
type GitHubIssueManager struct {
	RepoOwner string
	RepoName  string
	Token     string
	Client    *http.Client
}

// NewGitHubIssueManager creates a new GitHub issue manager
func NewGitHubIssueManager(repoOwner, repoName, token string) *GitHubIssueManager {
	return &GitHubIssueManager{
		RepoOwner: repoOwner,
		RepoName:  repoName,
		Token:     token,
		Client:    &http.Client{Timeout: 10 * time.Second},
	}
}

// CreateOrUpdateIssue creates or updates a GitHub issue with validation findings
func (g *GitHubIssueManager) CreateOrUpdateIssue(ctx context.Context, findings []ValidationFinding) error {
	if len(findings) == 0 {
		return nil
	}

	dedup := make(map[string]ValidationFinding)

	for _, f := range findings {
		key := fmt.Sprintf("%s|%s|%s|%v|%v|%s",
			f.ResourceType,
			strings.ReplaceAll(f.Path, "root.", ""),
			f.Name,
			f.IsBlock,
			f.IsDataSource,
			f.SubmoduleName,
		)
		dedup[key] = f
	}

	var newBody bytes.Buffer
	fmt.Fprint(&newBody)

	for _, f := range dedup {
		cleanPath := strings.ReplaceAll(f.Path, "root.", "")
		status := "optional"
		if f.Required {
			status = "required"
		}

		itemType := "property"
		if f.IsBlock {
			itemType = "block"
		}

		entityType := "resource"
		if f.IsDataSource {
			entityType = "data source"
		}

		if f.SubmoduleName == "" {
			fmt.Fprintf(&newBody, "`%s`: missing %s %s `%s` in `%s` (%s)\n\n",
				f.ResourceType, status, itemType, f.Name, cleanPath, entityType,
			)
		} else {
			fmt.Fprintf(&newBody, "`%s`: missing %s %s `%s` in `%s` in submodule `%s` (%s)\n\n",
				f.ResourceType, status, itemType, f.Name, cleanPath, f.SubmoduleName, entityType,
			)
		}
	}

	title := "Generated schema validation"
	issueNum, _, err := g.findExistingIssue(ctx, title)
	if err != nil {
		return err
	}

	finalBody := newBody.String()
	if issueNum > 0 {
		// Update existing issue
		return g.updateIssue(ctx, issueNum, finalBody)
	}

	// Create new issue
	return g.createIssue(ctx, title, finalBody)
}

// findExistingIssue finds an existing GitHub issue with the given title
func (g *GitHubIssueManager) findExistingIssue(ctx context.Context, title string) (int, string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues?state=open", g.RepoOwner, g.RepoName)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "token "+g.Token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := g.Client.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, "", fmt.Errorf("GitHub API error: %s, response: %s", resp.Status, string(body))
	}

	var issues []struct {
		Number int    `json:"number"`
		Title  string `json:"title"`
		Body   string `json:"body"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&issues); err != nil {
		return 0, "", err
	}

	for _, issue := range issues {
		if issue.Title == title {
			return issue.Number, issue.Body, nil
		}
	}

	return 0, "", nil
}

// updateIssue updates an existing GitHub issue
func (g *GitHubIssueManager) updateIssue(ctx context.Context, issueNumber int, body string) error {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues/%d", g.RepoOwner, g.RepoName, issueNumber)
	payload := struct {
		Body string `json:"body"`
	}{Body: body}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "PATCH", url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "token "+g.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("GitHub API error: %s, response: %s", resp.Status, string(body))
	}

	return nil
}

// createIssue creates a new GitHub issue
func (g *GitHubIssueManager) createIssue(ctx context.Context, title, body string) error {
	payload := struct {
		Title string `json:"title"`
		Body  string `json:"body"`
	}{
		Title: title,
		Body:  body,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues", g.RepoOwner, g.RepoName)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "token "+g.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("GitHub API error: %s, response: %s", resp.Status, string(body))
	}

	return nil
}
