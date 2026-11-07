package cmd

import (
	"errors"
	"io"
	"strings"
	"testing"
)

func TestReadLinePreservesFollowingCredential(t *testing.T) {
	t.Parallel()
	reader := strings.NewReader("first\nsecond\n")
	for _, want := range []string{"first", "second"} {
		got, err := readLine(reader)
		if err != nil || string(got) != want {
			t.Fatalf("readLine = %q, %v; want %q", got, err, want)
		}
	}
}

func TestReadLineEOF(t *testing.T) {
	t.Parallel()
	for _, input := range []string{"", "partial"} {
		got, err := readLine(strings.NewReader(input))
		if !errors.Is(err, io.EOF) || string(got) != input {
			t.Fatalf("readLine = %q, %v; want %q, EOF", got, err, input)
		}
	}
}
