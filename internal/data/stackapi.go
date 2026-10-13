package data

import (
	"fmt"
	"strings"

	gh "github.com/cli/go-gh/v2/pkg/api"

	"github.com/dlvhdr/gh-dash/v4/internal/stack"
)

type pullRequestStackNode struct {
	Stack *struct {
		Number int
		Size   int
	}
	StackEntry *struct {
		Position int
	}
}

// FetchPullRequestStacks returns the GitHub native stack of each PR in urls
// that belongs to one, keyed by PR URL.
func FetchPullRequestStacks(urls []string) (map[string]stack.Native, error) {
	res := map[string]stack.Native{}
	if len(urls) == 0 {
		return res, nil
	}

	c := client
	if c == nil {
		var err error
		if c, err = gh.DefaultGraphQLClient(); err != nil {
			return nil, err
		}
	}

	var params, fields []string
	variables := make(map[string]any, len(urls))
	for i, u := range urls {
		params = append(params, fmt.Sprintf("$u%d: URI!", i))
		fields = append(fields, fmt.Sprintf(
			"p%d: resource(url: $u%d) { ... on PullRequest { stack { number size } stackEntry { position } } }",
			i,
			i,
		))
		variables[fmt.Sprintf("u%d", i)] = u
	}
	query := fmt.Sprintf("query PullRequestStacks(%s) { %s }",
		strings.Join(params, ", "), strings.Join(fields, " "))

	var response map[string]*pullRequestStackNode
	if err := c.Do(query, variables, &response); err != nil {
		return nil, err
	}

	for i, u := range urls {
		node := response[fmt.Sprintf("p%d", i)]
		if node == nil || node.Stack == nil || node.StackEntry == nil {
			continue
		}
		res[u] = stack.Native{
			Number:   node.Stack.Number,
			Size:     node.Stack.Size,
			Position: node.StackEntry.Position,
		}
	}
	return res, nil
}
