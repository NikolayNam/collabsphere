package bootstrap

import (
	"reflect"
	"testing"

	authdto "github.com/NikolayNam/collabsphere/internal/auth/delivery/http/dto"
	catalogdto "github.com/NikolayNam/collabsphere/internal/catalog/delivery/http/dto"
	collabdto "github.com/NikolayNam/collabsphere/internal/collab/delivery/http/dto"
	"github.com/danielgtaylor/huma/v2"
)

func TestSchemaNamerDisambiguatesDuplicateNamedTypesAcrossPackages(t *testing.T) {
	registry := huma.NewMapRegistry("#/components/schemas/", newSchemaNamer())

	var authEmptySchema *huma.Schema
	var catalogEmptySchema *huma.Schema
	var collabEmptySchema *huma.Schema

	assertNotPanics(t, func() {
		authEmptySchema = registry.Schema(reflect.TypeOf(authdto.EmptyResponse{}), true, "")
		catalogEmptySchema = registry.Schema(reflect.TypeOf(catalogdto.EmptyResponse{}), true, "")
		collabEmptySchema = registry.Schema(reflect.TypeOf(collabdto.EmptyResponse{}), true, "")
	})

	if authEmptySchema.Ref == catalogEmptySchema.Ref {
		t.Fatalf("expected unique schema refs for duplicate EmptyResponse types, got %q", authEmptySchema.Ref)
	}
	if authEmptySchema.Ref == collabEmptySchema.Ref {
		t.Fatalf("expected unique schema refs for duplicate EmptyResponse types, got %q", authEmptySchema.Ref)
	}
}

func TestSchemaNamerDisambiguatesAnonymousBodyTypesWithSameHint(t *testing.T) {
	registry := huma.NewMapRegistry("#/components/schemas/", newSchemaNamer())

	type firstResponse struct {
		Status int
		Body   struct {
			ObjectID string `json:"objectId"`
		}
	}

	type secondResponse struct {
		Status int
		Body   struct {
			Name string `json:"name"`
		}
	}

	firstBodyField, ok := reflect.TypeOf(firstResponse{}).FieldByName("Body")
	if !ok {
		t.Fatal("firstResponse is missing Body field")
	}
	secondBodyField, ok := reflect.TypeOf(secondResponse{}).FieldByName("Body")
	if !ok {
		t.Fatal("secondResponse is missing Body field")
	}

	var firstBodySchema *huma.Schema
	var secondBodySchema *huma.Schema
	assertNotPanics(t, func() {
		firstBodySchema = registry.Schema(firstBodyField.Type, true, "UploadResponseBody")
		secondBodySchema = registry.Schema(secondBodyField.Type, true, "UploadResponseBody")
	})

	if firstBodySchema.Ref == secondBodySchema.Ref {
		t.Fatalf("expected unique schema refs for anonymous body types, got %q", firstBodySchema.Ref)
	}
}

func assertNotPanics(t *testing.T, fn func()) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("unexpected panic: %v", r)
		}
	}()
	fn()
}
