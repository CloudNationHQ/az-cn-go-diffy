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

type GitHubConfig struct {
	RepoOwner string
	RepoName  string
	Token     string
}

type GitHubIssueManager struct {
	GitHubConfig
	Client *http.Client
}

func NewGitHubIssueManager(repoOwner, repoName, token string) *GitHubIssueManager {
	return &GitHubIssueManager{
		GitHubConfig: GitHubConfig{
			RepoOwner: repoOwner,
			RepoName:  repoName,
			Token:     token,
		},
		Client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (manager *GitHubIssueManager) CreateOrUpdateIssue(ctx context.Context, findings []ValidationFinding) error {
	if len(findings) == 0 {
		return nil
	}

	dedup := make(map[string]ValidationFinding)

	for _, finding := range findings {
		key := fmt.Sprintf("%s|%s|%s|%v|%v|%s",
			finding.ResourceType,
			strings.ReplaceAll(finding.Path, "root.", ""),
			finding.Name,
			finding.IsBlock,
			finding.IsDataSource,
			finding.SubmoduleName,
		)
		dedup[key] = finding
	}

	var newBody bytes.Buffer

	for _, finding := range dedup {
		cleanPath := strings.ReplaceAll(finding.Path, "root.", "")
		status := "optional"
		if finding.Required {
			status = "required"
		}

		itemType := "property"
		if finding.IsBlock {
			itemType = "block"
		}

		entityType := "resource"
		if finding.IsDataSource {
			entityType = "data source"
		}

		if finding.SubmoduleName == "" {
			fmt.Fprintf(&newBody, "`%s`: missing %s %s `%s` in `%s` (%s)\n\n",
				finding.ResourceType, status, itemType, finding.Name, cleanPath, entityType,
			)
		} else {
			fmt.Fprintf(&newBody, "`%s`: missing %s %s `%s` in `%s` in submodule `%s` (%s)\n\n",
				finding.ResourceType, status, itemType, finding.Name, cleanPath, finding.SubmoduleName, entityType,
			)
		}
	}

	title := "Generated schema validation"
	issueNum, _, err := manager.findExistingIssue(ctx, title)
	if err != nil {
		return err
	}

	finalBody := newBody.String()
	if issueNum > 0 {
		return manager.updateIssue(ctx, issueNum, finalBody)
	}

	return manager.createIssue(ctx, title, finalBody)
}

func (manager *GitHubIssueManager) CloseExistingIssuesIfEmpty(ctx context.Context) error {
	title := "Generated schema validation"
	issueNum, _, err := manager.findExistingIssue(ctx, title)
	if err != nil {
		return fmt.Errorf("error finding existing issues: %w", err)
	}

	if issueNum <= 0 {
		return nil
	}

	return manager.closeIssue(ctx, issueNum, "All schema validation issues have been resolved. Closing this issue automatically.")
}

func (manager *GitHubIssueManager) findExistingIssue(ctx context.Context, title string) (int, string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues?state=open", manager.RepoOwner, manager.RepoName)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, "", &GitHubError{
			Operation: "find existing issue",
			Message:   "failed to create request",
			Err:       err,
		}
	}

	req.Header.Set("Authorization", "token "+manager.Token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := manager.Client.Do(req)
	if err != nil {
		return 0, "", &GitHubError{
			Operation: "find existing issue",
			Message:   "request failed",
			Err:       err,
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, "", &GitHubError{
			Operation: "find existing issue",
			Message:   fmt.Sprintf("API error: %s", resp.Status),
			Err:       fmt.Errorf("response: %s", string(body)),
		}
	}

	var issues []struct {
		Number int    `json:"number"`
		Title  string `json:"title"`
		Body   string `json:"body"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&issues); err != nil {
		return 0, "", &GitHubError{
			Operation: "find existing issue",
			Message:   "failed to decode response",
			Err:       err,
		}
	}

	for _, issue := range issues {
		if issue.Title == title {
			return issue.Number, issue.Body, nil
		}
	}

	return 0, "", nil
}

func (manager *GitHubIssueManager) updateIssue(ctx context.Context, issueNumber int, body string) error {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues/%d", manager.RepoOwner, manager.RepoName, issueNumber)
	payload := struct {
		Body string `json:"body"`
	}{Body: body}

	data, err := json.Marshal(payload)
	if err != nil {
		return &GitHubError{
			Operation: "update issue",
			Message:   "failed to marshal payload",
			Err:       err,
		}
	}

	req, err := http.NewRequestWithContext(ctx, "PATCH", url, bytes.NewReader(data))
	if err != nil {
		return &GitHubError{
			Operation: "update issue",
			Message:   "failed to create request",
			Err:       err,
		}
	}

	req.Header.Set("Authorization", "token "+manager.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := manager.Client.Do(req)
	if err != nil {
		return &GitHubError{
			Operation: "update issue",
			Message:   "request failed",
			Err:       err,
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return &GitHubError{
			Operation: "update issue",
			Message:   fmt.Sprintf("API error: %s", resp.Status),
			Err:       fmt.Errorf("response: %s", string(body)),
		}
	}

	return nil
}

func (manager *GitHubIssueManager) createIssue(ctx context.Context, title, body string) error {
	payload := struct {
		Title string `json:"title"`
		Body  string `json:"body"`
	}{
		Title: title,
		Body:  body,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return &GitHubError{
			Operation: "create issue",
			Message:   "failed to marshal payload",
			Err:       err,
		}
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues", manager.RepoOwner, manager.RepoName)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(data))
	if err != nil {
		return &GitHubError{
			Operation: "create issue",
			Message:   "failed to create request",
			Err:       err,
		}
	}

	req.Header.Set("Authorization", "token "+manager.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := manager.Client.Do(req)
	if err != nil {
		return &GitHubError{
			Operation: "create issue",
			Message:   "request failed",
			Err:       err,
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return &GitHubError{
			Operation: "create issue",
			Message:   fmt.Sprintf("API error: %s", resp.Status),
			Err:       fmt.Errorf("response: %s", string(body)),
		}
	}

	return nil
}

func (manager *GitHubIssueManager) closeIssue(ctx context.Context, issueNumber int, comment string) error {
	if comment != "" {
		if err := manager.addComment(ctx, issueNumber, comment); err != nil {
			return &GitHubError{
				Operation: "close issue",
				Message:   "failed to add closing comment",
				Err:       err,
			}
		}
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues/%d", manager.RepoOwner, manager.RepoName, issueNumber)
	payload := struct {
		State string `json:"state"`
	}{State: "closed"}

	data, err := json.Marshal(payload)
	if err != nil {
		return &GitHubError{
			Operation: "close issue",
			Message:   "failed to marshal payload",
			Err:       err,
		}
	}

	req, err := http.NewRequestWithContext(ctx, "PATCH", url, bytes.NewReader(data))
	if err != nil {
		return &GitHubError{
			Operation: "close issue",
			Message:   "failed to create request",
			Err:       err,
		}
	}

	req.Header.Set("Authorization", "token "+manager.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := manager.Client.Do(req)
	if err != nil {
		return &GitHubError{
			Operation: "close issue",
			Message:   "request failed",
			Err:       err,
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return &GitHubError{
			Operation: "close issue",
			Message:   fmt.Sprintf("API error: %s", resp.Status),
			Err:       fmt.Errorf("response: %s", string(body)),
		}
	}

	return nil
}

func (manager *GitHubIssueManager) addComment(ctx context.Context, issueNumber int, comment string) error {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues/%d/comments", manager.RepoOwner, manager.RepoName, issueNumber)
	payload := struct {
		Body string `json:"body"`
	}{Body: comment}

	data, err := json.Marshal(payload)
	if err != nil {
		return &GitHubError{
			Operation: "add comment",
			Message:   "failed to marshal payload",
			Err:       err,
		}
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(data))
	if err != nil {
		return &GitHubError{
			Operation: "add comment",
			Message:   "failed to create request",
			Err:       err,
		}
	}

	req.Header.Set("Authorization", "token "+manager.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := manager.Client.Do(req)
	if err != nil {
		return &GitHubError{
			Operation: "add comment",
			Message:   "request failed",
			Err:       err,
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return &GitHubError{
			Operation: "add comment",
			Message:   fmt.Sprintf("API error: %s", resp.Status),
			Err:       fmt.Errorf("response: %s", string(body)),
		}
	}
	return nil
}
