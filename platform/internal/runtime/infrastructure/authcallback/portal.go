package authcallback

import (
	"html/template"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

type metaEntry struct {
	Label string
	Value string
}

type pageData struct {
	Title             string
	StatusKind        string
	StatusTitle       string
	StatusDescription string
	RecoveryTitle     string
	RecoveryText      string
	Meta              []metaEntry
}

var callbackTemplate = template.Must(template.New("auth-callback").Parse(`<!doctype html>
<html lang="ru">
  <head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>{{ .Title }}</title>
    <style>
      :root {
        color-scheme: dark;
        --bg: #071019;
        --bg-grid: rgba(153, 198, 255, 0.06);
        --panel: rgba(10, 19, 31, 0.94);
        --panel-2: rgba(14, 26, 42, 0.98);
        --text: #f4f8ff;
        --muted: #a6b5c9;
        --line: rgba(151, 181, 220, 0.18);
        --accent: #72d8ff;
        --accent-2: #95ffcf;
        --warn: #ffb870;
        --error: #ff9a9a;
        --shadow: 0 26px 60px rgba(0, 0, 0, 0.42);
      }
      * { box-sizing: border-box; }
      body {
        margin: 0;
        min-height: 100vh;
        color: var(--text);
        font-family: Aptos, "Segoe UI", system-ui, sans-serif;
        background:
          radial-gradient(circle at top left, rgba(114, 216, 255, 0.18), transparent 26%),
          radial-gradient(circle at bottom right, rgba(149, 255, 207, 0.12), transparent 22%),
          linear-gradient(rgba(255,255,255,0.02) 1px, transparent 1px),
          linear-gradient(90deg, rgba(255,255,255,0.02) 1px, transparent 1px),
          var(--bg);
        background-size: auto, auto, 28px 28px, 28px 28px, auto;
        background-position: center, center, center, center, center;
      }
      .shell {
        width: min(1080px, 100%);
        margin: 0 auto;
        padding: 28px;
      }
      .frame {
        display: grid;
        grid-template-columns: minmax(0, 1.35fr) minmax(280px, 0.9fr);
        gap: 18px;
        align-items: start;
      }
      .panel {
        border: 1px solid var(--line);
        border-radius: 26px;
        background: linear-gradient(180deg, var(--panel), var(--panel-2));
        box-shadow: var(--shadow);
        overflow: hidden;
      }
      .hero {
        padding: 28px 30px 18px;
        border-bottom: 1px solid var(--line);
      }
      .eyebrow {
        display: inline-flex;
        align-items: center;
        gap: 8px;
        padding: 7px 11px;
        border: 1px solid rgba(114, 216, 255, 0.22);
        border-radius: 999px;
        color: var(--accent);
        background: rgba(114, 216, 255, 0.08);
        font-size: 0.76rem;
        letter-spacing: 0.08em;
        text-transform: uppercase;
        font-weight: 700;
      }
      h1 {
        margin: 16px 0 10px;
        font-family: Georgia, "Times New Roman", serif;
        font-size: clamp(2rem, 4vw, 3.1rem);
        letter-spacing: -0.04em;
        line-height: 0.98;
      }
      .subtitle {
        margin: 0;
        max-width: 58ch;
        color: var(--muted);
        line-height: 1.72;
        font-size: 1rem;
      }
      .body {
        padding: 24px 30px 30px;
      }
      .status {
        position: relative;
        padding: 18px 18px 18px 20px;
        border: 1px solid var(--line);
        border-radius: 18px;
        background: rgba(255,255,255,0.03);
        overflow: hidden;
      }
      .status::before {
        content: "";
        position: absolute;
        inset: 0 auto 0 0;
        width: 4px;
        background: rgba(114, 216, 255, 0.88);
      }
      .status.ok {
        border-color: rgba(149, 255, 207, 0.24);
        background: rgba(149, 255, 207, 0.07);
      }
      .status.ok::before {
        background: var(--accent-2);
      }
      .status.error {
        border-color: rgba(255, 154, 154, 0.25);
        background: rgba(255, 154, 154, 0.08);
      }
      .status.error::before {
        background: var(--error);
      }
      .status strong {
        display: block;
        margin-bottom: 8px;
        font-size: 1.02rem;
      }
      .status p {
        margin: 0;
        color: var(--muted);
        line-height: 1.68;
      }
      .recovery {
        margin-top: 14px;
        padding: 16px 18px;
        border: 1px solid rgba(255, 184, 112, 0.22);
        border-radius: 16px;
        background: rgba(255, 184, 112, 0.08);
      }
      .recovery strong {
        display: block;
        margin-bottom: 8px;
        color: #ffd9a6;
      }
      .recovery p {
        margin: 0;
        color: #f3d7b3;
        line-height: 1.68;
      }
      .meta {
        margin-top: 16px;
        padding: 16px 18px;
        border: 1px solid var(--line);
        border-radius: 16px;
        background: rgba(255,255,255,0.025);
      }
      .meta dl {
        margin: 0;
        display: grid;
        grid-template-columns: 180px 1fr;
        gap: 12px 14px;
      }
      .meta dt {
        color: var(--muted);
      }
      .meta dd {
        margin: 0;
        word-break: break-word;
      }
      .actions {
        display: flex;
        flex-wrap: wrap;
        gap: 12px;
        margin-top: 18px;
      }
      .button,
      button {
        appearance: none;
        border: 1px solid rgba(114, 216, 255, 0.2);
        border-radius: 14px;
        padding: 12px 16px;
        background: linear-gradient(135deg, rgba(114, 216, 255, 0.18), rgba(114, 216, 255, 0.08));
        color: var(--text);
        font: inherit;
        font-weight: 700;
        cursor: pointer;
        text-decoration: none;
        transition: transform 140ms ease, border-color 140ms ease, filter 140ms ease;
      }
      .button.secondary,
      button.secondary {
        border-color: var(--line);
        background: rgba(255,255,255,0.04);
        font-weight: 600;
      }
      .button:hover,
      button:hover {
        transform: translateY(-1px);
        filter: brightness(1.06);
      }
      .footnote {
        margin-top: 18px;
        color: var(--muted);
        line-height: 1.65;
        font-size: 0.95rem;
      }
      .aside {
        padding: 24px;
      }
      .aside h2 {
        margin: 0 0 14px;
        font-size: 1.1rem;
      }
      .steps {
        display: grid;
        gap: 12px;
      }
      .step {
        padding: 14px 14px 14px 16px;
        border: 1px solid var(--line);
        border-radius: 16px;
        background: rgba(255,255,255,0.025);
      }
      .step strong {
        display: block;
        margin-bottom: 6px;
      }
      .step p {
        margin: 0;
        color: var(--muted);
        line-height: 1.6;
      }
      code {
        padding: 2px 6px;
        border-radius: 8px;
        background: rgba(255,255,255,0.08);
        font-family: "Cascadia Code", "JetBrains Mono", monospace;
        font-size: 0.92em;
      }
      @media (max-width: 900px) {
        .frame {
          grid-template-columns: 1fr;
        }
      }
      @media (max-width: 680px) {
        .shell {
          padding: 16px;
        }
        .hero,
        .body,
        .aside {
          padding-left: 18px;
          padding-right: 18px;
        }
        .meta dl {
          grid-template-columns: 1fr;
        }
      }
    </style>
  </head>
  <body>
    <div class="shell">
      <main class="frame">
        <section class="panel">
          <header class="hero">
            <span class="eyebrow">CollabSphere Auth</span>
            <h1>Browser auth callback</h1>
            <p class="subtitle">Локальная dev-страница для входа через ZITADEL. Она принимает <code>ticket</code> после callback, меняет его на backend-токены и сохраняет результат в <code>localStorage</code>.</p>
          </header>
          <div class="body">
            <div id="status" class="status{{ if .StatusKind }} {{ .StatusKind }}{{ end }}">
              <strong>{{ .StatusTitle }}</strong>
              <p>{{ .StatusDescription }}</p>
            </div>
            {{ if .RecoveryText }}
            <div class="recovery">
              <strong>{{ .RecoveryTitle }}</strong>
              <p>{{ .RecoveryText }}</p>
            </div>
            {{ end }}
            <div id="meta" class="meta"{{ if not .Meta }} hidden{{ end }}>
              <dl id="meta-list">
                {{ range .Meta }}
                <dt>{{ .Label }}</dt>
                <dd>{{ .Value }}</dd>
                {{ end }}
              </dl>
            </div>
            <div class="actions">
              <a class="button" href="/v1/auth/zitadel/login?return_to=/auth/callback">Войти через ZITADEL</a>
              <a class="button secondary" href="/v1/auth/zitadel/signup?return_to=/auth/callback">Зарегистрироваться</a>
              <button id="me" class="secondary" type="button">Проверить <code>/auth/me</code></button>
              <button id="clear" class="secondary" type="button">Очистить токены</button>
            </div>
            <p class="footnote">Используемые ключи: <code>collabsphere.auth</code>, <code>collabsphere.accessToken</code>, <code>collabsphere.refreshToken</code>.</p>
          </div>
        </section>
        <aside class="panel aside">
          <h2>Как должен идти поток</h2>
          <div class="steps">
            <div class="step">
              <strong>1. Старт</strong>
              <p>Кнопка входа отправляет браузер в <code>/v1/auth/zitadel/login</code> и дальше в hosted UI ZITADEL.</p>
            </div>
            <div class="step">
              <strong>2. Callback</strong>
              <p>ZITADEL возвращает пользователя в backend на <code>/v1/auth/zitadel/callback</code>, а backend редиректит обратно сюда.</p>
            </div>
            <div class="step">
              <strong>3. Exchange</strong>
              <p>Если в URL есть <code>ticket</code>, страница автоматически вызывает <code>/v1/auth/exchange</code> и сохраняет локальные токены.</p>
            </div>
            <div class="step">
              <strong>4. Проверка</strong>
              <p>Кнопка <code>/auth/me</code> подтверждает, что backend уже видит вашу локальную сессию.</p>
            </div>
          </div>
        </aside>
      </main>
    </div>
    <script>
      const statusEl = document.getElementById('status');
      const metaEl = document.getElementById('meta');
      const metaListEl = document.getElementById('meta-list');
      const meButton = document.getElementById('me');
      const clearButton = document.getElementById('clear');
      const params = new URLSearchParams(window.location.search);

      function escapeHtml(value) {
        return String(value)
          .replaceAll('&', '&amp;')
          .replaceAll('<', '&lt;')
          .replaceAll('>', '&gt;')
          .replaceAll('"', '&quot;')
          .replaceAll("'", '&#39;');
      }

      function setStatus(kind, title, description) {
        statusEl.className = 'status' + (kind ? ' ' + kind : '');
        statusEl.innerHTML = '<strong>' + escapeHtml(title) + '</strong><p>' + escapeHtml(description) + '</p>';
      }

      function setMeta(entries) {
        metaListEl.innerHTML = '';
        if (!entries || entries.length === 0) {
          metaEl.hidden = true;
          return;
        }
        metaEl.hidden = false;
        for (const [label, value] of entries) {
          const dt = document.createElement('dt');
          dt.textContent = label;
          const dd = document.createElement('dd');
          dd.textContent = value;
          metaListEl.appendChild(dt);
          metaListEl.appendChild(dd);
        }
      }

      function storeAuth(payload) {
        localStorage.setItem('collabsphere.auth', JSON.stringify(payload));
        localStorage.setItem('collabsphere.accessToken', payload.accessToken);
        localStorage.setItem('collabsphere.refreshToken', payload.refreshToken);
      }

      function clearAuth() {
        localStorage.removeItem('collabsphere.auth');
        localStorage.removeItem('collabsphere.accessToken');
        localStorage.removeItem('collabsphere.refreshToken');
      }

      async function exchangeTicket(ticket) {
        const response = await fetch('/v1/auth/exchange', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ ticket }),
        });
        const text = await response.text();
        let payload = null;
        try {
          payload = text ? JSON.parse(text) : null;
        } catch (_) {
          payload = null;
        }
        if (!response.ok) {
          const message = payload && (payload.detail || payload.title || payload.error) ? (payload.detail || payload.title || payload.error) : ('HTTP ' + response.status);
          throw new Error(message);
        }
        return payload;
      }

      async function loadMe() {
        const token = localStorage.getItem('collabsphere.accessToken');
        if (!token) {
          setStatus('error', 'Токены не найдены', 'Сначала завершите вход через ZITADEL, чтобы получить локальный access token.');
          setMeta([]);
          return;
        }
        setStatus('', 'Проверяем текущего пользователя', 'Запрашиваем /v1/auth/me с сохранённым access token.');
        const response = await fetch('/v1/auth/me', {
          headers: { Authorization: 'Bearer ' + token }
        });
        const text = await response.text();
        let payload = null;
        try {
          payload = text ? JSON.parse(text) : null;
        } catch (_) {
          payload = null;
        }
        if (!response.ok) {
          const message = payload && (payload.detail || payload.title || payload.error) ? (payload.detail || payload.title || payload.error) : ('HTTP ' + response.status);
          throw new Error(message);
        }
        setStatus('ok', 'Сессия активна', 'Локальный backend token работает, /auth/me ответил успешно.');
        setMeta([
          ['Account ID', payload.id || ''],
          ['Email', payload.email || ''],
          ['Display Name', payload.displayName || ''],
          ['Active', String(payload.isActive)],
        ]);
      }

      async function boot() {
        clearButton.addEventListener('click', () => {
          clearAuth();
          setStatus('', 'Токены очищены', 'Сохранённые access/refresh токены удалены из localStorage.');
          setMeta([]);
          history.replaceState({}, '', '/auth/callback');
        });

        meButton.addEventListener('click', async () => {
          try {
            await loadMe();
          } catch (error) {
            setStatus('error', 'Проверка /auth/me не удалась', error instanceof Error ? error.message : 'Unknown error');
          }
        });

        const errorCode = params.get('error');
        if (errorCode) {
          setStatus('error', 'Вход завершился ошибкой', params.get('error_description') || errorCode);
          setMeta([['Error', errorCode]]);
          return;
        }

        const ticket = params.get('ticket');
        if (!ticket) {
          const stored = localStorage.getItem('collabsphere.auth');
          if (stored) {
            setStatus('ok', 'Токены уже сохранены', 'Можно проверить текущего пользователя через /auth/me или повторно запустить вход.');
          }
          return;
        }

        setStatus('', 'Завершаем вход', 'Обмениваем одноразовый ticket на локальные access и refresh токены.');
        setMeta([['Ticket', 'получен, выполняется exchange']]);
        try {
          const payload = await exchangeTicket(ticket);
          storeAuth(payload);
          history.replaceState({}, '', '/auth/callback');
          setStatus('ok', 'Вход выполнен', 'Токены сохранены в localStorage. Теперь можно вызывать backend API от имени текущего пользователя.');
          setMeta([
            ['Provider', payload.provider || ''],
            ['Intent', payload.intent || ''],
            ['Is New Account', String(Boolean(payload.isNewAccount))],
            ['Token Type', payload.tokenType || ''],
            ['Expires In', String(payload.expiresIn || '')],
          ]);
        } catch (error) {
          setStatus('error', 'Обмен ticket не удался', error instanceof Error ? error.message : 'Unknown error');
          setMeta([]);
        }
      }

      boot();
    </script>
  </body>
</html>`))

func defaultPageData(title string) pageData {
	return pageData{
		Title:             title + " Auth Callback",
		StatusTitle:       "Готово к входу",
		StatusDescription: "Запустите вход или регистрацию через ZITADEL. После возврата страница сама закончит обмен ticket на локальные токены.",
	}
}

func buildPageData(title string, r *http.Request) pageData {
	data := defaultPageData(title)
	query := r.URL.Query()

	if errorCode := query.Get("error"); errorCode != "" {
		description := query.Get("error_description")
		if description == "" {
			description = errorCode
		}
		data.StatusKind = "error"
		data.StatusTitle = "Вход завершился ошибкой"
		data.StatusDescription = description
		data.Meta = []metaEntry{{Label: "Error", Value: errorCode}}

		lowerDescription := strings.ToLower(description)
		if strings.Contains(lowerDescription, "verified email is required for first external login") {
			data.RecoveryTitle = "Почему это произошло"
			data.RecoveryText = "Backend допускает первый вход через внешний OIDC только для пользователей с подтверждённым email. Подтвердите email в ZITADEL или выполните platform force-verify, затем повторите вход."
		}
		return data
	}

	if ticket := query.Get("ticket"); ticket != "" {
		data.StatusTitle = "Завершаем вход"
		data.StatusDescription = "Одноразовый ticket уже получен. Если JavaScript включён, страница сейчас обменяет его на локальные access и refresh токены backend."
		data.Meta = []metaEntry{{Label: "Ticket", Value: "получен"}}
	}

	return data
}

func Register(router chi.Router, title string) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := callbackTemplate.Execute(w, buildPageData(title, r)); err != nil {
			http.Error(w, "failed to render auth callback", http.StatusInternalServerError)
		}
	}

	router.Get("/auth/callback", handler)
	router.Get("/auth/callback/", handler)
}
