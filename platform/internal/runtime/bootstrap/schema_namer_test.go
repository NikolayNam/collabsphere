package bootstrap

import (
	"reflect"
	"testing"

	accdto "github.com/NikolayNam/collabsphere/internal/accounts/delivery/http/dto"
	authdto "github.com/NikolayNam/collabsphere/internal/auth/delivery/http/dto"
	catalogdto "github.com/NikolayNam/collabsphere/internal/catalog/delivery/http/dto"
	orgdto "github.com/NikolayNam/collabsphere/internal/organizations/delivery/http/dto"
	"github.com/danielgtaylor/huma/v2"
)

func TestSchemaNamerDisambiguatesDuplicateNamedTypesAcrossPackages(t *testing.T) {
	registry := huma.NewMapRegistry("#/components/schemas/", newSchemaNamer())

	var accountUploadSchema *huma.Schema
	var organizationUploadSchema *huma.Schema
	var authEmptySchema *huma.Schema
	var catalogEmptySchema *huma.Schema

	assertNotPanics(t, func() {
		accountUploadSchema = registry.Schema(reflect.TypeOf(accdto.UploadResponse{}), true, "")
		organizationUploadSchema = registry.Schema(reflect.TypeOf(orgdto.UploadResponse{}), true, "")
		authEmptySchema = registry.Schema(reflect.TypeOf(authdto.EmptyResponse{}), true, "")
		catalogEmptySchema = registry.Schema(reflect.TypeOf(catalogdto.EmptyResponse{}), true, "")
	})

	if accountUploadSchema.Ref == organizationUploadSchema.Ref {
		t.Fatalf("expected unique schema refs for duplicate UploadResponse types, got %q", accountUploadSchema.Ref)
	}
	if authEmptySchema.Ref == catalogEmptySchema.Ref {
		t.Fatalf("expected unique schema refs for duplicate EmptyResponse types, got %q", authEmptySchema.Ref)
	}
}

func TestSchemaNamerDisambiguatesAnonymousBodyTypesWithSameHint(t *testing.T) {
	registry := huma.NewMapRegistry("#/components/schemas/", newSchemaNamer())

	accountBodyField, ok := reflect.TypeOf(accdto.UploadResponse{}).FieldByName("Body")
	if !ok {
		t.Fatal("accounts UploadResponse is missing Body field")
	}
	organizationBodyField, ok := reflect.TypeOf(orgdto.UploadResponse{}).FieldByName("Body")
	if !ok {
		t.Fatal("organizations UploadResponse is missing Body field")
	}

	var accountBodySchema *huma.Schema
	var organizationBodySchema *huma.Schema
	assertNotPanics(t, func() {
		accountBodySchema = registry.Schema(accountBodyField.Type, true, "UploadResponseBody")
		organizationBodySchema = registry.Schema(organizationBodyField.Type, true, "UploadResponseBody")
	})

	if accountBodySchema.Ref == organizationBodySchema.Ref {
		t.Fatalf("expected unique schema refs for anonymous body types, got %q", accountBodySchema.Ref)
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
