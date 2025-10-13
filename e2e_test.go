package mdtopdf_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// containsPDFMarker checks if the given bytes contain the PDF magic marker
func containsPDFMarker(data []byte) bool {
	return bytes.Contains(data, []byte("%PDF"))
}

func TestE2EConversions(t *testing.T) {
	// Build the binary first
	buildCmd := exec.Command("go", "build", "-o", "bin/md2pdf", "./cmd/md2pdf")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}

	binary := "./bin/md2pdf"
	if _, err := os.Stat(binary); err != nil {
		t.Fatalf("Binary not found after build: %v", err)
	}

	testCases := []struct {
		name       string
		inputFile  string
		outputFile string
		extraArgs  []string
		timeout    time.Duration
	}{
		{
			name:       "Basic MD conversion",
			inputFile:  "tests/md2pdf_test.md",
			outputFile: "tests/e2e_basic.pdf",
			timeout:    10 * time.Second,
		},
		{
			name:       "Chinese characters",
			inputFile:  "tests/chinese.md",
			outputFile: "tests/e2e_chinese.pdf",
			timeout:    10 * time.Second,
		},
		{
			name:       "Russian characters",
			inputFile:  "tests/russian_test.md",
			outputFile: "tests/e2e_russian.pdf",
			timeout:    10 * time.Second,
		},
		{
			name:       "Syntax highlighting",
			inputFile:  "tests/test_syntax_highlighting.md",
			outputFile: "tests/e2e_syntax.pdf",
			timeout:    10 * time.Second,
		},
		{
			name:       "Dark theme",
			inputFile:  "tests/md2pdf_test.md",
			outputFile: "tests/e2e_dark.pdf",
			extraArgs:  []string{"--theme", "dark"},
			timeout:    10 * time.Second,
		},
		{
			name:       "Helvetica font",
			inputFile:  "tests/md2pdf_test.md",
			outputFile: "tests/e2e_helvetica.pdf",
			extraArgs:  []string{"--default-font", "Helvetica"},
			timeout:    10 * time.Second,
		},
		{
			name:       "Auto output name",
			inputFile:  "tests/md2pdf_test.md",
			outputFile: "tests/md2pdf_test.pdf", // Auto-generated name
			timeout:    10 * time.Second,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Clean up output file before test
			os.Remove(tc.outputFile)

			// Build command arguments
			args := []string{"-i", tc.inputFile}

			// Only add -o if it's not testing auto-output
			if tc.name != "Auto output name" {
				args = append(args, "-o", tc.outputFile)
			}

			args = append(args, tc.extraArgs...)

			// Create command with timeout
			cmd := exec.Command(binary, args...)

			// Run with timeout
			done := make(chan error, 1)
			go func() {
				done <- cmd.Run()
			}()

			select {
			case err := <-done:
				if err != nil {
					t.Fatalf("Conversion failed: %v", err)
				}
			case <-time.After(tc.timeout):
				cmd.Process.Kill()
				t.Fatalf("Conversion timed out after %v", tc.timeout)
			}

			// Give fpdf time to flush and close the file
			time.Sleep(100 * time.Millisecond)

			// Verify output file exists
			if _, err := os.Stat(tc.outputFile); err != nil {
				t.Fatalf("Output file not created: %v", err)
			}

			// Verify file is not empty
			info, err := os.Stat(tc.outputFile)
			if err != nil {
				t.Fatalf("Cannot stat output file: %v", err)
			}
			if info.Size() == 0 {
				t.Fatalf("Output file is empty")
			}

			// Verify it's a PDF (check for %PDF marker in first 100 bytes)
			// Note: fpdf may write page content before PDF header due to SetHeaderFunc
			f, err := os.Open(tc.outputFile)
			if err != nil {
				t.Fatalf("Cannot open output file: %v", err)
			}
			defer f.Close()

			header := make([]byte, 100)
			n, err := f.Read(header)
			if err != nil {
				t.Fatalf("Cannot read file header: %v", err)
			}

			if !containsPDFMarker(header[:n]) {
				t.Fatalf("Output does not contain PDF marker in first 100 bytes")
			}

			t.Logf("✓ Generated %s (%d bytes)", tc.outputFile, info.Size())
		})
	}
}

func TestE2EDirectoryConversion(t *testing.T) {
	binary := "./bin/md2pdf"

	// Create temp directory with multiple MD files
	tempDir, err := os.MkdirTemp("", "md2pdf-e2e-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test MD files
	testFiles := []string{"doc1.md", "doc2.md", "doc3.md"}
	for _, filename := range testFiles {
		content := "# " + filename + "\n\nTest content for " + filename
		path := filepath.Join(tempDir, filename)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	outputFile := filepath.Join(tempDir, "combined.pdf")

	// Run conversion
	cmd := exec.Command(binary, "-i", tempDir, "-o", outputFile)

	done := make(chan error, 1)
	go func() {
		done <- cmd.Run()
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Directory conversion failed: %v", err)
		}
	case <-time.After(15 * time.Second):
		cmd.Process.Kill()
		t.Fatalf("Directory conversion timed out")
	}

	// Verify output
	if _, err := os.Stat(outputFile); err != nil {
		t.Fatalf("Combined PDF not created: %v", err)
	}

	info, err := os.Stat(outputFile)
	if err != nil {
		t.Fatalf("Cannot stat output file: %v", err)
	}

	t.Logf("✓ Combined directory to PDF (%d bytes)", info.Size())
}

func TestE2EErrorHandling(t *testing.T) {
	binary := "./bin/md2pdf"

	testCases := []struct {
		name      string
		args      []string
		shouldFail bool
	}{
		{
			name:      "Non-existent input file",
			args:      []string{"-i", "non-existent-file.md", "-o", "output.pdf"},
			shouldFail: true,
		},
		{
			name:      "Invalid theme",
			args:      []string{"-i", "tests/md2pdf_test.md", "-o", "output.pdf", "--theme", "invalid"},
			shouldFail: false, // Should use light theme as fallback
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := exec.Command(binary, tc.args...)
			err := cmd.Run()

			if tc.shouldFail && err == nil {
				t.Fatalf("Expected command to fail, but it succeeded")
			}
			if !tc.shouldFail && err != nil {
				t.Fatalf("Expected command to succeed, but it failed: %v", err)
			}

			t.Logf("✓ Error handling validated")
		})
	}
}
