package bootstrap

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

type scalarDocsPage struct {
	Title      string
	ConfigJSON template.JS
}

var scalarDocsTemplate = template.Must(template.New("scalar-docs").Parse(`<!doctype html>
<html lang="ru">
  <head>
    <meta charset="utf-8">
    <meta name="referrer" content="no-referrer">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>{{ .Title }}</title>
  </head>
  <body>
    <div id="app"></div>
    <script>
      const configuration = {{ .ConfigJSON }};
      window.__scalarApiReferenceConfiguration = configuration;
    </script>
    <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference@1.44.20/dist/browser/standalone.min.js"></script>
    <script>
      function parseImagePreviewPayload(text) {
        if (!text) {
          return null;
        }

        try {
          const payload = JSON.parse(text);
          if (!payload || typeof payload !== 'object') {
            return null;
          }

          if (typeof payload.downloadUrl !== 'string' || payload.downloadUrl.length === 0) {
            return null;
          }

          if (typeof payload.contentType !== 'string' || !payload.contentType.toLowerCase().startsWith('image/')) {
            return null;
          }

          return payload;
        } catch {
          return null;
        }
      }

      function createImagePreviewCard(payload) {
        const wrapper = document.createElement('div');
        wrapper.className = 'cs-image-preview';

        const heading = document.createElement('div');
        heading.className = 'cs-image-preview__heading';
        heading.textContent = 'Image preview';

        const image = document.createElement('img');
        image.className = 'cs-image-preview__image';
        image.src = payload.downloadUrl;
        image.alt = payload.fileName || 'Downloaded image preview';
        image.loading = 'lazy';
        image.decoding = 'async';
        image.referrerPolicy = 'no-referrer';

        const meta = document.createElement('div');
        meta.className = 'cs-image-preview__meta';
        meta.textContent = [payload.fileName || 'image', payload.contentType || ''].filter(Boolean).join(' • ');

        const link = document.createElement('a');
        link.className = 'cs-image-preview__link';
        link.href = payload.downloadUrl;
        link.target = '_blank';
        link.rel = 'noreferrer noopener';
        link.textContent = 'Open original image';

        wrapper.appendChild(heading);
        wrapper.appendChild(image);
        wrapper.appendChild(meta);
        wrapper.appendChild(link);
        return wrapper;
      }

      function attachImagePreview(pre) {
        if (!(pre instanceof HTMLElement)) {
          return;
        }

        if (pre.dataset.csImagePreviewProcessed === 'true') {
          return;
        }
        pre.dataset.csImagePreviewProcessed = 'true';

        const payload = parseImagePreviewPayload(pre.textContent || '');
        if (!payload) {
          return;
        }

        const anchor = pre.closest('.scalar-api-client__content') || pre.parentElement;
        if (!anchor || !(anchor instanceof HTMLElement)) {
          return;
        }

        const next = anchor.nextElementSibling;
        if (next && next.classList.contains('cs-image-preview')) {
          return;
        }

        anchor.insertAdjacentElement('afterend', createImagePreviewCard(payload));
      }

      function scanForImagePreviews(root) {
        if (!(root instanceof HTMLElement)) {
          return;
        }

        root.querySelectorAll('pre').forEach((pre) => attachImagePreview(pre));
      }

      function installImagePreviewObserver() {
        const app = document.getElementById('app');
        if (!app) {
          return;
        }

        let timer = null;
        const scheduleScan = () => {
          if (timer) {
            window.clearTimeout(timer);
          }
          timer = window.setTimeout(() => scanForImagePreviews(app), 120);
        };

        const observer = new MutationObserver(() => scheduleScan());
        observer.observe(app, {
          childList: true,
          subtree: true,
          characterData: true,
        });

        window.setTimeout(() => scanForImagePreviews(app), 250);
      }

      Scalar.createApiReference('#app', window.__scalarApiReferenceConfiguration)
      installImagePreviewObserver()
    </script>
  </body>
</html>`))

