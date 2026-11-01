package main

import (
	"bytes"
	"image"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCaptureBounds(t *testing.T) {
	laptop := image.Rect(0, 0, 1920, 1080)
	wide := image.Rect(-763, -1440, 2677, 0)
	displays := []image.Rectangle{laptop, wide}

	tests := []struct {
		name    string
		index   int
		all     bool
		want    image.Rectangle
		wantErr string
	}{
		{name: "all displays is the union", all: true, want: image.Rect(-763, -1440, 2677, 1080)},
		{name: "first display", index: 0, want: laptop},
		{name: "second display", index: 1, want: wide},
		{name: "index past the last display", index: 2, wantErr: "display index 2 out of range (have 2 display(s), valid range is 0..1)"},
		{name: "negative index", index: -1, wantErr: "display index -1 out of range (have 2 display(s), valid range is 0..1)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := captureBounds(displays, tt.index, tt.all)
			if tt.wantErr != "" {
				var ue *usageError
				require.ErrorAs(t, err, &ue)
				assert.EqualError(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestResolveName(t *testing.T) {
	now := time.Date(2026, 10, 31, 23, 5, 9, 0, time.Local)

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "default is timestamped", input: "", want: "screenshot-20261031-230509.png"},
		{name: "png appended without extension", input: "s4-08", want: "s4-08.png"},
		{name: "extension kept", input: "s4-08.screen.png", want: "s4-08.screen.png"},
		{name: "forward slash rejected", input: "a/b.png", wantErr: true},
		{name: "backslash rejected", input: `a\b.png`, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveName(tt.input, now)
			if tt.wantErr {
				var ue *usageError
				require.ErrorAs(t, err, &ue)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestExpandHome(t *testing.T) {
	home := filepath.Join("C:", "Users", "boss")

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "tilde alone", input: "~", want: home},
		{name: "tilde slash", input: "~/shots", want: filepath.Join(home, "shots")},
		{name: "tilde backslash", input: `~\shots`, want: filepath.Join(home, "shots")},
		{name: "absolute path untouched", input: `D:\shots`, want: `D:\shots`},
		{name: "tilde inside a name untouched", input: "~shots", want: "~shots"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, expandHome(tt.input, home))
		})
	}
}

func TestSavePNGCreatesDirectoryAndReturnsAbsolutePath(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested", "run")

	path, err := savePNG(dir, "shot.png", []byte("png"))

	require.NoError(t, err)
	assert.True(t, filepath.IsAbs(path))
	assert.Equal(t, filepath.Join(dir, "shot.png"), path)
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, []byte("png"), data)
}

func TestOutputDirDefaultsToTemp(t *testing.T) {
	assert.Equal(t, filepath.Join(os.TempDir(), "screenshot"), outputDir(""))
}

func TestRunExitCodes(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantCode   int
		wantStderr string
	}{
		{name: "no command", args: nil, wantCode: 2, wantStderr: "usage: screenshot <command>"},
		{name: "unknown command", args: []string{"snap"}, wantCode: 2, wantStderr: `unknown command "snap"`},
		{name: "help", args: []string{"help"}, wantCode: 0},
		{name: "capture help", args: []string{"capture", "-h"}, wantCode: 0, wantStderr: "-display"},
		{name: "unknown flag", args: []string{"capture", "--monitor", "1"}, wantCode: 2, wantStderr: "flag provided but not defined"},
		{name: "stray argument", args: []string{"displays", "extra"}, wantCode: 2, wantStderr: `unexpected argument "extra"`},
		{name: "name with separator", args: []string{"capture", "--name", "a/b"}, wantCode: 2, wantStderr: "--name must not contain path separators"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer

			code := run(tt.args, &stdout, &stderr)

			assert.Equal(t, tt.wantCode, code)
			assert.Contains(t, stderr.String(), tt.wantStderr)
		})
	}
}
