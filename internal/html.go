package internal

import (
	"os"

	"github.com/ivuorinen/gh-action-readme/appconstants"
)

// HTMLWriter writes HTML output with optional header/footer.
type HTMLWriter struct {
	Header string
	Footer string
}

func (w *HTMLWriter) Write(output, path string) error {
	// Use OpenFile with FilePermDefault (0600) rather than os.Create's 0644 so the
	// HTML output mode matches the markdown/JSON writers and is not world-readable.
	f, err := os.OpenFile( // #nosec G304 -- path from function parameter
		path,
		os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
		appconstants.FilePermDefault,
	)
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close() // Ignore close error in defer
	}()
	if w.Header != "" {
		if _, err := f.WriteString(w.Header); err != nil {
			return err
		}
	}
	if _, err := f.WriteString(output); err != nil {
		return err
	}
	if w.Footer != "" {
		if _, err := f.WriteString(w.Footer); err != nil {
			return err
		}
	}

	return nil
}
