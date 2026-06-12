// Command specutil is the deterministic CLI for projecting spec-framework change
// artifacts into other artifacts and visualizations. It performs no network I/O.
package main

import (
	"os"

	"github.com/roshbhatia/specutil/internal/cli"
)

func main() {
	if err := cli.NewRootCmd().Execute(); err != nil {
		// A missing lock entry is a distinguishable, non-fatal outcome (exit 3)
		// so `lock get` callers can branch on "absent" vs a real error.
		if cli.IsNoMapping(err) {
			os.Exit(3)
		}
		os.Exit(1)
	}
}
