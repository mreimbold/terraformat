package cli

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/spf13/cobra"

	"github.com/mreimbold/terraformat/config"
	"github.com/mreimbold/terraformat/format"
)

const (
	exitOK        = 0
	exitFlagError = 1
	exitError     = 2
	exitCheckDiff = 3
)

const (
	stdinArg  = "-"
	emptyPath = ""
)

const (
	shortFlagPrefix = "-"
	longFlagPrefix  = "--"
)

const (
	startLine   = 1
	startColumn = 1
	startByte   = 0
)

const (
	indexFirst = 0
)

const formattedFilePerm = 0o644

const (
	flagList      = "list"
	flagWrite     = "write"
	flagDiff      = "diff"
	flagCheck     = "check"
	flagNoColor   = "no-color"
	flagRecursive = "recursive"
	flagHelp      = "help"
)

const (
	diffCommand     = "diff"
	diffErrorFormat = "Failed to generate diff for %s: %s"
)

type exitCodeError struct {
	code int
}

// Error returns the error string.
func (err exitCodeError) Error() string {
	return fmt.Sprintf("exit %d", err.code)
}

type staticError string

// Error returns the error string.
func (err staticError) Error() string {
	return string(err)
}

type pathError struct {
	message string
	path    string
}

// Error returns the error string.
func (err pathError) Error() string {
	return fmt.Sprintf(err.message, err.path)
}

type diffError struct {
	path  string
	cause error
}

// Error returns the error string.
func (err diffError) Error() string {
	return fmt.Sprintf(diffErrorFormat, err.path, err.cause)
}

type diagError struct {
	message string
}

// Error returns the error string.
func (err diagError) Error() string {
	return err.message
}

const (
	errWriteWithStdin staticError = "Option -write cannot be used " +
		"when reading from stdin"
	errUnsupportedFile staticError = "Only .tf, .tfvars, and .tftest.hcl " +
		"files can be processed with terraform fmt"
	errWriteResult staticError = "Failed to write result"
)

type fmtOptions struct {
	list      bool
	write     bool
	diff      bool
	check     bool
	noColor   bool
	recursive bool
	targets   []string
}

type ioConfig struct {
	in  io.Reader
	out io.Writer
	err io.Writer
}

type checkPlan struct {
	enabled bool
	list    bool
	buffer  *bytes.Buffer
}

type colorlessWriter struct {
	writer io.Writer
}

// Write forwards bytes without adding color.
func (cw colorlessWriter) Write(p []byte) (int, error) {
	count, err := cw.writer.Write(p)

	return count, wrapExternalError(err)
}

// Execute runs the terraformat CLI and returns the exit code.
func Execute() int {
	cmd := newRootCommand()
	cmd.SetArgs(rewriteSingleDashArgs(os.Args[1:]))

	return executeCommand(cmd)
}

func newRootCommand() *cobra.Command {
	opts := defaultFmtOptions()

	cmd := new(cobra.Command)
	cmd.Use = "terraformat [options] [target...]"
	cmd.Short = "Rewrite Terraform files to a canonical format"
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		opts.targets = args
		ioCfg := ioConfig{
			in:  cmd.InOrStdin(),
			out: cmd.OutOrStdout(),
			err: cmd.ErrOrStderr(),
		}

		code := runFmt(config.Default(), opts, ioCfg)
		if code == exitOK {
			return nil
		}

		return exitCodeError{code: code}
	}
	cmd.SetHelpFunc(func(cmd *cobra.Command, _ []string) {
		_, _ = fmt.Fprint(cmd.OutOrStdout(), fmtHelpText())
	})
	cmd.SetUsageFunc(func(cmd *cobra.Command) error {
		_, _ = fmt.Fprint(cmd.OutOrStdout(), fmtHelpText())

		return nil
	})

	cmd.Flags().BoolVar(&opts.list, flagList, true, flagList)
	cmd.Flags().BoolVar(&opts.write, flagWrite, true, flagWrite)
	cmd.Flags().BoolVar(&opts.diff, flagDiff, false, flagDiff)
	cmd.Flags().BoolVar(&opts.check, flagCheck, false, flagCheck)
	cmd.Flags().BoolVar(&opts.noColor, flagNoColor, false, flagNoColor)
	cmd.Flags().BoolVar(&opts.recursive, flagRecursive, false, flagRecursive)
	cmd.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error {
		_, _ = fmt.Fprintf(
			cmd.ErrOrStderr(),
			"Error parsing command-line flags: %s\n",
			err.Error(),
		)
		_ = cmd.Usage()

		return exitCodeError{code: exitFlagError}
	})

	return cmd
}

