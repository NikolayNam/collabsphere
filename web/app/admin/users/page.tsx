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

type PlatformAccess = {
  accountId: string;
  bootstrapAdmin: boolean;
  effectiveRoles: string[];
  storedRoles: string[];
};

type MyOrganization = {
  id: string;
  name: string;
  slug: string;
};

type MemberPayload = {
  id: string;
  organizationId: string;
  accountId: string;
  role: string;
  isActive: boolean;
  createdAt: string;
};

type AccountPayload = {
  id: string;
  email: string;
  displayName?: string | null;
  isActive: boolean;
  zitadelUserId?: string | null;
  subject?: string | null;
  externalSubject?: string | null;
};

type MemberRow = {
  membershipId: string;
  accountId: string;
  role: string;
  isActive: boolean;
  createdAt: string;
  email?: string;
  displayName?: string;
  zitadelUserId?: string;
};

type KYCReview = {
  reviewId: string;
  scope: "account" | "organization" | string;
  subjectId: string;
  status: string;
  legalName?: string;
  countryCode?: string;
  registrationNumber?: string;
  taxId?: string;
  documentNumber?: string;
  residenceAddress?: string;
  reviewNote?: string;
  reviewerAccountId?: string;
  submittedAt?: string;
  reviewedAt?: string;
  createdAt?: string;
  updatedAt: string;
};

const initialAccessState: Status = {
  kind: "idle",
  title: "Доступ к control-plane",
  description: "Проверяем `GET /v1/platform/access/me` и определяем, доступна ли страница администратора.",
};

const initialUsersState: Status = {
  kind: "idle",
  title: "Список пользователей",
  description: "Выберите организацию, чтобы загрузить участников через `GET /v1/organizations/{id}/members`.",
};

const initialVerifyState: Status = {
  kind: "idle",
  title: "Верификация email",
  description: "Введите raw ZITADEL user id и отправьте `POST /v1/platform/users/{userId}/email/force-verify`.",
};

const initialKYCState: Status = {
  kind: "idle",
  title: "KYC reviews",
  description: "Загружаем проверки пользователей и организаций через `GET /v1/platform/kyc/reviews`.",
};

function problemMessage(error: unknown, fallback: string): string {
  if (error instanceof APIError) {
    return `${error.message}${error.code ? ` (${error.code})` : ""}`;
  }
  if (error instanceof Error) {
    return error.message;
  }
  return fallback;
}

function formatTimestamp(value?: string): string {
  if (!value) {
    return "n/a";
  }
  const parsed = new Date(value);
  if (Number.isNaN(parsed.getTime())) {
    return value;
  }
  return new Intl.DateTimeFormat("ru-RU", {
    dateStyle: "short",
    timeStyle: "short",
  }).format(parsed);
}

function extractZitadelUserID(account: AccountPayload): string | undefined {
  const typedCandidates = [account.zitadelUserId, account.externalSubject, account.subject];
  for (const value of typedCandidates) {
    if (typeof value === "string" && value.trim()) {
      return value.trim();
    }
  }

  const raw = account as unknown as Record<string, unknown>;
  const rawCandidates = [
    raw.zitadel_user_id,
    raw.zitadelUserId,
    raw.external_subject,
    raw.externalSubject,
    raw.providerSubject,
    raw.subject,
  ];
  for (const value of rawCandidates) {
    if (typeof value === "string" && value.trim()) {
      return value.trim();
    }
  }
  return undefined;
}

