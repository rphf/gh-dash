// Package attention tells which PRs need the viewer to act.
package attention

import (
	"fmt"
	"strings"
	"time"
)

// Bucket is how much a PR needs the viewer.
type Bucket int

const (
	// Waiting is a PR that waits on someone else.
	Waiting Bucket = iota
	// MyMove is a PR that needs the viewer to act.
	MyMove
	// ReadyToShip is a PR by the viewer that can be merged.
	ReadyToShip
	// Ignored is a draft by someone else.
	Ignored
)

// Rank orders buckets in a list: my move first, then ready to ship.
func (b Bucket) Rank() int {
	switch b {
	case MyMove:
		return 0
	case ReadyToShip:
		return 1
	default:
		return 2
	}
}

// Status is the bucket of a PR and why.
type Status struct {
	Bucket  Bucket
	Reasons []string
}

// Label returns the status as shown in a row, such as "your move: changes
// requested", or an empty string for waiting and ignored PRs.
func (s Status) Label() string {
	switch s.Bucket {
	case MyMove:
		return "your move: " + strings.Join(s.Reasons, ", ")
	case ReadyToShip:
		return "ready to ship"
	default:
		return ""
	}
}

// Event is a comment or a review.
type Event struct {
	Author string
	Bot    bool
	At     time.Time
	// State is the state of a review, such as COMMENTED or APPROVED. It is
	// empty for comments.
	State string
	// HasBody tells if a review has a top-level message.
	HasBody bool
}

// Thread is a review thread.
type Thread struct {
	Resolved bool
	Last     Event
}

// Input holds what a PR's status depends on.
type Input struct {
	Viewer string
	// MyAuthors holds other logins whose PRs, comments and reviews count as
	// the viewer's, such as the viewer's agents.
	MyAuthors      []string
	Author         string
	IsDraft        bool
	ReviewDecision string
	Mergeable      string
	// CIState is the state of the status checks of the last commit, such as
	// SUCCESS or FAILURE.
	CIState string
	// LastCommitAt is the author date of the last commit, which a rebase
	// keeps.
	LastCommitAt       time.Time
	RequestedReviewers []string
	Reviews            []Event
	Comments           []Event
	Threads            []Thread
}

// Compute returns the status of a PR.
func Compute(in Input) Status {
	if in.Viewer == "" {
		return Status{}
	}
	if !in.isMe(in.Author) {
		return reviewerStatus(in)
	}

	var reasons []string
	lastReply := lastReply(in)
	if in.ReviewDecision == "CHANGES_REQUESTED" && lastChangesRequest(in).After(lastReply) {
		reasons = append(reasons, "changes requested")
	}
	if n := newComments(in, lastReply); n > 0 {
		reasons = append(reasons, plural(n, "new comment"))
	}
	if n := openThreads(in); n > 0 {
		reasons = append(reasons, plural(n, "unresolved thread"))
	}
	if in.CIState == "FAILURE" || in.CIState == "ERROR" {
		reasons = append(reasons, "CI failing")
	}
	if in.Mergeable == "CONFLICTING" {
		reasons = append(reasons, "merge conflict")
	}
	if len(reasons) > 0 {
		return Status{Bucket: MyMove, Reasons: reasons}
	}
	if !in.IsDraft && in.ReviewDecision == "APPROVED" && in.CIState == "SUCCESS" &&
		in.Mergeable == "MERGEABLE" {
		return Status{Bucket: ReadyToShip}
	}
	return Status{Bucket: Waiting}
}

func reviewerStatus(in Input) Status {
	if in.IsDraft {
		return Status{Bucket: Ignored}
	}
	requested := false
	for _, r := range in.RequestedReviewers {
		if strings.EqualFold(r, in.Viewer) {
			requested = true
		}
	}
	if !requested {
		return Status{Bucket: Waiting}
	}
	var lastReview time.Time
	for _, r := range in.Reviews {
		if strings.EqualFold(r.Author, in.Viewer) && r.At.After(lastReview) {
			lastReview = r.At
		}
	}
	if lastReview.IsZero() || lastReview.Before(in.LastCommitAt) {
		return Status{Bucket: MyMove, Reasons: []string{"review requested"}}
	}
	return Status{Bucket: Waiting}
}

// lastReply returns the time of the viewer's last comment, review or
// review thread reply.
func lastReply(in Input) time.Time {
	var last time.Time
	for _, e := range append(append([]Event{}, in.Comments...), in.Reviews...) {
		if in.isMe(e.Author) && e.At.After(last) {
			last = e.At
		}
	}
	for _, t := range in.Threads {
		if in.isMe(t.Last.Author) && t.Last.At.After(last) {
			last = t.Last.At
		}
	}
	return last
}

// lastChangesRequest returns the time of the last review that requested
// changes.
func lastChangesRequest(in Input) time.Time {
	var last time.Time
	for _, r := range in.Reviews {
		if r.State == "CHANGES_REQUESTED" && r.At.After(last) {
			last = r.At
		}
	}
	return last
}

// newComments counts the comments by other people after lastReply, and their
// reviews with a top-level message. Review thread comments count through
// openThreads.
func newComments(in Input, lastReply time.Time) int {
	n := 0
	for _, e := range in.Comments {
		if isOther(in, e) && e.At.After(lastReply) {
			n++
		}
	}
	for _, e := range in.Reviews {
		if isOther(in, e) && e.State == "COMMENTED" && e.HasBody && e.At.After(lastReply) {
			n++
		}
	}
	return n
}

// openThreads counts the unresolved review threads whose last comment is by
// another person.
func openThreads(in Input) int {
	n := 0
	for _, t := range in.Threads {
		if !t.Resolved && isOther(in, t.Last) {
			n++
		}
	}
	return n
}

func isOther(in Input, e Event) bool {
	return e.Author != "" && !e.Bot && !in.isMe(e.Author)
}

// isMe tells if login is the viewer or one of MyAuthors. An "app/" prefix,
// as in search filters, is ignored.
func (in Input) isMe(login string) bool {
	if login == "" {
		return false
	}
	for _, l := range append([]string{in.Viewer}, in.MyAuthors...) {
		if strings.EqualFold(strings.TrimPrefix(l, "app/"), login) {
			return true
		}
	}
	return false
}

func plural(n int, word string) string {
	if n == 1 {
		return fmt.Sprintf("1 %s", word)
	}
	return fmt.Sprintf("%d %ss", n, word)
}
