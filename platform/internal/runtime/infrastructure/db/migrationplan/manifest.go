package migrationplan

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

type Manifest struct {
	Migrations []string `yaml:"migrations"`
}

func ReadManifest(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read manifest: %w", err)
	}

	var m Manifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("unmarshal manifest yaml: %w", err)
	}

	if len(m.Migrations) == 0 {
		return nil, errors.New("manifest is empty: no migrations declared")
	}

	return &m, nil
}

func ValidateManifest(m *Manifest, srcDir string) error {
	seenPaths := make(map[string]int)
	seenBaseNames := make(map[string][]string)

	var errs []string

	for idx, raw := range m.Migrations {
		p := strings.TrimSpace(raw)
		if p == "" {
			errs = append(errs, fmt.Sprintf("entry #%d is empty", idx+1))
			continue
		}

		if filepath.IsAbs(p) {
			errs = append(errs, fmt.Sprintf("entry %q must be relative, not absolute", p))
			continue
		}

		if strings.Contains(p, `\`) {
			errs = append(errs, fmt.Sprintf("entry %q must use '/' instead of '\\\\'", p))
		}

		clean := filepath.ToSlash(filepath.Clean(p))
		if clean == "." {
			errs = append(errs, fmt.Sprintf("entry %q is invalid", p))
			continue
		}

		if strings.HasPrefix(clean, "../") || clean == ".." {
			errs = append(errs, fmt.Sprintf("entry %q escapes migrations-src", p))
			continue
		}

		if filepath.Ext(clean) != ".sql" {
			errs = append(errs, fmt.Sprintf("entry %q must be a .sql file", p))
		}

		if prev, ok := seenPaths[clean]; ok {
			errs = append(errs, fmt.Sprintf("duplicate migration path %q at positions %d and %d", clean, prev, idx+1))
		} else {
			seenPaths[clean] = idx + 1
		}

		base := filepath.Base(clean)
		seenBaseNames[base] = append(seenBaseNames[base], clean)

		fullPath := filepath.Join(srcDir, filepath.FromSlash(clean))
		info, err := os.Stat(fullPath)
		if err != nil {
			if os.IsNotExist(err) {
				errs = append(errs, fmt.Sprintf("file %q does not exist", clean))
			} else {
				errs = append(errs, fmt.Sprintf("stat %q: %v", clean, err))
			}
			continue
		}

		if info.IsDir() {
			errs = append(errs, fmt.Sprintf("entry %q is a directory, expected file", clean))
			continue
		}

		content, err := os.ReadFile(fullPath)
		if err != nil {
			errs = append(errs, fmt.Sprintf("read %q: %v", clean, err))
			continue
		}

		if err := validateGooseDirectives(content); err != nil {
			errs = append(errs, fmt.Sprintf("file %q: %v", clean, err))
		}
	}

	var dupBaseNames []string
	for base, paths := range seenBaseNames {
		if len(paths) > 1 {
			sort.Strings(paths)
			dupBaseNames = append(dupBaseNames, fmt.Sprintf("duplicate basename %q -> %s", base, strings.Join(paths, ", ")))
		}
	}
	sort.Strings(dupBaseNames)
	errs = append(errs, dupBaseNames...)

	if len(errs) > 0 {
		return errors.New("manifest validation errors:\n - " + strings.Join(errs, "\n - "))
	}

	return nil
}

func validateGooseDirectives(content []byte) error {
	normalized := strings.TrimPrefix(string(content), "\ufeff")
	for _, rawLine := range strings.Split(normalized, "\n") {
		line := strings.TrimSpace(rawLine)
		if line == "" {
			continue
		}
		if !strings.EqualFold(line, "-- +goose Up") {
			return fmt.Errorf("must start with '-- +goose Up'")
		}
		return nil
	}

	return fmt.Errorf("is empty")
}
