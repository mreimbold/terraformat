//nolint:testpackage // need access to unexported helpers for CLI wiring.
package cli

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/mreimbold/terraformat/internal/config"
	"github.com/mreimbold/terraformat/internal/format"
)

const (
	testInput = "resource \"aws_instance\" \"b\" {\n" +
		"ami = \"ami-b\"\n" +
		"}\n"
	mainTF               = "main.tf"
	emptyString          = ""
	exitCodeFormat       = "exit code: %d"
	stderrNotEmptyFormat = "stderr not empty: %s"
	errFileUnchanged     = "file should be unchanged"
	errRootFormatted     = "root file should be formatted"
	errChildFormatted    = "child file should be formatted"
	errSubdirUnchanged   = "subdir file should be unchanged"
	permDir              = 0o750
)

type fmtResult struct {
	stdout string
	stderr string
	code   int
}

type childExpectation int

const (
	expectChildUnchanged childExpectation = iota
	expectChildFormatted
)

type recursiveCase struct {
	recursive   bool
	expectChild childExpectation
}

// TestRewriteSingleDashArgs ensures single-dash flags are normalized.
func TestRewriteSingleDashArgs(t *testing.T) {
	t.Parallel()

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

// TestRunFmtDefaultsWriteAndList checks defaults write files and list changes.
func TestRunFmtDefaultsWriteAndList(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, mainTF)
	input := []byte(testInput)
	mustWriteFile(t, path, input)

	opts := defaultFmtOptions()
	opts.targets = []string{dir}

	result := runFmtForTest(t, opts, bytes.NewBuffer(nil))
	if result.code != exitOK {
		t.Fatalf(exitCodeFormat, result.code)
	}

	if result.stderr != emptyString {
		t.Fatalf(stderrNotEmptyFormat, result.stderr)
	}

	if !strings.Contains(result.stdout, path) {
		t.Fatalf("expected list output to include %s", path)
	}

	want := mustFormat(t, input)

	got := mustReadFile(t, path)
	if !bytes.Equal(got, want) {
		t.Fatalf("formatted file mismatch:\nwant: %s\n got: %s", want, got)
	}
}

// TestRunFmtWriteFalsePrintsFormatted ensures no write prints formatted output.
func TestRunFmtWriteFalsePrintsFormatted(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, mainTF)
	input := []byte(testInput)
	mustWriteFile(t, path, input)

	opts := defaultFmtOptions()
	opts.list = false
	opts.write = false
	opts.targets = []string{path}

	result := runFmtForTest(t, opts, bytes.NewBuffer(nil))
	if result.code != exitOK {
		t.Fatalf(exitCodeFormat, result.code)
	}

	if result.stderr != emptyString {
		t.Fatalf(stderrNotEmptyFormat, result.stderr)
	}

	want := mustFormat(t, input)
	if !bytes.Equal([]byte(result.stdout), want) {
		t.Fatalf("stdout mismatch:\nwant: %s\n got: %s", want, result.stdout)
	}

	after := mustReadFile(t, path)
	if !bytes.Equal(after, input) {
		t.Fatal(errFileUnchanged)
	}
}

// TestRunFmtDiff verifies diff output is produced when requested.
func TestRunFmtDiff(t *testing.T) {
	t.Parallel()

	_, err := exec.LookPath(diffCommand)
	if err != nil {
		t.Skip("diff not available")
	}

	dir := t.TempDir()
	path := filepath.Join(dir, mainTF)
	input := []byte(testInput)
	mustWriteFile(t, path, input)

	opts := defaultFmtOptions()
	opts.list = false
	opts.write = false
	opts.diff = true
	opts.targets = []string{path}

	result := runFmtForTest(t, opts, bytes.NewBuffer(nil))
	if result.code != exitOK {
		t.Fatalf(exitCodeFormat, result.code)
	}

	if result.stderr != emptyString {
		t.Fatalf(stderrNotEmptyFormat, result.stderr)
	}

	if !strings.Contains(result.stdout, "old/") ||
		!strings.Contains(result.stdout, "new/") {
		t.Fatal("expected diff output to include old/new labels")
	}

	after := mustReadFile(t, path)
	if !bytes.Equal(after, input) {
		t.Fatal(errFileUnchanged)
	}
}

// TestRunFmtCheck ensures -check reports differences and uses exit code 3.
func TestRunFmtCheck(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, mainTF)
	input := []byte(testInput)
	mustWriteFile(t, path, input)

	opts := defaultFmtOptions()
	opts.check = true
	opts.targets = []string{path}

	result := runFmtForTest(t, opts, bytes.NewBuffer(nil))
	if result.code != exitCheckDiff {
		t.Fatalf(exitCodeFormat, result.code)
	}

	if result.stderr != emptyString {
		t.Fatalf(stderrNotEmptyFormat, result.stderr)
	}

	if !strings.Contains(result.stdout, path) {
		t.Fatalf("expected list output to include %s", path)
	}

	after := mustReadFile(t, path)
	if !bytes.Equal(after, input) {
		t.Fatal(errFileUnchanged)
	}
}

