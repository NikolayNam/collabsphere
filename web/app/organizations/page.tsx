"use client";

import { ChangeEvent, FormEvent, useEffect, useMemo, useState } from "react";

import { Panel } from "@/components/panel";
import { APIError, apiFetch } from "@/lib/api";
import { readStoredTokens } from "@/lib/auth";

type OrganizationDomain = {
  id: string;
  hostname: string;
  kind: string;
  isPrimary: boolean;
  isVerified: boolean;
};

type CreatedOrganization = {
  id: string;
  name: string;
  slug: string;
  isActive: boolean;
  domains?: OrganizationDomain[];
};

type ResolvedOrganization = {
  id: string;
  name: string;
  slug: string;
  domains?: Array<{
    hostname: string;
    kind: string;
    isPrimary: boolean;
    isVerified: boolean;
  }>;
};

type MyOrganization = {
  id: string;
  name: string;
  slug: string;
  logoObjectId?: string | null;
  isActive: boolean;
  createdAt: string;
  updatedAt?: string | null;
  membershipRole: string;
};

type OrganizationProfile = {
  id: string;
  name: string;
  slug: string;
  logoObjectId?: string | null;
  videoObjectIds?: string[];
  domains?: OrganizationDomain[];
  description?: string | null;
  website?: string | null;
  primaryEmail?: string | null;
  phone?: string | null;
  address?: string | null;
  industry?: string | null;
  isActive: boolean;
  createdAt: string;
  updatedAt?: string | null;
};

type ProfileDraft = {
  name: string;
  slug: string;
  description: string;
  website: string;
  primaryEmail: string;
  phone: string;
  address: string;
  industry: string;
};

type Status = {
  kind: "idle" | "working" | "success" | "error";
  title: string;
  description: string;
};

const initialMyOrganizationsState: Status = {
  kind: "idle",
  title: "Мои организации",
  description: "Этот блок читает `GET /v1/organizations/my` и даёт точку входа в реальный organization profile.",
};

const initialProfileState: Status = {
  kind: "idle",
  title: "Профиль организации",
  description: "Выберите organization слева, чтобы загрузить полный профиль через `GET /v1/organizations/{id}`.",
};

const initialSaveState: Status = {
  kind: "idle",
  title: "Сохранение профиля",
  description: "Форма пишет напрямую в `PATCH /v1/organizations/{id}` без второго backend-контракта.",
};

const initialLogoState: Status = {
  kind: "idle",
  title: "Logo upload",
  description: "Этот блок использует `POST /v1/organizations/{id}/logo` и сразу обновляет organization profile.",
};

const initialCreateState: Status = {
  kind: "idle",
  title: "Create organization",
  description: "Эта форма бьёт прямо в существующий backend endpoint `POST /v1/organizations`.",
};

const initialResolveState: Status = {
  kind: "idle",
  title: "Resolve by host",
  description: "Этот блок использует публичный `GET /v1/organizations/resolve-by-host` и не требует логина.",
};

function toDraft(profile: OrganizationProfile): ProfileDraft {
  return {
    name: profile.name || "",
    slug: profile.slug || "",
    description: profile.description || "",
    website: profile.website || "",
    primaryEmail: profile.primaryEmail || "",
    phone: profile.phone || "",
    address: profile.address || "",
    industry: profile.industry || "",
  };
}

