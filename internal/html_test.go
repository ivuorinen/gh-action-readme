package internal

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/ivuorinen/gh-action-readme/testutil"
)

// mustSafePath validates that a path is safe (no "..", matches cleaned version).
// Fails the test if path is unsafe.
func mustSafePath(t *testing.T, p string) string {
	t.Helper()
	cleaned := filepath.Clean(p)
	if cleaned != p {
		t.Fatalf("path %q does not match cleaned path %q", p, cleaned)
	}
	if strings.Contains(cleaned, "..") {
		t.Fatalf("path %q contains unsafe .. component", p)
	}

	return cleaned
}

// TestHTMLWriterWrite tests the HTMLWriter.Write function.
func TestHTMLWriterWrite(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		header     string
		footer     string
		content    string
		wantString string
	}{
		{
			name:       "no header or footer",
			header:     "",
			footer:     "",
			content:    "<h1>Test Content</h1>",
			wantString: "<h1>Test Content</h1>",
		},
		{
			name:       "with header only",
			header:     "<!DOCTYPE html>\n<html>\n",
			footer:     "",
			content:    "<body>Content</body>",
			wantString: "<!DOCTYPE html>\n<html>\n<body>Content</body>",
		},
		{
			name:       "with footer only",
			header:     "",
			footer:     testutil.TestHTMLClosingTag,
			content:    "<body>Content</body>",
			wantString: "<body>Content</body>\n</html>",
		},
		{
			name:       "with both header and footer",
			header:     "<!DOCTYPE html>\n<html>\n<body>\n",
			footer:     "\n</body>\n</html>",
			content:    "<h1>Main Content</h1>",
			wantString: "<!DOCTYPE html>\n<html>\n<body>\n<h1>Main Content</h1>\n</body>\n</html>",
		},
		{
			name:       "empty content",
			header:     "<header>",
			footer:     "</footer>",
			content:    "",
			wantString: "<header></footer>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			outputPath := filepath.Join(tmpDir, "test.html")

			writer := &HTMLWriter{
				Header: tt.header,
				Footer: tt.footer,
			}

			err := writer.Write(tt.content, outputPath)
			if err != nil {
				t.Errorf("Write() unexpected error = %v", err)

				return
			}

			// Read the file and verify content
			content, err := os.ReadFile(mustSafePath(t, outputPath))
			if err != nil {
				t.Fatalf(testutil.TestMsgFailedToReadOutput, err)
			}

			got := string(content)
			if got != tt.wantString {
				t.Errorf("Write() content = %q, want %q", got, tt.wantString)
			}
		})
	}
}