func fmtHelpText() string {
	return `Usage: terraformat [options] [target...]

  Rewrites all Terraform configuration files to a canonical format. All
  configuration files (.tf), variables files (.tfvars), and testing files
  (.tftest.hcl) are updated. JSON files (.tf.json, .tfvars.json, or
  .tftest.json) are not modified.

  By default, fmt scans the current directory for configuration files. If you
  provide a directory for the target argument, then fmt will scan that
  directory instead. If you provide a file, then fmt will process just that
  file. If you provide a single dash ("-"), then fmt will read from standard
  input (STDIN).

  The content must be in the Terraform language native syntax; JSON is not
  supported.

Options:

  -list=false    Don't list files whose formatting differs
                 (always disabled if using STDIN)

  -write=false   Don't write to source files
                 (always disabled if using STDIN or -check)

  -diff          Display diffs of formatting changes

  -check         Check if the input is formatted. Exit status will be 0 if all
                 input is properly formatted and non-zero otherwise.

  -no-color      If specified, output won't contain any color.

  -recursive     Also process files in subdirectories. By default, only the
                 given directory (or current directory) is processed.
`
}

func defaultFmtOptions() fmtOptions {
	return fmtOptions{
		list:      true,
		write:     true,
		diff:      false,
		check:     false,
		noColor:   false,
		recursive: false,
		targets:   nil,
	}
}

func executeCommand(cmd *cobra.Command) int {
	err := cmd.Execute()
	if err == nil {
		return exitOK
	}

	var exitErr exitCodeError
	if errors.As(err, &exitErr) {
		return exitErr.code
	}

	_, _ = fmt.Fprintln(cmd.ErrOrStderr(), err)

	return exitError
}

func runFmt(cfg config.Config, opts fmtOptions, ioCfg ioConfig) int {
	stdin, targets := normalizeTargets(opts.targets)
	resolved := applyFmtDefaults(opts)
	resolved.targets = targets

	if stdin {
		resolved = applyStdinDefaults(resolved)
	}

	colorlessCfg := applyNoColor(resolved, ioCfg)

	resolved, plan, runCfg := planCheck(resolved, colorlessCfg)

	var err error
	if stdin {
		err = formatStdin(cfg, resolved, runCfg)
	} else {
		err = formatTargets(cfg, resolved, runCfg)
	}

	if err != nil {
		_, _ = fmt.Fprintln(ioCfg.err, err)

		return exitError
	}

	return finalizeCheck(plan, ioCfg.out)
}

func normalizeTargets(targets []string) (bool, []string) {
	//nolint:revive // add-constant: len check is clear here.
	if len(targets) == 0 {
		return false, []string{"."}
	}

	//nolint:revive // add-constant: index check is clear here.
	if targets[0] == stdinArg {
		return true, nil
	}

	return false, targets
}

func rewriteSingleDashArgs(args []string) []string {
	if len(args) == indexFirst {
		return nil
	}

	rewritten := make([]string, indexFirst, len(args))
	for _, arg := range args {
		rewrite := rewriteSingleDashArg(arg)
		rewritten = append(rewritten, rewrite)
	}

	return rewritten
}

func rewriteSingleDashArg(arg string) string {
	trimmed, ok := trimSingleDashArg(arg)
	if !ok {
		return arg
	}

	return rewriteKnownFlag(trimmed, arg)
}

func trimSingleDashArg(arg string) (string, bool) {
	if arg == stdinArg {
		return emptyPath, false
	}

	if !strings.HasPrefix(arg, shortFlagPrefix) ||
		strings.HasPrefix(arg, longFlagPrefix) {
		return emptyPath, false
	}

	return strings.TrimPrefix(arg, shortFlagPrefix), true
}

func rewriteKnownFlag(trimmed string, original string) string {
	for _, name := range fmtLongFlags() {
		if trimmed == name {
			return longFlagPrefix + trimmed
		}

		if strings.HasPrefix(trimmed, name+"=") {
			return longFlagPrefix + trimmed
		}
	}

	return original
}

func fmtLongFlags() []string {
	return []string{
		flagList,
		flagWrite,
		flagDiff,
		flagCheck,
		flagNoColor,
		flagRecursive,
		flagHelp,
	}
}

func applyFmtDefaults(opts fmtOptions) fmtOptions {
	resolved := opts
	if resolved.check {
		resolved.write = false
	}

	return resolved
}

func applyStdinDefaults(opts fmtOptions) fmtOptions {
	resolved := opts
	resolved.list = false
	resolved.write = false

	return resolved
}