function formatTimestamp(value?: string | null): string {
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

function problemMessage(error: unknown, fallback: string): string {
  if (error instanceof APIError) {
    return `${error.message}${error.code ? ` (${error.code})` : ""}`;
  }
  if (error instanceof Error) {
    return error.message;
  }
  return fallback;
}

export default function OrganizationsPage() {
  const [name, setName] = useState("Acme Foods");
  const [slug, setSlug] = useState("acme-foods");
  const [hostname, setHostname] = useState("acme.collabsphere.ru");
  const [hostToResolve, setHostToResolve] = useState("https://acme.collabsphere.ru/");
  const [createState, setCreateState] = useState<Status>(initialCreateState);
  const [resolveState, setResolveState] = useState<Status>(initialResolveState);
  const [created, setCreated] = useState<CreatedOrganization | null>(null);
  const [resolved, setResolved] = useState<ResolvedOrganization | null>(null);
  const [myOrganizations, setMyOrganizations] = useState<MyOrganization[]>([]);
  const [myOrganizationsState, setMyOrganizationsState] = useState<Status>(initialMyOrganizationsState);
  const [selectedOrganizationId, setSelectedOrganizationId] = useState("");
  const [profile, setProfile] = useState<OrganizationProfile | null>(null);
  const [profileDraft, setProfileDraft] = useState<ProfileDraft>({
    name: "",
    slug: "",
    description: "",
    website: "",
    primaryEmail: "",
    phone: "",
    address: "",
    industry: "",
  });
  const [profileState, setProfileState] = useState<Status>(initialProfileState);
  const [saveState, setSaveState] = useState<Status>(initialSaveState);
  const [logoState, setLogoState] = useState<Status>(initialLogoState);
  const [logoFile, setLogoFile] = useState<File | null>(null);
  const [listRefreshKey, setListRefreshKey] = useState(0);

  const accessToken = useMemo(() => readStoredTokens()?.accessToken || null, []);
  const selectedOrganization = myOrganizations.find((item) => item.id === selectedOrganizationId) || null;

  useEffect(() => {
    let cancelled = false;

    async function loadMyOrganizations() {
      if (!accessToken) {
        if (!cancelled) {
          setMyOrganizations([]);
          setSelectedOrganizationId("");
          setMyOrganizationsState({
            kind: "error",
            title: "Нет локальной сессии",
            description: "Сначала завершите login через /login, чтобы backend смог вернуть ваши memberships.",
          });
        }
        return;
      }

      if (!cancelled) {
        setMyOrganizationsState({
          kind: "working",
          title: "Загружаем ваши организации",
          description: "Backend читает active memberships текущего account и возвращает связанный список organizations.",
        });
      }

      try {
        const payload = await apiFetch<{ data?: MyOrganization[] }>("/v1/organizations/my", { accessToken });
        if (cancelled) {
          return;
        }
        const items = Array.isArray(payload.data) ? payload.data : [];
        setMyOrganizations(items);
        setSelectedOrganizationId((current) => {
          if (current && items.some((item) => item.id === current)) {
            return current;
          }
          return items[0]?.id || "";
        });
        setMyOrganizationsState({
          kind: "success",
          title: items.length > 0 ? "Организации загружены" : "Организаций пока нет",
          description:
            items.length > 0
              ? "Теперь ниже можно открыть полный профиль конкретной organization и редактировать основные поля."
              : "У аккаунта пока нет active memberships. Можно создать первую organization прямо на этой странице.",
        });
      } catch (error) {
        if (cancelled) {
          return;
        }
        setMyOrganizations([]);
        setSelectedOrganizationId("");
        setMyOrganizationsState({
          kind: "error",
          title: "Не удалось загрузить организации",
          description: problemMessage(error, "Unknown my organizations error"),
        });
      }
    }

    void loadMyOrganizations();
    return () => {
      cancelled = true;
    };
  }, [accessToken, listRefreshKey]);

  useEffect(() => {
    let cancelled = false;

    async function loadProfile() {
      if (!selectedOrganizationId) {
        setProfile(null);
        setProfileState(initialProfileState);
        setSaveState(initialSaveState);
        setLogoState(initialLogoState);
        return;
      }

      setProfileState({
        kind: "working",
        title: "Загружаем профиль",
        description: "Читаем полную organization card из `GET /v1/organizations/{id}`.",
      });

      try {
        const payload = await apiFetch<OrganizationProfile>(`/v1/organizations/${selectedOrganizationId}`, { accessToken });
        if (cancelled) {
          return;
        }
        setProfile(payload);
        setProfileDraft(toDraft(payload));
        setProfileState({
          kind: "success",
          title: "Профиль загружен",
          description: "Форма ниже уже редактирует реальные organization fields из backend domain model.",
        });
      } catch (error) {
        if (cancelled) {
          return;
        }
        setProfile(null);
        setProfileState({
          kind: "error",
          title: "Не удалось загрузить профиль",
          description: problemMessage(error, "Unknown organization profile error"),
        });
      }
    }

    void loadProfile();
    return () => {
      cancelled = true;
    };
  }, [accessToken, selectedOrganizationId]);

  function handleDraftChange(field: keyof ProfileDraft, value: string) {
    setProfileDraft((current) => ({ ...current, [field]: value }));
  }

  function handleLogoSelection(event: ChangeEvent<HTMLInputElement>) {
    setLogoFile(event.target.files?.[0] || null);
  }

  async function handleCreate(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!accessToken) {
      setCreateState({
        kind: "error",
        title: "Нет локальной сессии",
        description: "Сначала завершите login через /login, затем вернитесь к созданию organization.",
      });
      return;
    }

    setCreateState({
      kind: "working",
      title: "Создаём organization",
      description: "Backend автоматически создаст organization, owner membership и initial subdomain binding.",
    });

    try {
      const payload: CreatedOrganization = await apiFetch("/v1/organizations", {
        method: "POST",
        accessToken,
        bodyJSON: {
          name,
          slug,
          domains: hostname
            ? [
                {
                  hostname,
                  kind: "subdomain",
                  isPrimary: true,
                },
              ]
            : [],
        },
      });

      setCreated(payload);
      setSelectedOrganizationId(payload.id);
      setListRefreshKey((value) => value + 1);
      setCreateState({
        kind: "success",
        title: "Organization создана",
        description: "Новая organization выбрана для дальнейшего редактирования профиля.",
      });
    } catch (error) {
      setCreateState({
        kind: "error",
        title: "Создание не удалось",
        description: problemMessage(error, "Unknown organization create error"),
      });
    }
  }

  async function handleSaveProfile(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!accessToken || !selectedOrganizationId) {
      setSaveState({
        kind: "error",
        title: "Нет выбранной organization",
        description: "Сначала выберите organization из списка слева.",
      });
      return;
    }

    setSaveState({
      kind: "working",
      title: "Сохраняем профиль",
      description: "Форма пишет в существующий backend endpoint `PATCH /v1/organizations/{id}`.",
    });

    try {
      const payload = await apiFetch<OrganizationProfile>(`/v1/organizations/${selectedOrganizationId}`, {
        method: "PATCH",
        accessToken,
        bodyJSON: {
          name: profileDraft.name,
          slug: profileDraft.slug,
          description: profileDraft.description,
          website: profileDraft.website,
          primaryEmail: profileDraft.primaryEmail,
          phone: profileDraft.phone,
          address: profileDraft.address,
          industry: profileDraft.industry,
        },
      });
      setProfile(payload);
      setProfileDraft(toDraft(payload));
      setListRefreshKey((value) => value + 1);
      setSaveState({
        kind: "success",
        title: "Профиль сохранён",
        description: "Organization profile уже обновлён в backend и повторно синхронизирован во frontend.",
      });
    } catch (error) {
      setSaveState({
        kind: "error",
        title: "Сохранение не удалось",
        description: problemMessage(error, "Unknown organization update error"),
      });
    }
  }

  async function handleUploadLogo(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!accessToken || !selectedOrganizationId) {
      setLogoState({
        kind: "error",
        title: "Нет выбранной organization",
        description: "Сначала выберите organization, затем загрузите logo.",
      });
      return;
    }
    if (!logoFile) {
      setLogoState({
        kind: "error",
        title: "Файл не выбран",
        description: "Выберите logo image перед отправкой.",
      });
      return;
    }

    setLogoState({
      kind: "working",
      title: "Загружаем logo",
      description: "Файл идёт в `POST /v1/organizations/{id}/logo` как multipart/form-data.",
    });

    try {
      const formData = new FormData();
      formData.append("file", logoFile);
      const payload = await apiFetch<OrganizationProfile>(`/v1/organizations/${selectedOrganizationId}/logo`, {
        method: "POST",
        accessToken,
        body: formData,
      });
      setProfile(payload);
      setProfileDraft(toDraft(payload));
      setLogoFile(null);
      setListRefreshKey((value) => value + 1);
      setLogoState({
        kind: "success",
        title: "Logo обновлён",
        description: "Backend сразу привязал uploaded object к profile и вернул обновлённую organization card.",
      });
    } catch (error) {
      setLogoState({
        kind: "error",
        title: "Upload не удался",
        description: problemMessage(error, "Unknown logo upload error"),
      });
    }
  }

  async function handleResolve(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setResolveState({
      kind: "working",
      title: "Ищем organization по host",
      description: "Backend нормализует raw host/URL и пытается найти активный verified домен.",
    });

    try {
      const params = new URLSearchParams({ host: hostToResolve });
      const payload: ResolvedOrganization = await apiFetch(`/v1/organizations/resolve-by-host?${params.toString()}`);
      setResolved(payload);
      setResolveState({
        kind: "success",
        title: "Organization найдена",
        description: "Resolve-by-host вернул живую tenant запись backend.",
      });
    } catch (error) {
      setResolveState({
        kind: "error",
        title: "Resolve не удался",
        description: problemMessage(error, "Unknown organization resolve error"),
      });
      setResolved(null);
    }
  }

  return (
    <>
      <Panel title="Organizations workbench" eyebrow="Existing backend flows">
        <div className="mini-card">
          <h3>Что здесь уже реально работает</h3>
          <p className="muted">
            Страница больше не ограничивается create/resolve. Теперь она использует существующие backend profile endpoints:
            список организаций аккаунта, полная загрузка organization card, редактирование профиля и upload logo.
          </p>
        </div>
      </Panel>

      <Panel title="Мои организации" eyebrow="GET /v1/organizations/my">
        <div className={`status-card ${myOrganizationsState.kind === "error" ? "error" : myOrganizationsState.kind === "success" ? "success" : "info"}`}>
          <strong>{myOrganizationsState.title}</strong>
          <p className="status-copy">{myOrganizationsState.description}</p>
        </div>
        {myOrganizations.length > 0 ? (
          <div className="selection-list">
            {myOrganizations.map((item) => (
              <button
                key={item.id}
                type="button"
                className={`selection-card ${item.id === selectedOrganizationId ? "active" : ""}`}
                onClick={() => setSelectedOrganizationId(item.id)}
              >
                <strong>{item.name}</strong>
                <span className="muted">
                  <code>{item.slug}</code> · {item.membershipRole}
                </span>
                <span className="muted">Status: {item.isActive ? "active" : "archived"}</span>
                <span className="muted">
                  Updated: {formatTimestamp(item.updatedAt || item.createdAt)}
                </span>
              </button>
            ))}
          </div>
        ) : null}
      </Panel>

      <section className="split">
        <Panel title="Organization profile" eyebrow={selectedOrganization ? selectedOrganization.name : "Select organization"}>
          <div className={`status-card ${profileState.kind === "error" ? "error" : profileState.kind === "success" ? "success" : "info"}`}>
            <strong>{profileState.title}</strong>
            <p className="status-copy">{profileState.description}</p>
          </div>

          {profile ? (
            <>
              <div className="cards">
                <div className="mini-card">
                  <h3>Identity</h3>
                  <p className="muted">
                    <code>{profile.id}</code>
                  </p>
                  <p className="muted">Status: {profile.isActive ? "active" : "archived"}</p>
                  <p className="muted">Created: {formatTimestamp(profile.createdAt)}</p>
                </div>
                <div className="mini-card">
                  <h3>Branding</h3>
                  <p className="muted">Logo object: {profile.logoObjectId || "not attached"}</p>
                  <p className="muted">Videos: {profile.videoObjectIds?.length || 0}</p>
                </div>
              </div>

              <form className="form-grid" onSubmit={handleSaveProfile}>
                <div className={`status-card ${saveState.kind === "error" ? "error" : saveState.kind === "success" ? "success" : "info"}`}>
                  <strong>{saveState.title}</strong>
                  <p className="status-copy">{saveState.description}</p>
                </div>

                <div className="form-row two">
                  <div className="form-row">
                    <label className="form-label" htmlFor="profile-name">
                      Name
                    </label>
                    <input
                      id="profile-name"
                      className="text-input"
                      value={profileDraft.name}
                      onChange={(event) => handleDraftChange("name", event.target.value)}
                      required
                    />
                  </div>
                  <div className="form-row">
                    <label className="form-label" htmlFor="profile-slug">
                      Slug
                    </label>
                    <input
                      id="profile-slug"
                      className="text-input"
                      value={profileDraft.slug}
                      onChange={(event) => handleDraftChange("slug", event.target.value)}
                      required
                    />
                  </div>
                </div>

                <div className="form-row">
                  <label className="form-label" htmlFor="profile-description">
                    Description
                  </label>
                  <textarea
                    id="profile-description"
                    className="textarea"
                    value={profileDraft.description}
                    onChange={(event) => handleDraftChange("description", event.target.value)}
                    rows={5}
                  />
                </div>

                <div className="form-row two">
                  <div className="form-row">
                    <label className="form-label" htmlFor="profile-website">
                      Website
                    </label>
                    <input
                      id="profile-website"
                      className="text-input"
                      value={profileDraft.website}
                      onChange={(event) => handleDraftChange("website", event.target.value)}
                      placeholder="https://example.com"
                    />
                  </div>
                  <div className="form-row">
                    <label className="form-label" htmlFor="profile-primary-email">
                      Primary email
                    </label>
                    <input
                      id="profile-primary-email"
                      className="text-input"
                      value={profileDraft.primaryEmail}
                      onChange={(event) => handleDraftChange("primaryEmail", event.target.value)}
                      placeholder="contact@example.com"
                    />
                  </div>
                </div>

                <div className="form-row two">
                  <div className="form-row">
                    <label className="form-label" htmlFor="profile-phone">
                      Phone
                    </label>
                    <input
                      id="profile-phone"
                      className="text-input"
                      value={profileDraft.phone}
                      onChange={(event) => handleDraftChange("phone", event.target.value)}
                      placeholder="+7 999 000 00 00"
                    />
                  </div>
                  <div className="form-row">
                    <label className="form-label" htmlFor="profile-industry">
                      Industry
                    </label>
                    <input
                      id="profile-industry"
                      className="text-input"
                      value={profileDraft.industry}
                      onChange={(event) => handleDraftChange("industry", event.target.value)}
                      placeholder="Wholesale"
                    />
                  </div>
                </div>

                <div className="form-row">
                  <label className="form-label" htmlFor="profile-address">
                    Address
                  </label>
                  <textarea
                    id="profile-address"
                    className="textarea"
                    value={profileDraft.address}
                    onChange={(event) => handleDraftChange("address", event.target.value)}
                    rows={4}
                  />
                </div>

                <div className="button-row">
                  <button className="button primary" type="submit">
                    Сохранить профиль
                  </button>
                </div>
              </form>

              <form className="form-grid" onSubmit={handleUploadLogo}>
                <div className={`status-card ${logoState.kind === "error" ? "error" : logoState.kind === "success" ? "success" : "info"}`}>
                  <strong>{logoState.title}</strong>
                  <p className="status-copy">{logoState.description}</p>
                </div>
                <div className="form-row">
                  <label className="form-label" htmlFor="profile-logo-file">
                    Logo image
                  </label>
                  <input id="profile-logo-file" type="file" accept="image/*" onChange={handleLogoSelection} />
                </div>
                <div className="button-row">
                  <button className="button secondary" type="submit">
                    Upload logo
                  </button>
                </div>
              </form>

              <div className="mini-card">
                <h3>Domains</h3>
                {profile.domains && profile.domains.length > 0 ? (
                  <div className="domain-list">
                    {profile.domains.map((domain) => (
                      <div key={domain.id} className="inline-panel">
                        <strong>{domain.hostname}</strong>
                        <p className="muted">
                          {domain.kind} · {domain.isPrimary ? "primary" : "secondary"} · {domain.isVerified ? "verified" : "pending"}
                        </p>
                      </div>
                    ))}
                  </div>
                ) : (
                  <p className="muted">Для этой organization пока не настроены домены.</p>
                )}
              </div>
            </>
          ) : (
            <div className="mini-card">
              <h3>Профиль ещё не выбран</h3>
              <p className="muted">Выберите organization из списка выше, чтобы открыть её profile editor.</p>
            </div>
          )}
        </Panel>

        <div className="page-grid">
          <Panel title="Create organization" eyebrow="POST /v1/organizations">
            <div className={`status-card ${createState.kind === "error" ? "error" : createState.kind === "success" ? "success" : "info"}`}>
              <strong>{createState.title}</strong>
              <p className="status-copy">{createState.description}</p>
            </div>
            <form className="form-grid" onSubmit={handleCreate}>
              <div className="form-row">
                <label className="form-label" htmlFor="org-name">
                  Name
                </label>
                <input id="org-name" className="text-input" value={name} onChange={(event) => setName(event.target.value)} required />
              </div>
              <div className="form-row two">
                <div className="form-row">
                  <label className="form-label" htmlFor="org-slug">
                    Slug
                  </label>
                  <input id="org-slug" className="text-input" value={slug} onChange={(event) => setSlug(event.target.value)} required />
                </div>
                <div className="form-row">
                  <label className="form-label" htmlFor="org-hostname">
                    Primary subdomain
                  </label>
                  <input
                    id="org-hostname"
                    className="text-input"
                    value={hostname}
                    onChange={(event) => setHostname(event.target.value)}
                    placeholder="acme.collabsphere.ru"
                  />
                </div>
              </div>
              <div className="button-row">
                <button className="button primary" type="submit">
                  Create
                </button>
              </div>
            </form>
            {created ? <textarea className="code-block" readOnly value={JSON.stringify(created, null, 2)} /> : null}
          </Panel>

          <Panel title="Resolve by host" eyebrow="GET /v1/organizations/resolve-by-host">
            <div className={`status-card ${resolveState.kind === "error" ? "error" : resolveState.kind === "success" ? "success" : "info"}`}>
              <strong>{resolveState.title}</strong>
              <p className="status-copy">{resolveState.description}</p>
            </div>
            <form className="form-grid" onSubmit={handleResolve}>
              <div className="form-row">
                <label className="form-label" htmlFor="resolve-host">
                  Host or URL
                </label>
                <input
                  id="resolve-host"
                  className="text-input"
                  value={hostToResolve}
                  onChange={(event) => setHostToResolve(event.target.value)}
                  placeholder="https://acme.collabsphere.ru/"
                  required
                />
              </div>
              <div className="button-row">
                <button className="button secondary" type="submit">
                  Resolve host
                </button>
              </div>
            </form>
            {resolved ? <textarea className="code-block" readOnly value={JSON.stringify(resolved, null, 2)} /> : null}
          </Panel>
        </div>
      </section>
    </>
  );
}
