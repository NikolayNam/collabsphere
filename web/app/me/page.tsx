"use client";

import { useEffect, useState } from "react";

import { Panel } from "@/components/panel";
import { APIError } from "@/lib/api";
import { clearTokens, loadMe, logout, readStoredTokens, rotateTokens, storeTokens, type MeResponse } from "@/lib/auth";

type LoadState =
  | { kind: "idle"; message: string }
  | { kind: "working"; message: string }
  | { kind: "error"; message: string }
  | { kind: "success"; message: string };

export default function MePage() {
  const [state, setState] = useState<LoadState>({ kind: "idle", message: "Ожидаем локальную сессию." });
  const [profile, setProfile] = useState<MeResponse | null>(null);
  const [tokenPreview, setTokenPreview] = useState<{ access: string; refresh: string } | null>(null);

  async function refreshProfile() {
    const stored = readStoredTokens();
    if (!stored?.accessToken) {
      setProfile(null);
      setTokenPreview(null);
      setState({
        kind: "error",
        message: "Локальные токены не найдены. Сначала завершите login через /login.",
      });
      return;
    }

    setState({
      kind: "working",
      message: "Запрашиваем /v1/auth/me с сохранённым access token.",
    });

    try {
      const currentProfile = await loadMe(stored.accessToken);
      setProfile(currentProfile);
      setTokenPreview({
        access: stored.accessToken,
        refresh: stored.refreshToken,
      });
      setState({
        kind: "success",
        message: "Backend principal загружен успешно.",
      });
    } catch (error) {
      const message =
        error instanceof APIError
          ? `${error.message}${error.code ? ` (${error.code})` : ""}`
          : error instanceof Error
            ? error.message
            : "Unknown /auth/me error";
      setProfile(null);
      setState({ kind: "error", message });
    }
  }

  async function handleRefreshTokens() {
    const stored = readStoredTokens();
    if (!stored?.refreshToken) {
      setState({ kind: "error", message: "Refresh token не найден." });
      return;
    }

    setState({
      kind: "working",
      message: "Выполняем /v1/auth/refresh и обновляем локальную сессию.",
    });

    try {
      const tokens = await rotateTokens(stored.refreshToken);
      storeTokens(tokens);
      await refreshProfile();
    } catch (error) {
      const message =
        error instanceof APIError
          ? `${error.message}${error.code ? ` (${error.code})` : ""}`
          : error instanceof Error
            ? error.message
            : "Unknown refresh error";
      setState({ kind: "error", message });
    }
  }

  async function handleLogout() {
    const stored = readStoredTokens();
    try {
      if (stored?.refreshToken) {
        await logout(stored.refreshToken);
      }
    } catch {
      // Intentionally ignore logout transport errors; we still clear local state.
    } finally {
      clearTokens();
      setProfile(null);
      setTokenPreview(null);
      setState({
        kind: "idle",
        message: "Локальная сессия очищена.",
      });
    }
  }

  useEffect(() => {
    void refreshProfile();
  }, []);

  return (
    <>
      <Panel
        title="Current principal"
        eyebrow="GET /v1/auth/me"
        actions={
          <div className="button-row">
            <button className="button secondary" onClick={() => void refreshProfile()} type="button">
              Reload
            </button>
            <button className="button secondary" onClick={() => void handleRefreshTokens()} type="button">
              Refresh tokens
            </button>
            <button className="button danger" onClick={() => void handleLogout()} type="button">
              Logout
            </button>
          </div>
        }
      >
        <div className={`status-card ${state.kind === "error" ? "error" : state.kind === "success" ? "success" : "info"}`}>
          <strong>{state.kind === "success" ? "Сессия активна" : state.kind === "error" ? "Сессия недоступна" : "Статус сессии"}</strong>
          <p className="status-copy">{state.message}</p>
        </div>
      </Panel>

      {profile ? (
        <>
          <section className="split">
            <Panel title="Profile" eyebrow="Backend response">
              <dl className="info-list">
                <dt>Account ID</dt>
                <dd>{profile.id}</dd>
                <dt>Email</dt>
                <dd>{profile.email}</dd>
                <dt>Display name</dt>
                <dd>{profile.displayName || "—"}</dd>
                <dt>Active</dt>
                <dd>{String(profile.isActive)}</dd>
                <dt>Created at</dt>
                <dd>{profile.createdAt}</dd>
                <dt>Updated at</dt>
                <dd>{profile.updatedAt || "—"}</dd>
              </dl>
            </Panel>
            <Panel title="Extended profile fields" eyebrow="Optional user metadata">
              <dl className="info-list">
                <dt>Bio</dt>
                <dd>{profile.bio || "—"}</dd>
                <dt>Phone</dt>
                <dd>{profile.phone || "—"}</dd>
                <dt>Locale</dt>
                <dd>{profile.locale || "—"}</dd>
                <dt>Timezone</dt>
                <dd>{profile.timezone || "—"}</dd>
                <dt>Website</dt>
                <dd>{profile.website || "—"}</dd>
                <dt>Avatar object</dt>
                <dd>{profile.avatarObjectId || "—"}</dd>
              </dl>
            </Panel>
          </section>

          {tokenPreview ? (
            <section className="split">
              <Panel title="Access token" eyebrow="Local storage preview">
                <textarea className="code-block" readOnly value={tokenPreview.access} />
              </Panel>
              <Panel title="Refresh token" eyebrow="Local storage preview">
                <textarea className="code-block" readOnly value={tokenPreview.refresh} />
              </Panel>
            </section>
          ) : null}
        </>
      ) : (
        <Panel title="No active profile" eyebrow="Session required">
          <div className="empty-state">
            У frontend сейчас нет валидной локальной сессии. Сначала откройте <code>/login</code> и
            завершите browser flow или legacy fallback login.
          </div>
        </Panel>
      )}
    </>
  );
}
