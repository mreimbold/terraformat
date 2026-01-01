package cli

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mreimbold/terraformat/config"
	"github.com/mreimbold/terraformat/format"
)

const (
	exitOK        = 0
	exitDiff      = 1
	exitError     = 2
)

func Run(binName string) int {
	flags := flag.NewFlagSet(binName, flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	write := flags.Bool("w", false, "write result to (source) file instead of stdout")
	check := flags.Bool("check", false, "exit with non-zero status if formatting would change any files")
	if err := flags.Parse(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		printUsage(binName)
		return exitError
	}
	if *write && *check {
		fmt.Fprintln(os.Stderr, "-w and -check are mutually exclusive")
		return exitError
	}

	cfg := config.Default()

	paths := flags.Args()
	if len(paths) == 0 {
		if *write {
			fmt.Fprintln(os.Stderr, "-w requires file or directory arguments")
			return exitError
		}
		input, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return exitError
		}
		output, err := format.Format(input, cfg)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return exitError
		}
		_, _ = os.Stdout.Write(output)
		return exitOK
	}

	files, err := collectFiles(paths)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return exitError
	}

	changed := false
	for _, path := range files {
		src, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "read %s: %v\n", path, err)
			return exitError
		}
		info, err := os.Stat(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "stat %s: %v\n", path, err)
			return exitError
		}
		out, err := format.Format(src, cfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "format %s: %v\n", path, err)
			return exitError
		}
		if !bytes.Equal(src, out) {
			changed = true
			if *check {
				continue
			}
			if *write {
				if err := os.WriteFile(path, out, info.Mode().Perm()); err != nil {
					fmt.Fprintf(os.Stderr, "write %s: %v\n", path, err)
					return exitError
				}
				continue
			}
		}
		if !*check && !*write {
			_, _ = os.Stdout.Write(out)
		}
	}

	if *check && changed {
		return exitDiff
	}
	return exitOK
}

func printUsage(binName string) {
	msg := fmt.Sprintf("usage: %s [flags] [path ...]\n\nflags:\n  -w\twrite result to (source) file instead of stdout\n  -check\texit with non-zero status if formatting would change any files\n", binName)
	_, _ = os.Stderr.WriteString(msg)
}

func collectFiles(paths []string) ([]string, error) {
	var files []string
	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			return nil, err
		}
		if !info.IsDir() {
			if isTerraformFile(p) {
				files = append(files, p)
			}
			continue
		}

		err = filepath.WalkDir(p, func(path string, d os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if d.IsDir() {
				base := filepath.Base(path)
				if strings.HasPrefix(base, ".") || base == ".terraform" {
					return filepath.SkipDir
				}
				return nil
			}
			if isTerraformFile(path) {
				files = append(files, path)
			}
			return nil
		})
		if err != nil && !errors.Is(err, filepath.SkipDir) {
			return nil, err
		}
	}
	sort.Strings(files)
	return files, nil
}

func isTerraformFile(path string) bool {
	name := strings.ToLower(path)
	return strings.HasSuffix(name, ".tf") || strings.HasSuffix(name, ".tfvars")
}
