package cli

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/mreimbold/terraformat/internal/config"
	"github.com/mreimbold/terraformat/internal/format"
)

func TestRewriteSingleDashArgs(t *testing.T) {
	args := []string{
		"-diff",
		"-list=false",
		"--check",
		"-",
		"-unknown",
		"-no-color",
		"-recursive",
	}
	want := []string{
		"--diff",
		"--list=false",
		"--check",
		"-",
		"-unknown",
		"--no-color",
		"--recursive",
	}

	got := rewriteSingleDashArgs(args)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("rewrite args mismatch:\nwant: %v\n got: %v", want, got)
	}
}

func TestRunFmtDefaultsWriteAndList(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "main.tf")
	input := []byte("resource \"aws_instance\" \"b\" {\nami = \"ami-b\"\n}\n")
	mustWriteFile(t, path, input)

	opts := defaultFmtOptions()
	opts.targets = []string{dir}

	stdout, stderr, code := runFmtForTest(t, opts, bytes.NewBuffer(nil))
	if code != exitOK {
		t.Fatalf("exit code: %d", code)
	}
	if len(stderr) != 0 {
		t.Fatalf("stderr not empty: %s", stderr)
	}
	if !bytes.Contains([]byte(stdout), []byte(path)) {
		t.Fatalf("expected list output to include %s", path)
	}

	want := mustFormat(t, input)
	got := mustReadFile(t, path)
	if !bytes.Equal(got, want) {
		t.Fatalf("formatted file mismatch:\nwant: %s\n got: %s", want, got)
	}
}

func TestRunFmtWriteFalsePrintsFormatted(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "main.tf")
	input := []byte("resource \"aws_instance\" \"b\" {\nami = \"ami-b\"\n}\n")
	mustWriteFile(t, path, input)

	opts := defaultFmtOptions()
	opts.list = false
	opts.write = false
	opts.targets = []string{path}

	stdout, stderr, code := runFmtForTest(t, opts, bytes.NewBuffer(nil))
	if code != exitOK {
		t.Fatalf("exit code: %d", code)
	}
	if len(stderr) != 0 {
		t.Fatalf("stderr not empty: %s", stderr)
	}

	want := mustFormat(t, input)
	if !bytes.Equal([]byte(stdout), want) {
		t.Fatalf("stdout mismatch:\nwant: %s\n got: %s", want, stdout)
	}

	after := mustReadFile(t, path)
	if !bytes.Equal(after, input) {
		t.Fatalf("file should be unchanged")
	}
}

func TestRunFmtDiff(t *testing.T) {
	t.Parallel()

	if _, err := exec.LookPath(diffCommand); err != nil {
		t.Skip("diff not available")
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "main.tf")
	input := []byte("resource \"aws_instance\" \"b\" {\nami = \"ami-b\"\n}\n")
	mustWriteFile(t, path, input)

	opts := defaultFmtOptions()
	opts.list = false
	opts.write = false
	opts.diff = true
	opts.targets = []string{path}

	stdout, stderr, code := runFmtForTest(t, opts, bytes.NewBuffer(nil))
	if code != exitOK {
		t.Fatalf("exit code: %d", code)
	}
	if len(stderr) != 0 {
		t.Fatalf("stderr not empty: %s", stderr)
	}
	if !bytes.Contains([]byte(stdout), []byte("old/")) ||
		!bytes.Contains([]byte(stdout), []byte("new/")) {
		t.Fatalf("expected diff output to include old/new labels")
	}

	after := mustReadFile(t, path)
	if !bytes.Equal(after, input) {
		t.Fatalf("file should be unchanged")
	}
}

func TestRunFmtCheck(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "main.tf")
	input := []byte("resource \"aws_instance\" \"b\" {\nami = \"ami-b\"\n}\n")
	mustWriteFile(t, path, input)

	opts := defaultFmtOptions()
	opts.check = true
	opts.targets = []string{path}

	stdout, stderr, code := runFmtForTest(t, opts, bytes.NewBuffer(nil))
	if code != exitCheckDiff {
		t.Fatalf("exit code: %d", code)
	}
	if len(stderr) != 0 {
		t.Fatalf("stderr not empty: %s", stderr)
	}
	if !bytes.Contains([]byte(stdout), []byte(path)) {
		t.Fatalf("expected list output to include %s", path)
	}

	after := mustReadFile(t, path)
	if !bytes.Equal(after, input) {
		t.Fatalf("file should be unchanged")
	}
}

