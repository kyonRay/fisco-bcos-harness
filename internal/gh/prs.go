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

// MyPRs lists the caller's open PRs.
func MyPRs() ([]PR, error) {
	out, err := Run("pr", "list", "--author", "@me", "--state", "open", "--json", prListFields)
	if err != nil {
		return nil, err
	}
	var prs []PR
	if err := json.Unmarshal([]byte(out), &prs); err != nil {
		return nil, fmt.Errorf("parse gh pr list output: %w", err)
	}
	return prs, nil
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
