package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
)

const usage = `usage: screenshot <command> [flags]

commands:
  displays   list active displays with index, size and position
  capture    capture the desktop to a PNG and print its absolute path

Run "screenshot <command> -h" for the command's flags.
`

var errInvalidArgs = errors.New("invalid arguments")

type usageError struct{ msg string }

func (e *usageError) Error() string { return e.msg }

func usageErrorf(format string, a ...any) error {
	return &usageError{msg: fmt.Sprintf(format, a...)}
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, usage)
		return 2
	}

	var err error
	switch args[0] {
	case "displays":
		err = runDisplays(args[1:], stdout, stderr)
	case "capture":
		err = runCapture(args[1:], stdout, stderr)
	case "help", "-h", "-help", "--help":
		fmt.Fprint(stdout, usage)
		return 0
	default:
		fmt.Fprintf(stderr, "screenshot: unknown command %q\n\n%s", args[0], usage)
		return 2
	}

	var ue *usageError
	switch {
	case err == nil, errors.Is(err, flag.ErrHelp):
		return 0
	case errors.Is(err, errInvalidArgs):
		return 2
	case errors.As(err, &ue):
		fmt.Fprintf(stderr, "screenshot %s: %v\n", args[0], err)
		return 2
	default:
		fmt.Fprintf(stderr, "screenshot %s: %v\n", args[0], err)
		return 1
	}
}

func newFlagSet(name, synopsis string, stderr io.Writer) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.Usage = func() {
		fmt.Fprintf(stderr, "usage: screenshot %s %s\n", name, synopsis)
		fs.PrintDefaults()
	}
	return fs
}

func parseFlags(fs *flag.FlagSet, args []string) error {
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return err
		}
		return errInvalidArgs
	}
	if fs.NArg() > 0 {
		fmt.Fprintf(fs.Output(), "unexpected argument %q\n", fs.Arg(0))
		fs.Usage()
		return errInvalidArgs
	}
	return nil
}
