package migrationplan

import (
	"fmt"
	"path/filepath"
	"strings"
)

type Config struct {
	SrcDir       string
	ManifestPath string
	LockPath     string
	OutDir       string
}

type LockEntry struct {
	Version string
	Source  string
	Output  string
}

func Build(cfg Config) error {
	manifest, err := ReadManifest(cfg.ManifestPath)
	if err != nil {
		return err
	}

	if err := ValidateManifest(manifest, cfg.SrcDir); err != nil {
		return err
	}

	if err := recreateDir(cfg.OutDir); err != nil {
		return fmt.Errorf("recreate output dir: %w", err)
	}

	entries := make([]LockEntry, 0, len(manifest.Migrations))

	for i, relPath := range manifest.Migrations {
		cleanRelPath := filepath.ToSlash(filepath.Clean(strings.TrimSpace(relPath)))
		srcPath := filepath.Join(cfg.SrcDir, filepath.FromSlash(cleanRelPath))

		baseName := filepath.Base(cleanRelPath)
		version := formatVersion(i + 1)
		dstName := version + "_" + baseName
		dstPath := filepath.Join(cfg.OutDir, dstName)

		if err := copyFile(srcPath, dstPath); err != nil {
			return fmt.Errorf("copy %q -> %q: %w", srcPath, dstPath, err)
		}

		entries = append(entries, LockEntry{
			Version: version,
			Source:  cleanRelPath,
			Output:  dstName,
		})
	}

	if err := WriteManifestLock(cfg.LockPath, entries); err != nil {
		return fmt.Errorf("write manifest.lock: %w", err)
	}

	return nil
}

func formatVersion(n int) string {
	return fmt.Sprintf("%04d", n)
}
