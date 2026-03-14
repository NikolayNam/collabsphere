"use client";

import { ChangeEvent, FormEvent, useEffect, useState } from "react";

import { Panel } from "@/components/panel";
import { APIError, apiFetch } from "@/lib/api";
import { clearTokens, loadMe, logout, readStoredTokens, rotateTokens, storeTokens, type MeResponse } from "@/lib/auth";

type LoadState =
  | { kind: "idle"; message: string }
  | { kind: "working"; message: string }
  | { kind: "error"; message: string }
  | { kind: "success"; message: string };

type AccountKYCDocument = {
  id: string;
  accountId: string;
  objectId: string;
  documentType: string;
  title: string;
  status: string;
  reviewNote?: string;
  reviewerAccountId?: string;
  createdAt: string;
  updatedAt?: string;
  reviewedAt?: string;
};

type AccountKYCProfile = {
  accountId: string;
  status: string;
  legalName?: string;
  countryCode?: string;
  documentNumber?: string;
  residenceAddress?: string;
  reviewNote?: string;
  reviewerAccountId?: string;
  submittedAt?: string;
  reviewedAt?: string;
  createdAt: string;
  updatedAt: string;
  documents: AccountKYCDocument[];
};

function kycStatusLabel(status?: string): string {
  const value = (status || "").trim().toLowerCase();
  switch (value) {
    case "draft":
      return "Черновик";
    case "submitted":
      return "Отправлено на проверку";
    case "in_review":
      return "На проверке";
    case "needs_info":
      return "Нужны уточнения";
    case "approved":
      return "Подтверждено";
    case "rejected":
      return "Отклонено";
    default:
      return status || "unknown";
  }
}

function isReviewSubmissionLocked(status?: string): boolean {
  const value = (status || "").trim().toLowerCase();
  return value === "submitted" || value === "in_review" || value === "approved";
}

