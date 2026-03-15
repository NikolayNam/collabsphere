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

type OrganizationKYCDocument = {
  id: string;
  organizationId: string;
  objectId: string;
  documentType: string;
  title: string;
  status?: string;
  reviewNote?: string;
  reviewerAccountId?: string;
  createdAt: string;
  updatedAt?: string;
  reviewedAt?: string;
};

type OrganizationKYCProfile = {
  organizationId: string;
  status?: string;
  legalName?: string;
  countryCode?: string;
  registrationNumber?: string;
  taxId?: string;
  reviewNote?: string;
  reviewerAccountId?: string;
  submittedAt?: string;
  reviewedAt?: string;
  createdAt: string;
  updatedAt: string;
  documents: OrganizationKYCDocument[];
};

type ProductCategory = {
  id: string;
  organizationId: string;
  parentId?: string | null;
  status: string;
  code: string;
  name: string;
  sortOrder: number;
  createdAt: string;
};

type ProductItem = {
  id: string;
  organizationId: string;
  categoryId?: string | null;
  status: string;
  name: string;
  description?: string | null;
  sku?: string | null;
  priceAmount?: string | null;
  currencyCode?: string | null;
  isActive: boolean;
  createdAt: string;
};

type OrganizationFileItem = {
  objectId: string;
  fileName: string;
  contentType?: string | null;
  sizeBytes: number;
  createdAt: string;
  sourceType: string;
  sourceId?: string | null;
};

type CooperationApplication = {
  organizationId: string;
  priceListObjectId?: string | null;
  priceListStatus?: string;
  status?: string;
  updatedAt?: string | null;
};

const CATALOG_STATUSES = ["draft", "validating", "verified", "published", "withdrawn", "archived"] as const;

function catalogStatusLabel(status?: string): string {
  const value = (status || "").trim().toLowerCase();
  switch (value) {
    case "draft":
      return "Черновик";
    case "validating":
      return "Валидируется";
    case "verified":
      return "Подтвержден";
    case "published":
      return "Опубликован";
    case "withdrawn":
      return "Отозван";
    case "archived":
      return "Архив";
    default:
      return status || "unknown";
  }
}

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

function kycDocumentStatusLabel(status?: string): string {
  const value = (status || "").trim().toLowerCase();
  switch (value) {
    case "pending_upload":
      return "На подготовке";
    case "uploaded":
      return "Загружен";
    case "verified":
      return "Подтвержден";
    case "rejected":
      return "Отклонен";
    default:
      return kycStatusLabel(status);
  }
}

function kycDocumentStatusBadgeClass(status?: string): string {
  const value = (status || "").trim().toLowerCase();
  switch (value) {
    case "pending_upload":
      return "status-badge pending-upload";
    case "uploaded":
      return "status-badge uploaded";
    case "verified":
      return "status-badge verified";
    case "rejected":
      return "status-badge rejected";
    default:
      return "status-badge";
  }
}

type KYCQueueBucket = "draft" | "submitted" | "in_review" | "verified" | "rejected";

function getKYCQueueBucket(documentStatus?: string, profileStatus?: string): KYCQueueBucket {
  const doc = (documentStatus || "").trim().toLowerCase();
  const profile = (profileStatus || "").trim().toLowerCase();
  if (doc === "pending_upload") {
    return "draft";
  }
  if (doc === "verified") {
    return "verified";
  }
  if (doc === "rejected") {
    return "rejected";
  }
  if (doc === "uploaded" && profile === "in_review") {
    return "in_review";
  }
  if (doc === "uploaded") {
    return "submitted";
  }
  return "submitted";
}

function kycQueueBucketLabel(bucket: KYCQueueBucket): string {
  switch (bucket) {
    case "draft":
      return "Черновики";
    case "submitted":
      return "Отправлено на верификацию";
    case "in_review":
      return "Проходит проверку";
    case "verified":
      return "Подтверждено";
    case "rejected":
      return "Отклонено";
    default:
      return bucket;
  }
}

function isReviewSubmissionLocked(status?: string): boolean {
  const value = (status || "").trim().toLowerCase();
  return value === "approved";
}

function kycSubmitActionLabel(status?: string): string {
  const value = (status || "").trim().toLowerCase();
  if (value === "needs_info" || value === "rejected") {
    return "Отправить повторную заявку";
  }
  if (value === "submitted" || value === "in_review") {
    return "Отправить дополнительный документ";
  }
  return "Submit for review";
}

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

type OrganizationSection = "profile" | "uploads" | "catalog" | "kyc" | "roles";

