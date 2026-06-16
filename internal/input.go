package internal

import (
	"bufio"
	"errors"
	"io"
	"os"
	"strings"
)

// InputReader reads a single line of user input. It lives in internal/ (not the
// CLI layer) so the I/O abstraction is reusable and testable across packages.
type InputReader interface {
	ReadLine() (string, error)
}

// StdinReader reads lines from the process's standard input. It lazily creates a
// buffered reader on first use and reuses it so the buffer is preserved across
// successive ReadLine calls on the same StdinReader. The guarantee is per
// instance: a caller that reads more than once must reuse a single StdinReader
// rather than constructing a fresh one per read, or a buffered-but-unconsumed
// tail of stdin can be dropped.
type StdinReader struct {
	reader *bufio.Reader
}

// ReadLine returns the next whitespace-trimmed line from stdin.
func (r *StdinReader) ReadLine() (string, error) {
	if r.reader == nil {
		r.reader = bufio.NewReader(os.Stdin)
	}

	line, err := r.reader.ReadString('\n')
	trimmed := strings.TrimSpace(line)

	// EOF on the last line with no trailing newline — the data is still valid
	// input, so return it with a nil error.
	if errors.Is(err, io.EOF) && trimmed != "" {
		return trimmed, nil
	}

	return trimmed, err
}
