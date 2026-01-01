package format_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mreimbold/terraformat/internal/config"
	tfmt "github.com/mreimbold/terraformat/internal/format"
)

const replaceOnce = 1

// TestFormatGolden verifies formatting against golden files.
func TestFormatGolden(t *testing.T) {
	t.Parallel()

	inputs := mustGlob(t, "testdata/*.input.tf")
	for _, inputPath := range inputs {
		name := strings.TrimSuffix(filepath.Base(inputPath), ".input.tf")
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			runGoldenCase(t, inputPath)
		})
	}
}

// TestIdempotent ensures formatting is idempotent.
func TestIdempotent(t *testing.T) {
	t.Parallel()

	src := mustReadFile(t, "testdata/basic.input.tf")
	first := mustFormat(t, src)
	second := mustFormat(t, first)

	if !bytes.Equal(first, second) {
		t.Fatal("formatting is not idempotent")
	}
}

func runGoldenCase(t *testing.T, inputPath string) {
	t.Helper()

	src := mustReadFile(t, inputPath)
	got := mustFormat(t, src)
	golden := strings.Replace(
		inputPath,
		".input.tf",
		".golden.tf",
		replaceOnce,
	)
	want := mustReadFile(t, golden)

	if !bytes.Equal(got, want) {
		message := "output did not match golden" +
			"\n--- got ---\n%s\n--- want ---\n%s"
		t.Fatalf(message, got, want)
	}
}

func mustFormat(t *testing.T, src []byte) []byte {
	t.Helper()

	formatted, err := tfmt.Format(src, config.Default())
	if err != nil {
		t.Fatalf("format: %v", err)
	}

	return formatted
}

func mustReadFile(t *testing.T, path string) []byte {
	t.Helper()

	//nolint:gosec // Test helper reads local testdata paths.
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}

	return data
}

func mustGlob(t *testing.T, pattern string) []string {
	t.Helper()

	matches, err := filepath.Glob(pattern)
	if err != nil {
		t.Fatalf("glob: %v", err)
	}

	//nolint:revive // add-constant: len check is clear here.
	if len(matches) == 0 {
		t.Fatal("no input files found")
	}

	return matches
}
