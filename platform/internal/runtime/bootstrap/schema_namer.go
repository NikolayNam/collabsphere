package bootstrap

import (
	"fmt"
	"hash/fnv"
	"reflect"
	"strconv"
	"strings"
	"unicode"

	"github.com/danielgtaylor/huma/v2"
)

const moduleImportPath = "github.com/NikolayNam/collabsphere/"

func newSchemaNamer() func(reflect.Type, string) string {
	assigned := map[reflect.Type]string{}
	used := map[string]reflect.Type{}

	return func(t reflect.Type, hint string) string {
		normalized := normalizeSchemaType(t)
		if name, ok := assigned[normalized]; ok {
			return name
		}

		base := huma.DefaultSchemaNamer(normalized, hint)
		if prefix := schemaPackagePrefix(normalized); prefix != "" {
			base = prefix + base
		} else if normalized.Name() == "" {
			base = base + shortSchemaHash(schemaFingerprint(normalized, hint))
		}

		name := base
		for suffix := 2; ; suffix++ {
			existing, ok := used[name]
			if !ok || existing == normalized {
				assigned[normalized] = name
				used[name] = normalized
				return name
			}
			name = base + strconv.Itoa(suffix)
		}
	}
}

func normalizeSchemaType(t reflect.Type) reflect.Type {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return t
}

func schemaPackagePrefix(t reflect.Type) string {
	t = normalizeSchemaType(t)
	if t.Name() == "" {
		return ""
	}

	pkg := strings.TrimPrefix(t.PkgPath(), moduleImportPath)
	if pkg == "" {
		return ""
	}

	parts := strings.FieldsFunc(pkg, func(r rune) bool {
		return r == '/' || r == '.' || r == '-' || r == '_'
	})

	var b strings.Builder
	for _, part := range parts {
		if part == "" {
			continue
		}
		b.WriteString(upperFirst(part))
	}
	return b.String()
}

func upperFirst(s string) string {
	if s == "" {
		return s
	}
	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

func shortSchemaHash(s string) string {
	h := fnv.New32a()
	_, _ = h.Write([]byte(s))
	return fmt.Sprintf("%08x", h.Sum32())
}

func schemaFingerprint(t reflect.Type, hint string) string {
	t = normalizeSchemaType(t)

	switch t.Kind() {
	case reflect.Slice:
		return "slice:" + schemaFingerprint(t.Elem(), hint)
	case reflect.Array:
		return fmt.Sprintf("array:%d:%s", t.Len(), schemaFingerprint(t.Elem(), hint))
	case reflect.Map:
		return "map:" + schemaFingerprint(t.Key(), hint) + ":" + schemaFingerprint(t.Elem(), hint)
	case reflect.Struct:
		if t.Name() != "" {
			return t.PkgPath() + "." + t.Name()
		}

		var b strings.Builder
		b.WriteString("struct:")
		b.WriteString(hint)
		b.WriteByte('{')
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			b.WriteString(f.Name)
			b.WriteByte(':')
			b.WriteString(schemaFingerprint(f.Type, hint+"."+f.Name))
			b.WriteByte('|')
			b.WriteString(string(f.Tag))
			b.WriteByte(';')
		}
		b.WriteByte('}')
		return b.String()
	default:
		if t.Name() != "" && t.PkgPath() != "" {
			return t.PkgPath() + "." + t.Name()
		}
		return t.String()
	}
}
