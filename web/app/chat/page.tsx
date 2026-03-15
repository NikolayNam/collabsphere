"use client";

import { FormEvent, useEffect, useMemo, useRef, useState } from "react";

import { Panel } from "@/components/panel";
import { APIError, apiFetch } from "@/lib/api";
import { readStoredTokens } from "@/lib/auth";

type MyGroup = {
  id: string;
  name: string;
  slug: string;
  description?: string | null;
  isActive: boolean;
  createdAt: string;
  membershipSource: string;
  membershipRole?: string | null;
};

type MyOrganization = {
  id: string;
  name: string;
  slug: string;
  membershipRole: string;
};

type AccessRequestOrganization = {
  id: string;
  name: string;
  slug: string;
};

type GroupAccountMember = {
  id: string;
  accountId: string;
  email: string;
  displayName?: string | null;
  role: string;
  isActive: boolean;
  createdAt: string;
};

type GroupOrganizationMember = {
  id: string;
  organizationId: string;
  name: string;
  slug: string;
  isActive: boolean;
  createdAt: string;
};

type OrganizationAccessRequest = {
  id: string;
  organizationId: string;
  requesterAccountId: string;
  requestedRole: string;
  message?: string | null;
  status: string;
  reviewerAccountId?: string | null;
  reviewNote?: string | null;
  reviewedAt?: string | null;
  createdAt: string;
  updatedAt?: string | null;
};

type Channel = {
  id: string;
  groupId: string;
  slug: string;
  name: string;
  description?: string | null;
  isDefault: boolean;
  lastMessageSeq: number;
  createdAt: string;
  updatedAt?: string | null;
};

type Message = {
  id: string;
  channelId: string;
  channelSeq: number;
  type: string;
  authorType: string;
  authorAccountId?: string | null;
  authorGuestId?: string | null;
  authorName?: string | null;
  body: string;
  createdAt: string;
  editedAt?: string | null;
  deletedAt?: string | null;
  attachments?: Array<{ objectId: string; fileName: string; sizeBytes: number }>;
};

type CreatedGroup = {
  id: string;
  name: string;
  slug: string;
  description?: string | null;
  isActive: boolean;
  createdAt: string;
};

type Status = {
  kind: "idle" | "working" | "success" | "error";
  title: string;
  description: string;
};

const initialGroupsState: Status = {
  kind: "idle",
  title: "Группы",
  description: "Этот экран использует `GET /v1/groups/my`, затем раскрывает каналы и сообщения без ручного UUID-ввода.",
};

const initialChannelsState: Status = {
  kind: "idle",
  title: "Каналы",
  description: "После выбора группы backend вернёт каналы через `GET /v1/groups/{groupId}/channels`.",
};

const initialJoinRequestState: Status = {
  kind: "idle",
  title: "Запрос на присоединение",
  description: "Можно отправить запрос в backend на добавление аккаунта в группу выбранного канала.",
};

const initialOrganizationLinkState: Status = {
  kind: "idle",
  title: "Группы организаций",
  description: "Можно добавить организацию в выбранную группу через `POST /v1/groups/{groupId}/organizations`.",
};

const initialMembersState: Status = {
  kind: "idle",
  title: "Участники группы",
  description: "Состав группы читается через `GET /v1/groups/{groupId}/members`.",
};

const initialOrganizationAccessRequestState: Status = {
  kind: "idle",
  title: "Доступ к организации",
  description: "Отправка self-service запроса в `POST /v1/organizations/{organizationId}/access-requests`.",
};

const initialOrganizationAccessQueueState: Status = {
  kind: "idle",
  title: "Очередь заявок",
  description: "Для вашей организации можно смотреть и разбирать заявки через `/access-requests`.",
};

const initialMessagesState: Status = {
  kind: "idle",
  title: "Лента сообщений",
  description: "После выбора канала timeline читается из `GET /v1/channels/{channelId}/messages` и обновляется polling-ом.",
};

const initialComposerState: Status = {
  kind: "idle",
  title: "Отправка сообщения",
  description: "Composer пишет напрямую в `POST /v1/channels/{channelId}/messages`.",
};

function slugify(value: string): string {
  return value
    .toLowerCase()
    .trim()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "")
    .slice(0, 60);
}

function formatDate(value: string): string {
  const parsed = new Date(value);
  if (Number.isNaN(parsed.getTime())) {
    return value;
  }
  return new Intl.DateTimeFormat("ru-RU", {
    dateStyle: "short",
    timeStyle: "short",
  }).format(parsed);
}

function authorLabel(message: Message): string {
  if (message.authorName && message.authorName.trim() !== "") {
    return message.authorName;
  }
  if (message.authorAccountId) {
    return `Account ${message.authorAccountId.slice(0, 8)}`;
  }
  if (message.authorGuestId) {
    return `Guest ${message.authorGuestId.slice(0, 8)}`;
  }
  return "Unknown author";
}

