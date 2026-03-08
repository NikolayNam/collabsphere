package migrationplan

import (
	"fmt"
	"path/filepath"
	"strconv"
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

	lockedEntries, err := ReadManifestLock(cfg.LockPath)
	if err != nil {
		return fmt.Errorf("read manifest.lock: %w", err)
	}

	entries, err := planEntries(manifest.Migrations, lockedEntries)
	if err != nil {
		return err
	}

	if err := recreateDir(cfg.OutDir); err != nil {
		return fmt.Errorf("recreate output dir: %w", err)
	}

	for _, entry := range entries {
		srcPath := filepath.Join(cfg.SrcDir, filepath.FromSlash(entry.Source))
		dstPath := filepath.Join(cfg.OutDir, entry.Output)

		if err := copyFile(srcPath, dstPath); err != nil {
			return fmt.Errorf("copy %q -> %q: %w", srcPath, dstPath, err)
		}
	}

	if err := WriteManifestLock(cfg.LockPath, entries); err != nil {
		return fmt.Errorf("write manifest.lock: %w", err)
	}

	return nil
}

func planEntries(manifestMigrations []string, lockedEntries []LockEntry) ([]LockEntry, error) {
	cleanedManifest := make([]string, 0, len(manifestMigrations))
	manifestIndex := make(map[string]int, len(manifestMigrations))
	for _, relPath := range manifestMigrations {
		clean := filepath.ToSlash(filepath.Clean(strings.TrimSpace(relPath)))
		if _, exists := manifestIndex[clean]; exists {
			return nil, fmt.Errorf("manifest contains duplicate migration %q", clean)
		}
		cleanedManifest = append(cleanedManifest, clean)
		manifestIndex[clean] = len(cleanedManifest) - 1
	}

	lockedBySource := make(map[string]LockEntry, len(lockedEntries))
	maxVersion := 0
	for _, entry := range lockedEntries {
		cleanSource := filepath.ToSlash(filepath.Clean(strings.TrimSpace(entry.Source)))
		entry.Source = cleanSource
		if _, ok := manifestIndex[cleanSource]; !ok {
			return nil, fmt.Errorf("manifest cannot remove locked migration %q", cleanSource)
		}
		if _, exists := lockedBySource[cleanSource]; exists {
			return nil, fmt.Errorf("manifest.lock contains duplicate source %q", cleanSource)
		}
		lockedBySource[cleanSource] = entry

		version, err := strconv.Atoi(entry.Version)
		if err != nil {
			return nil, fmt.Errorf("manifest.lock contains invalid version %q for %q", entry.Version, cleanSource)
		}
		if version > maxVersion {
			maxVersion = version
		}
	}

	lockedCursor := 0
	seenNew := false
	firstNewSource := ""
	for _, source := range cleanedManifest {
		_, isLocked := lockedBySource[source]
		if !isLocked {
			seenNew = true
			if firstNewSource == "" {
				firstNewSource = source
			}
			continue
		}
		if seenNew {
			return nil, fmt.Errorf("new migrations must be appended at the end; new migration %q appears before already locked migration %q", firstNewSource, source)
		}
		if lockedCursor >= len(lockedEntries) {
			return nil, fmt.Errorf("manifest reordered locked migration %q", source)
		}
		expected := filepath.ToSlash(filepath.Clean(strings.TrimSpace(lockedEntries[lockedCursor].Source)))
		if expected != source {
			return nil, fmt.Errorf("manifest reordered locked migrations: expected %q before %q", expected, source)
		}
		lockedCursor++
	}

	entries := make([]LockEntry, 0, len(cleanedManifest))
	nextVersion := maxVersion + 1
	for _, source := range cleanedManifest {
		if locked, ok := lockedBySource[source]; ok {
			entries = append(entries, locked)
			continue
		}

		version := formatVersion(nextVersion)
		nextVersion++
		entries = append(entries, LockEntry{
			Version: version,
			Source:  source,
			Output:  version + "_" + filepath.Base(source),
		})
	}

	return entries, nil
}

func formatVersion(n int) string {
	return fmt.Sprintf("%04d", n)
}
