package internal

import (
	"os"
)

// HTMLWriter writes HTML output with optional header/footer.
type HTMLWriter struct {
	Header string
	Footer string
}

func (w *HTMLWriter) Write(output string, path string) error {
	f, err := os.Create(path) // #nosec G304 -- path from function parameter
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
