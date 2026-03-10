package docsportal

import (
	"html/template"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type pageSection struct {
	ID          string
	Title       string
	Eyebrow     string
	Description string
	Bullets     []string
}

type pageData struct {
	Title    string
	Sections []pageSection
}

var docsTemplate = template.Must(template.New("docs-portal").Parse(`<!doctype html>
<html lang="ru">
  <head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>{{ .Title }}</title>
    <style>
      :root {
        color-scheme: dark;
        --bg: #07101d;
        --bg-2: #0c1728;
        --panel: rgba(13, 24, 42, 0.92);
        --panel-2: rgba(17, 30, 52, 0.92);
        --text: #edf3ff;
        --muted: #b7c4de;
        --line: rgba(150, 176, 255, 0.16);
        --accent: #7bb0ff;
        --accent-2: #86f0ff;
        --shadow: 0 18px 48px rgba(0, 0, 0, 0.28);
      }
      * { box-sizing: border-box; }
      body {
        margin: 0;
        font-family: Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
        color: var(--text);
        background:
          radial-gradient(circle at top left, rgba(95, 138, 255, 0.12), transparent 28%),
          radial-gradient(circle at top right, rgba(0, 216, 255, 0.08), transparent 24%),
          var(--bg);
      }
      a { color: var(--accent); text-decoration: none; }
      a:hover { text-decoration: underline; }
      .layout {
        display: grid;
        grid-template-columns: 280px minmax(0, 1fr);
        min-height: 100vh;
      }
      .sidebar {
        position: sticky;
        top: 0;
        height: 100vh;
        padding: 24px 20px;
        border-right: 1px solid var(--line);
        background: rgba(5, 11, 21, 0.92);
        backdrop-filter: blur(12px);
      }
      .sidebar h1 {
        margin: 0 0 18px;
        font-size: 1.15rem;
      }
      .sidebar .group {
        margin-top: 22px;
      }
      .sidebar .group-title {
        margin-bottom: 10px;
        font-size: 0.77rem;
        font-weight: 700;
        letter-spacing: 0.08em;
        text-transform: uppercase;
        color: #92a7cd;
      }
      .sidebar a {
        display: block;
        padding: 8px 10px;
        border-radius: 10px;
        color: var(--text);
      }
      .sidebar a:hover {
        background: rgba(123, 176, 255, 0.10);
        text-decoration: none;
      }
      .main {
        padding: 40px 40px 72px;
      }
      .hero {
        max-width: 960px;
        padding: 24px 28px;
        border: 1px solid var(--line);
        border-radius: 20px;
        background: linear-gradient(180deg, var(--panel), var(--panel-2));
        box-shadow: var(--shadow);
      }
      .eyebrow {
        display: inline-block;
        margin-bottom: 10px;
        font-size: 0.78rem;
        font-weight: 700;
        letter-spacing: 0.08em;
        text-transform: uppercase;
        color: #93b6ff;
      }
      .hero h2 {
        margin: 0;
        font-size: 2.2rem;
        letter-spacing: -0.03em;
      }
      .hero p {
        margin: 14px 0 0;
        max-width: 760px;
        line-height: 1.75;
        color: var(--muted);
      }
      .actions {
        display: flex;
        gap: 12px;
        flex-wrap: wrap;
        margin-top: 20px;
      }
      .button {
        display: inline-flex;
        align-items: center;
        justify-content: center;
        padding: 10px 16px;
        border-radius: 12px;
        border: 1px solid var(--line);
        background: rgba(123, 176, 255, 0.12);
        color: var(--text);
        font-weight: 600;
      }
      .button.secondary {
        background: rgba(255, 255, 255, 0.04);
      }
      .section-grid {
        display: grid;
        grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
        gap: 18px;
        margin-top: 28px;
        max-width: 1100px;
      }
      .section-card {
        padding: 20px;
        border-radius: 18px;
        border: 1px solid var(--line);
        background: linear-gradient(180deg, rgba(15, 26, 45, 0.94), rgba(9, 16, 29, 0.94));
        box-shadow: var(--shadow);
      }
      .section-card h3 {
        margin: 8px 0 10px;
        font-size: 1.15rem;
      }
      .section-card p {
        margin: 0 0 12px;
        line-height: 1.65;
        color: var(--muted);
      }
      .section-card ul {
        margin: 0;
        padding-left: 18px;
        color: var(--muted);
      }
      .section-card li { margin: 8px 0; }
      @media (max-width: 980px) {
        .layout { grid-template-columns: 1fr; }
        .sidebar {
          position: static;
          height: auto;
          border-right: none;
          border-bottom: 1px solid var(--line);
        }
        .main { padding: 24px 18px 48px; }
      }
    </style>
  </head>
  <body>
    <div class="layout">
      <aside class="sidebar">
        <h1>CollabSphere Docs</h1>
        <div class="group">
          <div class="group-title">Reference</div>
          <a href="/api/v1/docs">Open API Reference</a>
          <a href="/api/v1/openapi.yaml">OpenAPI YAML</a>
        </div>
        <div class="group">
          <div class="group-title">Sections</div>
          {{ range .Sections }}<a href="#{{ .ID }}">{{ .Title }}</a>{{ end }}
        </div>
      </aside>
      <main class="main">
        <section class="hero">
          <span class="eyebrow">Documentation Portal</span>
          <h2>Product docs over the live CollabSphere API</h2>
          <p>This portal is aligned with the live OpenAPI generated by the backend. Use it as the human-oriented entrypoint, and open the API reference when you need exact request and response contracts.</p>
          <div class="actions">
            <a class="button" href="/api/v1/docs">Open API Reference</a>
            <a class="button secondary" href="/api/v1/openapi.json">Open OpenAPI JSON</a>
          </div>
        </section>
        <section class="section-grid">
          {{ range .Sections }}
          <article class="section-card" id="{{ .ID }}">
            <span class="eyebrow">{{ .Eyebrow }}</span>
            <h3>{{ .Title }}</h3>
            <p>{{ .Description }}</p>
            <ul>
              {{ range .Bullets }}<li>{{ . }}</li>{{ end }}
            </ul>
          </article>
          {{ end }}
        </section>
      </main>
    </div>
  </body>
</html>`))

func Register(router chi.Router, title string) {
	data := pageData{
		Title: title + " Docs",
		Sections: []pageSection{
			{
				ID:          "accounts",
				Title:       "Accounts",
				Eyebrow:     "Identity",
				Description: "Sign up, authenticate, inspect the current principal, and manage the current account profile and avatar.",
				Bullets: []string{
					"Create an account and sign in with email/password credentials.",
					"Inspect the authenticated principal and update self profile fields.",
					"Upload an avatar and download account-owned files.",
				},
			},
			{
				ID:          "organizations",
				Title:       "Organizations",
				Eyebrow:     "Business Workspace",
				Description: "Create organizations, manage members, update branding, process onboarding documents, and work with organization-bound files.",
				Bullets: []string{
					"Create an organization with the first owner membership provisioned automatically.",
					"Manage roles, profile fields, logo uploads, and organization file downloads.",
					"Handle cooperation applications, price lists, legal documents, and machine analysis.",
				},
			},
			{
				ID:          "catalog",
				Title:       "Catalog and Imports",
				Eyebrow:     "Organization Catalog",
				Description: "Manage organization-scoped categories and products, and run direct multipart catalog imports.",
				Bullets: []string{
					"Create, update, list, and delete categories and products.",
					"Upload catalog CSV files directly and inspect import batches.",
					"Download the original import source file for an import batch.",
				},
			},
			{
				ID:          "groups",
				Title:       "Groups",
				Eyebrow:     "ACL Containers",
				Description: "Use groups to collect accounts and organizations into shared collaboration spaces.",
				Bullets: []string{
					"Create groups and attach account or organization members.",
					"Use groups as the permission boundary for channels and conferences.",
					"Inspect group membership to understand inherited collaboration access.",
				},
			},
			{
				ID:          "collaboration",
				Title:       "Collaboration",
				Eyebrow:     "Channels and Conferences",
				Description: "Create channels, exchange messages, upload attachments, invite guests, and manage conferences plus recordings.",
				Bullets: []string{
					"Work with channels, messages, reactions, read cursors, and guest invites.",
					"Upload channel attachments and download them by channel context.",
					"Create conferences, inspect transcripts, list recordings, and download them individually.",
				},
			},
		},
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := docsTemplate.Execute(w, data); err != nil {
			http.Error(w, "failed to render docs portal", http.StatusInternalServerError)
		}
	}

	router.Get("/docs", handler)
	router.Get("/docs/", handler)
}
