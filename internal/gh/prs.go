package gh

import (
	"encoding/json"
	"fmt"
	"time"
)

// PR is the slice of gh pr list output the harness needs.
type PR struct {
	Number int    `json:"number"`
	URL    string `json:"url"`
	Title  string `json:"title"`
	State  string `json:"state"`
	Author struct {
		Login string `json:"login"`
	} `json:"author"`
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

const prViewFields = "url,title,state,author,createdAt,updatedAt,reviewDecision,reviewRequests,latestReviews"

// ViewPR fetches one PR's review-relevant state.
func ViewPR(ref string) (PR, error) {
	out, err := Run("pr", "view", ref, "--json", prViewFields)
	if err != nil {
		return PR{}, err
	}
	var pr PR
	if err := json.Unmarshal([]byte(out), &pr); err != nil {
		return PR{}, fmt.Errorf("parse gh pr view output: %w", err)
	}
	return pr, nil
}

// ApprovedBy returns reviewers whose latest review is APPROVED.
func (p PR) ApprovedBy() []string {
	var out []string
	for _, r := range p.LatestReviews {
		if r.State == "APPROVED" {
			out = append(out, r.Author.Login)
		}
	}
	return out
}

// PendingReviewers returns requested reviewers plus past reviewers who
// have not approved yet.
func (p PR) PendingReviewers() []string {
	approved := map[string]bool{}
	for _, login := range p.ApprovedBy() {
		approved[login] = true
	}
	seen := map[string]bool{}
	var out []string
	add := func(login string) {
		if login == "" || approved[login] || seen[login] {
			return
		}
		seen[login] = true
		out = append(out, login)
	}
	for _, r := range p.ReviewRequests {
		add(r.Login)
	}
	for _, r := range p.LatestReviews {
		add(r.Author.Login)
	}
	return out
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
