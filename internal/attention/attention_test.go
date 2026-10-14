package attention

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var t0 = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

func at(h int) time.Time {
	return t0.Add(time.Duration(h) * time.Hour)
}

func mine() Input {
	return Input{Viewer: "me", Author: "me", CIState: "PENDING", Mergeable: "MERGEABLE"}
}

func TestAuthorMyMove(t *testing.T) {
	in := mine()
	require.Equal(t, Status{Bucket: Waiting}, Compute(in))

	in.ReviewDecision = "CHANGES_REQUESTED"
	in.Reviews = []Event{{Author: "bob", State: "CHANGES_REQUESTED", At: at(1)}}
	in.CIState = "FAILURE"
	in.Mergeable = "CONFLICTING"
	require.Equal(t, "your move: changes requested, CI failing, merge conflict",
		Compute(in).Label())
}

func TestAuthorNewComments(t *testing.T) {
	in := mine()
	in.Comments = []Event{
		{Author: "bob", At: at(1)},
		{Author: "me", At: at(2)},
		{Author: "bob", At: at(3)},
		{Author: "ci", Bot: true, At: at(4)},
	}
	in.Reviews = []Event{
		{Author: "ann", State: "COMMENTED", HasBody: true, At: at(5)},
		{Author: "ann", State: "COMMENTED", At: at(5)},
		{Author: "ann", State: "APPROVED", At: at(6)},
	}
	require.Equal(t, "your move: 2 new comments", Compute(in).Label())

	in.Comments = append(in.Comments, Event{Author: "me", At: at(7)})
	require.Equal(t, Waiting, Compute(in).Bucket)
}

func TestChangesRequestedAfterMyReply(t *testing.T) {
	in := mine()
	in.ReviewDecision = "CHANGES_REQUESTED"
	in.Reviews = []Event{
		{Author: "bob", State: "CHANGES_REQUESTED", At: at(1)},
		{Author: "me", State: "COMMENTED", At: at(2)},
	}
	require.Equal(t, Waiting, Compute(in).Bucket, "I replied after the changes request")

	in.Reviews = append(in.Reviews, Event{Author: "bob", State: "CHANGES_REQUESTED", At: at(3)})
	require.Equal(t, "your move: changes requested", Compute(in).Label())
}

func TestAuthorUnresolvedThreads(t *testing.T) {
	in := mine()
	in.Threads = []Thread{
		{Last: Event{Author: "bob", At: at(1)}},
		{Resolved: true, Last: Event{Author: "bob", At: at(1)}},
		{Last: Event{Author: "me", At: at(2)}},
	}
	require.Equal(t, "your move: 1 unresolved thread", Compute(in).Label())
}

func TestReadyToShip(t *testing.T) {
	in := mine()
	in.ReviewDecision = "APPROVED"
	in.CIState = "SUCCESS"
	require.Equal(t, "ready to ship", Compute(in).Label())

	in.IsDraft = true
	require.Equal(t, Waiting, Compute(in).Bucket)
}

func TestReviewer(t *testing.T) {
	in := Input{Viewer: "me", Author: "bob", LastCommitAt: at(2)}
	require.Equal(t, Waiting, Compute(in).Bucket)

	in.RequestedReviewers = []string{"me"}
	require.Equal(t, "your move: review requested", Compute(in).Label())

	in.Reviews = []Event{{Author: "me", At: at(1)}}
	require.Equal(t, MyMove, Compute(in).Bucket, "new push since my review")

	in.Reviews = []Event{{Author: "me", At: at(3)}}
	require.Equal(t, Waiting, Compute(in).Bucket)

	in.IsDraft = true
	require.Equal(t, Ignored, Compute(in).Bucket)
}

func TestMyAuthors(t *testing.T) {
	in := Input{
		Viewer:         "me",
		MyAuthors:      []string{"app/my-agent"},
		Author:         "my-agent",
		ReviewDecision: "CHANGES_REQUESTED",
		Reviews:        []Event{{Author: "bob", State: "CHANGES_REQUESTED", At: at(1)}},
	}
	require.Equal(t, "your move: changes requested", Compute(in).Label())

	in.ReviewDecision = ""
	in.Reviews = nil
	in.Comments = []Event{
		{Author: "bob", At: at(1)},
		{Author: "my-agent", Bot: true, At: at(2)},
	}
	require.Equal(t, Waiting, Compute(in).Bucket, "the agent's reply counts as mine")
}

func TestNoViewer(t *testing.T) {
	require.Equal(t, Status{}, Compute(Input{Author: "me"}))
}