// TestRunFmtCheckClean verifies -check exits clean when already formatted.
func TestRunFmtCheckClean(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, mainTF)
	formatted := mustFormat(t, []byte(testInput))
	mustWriteFile(t, path, formatted)

	opts := defaultFmtOptions()
	opts.check = true
	opts.targets = []string{path}

	result := runFmtForTest(t, opts, bytes.NewBuffer(nil))
	if result.code != exitOK {
		t.Fatalf(exitCodeFormat, result.code)
	}

	if result.stderr != emptyString {
		t.Fatalf(stderrNotEmptyFormat, result.stderr)
	}

	if result.stdout != emptyString {
		t.Fatal("stdout should be empty")
	}
}

// TestRunFmtRecursive ensures recursive is opt-in for subdirectories.
func TestRunFmtRecursive(t *testing.T) {
	t.Parallel()

	verifyRecursiveFormatting(t, recursiveCase{
		recursive:   false,
		expectChild: expectChildUnchanged,
	})
}

// TestRunFmtRecursiveEnabled verifies -recursive formats subdirectories too.
func TestRunFmtRecursiveEnabled(t *testing.T) {
	t.Parallel()

	verifyRecursiveFormatting(t, recursiveCase{
		recursive:   true,
		expectChild: expectChildFormatted,
	})
}

// TestRunFmtStdin ensures stdin formatting works without writing files.
func TestRunFmtStdin(t *testing.T) {
	t.Parallel()

	input := []byte(testInput)
	opts := defaultFmtOptions()
	opts.targets = []string{stdinArg}

	result := runFmtForTest(t, opts, bytes.NewBuffer(input))
	if result.code != exitOK {
		t.Fatalf(exitCodeFormat, result.code)
	}

	if result.stderr != emptyString {
		t.Fatalf(stderrNotEmptyFormat, result.stderr)
	}

	want := mustFormat(t, input)
	if !bytes.Equal([]byte(result.stdout), want) {
		t.Fatalf("stdout mismatch:\nwant: %s\n got: %s", want, result.stdout)
	}
}

func verifyRecursiveFormatting(t *testing.T, testCase recursiveCase) {
	t.Helper()

	fixture := newRecursiveFixture(t)

	opts := defaultFmtOptions()
	opts.targets = []string{fixture.dir}
	opts.recursive = testCase.recursive

	result := runFmtForTest(t, opts, bytes.NewBuffer(nil))
	if result.code != exitOK {
		t.Fatalf(exitCodeFormat, result.code)
	}

	want := mustFormat(t, fixture.input)

	rootAfter := mustReadFile(t, fixture.rootPath)
	if !bytes.Equal(rootAfter, want) {
		t.Fatal(errRootFormatted)
	}

	childAfter := mustReadFile(t, fixture.childPath)
	if testCase.expectChild == expectChildFormatted {
		if !bytes.Equal(childAfter, want) {
			t.Fatal(errChildFormatted)
		}

		return
	}

	if !bytes.Equal(childAfter, fixture.input) {
		t.Fatal(errSubdirUnchanged)
	}
}

type recursiveFixture struct {
	dir       string
	rootPath  string
	childPath string
	input     []byte
}

func newRecursiveFixture(t *testing.T) recursiveFixture {
	t.Helper()

	dir := t.TempDir()
	rootPath := filepath.Join(dir, mainTF)
	subDir := filepath.Join(dir, "sub")
	childPath := filepath.Join(subDir, "child.tf")
	input := []byte(testInput)

	mustWriteFile(t, rootPath, input)
	mustMkdirAll(t, subDir)
	mustWriteFile(t, childPath, input)

	return recursiveFixture{
		dir:       dir,
		rootPath:  rootPath,
		childPath: childPath,
		input:     input,
	}
}

func runFmtForTest(
	t *testing.T,
	opts fmtOptions,
	input *bytes.Buffer,
) fmtResult {
	t.Helper()

	var stdout bytes.Buffer

	var stderr bytes.Buffer

	ioCfg := ioConfig{in: input, out: &stdout, err: &stderr}

	code := runFmt(config.Default(), opts, ioCfg)

	return fmtResult{
		stdout: stdout.String(),
		stderr: stderr.String(),
		code:   code,
	}
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

	//nolint:gosec // test helper reads temporary files.
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

	err := os.MkdirAll(path, permDir)
	if err != nil {
		t.Fatalf("mkdir: %v", err)
	}
}
