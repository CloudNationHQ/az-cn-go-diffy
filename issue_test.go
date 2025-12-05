package diffy

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestCreateOrUpdateIssue_CreatesNew(t *testing.T) {
	var calls []recordedCall
	client := newStubHTTPClient(t, &calls, []httpHandlerStep{
		{
			method: "GET",
			path:   "/repos/o/r/issues",
			status: http.StatusOK,
			body:   `[]`,
		},
		{
			method: "POST",
			path:   "/repos/o/r/issues",
			status: http.StatusCreated,
			check: func(t *testing.T, r *http.Request) {
				if got := r.Header.Get("Authorization"); got != "token TOKEN" {
					t.Fatalf("expected auth header, got %q", got)
				}
			},
		},
	})

	manager := &GitHubIssueManager{
		GitHubConfig: GitHubConfig{
			RepoOwner: "o",
			RepoName:  "r",
			Token:     "TOKEN",
		},
		Client: client,
	}

	findings := []ValidationFinding{
		{ResourceType: "r1", Path: "root.attr", Name: "foo", Required: true},
		{ResourceType: "r1", Path: "root.attr", Name: "foo", Required: true}, // duplicate should be deduped
	}

	if err := manager.CreateOrUpdateIssue(context.Background(), findings); err != nil {
		t.Fatalf("CreateOrUpdateIssue returned error: %v", err)
	}

	if len(calls) != 2 {
		t.Fatalf("expected 2 API calls, got %d", len(calls))
	}

	body := calls[1].body.String()
	if !strings.Contains(body, "r1") || !strings.Contains(body, "`foo`") {
		t.Fatalf("issue body missing content: %q", body)
	}
}

func TestCreateOrUpdateIssue_UpdatesExisting(t *testing.T) {
	var calls []recordedCall
	client := newStubHTTPClient(t, &calls, []httpHandlerStep{
		{
			method: "GET",
			path:   "/repos/o/r/issues",
			status: http.StatusOK,
			body:   `[{"number":5,"title":"Generated schema validation","body":""}]`,
		},
		{
			method: "PATCH",
			path:   "/repos/o/r/issues/5",
			status: http.StatusOK,
		},
	})

	manager := &GitHubIssueManager{
		GitHubConfig: GitHubConfig{
			RepoOwner: "o",
			RepoName:  "r",
			Token:     "TOKEN",
		},
		Client: client,
	}

	if err := manager.CreateOrUpdateIssue(context.Background(), []ValidationFinding{
		{ResourceType: "r1", Path: "root.attr", Name: "bar"},
	}); err != nil {
		t.Fatalf("CreateOrUpdateIssue returned error: %v", err)
	}

	if len(calls) != 2 {
		t.Fatalf("expected 2 API calls, got %d", len(calls))
	}
}

func TestCloseExistingIssuesIfEmpty_AddsCommentAndCloses(t *testing.T) {
	var calls []recordedCall
	client := newStubHTTPClient(t, &calls, []httpHandlerStep{
		{
			method: "GET",
			path:   "/repos/o/r/issues",
			status: http.StatusOK,
			body:   `[{"number":9,"title":"Generated schema validation","body":""}]`,
		},
		{
			method: "POST",
			path:   "/repos/o/r/issues/9/comments",
			status: http.StatusCreated,
		},
		{
			method: "PATCH",
			path:   "/repos/o/r/issues/9",
			status: http.StatusOK,
		},
	})

	manager := &GitHubIssueManager{
		GitHubConfig: GitHubConfig{
			RepoOwner: "o",
			RepoName:  "r",
			Token:     "TOKEN",
		},
		Client: client,
	}

	if err := manager.CloseExistingIssuesIfEmpty(context.Background()); err != nil {
		t.Fatalf("CloseExistingIssuesIfEmpty returned error: %v", err)
	}

	if len(calls) != 3 {
		t.Fatalf("expected 3 API calls, got %d", len(calls))
	}
}

// --- helpers ---

type recordedCall struct {
	method string
	path   string
	body   bytes.Buffer
}

type httpHandlerStep struct {
	method string
	path   string
	status int
	body   string
	check  func(*testing.T, *http.Request)
}

type stubTransport struct {
	t     *testing.T
	steps []httpHandlerStep
	calls *[]recordedCall
}

func (st *stubTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	if len(st.steps) == 0 {
		st.t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
	}
	step := st.steps[0]
	st.steps = st.steps[1:]

	if r.Method != step.method || r.URL.Path != step.path {
		st.t.Fatalf("expected %s %s, got %s %s", step.method, step.path, r.Method, r.URL.Path)
	}

	var buf bytes.Buffer
	if r.Body != nil {
		_, _ = buf.ReadFrom(r.Body)
	}
	*st.calls = append(*st.calls, recordedCall{method: r.Method, path: r.URL.Path, body: buf})

	if step.check != nil {
		step.check(st.t, r)
	}

	resp := &http.Response{
		StatusCode: step.status,
		Body:       io.NopCloser(strings.NewReader(step.body)),
		Header:     make(http.Header),
	}
	return resp, nil
}

func newStubHTTPClient(t *testing.T, calls *[]recordedCall, steps []httpHandlerStep) *http.Client {
	t.Helper()
	return &http.Client{
		Transport: &stubTransport{
			t:     t,
			steps: steps,
			calls: calls,
		},
	}
}
