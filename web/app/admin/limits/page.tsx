"use client";

import { FormEvent, useEffect, useMemo, useState } from "react";

import { Panel } from "@/components/panel";
import { APIError, apiFetch } from "@/lib/api";
import { readStoredTokens } from "@/lib/auth";

type Status = {
  kind: "idle" | "working" | "success" | "error";
  title: string;
  description: string;
};

type AttachmentLimit = {
  id: string;
  scopeType: string;
  scopeId?: string | null;
  documentLimitBytes: number;
  photoLimitBytes: number;
  videoLimitBytes: number;
  totalLimitBytes: number;
  createdAt: string;
  updatedAt: string;
};

type PlatformAccess = {
  accountId: string;
  bootstrapAdmin: boolean;
  effectiveRoles: string[];
  storedRoles: string[];
};

const initialAccessState: Status = {
  kind: "idle",
  title: "Доступ к control-plane",
  description: "Проверяем `GET /v1/platform/access/me` для доступа к управлению лимитами.",
};

const initialPlatformLimitState: Status = {
  kind: "idle",
  title: "Лимит по умолчанию",
  description: "Платформенный лимит используется, когда нет переопределения для организации или аккаунта.",
};

const initialListState: Status = {
  kind: "idle",
  title: "Переопределения",
  description: "Лимиты для организаций и аккаунтов. Приоритет: account > organization > platform.",
};

function formatBytes(value: number): string {
  if (value >= 1_073_741_824) {
    return `${(value / 1_073_741_824).toFixed(1)} GB`;
  }
  if (value >= 1_048_576) {
    return `${(value / 1_048_576).toFixed(1)} MB`;
  }
  if (value >= 1024) {
    return `${(value / 1024).toFixed(1)} KB`;
  }
  return `${value} B`;
}

function parseBytesInput(value: string): number {
  const trimmed = value.trim().toLowerCase();
  if (!trimmed) return 0;
  const num = parseFloat(trimmed.replace(/[^0-9.]/g, ""));
  if (Number.isNaN(num)) return 0;
  if (trimmed.endsWith("gb")) return Math.round(num * 1_073_741_824);
  if (trimmed.endsWith("mb")) return Math.round(num * 1_048_576);
  if (trimmed.endsWith("kb")) return Math.round(num * 1024);
  return Math.round(num);
}

function problemMessage(error: unknown, fallback: string): string {
  if (error instanceof APIError) {
    return `${error.message}${error.code ? ` (${error.code})` : ""}`;
  }
  if (error instanceof Error) {
    return error.message;
  }
  return fallback;
}

