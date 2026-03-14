"use client";

import { FormEvent, useState } from "react";

import { Panel } from "@/components/panel";
import { APIError } from "@/lib/api";
import { beginBrowserLogin, loginWithPassword, storeTokens } from "@/lib/auth";

type LoginState =
  | { kind: "idle"; title: string; description: string }
  | { kind: "working"; title: string; description: string }
  | { kind: "success"; title: string; description: string }
  | { kind: "error"; title: string; description: string };

const initialState: LoginState = {
  kind: "idle",
  title: "Выберите путь входа",
  description: "Основной путь — browser login через ZITADEL с custom login UI из web-login. Legacy password login оставлен как dev/fallback и зависит от backend feature flags.",
};

export default function LoginPage() {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [state, setState] = useState<LoginState>(initialState);

  async function handleLegacyLogin(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setState({
      kind: "working",
      title: "Входим через backend password login",
      description: "Если backend fallback отключён, вы увидите честную ошибку feature-flag policy.",
    });

    try {
      const tokens = await loginWithPassword(email, password);
      storeTokens(tokens);
      setState({
        kind: "success",
        title: "Legacy login завершён",
        description: "Токены сохранены в localStorage. Теперь можно открыть /me или /organizations.",
      });
    } catch (error) {
      const message =
        error instanceof APIError
          ? `${error.message}${error.code ? ` (${error.code})` : ""}`
          : error instanceof Error
            ? error.message
            : "Unknown login error";
      setState({
        kind: "error",
        title: "Legacy login не удался",
        description: message,
      });
    }
  }

  return (
    <>
      <Panel
        title="Authentication entry"
        eyebrow="Browser-first auth"
        actions={
          <button className="button primary" onClick={() => beginBrowserLogin("login")} type="button">
            Войти через ZITADEL
          </button>
        }
      >
        <div className={`status-card ${state.kind === "error" ? "error" : state.kind === "success" ? "success" : "info"}`}>
          <strong>{state.title}</strong>
          <p className="status-copy">{state.description}</p>
        </div>
      </Panel>

      <section className="split">
        <Panel title="Основной путь" eyebrow="ZITADEL browser flow">
          <div className="mini-card">
            <h3>Что произойдёт</h3>
            <p className="muted">
              Frontend отправит браузер в backend `GET /v1/auth/zitadel/login`, backend
              отдаст control flow в ZITADEL authorization flow, затем пользователь пройдёт через
              custom login UI на `auth.localhost:3000` и вернётся обратно в
              frontend callback с `ticket`.
            </p>
          </div>
          <div className="button-row">
            <button className="button primary" onClick={() => beginBrowserLogin("login")} type="button">
              Login
            </button>
            <button className="button secondary" onClick={() => beginBrowserLogin("signup")} type="button">
              Signup
            </button>
          </div>
        </Panel>

        <Panel title="Fallback путь" eyebrow="Legacy password login">
          <form className="form-grid" onSubmit={handleLegacyLogin}>
            <div className="form-row">
              <label className="form-label" htmlFor="email">
                Email
              </label>
              <input
                id="email"
                className="text-input"
                autoComplete="email"
                value={email}
                onChange={(event) => setEmail(event.target.value)}
                placeholder="owner@example.com"
                type="email"
                required
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
                value={password}
                onChange={(event) => setPassword(event.target.value)}
                placeholder="Enter your password"
                type="password"
                required
              />
              <p className="field-help">Этот путь сработает только если backend разрешает `AUTH_PASSWORD_LOGIN_ENABLED=true`.</p>
            </div>
            <div className="button-row">
              <button className="button secondary" type="submit">
                Войти по email/password
              </button>
            </div>
          </form>
        </Panel>
      </section>
    </>
  );
}
