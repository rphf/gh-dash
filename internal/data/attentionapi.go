package data

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	gh "github.com/cli/go-gh/v2/pkg/api"

	"github.com/dlvhdr/gh-dash/v4/internal/attention"
)

type attentionActor struct {
	Typename string `json:"__typename"`
	Login    string `json:"login"`
}

func (a *attentionActor) event(at time.Time, state string) attention.Event {
	if a == nil {
		return attention.Event{At: at, State: state}
	}
	return attention.Event{Author: a.Login, Bot: a.Typename == "Bot", At: at, State: state}
}

type attentionNode struct {
	Commits struct {
		Nodes []struct {
			Commit struct {
				AuthoredDate time.Time `json:"authoredDate"`
			} `json:"commit"`
		} `json:"nodes"`
	} `json:"commits"`
	ReviewRequests struct {
		Nodes []struct {
			RequestedReviewer *attentionActor `json:"requestedReviewer"`
		} `json:"nodes"`
	} `json:"reviewRequests"`
	Reviews struct {
		Nodes []struct {
			Author      *attentionActor `json:"author"`
			SubmittedAt time.Time       `json:"submittedAt"`
			State       string          `json:"state"`
			Body        string          `json:"body"`
		} `json:"nodes"`
	} `json:"reviews"`
	Comments struct {
		Nodes []struct {
			Author    *attentionActor `json:"author"`
			CreatedAt time.Time       `json:"createdAt"`
		} `json:"nodes"`
	} `json:"comments"`
	ReviewThreads struct {
		Nodes []struct {
			IsResolved bool `json:"isResolved"`
			Comments   struct {
				Nodes []struct {
					Author    *attentionActor `json:"author"`
					CreatedAt time.Time       `json:"createdAt"`
				} `json:"nodes"`
			} `json:"comments"`
		} `json:"nodes"`
	} `json:"reviewThreads"`
}

const attentionFields = `... on PullRequest {
  commits(last: 1) { nodes { commit { authoredDate } } }
  reviewRequests(first: 20) { nodes { requestedReviewer { __typename ... on User { login } } } }
  reviews(last: 50) { nodes { author { __typename login } submittedAt state body } }
  comments(last: 50) { nodes { author { __typename login } createdAt } }
  reviewThreads(last: 50) { nodes { isResolved comments(last: 1) { nodes { author { __typename login } createdAt } } } }
}`

// FetchAttentionInputs returns the login of the viewer and, for each PR in
// urls, the reviews, comments, review threads, review requests and last
// commit time its attention status depends on, keyed by PR URL.
func FetchAttentionInputs(urls []string) (string, map[string]attention.Input, error) {
	res := map[string]attention.Input{}
	if len(urls) == 0 {
		return "", res, nil
	}

	c := client
	if c == nil {
		var err error
		if c, err = gh.DefaultGraphQLClient(); err != nil {
			return "", nil, err
		}
	}

	var params, fields []string
	variables := make(map[string]any, len(urls))
	for i, u := range urls {
		params = append(params, fmt.Sprintf("$u%d: URI!", i))
		fields = append(
			fields,
			fmt.Sprintf("p%d: resource(url: $u%d) { %s }", i, i, attentionFields),
		)
		variables[fmt.Sprintf("u%d", i)] = u
	}
	query := fmt.Sprintf("query PullRequestAttention(%s) { viewer { login } %s }",
		strings.Join(params, ", "), strings.Join(fields, " "))

	var response map[string]json.RawMessage
	if err := c.Do(query, variables, &response); err != nil {
		return "", nil, err
	}

	var viewer struct {
		Login string `json:"login"`
	}
	if err := json.Unmarshal(response["viewer"], &viewer); err != nil {
		return "", nil, err
	}

	for i, u := range urls {
		raw := response[fmt.Sprintf("p%d", i)]
		if len(raw) == 0 || string(raw) == "null" {
			continue
		}
		var node attentionNode
		if err := json.Unmarshal(raw, &node); err != nil {
			return "", nil, err
		}
		res[u] = node.input()
	}
	return viewer.Login, res, nil
}

func (n attentionNode) input() attention.Input {
	var in attention.Input
	if len(n.Commits.Nodes) > 0 {
		in.LastCommitAt = n.Commits.Nodes[0].Commit.AuthoredDate
	}
	for _, r := range n.ReviewRequests.Nodes {
		if r.RequestedReviewer != nil && r.RequestedReviewer.Login != "" {
			in.RequestedReviewers = append(in.RequestedReviewers, r.RequestedReviewer.Login)
		}
	}
	for _, r := range n.Reviews.Nodes {
		e := r.Author.event(r.SubmittedAt, r.State)
		e.HasBody = strings.TrimSpace(r.Body) != ""
		in.Reviews = append(in.Reviews, e)
	}
	for _, c := range n.Comments.Nodes {
		in.Comments = append(in.Comments, c.Author.event(c.CreatedAt, ""))
	}
	for _, t := range n.ReviewThreads.Nodes {
		thread := attention.Thread{Resolved: t.IsResolved}
		if len(t.Comments.Nodes) > 0 {
			last := t.Comments.Nodes[0]
			thread.Last = last.Author.event(last.CreatedAt, "")
		}
		in.Threads = append(in.Threads, thread)
	}
	return in
}
