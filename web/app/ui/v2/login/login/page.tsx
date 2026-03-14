import Link from "next/link";

import { Panel } from "@/components/panel";
import { getAuthRequest } from "@/lib/login/zitadel";

type SearchParams = Promise<Record<string, string | string[] | undefined>>;

function readParam(params: Record<string, string | string[] | undefined>, ...names: string[]): string {
  for (const name of names) {
    const value = params[name];
    if (typeof value === "string" && value.trim()) {
      return value.trim();
    }
  }
  return "";
}

function hiddenAuthRequestFields(authRequestId: string, loginHint: string) {
  return (
    <>
      {authRequestId ? <input type="hidden" name="authRequest" value={authRequestId} /> : null}
      {loginHint ? <input type="hidden" name="loginHint" value={loginHint} /> : null}
    </>
  );
}

export default async function ZitadelLoginPage({ searchParams }: { searchParams: SearchParams }) {
  const params = await searchParams;
  const authRequestId = readParam(params, "authRequest", "requestId");
  const mode = readParam(params, "mode");
  const loginHintFromQuery = readParam(params, "loginHint");
  const message = readParam(params, "message");
  const error = readParam(params, "error", "error_description");

  let prompt = "";
  let loginHint = loginHintFromQuery;
  if (authRequestId) {
    try {
      const authRequest = await getAuthRequest(authRequestId);
      prompt = authRequest.prompt || "";
      if (!loginHint) {
        loginHint = authRequest.loginHint || "";
      }
    } catch (cause) {
      console.error("load auth request", cause);
    }
  }

  const signupMode = mode === "signup" || prompt === "create";
  const baseQuery = authRequestId ? `authRequest=${encodeURIComponent(authRequestId)}` : "";
  const loginHintQuery = loginHint ? `loginHint=${encodeURIComponent(loginHint)}` : "";
  const joinQuery = [baseQuery, loginHintQuery].filter(Boolean).join("&");
  const loginHref = joinQuery ? `/ui/v2/login/login?${joinQuery}` : "/ui/v2/login/login";
  const signupHref = joinQuery ? `/ui/v2/login/login?mode=signup&${joinQuery}` : "/ui/v2/login/login?mode=signup";
  const verifyHref = joinQuery ? `/ui/v2/login/verify?${joinQuery}` : "/ui/v2/login/verify";

  return (
    <>
      <Panel
        title={signupMode ? "Создание ZITADEL account" : "ZITADEL login"}
        eyebrow={signupMode ? "Browser signup" : "Browser login"}
        actions={
          <div className="button-row">
            <Link className={`button-link ${signupMode ? "secondary" : "primary"}`} href={loginHref}>
              Login
            </Link>
            <Link className={`button-link ${signupMode ? "primary" : "secondary"}`} href={signupHref}>
              Signup
            </Link>
            <Link className="button-link secondary" href={verifyHref}>
              Verify email
            </Link>
          </div>
        }
      >
        <div className={`status-card ${error ? "error" : message ? "success" : "info"}`}>
          <strong>{error ? "Flow требует вмешательства" : message ? "Flow продолжен" : "Login UI обслуживается из web/"}</strong>
          <p className="status-copy">
            {error ||
              message ||
              "CollabSphere заменяет отдельный zitadel-login контейнер собственным login-контуром внутри Next.js web-приложения."}
          </p>
        </div>
        {authRequestId ? (
          <div className="mini-card">
            <h3>OIDC request</h3>
            <p className="muted">
              Текущий auth request: <code>{authRequestId}</code>
            </p>
          </div>
        ) : null}
      </Panel>

      <section className="split">
        <Panel title={signupMode ? "Signup form" : "Login form"} eyebrow="Username + password">
          {signupMode ? (
            <form className="form-grid" action="/ui/v2/login/actions/signup" method="post">
              {hiddenAuthRequestFields(authRequestId, loginHint)}
              <div className="form-row">
                <label className="form-label" htmlFor="displayName">
                  Display name
                </label>
                <input id="displayName" className="text-input" name="displayName" placeholder="Smoke User" required type="text" />
              </div>
              <div className="form-row">
                <label className="form-label" htmlFor="signupEmail">
                  Email
                </label>
                <input
                  id="signupEmail"
                  className="text-input"
                  autoComplete="email"
                  defaultValue={loginHint}
                  name="email"
                  placeholder="user@example.com"
                  required
                  type="email"
                />
              </div>
              <div className="form-row">
                <label className="form-label" htmlFor="signupPassword">
                  Password
                </label>
                <input
                  id="signupPassword"
                  className="text-input"
                  autoComplete="new-password"
                  name="password"
                  placeholder="Choose a password"
                  required
                  type="password"
                />
              </div>
              <div className="button-row">
                <button className="button primary" type="submit">
                  Create account
                </button>
              </div>
            </form>
          ) : (
            <form className="form-grid" action="/ui/v2/login/actions/login" method="post">
              {hiddenAuthRequestFields(authRequestId, loginHint)}
              <div className="form-row">
                <label className="form-label" htmlFor="loginName">
                  Email or username
                </label>
                <input
                  id="loginName"
                  className="text-input"
                  autoComplete="username"
                  defaultValue={loginHint}
                  name="loginName"
                  placeholder="user@example.com"
                  required
                  type="text"
                />
              </div>
              <div className="form-row">
                <label className="form-label" htmlFor="password">
                  Password
                </label>
                <input
                  id="password"
                  className="text-input"
                  autoComplete="current-password"
                  name="password"
                  placeholder="Enter your password"
                  required
                  type="password"
                />
              </div>
              <div className="button-row">
                <button className="button primary" type="submit">
                  Continue
                </button>
                <Link
                  className="button-link secondary"
                  href={`/ui/v2/login/password-reset${authRequestId ? `?authRequest=${encodeURIComponent(authRequestId)}&loginHint=${encodeURIComponent(loginHint)}` : ""}`}
                >
                  Forgot password
                </Link>
              </div>
            </form>
          )}
        </Panel>

        <Panel title="Flow notes" eyebrow="Current implementation">
          <div className="cards">
            <div className="mini-card">
              <h3>Backend contract unchanged</h3>
              <p className="muted">
                После удачного login/signup ZITADEL callback всё так же уходит в backend `/v1/auth/zitadel/callback`, а локальные токены выпускаются через `/v1/auth/exchange`.
              </p>
            </div>
            <div className="mini-card">
              <h3>Email verification</h3>
              <p className="muted">
                Signup создаёт нового пользователя с email verification code и переводит его в экран активации, не меняя backend-side `email_verified` policy.
              </p>
            </div>
            <div className="mini-card">
              <h3>Password reset</h3>
              <p className="muted">Смена пароля поддерживается отдельным экраном через verification code flow ZITADEL User API.</p>
            </div>
            <div className="mini-card">
              <h3>Advanced factors</h3>
              <p className="muted">
                OIDC proxy и login origin уже вынесены в `web-login`, поэтому дальше сюда можно безопасно встраивать MFA, passkeys и IDP-specific steps без смены CollabSphere auth surface.
              </p>
            </div>
          </div>
        </Panel>
      </section>
    </>
  );
}
