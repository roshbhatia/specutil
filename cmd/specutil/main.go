// Command specutil is the deterministic CLI for projecting spec-framework change
// artifacts into other artifacts and visualizations. It performs no network I/O.
package main

import (
	"os"

	"github.com/roshbhatia/specutil/internal/cli"
)

func main() {
	if err := cli.NewRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
