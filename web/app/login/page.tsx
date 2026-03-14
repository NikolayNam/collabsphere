"use client";

import { FormEvent, useState } from "react";

import { Panel } from "@/components/panel";
import { beginBrowserLogin } from "@/lib/auth";

type LoginState =
  | { kind: "idle"; title: string; description: string }
  | { kind: "working"; title: string; description: string }
  | { kind: "success"; title: string; description: string }
  | { kind: "error"; title: string; description: string };

const initialState: LoginState = {
  kind: "idle",
  title: "Выберите путь входа",
  description: "Основной путь — browser login через ZITADEL с custom login UI из web-login.",
};

export default function LoginPage() {
  const [state, setState] = useState<LoginState>(initialState);

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
      </section>
    </>
  );
}
