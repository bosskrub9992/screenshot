# screenshot

Capture the desktop to a PNG from the command line. A single binary with no
server and no background process, so scripts and coding agents can call it
straight from a shell.

## Install

```bash
go install github.com/bosskrub9992/screenshot@latest
```

The binary lands in `$(go env GOPATH)/bin` (or `$GOBIN`); put that directory on
`PATH`.

## Usage

```console
$ screenshot displays
[0] 1920x1080 at (0,0)
[1] 3440x1440 at (-763,-1440)

$ screenshot capture --display 1 --dir ~/shots --name login
C:\Users\me\shots\login.png
```

| Command | What it does |
|---|---|
| `screenshot displays` | One line per active display: 0-based index, size and position in desktop coordinates |
| `screenshot capture` | Capture to a PNG and print its absolute path on stdout |

`capture` flags:

| Flag | Default | Meaning |
|---|---|---|
| `--display N` | all displays as one image | 0-based display index from `displays` |
| `--dir DIR` | `screenshot` in the OS temp directory | Created if missing; a leading `~` expands to the home directory |
| `--name NAME` | `screenshot-YYYYMMDD-HHMMSS.png` | File name only, no path separators; `.png` is appended when it has no extension |

`-h` works on every command.

## Exit codes

| Code | Meaning |
|---|---|
| `0` | Success |
| `1` | Runtime failure: no display found, capture failed, file not writable |
| `2` | Usage error: unknown command or flag, display index out of range, `--name` with a path separator |

Errors go to stderr; stdout carries only the result.

## Platforms

Built on [`kbinani/screenshot`](https://github.com/kbinani/screenshot).
Windows is the platform it is used and tested on. Linux (X11) builds without
cgo; macOS needs cgo.

## Development

```bash
go vet ./...
go test ./...
```

Release by pushing a semver tag (`vX.Y.Z`); `go install ...@latest` picks it up.
