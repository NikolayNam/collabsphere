package http

import (
	"testing"

	platformapp "github.com/NikolayNam/collabsphere/internal/platformops/application"
	"github.com/NikolayNam/collabsphere/internal/runtime/bootstrap"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/config"
	"github.com/go-chi/chi/v5"
)

func TestRegisterDoesNotPanic(t *testing.T) {
	t.Helper()

	api := bootstrap.NewAPI(chi.NewRouter(), &config.Config{
		APP: config.App{
			Title:   "test",
			Version: "test",
		},
	})
	handler := NewHandler(platformapp.New(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil))

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Register panicked: %v", r)
		}
	}()

	Register(api, handler, nil)
}

