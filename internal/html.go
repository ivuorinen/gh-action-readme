package internal

import (
	"github.com/ivuorinen/gh-action-readme/appconstants"
)

// HTMLWriter writes HTML output with optional header/footer.
type HTMLWriter struct {
	Header string
	Footer string
}

func (w *HTMLWriter) Write(output, path string) error {
	// Write through writeFileTightMode (FilePermDefault 0600) so the HTML output
	// matches the markdown/JSON writers and is not world-readable, even when
	// regenerating over an existing file.
	return writeFileTightMode(path, []byte(w.Header+output+w.Footer), appconstants.FilePermDefault)
}
