package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"image"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kbinani/screenshot"
)

func runDisplays(args []string, stdout, stderr io.Writer) error {
	fs := newFlagSet("displays", "", stderr)
	if err := parseFlags(fs, args); err != nil {
		return err
	}

	displays, err := activeDisplays()
	if err != nil {
		return err
	}
	for i, b := range displays {
		fmt.Fprintf(stdout, "[%d] %dx%d at (%d,%d)\n", i, b.Dx(), b.Dy(), b.Min.X, b.Min.Y)
	}
	return nil
}

func runCapture(args []string, stdout, stderr io.Writer) error {
	fs := newFlagSet("capture", "[--display N] [--dir DIR] [--name NAME]", stderr)
	display := fs.Int("display", 0, "0-based display index (default: all displays as one image)")
	dir := fs.String("dir", "", `directory to save into, created if missing; a leading ~ expands to the home directory (default: screenshot in the OS temp directory)`)
	name := fs.String("name", "", "file name without path separators; .png is appended when it has no extension (default: screenshot-YYYYMMDD-HHMMSS.png)")
	if err := parseFlags(fs, args); err != nil {
		return err
	}

	allDisplays := true
	fs.Visit(func(f *flag.Flag) {
		if f.Name == "display" {
			allDisplays = false
		}
	})

	fileName, err := resolveName(*name, time.Now())
	if err != nil {
		return err
	}
	displays, err := activeDisplays()
	if err != nil {
		return err
	}
	bounds, err := captureBounds(displays, *display, allDisplays)
	if err != nil {
		return err
	}

	img, err := screenshot.CaptureRect(bounds)
	if err != nil {
		return fmt.Errorf("capture failed: %w", err)
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return fmt.Errorf("png encode failed: %w", err)
	}

	path, err := savePNG(outputDir(*dir), fileName, buf.Bytes())
	if err != nil {
		return err
	}
	fmt.Fprintln(stdout, path)
	return nil
}

func activeDisplays() ([]image.Rectangle, error) {
	n := screenshot.NumActiveDisplays()
	if n == 0 {
		return nil, errors.New("no active displays found")
	}
	displays := make([]image.Rectangle, n)
	for i := range displays {
		displays[i] = screenshot.GetDisplayBounds(i)
	}
	return displays, nil
}

func captureBounds(displays []image.Rectangle, index int, all bool) (image.Rectangle, error) {
	if all {
		bounds := displays[0]
		for _, d := range displays[1:] {
			bounds = bounds.Union(d)
		}
		return bounds, nil
	}
	if index < 0 || index >= len(displays) {
		return image.Rectangle{}, usageErrorf("display index %d out of range (have %d display(s), valid range is 0..%d)", index, len(displays), len(displays)-1)
	}
	return displays[index], nil
}

func resolveName(name string, now time.Time) (string, error) {
	if name == "" {
		name = "screenshot-" + now.Format("20060102-150405")
	}
	if strings.ContainsAny(name, `/\`) {
		return "", usageErrorf("--name must not contain path separators; pass the directory with --dir")
	}
	if filepath.Ext(name) == "" {
		name += ".png"
	}
	return name, nil
}

func outputDir(dir string) string {
	if dir == "" {
		return filepath.Join(os.TempDir(), "screenshot")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return dir
	}
	return expandHome(dir, home)
}

func expandHome(p, home string) string {
	switch {
	case p == "~":
		return home
	case strings.HasPrefix(p, "~/"), strings.HasPrefix(p, `~\`):
		return filepath.Join(home, p[2:])
	default:
		return p
	}
}

func savePNG(dir, name string, data []byte) (string, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("failed to create directory %q: %w", dir, err)
	}
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return "", fmt.Errorf("failed to write %q: %w", path, err)
	}
	return filepath.Abs(path)
}