func applyNoColor(opts fmtOptions, ioCfg ioConfig) ioConfig {
	if !opts.noColor {
		return ioCfg
	}

	ioCfg.out = colorlessWriter{writer: ioCfg.out}
	ioCfg.err = colorlessWriter{writer: ioCfg.err}

	return ioCfg
}

func readAll(reader io.Reader, path string) ([]byte, error) {
	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, pathError{message: "Failed to read %s", path: path}
	}

	return content, nil
}

func formatInput(cfg config.Config, src []byte, path string) ([]byte, error) {
	err := validateHCL(src, path)
	if err != nil {
		return nil, err
	}

	out, err := format.Format(src, cfg)
	if err != nil {
		return nil, wrapExternalError(err)
	}

	return out, nil
}

func handleFormattedOutput(
	path string,
	src []byte,
	out []byte,
	opts fmtOptions,
	ioCfg ioConfig,
) error {
	changed := !bytes.Equal(src, out)
	if changed {
		err := handleChangedOutput(path, src, out, opts, ioCfg)
		if err != nil {
			return err
		}
	}

	if shouldPrintFormatted(opts) {
		_, err := ioCfg.out.Write(out)
		if err != nil {
			return errWriteResult
		}
	}

	return nil
}

func handleChangedOutput(
	path string,
	src []byte,
	out []byte,
	opts fmtOptions,
	ioCfg ioConfig,
) error {
	writeListOutput(path, opts, ioCfg)

	err := writeChangedFile(path, out, opts)
	if err != nil {
		return err
	}

	return writeDiffOutput(path, src, out, opts, ioCfg)
}

func writeListOutput(path string, opts fmtOptions, ioCfg ioConfig) {
	if opts.list && path != emptyPath {
		_, _ = fmt.Fprintln(ioCfg.out, path)
	}
}

func writeChangedFile(path string, out []byte, opts fmtOptions) error {
	if !opts.write {
		return nil
	}

	return writeFormattedFile(path, out)
}

func writeDiffOutput(
	path string,
	src []byte,
	out []byte,
	opts fmtOptions,
	ioCfg ioConfig,
) error {
	if !opts.diff {
		return nil
	}

	return writeDiff(path, src, out, ioCfg)
}

func shouldPrintFormatted(opts fmtOptions) bool {
	return !opts.list && !opts.write && !opts.diff
}

func writeFormattedFile(path string, data []byte) error {
	err := os.WriteFile(path, data, formattedFilePerm)
	if err != nil {
		return pathError{message: "Failed to write %s", path: path}
	}

	return nil
}

func wrapExternalError(err error) error {
	if err == nil {
		return nil
	}

	return fmt.Errorf("%w", err)
}

func planCheck(
	opts fmtOptions,
	ioCfg ioConfig,
) (fmtOptions, checkPlan, ioConfig) {
	resolved := opts
	plan := checkPlan{
		enabled: resolved.check,
		list:    resolved.list,
		buffer:  nil,
	}

	if !resolved.check {
		return resolved, plan, ioCfg
	}

	buffer := new(bytes.Buffer)
	plan.buffer = buffer

	resolved.list = true
	resolved.write = false

	ioCfg.out = buffer

	return resolved, plan, ioCfg
}

func finalizeCheck(plan checkPlan, out io.Writer) int {
	if !plan.enabled {
		return exitOK
	}

	buffer := plan.buffer
	if buffer == nil {
		return exitError
	}

	//nolint:revive // add-constant: len check is clear here.
	ok := buffer.Len() == 0
	if plan.list {
		_, _ = io.Copy(out, buffer)
	}

	if ok {
		return exitOK
	}

	return exitCheckDiff
}

func formatStdin(cfg config.Config, opts fmtOptions, ioCfg ioConfig) error {
	if opts.write {
		return errWriteWithStdin
	}

	input, err := readAll(ioCfg.in, emptyPath)
	if err != nil {
		return err
	}

	output, err := formatInput(cfg, input, emptyPath)
	if err != nil {
		return err
	}

	return handleFormattedOutput(emptyPath, input, output, opts, ioCfg)
}

