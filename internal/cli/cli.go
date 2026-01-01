package cli

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/mreimbold/terraformat/config"
	"github.com/mreimbold/terraformat/format"
)

const (
	exitOK    = 0
	exitDiff  = 1
	exitError = 2
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

const (
	errWriteAndCheck       staticError = "-w and -check are mutually exclusive"
	errWriteNeedsArguments staticError = "-w requires file or dir arguments"
	errReadStdin           staticError = "read stdin"
	errFormatInput         staticError = "format input"
	errCollectFiles        staticError = "collect files"
	errReadFile            staticError = "read file"
	errStatFile            staticError = "stat file"
	errWriteFile           staticError = "write file"
	errWalkDir             staticError = "walk dir"
)

type runOptions struct {
	write bool
	check bool
	paths []string
}

type ioConfig struct {
	in  io.Reader
	out io.Writer
	err io.Writer
}

// Execute runs the terraformat CLI and returns the exit code.
func Execute() int {
	cmd := newRootCommand()

	return executeCommand(cmd)
}

func newRootCommand() *cobra.Command {
	var (
		write bool
		check bool
	)

	cmd := new(cobra.Command)
	cmd.Use = "terraformat [flags] [path ...]"
	cmd.Short = "Format Terraform files beyond terraform fmt"
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		opts := runOptions{
			write: write,
			check: check,
			paths: args,
		}
		ioCfg := ioConfig{
			in:  cmd.InOrStdin(),
			out: cmd.OutOrStdout(),
			err: cmd.ErrOrStderr(),
		}

		code := run(config.Default(), opts, ioCfg)
		if code == exitOK {
			return nil
		}

		return exitCodeError{code: code}
	}

	cmd.Flags().BoolVarP(
		&write,
		"write",
		"w",
		false,
		"write result to (source) file instead of stdout",
	)
	cmd.Flags().BoolVar(
		&check,
		"check",
		false,
		"exit with non-zero status if formatting would change any files",
	)
	cmd.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error {
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), err)
		_ = cmd.Usage()

		return exitCodeError{code: exitError}
	})

	return cmd
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

func run(cfg config.Config, opts runOptions, ioCfg ioConfig) int {
	err := validateOptions(opts)
	if err != nil {
		_, _ = fmt.Fprintln(ioCfg.err, err)

		return exitError
	}

	//nolint:revive // add-constant: len check is clear here.
	if len(opts.paths) == 0 {
		return runOnStdin(cfg, opts, ioCfg)
	}

	return runOnPaths(cfg, opts, ioCfg)
}

func validateOptions(opts runOptions) error {
	if opts.write && opts.check {
		return errWriteAndCheck
	}

	//nolint:revive // add-constant: len check is clear here.
	if len(opts.paths) == 0 && opts.write {
		return errWriteNeedsArguments
	}

	return nil
}

func runOnStdin(cfg config.Config, opts runOptions, ioCfg ioConfig) int {
	input, err := io.ReadAll(ioCfg.in)
	if err != nil {
		_, _ = fmt.Fprintln(ioCfg.err, wrapError(errReadStdin, err))

		return exitError
	}

	output, err := format.Format(input, cfg)
	if err != nil {
		_, _ = fmt.Fprintln(ioCfg.err, wrapError(errFormatInput, err))

		return exitError
	}

	if opts.check {
		if !bytes.Equal(input, output) {
			return exitDiff
		}

		return exitOK
	}

	_, _ = ioCfg.out.Write(output)

	return exitOK
}

func runOnPaths(cfg config.Config, opts runOptions, ioCfg ioConfig) int {
	files, err := collectFiles(opts.paths)
	if err != nil {
		_, _ = fmt.Fprintln(ioCfg.err, wrapError(errCollectFiles, err))

		return exitError
	}

	changed, err := processFiles(files, cfg, opts, ioCfg)
	if err != nil {
		_, _ = fmt.Fprintln(ioCfg.err, err)

		return exitError
	}

	if opts.check && changed {
		return exitDiff
	}

	return exitOK
}

