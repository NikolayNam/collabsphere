"use client";

import Link from "next/link";
import { useSearchParams } from "next/navigation";
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
  message: "Проверяем локальную сессию и готовим стартовый экран.",
  tokens: null,
  profile: null,
};

export function HomeDashboard() {
  const [session, setSession] = useState<SessionView>(initialState);
  const searchParams = useSearchParams();
  const apiBaseURL = getAPIBaseURL();
  const appBaseURL = getAppBaseURL();

  async function hydrateSession() {
    const stored = readStoredTokens();
    if (!stored?.accessToken) {
      setSession({
        kind: "guest",
        message: "Сессия не найдена. Начните с входа через ZITADEL или создайте аккаунт.",
        tokens: null,
        profile: null,
      });
      return;
    }

    setSession({
      kind: "loading",
      message: "Проверяем пользователя через /v1/auth/me.",
      tokens: stored,
      profile: null,
    });

    try {
      const profile = await loadMe(stored.accessToken);
      setSession({
        kind: "ready",
        message: "Сессия активна: главная уже знает ваш профиль и готова к работе.",
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
        message: "Вы вышли из аккаунта. Можно войти снова в любой момент.",
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
  const signupCreated = searchParams.get("signup") === "created" || Boolean(searchParams.get("verificationCode"));

  return (
    <>
      <section className="hero hero-home">
        <div className="panel hero-copy home-hero-copy">
          {signupCreated ? (
            <div className="status-card success">
              <strong>Аккаунт создан</strong>
              <p className="status-copy">Регистрация завершена. Проверьте email и подтвердите адрес перед первым входом.</p>
              <div className="hero-actions">
                <Link href="/ui/v2/login/verify" className="button-link secondary">
                  Перейти к Verify email
                </Link>
              </div>
            </div>
          ) : null}
          <p className="kicker">{isAuthenticated ? "Session live" : "CollabSphere Web"}</p>
          <h2 className="hero-title">
            {isAuthenticated
              ? `Добро пожаловать, ${heroName}.`
              : "Главная страница может быть вашим первым экраном после авторизации."}
          </h2>
          <p className="hero-text">
            {isAuthenticated
              ? "Вход завершён, локальные токены выпущены backend-ом. Дальше можно сразу переходить в организации, профиль и рабочие экраны."
              : "Это главная страница нашей платформы CollabSphere: она понимает состояние сессии, показывает backend principal и помогает быстро стартовать."}
          </p>
          <img className="hero-illustration" src="/illustrations/orbit.svg" alt="Схема совместной работы в CollabSphere" />
          <div className="hero-actions">
            {isAuthenticated ? (
              <>
                <Link href="/organizations" className="button-link primary">
                  Открыть organizations
                </Link>
                <Link href="/chat" className="button-link secondary">
                  Открыть chat
                </Link>
                <Link href="/offers-requests" className="button-link secondary">
                  Заявки и предложения
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
                <Link href="/ui/v2/login/verify" className="button-link secondary">
                  Verify email
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
              <img className="feature-illustration" src="/illustrations/workspace.svg" alt="Рабочее пространство организации" />
              <strong>Organizations</strong>
              <p className="muted">
                Создайте organization, обновите профиль, загрузите logo и проверьте resolve-by-host в одном рабочем потоке.
              </p>
              <Link href="/organizations" className="button-link secondary">
                Перейти в organizations
              </Link>
            </div>
            <div className="feature-card">
              <img className="feature-illustration" src="/illustrations/mail-check.svg" alt="Подтверждение email пользователя" />
              <strong>Profile</strong>
              <p className="muted">
                Посмотрите текущий backend principal, обновите сессию и проверьте, что email и токены в порядке.
              </p>
              <Link href="/me" className="button-link secondary">
                Открыть /me
              </Link>
            </div>
            <div className="feature-card">
              <img className="feature-illustration" src="/illustrations/orbit.svg" alt="Командное общение и каналы" />
              <strong>Chat</strong>
              <p className="muted">
                Каналы и сообщения уже работают в backend, а этот экран показывает их в продуктовой форме без отдельного chat-backend.
              </p>
              <Link href="/chat" className="button-link secondary">
                Открыть chat
              </Link>
            </div>
            <div className="feature-card">
              <img className="feature-illustration" src="/illustrations/workspace.svg" alt="Доска заявок и предложений" />
              <strong>Заявки и предложения</strong>
              <p className="muted">Отдельная страница с витриной заявок (orders) и предложений (offers) для быстрого просмотра.</p>
              <Link href="/offers-requests" className="button-link secondary">
                Открыть доску
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