// TestHTMLWriterWriteErrorPaths tests error handling in HTMLWriter.Write.
func TestHTMLWriterWriteErrorPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		setupPath  func(t *testing.T) string
		skipReason string
		wantErr    bool
	}{
		{
			name: "invalid path - directory doesn't exist",
			setupPath: func(t *testing.T) string {
				t.Helper()
				tmpDir := t.TempDir()

				return filepath.Join(tmpDir, "nonexistent", "file.html")
			},
			wantErr: true,
		},
		{
			name: "permission denied - unwritable directory",
			setupPath: func(t *testing.T) string {
				t.Helper()
				// Skip on Windows (chmod behavior differs)
				if runtime.GOOS == "windows" {
					return ""
				}
				// Skip if running as root (can write anywhere)
				if os.Geteuid() == 0 {
					return ""
				}

				tmpDir := t.TempDir()
				restrictedDir := filepath.Join(tmpDir, "restricted")
				if err := os.Mkdir(restrictedDir, 0700); err != nil {
					t.Fatalf("failed to create restricted dir: %v", err)
				}

				// Make directory unwritable
				if err := os.Chmod(restrictedDir, 0000); err != nil {
					t.Fatalf("failed to chmod: %v", err)
				}

				// Restore permissions in cleanup
				t.Cleanup(func() {
					_ = os.Chmod(restrictedDir, 0700) // #nosec G302 -- directory needs exec bit for cleanup
				})

				return filepath.Join(restrictedDir, "file.html")
			},
			skipReason: "skipped on Windows or when running as root",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			path := tt.setupPath(t)
			if path == "" {
				t.Skip(tt.skipReason)
			}

			writer := &HTMLWriter{
				Header: "<header>",
				Footer: "</footer>",
			}

			err := writer.Write("<content>", path)
			if (err != nil) != tt.wantErr {
				t.Errorf("Write() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestHTMLWriterWriteLargeContent tests writing large HTML content.
func TestHTMLWriterWriteLargeContent(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "large.html")

	// Create large content (10MB)
	largeContent := strings.Repeat("<p>Test content line</p>\n", 500000)

	writer := &HTMLWriter{
		Header: "<!DOCTYPE html>\n",
		Footer: testutil.TestHTMLClosingTag,
	}

	err := writer.Write(largeContent, outputPath)
	if err != nil {
		t.Errorf("Write() failed for large content: %v", err)
	}

	// Verify file was created and has correct size
	info, err := os.Stat(outputPath)
	if err != nil {
		t.Fatalf("Failed to stat output file: %v", err)
	}

	expectedSize := len("<!DOCTYPE html>\n") + len(largeContent) + len(testutil.TestHTMLClosingTag)
	if int(info.Size()) != expectedSize {
		t.Errorf("File size = %d, want %d", info.Size(), expectedSize)
	}
}

// TestHTMLWriterWriteSpecialCharacters tests writing HTML with special characters.
func TestHTMLWriterWriteSpecialCharacters(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "special.html")

	// Content with HTML entities and special characters
	content := `<div>&lt;script&gt;alert("test")&lt;/script&gt;</div>
<p>Special chars: &amp; &quot; &apos; &lt; &gt;</p>
<p>Unicode: 你好 مرحبا привет 🎉</p>`

	writer := &HTMLWriter{}
	err := writer.Write(content, outputPath)
	if err != nil {
		t.Errorf("Write() failed for special characters: %v", err)
	}

	// Verify content was written correctly
	readContent, err := os.ReadFile(mustSafePath(t, outputPath))
	if err != nil {
		t.Fatalf(testutil.TestMsgFailedToReadOutput, err)
	}

	if string(readContent) != content {
		t.Errorf("Content mismatch for special characters")
	}
}

// TestHTMLWriterWriteOverwrite tests overwriting an existing file.
func TestHTMLWriterWriteOverwrite(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "overwrite.html")

	// Write initial content
	writer := &HTMLWriter{}
	err := writer.Write("Initial content", outputPath)
	if err != nil {
		t.Fatalf("Initial write failed: %v", err)
	}

	// Overwrite with new content
	err = writer.Write(testutil.TestHTMLNewContent, outputPath)
	if err != nil {
		t.Errorf("Overwrite failed: %v", err)
	}

	// Verify new content
	content, err := os.ReadFile(mustSafePath(t, outputPath))
	if err != nil {
		t.Fatalf(testutil.TestMsgFailedToReadOutput, err)
	}

	if string(content) != testutil.TestHTMLNewContent {
		t.Errorf("Content = %q, want %q", string(content), testutil.TestHTMLNewContent)
	}
}

// TestHTMLWriterWriteEmptyPath tests writing to an empty path.
func TestHTMLWriterWriteEmptyPath(t *testing.T) {
	t.Parallel()

	writer := &HTMLWriter{}
	err := writer.Write("content", "")

	// Empty path should cause an error
	if err == nil {
		t.Error("Write() with empty path should return error")
	}
}

// TestHTMLWriterWriteValidPath tests writing to a valid nested path.
func TestHTMLWriterWriteValidPath(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Create nested directory structure
	nestedDir := filepath.Join(tmpDir, "nested", "directory")
	err := os.MkdirAll(nestedDir, 0750) // Use secure permissions
	if err != nil {
		t.Fatalf("Failed to create nested directory: %v", err)
	}

	outputPath := filepath.Join(nestedDir, "nested.html")

	writer := &HTMLWriter{
		Header: "<html>",
		Footer: "</html>",
	}

	err = writer.Write("<body>Nested content</body>", outputPath)
	if err != nil {
		t.Errorf("Write() to nested path failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("File was not created in nested path")
	}
}