func processFiles(
	files []string,
	cfg config.Config,
	opts runOptions,
	ioCfg ioConfig,
) (bool, error) {
	changed := false

	for _, path := range files {
		fileChanged, err := processFile(path, cfg, opts, ioCfg)
		if err != nil {
			return false, err
		}

		if fileChanged {
			changed = true
		}
	}

	return changed, nil
}

func processFile(
	path string,
	cfg config.Config,
	opts runOptions,
	ioCfg ioConfig,
) (bool, error) {
	src, info, err := readFile(path)
	if err != nil {
		return false, err
	}

	out, err := format.Format(src, cfg)
	if err != nil {
		return false, wrapPathError(errFormatInput, path, err)
	}

	if bytes.Equal(src, out) {
		return handleUnchangedOutput(out, opts, ioCfg)
	}

	return handleChangedOutput(path, info, out, opts, ioCfg)
}

//nolint:gosec // CLI intentionally reads user-provided paths.
func readFile(path string) ([]byte, os.FileInfo, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, wrapPathError(errReadFile, path, err)
	}

	info, err := os.Stat(path)
	if err != nil {
		return nil, nil, wrapPathError(errStatFile, path, err)
	}

	return src, info, nil
}

func handleUnchangedOutput(
	out []byte,
	opts runOptions,
	ioCfg ioConfig,
) (bool, error) {
	if !opts.check && !opts.write {
		_, _ = ioCfg.out.Write(out)
	}

	return false, nil
}

func handleChangedOutput(
	path string,
	info os.FileInfo,
	out []byte,
	opts runOptions,
	ioCfg ioConfig,
) (bool, error) {
	if opts.check {
		return true, nil
	}

	if opts.write {
		err := os.WriteFile(path, out, info.Mode().Perm())
		if err != nil {
			return false, wrapPathError(errWriteFile, path, err)
		}

		return true, nil
	}

	_, _ = ioCfg.out.Write(out)

	return true, nil
}

func collectFiles(paths []string) ([]string, error) {
	var files []string

	for _, path := range paths {
		pathFiles, err := collectPath(path)
		if err != nil {
			return nil, err
		}

		files = append(files, pathFiles...)
	}

	sort.Strings(files)

	return files, nil
}

func collectPath(path string) ([]string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, wrapPathError(errStatFile, path, err)
	}

	if !info.IsDir() {
		if isTerraformFile(path) {
			return []string{path}, nil
		}

		return nil, nil
	}

	return collectDir(path)
}

func collectDir(root string) ([]string, error) {
	var files []string

	err := walkTerraformDir(root, &files)
	if err != nil {
		return nil, err
	}

	return files, nil
}

func walkTerraformDir(root string, files *[]string) error {
	err := filepath.WalkDir(
		root,
		func(path string, entry os.DirEntry, walkErr error) error {
			return handleWalkEntry(path, entry, walkErr, files)
		},
	)
	if err != nil && !errors.Is(err, filepath.SkipDir) {
		return wrapError(errWalkDir, err)
	}

	return nil
}

func handleWalkEntry(
	path string,
	entry os.DirEntry,
	walkErr error,
	files *[]string,
) error {
	if walkErr != nil {
		return wrapPathError(errWalkDir, path, walkErr)
	}

	if entry.IsDir() {
		if shouldSkipDir(filepath.Base(path)) {
			return filepath.SkipDir
		}

		return nil
	}

	if isTerraformFile(path) {
		*files = append(*files, path)
	}

	return nil
}

func shouldSkipDir(name string) bool {
	return name == ".terraform" || strings.HasPrefix(name, ".")
}

func isTerraformFile(path string) bool {
	name := strings.ToLower(path)

	return strings.HasSuffix(name, ".tf") || strings.HasSuffix(name, ".tfvars")
}

func wrapError(errType error, err error) error {
	return fmt.Errorf(errWrapFormat, errType, err)
}

func wrapPathError(errType error, path string, err error) error {
	return fmt.Errorf(errWrapPathFormat, errType, path, err)
}