function kycStatusLabel(status?: string): string {
  const value = (status || "").trim().toLowerCase();
  switch (value) {
    case "draft":
      return "Черновик";
    case "submitted":
      return "Отправлено";
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

function kycStatusBadgeClass(status?: string): string {
  const value = (status || "").trim().toLowerCase();
  switch (value) {
    case "approved":
      return "status-badge approved";
    case "rejected":
      return "status-badge rejected";
    case "in_review":
      return "status-badge in-review";
    case "needs_info":
      return "status-badge needs-info";
    case "submitted":
      return "status-badge submitted";
    case "draft":
      return "status-badge draft";
    default:
      return "status-badge";
  }
}

function normalizeKYCReviewListPayload(payload: unknown): KYCReview[] {
  const root = payload as Record<string, unknown> | null;
  const directItems = Array.isArray(root?.items) ? (root?.items as KYCReview[]) : null;
  if (directItems) {
    return directItems;
  }
  const body = root?.body as Record<string, unknown> | undefined;
  const nestedItems = Array.isArray(body?.items) ? (body?.items as KYCReview[]) : null;
  return nestedItems || [];
}

function normalizeKYCReviewPayload(payload: unknown): KYCReview | null {
  const root = payload as Record<string, unknown> | null;
  if (root && typeof root.reviewId === "string") {
    return root as unknown as KYCReview;
  }
  const body = root?.body as Record<string, unknown> | undefined;
  if (body && typeof body.reviewId === "string") {
    return body as unknown as KYCReview;
  }
  return null;
}

export default function AdminUsersPage() {
  const accessToken = useMemo(() => readStoredTokens()?.accessToken || null, []);
  const [platformAccess, setPlatformAccess] = useState<PlatformAccess | null>(null);
  const [accessState, setAccessState] = useState<Status>(initialAccessState);
  const [organizations, setOrganizations] = useState<MyOrganization[]>([]);
  const [selectedOrganizationId, setSelectedOrganizationId] = useState("");
  const [usersState, setUsersState] = useState<Status>(initialUsersState);
  const [members, setMembers] = useState<MemberRow[]>([]);
  const [zitadelUserId, setZitadelUserId] = useState("");
  const [verifyState, setVerifyState] = useState<Status>(initialVerifyState);
  const [verifyResult, setVerifyResult] = useState<unknown | null>(null);
  const [kycState, setKYCState] = useState<Status>(initialKYCState);
  const [kycReviews, setKYCReviews] = useState<KYCReview[]>([]);
  const [selectedReviewId, setSelectedReviewId] = useState("");
  const [selectedReview, setSelectedReview] = useState<KYCReview | null>(null);
  const [kycScopeFilter, setKYCScopeFilter] = useState("");
  const [kycStatusFilter, setKYCStatusFilter] = useState("");
  const [kycDecision, setKYCDecision] = useState<"approve" | "reject" | "request_info">("approve");
  const [kycReason, setKYCReason] = useState("");

  const isPlatformAdmin = Boolean(
    platformAccess && (platformAccess.bootstrapAdmin || platformAccess.effectiveRoles.includes("platform_admin")),
  );
  const canManageKYC = Boolean(
    platformAccess &&
      (platformAccess.bootstrapAdmin ||
        platformAccess.effectiveRoles.includes("platform_admin") ||
        platformAccess.effectiveRoles.includes("review_operator")),
  );

  useEffect(() => {
    let cancelled = false;

    async function loadAccess() {
      if (!accessToken) {
        if (!cancelled) {
          setAccessState({
            kind: "error",
            title: "Нет локальной сессии",
            description: "Сначала войдите через /login, затем откройте страницу admin users.",
          });
        }
        return;
      }

      setAccessState({
        kind: "working",
        title: "Проверяем права",
        description: "Читаем platform access профиль текущего аккаунта.",
      });

      try {
        const access = await apiFetch<PlatformAccess>("/v1/platform/access/me", { accessToken });
        if (cancelled) {
          return;
        }
        setPlatformAccess(access);
        setAccessState({
          kind: "success",
          title: "Доступ загружен",
          description: access.bootstrapAdmin || access.effectiveRoles.includes("platform_admin")
            ? "Аккаунт имеет admin-level доступ. Можно управлять верификацией и смотреть список пользователей."
            : "Аккаунт без platform_admin роли. Страница работает в режиме только чтения статуса доступа.",
        });
      } catch (error) {
        if (cancelled) {
          return;
        }
        setAccessState({
          kind: "error",
          title: "Не удалось проверить права",
          description: problemMessage(error, "Unknown platform access error"),
        });
      }
    }

    void loadAccess();
    return () => {
      cancelled = true;
    };
  }, [accessToken]);

  useEffect(() => {
    let cancelled = false;

    async function loadOrganizations() {
      if (!accessToken || !isPlatformAdmin) {
        setOrganizations([]);
        setSelectedOrganizationId("");
        return;
      }

      try {
        const payload = await apiFetch<{ data?: MyOrganization[] }>("/v1/organizations/my", { accessToken });
        if (cancelled) {
          return;
        }
        const items = Array.isArray(payload.data) ? payload.data : [];
        setOrganizations(items);
        setSelectedOrganizationId(items[0]?.id || "");
      } catch {
        if (cancelled) {
          return;
        }
        setOrganizations([]);
        setSelectedOrganizationId("");
      }
    }

    void loadOrganizations();
    return () => {
      cancelled = true;
    };
  }, [accessToken, isPlatformAdmin]);

  useEffect(() => {
    let cancelled = false;

    async function loadMembers() {
      if (!accessToken || !isPlatformAdmin) {
        setMembers([]);
        setUsersState(initialUsersState);
        return;
      }
      if (!selectedOrganizationId) {
        setMembers([]);
        setUsersState({
          kind: "idle",
          title: "Список пользователей",
          description: "Выберите организацию, чтобы увидеть пользователей (участников).",
        });
        return;
      }

      setUsersState({
        kind: "working",
        title: "Загружаем пользователей",
        description: "Читаем memberships и профиль аккаунтов участников выбранной организации.",
      });

      try {
        const payload = await apiFetch<{ members?: MemberPayload[] }>(`/v1/organizations/${selectedOrganizationId}/members`, {
          accessToken,
        });
        const rawMembers = Array.isArray(payload.members) ? payload.members : [];
        const enriched = await Promise.all(
          rawMembers.map(async (member) => {
            try {
              const account = await apiFetch<AccountPayload>(`/v1/accounts/${member.accountId}`, { accessToken });
              return {
                membershipId: member.id,
                accountId: member.accountId,
                role: member.role,
                isActive: member.isActive,
                createdAt: member.createdAt,
                email: account.email,
                displayName: account.displayName || undefined,
                zitadelUserId: extractZitadelUserID(account),
              } satisfies MemberRow;
            } catch {
              return {
                membershipId: member.id,
                accountId: member.accountId,
                role: member.role,
                isActive: member.isActive,
                createdAt: member.createdAt,
              } satisfies MemberRow;
            }
          }),
        );

        if (cancelled) {
          return;
        }
        setMembers(enriched);
        setUsersState({
          kind: "success",
          title: enriched.length > 0 ? "Пользователи загружены" : "Пользователи не найдены",
          description:
            enriched.length > 0
              ? "Ниже показаны участники выбранной организации с accountId и ролью."
              : "Для выбранной организации пока нет участников.",
        });
      } catch (error) {
        if (cancelled) {
          return;
        }
        setMembers([]);
        setUsersState({
          kind: "error",
          title: "Не удалось загрузить список пользователей",
          description: problemMessage(error, "Unknown members list error"),
        });
      }
    }

    void loadMembers();
    return () => {
      cancelled = true;
    };
  }, [accessToken, isPlatformAdmin, selectedOrganizationId]);

  useEffect(() => {
    let cancelled = false;

    async function loadKYCReviews() {
      if (!accessToken || !canManageKYC) {
        setKYCReviews([]);
        setSelectedReviewId("");
        setSelectedReview(null);
        setKYCState(initialKYCState);
        return;
      }

      setKYCState({
        kind: "working",
        title: "KYC reviews",
        description: "Загружаем reviews из platform control-plane.",
      });
      try {
        const params = new URLSearchParams();
        if (kycScopeFilter) {
          params.set("scope", kycScopeFilter);
        }
        if (kycStatusFilter) {
          params.set("status", kycStatusFilter);
        }
        params.set("limit", "100");

        const payload = await apiFetch<unknown>(`/v1/platform/kyc/reviews?${params.toString()}`, { accessToken });
        if (cancelled) {
          return;
        }
        const items = normalizeKYCReviewListPayload(payload);
        setKYCReviews(items);
        setSelectedReviewId((current) => {
          if (current && items.some((item) => item.reviewId === current)) {
            return current;
          }
          return items[0]?.reviewId || "";
        });
        setKYCState({
          kind: "success",
          title: items.length > 0 ? "KYC reviews загружены" : "KYC reviews не найдены",
          description: items.length > 0 ? "Выберите review и примите решение ниже." : "По текущим фильтрам записей нет.",
        });
      } catch (error) {
        if (cancelled) {
          return;
        }
        setKYCReviews([]);
        setSelectedReviewId("");
        setSelectedReview(null);
        setKYCState({
          kind: "error",
          title: "Не удалось загрузить KYC reviews",
          description: problemMessage(error, "Unknown KYC list error"),
        });
      }
    }

    void loadKYCReviews();
    return () => {
      cancelled = true;
    };
  }, [accessToken, canManageKYC, kycScopeFilter, kycStatusFilter]);

  useEffect(() => {
    let cancelled = false;

    async function loadSelectedReview() {
      if (!accessToken || !canManageKYC || !selectedReviewId) {
        setSelectedReview(null);
        return;
      }
      try {
        const payload = await apiFetch<unknown>(`/v1/platform/kyc/reviews/${encodeURIComponent(selectedReviewId)}`, { accessToken });
        if (cancelled) {
          return;
        }
        setSelectedReview(normalizeKYCReviewPayload(payload));
      } catch {
        if (cancelled) {
          return;
        }
        setSelectedReview(null);
      }
    }

    void loadSelectedReview();
    return () => {
      cancelled = true;
    };
  }, [accessToken, canManageKYC, selectedReviewId]);

  async function handleForceVerify(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!accessToken || !isPlatformAdmin) {
      setVerifyState({
        kind: "error",
        title: "Верификация email",
        description: "Нужна admin-учётка с platform_admin доступом.",
      });
      return;
    }
    const userId = zitadelUserId.trim();
    if (!userId) {
      setVerifyState({
        kind: "error",
        title: "Верификация email",
        description: "Введите raw ZITADEL user id.",
      });
      return;
    }

    setVerifyState({
      kind: "working",
      title: "Верификация email",
      description: "Отправляем force-verify запрос в backend control-plane.",
    });

    try {
      const payload = await apiFetch(`/v1/platform/users/${encodeURIComponent(userId)}/email/force-verify`, {
        method: "POST",
        accessToken,
      });
      setVerifyResult(payload);
      setVerifyState({
        kind: "success",
        title: "Верификация email",
        description: "Email пользователя успешно подтверждён через backend admin flow.",
      });
    } catch (error) {
      setVerifyResult(null);
      setVerifyState({
        kind: "error",
        title: "Верификация email",
        description: problemMessage(error, "Unknown force-verify error"),
      });
    }
  }

  async function handleKYCDecision(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!accessToken || !canManageKYC || !selectedReviewId) {
      setKYCState({
        kind: "error",
        title: "KYC review",
        description: "Выберите review и убедитесь, что есть admin доступ.",
      });
      return;
    }
    setKYCState({
      kind: "working",
      title: "KYC review",
      description: "Отправляем решение по KYC review.",
    });
    try {
      const payload = await apiFetch<unknown>(`/v1/platform/kyc/reviews/${encodeURIComponent(selectedReviewId)}/decision`, {
        method: "POST",
        accessToken,
        bodyJSON: {
          decision: kycDecision,
          reason: kycReason.trim() || undefined,
        },
      });
      const next = normalizeKYCReviewPayload(payload);
      setSelectedReview(next);
      setKYCReason("");

      const params = new URLSearchParams();
      if (kycScopeFilter) {
        params.set("scope", kycScopeFilter);
      }
      if (kycStatusFilter) {
        params.set("status", kycStatusFilter);
      }
      params.set("limit", "100");
      const listPayload = await apiFetch<unknown>(`/v1/platform/kyc/reviews?${params.toString()}`, { accessToken });
      setKYCReviews(normalizeKYCReviewListPayload(listPayload));

      setKYCState({
        kind: "success",
        title: "KYC review обновлён",
        description: "Решение применено и список review синхронизирован.",
      });
    } catch (error) {
      setKYCState({
        kind: "error",
        title: "Не удалось применить решение",
        description: problemMessage(error, "Unknown KYC decision error"),
      });
    }
  }

  return (
    <>
      <Panel title="Администрирование пользователей" eyebrow="Platform / Users">
        <div className={`status-card ${accessState.kind === "error" ? "error" : accessState.kind === "success" ? "success" : "info"}`}>
          <strong>{accessState.title}</strong>
          <p className="status-copy">{accessState.description}</p>
        </div>
        {platformAccess ? (
          <div className="mini-card">
            <p className="muted">Account: {platformAccess.accountId}</p>
            <p className="muted">Bootstrap admin: {platformAccess.bootstrapAdmin ? "yes" : "no"}</p>
            <p className="muted">Effective roles: {(platformAccess.effectiveRoles || []).join(", ") || "none"}</p>
          </div>
        ) : null}
      </Panel>

      {isPlatformAdmin || canManageKYC ? (
        <>
        {isPlatformAdmin ? (
        <section className="split">
          <Panel title="Список пользователей" eyebrow="По выбранной организации">
            <div className={`status-card ${usersState.kind === "error" ? "error" : usersState.kind === "success" ? "success" : "info"}`}>
              <strong>{usersState.title}</strong>
              <p className="status-copy">{usersState.description}</p>
            </div>

            <div className="form-row">
              <label className="form-label" htmlFor="admin-users-org">
                Организация
              </label>
              <select
                id="admin-users-org"
                className="text-input"
                value={selectedOrganizationId}
                onChange={(event) => setSelectedOrganizationId(event.target.value)}
              >
                {organizations.length === 0 ? <option value="">Нет доступных организаций</option> : null}
                {organizations.map((item) => (
                  <option key={item.id} value={item.id}>
                    {item.name} ({item.slug})
                  </option>
                ))}
              </select>
            </div>

            {members.length > 0 ? (
              <div className="selection-list">
                {members.map((member) => (
                  <div key={member.membershipId} className="selection-card">
                    <strong>{member.displayName || member.email || member.accountId}</strong>
                    <span className="muted">accountId: {member.accountId}</span>
                    <span className="muted">zitadel_user_id: {member.zitadelUserId || "n/a"}</span>
                    {member.email ? <span className="muted">email: {member.email}</span> : null}
                    <span className="muted">role: {member.role}</span>
                    <span className="muted">
                      status: {member.isActive ? "active" : "inactive"} · added: {formatTimestamp(member.createdAt)}
                    </span>
                  </div>
                ))}
              </div>
            ) : null}
          </Panel>

          <Panel title="Верификация email" eyebrow="Force verify">
            <form className="form-grid" onSubmit={handleForceVerify}>
              <div className={`status-card ${verifyState.kind === "error" ? "error" : verifyState.kind === "success" ? "success" : "info"}`}>
                <strong>{verifyState.title}</strong>
                <p className="status-copy">{verifyState.description}</p>
              </div>
              <div className="mini-card">
                <p className="muted">
                  Для force-verify нужен raw ZITADEL user id (не accountId из CollabSphere). Возьмите его из ZITADEL admin UI или audit/log потока.
                </p>
              </div>
              <div className="form-row">
                <label className="form-label" htmlFor="zitadel-user-id">
                  ZITADEL user id
                </label>
                <input
                  id="zitadel-user-id"
                  className="text-input"
                  value={zitadelUserId}
                  onChange={(event) => setZitadelUserId(event.target.value)}
                  placeholder="364110152871182339"
                  required
                />
              </div>
              <div className="button-row">
                <button className="button primary" type="submit">
                  Force verify email
                </button>
              </div>
              {verifyResult ? <textarea className="code-block" readOnly value={JSON.stringify(verifyResult, null, 2)} /> : null}
            </form>
          </Panel>
        </section>
        ) : (
          <Panel title="Пользователи" eyebrow="Platform admin only">
            <p className="muted">
              Блок управления пользователями и force-verify email доступен роли `platform_admin` (или bootstrap admin).
            </p>
          </Panel>
        )}

        {canManageKYC ? (
        <Panel title="KYC верификация документов" eyebrow="Platform / KYC Reviews">
          <div className={`status-card ${kycState.kind === "error" ? "error" : kycState.kind === "success" ? "success" : "info"}`}>
            <strong>{kycState.title}</strong>
            <p className="status-copy">{kycState.description}</p>
          </div>

          <div className="form-row two">
            <div className="form-row">
              <label className="form-label" htmlFor="kyc-scope-filter">
                Scope
              </label>
              <select id="kyc-scope-filter" className="text-input" value={kycScopeFilter} onChange={(event) => setKYCScopeFilter(event.target.value)}>
                <option value="">Все</option>
                <option value="account">Users</option>
                <option value="organization">Organizations</option>
              </select>
            </div>
            <div className="form-row">
              <label className="form-label" htmlFor="kyc-status-filter">
                Status
              </label>
              <select id="kyc-status-filter" className="text-input" value={kycStatusFilter} onChange={(event) => setKYCStatusFilter(event.target.value)}>
                <option value="">Все</option>
                <option value="draft">draft</option>
                <option value="submitted">submitted</option>
                <option value="in_review">in_review</option>
                <option value="needs_info">needs_info</option>
                <option value="approved">approved</option>
                <option value="rejected">rejected</option>
              </select>
            </div>
          </div>

          <div className="split">
            <div className="selection-list kyc-review-list">
              {kycReviews.map((item) => (
                <button
                  key={item.reviewId}
                  type="button"
                  className={`selection-card ${item.reviewId === selectedReviewId ? "active" : ""}`}
                  onClick={() => setSelectedReviewId(item.reviewId)}
                >
                  <strong>{item.scope === "account" ? "User KYC" : "Organization KYC"}</strong>
                  <span className="muted">reviewId: {item.reviewId}</span>
                  <span className="muted">
                    status: <span className={kycStatusBadgeClass(item.status)}>{kycStatusLabel(item.status)}</span>
                  </span>
                  <span className="muted">updated: {formatTimestamp(item.updatedAt)}</span>
                </button>
              ))}
              {kycReviews.length === 0 ? <p className="muted">Нет записей по выбранным фильтрам.</p> : null}
            </div>

            <div className="form-grid">
              {selectedReview ? (
                <>
                  <div className="mini-card">
                    <p className="muted">Scope: {selectedReview.scope}</p>
                    <p className="muted">Subject: {selectedReview.subjectId}</p>
                    <p className="muted">
                      Status:{" "}
                      <span className={kycStatusBadgeClass(selectedReview.status)}>{kycStatusLabel(selectedReview.status)}</span>
                    </p>
                    <p className="muted">Legal name: {selectedReview.legalName || "n/a"}</p>
                    <p className="muted">Country: {selectedReview.countryCode || "n/a"}</p>
                    {selectedReview.documentNumber ? <p className="muted">Document number: {selectedReview.documentNumber}</p> : null}
                    {selectedReview.registrationNumber ? <p className="muted">Registration: {selectedReview.registrationNumber}</p> : null}
                    {selectedReview.taxId ? <p className="muted">Tax ID: {selectedReview.taxId}</p> : null}
                    {selectedReview.reviewNote ? <p className="muted">Review note: {selectedReview.reviewNote}</p> : null}
                  </div>

                  <form className="form-grid" onSubmit={handleKYCDecision}>
                    <div className="form-row">
                      <label className="form-label" htmlFor="kyc-decision">
                        Решение
                      </label>
                      <select id="kyc-decision" className="text-input" value={kycDecision} onChange={(event) => setKYCDecision(event.target.value as "approve" | "reject" | "request_info")}>
                        <option value="approve">approve</option>
                        <option value="reject">reject</option>
                        <option value="request_info">request_info</option>
                      </select>
                    </div>
                    <div className="form-row">
                      <label className="form-label" htmlFor="kyc-reason">
                        Комментарий
                      </label>
                      <textarea
                        id="kyc-reason"
                        className="textarea"
                        rows={4}
                        value={kycReason}
                        onChange={(event) => setKYCReason(event.target.value)}
                        placeholder="Опционально: причина отклонения или уточнение для needs_info"
                      />
                    </div>
                    <div className="button-row">
                      <button className="button primary" type="submit">
                        Применить решение
                      </button>
                    </div>
                  </form>
                </>
              ) : (
                <div className="empty-state">Выберите review слева для детализации и принятия решения.</div>
              )}
            </div>
          </div>
        </Panel>
        ) : null}
        </>
      ) : (
        <Panel title="Доступ ограничен" eyebrow="Admin only">
          <p className="muted">
            Эта страница доступна ролям `platform_admin`, `review_operator` или bootstrap admin. Если роль уже выдана, обновите сессию и откройте страницу снова.
          </p>
        </Panel>
      )}
    </>
  );
}