export default function ChatPage() {
  const [groups, setGroups] = useState<MyGroup[]>([]);
  const [organizations, setOrganizations] = useState<MyOrganization[]>([]);
  const [requestOrganizations, setRequestOrganizations] = useState<AccessRequestOrganization[]>([]);
  const [groupAccountMembers, setGroupAccountMembers] = useState<GroupAccountMember[]>([]);
  const [groupOrganizationMembers, setGroupOrganizationMembers] = useState<GroupOrganizationMember[]>([]);
  const [channels, setChannels] = useState<Channel[]>([]);
  const [messages, setMessages] = useState<Message[]>([]);
  const [selectedGroupId, setSelectedGroupId] = useState("");
  const [selectedOrganizationId, setSelectedOrganizationId] = useState("");
  const [requestOrganizationId, setRequestOrganizationId] = useState("");
  const [requestRole, setRequestRole] = useState("member");
  const [requestMessage, setRequestMessage] = useState("");
  const [selectedChannelId, setSelectedChannelId] = useState("");
  const [groupName, setGroupName] = useState("Product Team");
  const [groupSlug, setGroupSlug] = useState("product-team");
  const [draft, setDraft] = useState("");
  const [pendingAttachments, setPendingAttachments] = useState<Array<{ objectId: string; fileName: string; sizeBytes: number }>>([]);
  const [attachmentUploadState, setAttachmentUploadState] = useState<Status>({ kind: "idle", title: "", description: "" });
  const [groupsState, setGroupsState] = useState<Status>(initialGroupsState);
  const [channelsState, setChannelsState] = useState<Status>(initialChannelsState);
  const [messagesState, setMessagesState] = useState<Status>(initialMessagesState);
  const [composerState, setComposerState] = useState<Status>(initialComposerState);
  const [joinRequestState, setJoinRequestState] = useState<Status>(initialJoinRequestState);
  const [organizationLinkState, setOrganizationLinkState] = useState<Status>(initialOrganizationLinkState);
  const [membersState, setMembersState] = useState<Status>(initialMembersState);
  const [organizationAccessRequestState, setOrganizationAccessRequestState] = useState<Status>(initialOrganizationAccessRequestState);
  const [organizationAccessQueueState, setOrganizationAccessQueueState] = useState<Status>(initialOrganizationAccessQueueState);
  const [organizationAccessQueue, setOrganizationAccessQueue] = useState<OrganizationAccessRequest[]>([]);
  const [groupsRefreshKey, setGroupsRefreshKey] = useState(0);
  const [messagesRefreshKey, setMessagesRefreshKey] = useState(0);
  const fileInputRef = useRef<HTMLInputElement | null>(null);

  const accessToken = useMemo(() => readStoredTokens()?.accessToken || null, []);
  const selectedGroup = groups.find((group) => group.id === selectedGroupId) || null;
  const selectedOrganization = organizations.find((organization) => organization.id === selectedOrganizationId) || null;
  const selectedChannel = channels.find((channel) => channel.id === selectedChannelId) || null;

  useEffect(() => {
    let cancelled = false;

    async function loadGroups() {
      if (!accessToken) {
        if (!cancelled) {
          setGroups([]);
          setSelectedGroupId("");
          setGroupsState({
            kind: "error",
            title: "Нет локальной сессии",
            description: "Сначала завершите login через /login. Чат использует текущий bearer token.",
          });
        }
        return;
      }

      if (!cancelled) {
        setGroupsState({
          kind: "working",
          title: "Загружаем группы",
          description: "Backend объединяет прямые membership и organization-derived access paths для текущего account.",
        });
      }

      try {
        const payload = await apiFetch<{ data?: MyGroup[]; body?: { data?: MyGroup[] } }>("/v1/groups/my", { accessToken });
        if (cancelled) {
          return;
        }
        const items = Array.isArray(payload.data) ? payload.data : Array.isArray(payload.body?.data) ? payload.body.data : [];
        setGroups(items);
        setSelectedGroupId((current) => {
          if (current && items.some((item) => item.id === current)) {
            return current;
          }
          return items[0]?.id || "";
        });
        setGroupsState({
          kind: "success",
          title: items.length > 0 ? "Группы загружены" : "Групп пока нет",
          description:
            items.length > 0
              ? "Выберите группу слева. Для новых групп default channel должен появляться автоматически."
              : "Создайте первую группу. Backend автоматически провиженит owner membership и default channel.",
        });
      } catch (error) {
        if (cancelled) {
          return;
        }
        const message =
          error instanceof APIError
            ? `${error.message}${error.code ? ` (${error.code})` : ""}`
            : error instanceof Error
              ? error.message
              : "Unknown groups error";
        setGroups([]);
        setSelectedGroupId("");
        setGroupsState({
          kind: "error",
          title: "Не удалось загрузить группы",
          description: message,
        });
      }
    }

    void loadGroups();
    return () => {
      cancelled = true;
    };
  }, [accessToken, groupsRefreshKey]);

  useEffect(() => {
    let cancelled = false;

    async function loadOrganizations() {
      if (!accessToken) {
        if (!cancelled) {
          setOrganizations([]);
          setSelectedOrganizationId("");
          setOrganizationLinkState({
            kind: "error",
            title: "Нет локальной сессии",
            description: "Сначала завершите login через /login, затем можно добавлять организации в группы.",
          });
        }
        return;
      }

      try {
        const payload = await apiFetch<{ data?: MyOrganization[] }>("/v1/organizations/my", { accessToken });
        if (cancelled) {
          return;
        }
        const items = Array.isArray(payload.data) ? payload.data : [];
        setOrganizations(items);
        setSelectedOrganizationId((current) => {
          if (current && items.some((item) => item.id === current)) {
            return current;
          }
          return items[0]?.id || "";
        });
        setOrganizationLinkState((current) => (current.kind === "error" ? initialOrganizationLinkState : current));
      } catch (error) {
        if (cancelled) {
          return;
        }
        const message =
          error instanceof APIError
            ? `${error.message}${error.code ? ` (${error.code})` : ""}`
            : error instanceof Error
              ? error.message
              : "Unknown organizations error";
        setOrganizations([]);
        setSelectedOrganizationId("");
        setOrganizationLinkState({
          kind: "error",
          title: "Не удалось загрузить организации",
          description: message,
        });
      }
    }

    void loadOrganizations();
    return () => {
      cancelled = true;
    };
  }, [accessToken]);

  useEffect(() => {
    let cancelled = false;

    async function loadAccessRequestOrganizations() {
      try {
        const payload = await apiFetch<{
          items?: AccessRequestOrganization[];
          body?: { items?: AccessRequestOrganization[] };
        }>("/v1/organizations/list?limit=500", { accessToken });
        if (cancelled) {
          return;
        }
        const items = Array.isArray(payload.items)
          ? payload.items
          : Array.isArray(payload.body?.items)
            ? payload.body.items
            : [];
        setRequestOrganizations(items);
        setRequestOrganizationId((current) => (current && items.some((item) => item.id === current) ? current : items[0]?.id || ""));
      } catch {
        if (cancelled) {
          return;
        }
        const fallback = organizations.map((item) => ({ id: item.id, name: item.name, slug: item.slug }));
        setRequestOrganizations(fallback);
        setRequestOrganizationId((current) => (current && fallback.some((item) => item.id === current) ? current : fallback[0]?.id || ""));
      }
    }

    void loadAccessRequestOrganizations();
    return () => {
      cancelled = true;
    };
  }, [accessToken, organizations]);

  useEffect(() => {
    let cancelled = false;

    async function loadOrganizationAccessQueue() {
      if (!accessToken || !selectedOrganizationId) {
        setOrganizationAccessQueue([]);
        setOrganizationAccessQueueState(initialOrganizationAccessQueueState);
        return;
      }
      setOrganizationAccessQueueState({
        kind: "working",
        title: "Загружаем заявки",
        description: "Читаем очередь заявок для выбранной организации.",
      });
      try {
        const payload = await apiFetch<{ requests?: OrganizationAccessRequest[]; body?: { requests?: OrganizationAccessRequest[] } }>(
          `/v1/organizations/${selectedOrganizationId}/access-requests`,
          { accessToken },
        );
        if (cancelled) {
          return;
        }
        const items = Array.isArray(payload.requests)
          ? payload.requests
          : Array.isArray(payload.body?.requests)
            ? payload.body.requests
            : [];
        setOrganizationAccessQueue(items);
        setOrganizationAccessQueueState({
          kind: "success",
          title: "Очередь заявок загружена",
          description: `Найдено заявок: ${items.length}.`,
        });
      } catch (error) {
        if (cancelled) {
          return;
        }
        const message =
          error instanceof APIError
            ? `${error.message}${error.code ? ` (${error.code})` : ""}`
            : error instanceof Error
              ? error.message
              : "Unknown access requests error";
        setOrganizationAccessQueue([]);
        setOrganizationAccessQueueState({
          kind: "error",
          title: "Не удалось загрузить заявки",
          description: message,
        });
      }
    }

    void loadOrganizationAccessQueue();
    return () => {
      cancelled = true;
    };
  }, [accessToken, selectedOrganizationId, groupsRefreshKey]);

  useEffect(() => {
    let cancelled = false;

    async function loadMembers() {
      if (!accessToken || !selectedGroupId) {
        setGroupAccountMembers([]);
        setGroupOrganizationMembers([]);
        setMembersState(initialMembersState);
        return;
      }

      setMembersState({
        kind: "working",
        title: "Загружаем участников",
        description: "Проверяем текущий состав аккаунтов и организаций в выбранной группе.",
      });

      try {
        const payload = await apiFetch<{
          accounts?: GroupAccountMember[];
          organizations?: GroupOrganizationMember[];
          body?: { accounts?: GroupAccountMember[]; organizations?: GroupOrganizationMember[] };
        }>(`/v1/groups/${selectedGroupId}/members`, { accessToken });
        if (cancelled) {
          return;
        }
        const accounts = Array.isArray(payload.accounts)
          ? payload.accounts
          : Array.isArray(payload.body?.accounts)
            ? payload.body.accounts
            : [];
        const organizations = Array.isArray(payload.organizations)
          ? payload.organizations
          : Array.isArray(payload.body?.organizations)
            ? payload.body.organizations
            : [];
        setGroupAccountMembers(accounts);
        setGroupOrganizationMembers(organizations);
        setMembersState({
          kind: "success",
          title: "Участники загружены",
          description: `Аккаунтов: ${accounts.length}, организаций: ${organizations.length}.`,
        });
      } catch (error) {
        if (cancelled) {
          return;
        }
        const message =
          error instanceof APIError
            ? `${error.message}${error.code ? ` (${error.code})` : ""}`
            : error instanceof Error
              ? error.message
              : "Unknown group members error";
        setGroupAccountMembers([]);
        setGroupOrganizationMembers([]);
        setMembersState({
          kind: "error",
          title: "Не удалось загрузить участников",
          description: message,
        });
      }
    }

    void loadMembers();
    return () => {
      cancelled = true;
    };
  }, [accessToken, selectedGroupId, groupsRefreshKey]);

  useEffect(() => {
    let cancelled = false;

    async function loadChannels() {
      if (!accessToken || !selectedGroupId) {
        setChannels([]);
        setSelectedChannelId("");
        setChannelsState(initialChannelsState);
        return;
      }

      setChannelsState({
        kind: "working",
        title: "Загружаем каналы",
        description: "Список каналов читается из выбранной collaboration group.",
      });

      try {
        const payload = await apiFetch<{ channels?: Channel[]; body?: { channels?: Channel[] } }>(`/v1/groups/${selectedGroupId}/channels`, { accessToken });
        if (cancelled) {
          return;
        }
        const items = Array.isArray(payload.channels) ? payload.channels : Array.isArray(payload.body?.channels) ? payload.body.channels : [];
        setChannels(items);
        setSelectedChannelId((current) => {
          if (current && items.some((item) => item.id === current)) {
            return current;
          }
          return items[0]?.id || "";
        });
        setChannelsState({
          kind: "success",
          title: items.length > 0 ? "Каналы готовы" : "В группе пока нет каналов",
          description:
            items.length > 0
              ? "Выберите канал, и справа появится лента сообщений."
              : "Если это новая группа, проверьте, что default channel уже провиженился.",
        });
      } catch (error) {
        if (cancelled) {
          return;
        }
        const message =
          error instanceof APIError
            ? `${error.message}${error.code ? ` (${error.code})` : ""}`
            : error instanceof Error
              ? error.message
              : "Unknown channels error";
        setChannels([]);
        setSelectedChannelId("");
        setChannelsState({
          kind: "error",
          title: "Не удалось загрузить каналы",
          description: message,
        });
      }
    }

    void loadChannels();
    return () => {
      cancelled = true;
    };
  }, [accessToken, selectedGroupId]);

  useEffect(() => {
    let cancelled = false;

    async function loadMessages() {
      if (!accessToken || !selectedChannelId) {
        setMessages([]);
        setMessagesState(initialMessagesState);
        return;
      }

      setMessagesState((current) =>
        current.kind === "success"
          ? current
          : {
              kind: "working",
              title: "Загружаем сообщения",
              description: "Timeline подтягивается из backend и будет обновляться каждые 5 секунд.",
            },
      );

      try {
        const payload = await apiFetch<{ messages?: Message[]; body?: { messages?: Message[] } }>(`/v1/channels/${selectedChannelId}/messages?limit=100`, { accessToken });
        if (cancelled) {
          return;
        }
        const items = Array.isArray(payload.messages) ? payload.messages : Array.isArray(payload.body?.messages) ? payload.body.messages : [];
        setMessages(items);
        setMessagesState({
          kind: "success",
          title: items.length > 0 ? "Лента синхронизирована" : "Канал пока пуст",
          description:
            items.length > 0
              ? "Новые сообщения можно отправлять снизу. Сейчас обновление работает через polling."
              : "Отправьте первое сообщение в этот канал.",
        });
      } catch (error) {
        if (cancelled) {
          return;
        }
        const message =
          error instanceof APIError
            ? `${error.message}${error.code ? ` (${error.code})` : ""}`
            : error instanceof Error
              ? error.message
              : "Unknown messages error";
        setMessages([]);
        setMessagesState({
          kind: "error",
          title: "Не удалось загрузить сообщения",
          description: message,
        });
      }
    }

    void loadMessages();

    if (!selectedChannelId) {
      return () => {
        cancelled = true;
      };
    }

    const timer = window.setInterval(() => {
      void loadMessages();
    }, 5000);

    return () => {
      cancelled = true;
      window.clearInterval(timer);
    };
  }, [accessToken, selectedChannelId, messagesRefreshKey]);

  async function handleCreateGroup(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!accessToken) {
      setGroupsState({
        kind: "error",
        title: "Нет локальной сессии",
        description: "Сначала завершите login, затем создайте группу.",
      });
      return;
    }

    setGroupsState({
      kind: "working",
      title: "Создаём группу",
      description: "Backend создаст collaboration group, owner membership и default channel.",
    });

    try {
      const payload = await apiFetch<CreatedGroup>("/v1/groups", {
        method: "POST",
        accessToken,
        bodyJSON: {
          name: groupName,
          slug: groupSlug,
        },
      });
      setSelectedGroupId(payload.id);
      setGroupsRefreshKey((value) => value + 1);
      setGroupsState({
        kind: "success",
        title: "Группа создана",
        description: "Список групп будет перечитан. После этого можно сразу открыть default channel.",
      });
    } catch (error) {
      if (error instanceof APIError && error.status === 403) {
        const groupHint = selectedGroup ? ` (${selectedGroup.name})` : "";
        const channelHint = selectedChannel ? `Канал: ${selectedChannel.name}. ` : "";
        setJoinRequestState({
          kind: "error",
          title: "Нужны права owner",
          description:
            `${channelHint}Этот endpoint доступен только owner группы. Передайте owner уже выбранный groupId: \`${selectedGroupId}\`${groupHint}. Он сможет добавить вас через \`POST /v1/groups/${selectedGroupId}/accounts\`.`,
        });
        return;
      }
      if (error instanceof APIError && error.status === 409) {
        setJoinRequestState({
          kind: "success",
          title: "Уже в группе",
          description: "Аккаунт уже является участником группы выбранного канала. Можно просто открыть канал и писать сообщения.",
        });
        return;
      }
      const message =
        error instanceof APIError
          ? `${error.message}${error.code ? ` (${error.code})` : ""}`
          : error instanceof Error
            ? error.message
            : "Unknown create group error";
      setGroupsState({
        kind: "error",
        title: "Создание группы не удалось",
        description: message,
      });
    }
  }

  async function handleSendMessage(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!accessToken || !selectedChannelId) {
      setComposerState({
        kind: "error",
        title: "Нет активного канала",
        description: "Сначала выберите группу и канал.",
      });
      return;
    }
    if (draft.trim() === "" && pendingAttachments.length === 0) {
      setComposerState({
        kind: "error",
        title: "Пустое сообщение",
        description: "Введите текст или прикрепите файл.",
      });
      return;
    }

    setComposerState({
      kind: "working",
      title: "Отправляем сообщение",
      description: "Запрос идёт в `POST /v1/channels/{channelId}/messages`.",
    });

    try {
      const bodyJSON: { body: string; attachmentObjectIds?: string[] } = {
        body: draft.trim(),
      };
      if (pendingAttachments.length > 0) {
        bodyJSON.attachmentObjectIds = pendingAttachments.map((a) => a.objectId);
      }
      await apiFetch<Message>(`/v1/channels/${selectedChannelId}/messages`, {
        method: "POST",
        accessToken,
        bodyJSON,
      });
      setDraft("");
      setPendingAttachments([]);
      setMessagesRefreshKey((value) => value + 1);
      setComposerState({
        kind: "success",
        title: "Сообщение отправлено",
        description: "Лента перечитается немедленно и продолжит обновляться polling-ом.",
      });
    } catch (error) {
      if (error instanceof APIError) {
        const code = (error.code || "").trim().toUpperCase();
        const message = error.message.toLowerCase();
        if (error.status === 409 || code === "MEMBER_EXIST" || message.includes("member already exists")) {
          setOrganizationAccessRequestState({
            kind: "success",
            title: "Доступ уже есть",
            description: "Вы уже состоите в этой организации. Повторная заявка не требуется.",
          });
          return;
        }
      }
      const message =
        error instanceof APIError
          ? `${error.message}${error.code ? ` (${error.code})` : ""}`
          : error instanceof Error
            ? error.message
            : "Unknown send message error";
      setComposerState({
        kind: "error",
        title: "Не удалось отправить сообщение",
        description: message,
      });
    }
  }

  async function handleFileSelect(event: React.ChangeEvent<HTMLInputElement>) {
    const files = event.target.files;
    if (!files?.length || !accessToken || !selectedChannelId) return;

    setAttachmentUploadState({ kind: "working", title: "Загружаем файлы", description: "..." });

    const uploaded: Array<{ objectId: string; fileName: string; sizeBytes: number }> = [];
    for (let i = 0; i < files.length; i++) {
      const file = files[i];
      const isImage = file.type.startsWith("image/");
      const isVideo = file.type.startsWith("video/");
      const limitBytes = isImage ? 15 * 1024 * 1024 : isVideo ? 100 * 1024 * 1024 : 10 * 1024 * 1024;
      if (file.size > limitBytes) {
        const limitMB = (limitBytes / (1024 * 1024)).toFixed(0);
        setAttachmentUploadState({
          kind: "error",
          title: `Файл слишком большой: ${file.name}`,
          description: `Лимит: ${limitMB} MB (документы 10 MB, фото 15 MB, видео 100 MB). Ваш файл: ${(file.size / (1024 * 1024)).toFixed(1)} MB.`,
        });
        return;
      }

      const formData = new FormData();
      formData.append("file", file);

      try {
        // Use /api/upload proxy: Next.js rewrites can fail to forward multipart/form-data correctly
        const payload = await apiFetch<{ objectId?: string; fileName?: string; sizeBytes?: number; body?: { objectId: string; fileName: string; sizeBytes: number } }>(
          `/api/upload/v1/channels/${selectedChannelId}/attachments/upload`,
          {
            method: "POST",
            accessToken,
            body: formData,
          },
        );
        const obj = (payload as { body?: { objectId: string; fileName: string; sizeBytes: number } }).body ?? payload;
        const objectId = obj.objectId;
        if (objectId) {
          uploaded.push({
            objectId,
            fileName: obj.fileName ?? file.name,
            sizeBytes: obj.sizeBytes ?? file.size,
          });
        }
      } catch (error) {
        const msg = error instanceof APIError ? error.message : error instanceof Error ? error.message : "Upload failed";
        setAttachmentUploadState({ kind: "error", title: `Ошибка: ${file.name}`, description: msg });
        return;
      }
    }

    setPendingAttachments((prev) => [...prev, ...uploaded]);
    setAttachmentUploadState({ kind: "success", title: "Файлы загружены", description: `${uploaded.length} файл(ов) готовы к отправке.` });
    event.target.value = "";
  }

  function removePendingAttachment(objectId: string) {
    setPendingAttachments((prev) => prev.filter((a) => a.objectId !== objectId));
  }

  useEffect(() => {
    setPendingAttachments([]);
  }, [selectedChannelId]);

  async function handleRequestJoinSelectedChannel() {
    if (!accessToken || !selectedGroupId || !selectedChannelId) {
      setJoinRequestState({
        kind: "error",
        title: "Нет выбранного канала",
        description: "Сначала выберите группу и канал, затем отправьте запрос на присоединение.",
      });
      return;
    }

    if (selectedGroup?.membershipSource === "account") {
      setJoinRequestState({
        kind: "success",
        title: "Вы уже в группе",
        description: "Для выбранной группы у вас уже есть прямой доступ. Можно просто открыть канал и писать сообщения.",
      });
      return;
    }

    setJoinRequestState({
      kind: "working",
      title: "Отправляем запрос",
      description: "Запрос уходит в `POST /v1/groups/{groupId}/accounts` с ролью member.",
    });

    try {
      const me = await apiFetch<{ id: string }>("/v1/auth/me", { accessToken });
      await apiFetch(`/v1/groups/${selectedGroupId}/accounts`, {
        method: "POST",
        accessToken,
        bodyJSON: {
          accountId: me.id,
          role: "member",
        },
      });

      setGroupsRefreshKey((value) => value + 1);
      setJoinRequestState({
        kind: "success",
        title: "Запрос отправлен",
        description: "Backend принял запрос на добавление в группу выбранного канала.",
      });
    } catch (error) {
      if (error instanceof APIError && error.status === 409) {
        setJoinRequestState({
          kind: "success",
          title: "Вы уже в группе",
          description: "Backend подтвердил, что аккаунт уже состоит в этой группе. Дополнительный запрос не нужен.",
        });
        return;
      }

      if (error instanceof APIError && error.status === 403) {
        setJoinRequestState({
          kind: "error",
          title: "Нужны права owner",
          description:
            "Этот endpoint может вызывать только owner группы. Если нужен доступ другому аккаунту, owner должен добавить его через `POST /v1/groups/{groupId}/accounts`.",
        });
        return;
      }

      const message =
        error instanceof APIError
          ? `${error.message}${error.code ? ` (${error.code})` : ""}`
          : error instanceof Error
            ? error.message
            : "Unknown join request error";
      setJoinRequestState({
        kind: "error",
        title: "Не удалось отправить запрос",
        description: message,
      });
    }
  }

  async function handleAttachOrganizationToGroup() {
    if (!accessToken || !selectedGroupId) {
      setOrganizationLinkState({
        kind: "error",
        title: "Нет выбранной группы",
        description: "Сначала выберите группу, затем добавьте организацию.",
      });
      return;
    }
    if (!selectedOrganizationId) {
      setOrganizationLinkState({
        kind: "error",
        title: "Нет выбранной организации",
        description: "Выберите организацию из списка и повторите.",
      });
      return;
    }

    setOrganizationLinkState({
      kind: "working",
      title: "Добавляем организацию",
      description: "Запрос уходит в `POST /v1/groups/{groupId}/organizations`.",
    });

    try {
      await apiFetch(`/v1/groups/${selectedGroupId}/organizations`, {
        method: "POST",
        accessToken,
        bodyJSON: { organizationId: selectedOrganizationId },
      });
      setGroupsRefreshKey((value) => value + 1);
      setOrganizationLinkState({
        kind: "success",
        title: "Организация добавлена",
        description: `${selectedOrganization?.name || "Organization"} теперь подключена к выбранной группе.`,
      });
    } catch (error) {
      if (error instanceof APIError && error.status === 409) {
        setOrganizationLinkState({
          kind: "success",
          title: "Организация уже в группе",
          description: "Повторно добавлять не нужно: backend подтвердил существующее членство организации.",
        });
        return;
      }
      if (error instanceof APIError && error.status === 403) {
        setOrganizationLinkState({
          kind: "error",
          title: "Нужны права owner",
          description: "Добавлять организации в группу может только owner этой группы.",
        });
        return;
      }
      const message =
        error instanceof APIError
          ? `${error.message}${error.code ? ` (${error.code})` : ""}`
          : error instanceof Error
            ? error.message
            : "Unknown add organization error";
      setOrganizationLinkState({
        kind: "error",
        title: "Не удалось добавить организацию",
        description: message,
      });
    }
  }

  async function handleCreateOrganizationAccessRequest() {
    const organizationId = requestOrganizationId.trim();
    if (!accessToken || organizationId === "") {
      setOrganizationAccessRequestState({
        kind: "error",
        title: "Нужен organizationId",
        description: "Укажите UUID организации и повторите запрос.",
      });
      return;
    }
    setOrganizationAccessRequestState({
      kind: "working",
      title: "Отправляем заявку",
      description: "Создаём заявку на доступ к организации.",
    });
    try {
      await apiFetch(`/v1/organizations/${organizationId}/access-requests`, {
        method: "POST",
        accessToken,
        bodyJSON: {
          role: requestRole,
          message: requestMessage.trim() || undefined,
        },
      });
      setOrganizationAccessRequestState({
        kind: "success",
        title: "Заявка отправлена",
        description: "Owner/admin организации сможет рассмотреть заявку в очереди access requests.",
      });
      setRequestMessage("");
      if (selectedOrganizationId === organizationId) {
        setGroupsRefreshKey((value) => value + 1);
      }
    } catch (error) {
      const message =
        error instanceof APIError
          ? `${error.message}${error.code ? ` (${error.code})` : ""}`
          : error instanceof Error
            ? error.message
            : "Unknown create access request error";
      setOrganizationAccessRequestState({
        kind: "error",
        title: "Не удалось отправить заявку",
        description: message,
      });
    }
  }

  async function handleReviewOrganizationAccessRequest(requestId: string, decision: "approve" | "reject") {
    if (!accessToken || !selectedOrganizationId) {
      return;
    }
    setOrganizationAccessQueueState({
      kind: "working",
      title: decision === "approve" ? "Одобряем заявку" : "Отклоняем заявку",
      description: "Обновляем статус заявки в backend.",
    });
    try {
      await apiFetch(`/v1/organizations/${selectedOrganizationId}/access-requests/${requestId}/${decision}`, {
        method: "POST",
        accessToken,
        bodyJSON: {},
      });
      setGroupsRefreshKey((value) => value + 1);
      setOrganizationAccessQueueState({
        kind: "success",
        title: "Заявка обработана",
        description: "Очередь будет перечитана автоматически.",
      });
    } catch (error) {
      const message =
        error instanceof APIError
          ? `${error.message}${error.code ? ` (${error.code})` : ""}`
          : error instanceof Error
            ? error.message
            : "Unknown review access request error";
      setOrganizationAccessQueueState({
        kind: "error",
        title: "Не удалось обработать заявку",
        description: message,
      });
    }
  }

  return (
    <section className="chat-layout">
      <div className="chat-sidebar">
        <Panel title="Группы" eyebrow="My groups">
          <div className={`status-card ${groupsState.kind === "error" ? "error" : groupsState.kind === "success" ? "success" : "info"}`}>
            <strong>{groupsState.title}</strong>
            <p className="status-copy">{groupsState.description}</p>
          </div>

          <form className="form-grid chat-create-form" onSubmit={handleCreateGroup}>
            <div className="form-row">
              <label className="form-label" htmlFor="chat-group-name">
                Новая группа
              </label>
              <input
                id="chat-group-name"
                className="text-input"
                value={groupName}
                onChange={(event) => {
                  const value = event.target.value;
                  setGroupName(value);
                  setGroupSlug((current) => (current === "" || current === slugify(groupName) ? slugify(value) : current));
                }}
                placeholder="Product Team"
              />
            </div>
            <div className="form-row">
              <label className="form-label" htmlFor="chat-group-slug">
                Slug
              </label>
              <input
                id="chat-group-slug"
                className="text-input"
                value={groupSlug}
                onChange={(event) => setGroupSlug(event.target.value)}
                placeholder="product-team"
              />
            </div>
            <button className="button primary" type="submit">
              Создать группу
            </button>
          </form>

          <div className="chat-list">
            {groups.map((group) => (
              <button
                key={group.id}
                type="button"
                className={`chat-list-item ${group.id === selectedGroupId ? "active" : ""}`}
                onClick={() => setSelectedGroupId(group.id)}
              >
                <strong>{group.name}</strong>
                <span className="muted">
                  {group.membershipSource === "account"
                    ? `Прямой доступ${group.membershipRole ? `: ${group.membershipRole}` : ""}`
                    : "Доступ через organization"}
                </span>
              </button>
            ))}
            {groups.length === 0 ? <p className="muted">Групп пока нет.</p> : null}
          </div>
        </Panel>

        <Panel title="Каналы" eyebrow={selectedGroup ? selectedGroup.slug : "No group selected"}>
          <div className={`status-card ${channelsState.kind === "error" ? "error" : channelsState.kind === "success" ? "success" : "info"}`}>
            <strong>{channelsState.title}</strong>
            <p className="status-copy">{channelsState.description}</p>
          </div>
          <div className={`status-card ${joinRequestState.kind === "error" ? "error" : joinRequestState.kind === "success" ? "success" : "info"}`}>
            <strong>{joinRequestState.title}</strong>
            <p className="status-copy">{joinRequestState.description}</p>
          </div>
          <div className="button-row">
            <button className="button secondary" type="button" onClick={() => void handleRequestJoinSelectedChannel()} disabled={!selectedChannelId}>
              Запросить присоединение к каналу
            </button>
          </div>

          <div className="chat-list">
            {channels.map((channel) => (
              <button
                key={channel.id}
                type="button"
                className={`chat-list-item ${channel.id === selectedChannelId ? "active" : ""}`}
                onClick={() => setSelectedChannelId(channel.id)}
              >
                <strong>{channel.name}</strong>
                <span className="muted">{channel.isDefault ? "Default channel" : channel.slug}</span>
              </button>
            ))}
            {selectedGroup && channels.length === 0 ? <p className="muted">У выбранной группы нет доступных каналов.</p> : null}
          </div>
        </Panel>

        <Panel title="Группы организаций" eyebrow={selectedGroup ? selectedGroup.name : "Сначала выберите группу"}>
          <div
            className={`status-card ${organizationLinkState.kind === "error" ? "error" : organizationLinkState.kind === "success" ? "success" : "info"}`}
          >
            <strong>{organizationLinkState.title}</strong>
            <p className="status-copy">{organizationLinkState.description}</p>
          </div>

          <div className="form-grid">
            <div className="form-row">
              <label className="form-label" htmlFor="chat-organization-selector">
                Организация
              </label>
              <select
                id="chat-organization-selector"
                className="text-input"
                value={selectedOrganizationId}
                onChange={(event) => setSelectedOrganizationId(event.target.value)}
              >
                <option value="">Выберите организацию</option>
                {organizations.map((organization) => (
                  <option key={organization.id} value={organization.id}>
                    {organization.name} ({organization.membershipRole})
                  </option>
                ))}
              </select>
            </div>
            <button
              className="button secondary"
              type="button"
              onClick={() => void handleAttachOrganizationToGroup()}
              disabled={!selectedGroupId || !selectedOrganizationId}
            >
              Добавить организацию в группу
            </button>
          </div>
        </Panel>

        <Panel title="Запрос доступа к организации" eyebrow="Self-service">
          <div
            className={`status-card ${organizationAccessRequestState.kind === "error" ? "error" : organizationAccessRequestState.kind === "success" ? "success" : "info"}`}
          >
            <strong>{organizationAccessRequestState.title}</strong>
            <p className="status-copy">{organizationAccessRequestState.description}</p>
          </div>

          <div className="form-grid">
            <div className="form-row">
              <label className="form-label" htmlFor="chat-request-organization">
                Организация
              </label>
              <select
                id="chat-request-organization"
                className="text-input"
                value={requestOrganizationId}
                onChange={(event) => setRequestOrganizationId(event.target.value)}
              >
                <option value="">Выберите организацию</option>
                {requestOrganizations.map((organization) => (
                  <option key={organization.id} value={organization.id}>
                    {organization.name} ({organization.slug})
                  </option>
                ))}
              </select>
            </div>
            <div className="form-row">
              <label className="form-label" htmlFor="chat-request-role">
                Роль
              </label>
              <select id="chat-request-role" className="text-input" value={requestRole} onChange={(event) => setRequestRole(event.target.value)}>
                <option value="member">member</option>
                <option value="viewer">viewer</option>
                <option value="manager">manager</option>
                <option value="admin">admin</option>
                <option value="owner">owner</option>
              </select>
            </div>
            <div className="form-row">
              <label className="form-label" htmlFor="chat-request-message">
                Комментарий
              </label>
              <textarea
                id="chat-request-message"
                className="textarea"
                value={requestMessage}
                onChange={(event) => setRequestMessage(event.target.value)}
                rows={3}
                placeholder="Почему вам нужен доступ"
              />
            </div>
            <button
              className="button secondary"
              type="button"
              onClick={() => void handleCreateOrganizationAccessRequest()}
              disabled={!requestOrganizationId}
            >
              Отправить заявку на доступ
            </button>
          </div>
        </Panel>

        <Panel title="Очередь заявок организации" eyebrow={selectedOrganization ? selectedOrganization.name : "Выберите организацию"}>
          <div
            className={`status-card ${organizationAccessQueueState.kind === "error" ? "error" : organizationAccessQueueState.kind === "success" ? "success" : "info"}`}
          >
            <strong>{organizationAccessQueueState.title}</strong>
            <p className="status-copy">{organizationAccessQueueState.description}</p>
          </div>
          <div className="chat-list">
            {organizationAccessQueue.map((request) => (
              <div key={request.id} className="chat-list-item">
                <strong>{request.requesterAccountId.slice(0, 8)}</strong>
                <span className="muted">
                  {request.requestedRole} · {request.status}
                </span>
                {request.message ? <span className="muted">{request.message}</span> : null}
                {request.status === "pending" ? (
                  <div className="button-row">
                    <button className="button secondary" type="button" onClick={() => void handleReviewOrganizationAccessRequest(request.id, "approve")}>
                      Одобрить
                    </button>
                    <button className="button secondary" type="button" onClick={() => void handleReviewOrganizationAccessRequest(request.id, "reject")}>
                      Отклонить
                    </button>
                  </div>
                ) : null}
              </div>
            ))}
            {selectedOrganizationId && organizationAccessQueue.length === 0 ? <p className="muted">Заявок пока нет.</p> : null}
          </div>
        </Panel>

        <Panel title="Участники группы" eyebrow={selectedGroup ? selectedGroup.slug : "No group selected"}>
          <div className={`status-card ${membersState.kind === "error" ? "error" : membersState.kind === "success" ? "success" : "info"}`}>
            <strong>{membersState.title}</strong>
            <p className="status-copy">{membersState.description}</p>
          </div>

          <div className="chat-list">
            <p className="muted">Account members</p>
            {groupAccountMembers.map((member) => (
              <div key={member.id} className="chat-list-item">
                <strong>{member.displayName?.trim() || member.email}</strong>
                <span className="muted">
                  {member.role} · {member.isActive ? "active" : "inactive"}
                </span>
              </div>
            ))}
            {selectedGroup && groupAccountMembers.length === 0 ? <p className="muted">Прямых account-участников нет.</p> : null}
          </div>

          <div className="chat-list">
            <p className="muted">Organization members</p>
            {groupOrganizationMembers.map((member) => (
              <div key={member.id} className="chat-list-item">
                <strong>{member.name}</strong>
                <span className="muted">
                  {member.slug} · {member.isActive ? "active" : "inactive"}
                </span>
              </div>
            ))}
            {selectedGroup && groupOrganizationMembers.length === 0 ? <p className="muted">Организации пока не подключены.</p> : null}
          </div>
        </Panel>
      </div>

      <div className="chat-main">
        <Panel
          title={selectedChannel ? selectedChannel.name : "Лента"}
          eyebrow={selectedGroup ? selectedGroup.name : "Select a group"}
          actions={
            <button className="button secondary" type="button" onClick={() => setMessagesRefreshKey((value) => value + 1)}>
              Обновить
            </button>
          }
        >
          <div className="chat-channel-view">
            <div className="chat-timeline-wrap">
              {messagesState.kind === "error" ? (
                <div className={`status-card error`}>
                  <strong>{messagesState.title}</strong>
                  <p className="status-copy">{messagesState.description}</p>
                </div>
              ) : null}
              <div className="chat-timeline">
                {messages.map((message) => (
                  <article key={message.id} className="chat-message">
                    <div className="chat-message-head">
                      <strong>{authorLabel(message)}</strong>
                      <span className="muted">
                        #{message.channelSeq} · {formatDate(message.createdAt)}
                        {message.editedAt ? " · edited" : ""}
                      </span>
                    </div>
                    <p className={`chat-message-body ${message.deletedAt ? "deleted" : ""}`}>
                      {message.deletedAt ? "Сообщение удалено" : message.body || "Вложение без текста"}
                    </p>
                    {Array.isArray(message.attachments) && message.attachments.length > 0 ? (
                      <p className="muted">Вложений: {message.attachments.length}</p>
                    ) : null}
                  </article>
                ))}
                {selectedChannel && messages.length === 0 && messagesState.kind !== "error" ? <p className="muted">Сообщений пока нет.</p> : null}
                {!selectedChannel ? <p className="muted">Слева выберите группу и канал, чтобы открыть chat timeline.</p> : null}
              </div>
            </div>
            <div className="chat-composer-wrap">
              {composerState.kind === "error" ? (
                <div className={`status-card error`}>
                  <strong>{composerState.title}</strong>
                  <p className="status-copy">{composerState.description}</p>
                </div>
              ) : null}
              <form className="chat-composer" onSubmit={handleSendMessage}>
                <div className="form-row">
                  <textarea
                    id="chat-draft"
                    className="textarea chat-composer-input"
                    value={draft}
                    onChange={(event) => setDraft(event.target.value)}
                    placeholder="Напишите сообщение или прикрепите файл"
                    rows={3}
                  />
                </div>
                <div className="form-row">
                  <input
                    ref={fileInputRef}
                    type="file"
                    multiple
                    className="visually-hidden"
                    accept="*/*"
                    onChange={handleFileSelect}
                  />
                  <div className="button-row">
                    <button
                      type="button"
                      className="button secondary"
                      onClick={() => fileInputRef.current?.click()}
                      disabled={!selectedChannelId}
                    >
                      Прикрепить файл
                    </button>
                    <button className="button primary" type="submit" disabled={!selectedChannelId || (draft.trim() === "" && pendingAttachments.length === 0)}>
                      Отправить
                    </button>
                  </div>
                </div>
                {attachmentUploadState.kind !== "idle" && attachmentUploadState.title ? (
                  <div className={`status-card ${attachmentUploadState.kind === "error" ? "error" : attachmentUploadState.kind === "success" ? "success" : "info"}`}>
                    <strong>{attachmentUploadState.title}</strong>
                    {attachmentUploadState.description ? <p className="status-copy">{attachmentUploadState.description}</p> : null}
                  </div>
                ) : null}
                {pendingAttachments.length > 0 ? (
                  <div className="chat-attachments-pending">
                    <p className="form-label">Вложения ({pendingAttachments.length}):</p>
                    <ul className="chat-attachment-list">
                      {pendingAttachments.map((a) => (
                        <li key={a.objectId} className="chat-attachment-item">
                          <span>{a.fileName}</span>
                          <span className="muted">{(a.sizeBytes / 1024).toFixed(1)} KB</span>
                          <button type="button" className="button secondary" onClick={() => removePendingAttachment(a.objectId)}>
                            Удалить
                          </button>
                        </li>
                      ))}
                    </ul>
                  </div>
                ) : null}
              </form>
            </div>
          </div>
        </Panel>
      </div>
    </section>
  );
}
