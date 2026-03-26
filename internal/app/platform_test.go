package app

import (
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
