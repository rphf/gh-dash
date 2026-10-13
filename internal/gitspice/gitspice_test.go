package gitspice

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	out := []byte(
		`{"name":"a","down":{"name":"main"},"change":{"id":"#1","url":"https://github.com/o/r/pull/1"}}
{"name":"b","down":{"name":"a","needsRestack":true},"change":{"id":"#2","url":"https://github.com/o/r/pull/2"}}
{"name":"c","down":{"name":"b"}}
{"name":"d","down":{"name":"c"},"change":{"id":"#4","url":"https://github.com/o/r/pull/4"}}
{"name":"main","ups":[{"name":"a"}]}
`,
	)
	require.Equal(t, map[string]string{
		"https://github.com/o/r/pull/2": "https://github.com/o/r/pull/1",
	}, parse(out))
}
