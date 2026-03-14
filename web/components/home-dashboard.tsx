"use client";

import Link from "next/link";
import { useEffect, useState } from "react";

import { Panel } from "@/components/panel";
import { APIError } from "@/lib/api";
import { clearTokens, loadMe, logout, readStoredTokens, type MeResponse, type TokenBundle } from "@/lib/auth";
import { getAPIBaseURL, getAppBaseURL } from "@/lib/env";

type SessionView =
  | { kind: "loading"; message: string; tokens: TokenBundle | null; profile: MeResponse | null }
  | { kind: "guest"; message: string; tokens: TokenBundle | null; profile: MeResponse | null }
  | { kind: "ready"; message: string; tokens: TokenBundle; profile: MeResponse }
  | { kind: "error"; message: string; tokens: TokenBundle | null; profile: MeResponse | null };

const initialState: SessionView = {
  kind: "loading",
  message: "Проверяем локальную сессию и готовим домашнюю страницу.",
  tokens: null,
  profile: null,
};

export function HomeDashboard() {
  const [session, setSession] = useState<SessionView>(initialState);
  const apiBaseURL = getAPIBaseURL();
  const appBaseURL = getAppBaseURL();

  async function hydrateSession() {
    const stored = readStoredTokens();
    if (!stored?.accessToken) {
      setSession({
        kind: "guest",
        message: "Локальная сессия не найдена. Можно начать с browser login или signup.",
        tokens: null,
        profile: null,
      });
      return;
    }

    setSession({
      kind: "loading",
      message: "Запрашиваем backend principal через /v1/auth/me.",
      tokens: stored,
      profile: null,
    });

    try {
      const profile = await loadMe(stored.accessToken);
      setSession({
        kind: "ready",
        message: "Сессия активна, домашняя страница уже знает ваш backend principal.",
        tokens: stored,
        profile,
      });
    } catch (error) {
      const message =
        error instanceof APIError
          ? `${error.message}${error.code ? ` (${error.code})` : ""}`
          : error instanceof Error
            ? error.message
            : "Не удалось загрузить профиль";

      setSession({
        kind: "error",
        message,
        tokens: stored,
        profile: null,
      });
    }
  }

  async function handleLogout() {
    const stored = readStoredTokens();
    try {
      if (stored?.refreshToken) {
        await logout(stored.refreshToken);
      }
    } catch {
      // ignore transport errors; local cleanup still matters
    } finally {
      clearTokens();
      setSession({
        kind: "guest",
        message: "Сессия очищена. Можно войти снова или открыть signup.",
        tokens: null,
        profile: null,
      });
    }
  }

  useEffect(() => {
    void hydrateSession();
  }, []);

  const isAuthenticated = session.kind === "ready";
  const profile = session.profile;
  const tokens = session.tokens;
  const heroName = isAuthenticated ? profile?.displayName || profile?.email || "внутри платформы" : "";

  return (
    <>
      <section className="hero hero-home">
        <div className="panel hero-copy home-hero-copy">
          <p className="kicker">{isAuthenticated ? "Session live" : "CollabSphere Web"}</p>
          <h2 className="hero-title">
            {isAuthenticated
              ? `Добро пожаловать, ${heroName}.`
              : "Главная страница теперь может быть первой точкой входа после авторизации."}
          </h2>
          <p className="hero-text">
            {isAuthenticated
              ? "Browser login завершён, локальные токены выпущены backend’ом, и дальше можно уже двигаться в organizations, профиль и следующий продуктовый UI."
              : "Это не просто экран со ссылками. Главная уже понимает, есть ли локальная сессия, умеет показать backend principal и остаётся поверх существующего Go API без второго backend-контракта."}
          </p>
          <div className="hero-actions">
            {isAuthenticated ? (
              <>
                <Link href="/organizations" className="button-link primary">
                  Открыть organizations
                </Link>
                <Link href="/chat" className="button-link secondary">
                  Открыть chat
                </Link>
                <Link href="/me" className="button-link secondary">
                  Профиль и токены
                </Link>
                <button className="button secondary" onClick={() => void hydrateSession()} type="button">
                  Обновить статус
                </button>
                <button className="button danger" onClick={() => void handleLogout()} type="button">
                  Выйти
                </button>
              </>
            ) : (
              <>
                <Link href="/login" className="button-link primary">
                  Войти через ZITADEL
                </Link>
                <Link href="/login" className="button-link secondary">
                  Открыть login shell
                </Link>
                <a href={`${apiBaseURL}/v1/docs`} className="button-link secondary" target="_blank" rel="noreferrer">
                  API docs
                </a>
              </>
            )}
          </div>
        </div>

        <Panel title="Session overview" eyebrow="Current browser state">
          <div
            className={`status-card ${
              session.kind === "ready" ? "success" : session.kind === "error" ? "error" : "info"
            }`}
          >
            <strong>
              {session.kind === "ready"
                ? "Авторизация подтверждена"
                : session.kind === "error"
                  ? "Сессия требует внимания"
                  : session.kind === "guest"
                    ? "Гостевой режим"
                    : "Проверяем сессию"}
            </strong>
            <p className="status-copy">{session.message}</p>
          </div>

          <div className="cards home-mini-grid">
            <div className="mini-card">
              <h3>Frontend origin</h3>
              <p className="muted">{appBaseURL}</p>
            </div>
            <div className="mini-card">
              <h3>Backend API</h3>
              <p className="muted">{apiBaseURL}</p>
            </div>
            <div className="mini-card">
              <h3>Login origin</h3>
              <p className="muted">auth.localhost:3000 остаётся отдельным login surface.</p>
            </div>
            <div className="mini-card">
              <h3>Current user</h3>
              <p className="muted">{profile?.email || "Гость без локальной сессии"}</p>
            </div>
          </div>
        </Panel>
      </section>

      <section className="home-grid">
        <Panel title="Что делать дальше" eyebrow="Immediate paths">
          <div className="feature-grid">
            <div className="feature-card">
              <strong>Organizations</strong>
              <p className="muted">
                Создать organization, открыть её профиль, обновить основные поля, загрузить logo и проверить resolve-by-host.
              </p>
              <Link href="/organizations" className="button-link secondary">
                Перейти в organizations
              </Link>
            </div>
            <div className="feature-card">
              <strong>Profile</strong>
              <p className="muted">
                Посмотреть текущий backend principal, refresh/logout и убедиться, что локальная сессия жива.
              </p>
              <Link href="/me" className="button-link secondary">
                Открыть /me
              </Link>
            </div>
            <div className="feature-card">
              <strong>Chat</strong>
              <p className="muted">
                Channels и messages уже существуют в backend. Новый экран выводит их в продуктовый UI без отдельного chat-backend.
              </p>
              <Link href="/chat" className="button-link secondary">
                Открыть chat
              </Link>
            </div>
          </div>
        </Panel>

        <Panel title="Главная теперь знает контекст" eyebrow="Guest vs signed-in states">
          <div className="cards home-mini-grid">
            <div className="mini-card">
              <h3>После login</h3>
              <p className="muted">Frontend callback теперь возвращает пользователя на `/`, а не на техническую страницу `/me`.</p>
            </div>
            <div className="mini-card">
              <h3>Когда токены есть</h3>
              <p className="muted">
                Главная запрашивает <code>/v1/auth/me</code> и превращается в домашний экран, а не в маркетинговую заглушку.
              </p>
            </div>
            <div className="mini-card">
              <h3>Когда токенов нет</h3>
              <p className="muted">Показываем понятный стартовый сценарий: login, signup, docs и topology без ложного ощущения “вы уже внутри”.</p>
            </div>
            <div className="mini-card">
              <h3>Auth boundary сохранена</h3>
              <p className="muted">`auth.localhost:3000` остаётся login-origin. Продуктовая главная живёт на `collabsphere.localhost:3002`.</p>
            </div>
          </div>
        </Panel>
      </section>

      {isAuthenticated && profile && tokens ? (
        <section className="stat-grid">
          <div className="stat">
            <strong>{profile.email}</strong>
            <p className="muted">Подтверждённый backend principal уже подхвачен домашней страницей.</p>
          </div>
          <div className="stat">
            <strong>{tokens.provider || "local tokens"}</strong>
            <p className="muted">Локальная сессия уже выпущена через backend exchange и готова для следующих экранов.</p>
          </div>
          <div className="stat">
            <strong>{tokens.expiresIn || 0}s</strong>
            <p className="muted">TTL access token из последнего exchange/refresh виден уже на главной без захода в debug-only экран.</p>
          </div>
        </section>
      ) : (
        <section className="cards">
          <Panel title="Почему не `localhost:3000`" eyebrow="Routing boundary">
            <p className="muted">
              `3000` уже используется как отдельный login-origin. Это хорошо: login UI, OIDC proxy и product homepage не смешиваются в одном surface.
            </p>
          </Panel>
          <Panel title="Что уже реально работает" eyebrow="Current web shell">
            <ul className="list">
              <li>Browser login через ZITADEL и self-hosted login UI.</li>
              <li>
                <code>ticket -&gt; /v1/auth/exchange -&gt; local tokens</code>.
              </li>
              <li>Домашняя страница, которая меняется после входа.</li>
              <li>Organizations workbench и профиль текущего пользователя.</li>
            </ul>
          </Panel>
        </section>
      )}
    </>
  );
}
