package app

import (
	"os/exec"
	"path/filepath"
	"testing"
)

func TestOpenCommandUsesPlatformHandler(t *testing.T) {
	tests := []struct {
		name       string
		goos       string
		path       string
		wantBinary string
		wantArg    string
		wantErr    string
	}{
		{name: "macos", goos: "darwin", path: "/tmp/notes", wantBinary: "open", wantArg: filepath.Clean("/tmp/notes")},
		{name: "linux", goos: "linux", path: "/tmp/notes", wantBinary: "xdg-open", wantArg: filepath.Clean("/tmp/notes")},
		{name: "windows", goos: "windows", path: `C:\notes`, wantBinary: "explorer", wantArg: filepath.Clean(`C:\notes`)},
		{name: "unsupported", goos: "plan9", path: "/tmp/notes", wantErr: "unsupported platform: plan9"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := openCommand(tt.goos, tt.path)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("openCommand(%q, %q) error = nil, want %q", tt.goos, tt.path, tt.wantErr)
				}
				if err.Error() != tt.wantErr {
					t.Fatalf("openCommand(%q, %q) error = %q, want %q", tt.goos, tt.path, err.Error(), tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("openCommand(%q, %q) error = %v", tt.goos, tt.path, err)
			}
			if filepath.Base(got.Path) != tt.wantBinary {
				t.Fatalf("filepath.Base(command.Path) = %q, want %q", filepath.Base(got.Path), tt.wantBinary)
			}
			if len(got.Args) != 2 {
				t.Fatalf("len(command.Args) = %d, want 2", len(got.Args))
			}
			if got.Args[1] != tt.wantArg {
				t.Fatalf("command.Args[1] = %q, want %q", got.Args[1], tt.wantArg)
			}
		})
	}
}

func TestOpenURLCommandUsesPlatformHandler(t *testing.T) {
	tests := []struct {
		name       string
		goos       string
		rawURL     string
		wantBinary string
		wantArg    string
		wantErr    string
	}{
		{name: "macos", goos: "darwin", rawURL: "https://example.com/docs?q=1", wantBinary: "open", wantArg: "https://example.com/docs?q=1"},
		{name: "linux", goos: "linux", rawURL: "https://example.com/docs?q=1", wantBinary: "xdg-open", wantArg: "https://example.com/docs?q=1"},
		{name: "windows", goos: "windows", rawURL: "https://example.com/docs?q=1", wantBinary: "explorer", wantArg: "https://example.com/docs?q=1"},
		{name: "unsupported", goos: "plan9", rawURL: "https://example.com/docs?q=1", wantErr: "unsupported platform: plan9"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := openURLCommand(tt.goos, tt.rawURL)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("openURLCommand(%q, %q) error = nil, want %q", tt.goos, tt.rawURL, tt.wantErr)
				}
				if err.Error() != tt.wantErr {
					t.Fatalf("openURLCommand(%q, %q) error = %q, want %q", tt.goos, tt.rawURL, err.Error(), tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("openURLCommand(%q, %q) error = %v", tt.goos, tt.rawURL, err)
			}
			if filepath.Base(got.Path) != tt.wantBinary {
				t.Fatalf("filepath.Base(command.Path) = %q, want %q", filepath.Base(got.Path), tt.wantBinary)
			}
			if len(got.Args) != 2 {
				t.Fatalf("len(command.Args) = %d, want 2", len(got.Args))
			}
			if got.Args[1] != tt.wantArg {
				t.Fatalf("command.Args[1] = %q, want %q", got.Args[1], tt.wantArg)
			}
		})
	}
}

func TestClipboardCommandUsesPlatformHandler(t *testing.T) {
	originalLookPath := lookPath
	t.Cleanup(func() {
		lookPath = originalLookPath
	})

	tests := []struct {
		name       string
		goos       string
		available  map[string]bool
		wantBinary string
		wantArgs   []string
		wantErr    string
	}{
		{name: "macos", goos: "darwin", wantBinary: "pbcopy", wantArgs: []string{"pbcopy"}},
		{name: "linux wayland", goos: "linux", available: map[string]bool{"wl-copy": true}, wantBinary: "wl-copy", wantArgs: []string{"wl-copy"}},
		{name: "linux xclip fallback", goos: "linux", available: map[string]bool{"xclip": true}, wantBinary: "xclip", wantArgs: []string{"xclip", "-selection", "clipboard"}},
		{name: "linux xsel fallback", goos: "linux", available: map[string]bool{"xsel": true}, wantBinary: "xsel", wantArgs: []string{"xsel", "--clipboard", "--input"}},
		{name: "linux missing clipboard tools", goos: "linux", wantErr: "clipboard support requires wl-copy, xclip, or xsel"},
		{name: "windows", goos: "windows", wantBinary: "cmd", wantArgs: []string{"cmd", "/c", "clip"}},
		{name: "unsupported", goos: "plan9", wantErr: "unsupported platform: plan9"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lookPath = func(file string) (string, error) {
				if tt.available[file] {
					return "/usr/bin/" + file, nil
				}
				return "", exec.ErrNotFound
			}

			got, err := clipboardCommand(tt.goos)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("clipboardCommand(%q) error = nil, want %q", tt.goos, tt.wantErr)
				}
				if err.Error() != tt.wantErr {
					t.Fatalf("clipboardCommand(%q) error = %q, want %q", tt.goos, err.Error(), tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("clipboardCommand(%q) error = %v", tt.goos, err)
			}
			if filepath.Base(got.Path) != tt.wantBinary {
				t.Fatalf("filepath.Base(command.Path) = %q, want %q", filepath.Base(got.Path), tt.wantBinary)
			}
			if len(got.Args) != len(tt.wantArgs) {
				t.Fatalf("len(command.Args) = %d, want %d", len(got.Args), len(tt.wantArgs))
			}
			for i := range tt.wantArgs {
				if got.Args[i] != tt.wantArgs[i] {
					t.Fatalf("command.Args[%d] = %q, want %q", i, got.Args[i], tt.wantArgs[i])
				}
			}
		})
	}
}