export default function AdminLimitsPage() {
  const accessToken = useMemo(() => readStoredTokens()?.accessToken || null, []);
  const [platformAccess, setPlatformAccess] = useState<PlatformAccess | null>(null);
  const [accessState, setAccessState] = useState<Status>(initialAccessState);
  const [platformLimit, setPlatformLimit] = useState<AttachmentLimit | null>(null);
  const [platformLimitState, setPlatformLimitState] = useState<Status>(initialPlatformLimitState);
  const [limitsList, setLimitsList] = useState<AttachmentLimit[]>([]);
  const [listState, setListState] = useState<Status>(initialListState);
  const [upsertState, setUpsertState] = useState<Status>({ kind: "idle", title: "", description: "" });

  const [docLimitInput, setDocLimitInput] = useState("10 MB");
  const [photoLimitInput, setPhotoLimitInput] = useState("15 MB");
  const [videoLimitInput, setVideoLimitInput] = useState("100 MB");
  const [totalLimitInput, setTotalLimitInput] = useState("1 GB");

  const [newScopeType, setNewScopeType] = useState<"organization" | "account">("organization");
  const [newScopeId, setNewScopeId] = useState("");
  const [newDocLimit, setNewDocLimit] = useState("10 MB");
  const [newPhotoLimit, setNewPhotoLimit] = useState("15 MB");
  const [newVideoLimit, setNewVideoLimit] = useState("100 MB");
  const [newTotalLimit, setNewTotalLimit] = useState("1 GB");

  const isPlatformAdmin = Boolean(
    platformAccess && (platformAccess.bootstrapAdmin || platformAccess.effectiveRoles.includes("platform_admin")),
  );

  useEffect(() => {
    let cancelled = false;

    async function loadAccess() {
      if (!accessToken) {
        if (!cancelled) {
          setAccessState({
            kind: "error",
            title: "Нет сессии",
            description: "Войдите через /login для доступа к админке.",
          });
        }
        return;
      }

      if (!cancelled) {
        setAccessState({ kind: "working", title: "Проверяем доступ", description: "..." });
      }

      try {
        const payload = await apiFetch<PlatformAccess | { body?: PlatformAccess }>(`/v1/platform/access/me`, { accessToken });
        const access = (payload as { body?: PlatformAccess }).body ?? (payload as PlatformAccess);
        if (cancelled) return;
        setPlatformAccess(access);
        setAccessState({
          kind: "success",
          title: "Доступ получен",
          description: access.effectiveRoles?.length ? `Роли: ${access.effectiveRoles.join(", ")}` : "Нет platform-ролей.",
        });
      } catch (error) {
        if (cancelled) return;
        setPlatformAccess(null);
        setAccessState({
          kind: "error",
          title: "Нет доступа",
          description: problemMessage(error, "Не удалось загрузить platform access."),
        });
      }
    }

    void loadAccess();
    return () => {
      cancelled = true;
    };
  }, [accessToken]);

  useEffect(() => {
    if (!accessToken || !isPlatformAdmin) return;

    let cancelled = false;

    async function loadPlatformLimit() {
      setPlatformLimitState((s) => ({ ...s, kind: "working" as const, title: "Загружаем лимит", description: "..." }));
      try {
        const payload = await apiFetch<AttachmentLimit | { body?: AttachmentLimit }>(`/v1/platform/attachment-limits/platform`, { accessToken });
        const limit = (payload as { body?: AttachmentLimit }).body ?? (payload as AttachmentLimit);
        if (cancelled) return;
        setPlatformLimit(limit);
        setDocLimitInput(formatBytes(limit.documentLimitBytes));
        setPhotoLimitInput(formatBytes(limit.photoLimitBytes));
        setVideoLimitInput(formatBytes(limit.videoLimitBytes));
        setTotalLimitInput(formatBytes(limit.totalLimitBytes));
        setPlatformLimitState({
          kind: "success",
          title: "Лимит по умолчанию",
          description: `Документы: ${formatBytes(limit.documentLimitBytes)}, фото: ${formatBytes(limit.photoLimitBytes)}, видео: ${formatBytes(limit.videoLimitBytes)}, всего: ${formatBytes(limit.totalLimitBytes)}`,
        });
      } catch (error) {
        if (cancelled) return;
        setPlatformLimitState({
          kind: "error",
          title: "Ошибка загрузки",
          description: problemMessage(error, "Не удалось загрузить platform limit."),
        });
      }
    }

    void loadPlatformLimit();
    return () => {
      cancelled = true;
    };
  }, [accessToken, isPlatformAdmin]);

  useEffect(() => {
    if (!accessToken || !isPlatformAdmin) return;

    let cancelled = false;

    async function loadList() {
      setListState((s) => ({ ...s, kind: "working" as const, title: "Загружаем список", description: "..." }));
      try {
        const payload = await apiFetch<{ body?: { items?: AttachmentLimit[] }; items?: AttachmentLimit[] }>(
          `/v1/platform/attachment-limits`,
          { accessToken },
        );
        const items = payload.body?.items ?? payload.items ?? [];
        if (cancelled) return;
        setLimitsList(Array.isArray(items) ? items : []);
        setListState({
          kind: "success",
          title: "Переопределения",
          description: `${items.length} записей (organization/account).`,
        });
      } catch (error) {
        if (cancelled) return;
        setListState({
          kind: "error",
          title: "Ошибка загрузки",
          description: problemMessage(error, "Не удалось загрузить список."),
        });
      }
    }

    void loadList();
    return () => {
      cancelled = true;
    };
  }, [accessToken, isPlatformAdmin]);

  async function handleUpsertPlatform(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!accessToken) return;

    const docBytes = parseBytesInput(docLimitInput);
    const photoBytes = parseBytesInput(photoLimitInput);
    const videoBytes = parseBytesInput(videoLimitInput);
    const totalBytes = parseBytesInput(totalLimitInput);

    if (docBytes <= 0 || photoBytes <= 0 || videoBytes <= 0 || totalBytes <= 0) {
      setUpsertState({
        kind: "error",
        title: "Неверные значения",
        description: "Все лимиты должны быть положительными (например: 10 MB, 1 GB).",
      });
      return;
    }

    setUpsertState({ kind: "working", title: "Сохраняем...", description: "" });

    try {
      await apiFetch(`/v1/platform/attachment-limits/platform`, {
        method: "PUT",
        accessToken,
        bodyJSON: {
          documentLimitBytes: docBytes,
          photoLimitBytes: photoBytes,
          videoLimitBytes: videoBytes,
          totalLimitBytes: totalBytes,
        },
      });
      setUpsertState({ kind: "success", title: "Лимит обновлён", description: "" });
      setPlatformLimit({
        id: platformLimit?.id ?? "",
        scopeType: "platform",
        documentLimitBytes: docBytes,
        photoLimitBytes: photoBytes,
        videoLimitBytes: videoBytes,
        totalLimitBytes: totalBytes,
        createdAt: platformLimit?.createdAt ?? "",
        updatedAt: new Date().toISOString(),
      });
    } catch (error) {
      setUpsertState({
        kind: "error",
        title: "Ошибка сохранения",
        description: problemMessage(error, "Не удалось обновить лимит."),
      });
    }
  }

  async function handleAddOverride(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!accessToken || !newScopeId.trim()) return;

    const docBytes = parseBytesInput(newDocLimit);
    const photoBytes = parseBytesInput(newPhotoLimit);
    const videoBytes = parseBytesInput(newVideoLimit);
    const totalBytes = parseBytesInput(newTotalLimit);

    if (docBytes <= 0 || photoBytes <= 0 || videoBytes <= 0 || totalBytes <= 0) {
      setUpsertState({
        kind: "error",
        title: "Неверные значения",
        description: "Все лимиты должны быть положительными.",
      });
      return;
    }

    const path =
      newScopeType === "organization"
        ? `/v1/platform/attachment-limits/organizations/${newScopeId.trim()}`
        : `/v1/platform/attachment-limits/accounts/${newScopeId.trim()}`;

    setUpsertState({ kind: "working", title: "Добавляем переопределение...", description: "" });

    try {
      await apiFetch(path, {
        method: "PUT",
        accessToken,
        bodyJSON: {
          documentLimitBytes: docBytes,
          photoLimitBytes: photoBytes,
          videoLimitBytes: videoBytes,
          totalLimitBytes: totalBytes,
        },
      });
      setUpsertState({ kind: "success", title: "Переопределение добавлено", description: "" });
      setNewScopeId("");
      const listPayload = await apiFetch<{ body?: { items?: AttachmentLimit[] }; items?: AttachmentLimit[] }>(
        `/v1/platform/attachment-limits`,
        { accessToken },
      );
      const items = listPayload.body?.items ?? listPayload.items ?? [];
      setLimitsList(Array.isArray(items) ? items : []);
    } catch (error) {
      setUpsertState({
        kind: "error",
        title: "Ошибка",
        description: problemMessage(error, "Не удалось добавить переопределение."),
      });
    }
  }

  async function handleDeleteOverride(scopeType: string, scopeId: string) {
    if (!accessToken) return;

    const path =
      scopeType === "organization"
        ? `/v1/platform/attachment-limits/organizations/${scopeId}`
        : `/v1/platform/attachment-limits/accounts/${scopeId}`;

    setUpsertState({ kind: "working", title: "Удаляем...", description: "" });

    try {
      await apiFetch(path, { method: "DELETE", accessToken });
      setUpsertState({ kind: "success", title: "Удалено", description: "" });
      setLimitsList((prev) => prev.filter((l) => l.scopeId !== scopeId));
    } catch (error) {
      setUpsertState({
        kind: "error",
        title: "Ошибка удаления",
        description: problemMessage(error, "Не удалось удалить."),
      });
    }
  }

  return (
    <>
      <Panel title="Лимиты вложений" eyebrow="Platform / Attachment Limits">
        <div className={`status-card ${accessState.kind === "error" ? "error" : accessState.kind === "success" ? "success" : "info"}`}>
          <strong>{accessState.title}</strong>
          <p className="status-copy">{accessState.description}</p>
        </div>
        {platformAccess ? (
          <div className="mini-card">
            <p className="muted">Account: {platformAccess.accountId}</p>
            <p className="muted">Effective roles: {(platformAccess.effectiveRoles || []).join(", ") || "none"}</p>
          </div>
        ) : null}
      </Panel>

      {isPlatformAdmin ? (
        <section className="split">
          <Panel title="Платформенный лимит по умолчанию" eyebrow="Platform scope">
            <div className={`status-card ${platformLimitState.kind === "error" ? "error" : platformLimitState.kind === "success" ? "success" : "info"}`}>
              <strong>{platformLimitState.title}</strong>
              <p className="status-copy">{platformLimitState.description}</p>
            </div>

            <form className="form-grid" onSubmit={handleUpsertPlatform}>
              <div className="form-row two">
                <div className="form-row">
                  <label className="form-label" htmlFor="limits-doc">
                    Документы (макс. размер файла)
                  </label>
                  <input
                    id="limits-doc"
                    className="text-input"
                    value={docLimitInput}
                    onChange={(e) => setDocLimitInput(e.target.value)}
                    placeholder="10 MB"
                  />
                </div>
                <div className="form-row">
                  <label className="form-label" htmlFor="limits-photo">
                    Фото
                  </label>
                  <input
                    id="limits-photo"
                    className="text-input"
                    value={photoLimitInput}
                    onChange={(e) => setPhotoLimitInput(e.target.value)}
                    placeholder="15 MB"
                  />
                </div>
              </div>
              <div className="form-row two">
                <div className="form-row">
                  <label className="form-label" htmlFor="limits-video">
                    Видео
                  </label>
                  <input
                    id="limits-video"
                    className="text-input"
                    value={videoLimitInput}
                    onChange={(e) => setVideoLimitInput(e.target.value)}
                    placeholder="100 MB"
                  />
                </div>
                <div className="form-row">
                  <label className="form-label" htmlFor="limits-total">
                    Общий лимит на пользователя
                  </label>
                  <input
                    id="limits-total"
                    className="text-input"
                    value={totalLimitInput}
                    onChange={(e) => setTotalLimitInput(e.target.value)}
                    placeholder="1 GB"
                  />
                </div>
              </div>
              <div className="button-row">
                <button className="button primary" type="submit">
                  Сохранить лимит по умолчанию
                </button>
              </div>
            </form>

            {upsertState.kind !== "idle" && upsertState.title ? (
              <div className={`status-card ${upsertState.kind === "error" ? "error" : upsertState.kind === "success" ? "success" : "info"}`} style={{ marginTop: 16 }}>
                <strong>{upsertState.title}</strong>
                {upsertState.description ? <p className="status-copy">{upsertState.description}</p> : null}
              </div>
            ) : null}
          </Panel>

          <div>
            <Panel title="Переопределения по организации/аккаунту" eyebrow="Overrides">
              <div className={`status-card ${listState.kind === "error" ? "error" : listState.kind === "success" ? "success" : "info"}`}>
                <strong>{listState.title}</strong>
                <p className="status-copy">{listState.description}</p>
              </div>

              <form className="form-grid" onSubmit={handleAddOverride} style={{ marginTop: 16 }}>
                <div className="form-row two">
                  <div className="form-row">
                    <label className="form-label" htmlFor="new-scope-type">
                      Scope
                    </label>
                    <select
                      id="new-scope-type"
                      className="text-input"
                      value={newScopeType}
                      onChange={(e) => setNewScopeType(e.target.value as "organization" | "account")}
                    >
                      <option value="organization">organization</option>
                      <option value="account">account</option>
                    </select>
                  </div>
                  <div className="form-row">
                    <label className="form-label" htmlFor="new-scope-id">
                      ID (UUID)
                    </label>
                    <input
                      id="new-scope-id"
                      className="text-input"
                      value={newScopeId}
                      onChange={(e) => setNewScopeId(e.target.value)}
                      placeholder="00000000-0000-0000-0000-000000000000"
                    />
                  </div>
                </div>
                <div className="form-row two">
                  <div className="form-row">
                    <label className="form-label">Документы</label>
                    <input className="text-input" value={newDocLimit} onChange={(e) => setNewDocLimit(e.target.value)} placeholder="10 MB" />
                  </div>
                  <div className="form-row">
                    <label className="form-label">Фото</label>
                    <input className="text-input" value={newPhotoLimit} onChange={(e) => setNewPhotoLimit(e.target.value)} placeholder="15 MB" />
                  </div>
                </div>
                <div className="form-row two">
                  <div className="form-row">
                    <label className="form-label">Видео</label>
                    <input className="text-input" value={newVideoLimit} onChange={(e) => setNewVideoLimit(e.target.value)} placeholder="100 MB" />
                  </div>
                  <div className="form-row">
                    <label className="form-label">Всего</label>
                    <input className="text-input" value={newTotalLimit} onChange={(e) => setNewTotalLimit(e.target.value)} placeholder="1 GB" />
                  </div>
                </div>
                <div className="button-row">
                  <button className="button secondary" type="submit">
                    Добавить переопределение
                  </button>
                </div>
              </form>

              <div className="selection-list" style={{ marginTop: 16 }}>
                {limitsList
                  .filter((l) => l.scopeType !== "platform")
                  .map((limit) => (
                    <div key={`${limit.scopeType}-${limit.scopeId}`} className="selection-card">
                      <strong>
                        {limit.scopeType}: {limit.scopeId ?? "—"}
                      </strong>
                      <span className="muted">
                        doc: {formatBytes(limit.documentLimitBytes)}, photo: {formatBytes(limit.photoLimitBytes)}, video: {formatBytes(limit.videoLimitBytes)}, total: {formatBytes(limit.totalLimitBytes)}
                      </span>
                      <button
                        type="button"
                        className="button secondary"
                        style={{ marginTop: 8 }}
                        onClick={() => limit.scopeId && handleDeleteOverride(limit.scopeType, limit.scopeId)}
                      >
                        Удалить
                      </button>
                    </div>
                  ))}
                {limitsList.filter((l) => l.scopeType !== "platform").length === 0 ? (
                  <p className="muted">Нет переопределений. Используется платформенный лимит.</p>
                ) : null}
              </div>
            </Panel>
          </div>
        </section>
      ) : (
        <Panel title="Доступ ограничен" eyebrow="Admin only">
          <p className="muted">Управление лимитами доступно роли platform_admin или bootstrap admin.</p>
        </Panel>
      )}
    </>
  );
}
