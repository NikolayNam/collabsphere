package main

import (
	"fmt"
	"os"

	"github.com/NikolayNam/collabsphere/internal/runtime/infrastructure/db/migrationplan"
)

func main() {
	cfg := migrationplan.Config{
		SrcDir:       "internal/runtime/infrastructure/db/migrations-src",
		ManifestPath: "internal/runtime/infrastructure/db/migrations-src/manifest.yaml",
		LockPath:     "internal/runtime/infrastructure/db/migrations-src/manifest.lock",
		OutDir:       "internal/runtime/infrastructure/db/migrations",
	}

	if err := migrationplan.Build(cfg); err != nil {
		_, err := fmt.Fprintf(os.Stderr, "build migrations failed: %v\n", err)
		if err != nil {
			return
		}
		os.Exit(1)
	}
}