func formatTargets(cfg config.Config, opts fmtOptions, ioCfg ioConfig) error {
	var errs []error

	for _, target := range opts.targets {
		err := processTarget(target, opts, ioCfg, cfg)
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func processTarget(
	target string,
	opts fmtOptions,
	ioCfg ioConfig,
	cfg config.Config,
) error {
	normPath := normalizePath(target)

	info, err := os.Stat(normPath)
	if err != nil {
		return pathError{
			message: "No file or directory at %s",
			path:    normPath,
		}
	}

	if info.IsDir() {
		return processDir(normPath, opts, ioCfg, cfg)
	}

	return processFilePath(normPath, opts, ioCfg, cfg)
}

func processFilePath(
	path string,
	opts fmtOptions,
	ioCfg ioConfig,
	cfg config.Config,
) error {
	if !isTerraformFile(path) {
		return errUnsupportedFile
	}

	//nolint:gosec // CLI intentionally reads user-provided paths.
	inputFile, err := os.Open(path)
	if err != nil {
		return pathError{message: "Failed to read file %s", path: path}
	}

	defer func() {
		_ = inputFile.Close()
	}()

	return processFile(path, inputFile, opts, ioCfg, cfg)
}

func processFile(
	path string,
	reader io.Reader,
	opts fmtOptions,
	ioCfg ioConfig,
	cfg config.Config,
) error {
	src, err := readAll(reader, path)
	if err != nil {
		return err
	}

	out, err := formatInput(cfg, src, path)
	if err != nil {
		return err
	}

	return handleFormattedOutput(path, src, out, opts, ioCfg)
}

func validateHCL(src []byte, path string) error {
	pos := hcl.Pos{Line: startLine, Column: startColumn, Byte: startByte}

	_, diags := hclsyntax.ParseConfig(src, path, pos)
	if diags.HasErrors() {
		return diagError{message: diags.Error()}
	}

	return nil
}

func bytesDiff(before []byte, after []byte, path string) ([]byte, error) {
	beforeFile, err := os.CreateTemp(emptyPath, emptyPath)
	if err != nil {
		return nil, wrapExternalError(err)
	}

	defer func() {
		_ = os.Remove(beforeFile.Name())
		_ = beforeFile.Close()
	}()

	afterFile, err := os.CreateTemp(emptyPath, emptyPath)
	if err != nil {
		return nil, wrapExternalError(err)
	}

	defer func() {
		_ = os.Remove(afterFile.Name())
		_ = afterFile.Close()
	}()

	_, _ = beforeFile.Write(before)
	_, _ = afterFile.Write(after)

	//nolint:gosec // local diff command.
	cmd := exec.CommandContext(
		context.Background(),
		diffCommand,
		"--label=old/"+path,
		"--label=new/"+path,
		"-u",
		beforeFile.Name(),
		afterFile.Name(),
	)
	data, err := cmd.CombinedOutput()
	//nolint:revive // add-constant: zero-length check is clear here.
	if len(data) > 0 {
		// diff exits with a non-zero status when the files don't match.
		// Ignore that failure as long as we get output.
		err = nil
	}

	return data, err
}

func writeDiff(path string, src []byte, out []byte, ioCfg ioConfig) error {
	diff, err := bytesDiff(src, out, path)
	if err != nil {
		return diffError{path: path, cause: err}
	}

	_, _ = ioCfg.out.Write(diff)

	return nil
}

func processDir(
	path string,
	opts fmtOptions,
	ioCfg ioConfig,
	cfg config.Config,
) error {
	entries, err := os.ReadDir(path)
	if err != nil {
		if os.IsNotExist(err) {
			return pathError{
				message: "There is no configuration directory at %s",
				path:    path,
			}
		}

		return pathError{message: "Cannot read directory %s", path: path}
	}

	var errs []error

	for _, entry := range entries {
		err := processDirEntry(path, entry, opts, ioCfg, cfg)
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func processDirEntry(
	root string,
	entry os.DirEntry,
	opts fmtOptions,
	ioCfg ioConfig,
	cfg config.Config,
) error {
	name := entry.Name()
	if shouldSkipFile(name) {
		return nil
	}

	subPath := filepath.Join(root, name)

	if entry.IsDir() {
		if opts.recursive {
			return processDir(subPath, opts, ioCfg, cfg)
		}

		return nil
	}

	if !isTerraformFile(name) {
		return nil
	}

	return processFilePath(subPath, opts, ioCfg, cfg)
}

func normalizePath(path string) string {
	return filepath.Clean(path)
}

func shouldSkipFile(name string) bool {
	return strings.HasPrefix(name, ".")
}

func isTerraformFile(path string) bool {
	name := strings.ToLower(filepath.Base(path))
	if strings.HasSuffix(name, ".tftest.hcl") {
		return true
	}

	if strings.HasSuffix(name, ".tfmock.hcl") {
		return true
	}

	if strings.HasSuffix(name, ".tf") {
		return true
	}

	return strings.HasSuffix(name, ".tfvars")
}