func RegisterScalarDocs(router chi.Router, title, openAPIURL string) {
	config := map[string]any{
		"url":                     openAPIURL,
		"theme":                   "deepSpace",
		"darkMode":                true,
		"layout":                  "modern",
		"showSidebar":             true,
		"hideModels":              true,
		"defaultOpenAllTags":      false,
		"showOperationId":         false,
		"showDeveloperTools":      "never",
		"operationTitleSource":    "summary",
		"orderSchemaPropertiesBy": "preserve",
		"withDefaultFonts":        true,
		"hideDarkModeToggle":      true,
		"defaultHttpClient": map[string]any{
			"targetKey": "http",
			"clientKey": "http1.1",
		},
		"metaData": map[string]any{
			"title":       title + " API Reference",
			"description": "Interactive API reference for " + title + " with grouped navigation, consistent summaries, and HTTP/1.1 selected by default",
		},
		"customCss": strings.TrimSpace(`
:root {
  --scalar-font: "Inter", ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
  --scalar-font-code: "JetBrains Mono", ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
}

.light-mode,
.dark-mode {
  --scalar-sidebar-indent-border: color-mix(in srgb, var(--scalar-border-color) 70%, transparent);
}

.dark-mode {
  --scalar-background-1: #070b14;
  --scalar-background-2: #0d1324;
  --scalar-background-3: #11182b;
  --scalar-color-1: #eef4ff;
  --scalar-color-2: #c0cae0;
  --scalar-border-color: rgba(179, 196, 255, 0.12);
  --scalar-sidebar-background-1: #060a13;
  --scalar-sidebar-item-hover-background: rgba(138, 180, 255, 0.10);
  --scalar-sidebar-item-active-background: rgba(138, 180, 255, 0.16);
}

.dark-mode body,
.dark-mode #app {
  background:
    radial-gradient(circle at top left, rgba(82, 123, 255, 0.10), transparent 28%),
    radial-gradient(circle at top right, rgba(0, 201, 255, 0.08), transparent 22%),
    #070b14;
}

.dark-mode .sidebar {
  border-right: 1px solid rgba(179, 196, 255, 0.10);
  box-shadow: inset -1px 0 0 rgba(179, 196, 255, 0.04);
  width: 320px;
}

.dark-mode .sidebar a,
.dark-mode .sidebar button {
  border-radius: 10px;
}

.dark-mode .sidebar [data-slot="tag-group"] {
  font-size: 0.78rem;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: #9fb0d1;
  margin-top: 1.25rem;
  margin-bottom: 0.5rem;
}

.dark-mode .sidebar [data-slot="tag"] {
  font-weight: 600;
}

.dark-mode .sidebar [data-slot="tag"] small {
  opacity: 0.72;
}

.dark-mode .sidebar [data-active="true"] {
  box-shadow: inset 0 0 0 1px rgba(138, 180, 255, 0.18);
}

.t-doc__content {
  max-width: 1040px;
}

.t-doc__content h1 {
  font-size: 2.2rem;
  letter-spacing: -0.03em;
}

.t-doc__content h2 {
  margin-top: 2rem;
}

.t-doc__content p,
.t-doc__content li {
  line-height: 1.7;
}

section[aria-label="Request"] pre,
section[aria-label="Response"] pre,
.scalar-card,
.scalar-api-client__content {
  border-radius: 14px;
  border: 1px solid rgba(179, 196, 255, 0.12);
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.16);
}

.dark-mode .scalar-api-client__container,
.dark-mode .scalar-card {
  background: linear-gradient(180deg, rgba(17, 24, 43, 0.96), rgba(11, 16, 30, 0.96));
}

.dark-mode .scalar-api-client__url,
.dark-mode .scalar-api-client__method,
.dark-mode pre,
.dark-mode code {
  font-feature-settings: "liga" 0;
}

.dark-mode input,
.dark-mode textarea,
.dark-mode select {
  border-radius: 10px !important;
}

.dark-mode .scalar-search-input {
  border-radius: 12px;
  box-shadow: inset 0 0 0 1px rgba(179, 196, 255, 0.10);
}

.cs-image-preview {
  margin-top: 1rem;
  padding: 1rem;
  border-radius: 14px;
  border: 1px solid rgba(179, 196, 255, 0.12);
  background: linear-gradient(180deg, rgba(17, 24, 43, 0.92), rgba(11, 16, 30, 0.92));
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.18);
}

.cs-image-preview__heading {
  margin-bottom: 0.75rem;
  font-size: 0.78rem;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: #9fb0d1;
}

.cs-image-preview__image {
  display: block;
  max-width: 100%;
  max-height: 320px;
  border-radius: 12px;
  border: 1px solid rgba(179, 196, 255, 0.10);
  background: rgba(7, 11, 20, 0.85);
}

.cs-image-preview__meta {
  margin-top: 0.75rem;
  font-size: 0.92rem;
  color: #c0cae0;
}

.cs-image-preview__link {
  display: inline-block;
  margin-top: 0.5rem;
  color: #8ab4ff;
  text-decoration: none;
  font-weight: 600;
}

.cs-image-preview__link:hover {
  text-decoration: underline;
}

button[aria-label="Ask AI Agent"] {
  display: none !important;
}
`),
	}

	configJSON, err := json.Marshal(config)
	if err != nil {
		panic("marshal scalar docs config: " + err.Error())
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		csp := []string{
			"default-src 'none'",
			"base-uri 'none'",
			"connect-src 'self'",
			"img-src 'self' data: blob: http: https:",
			"font-src https://fonts.scalar.com data:",
			"form-action 'none'",
			"frame-ancestors 'none'",
			"sandbox allow-same-origin allow-scripts",
			"script-src 'unsafe-inline' 'unsafe-eval' https://cdn.jsdelivr.net",
			"style-src 'unsafe-inline'",
		}
		w.Header().Set("Content-Security-Policy", strings.Join(csp, "; "))
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := scalarDocsTemplate.Execute(w, scalarDocsPage{
			Title:      title + " API Reference",
			ConfigJSON: template.JS(configJSON),
		}); err != nil {
			http.Error(w, "failed to render API reference", http.StatusInternalServerError)
		}
	}

	router.Get("/docs", handler)
	router.Get("/docs/", handler)
}
