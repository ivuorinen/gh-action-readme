package internal

import (
	"errors"
	"io"
	"os"
	"testing"
)

// withStdin points os.Stdin at a pipe carrying content for the duration of fn.
// Not parallel-safe: os.Stdin is process-global, so tests using it must not call
// t.Parallel().
func withStdin(t *testing.T, content string, fn func()) {
	t.Helper()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}

	orig := os.Stdin
	os.Stdin = r

	defer func() {
		os.Stdin = orig
		_ = r.Close()
	}()

	if _, err := w.WriteString(content); err != nil {
		t.Fatalf("write to pipe: %v", err)
	}
	// Close before reading: ReadLine's EOF handling only triggers once the writer
	// is gone, and an open writer would block the read forever.
	if err := w.Close(); err != nil {
		t.Fatalf("close pipe writer: %v", err)
	}

	fn()
}

// TestStdinReaderReadLineTrims pins the trimming contract: surrounding whitespace
// and the line terminator are stripped, so callers comparing against "y"/"n" do not
// have to defend against "\r\n" or stray spaces.
func TestStdinReaderReadLineTrims(t *testing.T) {
	withStdin(t, "  hello world  \n", func() {
		r := &StdinReader{}

		got, err := r.ReadLine()
		if err != nil {
			t.Fatalf("ReadLine() unexpected error: %v", err)
		}
		if got != "hello world" {
			t.Errorf("ReadLine() = %q, want %q", got, "hello world")
		}
	})
}

// TestStdinReaderReadLineEOFWithoutNewline is the regression guard for the
// documented EOF special case: a final line with no trailing newline is still
// valid input and must come back with a nil error, not io.EOF. Without it, piping
// `printf 'y'` into a prompt would be treated as a read failure.
func TestStdinReaderReadLineEOFWithoutNewline(t *testing.T) {
	withStdin(t, "answer", func() {
		r := &StdinReader{}

		got, err := r.ReadLine()
		if err != nil {
			t.Fatalf("ReadLine() at EOF without newline must return nil error, got: %v", err)
		}
		if got != "answer" {
			t.Errorf("ReadLine() = %q, want %q", got, "answer")
		}
	})
}

// TestStdinReaderReadLineEmptyAtEOF covers the other side of that branch: with
// nothing left to read, the trimmed value is empty and the EOF must surface so the
// caller can stop prompting instead of looping on blank input forever.
func TestStdinReaderReadLineEmptyAtEOF(t *testing.T) {
	withStdin(t, "", func() {
		r := &StdinReader{}

		got, err := r.ReadLine()
		if !errors.Is(err, io.EOF) {
			t.Errorf("ReadLine() on empty stdin error = %v, want io.EOF", err)
		}
		if got != "" {
			t.Errorf("ReadLine() = %q, want empty", got)
		}
	})
}

// TestStdinReaderReusesBuffer pins the per-instance guarantee stated on
// StdinReader: successive reads on one instance keep the buffer, so the second
// line is not lost to the first read's buffering.
func TestStdinReaderReusesBuffer(t *testing.T) {
	withStdin(t, "first\nsecond\n", func() {
		r := &StdinReader{}

		first, err := r.ReadLine()
		if err != nil {
			t.Fatalf("first ReadLine() error: %v", err)
		}

		second, err := r.ReadLine()
		if err != nil {
			t.Fatalf("second ReadLine() error: %v", err)
		}

		if first != "first" || second != "second" {
			t.Errorf("ReadLine() sequence = %q, %q; want \"first\", \"second\"", first, second)
		}
	})
}
