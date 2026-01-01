package format

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mreimbold/terraformat/config"
)

func TestFormatGolden(t *testing.T) {
	inputs, err := filepath.Glob("testdata/*.input.tf")
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	if len(inputs) == 0 {
		t.Fatal("no input files found")
	}

	for _, input := range inputs {
		name := strings.TrimSuffix(filepath.Base(input), ".input.tf")
		t.Run(name, func(t *testing.T) {
			src, err := os.ReadFile(input)
			if err != nil {
				t.Fatalf("read input: %v", err)
			}
			got, err := Format(src, config.Default())
			if err != nil {
				t.Fatalf("format: %v", err)
			}
			golden := strings.Replace(input, ".input.tf", ".golden.tf", 1)
			want, err := os.ReadFile(golden)
			if err != nil {
				t.Fatalf("read golden: %v", err)
			}
			if !bytes.Equal(got, want) {
				t.Fatalf("output did not match golden\n--- got ---\n%s\n--- want ---\n%s", got, want)
			}
		})
	}
}

func TestIdempotent(t *testing.T) {
	src, err := os.ReadFile("testdata/basic.input.tf")
	if err != nil {
		t.Fatalf("read input: %v", err)
	}
	first, err := Format(src, config.Default())
	if err != nil {
		t.Fatalf("format: %v", err)
	}
	second, err := Format(first, config.Default())
	if err != nil {
		t.Fatalf("format second: %v", err)
	}
	if !bytes.Equal(first, second) {
		t.Fatalf("formatting is not idempotent")
	}
}