export default function MePage() {
  const [state, setState] = useState<LoadState>({ kind: "idle", message: "Ожидаем локальную сессию." });
  const [profile, setProfile] = useState<MeResponse | null>(null);
  const [tokenPreview, setTokenPreview] = useState<{ access: string; refresh: string } | null>(null);
  const [kyc, setKYC] = useState<AccountKYCProfile | null>(null);
  const [kycState, setKYCState] = useState<LoadState>({
    kind: "idle",
    message: "KYC профиль ещё не загружен.",
  });
  const [kycDraft, setKYCDraft] = useState({
    status: "draft",
    legalName: "",
    countryCode: "",
    documentNumber: "",
    residenceAddress: "",
  });
  const [kycFile, setKYCFile] = useState<File | null>(null);
  const [kycDocumentType, setKYCDocumentType] = useState("identity_document");
  const [kycDocumentTitle, setKYCDocumentTitle] = useState("");

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
      await refreshKYC(stored.accessToken);
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

  async function refreshKYC(accessToken?: string | null) {
    const token = accessToken || readStoredTokens()?.accessToken || null;
    if (!token) {
      setKYC(null);
      setKYCState({ kind: "error", message: "Нет токена для чтения KYC." });
      return;
    }
    try {
      const payload = await apiFetch<AccountKYCProfile>("/v1/accounts/me/kyc", { accessToken: token });
      setKYC(payload);
      setKYCDraft({
        status: payload.status || "draft",
        legalName: payload.legalName || "",
        countryCode: payload.countryCode || "",
        documentNumber: payload.documentNumber || "",
        residenceAddress: payload.residenceAddress || "",
      });
      setKYCState({ kind: "success", message: "KYC профиль загружен." });
    } catch (error) {
      const message =
        error instanceof APIError
          ? `${error.message}${error.code ? ` (${error.code})` : ""}`
          : error instanceof Error
            ? error.message
            : "Unknown KYC error";
      setKYC(null);
      setKYCState({ kind: "error", message });
    }
  }

  async function handleSaveKYC(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const accessToken = readStoredTokens()?.accessToken || null;
    if (!accessToken) {
      setKYCState({ kind: "error", message: "Сначала выполните login." });
      return;
    }
    setKYCState({ kind: "working", message: "Сохраняем KYC профиль..." });
    try {
      await apiFetch<AccountKYCProfile>("/v1/accounts/me/kyc", {
        method: "PATCH",
        accessToken,
        bodyJSON: {
          status: kycDraft.status,
          legalName: kycDraft.legalName,
          countryCode: kycDraft.countryCode,
          documentNumber: kycDraft.documentNumber,
          residenceAddress: kycDraft.residenceAddress,
        },
      });
      await refreshKYC(accessToken);
      setKYCState({ kind: "success", message: "KYC профиль сохранён." });
    } catch (error) {
      const message =
        error instanceof APIError
          ? `${error.message}${error.code ? ` (${error.code})` : ""}`
          : error instanceof Error
            ? error.message
            : "Unknown KYC update error";
      setKYCState({ kind: "error", message });
    }
  }

  async function handleSubmitKYCForReview() {
    const accessToken = readStoredTokens()?.accessToken || null;
    if (!accessToken) {
      setKYCState({ kind: "error", message: "Сначала выполните login." });
      return;
    }
    setKYCState({ kind: "working", message: "Отправляем KYC профиль на review..." });
    try {
      await apiFetch<AccountKYCProfile>("/v1/accounts/me/kyc", {
        method: "PATCH",
        accessToken,
        bodyJSON: {
          status: "submitted",
          legalName: kycDraft.legalName,
          countryCode: kycDraft.countryCode,
          documentNumber: kycDraft.documentNumber,
          residenceAddress: kycDraft.residenceAddress,
        },
      });
      await refreshKYC(accessToken);
      setKYCState({ kind: "success", message: "KYC профиль отправлен на проверку." });
    } catch (error) {
      const message =
        error instanceof APIError
          ? `${error.message}${error.code ? ` (${error.code})` : ""}`
          : error instanceof Error
            ? error.message
            : "Unknown KYC submit error";
      setKYCState({ kind: "error", message });
    }
  }

  function handleKYCFileSelection(event: ChangeEvent<HTMLInputElement>) {
    setKYCFile(event.target.files?.[0] || null);
  }

  async function handleUploadKYCDocument(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const accessToken = readStoredTokens()?.accessToken || null;
    if (!accessToken) {
      setKYCState({ kind: "error", message: "Сначала выполните login." });
      return;
    }
    if (!kycFile) {
      setKYCState({ kind: "error", message: "Выберите файл KYC документа." });
      return;
    }
    setKYCState({ kind: "working", message: "Создаём upload session..." });
    try {
      const upload = await apiFetch<{
        id: string;
        uploadUrl: string;
      }>("/v1/accounts/me/kyc/documents/uploads", {
        method: "POST",
        accessToken,
        bodyJSON: {
          documentType: kycDocumentType,
          title: kycDocumentTitle || undefined,
          fileName: kycFile.name,
          contentType: kycFile.type || "application/octet-stream",
          sizeBytes: kycFile.size,
        },
      });

      const putResponse = await fetch(upload.uploadUrl, {
        method: "PUT",
        headers: kycFile.type ? { "Content-Type": kycFile.type } : undefined,
        body: kycFile,
      });
      if (!putResponse.ok) {
        throw new Error(`KYC upload PUT failed: HTTP ${putResponse.status}`);
      }

      await apiFetch(`/v1/accounts/me/kyc/documents/uploads/${encodeURIComponent(upload.id)}/complete`, {
        method: "POST",
        accessToken,
      });

      setKYCFile(null);
      setKYCDocumentTitle("");
      await refreshKYC(accessToken);
      setKYCState({ kind: "success", message: "KYC документ загружен и зарегистрирован." });
    } catch (error) {
      const message =
        error instanceof APIError
          ? `${error.message}${error.code ? ` (${error.code})` : ""}`
          : error instanceof Error
            ? error.message
            : "Unknown KYC upload error";
      setKYCState({ kind: "error", message });
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

          <Panel title="Account KYC" eyebrow="Accounts / KYC">
            <div className={`status-card ${kycState.kind === "error" ? "error" : kycState.kind === "success" ? "success" : "info"}`}>
              <strong>KYC status</strong>
              <p className="status-copy">{kycState.message}</p>
              {kyc?.status ? (
                <p className="status-copy">
                  Текущий статус: <strong>{kycStatusLabel(kyc.status)}</strong>
                </p>
              ) : null}
            </div>
            <form className="form-grid" onSubmit={handleSaveKYC}>
              <div className="form-row two">
                <div className="form-row">
                  <label className="form-label">Review status</label>
                  <input className="text-input" value={kycStatusLabel(kyc?.status || kycDraft.status)} readOnly />
                </div>
                <div className="form-row">
                  <label className="form-label">Country code</label>
                  <input className="text-input" value={kycDraft.countryCode} onChange={(e) => setKYCDraft((v) => ({ ...v, countryCode: e.target.value }))} />
                </div>
              </div>
              <div className="form-row two">
                <div className="form-row">
                  <label className="form-label">Legal name</label>
                  <input className="text-input" value={kycDraft.legalName} onChange={(e) => setKYCDraft((v) => ({ ...v, legalName: e.target.value }))} />
                </div>
                <div className="form-row">
                  <label className="form-label">Document number</label>
                  <input className="text-input" value={kycDraft.documentNumber} onChange={(e) => setKYCDraft((v) => ({ ...v, documentNumber: e.target.value }))} />
                </div>
              </div>
              <div className="form-row">
                <label className="form-label">Residence address</label>
                <textarea className="textarea" rows={3} value={kycDraft.residenceAddress} onChange={(e) => setKYCDraft((v) => ({ ...v, residenceAddress: e.target.value }))} />
              </div>
              <div className="button-row">
                <button className="button primary" type="submit">Save KYC profile</button>
                <button
                  className="button secondary"
                  onClick={() => void handleSubmitKYCForReview()}
                  type="button"
                  disabled={isReviewSubmissionLocked(kyc?.status)}
                >
                  Submit for review
                </button>
              </div>
            </form>

            <form className="form-grid" onSubmit={handleUploadKYCDocument}>
              <div className="form-row two">
                <div className="form-row">
                  <label className="form-label">Document type</label>
                  <input className="text-input" value={kycDocumentType} onChange={(e) => setKYCDocumentType(e.target.value)} />
                </div>
                <div className="form-row">
                  <label className="form-label">Title</label>
                  <input className="text-input" value={kycDocumentTitle} onChange={(e) => setKYCDocumentTitle(e.target.value)} />
                </div>
              </div>
              <div className="form-row">
                <label className="form-label">File</label>
                <input type="file" onChange={handleKYCFileSelection} />
              </div>
              <div className="button-row">
                <button className="button secondary" type="submit">Upload KYC document</button>
              </div>
            </form>

            {kyc?.documents?.length ? (
              <div className="domain-list">
                {kyc.documents.map((item) => (
                  <div key={item.id} className="inline-panel">
                    <strong>{item.title || item.documentType}</strong>
                    <p className="muted">
                      {item.documentType} · {kycStatusLabel(item.status)}
                    </p>
                    <p className="muted">Created: {item.createdAt}</p>
                    {item.reviewNote ? <p className="muted">Note: {item.reviewNote}</p> : null}
                  </div>
                ))}
              </div>
            ) : (
              <div className="empty-state">Пока нет загруженных KYC документов.</div>
            )}
          </Panel>
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