func TestRunFmtCheckClean(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "main.tf")
	formatted := mustFormat(t, []byte("resource \"aws_instance\" \"b\" {\nami = \"ami-b\"\n}\n"))
	mustWriteFile(t, path, formatted)

	opts := defaultFmtOptions()
	opts.check = true
	opts.targets = []string{path}

	stdout, stderr, code := runFmtForTest(t, opts, bytes.NewBuffer(nil))
	if code != exitOK {
		t.Fatalf("exit code: %d", code)
	}
	if len(stderr) != 0 {
		t.Fatalf("stderr not empty: %s", stderr)
	}
	if len(stdout) != 0 {
		t.Fatalf("stdout should be empty")
	}
}

func TestRunFmtRecursive(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	rootPath := filepath.Join(dir, "main.tf")
	subDir := filepath.Join(dir, "sub")
	subPath := filepath.Join(subDir, "child.tf")
	input := []byte("resource \"aws_instance\" \"b\" {\nami = \"ami-b\"\n}\n")
	mustWriteFile(t, rootPath, input)
	mustMkdirAll(t, subDir)
	mustWriteFile(t, subPath, input)

	opts := defaultFmtOptions()
	opts.targets = []string{dir}
	opts.recursive = false

	_, _, code := runFmtForTest(t, opts, bytes.NewBuffer(nil))
	if code != exitOK {
		t.Fatalf("exit code: %d", code)
	}

	rootFormatted := mustFormat(t, input)
	rootAfter := mustReadFile(t, rootPath)
	if !bytes.Equal(rootAfter, rootFormatted) {
		t.Fatalf("root file should be formatted")
	}

	subAfter := mustReadFile(t, subPath)
	if !bytes.Equal(subAfter, input) {
		t.Fatalf("subdir file should be unchanged")
	}
}

func TestRunFmtRecursiveEnabled(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	rootPath := filepath.Join(dir, "main.tf")
	subDir := filepath.Join(dir, "sub")
	subPath := filepath.Join(subDir, "child.tf")
	input := []byte("resource \"aws_instance\" \"b\" {\nami = \"ami-b\"\n}\n")
	mustWriteFile(t, rootPath, input)
	mustMkdirAll(t, subDir)
	mustWriteFile(t, subPath, input)

	opts := defaultFmtOptions()
	opts.targets = []string{dir}
	opts.recursive = true

	_, _, code := runFmtForTest(t, opts, bytes.NewBuffer(nil))
	if code != exitOK {
		t.Fatalf("exit code: %d", code)
	}

	want := mustFormat(t, input)
	rootAfter := mustReadFile(t, rootPath)
	if !bytes.Equal(rootAfter, want) {
		t.Fatalf("root file should be formatted")
	}
	childAfter := mustReadFile(t, subPath)
	if !bytes.Equal(childAfter, want) {
		t.Fatalf("child file should be formatted")
	}
}

func TestRunFmtStdin(t *testing.T) {
	t.Parallel()

	input := []byte("resource \"aws_instance\" \"b\" {\nami = \"ami-b\"\n}\n")
	opts := defaultFmtOptions()
	opts.targets = []string{stdinArg}

	stdout, stderr, code := runFmtForTest(t, opts, bytes.NewBuffer(input))
	if code != exitOK {
		t.Fatalf("exit code: %d", code)
	}
	if len(stderr) != 0 {
		t.Fatalf("stderr not empty: %s", stderr)
	}

	want := mustFormat(t, input)
	if !bytes.Equal([]byte(stdout), want) {
		t.Fatalf("stdout mismatch:\nwant: %s\n got: %s", want, stdout)
	}
}

func runFmtForTest(t *testing.T, opts fmtOptions, in *bytes.Buffer) (string, string, int) {
	t.Helper()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	ioCfg := ioConfig{in: in, out: &stdout, err: &stderr}

	code := runFmt(config.Default(), opts, ioCfg)

	return stdout.String(), stderr.String(), code
}

func mustWriteFile(t *testing.T, path string, data []byte) {
	t.Helper()

	err := os.WriteFile(path, data, formattedFilePerm)
	if err != nil {
		t.Fatalf("write file: %v", err)
	}
}

func mustReadFile(t *testing.T, path string) []byte {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}

	return data
}

func mustFormat(t *testing.T, src []byte) []byte {
	t.Helper()

	out, err := format.Format(src, config.Default())
	if err != nil {
		t.Fatalf("format: %v", err)
	}

	return out
}

func mustMkdirAll(t *testing.T, path string) {
	t.Helper()

	err := os.MkdirAll(path, 0o755)
	if err != nil {
		t.Fatalf("mkdir: %v", err)
	}
}
