package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/go-chi/chi/v5"

	appbootstrap "github.com/NikolayNam/collabsphere/internal/runtime/bootstrap/app"
	"github.com/NikolayNam/collabsphere/internal/runtime/foundation/config"
)

type routeSpec struct {
	Method string `json:"method"`
	Path   string `json:"path"`
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(2)
	}

	conf := config.NewFor(config.ProfileContracts)
	application := appbootstrap.NewContracts(conf)

	switch os.Args[1] {
	case "openapi-json":
		if err := writeOpenAPIJSON(application); err != nil {
			fail(err)
		}
	case "openapi-yaml":
		if err := writeOpenAPIYAML(application); err != nil {
			fail(err)
		}
	case "routes":
		if err := writeRoutes(application); err != nil {
			fail(err)
		}
	case "check-parity":
		if err := checkParity(application); err != nil {
			fail(err)
		}
	default:
		printUsage()
		os.Exit(2)
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "usage: go run ./cmd/contracts <openapi-json|openapi-yaml|routes|check-parity>")
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err.Error())
	os.Exit(1)
}

func writeOpenAPIJSON(application *appbootstrap.App) error {
	data, err := json.MarshalIndent(application.API.OpenAPI(), "", "  ")
	if err != nil {
		return fmt.Errorf("marshal openapi json: %w", err)
	}
	_, err = os.Stdout.Write(append(data, '\n'))
	return err
}

func writeOpenAPIYAML(application *appbootstrap.App) error {
	data, err := application.API.OpenAPI().YAML()
	if err != nil {
		return fmt.Errorf("marshal openapi yaml: %w", err)
	}
	_, err = os.Stdout.Write(data)
	return err
}

func writeRoutes(application *appbootstrap.App) error {
	routes, err := collectV1Routes(application)
	if err != nil {
		return err
	}
	for _, route := range routes {
		if _, err := fmt.Fprintf(os.Stdout, "%s %s\n", route.Method, route.Path); err != nil {
			return err
		}
	}
	return nil
}

func checkParity(application *appbootstrap.App) error {
	routerRoutes, err := collectV1Routes(application)
	if err != nil {
		return err
	}
	openapiRoutes := collectOpenAPIRoutes(application)

	routerSet := make(map[string]struct{}, len(routerRoutes))
	for _, route := range routerRoutes {
		routerSet[route.Method+" "+route.Path] = struct{}{}
	}
	openapiSet := make(map[string]struct{}, len(openapiRoutes))
	for _, route := range openapiRoutes {
		openapiSet[route.Method+" "+route.Path] = struct{}{}
	}

	missingInOpenAPI := make([]string, 0)
	for key := range routerSet {
		if _, ok := openapiSet[key]; !ok {
			missingInOpenAPI = append(missingInOpenAPI, key)
		}
	}
	missingInRouter := make([]string, 0)
	for key := range openapiSet {
		if _, ok := routerSet[key]; !ok {
			missingInRouter = append(missingInRouter, key)
		}
	}

	sort.Strings(missingInOpenAPI)
	sort.Strings(missingInRouter)

	if len(missingInOpenAPI) == 0 && len(missingInRouter) == 0 {
		return nil
	}

	var buf bytes.Buffer
	buf.WriteString("route/openapi parity check failed\n")
	if len(missingInOpenAPI) > 0 {
		buf.WriteString("missing in OpenAPI:\n")
		for _, item := range missingInOpenAPI {
			buf.WriteString("  - " + item + "\n")
		}
	}
	if len(missingInRouter) > 0 {
		buf.WriteString("missing in router:\n")
		for _, item := range missingInRouter {
			buf.WriteString("  - " + item + "\n")
		}
	}
	return fmt.Errorf("%s", strings.TrimRight(buf.String(), "\n"))
}

func collectV1Routes(application *appbootstrap.App) ([]routeSpec, error) {
	routes := make([]routeSpec, 0, 128)
	if err := chi.Walk(application.Router, func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		_ = handler
		_ = middlewares
		method = strings.ToUpper(strings.TrimSpace(method))
		route = normalizeRoutePath(route)
		if method == "" || route == "" {
			return nil
		}
		if !strings.HasPrefix(route, "/v1/") {
			return nil
		}
		if shouldIgnoreContractRoute(method, route) {
			return nil
		}
		routes = append(routes, routeSpec{Method: method, Path: route})
		return nil
	}); err != nil {
		return nil, fmt.Errorf("walk router: %w", err)
	}

	sort.Slice(routes, func(i, j int) bool {
		if routes[i].Path == routes[j].Path {
			return routes[i].Method < routes[j].Method
		}
		return routes[i].Path < routes[j].Path
	})
	return routes, nil
}

func collectOpenAPIRoutes(application *appbootstrap.App) []routeSpec {
	oapi := application.API.OpenAPI()
	routes := make([]routeSpec, 0, len(oapi.Paths)*2)
	for path, item := range oapi.Paths {
		path = normalizeRoutePath("/v1" + path)
		appendOperationRoute(&routes, path, "GET", item.Get)
		appendOperationRoute(&routes, path, "POST", item.Post)
		appendOperationRoute(&routes, path, "PUT", item.Put)
		appendOperationRoute(&routes, path, "PATCH", item.Patch)
		appendOperationRoute(&routes, path, "DELETE", item.Delete)
		appendOperationRoute(&routes, path, "OPTIONS", item.Options)
		appendOperationRoute(&routes, path, "HEAD", item.Head)
	}
	sort.Slice(routes, func(i, j int) bool {
		if routes[i].Path == routes[j].Path {
			return routes[i].Method < routes[j].Method
		}
		return routes[i].Path < routes[j].Path
	})
	return routes
}

func appendOperationRoute(routes *[]routeSpec, path, method string, operation *huma.Operation) {
	if operation == nil {
		return
	}
	if shouldIgnoreContractRoute(method, path) {
		return
	}
	*routes = append(*routes, routeSpec{Method: method, Path: path})
}

func shouldIgnoreContractRoute(method, path string) bool {
	if method == "OPTIONS" {
		return true
	}
	switch {
	case path == "/v1/docs":
		return true
	case strings.HasPrefix(path, "/v1/docs/"):
		return true
	case strings.HasPrefix(path, "/v1/openapi"):
		return true
	case strings.HasPrefix(path, "/v1/schemas"):
		return true
	default:
		return false
	}
}

func normalizeRoutePath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	for strings.Contains(path, "//") {
		path = strings.ReplaceAll(path, "//", "/")
	}
	if len(path) > 1 {
		path = strings.TrimRight(path, "/")
	}
	return path
}
