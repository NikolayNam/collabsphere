"use client";

import { FormEvent, useEffect, useMemo, useState } from "react";

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
  const [channels, setChannels] = useState<Channel[]>([]);
  const [messages, setMessages] = useState<Message[]>([]);
  const [selectedGroupId, setSelectedGroupId] = useState("");
  const [selectedChannelId, setSelectedChannelId] = useState("");
  const [groupName, setGroupName] = useState("Product Team");
  const [groupSlug, setGroupSlug] = useState("product-team");
  const [draft, setDraft] = useState("");
  const [groupsState, setGroupsState] = useState<Status>(initialGroupsState);
  const [channelsState, setChannelsState] = useState<Status>(initialChannelsState);
  const [messagesState, setMessagesState] = useState<Status>(initialMessagesState);
  const [composerState, setComposerState] = useState<Status>(initialComposerState);
  const [joinRequestState, setJoinRequestState] = useState<Status>(initialJoinRequestState);
  const [groupsRefreshKey, setGroupsRefreshKey] = useState(0);
  const [messagesRefreshKey, setMessagesRefreshKey] = useState(0);

  const accessToken = useMemo(() => readStoredTokens()?.accessToken || null, []);
  const selectedGroup = groups.find((group) => group.id === selectedGroupId) || null;
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
        const payload = await apiFetch<{ data?: MyGroup[] }>("/v1/groups/my", { accessToken });
        if (cancelled) {
          return;
        }
        const items = Array.isArray(payload.data) ? payload.data : [];
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
        const payload = await apiFetch<{ channels?: Channel[] }>(`/v1/groups/${selectedGroupId}/channels`, { accessToken });
        if (cancelled) {
          return;
        }
        const items = Array.isArray(payload.channels) ? payload.channels : [];
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
        const payload = await apiFetch<{ messages?: Message[] }>(`/v1/channels/${selectedChannelId}/messages?limit=100`, { accessToken });
        if (cancelled) {
          return;
        }
        const items = Array.isArray(payload.messages) ? payload.messages : [];
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
    if (draft.trim() === "") {
      setComposerState({
        kind: "error",
        title: "Пустое сообщение",
        description: "Введите текст перед отправкой.",
      });
      return;
    }

    setComposerState({
      kind: "working",
      title: "Отправляем сообщение",
      description: "Запрос идёт в `POST /v1/channels/{channelId}/messages`.",
    });

    try {
      await apiFetch<Message>(`/v1/channels/${selectedChannelId}/messages`, {
        method: "POST",
        accessToken,
        bodyJSON: {
          body: draft.trim(),
        },
      });
      setDraft("");
      setMessagesRefreshKey((value) => value + 1);
      setComposerState({
        kind: "success",
        title: "Сообщение отправлено",
        description: "Лента перечитается немедленно и продолжит обновляться polling-ом.",
      });
    } catch (error) {
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

  async function handleRequestJoinSelectedChannel() {
    if (!accessToken || !selectedGroupId || !selectedChannelId) {
      setJoinRequestState({
        kind: "error",
        title: "Нет выбранного канала",
        description: "Сначала выберите группу и канал, затем отправьте запрос на присоединение.",
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
      </div>

      <div className="chat-main">
        <Panel title={selectedChannel ? selectedChannel.name : "Лента"} eyebrow={selectedGroup ? selectedGroup.name : "Select a group"}>
          <div className={`status-card ${messagesState.kind === "error" ? "error" : messagesState.kind === "success" ? "success" : "info"}`}>
            <strong>{messagesState.title}</strong>
            <p className="status-copy">{messagesState.description}</p>
          </div>

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
            {selectedChannel && messages.length === 0 ? <p className="muted">Сообщений пока нет.</p> : null}
            {!selectedChannel ? <p className="muted">Слева выберите группу и канал, чтобы открыть chat timeline.</p> : null}
          </div>
        </Panel>

        <Panel
          title="Composer"
          eyebrow={selectedChannel ? `Channel ${selectedChannel.slug}` : "No channel"}
          actions={
            <button className="button secondary" type="button" onClick={() => setMessagesRefreshKey((value) => value + 1)}>
              Обновить
            </button>
          }
        >
          <div className={`status-card ${composerState.kind === "error" ? "error" : composerState.kind === "success" ? "success" : "info"}`}>
            <strong>{composerState.title}</strong>
            <p className="status-copy">{composerState.description}</p>
          </div>

          <form className="chat-composer" onSubmit={handleSendMessage}>
            <div className="form-row">
              <label className="form-label" htmlFor="chat-draft">
                Новое сообщение
              </label>
              <textarea
                id="chat-draft"
                className="textarea"
                value={draft}
                onChange={(event) => setDraft(event.target.value)}
                placeholder="Напишите сообщение в выбранный канал"
                rows={5}
              />
            </div>
            <button className="button primary" type="submit" disabled={!selectedChannelId}>
              Отправить
            </button>
          </form>
        </Panel>
      </div>
    </section>
  );
}
