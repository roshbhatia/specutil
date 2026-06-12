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

// TestWebRuntimeIsVendoredNotFetched guards the offline half of the determinism
// boundary for the web viewer: the Cytoscape/dagre runtime must be vendored on
// disk and inlined into the page, never pulled from a CDN at view time. A drift
// back to a `<script src="https://…">` reference would make the "no network"
// promise a lie even though no Go network import changed.
func TestWebRuntimeIsVendoredNotFetched(t *testing.T) {
	root := moduleRoot(t)
	assets := filepath.Join(root, "internal", "web", "assets")

	for _, bundle := range []string{"cytoscape.min.js", "dagre.min.js", "cytoscape-dagre.min.js"} {
		if _, err := os.Stat(filepath.Join(assets, bundle)); err != nil {
			t.Errorf("web runtime bundle %s is not vendored on disk: %v", bundle, err)
		}
	}

	// The Mac OS theme (system.css) is a vendored client-side dependency too: its
	// fonts and button SVGs are inlined as data URIs and the stylesheet itself is
	// inlined via a template field, so it never reaches for a CDN either.
	if _, err := os.Stat(filepath.Join(assets, "system.css")); err != nil {
		t.Errorf("system.css theme is not vendored on disk: %v", err)
	}

	tmpl, err := os.ReadFile(filepath.Join(assets, "page.html.tmpl"))
	if err != nil {
		t.Fatalf("reading page template: %v", err)
	}
	src := string(tmpl)
	// The runtime must be inlined via template fields, not referenced externally.
	for _, want := range []string{"{{.CytoscapeJS}}", "{{.DagreJS}}", "{{.CytoscapeDagreJS}}", "{{.SystemCSS}}"} {
		if !strings.Contains(src, want) {
			t.Errorf("page template must inline %s; the runtime cannot be fetched at view time", want)
		}
	}
	for _, bad := range []string{"<script src", "<link ", "cdn.jsdelivr", "unpkg.com", "cdnjs"} {
		if strings.Contains(src, bad) {
			t.Errorf("page template references external runtime %q; the viewer must be self-contained and offline", bad)
		}
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
