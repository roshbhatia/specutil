// Package guard hosts the determinism guard: a test asserting that no
// non-test code path in the binary imports a network package. This is the
// enforcement arm of the core invariant — the binary is pure and offline; all
// remote I/O is delegated to the agent via the shipped sync skills.
package guard

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// networkPackages are stdlib import paths that perform outbound network I/O.
// net/url is intentionally excluded: it is pure parsing with no I/O.
var networkPackages = map[string]bool{
	"net":           true,
	"net/http":      true,
	"net/http/cgi":  true,
	"net/http/fcgi": true,
	"net/rpc":       true,
	"net/smtp":      true,
	"net/mail":      true,
	"net/textproto": true,
}

func TestNoNetworkImportsInBinary(t *testing.T) {
	root := moduleRoot(t)
	fset := token.NewFileSet()

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			// Skip vendored, hidden, and the guard package's own dir is fine to scan
			// (this file is a _test.go and is excluded below).
			base := d.Name()
			if base == ".git" || base == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		// The determinism boundary governs binary code paths, not tests.
		if strings.HasSuffix(path, "_test.go") {
			return nil
		}

		f, perr := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if perr != nil {
			t.Errorf("parsing %s: %v", path, perr)
			return nil
		}
		for _, imp := range f.Imports {
			p, uerr := strconv.Unquote(imp.Path.Value)
			if uerr != nil {
				continue
			}
			if networkPackages[p] {
				rel, _ := filepath.Rel(root, path)
				t.Errorf("network import %q in binary code path %s (line %d): the binary must perform no network I/O — delegate remote writes to the sync skills",
					p, rel, fset.Position(imp.Pos()).Line)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walking module: %v", err)
	}
}

// moduleRoot walks up from the test's working directory until it finds go.mod.
func moduleRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("could not locate go.mod above %s", dir)
		}
		dir = parent
	}
}
