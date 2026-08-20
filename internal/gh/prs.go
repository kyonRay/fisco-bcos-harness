package gh

import (
	"encoding/json"
	"fmt"
	"time"
)

// PR is the slice of gh pr list output the harness needs.
type PR struct {
	Number         int       `json:"number"`
	URL            string    `json:"url"`
	Title          string    `json:"title"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
	ReviewDecision string    `json:"reviewDecision"`
	ReviewRequests []struct {
		Login string `json:"login"`
	} `json:"reviewRequests"`
	LatestReviews []struct {
		Author struct {
			Login string `json:"login"`
		} `json:"author"`
		State       string    `json:"state"`
		SubmittedAt time.Time `json:"submittedAt"`
	} `json:"latestReviews"`
}

const prListFields = "number,url,title,createdAt,updatedAt,reviewDecision,reviewRequests,latestReviews"

func listPRs(args ...string) ([]PR, error) {
	out, err := Run(args...)
	if err != nil {
		return nil, err
	}
	var prs []PR
	if err := json.Unmarshal([]byte(out), &prs); err != nil {
		return nil, fmt.Errorf("parse gh pr list output: %w", err)
	}
	return prs, nil
}

// MyPRs lists the caller's open PRs.
func MyPRs() ([]PR, error) {
	return listPRs("pr", "list", "--author", "@me", "--state", "open", "--json", prListFields)
}

// Login returns the caller's GitHub login.
func Login() (string, error) {
	return Run("api", "user", "--jq", ".login")
}

// MyReviewTasks merges PRs where the caller is a requested reviewer
// with PRs the caller has already reviewed (still open), deduped by URL.
func MyReviewTasks() ([]PR, error) {
	requested, err := listPRs("pr", "list", "--search", "review-requested:@me", "--state", "open", "--json", prListFields)
	if err != nil {
		return nil, err
	}
	reviewed, err := listPRs("pr", "list", "--search", "reviewed-by:@me", "--state", "open", "--json", prListFields)
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	var out []PR
	for _, pr := range append(requested, reviewed...) {
		if seen[pr.URL] {
			continue
		}
		seen[pr.URL] = true
		out = append(out, pr)
	}
	return out, nil
}

// MyLastReview returns the state and time of login's newest review on
// this PR ("" if login hasn't reviewed it).
func (p PR) MyLastReview(login string) (string, time.Time) {
	state, at := "", time.Time{}
	for _, r := range p.LatestReviews {
		if r.Author.Login == login && r.SubmittedAt.After(at) {
			state, at = r.State, r.SubmittedAt
		}
	}
	return state, at
}

// LastReviewAt returns the newest review submission time (zero if none).
func (p PR) LastReviewAt() time.Time {
	var last time.Time
	for _, r := range p.LatestReviews {
		if r.SubmittedAt.After(last) {
			last = r.SubmittedAt
		}
	}
	return last
}

// Reviewers returns pending requested reviewers, falling back to past
// review authors (the re-review case).
func (p PR) Reviewers() []string {
	var out []string
	for _, r := range p.ReviewRequests {
		out = append(out, r.Login)
	}
	if len(out) > 0 {
		return out
	}
	for _, r := range p.LatestReviews {
		out = append(out, r.Author.Login)
	}
	return out
}

// NormalizedDecision maps gh reviewDecision to the harness's三态.
func (p PR) NormalizedDecision() string {
	switch p.ReviewDecision {
	case "APPROVED":
		return "approved"
	case "CHANGES_REQUESTED":
		return "changes_requested"
	default:
		return "none"
	}
}