type OrganizationRole = {
  id: string;
  organizationId: string;
  code: string;
  name: string;
  description?: string;
  baseRole: string;
  isSystem: boolean;
  createdAt?: string;
  updatedAt?: string;
  deletedAt?: string | null;
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

const initialCategoriesUploadState: Status = {
  kind: "idle",
  title: "Категории",
  description: "Загрузите CSV для импорта категорий каталога через `POST /v1/organizations/{id}/product-imports/upload`.",
};

const initialProductsUploadState: Status = {
  kind: "idle",
  title: "Продукция",
  description: "Загрузите CSV для импорта продукции через `POST /v1/organizations/{id}/product-imports/upload`.",
};

const initialPriceListUploadState: Status = {
  kind: "idle",
  title: "Прайс-лист",
  description: "Загрузите прайс-лист организации через `POST /v1/organizations/{id}/cooperation-application/price-list`.",
};

const initialCatalogState: Status = {
  kind: "idle",
  title: "Каталог",
  description: "Списки категорий и продукции загружаются из `/v1/organizations/{id}/product-categories` и `/v1/organizations/{id}/products`.",
};

const initialPriceListsState: Status = {
  kind: "idle",
  title: "Прайс-листы",
  description: "Список файлов организации читается из `/v1/organizations/{id}/files`.",
};

const initialKYCState: Status = {
  kind: "idle",
  title: "KYC",
  description: "KYC профиль организации загружается из `GET /v1/organizations/{id}/kyc`.",
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

function downloadCSVTemplate(fileName: string, header: string[], rows: string[][]) {
  const lines = [header.join(","), ...rows.map((row) => row.map((cell) => `"${String(cell).replaceAll('"', '""')}"`).join(","))];
  const csv = "\uFEFF" + lines.join("\n");
  const blob = new Blob([csv], { type: "text/csv;charset=utf-8" });
  const url = URL.createObjectURL(blob);
  const link = document.createElement("a");
  link.href = url;
  link.download = fileName;
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
  URL.revokeObjectURL(url);
}

function hasCSVExtension(fileName: string): boolean {
  return fileName.trim().toLowerCase().endsWith(".csv");
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
  const [categoriesFile, setCategoriesFile] = useState<File | null>(null);
  const [productsFile, setProductsFile] = useState<File | null>(null);
  const [priceListFile, setPriceListFile] = useState<File | null>(null);
  const [categoriesUploadState, setCategoriesUploadState] = useState<Status>(initialCategoriesUploadState);
  const [productsUploadState, setProductsUploadState] = useState<Status>(initialProductsUploadState);
  const [priceListUploadState, setPriceListUploadState] = useState<Status>(initialPriceListUploadState);
  const [categoriesUploadResult, setCategoriesUploadResult] = useState<unknown | null>(null);
  const [productsUploadResult, setProductsUploadResult] = useState<unknown | null>(null);
  const [priceListUploadResult, setPriceListUploadResult] = useState<unknown | null>(null);
  const [catalogState, setCatalogState] = useState<Status>(initialCatalogState);
  const [categories, setCategories] = useState<ProductCategory[]>([]);
  const [products, setProducts] = useState<ProductItem[]>([]);
  const [priceListFiles, setPriceListFiles] = useState<OrganizationFileItem[]>([]);
  const [priceListsState, setPriceListsState] = useState<Status>(initialPriceListsState);
  const [cooperationApplication, setCooperationApplication] = useState<CooperationApplication | null>(null);
  const [newCategoryCode, setNewCategoryCode] = useState("");
  const [newCategoryName, setNewCategoryName] = useState("");
  const [newCategoryStatus, setNewCategoryStatus] = useState("draft");
  const [newProductName, setNewProductName] = useState("");
  const [newProductCategoryId, setNewProductCategoryId] = useState("");
  const [newProductPriceAmount, setNewProductPriceAmount] = useState("");
  const [newProductCurrency, setNewProductCurrency] = useState("RUB");
  const [newProductStatus, setNewProductStatus] = useState("draft");
  const [editingCategory, setEditingCategory] = useState<Record<string, { code: string; name: string; sortOrder: number; status: string }>>({});
  const [editingProduct, setEditingProduct] = useState<Record<string, { name: string; categoryId: string; status: string; priceAmount: string; currencyCode: string; isActive: boolean }>>(
    {}
  );
  const [kycState, setKYCState] = useState<Status>(initialKYCState);
  const [kycProfile, setKYCProfile] = useState<OrganizationKYCProfile | null>(null);
  const [kycDraft, setKYCDraft] = useState({
    status: "draft",
    legalName: "",
    countryCode: "",
    registrationNumber: "",
    taxId: "",
  });
  const [kycFiles, setKYCFiles] = useState<File[]>([]);
  const [kycDocumentType, setKYCDocumentType] = useState("registration_document");
  const [kycDocumentTitle, setKYCDocumentTitle] = useState("");
  const [organizationSection, setOrganizationSection] = useState<OrganizationSection>("profile");
  const [listRefreshKey, setListRefreshKey] = useState(0);
  const [roles, setRoles] = useState<OrganizationRole[]>([]);
  const [rolesState, setRolesState] = useState<{ kind: "idle" | "working" | "success" | "error"; title: string; description: string }>({
    kind: "idle",
    title: "",
    description: "",
  });
  const [roleForm, setRoleForm] = useState<{ code: string; name: string; description: string; baseRole: string } | null>(null);
  const [editingRoleId, setEditingRoleId] = useState<string | null>(null);

  const accessToken = useMemo(() => readStoredTokens()?.accessToken || null, []);
  const selectedOrganization = myOrganizations.find((item) => item.id === selectedOrganizationId) || null;
  const canManageCatalog = useMemo(() => {
    const role = (selectedOrganization?.membershipRole || "").toLowerCase();
    return role === "owner" || role === "admin" || role === "manager";
  }, [selectedOrganization?.membershipRole]);
  const canManageOrganization = useMemo(() => {
    const role = (selectedOrganization?.membershipRole || "").toLowerCase();
    return role === "owner" || role === "admin";
  }, [selectedOrganization?.membershipRole]);

  useEffect(() => {
    setOrganizationSection("uploads");
  }, [selectedOrganizationId]);

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
        setKYCProfile(null);
        setCategories([]);
        setProducts([]);
        setPriceListFiles([]);
        setCooperationApplication(null);
        setCatalogState(initialCatalogState);
        setPriceListsState(initialPriceListsState);
        setKYCState(initialKYCState);
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
        void refreshOrganizationKYC(payload.id);
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

  useEffect(() => {
    let cancelled = false;
    async function loadCatalog() {
      if (!accessToken || !selectedOrganizationId) {
        return;
      }
      setCatalogState({
        kind: "working",
        title: "Каталог",
        description: "Загружаем категории и продукцию организации...",
      });
      try {
        const [categoriesPayload, productsPayload] = await Promise.all([
          apiFetch<{ items?: ProductCategory[] }>(`/v1/organizations/${selectedOrganizationId}/product-categories`, { accessToken }),
          apiFetch<{ items?: ProductItem[] }>(`/v1/organizations/${selectedOrganizationId}/products`, { accessToken }),
        ]);
        if (cancelled) {
          return;
        }
        const loadedCategories = Array.isArray(categoriesPayload.items) ? categoriesPayload.items : [];
        const loadedProducts = Array.isArray(productsPayload.items) ? productsPayload.items : [];
        setCategories(loadedCategories);
        setProducts(loadedProducts);

        setEditingCategory(
          Object.fromEntries(
            loadedCategories.map((item) => [item.id, { code: item.code || "", name: item.name || "", sortOrder: item.sortOrder || 0, status: item.status || "draft" }])
          )
        );
        setEditingProduct(
          Object.fromEntries(
            loadedProducts.map((item) => [
              item.id,
              {
                name: item.name || "",
                categoryId: item.categoryId || "",
                status: item.status || "draft",
                priceAmount: item.priceAmount || "",
                currencyCode: item.currencyCode || "RUB",
                isActive: Boolean(item.isActive),
              },
            ])
          )
        );
        setCatalogState({
          kind: "success",
          title: "Каталог",
          description: `Загружено: категорий ${loadedCategories.length}, продукции ${loadedProducts.length}.`,
        });
      } catch (error) {
        if (cancelled) {
          return;
        }
        setCatalogState({
          kind: "error",
          title: "Каталог",
          description: problemMessage(error, "Не удалось загрузить категории и продукцию"),
        });
      }
    }

    void loadCatalog();
    return () => {
      cancelled = true;
    };
  }, [accessToken, selectedOrganizationId, listRefreshKey]);

  useEffect(() => {
    let cancelled = false;
    async function loadRoles() {
      if (!accessToken || !selectedOrganizationId || organizationSection !== "roles") {
        setRoles([]);
        setRolesState({ kind: "idle", title: "", description: "" });
        return;
      }
      setRolesState({ kind: "working", title: "Роли", description: "Загружаем роли организации..." });
      try {
        const payload = await apiFetch<{ body?: { roles?: OrganizationRole[] }; roles?: OrganizationRole[] }>(
          `/v1/organizations/${selectedOrganizationId}/roles`,
          { accessToken }
        );
        if (cancelled) return;
        const list = payload?.body?.roles ?? payload?.roles ?? [];
        setRoles(Array.isArray(list) ? list : []);
        setRolesState({ kind: "success", title: "Роли", description: `Загружено ролей: ${Array.isArray(list) ? list.length : 0}.` });
      } catch (error) {
        if (cancelled) return;
        setRoles([]);
        setRolesState({ kind: "error", title: "Роли", description: problemMessage(error, "Не удалось загрузить роли") });
      }
    }
    void loadRoles();
    return () => { cancelled = true; };
  }, [accessToken, selectedOrganizationId, organizationSection, listRefreshKey]);

  useEffect(() => {
    if (organizationSection !== "roles") {
      setRoleForm(null);
      setEditingRoleId(null);
    }
  }, [organizationSection]);

  useEffect(() => {
    let cancelled = false;

    async function loadPriceLists() {
      if (!accessToken || !selectedOrganizationId) {
        setPriceListFiles([]);
        setCooperationApplication(null);
        setPriceListsState(initialPriceListsState);
        return;
      }

      setPriceListsState({
        kind: "working",
        title: "Прайс-листы",
        description: "Загружаем список прайс-листов организации...",
      });

      try {
        const [filesPayload, cooperationPayload] = await Promise.all([
          apiFetch<{ body?: { items?: OrganizationFileItem[] }; items?: OrganizationFileItem[] }>(
            `/v1/organizations/${selectedOrganizationId}/files`,
            { accessToken }
          ),
          apiFetch<{ body?: CooperationApplication }>(`/v1/organizations/${selectedOrganizationId}/cooperation-application`, { accessToken }),
        ]);

        if (cancelled) {
          return;
        }

        const fileItems = Array.isArray(filesPayload.body?.items) ? filesPayload.body.items : Array.isArray(filesPayload.items) ? filesPayload.items : [];
        const onlyPriceLists = fileItems
          .filter((item) => (item.sourceType || "").toLowerCase() === "cooperation_price_list")
          .sort((a, b) => (a.createdAt < b.createdAt ? 1 : -1));
        const cooperation = cooperationPayload.body || null;

        setPriceListFiles(onlyPriceLists);
        setCooperationApplication(cooperation);
        setPriceListsState({
          kind: "success",
          title: "Прайс-листы",
          description: `Найдено прайс-листов: ${onlyPriceLists.length}.`,
        });
      } catch (error) {
        if (cancelled) {
          return;
        }
        if (error instanceof APIError && error.status === 404) {
          setPriceListFiles([]);
          setCooperationApplication(null);
          setPriceListsState({
            kind: "success",
            title: "Прайс-листы",
            description: "Прайс-листы пока не загружены.",
          });
          return;
        }
        setPriceListFiles([]);
        setCooperationApplication(null);
        setPriceListsState({
          kind: "error",
          title: "Прайс-листы",
          description: problemMessage(error, "Не удалось загрузить прайс-листы"),
        });
      }
    }

    void loadPriceLists();
    return () => {
      cancelled = true;
    };
  }, [accessToken, selectedOrganizationId, priceListUploadResult, listRefreshKey]);

  async function refreshCatalog() {
    if (!accessToken || !selectedOrganizationId) {
      return;
    }
    const [categoriesPayload, productsPayload] = await Promise.all([
      apiFetch<{ items?: ProductCategory[] }>(`/v1/organizations/${selectedOrganizationId}/product-categories`, { accessToken }),
      apiFetch<{ items?: ProductItem[] }>(`/v1/organizations/${selectedOrganizationId}/products`, { accessToken }),
    ]);
    const loadedCategories = Array.isArray(categoriesPayload.items) ? categoriesPayload.items : [];
    const loadedProducts = Array.isArray(productsPayload.items) ? productsPayload.items : [];
    setCategories(loadedCategories);
    setProducts(loadedProducts);
    setEditingCategory(
      Object.fromEntries(
        loadedCategories.map((item) => [item.id, { code: item.code || "", name: item.name || "", sortOrder: item.sortOrder || 0, status: item.status || "draft" }])
      )
    );
    setEditingProduct(
      Object.fromEntries(
        loadedProducts.map((item) => [
          item.id,
          {
            name: item.name || "",
            categoryId: item.categoryId || "",
            status: item.status || "draft",
            priceAmount: item.priceAmount || "",
            currencyCode: item.currencyCode || "RUB",
            isActive: Boolean(item.isActive),
          },
        ])
      )
    );
  }

  function handleDraftChange(field: keyof ProfileDraft, value: string) {
    setProfileDraft((current) => ({ ...current, [field]: value }));
  }

  function handleLogoSelection(event: ChangeEvent<HTMLInputElement>) {
    setLogoFile(event.target.files?.[0] || null);
  }

  function handleCategoriesSelection(event: ChangeEvent<HTMLInputElement>) {
    setCategoriesFile(event.target.files?.[0] || null);
  }

  function handleProductsSelection(event: ChangeEvent<HTMLInputElement>) {
    setProductsFile(event.target.files?.[0] || null);
  }

  function handlePriceListSelection(event: ChangeEvent<HTMLInputElement>) {
    setPriceListFile(event.target.files?.[0] || null);
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

  async function handleUploadCategories(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!accessToken || !selectedOrganizationId) {
      setCategoriesUploadState({
        kind: "error",
        title: "Категории",
        description: "Сначала выберите organization и убедитесь, что есть локальная сессия.",
      });
      return;
    }
    if (!categoriesFile) {
      setCategoriesUploadState({
        kind: "error",
        title: "Категории",
        description: "Выберите CSV-файл с категориями перед отправкой.",
      });
      return;
    }
    if (!hasCSVExtension(categoriesFile.name)) {
      setCategoriesUploadState({
        kind: "error",
        title: "Категории",
        description: "Для импорта категорий нужен CSV-файл (.csv). Скачайте шаблон ниже и заполните его.",
      });
      return;
    }

    setCategoriesUploadState({
      kind: "working",
      title: "Категории",
      description: "Импортируем файл в каталог организации. Backend выполнит upsert категорий/продукции из CSV.",
    });
    try {
      const formData = new FormData();
      formData.append("file", categoriesFile);
      const payload = await apiFetch(`/v1/organizations/${selectedOrganizationId}/product-imports/upload`, {
        method: "POST",
        accessToken,
        body: formData,
      });
      setCategoriesUploadResult(payload);
      setCategoriesFile(null);
      setCategoriesUploadState({
        kind: "success",
        title: "Категории",
        description: "Файл принят, import batch создан и обработан.",
      });
    } catch (error) {
      setCategoriesUploadState({
        kind: "error",
        title: "Категории",
        description: problemMessage(error, "Не удалось загрузить CSV категорий"),
      });
    }
  }

  async function handleUploadProducts(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!accessToken || !selectedOrganizationId) {
      setProductsUploadState({
        kind: "error",
        title: "Продукция",
        description: "Сначала выберите organization и убедитесь, что есть локальная сессия.",
      });
      return;
    }
    if (!productsFile) {
      setProductsUploadState({
        kind: "error",
        title: "Продукция",
        description: "Выберите CSV-файл с продукцией перед отправкой.",
      });
      return;
    }
    if (!hasCSVExtension(productsFile.name)) {
      setProductsUploadState({
        kind: "error",
        title: "Продукция",
        description: "Для импорта продукции нужен CSV-файл (.csv). Используйте шаблон ниже.",
      });
      return;
    }

    setProductsUploadState({
      kind: "working",
      title: "Продукция",
      description: "Импортируем файл в каталог организации. Backend выполнит upsert категорий/продукции из CSV.",
    });
    try {
      const formData = new FormData();
      formData.append("file", productsFile);
      const payload = await apiFetch(`/v1/organizations/${selectedOrganizationId}/product-imports/upload`, {
        method: "POST",
        accessToken,
        body: formData,
      });
      setProductsUploadResult(payload);
      setProductsFile(null);
      setProductsUploadState({
        kind: "success",
        title: "Продукция",
        description: "Файл принят, import batch создан и обработан.",
      });
    } catch (error) {
      setProductsUploadState({
        kind: "error",
        title: "Продукция",
        description: problemMessage(error, "Не удалось загрузить CSV продукции"),
      });
    }
  }

  async function handleUploadPriceList(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!accessToken || !selectedOrganizationId) {
      setPriceListUploadState({
        kind: "error",
        title: "Прайс-лист",
        description: "Сначала выберите organization и убедитесь, что есть локальная сессия.",
      });
      return;
    }
    if (!priceListFile) {
      setPriceListUploadState({
        kind: "error",
        title: "Прайс-лист",
        description: "Выберите файл прайс-листа перед отправкой.",
      });
      return;
    }

    setPriceListUploadState({
      kind: "working",
      title: "Прайс-лист",
      description: "Загружаем прайс-лист в cooperation-application организации.",
    });
    try {
      const formData = new FormData();
      formData.append("file", priceListFile);
      const payload = await apiFetch(`/v1/organizations/${selectedOrganizationId}/cooperation-application/price-list`, {
        method: "POST",
        accessToken,
        body: formData,
      });
      setPriceListUploadResult(payload);
      setPriceListFile(null);
      setPriceListUploadState({
        kind: "success",
        title: "Прайс-лист",
        description: "Прайс-лист успешно загружен и привязан к cooperation application.",
      });
    } catch (error) {
      setPriceListUploadState({
        kind: "error",
        title: "Прайс-лист",
        description: problemMessage(error, "Не удалось загрузить прайс-лист"),
      });
    }
  }

  async function handleDownloadCurrentPriceList() {
    if (!accessToken || !selectedOrganizationId) {
      return;
    }
    setPriceListsState({
      kind: "working",
      title: "Прайс-листы",
      description: "Готовим ссылку для скачивания текущего прайс-листа...",
    });
    try {
      const payload = await apiFetch<{
        downloadUrl?: string;
        body?: { downloadUrl?: string };
      }>(`/v1/organizations/${selectedOrganizationId}/cooperation-application/price-list/download`, { accessToken });
      const downloadUrl = payload.downloadUrl || payload.body?.downloadUrl;
      if (!downloadUrl) {
        throw new Error("downloadUrl is empty");
      }
      window.open(downloadUrl, "_blank", "noopener,noreferrer");
      setPriceListsState({
        kind: "success",
        title: "Прайс-листы",
        description: "Ссылка для скачивания открыта в новой вкладке.",
      });
    } catch (error) {
      setPriceListsState({
        kind: "error",
        title: "Прайс-листы",
        description: problemMessage(error, "Не удалось получить ссылку для скачивания прайс-листа"),
      });
    }
  }

  async function handleUpdatePriceListStatus(nextStatus: string) {
    if (!accessToken || !selectedOrganizationId) {
      return;
    }
    setPriceListsState({
      kind: "working",
      title: "Прайс-листы",
      description: "Обновляем статус прайс-листа...",
    });
    try {
      await apiFetch(`/v1/organizations/${selectedOrganizationId}/cooperation-application`, {
        method: "PATCH",
        accessToken,
        bodyJSON: {
          priceListStatus: nextStatus,
        },
      });
      setCooperationApplication((current) => ({ ...(current || { organizationId: selectedOrganizationId }), priceListStatus: nextStatus }));
      setPriceListsState({
        kind: "success",
        title: "Прайс-листы",
        description: "Статус прайс-листа обновлен.",
      });
    } catch (error) {
      setPriceListsState({
        kind: "error",
        title: "Прайс-листы",
        description: problemMessage(error, "Не удалось обновить статус прайс-листа"),
      });
    }
  }

  const [publishAllState, setPublishAllState] = useState<Status>({ kind: "idle", title: "", description: "" });

  const canAutoPublish = useMemo(() => {
    if (!canManageCatalog) return false;
    const allCategoriesVerified =
      categories.length === 0 ||
      categories.every((c) => {
        const s = (c.status || "").toLowerCase();
        return s === "verified" || s === "published";
      });
    const allProductsVerified =
      products.length === 0 ||
      products.every((p) => {
        const s = (p.status || "").toLowerCase();
        return s === "verified" || s === "published";
      });
    const priceListOk =
      !cooperationApplication?.priceListObjectId ||
      (() => {
        const s = (cooperationApplication?.priceListStatus || "").toLowerCase();
        return s === "verified" || s === "published";
      })();
    const hasSomething = categories.length > 0 || products.length > 0 || !!cooperationApplication?.priceListObjectId;
    return hasSomething && allCategoriesVerified && allProductsVerified && priceListOk;
  }, [canManageCatalog, categories, products, cooperationApplication]);

  async function handlePublishAllCatalog() {
    if (!accessToken || !selectedOrganizationId || !canManageCatalog) {
      return;
    }
    setPublishAllState({
      kind: "working",
      title: "Автопубликация",
      description: "Публикуем категории, продукцию и прайс-лист...",
    });
    try {
      await apiFetch(`/v1/organizations/${selectedOrganizationId}/catalog/publish-all`, {
        method: "POST",
        accessToken,
      });
      setPublishAllState({
        kind: "success",
        title: "Автопубликация",
        description: "Всё успешно опубликовано.",
      });
      setListRefreshKey((k) => k + 1);
      void refreshCatalog();
    } catch (error) {
      setPublishAllState({
        kind: "error",
        title: "Автопубликация",
        description: problemMessage(error, "Не удалось опубликовать каталог"),
      });
    }
  }

  function handleDownloadCategoriesTemplate() {
    downloadCSVTemplate(
      "categories_template.csv",
      [
        "categoryCode",
        "categoryName",
        "categoryParentCode",
        "categorySortOrder",
        "productName",
        "description",
        "sku",
        "priceAmount",
        "currencyCode",
        "isActive",
      ],
      [
        ["milk", "Молочная продукция", "", "10", "", "", "", "", "", ""],
        ["cheese", "Сыры", "milk", "20", "", "", "", "", "", ""],
      ]
    );
  }

  function handleDownloadProductsTemplate() {
    downloadCSVTemplate(
      "products_template.csv",
      [
        "categoryCode",
        "categoryName",
        "categoryParentCode",
        "categorySortOrder",
        "productName",
        "description",
        "sku",
        "priceAmount",
        "currencyCode",
        "isActive",
      ],
      [
        ["milk", "Молочная продукция", "", "10", "Молоко 1л", "Пастеризованное 3.2%", "MILK-1L-32", "120.00", "RUB", "true"],
        ["cheese", "Сыры", "milk", "20", "Сыр Гауда 200г", "Полутвердый", "CHEESE-GOUDA-200", "290.00", "RUB", "true"],
      ]
    );
  }

  function handleDownloadPriceListTemplate() {
    downloadCSVTemplate(
      "price_list_template.csv",
      ["sku", "productName", "priceAmount", "currencyCode", "unit", "minBatch", "note"],
      [
        ["MILK-1L-32", "Молоко 1л", "120.00", "RUB", "шт", "10", "Цена с НДС"],
        ["CHEESE-GOUDA-200", "Сыр Гауда 200г", "290.00", "RUB", "шт", "20", "Опт от 20 шт"],
      ]
    );
  }

  async function handleCreateCategory(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!accessToken || !selectedOrganizationId) {
      return;
    }
    setCatalogState({
      kind: "working",
      title: "Каталог",
      description: "Создаём категорию...",
    });
    try {
      await apiFetch(`/v1/organizations/${selectedOrganizationId}/product-categories`, {
        method: "POST",
        accessToken,
        bodyJSON: {
          code: newCategoryCode.trim(),
          name: newCategoryName.trim(),
          status: newCategoryStatus,
          sortOrder: 0,
        },
      });
      setNewCategoryCode("");
      setNewCategoryName("");
      setNewCategoryStatus("draft");
      await refreshCatalog();
      setCatalogState({
        kind: "success",
        title: "Каталог",
        description: "Категория создана.",
      });
    } catch (error) {
      setCatalogState({
        kind: "error",
        title: "Каталог",
        description: problemMessage(error, "Не удалось создать категорию"),
      });
    }
  }

  async function handleUpdateCategory(categoryId: string) {
    if (!accessToken || !selectedOrganizationId) {
      return;
    }
    const current = editingCategory[categoryId];
    if (!current) {
      return;
    }
    setCatalogState({
      kind: "working",
      title: "Каталог",
      description: "Сохраняем категорию...",
    });
    try {
      await apiFetch(`/v1/organizations/${selectedOrganizationId}/product-categories/${categoryId}`, {
        method: "PATCH",
        accessToken,
        bodyJSON: {
          code: current.code,
          name: current.name,
          status: current.status,
          sortOrder: current.sortOrder,
        },
      });
      await refreshCatalog();
      setCatalogState({
        kind: "success",
        title: "Каталог",
        description: "Категория обновлена.",
      });
    } catch (error) {
      setCatalogState({
        kind: "error",
        title: "Каталог",
        description: problemMessage(error, "Не удалось обновить категорию"),
      });
    }
  }

  async function handleCreateProduct(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!accessToken || !selectedOrganizationId) {
      return;
    }
    setCatalogState({
      kind: "working",
      title: "Каталог",
      description: "Создаём продукцию...",
    });
    try {
      await apiFetch(`/v1/organizations/${selectedOrganizationId}/products`, {
        method: "POST",
        accessToken,
        bodyJSON: {
          name: newProductName.trim(),
          categoryId: newProductCategoryId || undefined,
          status: newProductStatus,
          priceAmount: newProductPriceAmount.trim() || undefined,
          currencyCode: newProductCurrency.trim() || undefined,
          isActive: true,
        },
      });
      setNewProductName("");
      setNewProductCategoryId("");
      setNewProductPriceAmount("");
      setNewProductCurrency("RUB");
      setNewProductStatus("draft");
      await refreshCatalog();
      setCatalogState({
        kind: "success",
        title: "Каталог",
        description: "Продукция создана.",
      });
    } catch (error) {
      setCatalogState({
        kind: "error",
        title: "Каталог",
        description: problemMessage(error, "Не удалось создать продукцию"),
      });
    }
  }

  async function handleUpdateProduct(productId: string) {
    if (!accessToken || !selectedOrganizationId) {
      return;
    }
    const current = editingProduct[productId];
    if (!current) {
      return;
    }
    setCatalogState({
      kind: "working",
      title: "Каталог",
      description: "Сохраняем продукцию...",
    });
    try {
      await apiFetch(`/v1/organizations/${selectedOrganizationId}/products/${productId}`, {
        method: "PATCH",
        accessToken,
        bodyJSON: {
          name: current.name,
          categoryId: current.categoryId || "",
          status: current.status,
          priceAmount: current.priceAmount || "",
          currencyCode: current.currencyCode || "",
          isActive: current.isActive,
        },
      });
      await refreshCatalog();
      setCatalogState({
        kind: "success",
        title: "Каталог",
        description: "Продукция обновлена.",
      });
    } catch (error) {
      setCatalogState({
        kind: "error",
        title: "Каталог",
        description: problemMessage(error, "Не удалось обновить продукцию"),
      });
    }
  }

  async function refreshOrganizationKYC(organizationId = selectedOrganizationId) {
    if (!accessToken || !organizationId) {
      setKYCProfile(null);
      setKYCState(initialKYCState);
      return;
    }
    try {
      const payload = await apiFetch<OrganizationKYCProfile>(`/v1/organizations/${organizationId}/kyc`, { accessToken });
      setKYCProfile(payload);
      setKYCDraft({
        status: payload.status || "draft",
        legalName: payload.legalName || "",
        countryCode: payload.countryCode || "",
        registrationNumber: payload.registrationNumber || "",
        taxId: payload.taxId || "",
      });
      setKYCState({
        kind: "success",
        title: "KYC",
        description: "KYC профиль организации загружен.",
      });
    } catch (error) {
      setKYCProfile(null);
      setKYCState({
        kind: "error",
        title: "KYC",
        description: problemMessage(error, "Не удалось загрузить KYC профиль"),
      });
    }
  }

  function handleKYCFileSelection(event: ChangeEvent<HTMLInputElement>) {
    setKYCFiles(Array.from(event.target.files || []));
  }

  async function handleSaveKYCProfile(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!accessToken || !selectedOrganizationId) {
      setKYCState({
        kind: "error",
        title: "KYC",
        description: "Сначала выберите organization и убедитесь, что есть локальная сессия.",
      });
      return;
    }
    setKYCState({
      kind: "working",
      title: "KYC",
      description: "Сохраняем KYC профиль организации...",
    });
    try {
      await apiFetch(`/v1/organizations/${selectedOrganizationId}/kyc`, {
        method: "PATCH",
        accessToken,
        bodyJSON: {
          status: kycDraft.status,
          legalName: kycDraft.legalName,
          countryCode: kycDraft.countryCode,
          registrationNumber: kycDraft.registrationNumber,
          taxId: kycDraft.taxId,
        },
      });
      await refreshOrganizationKYC(selectedOrganizationId);
    } catch (error) {
      setKYCState({
        kind: "error",
        title: "KYC",
        description: problemMessage(error, "Не удалось сохранить KYC профиль"),
      });
    }
  }

  async function handleSubmitKYCProfile() {
    if (!accessToken || !selectedOrganizationId) {
      setKYCState({
        kind: "error",
        title: "KYC",
        description: "Сначала выберите organization и убедитесь, что есть локальная сессия.",
      });
      return;
    }
    setKYCState({
      kind: "working",
      title: "KYC",
      description: "Создаём/обновляем заявку на KYC review...",
    });
    try {
      await apiFetch(`/v1/organizations/${selectedOrganizationId}/kyc`, {
        method: "PATCH",
        accessToken,
        bodyJSON: {
          status: "submitted",
          legalName: kycDraft.legalName,
          countryCode: kycDraft.countryCode,
          registrationNumber: kycDraft.registrationNumber,
          taxId: kycDraft.taxId,
        },
      });
      await refreshOrganizationKYC(selectedOrganizationId);
      setKYCState({
        kind: "success",
        title: "KYC",
        description: "Заявка отправлена на проверку. Можно догружать документы и отправлять повторно при необходимости.",
      });
    } catch (error) {
      setKYCState({
        kind: "error",
        title: "KYC",
        description: problemMessage(error, "Не удалось отправить KYC профиль на review"),
      });
    }
  }

  async function handleUploadKYCDocument(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!accessToken || !selectedOrganizationId) {
      setKYCState({
        kind: "error",
        title: "KYC",
        description: "Сначала выберите organization и убедитесь, что есть локальная сессия.",
      });
      return;
    }
    if (kycFiles.length === 0) {
      setKYCState({
        kind: "error",
        title: "KYC",
        description: "Выберите хотя бы один файл KYC документа.",
      });
      return;
    }
    setKYCState({
      kind: "working",
      title: "KYC",
      description: `Загружаем KYC документы: ${kycFiles.length} шт...`,
    });
    try {
      const failures: string[] = [];
      const baseTitle = kycDocumentTitle.trim();

      for (let index = 0; index < kycFiles.length; index += 1) {
        const file = kycFiles[index];
        const title =
          baseTitle.length > 0
            ? kycFiles.length > 1
              ? `${baseTitle} (${index + 1})`
              : baseTitle
            : file.name;

        try {
          const upload = await apiFetch<{ id: string; uploadUrl: string }>(`/v1/organizations/${selectedOrganizationId}/kyc/documents/uploads`, {
            method: "POST",
            accessToken,
            bodyJSON: {
              documentType: kycDocumentType,
              title,
              fileName: file.name,
              contentType: file.type || "application/octet-stream",
              sizeBytes: file.size,
            },
          });

          const putResponse = await fetch(upload.uploadUrl, {
            method: "PUT",
            headers: file.type ? { "Content-Type": file.type } : undefined,
            body: file,
          });
          if (!putResponse.ok) {
            throw new Error(`HTTP ${putResponse.status}`);
          }

          await apiFetch(`/v1/organizations/${selectedOrganizationId}/kyc/documents/uploads/${encodeURIComponent(upload.id)}/complete`, {
            method: "POST",
            accessToken,
          });
        } catch (error) {
          failures.push(`${file.name}: ${problemMessage(error, "upload failed")}`);
        }
      }
      await refreshOrganizationKYC(selectedOrganizationId);
      if (failures.length > 0) {
        setKYCState({
          kind: "error",
          title: "KYC",
          description: `Часть файлов не загрузилась (${failures.length}/${kycFiles.length}). Первый сбой: ${failures[0]}`,
        });
        return;
      }
      setKYCFiles([]);
      setKYCDocumentTitle("");
    } catch (error) {
      setKYCState({
        kind: "error",
        title: "KYC",
        description: problemMessage(error, "Не удалось загрузить KYC документ"),
      });
    }
  }

  async function handleCreateRole(e: FormEvent) {
    e.preventDefault();
    if (!accessToken || !selectedOrganizationId || !roleForm) return;
    setRolesState({ kind: "working", title: "Роли", description: "Создаём роль..." });
    try {
      await apiFetch(`/v1/organizations/${selectedOrganizationId}/roles`, {
        method: "POST",
        accessToken,
        bodyJSON: { code: roleForm.code.trim(), name: roleForm.name.trim(), description: roleForm.description.trim() || undefined, baseRole: roleForm.baseRole },
      });
      setRoleForm(null);
      setListRefreshKey((k) => k + 1);
      setRolesState({ kind: "success", title: "Роль создана", description: "Роль успешно добавлена." });
    } catch (error) {
      setRolesState({ kind: "error", title: "Ошибка", description: problemMessage(error, "Не удалось создать роль") });
    }
  }

  async function handleUpdateRole(e: FormEvent) {
    e.preventDefault();
    if (!accessToken || !selectedOrganizationId || !editingRoleId || !roleForm) return;
    setRolesState({ kind: "working", title: "Роли", description: "Сохраняем изменения..." });
    try {
      await apiFetch(`/v1/organizations/${selectedOrganizationId}/roles/${editingRoleId}`, {
        method: "PATCH",
        accessToken,
        bodyJSON: { name: roleForm.name.trim(), description: roleForm.description.trim() || undefined, baseRole: roleForm.baseRole },
      });
      setEditingRoleId(null);
      setRoleForm(null);
      setListRefreshKey((k) => k + 1);
      setRolesState({ kind: "success", title: "Роль обновлена", description: "Изменения сохранены." });
    } catch (error) {
      setRolesState({ kind: "error", title: "Ошибка", description: problemMessage(error, "Не удалось обновить роль") });
    }
  }

  async function handleDeleteRole(roleId: string) {
    if (!accessToken || !selectedOrganizationId) return;
    if (!confirm("Удалить роль? Участники с этой ролью должны быть переназначены до удаления.")) return;
    setRolesState({ kind: "working", title: "Роли", description: "Удаляем роль..." });
    try {
      await apiFetch(`/v1/organizations/${selectedOrganizationId}/roles/${roleId}`, { method: "DELETE", accessToken });
      setEditingRoleId(null);
      setRoleForm(null);
      setListRefreshKey((k) => k + 1);
      setRolesState({ kind: "success", title: "Роль удалена", description: "Роль помечена как удалённая (soft-delete)." });
    } catch (error) {
      setRolesState({ kind: "error", title: "Ошибка", description: problemMessage(error, "Не удалось удалить роль") });
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
        <Panel
          title="Organization profile"
          eyebrow={selectedOrganization ? selectedOrganization.name : "Select organization"}
          actions={
            <div className="button-row">
              <button
                className={`button ${organizationSection === "profile" ? "primary" : "secondary"}`}
                type="button"
                onClick={() => setOrganizationSection("profile")}
              >
                Профиль
              </button>
              <button
                className={`button ${organizationSection === "uploads" ? "primary" : "secondary"}`}
                type="button"
                onClick={() => setOrganizationSection("uploads")}
              >
                Категории / Прайс / Продукция
              </button>
              <button
                className={`button ${organizationSection === "catalog" ? "primary" : "secondary"}`}
                type="button"
                onClick={() => setOrganizationSection("catalog")}
              >
                Списки каталога
              </button>
              <button
                className={`button ${organizationSection === "kyc" ? "primary" : "secondary"}`}
                type="button"
                onClick={() => setOrganizationSection("kyc")}
              >
                KYC
              </button>
              {canManageOrganization ? (
                <button
                  className={`button ${organizationSection === "roles" ? "primary" : "secondary"}`}
                  type="button"
                  onClick={() => setOrganizationSection("roles")}
                >
                  Роли
                </button>
              ) : null}
            </div>
          }
        >
          <div className={`status-card ${profileState.kind === "error" ? "error" : profileState.kind === "success" ? "success" : "info"}`}>
            <strong>{profileState.title}</strong>
            <p className="status-copy">{profileState.description}</p>
          </div>

          {profile ? (
            <>
              {organizationSection === "profile" ? (
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
                  <button className="button primary" type="submit" disabled={!canManageOrganization}>
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
                  <button className="button secondary" type="submit" disabled={!canManageOrganization}>
                    Upload logo
                  </button>
                </div>
              </form>
                </>
              ) : null}

              {organizationSection === "uploads" ? (
                <div className="mini-card">
                  <h3>Загрузка категорий, прайс-листов и продукции</h3>
                  <p className="muted">
                    Это внутренний подраздел выбранной организации. Он использует существующие backend upload/import endpoints без отдельного контракта.
                  </p>
                  <div className="cards">
                    <form className="form-grid" onSubmit={handleUploadCategories}>
                      <div
                        className={`status-card ${categoriesUploadState.kind === "error" ? "error" : categoriesUploadState.kind === "success" ? "success" : "info"}`}
                      >
                        <strong>{categoriesUploadState.title}</strong>
                        <p className="status-copy">{categoriesUploadState.description}</p>
                      </div>
                      <div className="form-row">
                        <label className="form-label" htmlFor="categories-file">
                          CSV категорий
                        </label>
                        <input id="categories-file" type="file" accept=".csv,text/csv" onChange={handleCategoriesSelection} />
                      </div>
                      <div className="button-row">
                        <button className="button secondary" type="submit" disabled={!canManageCatalog}>
                          Загрузить категории
                        </button>
                        <button className="button secondary" type="button" onClick={handleDownloadCategoriesTemplate}>
                          Скачать шаблон CSV
                        </button>
                      </div>
                      <p className="muted">
                        Колонки: <code>categoryCode</code> (обязательно для категории), <code>categoryName</code> (обязательно для новой категории),{" "}
                        <code>categoryParentCode</code>, <code>categorySortOrder</code>. Поля продукции в этом же CSV можно оставить пустыми.
                      </p>
                      {categoriesUploadResult ? <textarea className="code-block" readOnly value={JSON.stringify(categoriesUploadResult, null, 2)} /> : null}
                    </form>

                    <form className="form-grid" onSubmit={handleUploadProducts}>
                      <div
                        className={`status-card ${productsUploadState.kind === "error" ? "error" : productsUploadState.kind === "success" ? "success" : "info"}`}
                      >
                        <strong>{productsUploadState.title}</strong>
                        <p className="status-copy">{productsUploadState.description}</p>
                      </div>
                      <div className="form-row">
                        <label className="form-label" htmlFor="products-file">
                          CSV продукции
                        </label>
                        <input id="products-file" type="file" accept=".csv,text/csv" onChange={handleProductsSelection} />
                      </div>
                      <div className="button-row">
                        <button className="button secondary" type="submit" disabled={!canManageCatalog}>
                          Загрузить продукцию
                        </button>
                        <button className="button secondary" type="button" onClick={handleDownloadProductsTemplate}>
                          Скачать шаблон CSV
                        </button>
                      </div>
                      <p className="muted">
                        Колонки: <code>productName</code> (обязательно), <code>categoryCode</code> (рекомендуется), <code>sku</code>, <code>description</code>,{" "}
                        <code>priceAmount</code>, <code>currencyCode</code>, <code>isActive</code>.
                      </p>
                      {productsUploadResult ? <textarea className="code-block" readOnly value={JSON.stringify(productsUploadResult, null, 2)} /> : null}
                    </form>
                  </div>

                  <form className="form-grid" onSubmit={handleUploadPriceList}>
                    <div
                      className={`status-card ${priceListUploadState.kind === "error" ? "error" : priceListUploadState.kind === "success" ? "success" : "info"}`}
                    >
                      <strong>{priceListUploadState.title}</strong>
                      <p className="status-copy">{priceListUploadState.description}</p>
                    </div>
                    <div className="form-row">
                      <label className="form-label" htmlFor="price-list-file">
                        Файл прайс-листа
                      </label>
                      <input
                        id="price-list-file"
                        type="file"
                        accept=".csv,.xls,.xlsx,.pdf,application/vnd.ms-excel,application/vnd.openxmlformats-officedocument.spreadsheetml.sheet,application/pdf"
                        onChange={handlePriceListSelection}
                      />
                    </div>
                    <div className="button-row">
                      <button className="button secondary" type="submit" disabled={!canManageOrganization}>
                        Загрузить прайс-лист
                      </button>
                      <button className="button secondary" type="button" onClick={handleDownloadPriceListTemplate}>
                        Скачать шаблон прайс-листа
                      </button>
                    </div>
                    <p className="muted">
                      Шаблон прайс-листа: <code>sku</code>, <code>productName</code>, <code>priceAmount</code>, <code>currencyCode</code>, <code>unit</code>,{" "}
                      <code>minBatch</code>, <code>note</code>. Это вспомогательный шаблон для удобного обмена с организацией.
                    </p>
                    {priceListUploadResult ? <textarea className="code-block" readOnly value={JSON.stringify(priceListUploadResult, null, 2)} /> : null}
                  </form>

                  <div className={`status-card ${catalogState.kind === "error" ? "error" : catalogState.kind === "success" ? "success" : "info"}`}>
                    <strong>{catalogState.title}</strong>
                    <p className="status-copy">{catalogState.description}</p>
                  </div>

                  <div className="cards">
                    <form className="form-grid" onSubmit={handleCreateCategory}>
                      <h3>Категории</h3>
                      <div className="form-row two">
                        <div className="form-row">
                          <label className="form-label">Код</label>
                          <input className="text-input" value={newCategoryCode} onChange={(event) => setNewCategoryCode(event.target.value)} required />
                        </div>
                        <div className="form-row">
                          <label className="form-label">Название</label>
                          <input className="text-input" value={newCategoryName} onChange={(event) => setNewCategoryName(event.target.value)} required />
                        </div>
                      </div>
                      <div className="form-row">
                        <label className="form-label">Статус категории</label>
                        <select className="text-input" value={newCategoryStatus} onChange={(event) => setNewCategoryStatus(event.target.value)}>
                          {CATALOG_STATUSES.map((status) => (
                            <option key={status} value={status}>
                              {catalogStatusLabel(status)}
                            </option>
                          ))}
                        </select>
                      </div>
                      <div className="button-row">
                        <button className="button secondary" type="submit" disabled={!canManageCatalog}>
                          Добавить категорию
                        </button>
                      </div>
                      <div className="domain-list">
                        {categories.map((item) => {
                          const draft = editingCategory[item.id] || { code: item.code, name: item.name, sortOrder: item.sortOrder || 0, status: item.status || "draft" };
                          return (
                            <div key={item.id} className="inline-panel">
                              <div className="form-row two">
                                <input
                                  className="text-input"
                                  value={draft.code}
                                  onChange={(event) =>
                                    setEditingCategory((prev) => ({ ...prev, [item.id]: { ...draft, code: event.target.value } }))
                                  }
                                />
                                <input
                                  className="text-input"
                                  value={draft.name}
                                  onChange={(event) =>
                                    setEditingCategory((prev) => ({ ...prev, [item.id]: { ...draft, name: event.target.value } }))
                                  }
                                />
                              </div>
                              <div className="form-row">
                                <select
                                  className="text-input"
                                  value={draft.status}
                                  onChange={(event) => setEditingCategory((prev) => ({ ...prev, [item.id]: { ...draft, status: event.target.value } }))}
                                >
                                  {CATALOG_STATUSES.map((status) => (
                                    <option key={status} value={status}>
                                      {catalogStatusLabel(status)}
                                    </option>
                                  ))}
                                </select>
                              </div>
                              <div className="button-row">
                                <button className="button secondary" type="button" onClick={() => void handleUpdateCategory(item.id)} disabled={!canManageCatalog}>
                                  Сохранить категорию
                                </button>
                              </div>
                            </div>
                          );
                        })}
                      </div>
                    </form>

                    <form className="form-grid" onSubmit={handleCreateProduct}>
                      <h3>Продукция</h3>
                      <div className="form-row">
                        <label className="form-label">Название</label>
                        <input className="text-input" value={newProductName} onChange={(event) => setNewProductName(event.target.value)} required />
                      </div>
                      <div className="form-row">
                        <label className="form-label">Статус продукции</label>
                        <select className="text-input" value={newProductStatus} onChange={(event) => setNewProductStatus(event.target.value)}>
                          {CATALOG_STATUSES.map((status) => (
                            <option key={status} value={status}>
                              {catalogStatusLabel(status)}
                            </option>
                          ))}
                        </select>
                      </div>
                      <div className="form-row two">
                        <div className="form-row">
                          <label className="form-label">Категория</label>
                          <select className="text-input" value={newProductCategoryId} onChange={(event) => setNewProductCategoryId(event.target.value)}>
                            <option value="">Без категории</option>
                            {categories.map((item) => (
                              <option key={item.id} value={item.id}>
                                {item.name}
                              </option>
                            ))}
                          </select>
                        </div>
                        <div className="form-row">
                          <label className="form-label">Цена</label>
                          <input className="text-input" value={newProductPriceAmount} onChange={(event) => setNewProductPriceAmount(event.target.value)} />
                        </div>
                      </div>
                      <div className="button-row">
                        <button className="button secondary" type="submit" disabled={!canManageCatalog}>
                          Добавить продукцию
                        </button>
                      </div>
                      <div className="domain-list">
                        {products.map((item) => {
                          const draft = editingProduct[item.id] || {
                            name: item.name,
                            categoryId: item.categoryId || "",
                            status: item.status || "draft",
                            priceAmount: item.priceAmount || "",
                            currencyCode: item.currencyCode || "RUB",
                            isActive: item.isActive,
                          };
                          return (
                            <div key={item.id} className="inline-panel">
                              <div className="form-row">
                                <input
                                  className="text-input"
                                  value={draft.name}
                                  onChange={(event) =>
                                    setEditingProduct((prev) => ({ ...prev, [item.id]: { ...draft, name: event.target.value } }))
                                  }
                                />
                              </div>
                              <div className="form-row two">
                                <select
                                  className="text-input"
                                  value={draft.categoryId}
                                  onChange={(event) =>
                                    setEditingProduct((prev) => ({ ...prev, [item.id]: { ...draft, categoryId: event.target.value } }))
                                  }
                                >
                                  <option value="">Без категории</option>
                                  {categories.map((category) => (
                                    <option key={category.id} value={category.id}>
                                      {category.name}
                                    </option>
                                  ))}
                                </select>
                                <input
                                  className="text-input"
                                  value={draft.priceAmount}
                                  onChange={(event) =>
                                    setEditingProduct((prev) => ({ ...prev, [item.id]: { ...draft, priceAmount: event.target.value } }))
                                  }
                                  placeholder="0.00"
                                />
                              </div>
                              <div className="form-row two">
                                <select
                                  className="text-input"
                                  value={draft.status}
                                  onChange={(event) => setEditingProduct((prev) => ({ ...prev, [item.id]: { ...draft, status: event.target.value } }))}
                                >
                                  {CATALOG_STATUSES.map((status) => (
                                    <option key={status} value={status}>
                                      {catalogStatusLabel(status)}
                                    </option>
                                  ))}
                                </select>
                                <input
                                  className="text-input"
                                  value={draft.currencyCode}
                                  onChange={(event) =>
                                    setEditingProduct((prev) => ({ ...prev, [item.id]: { ...draft, currencyCode: event.target.value.toUpperCase() } }))
                                  }
                                  placeholder="RUB"
                                />
                                <label className="muted">
                                  <input
                                    type="checkbox"
                                    checked={draft.isActive}
                                    onChange={(event) =>
                                      setEditingProduct((prev) => ({ ...prev, [item.id]: { ...draft, isActive: event.target.checked } }))
                                    }
                                  />{" "}
                                  active
                                </label>
                              </div>
                              <div className="button-row">
                                <button className="button secondary" type="button" onClick={() => void handleUpdateProduct(item.id)} disabled={!canManageCatalog}>
                                  Сохранить продукцию
                                </button>
                              </div>
                            </div>
                          );
                        })}
                      </div>
                    </form>
                  </div>
                </div>
              ) : null}

              {organizationSection === "catalog" ? (
                <div className="mini-card">
                  <h3>Списки каталога организации</h3>
                  <div className={`status-card ${catalogState.kind === "error" ? "error" : catalogState.kind === "success" ? "success" : "info"}`}>
                    <strong>{catalogState.title}</strong>
                    <p className="status-copy">{catalogState.description}</p>
                  </div>
                  {canManageCatalog && canAutoPublish ? (
                    <div className={`status-card ${publishAllState.kind === "error" ? "error" : publishAllState.kind === "success" ? "success" : "info"}`}>
                      <strong>{publishAllState.title || "Автопубликация"}</strong>
                      <p className="status-copy">
                        {publishAllState.kind === "idle"
                          ? "Всё загружено корректно. Можно опубликовать категории, продукцию и прайс-лист одной кнопкой."
                          : publishAllState.description}
                      </p>
                      <div className="button-row">
                        <button
                          className="button primary"
                          type="button"
                          onClick={() => void handlePublishAllCatalog()}
                          disabled={publishAllState.kind === "working"}
                        >
                          {publishAllState.kind === "working" ? "Публикуем..." : "Опубликовать всё"}
                        </button>
                      </div>
                    </div>
                  ) : null}
                  <div className={`status-card ${priceListsState.kind === "error" ? "error" : priceListsState.kind === "success" ? "success" : "info"}`}>
                    <strong>{priceListsState.title}</strong>
                    <p className="status-copy">{priceListsState.description}</p>
                  </div>

                  <div className="cards">
                    <div className="mini-card">
                      <h3>Категории ({categories.length})</h3>
                      {categories.length > 0 ? (
                        <div className="domain-list">
                          {categories.map((item) => (
                            <div key={item.id} className="inline-panel">
                              <strong>{item.name}</strong>
                              <p className="muted">
                                <code>{item.code}</code> · {catalogStatusLabel(item.status)}
                              </p>
                            </div>
                          ))}
                        </div>
                      ) : (
                        <p className="muted">Категорий пока нет.</p>
                      )}
                    </div>

                    <div className="mini-card">
                      <h3>Продукция ({products.length})</h3>
                      {products.length > 0 ? (
                        <div className="domain-list">
                          {products.map((item) => (
                            <div key={item.id} className="inline-panel">
                              <strong>{item.name}</strong>
                              <p className="muted">
                                {item.priceAmount ? `${item.priceAmount} ${item.currencyCode || ""}` : "Цена не указана"} · {catalogStatusLabel(item.status)} ·{" "}
                                {item.isActive ? "active" : "inactive"}
                              </p>
                            </div>
                          ))}
                        </div>
                      ) : (
                        <p className="muted">Продукции пока нет.</p>
                      )}
                    </div>
                  </div>

                  <div className="mini-card">
                    <h3>Прайс-листы ({priceListFiles.length})</h3>
                    <p className="muted">
                      Текущий прайс-лист:{" "}
                      {cooperationApplication?.priceListObjectId ? <code>{cooperationApplication.priceListObjectId}</code> : "не назначен"}
                    </p>
                    <div className="form-row">
                      <label className="form-label">Статус прайс-листа</label>
                      <select
                        className="text-input"
                        value={cooperationApplication?.priceListStatus || "draft"}
                        onChange={(event) => void handleUpdatePriceListStatus(event.target.value)}
                        disabled={!cooperationApplication || !canManageOrganization}
                      >
                        {CATALOG_STATUSES.map((status) => (
                          <option key={status} value={status}>
                            {catalogStatusLabel(status)}
                          </option>
                        ))}
                      </select>
                    </div>
                    <div className="button-row">
                      <button
                        className="button secondary"
                        type="button"
                        onClick={() => void handleDownloadCurrentPriceList()}
                        disabled={!cooperationApplication?.priceListObjectId}
                      >
                        Скачать текущий прайс-лист
                      </button>
                    </div>
                    {priceListFiles.length > 0 ? (
                      <div className="domain-list">
                        {priceListFiles.map((item) => (
                          <div key={item.objectId} className="inline-panel">
                            <strong>{item.fileName}</strong>
                            <p className="muted">
                              <code>{item.objectId}</code>
                            </p>
                            <p className="muted">
                              {Math.max(1, Math.round(item.sizeBytes / 1024))} KB · {formatTimestamp(item.createdAt)}
                            </p>
                          </div>
                        ))}
                      </div>
                    ) : (
                      <p className="muted">Прайс-листы пока не загружались.</p>
                    )}
                  </div>
                </div>
              ) : null}

              {organizationSection === "kyc" ? (
                <div className="mini-card">
                  <h3>KYC профиль организации</h3>
                  <div className={`status-card ${kycState.kind === "error" ? "error" : kycState.kind === "success" ? "success" : "info"}`}>
                    <strong>{kycState.title}</strong>
                    <p className="status-copy">{kycState.description}</p>
                    {kycProfile?.status ? (
                      <p className="status-copy">
                        Текущий статус: <strong>{kycStatusLabel(kycProfile.status)}</strong>
                      </p>
                    ) : null}
                  </div>
                  {kycProfile ? (
                    <div className="cards">
                      {(["draft", "submitted", "in_review", "verified", "rejected"] as KYCQueueBucket[]).map((bucket) => {
                        const items = (kycProfile.documents || []).filter((item) => getKYCQueueBucket(item.status, kycProfile.status) === bucket);
                        return (
                          <div key={bucket} className="mini-card">
                            <h3>{kycQueueBucketLabel(bucket)}</h3>
                            <p className="muted">Документов: {items.length}</p>
                          </div>
                        );
                      })}
                    </div>
                  ) : null}
                  <form className="form-grid" onSubmit={handleSaveKYCProfile}>
                    <div className="form-row two">
                      <div className="form-row">
                        <label className="form-label">Review status</label>
                        <input className="text-input" value={kycStatusLabel(kycProfile?.status || kycDraft.status)} readOnly />
                      </div>
                      <div className="form-row">
                        <label className="form-label">Country code</label>
                        <input className="text-input" value={kycDraft.countryCode} onChange={(event) => setKYCDraft((value) => ({ ...value, countryCode: event.target.value }))} />
                      </div>
                    </div>
                    <div className="form-row two">
                      <div className="form-row">
                        <label className="form-label">Legal name</label>
                        <input className="text-input" value={kycDraft.legalName} onChange={(event) => setKYCDraft((value) => ({ ...value, legalName: event.target.value }))} />
                      </div>
                      <div className="form-row">
                        <label className="form-label">Registration number</label>
                        <input
                          className="text-input"
                          value={kycDraft.registrationNumber}
                          onChange={(event) => setKYCDraft((value) => ({ ...value, registrationNumber: event.target.value }))}
                        />
                      </div>
                    </div>
                    <div className="form-row">
                      <label className="form-label">Tax ID</label>
                      <input className="text-input" value={kycDraft.taxId} onChange={(event) => setKYCDraft((value) => ({ ...value, taxId: event.target.value }))} />
                    </div>
                    <div className="button-row">
                      <button className="button primary" type="submit">
                        Сохранить KYC профиль
                      </button>
                      <button
                        className="button secondary"
                        type="button"
                        onClick={() => void handleSubmitKYCProfile()}
                        disabled={isReviewSubmissionLocked(kycProfile?.status)}
                      >
                        {kycSubmitActionLabel(kycProfile?.status)}
                      </button>
                    </div>
                    <p className="muted">
                      KYC поддерживает частичную верификацию: можно досылать дополнительные документы в текущую заявку и проходить проверку поэтапно.
                    </p>
                  </form>

                  <form className="form-grid" onSubmit={handleUploadKYCDocument}>
                    <div className="form-row two">
                      <div className="form-row">
                        <label className="form-label">Document type</label>
                        <input className="text-input" value={kycDocumentType} onChange={(event) => setKYCDocumentType(event.target.value)} />
                      </div>
                      <div className="form-row">
                        <label className="form-label">Title (опционально)</label>
                        <input className="text-input" value={kycDocumentTitle} onChange={(event) => setKYCDocumentTitle(event.target.value)} />
                      </div>
                    </div>
                    <div className="form-row">
                      <label className="form-label" htmlFor="org-kyc-file">
                        KYC документы
                      </label>
                      <input id="org-kyc-file" type="file" multiple onChange={handleKYCFileSelection} />
                      {kycFiles.length > 0 ? (
                        <p className="muted">
                          Выбрано файлов: {kycFiles.length}. {kycFiles.map((file) => file.name).join(", ")}
                        </p>
                      ) : null}
                    </div>
                    <div className="button-row">
                      <button className="button secondary" type="submit">
                        Upload KYC документы
                      </button>
                    </div>
                  </form>

                  {kycProfile?.documents?.length ? (
                    <div className="domain-list">
                      {(["draft", "submitted", "in_review", "verified", "rejected"] as KYCQueueBucket[])
                        .map((bucket) => ({
                          bucket,
                          items: kycProfile.documents
                            .filter((item) => getKYCQueueBucket(item.status, kycProfile.status) === bucket)
                            .sort((a, b) => (a.createdAt < b.createdAt ? 1 : -1)),
                        }))
                        .filter((group) => group.items.length > 0)
                        .map((group) => (
                          <div key={group.bucket} className="mini-card">
                            <p className="muted">
                              <span className="status-badge">{kycQueueBucketLabel(group.bucket)}</span>{" "}
                              · {group.items.length} шт
                            </p>
                            <div className="domain-list">
                              {group.items.map((item) => (
                                <div key={item.id} className="inline-panel">
                                  <strong>{item.title || item.documentType}</strong>
                                  <p className="muted">
                                    {item.documentType} · <span className={kycDocumentStatusBadgeClass(item.status)}>{kycDocumentStatusLabel(item.status)}</span>
                                  </p>
                                  <p className="muted">Created: {formatTimestamp(item.createdAt)}</p>
                                  {item.reviewNote ? <p className="muted">Note: {item.reviewNote}</p> : null}
                                </div>
                              ))}
                            </div>
                          </div>
                        ))}
                    </div>
                  ) : (
                    <div className="empty-state">Пока нет загруженных KYC документов.</div>
                  )}
                </div>
              ) : null}

              {organizationSection === "roles" ? (
                <div className="mini-card">
                  <h3>Роли организации</h3>
                  <div className={`status-card ${rolesState.kind === "error" ? "error" : rolesState.kind === "success" ? "success" : "info"}`}>
                    <strong>{rolesState.title}</strong>
                    <p className="status-copy">{rolesState.description}</p>
                  </div>
                  <div className="domain-list">
                    {roles.map((r) => (
                      <div key={r.id || r.code} className="inline-panel">
                        <strong>{r.name}</strong>
                        <p className="muted">
                          <code>{r.code}</code>
                          {r.isSystem ? " · системная" : " · базовая: " + r.baseRole}
                          {r.deletedAt ? " · удалена" : ""}
                        </p>
                        {r.description ? <p className="muted">{r.description}</p> : null}
                        {!r.isSystem && !r.deletedAt ? (
                          <div className="button-row">
                            <button
                              className="button secondary"
                              type="button"
                              onClick={() => {
                                setEditingRoleId(r.id);
                                setRoleForm({ code: r.code, name: r.name, description: r.description || "", baseRole: r.baseRole });
                              }}
                            >
                              Редактировать
                            </button>
                            <button className="button secondary" type="button" onClick={() => void handleDeleteRole(r.id)}>
                              Удалить
                            </button>
                          </div>
                        ) : null}
                      </div>
                    ))}
                  </div>
                  {!roleForm && !editingRoleId ? (
                    <button className="button primary" type="button" onClick={() => setRoleForm({ code: "", name: "", description: "", baseRole: "member" })}>
                      Добавить роль
                    </button>
                  ) : null}
                  {roleForm ? (
                    <form
                      className="form-grid"
                      onSubmit={editingRoleId ? handleUpdateRole : handleCreateRole}
                      style={{ marginTop: "1rem" }}
                    >
                      <div className="form-row two">
                        <div className="form-row">
                          <label className="form-label">Код (латиница, lowercase)</label>
                          <input
                            className="text-input"
                            value={roleForm.code}
                            onChange={(e) => setRoleForm((f) => ({ ...f!, code: e.target.value }))}
                            placeholder="project_manager"
                            disabled={!!editingRoleId}
                          />
                        </div>
                        <div className="form-row">
                          <label className="form-label">Название</label>
                          <input
                            className="text-input"
                            value={roleForm.name}
                            onChange={(e) => setRoleForm((f) => ({ ...f!, name: e.target.value }))}
                            placeholder="Менеджер проекта"
                            required
                          />
                        </div>
                      </div>
                      <div className="form-row">
                        <label className="form-label">Описание</label>
                        <input
                          className="text-input"
                          value={roleForm.description}
                          onChange={(e) => setRoleForm((f) => ({ ...f!, description: e.target.value }))}
                          placeholder="Опционально"
                        />
                      </div>
                      <div className="form-row">
                        <label className="form-label">Базовая роль</label>
                        <select
                          className="text-input"
                          value={roleForm.baseRole}
                          onChange={(e) => setRoleForm((f) => ({ ...f!, baseRole: e.target.value }))}
                        >
                          <option value="owner">Owner</option>
                          <option value="admin">Administrator</option>
                          <option value="manager">Manager</option>
                          <option value="member">Member</option>
                          <option value="viewer">Viewer</option>
                        </select>
                      </div>
                      <div className="button-row">
                        <button className="button primary" type="submit">
                          {editingRoleId ? "Сохранить" : "Создать"}
                        </button>
                        <button
                          className="button secondary"
                          type="button"
                          onClick={() => {
                            setRoleForm(null);
                            setEditingRoleId(null);
                          }}
                        >
                          Отмена
                        </button>
                      </div>
                    </form>
                  ) : null}
                </div>
              ) : null}

              {organizationSection === "profile" ? (
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
              ) : null}
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
