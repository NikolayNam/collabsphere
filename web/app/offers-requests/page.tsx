"use client";

import { FormEvent, useEffect, useMemo, useState } from "react";

import { Panel } from "@/components/panel";
import { APIError, apiFetch } from "@/lib/api";
import { readStoredTokens } from "@/lib/auth";

type MyOrganization = {
  id: string;
  name: string;
  slug: string;
  membershipRole: string;
};

type BoardOrder = {
  id: string;
  number: string;
  title: string;
  status: string;
  organizationId: string;
  createdAt: string;
  items?: Array<{ id: string; productName?: string; categoryId?: string; note?: string }>;
  comments?: Array<{ id: string; organizationName?: string; organizationId: string; comment: string; createdAt: string }>;
};

type BoardOffer = {
  id: string;
  orderId: string;
  organizationId: string;
  status: string;
  comment?: string;
  createdAt: string;
  items?: Array<{ id: string; customTitle?: string; categoryId?: string; productId?: string; priceAmount?: string; currencyCode?: string }>;
  comments?: Array<{ id: string; organizationName?: string; organizationId: string; comment: string; createdAt: string }>;
};

function toErrorText(error: unknown): string {
  if (error instanceof APIError) {
    return `${error.message}${error.code ? ` (${error.code})` : ""}`;
  }
  if (error instanceof Error) {
    return error.message;
  }
  return "Unknown error";
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

export default function OffersRequestsPage() {
  const accessToken = useMemo(() => readStoredTokens()?.accessToken || null, []);

  const [organizations, setOrganizations] = useState<MyOrganization[]>([]);
  const [creatorOrganizationId, setCreatorOrganizationId] = useState("");
  const [createMode, setCreateMode] = useState<"order" | "offer">("order");
  const [orderTitle, setOrderTitle] = useState("");
  const [orderDescription, setOrderDescription] = useState("");
  const [orderItemsText, setOrderItemsText] = useState("");
  const [orderComment, setOrderComment] = useState("");
  const [offerItemsText, setOfferItemsText] = useState("");
  const [offerComment, setOfferComment] = useState("");
  const [orders, setOrders] = useState<BoardOrder[]>([]);
  const [selectedOrderId, setSelectedOrderId] = useState("");
  const [selectedOrder, setSelectedOrder] = useState<BoardOrder | null>(null);
  const [offers, setOffers] = useState<BoardOffer[]>([]);
  const [status, setStatus] = useState("Загрузка заявок...");

  useEffect(() => {
    let cancelled = false;

    async function loadOrganizations() {
      if (!accessToken) {
        if (!cancelled) {
          setOrganizations([]);
          setCreatorOrganizationId("");
        }
        return;
      }
      try {
        const payload = await apiFetch<{ data?: MyOrganization[]; body?: { data?: MyOrganization[] } }>("/v1/organizations/my", { accessToken });
        if (cancelled) {
          return;
        }
        const items = Array.isArray(payload.data) ? payload.data : Array.isArray(payload.body?.data) ? payload.body.data : [];
        setOrganizations(items);
        setCreatorOrganizationId((current) => (current && items.some((item) => item.id === current) ? current : items[0]?.id || ""));
      } catch (error) {
        if (!cancelled) {
          setStatus(`Не удалось загрузить организации: ${toErrorText(error)}`);
        }
      }
    }

    async function loadOrders() {
      if (!accessToken) {
        if (!cancelled) {
          setStatus("Нужна авторизация: войдите через /login.");
          setOrders([]);
          setSelectedOrderId("");
        }
        return;
      }
      try {
        const payload = await apiFetch<{ items?: BoardOrder[]; body?: { items?: BoardOrder[] } }>("/v1/sales/orders?limit=100", { accessToken });
        if (cancelled) {
          return;
        }
        const items = Array.isArray(payload.items) ? payload.items : Array.isArray(payload.body?.items) ? payload.body.items : [];
        setOrders(items);
        setSelectedOrderId((current) => (current && items.some((item) => item.id === current) ? current : items[0]?.id || ""));
        setStatus(`Заявок: ${items.length}`);
      } catch (error) {
        if (!cancelled) {
          setStatus(`Не удалось загрузить заявки: ${toErrorText(error)}`);
        }
      }
    }

    void loadOrganizations();
    void loadOrders();
    return () => {
      cancelled = true;
    };
  }, [accessToken]);

  useEffect(() => {
    let cancelled = false;

    async function loadOrderDetails() {
      if (!accessToken || !selectedOrderId) {
        setSelectedOrder(null);
        setOffers([]);
        return;
      }
      try {
        const [order, offersPayload] = await Promise.all([
          apiFetch<BoardOrder>(`/v1/sales/orders/${selectedOrderId}`, { accessToken }),
          apiFetch<{ items?: BoardOffer[]; body?: { items?: BoardOffer[] } }>(`/v1/sales/orders/${selectedOrderId}/offers`, { accessToken }),
        ]);
        if (cancelled) {
          return;
        }
        const items = Array.isArray(offersPayload.items)
          ? offersPayload.items
          : Array.isArray(offersPayload.body?.items)
            ? offersPayload.body.items
            : [];
        setSelectedOrder(order);
        setOffers(items);
      } catch (error) {
        if (!cancelled) {
          setStatus(`Не удалось загрузить предложения: ${toErrorText(error)}`);
        }
      }
    }

    void loadOrderDetails();
    return () => {
      cancelled = true;
    };
  }, [accessToken, selectedOrderId]);

  async function handleCreateOrder(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!accessToken || !creatorOrganizationId.trim()) {
      setStatus("Выберите организацию для создания заявки.");
      return;
    }
    try {
      const lines = orderItemsText
        .split("\n")
        .map((item) => item.trim())
        .filter(Boolean);
      await apiFetch("/v1/sales/orders", {
        method: "POST",
        accessToken,
        bodyJSON: {
          organizationId: creatorOrganizationId.trim(),
          title: orderTitle.trim(),
          description: orderDescription.trim() || undefined,
          comment: orderComment.trim() || undefined,
          items: lines.map((line) => ({ productName: line })),
        },
      });
      setOrderTitle("");
      setOrderDescription("");
      setOrderItemsText("");
      setOrderComment("");

      const refreshed = await apiFetch<{ items?: BoardOrder[]; body?: { items?: BoardOrder[] } }>("/v1/sales/orders?limit=100", { accessToken });
      const items = Array.isArray(refreshed.items) ? refreshed.items : Array.isArray(refreshed.body?.items) ? refreshed.body.items : [];
      setOrders(items);
      if (items.length > 0) {
        setSelectedOrderId(items[0].id);
      }
      setStatus("Заявка создана.");
    } catch (error) {
      setStatus(`Не удалось создать заявку: ${toErrorText(error)}`);
    }
  }

  async function handleCreateOffer(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!accessToken || !creatorOrganizationId.trim()) {
      setStatus("Выберите организацию для создания предложения.");
      return;
    }
    if (!selectedOrderId) {
      setStatus("Сначала выберите заявку, к которой отправить предложение.");
      return;
    }
    try {
      const lines = offerItemsText
        .split("\n")
        .map((item) => item.trim())
        .filter(Boolean);
      await apiFetch(`/v1/sales/orders/${selectedOrderId}/offers`, {
        method: "POST",
        accessToken,
        bodyJSON: {
          organizationId: creatorOrganizationId.trim(),
          comment: offerComment.trim() || undefined,
          items: lines.map((line) => ({ customTitle: line })),
        },
      });
      setOfferItemsText("");
      setOfferComment("");

      const offersPayload = await apiFetch<{ items?: BoardOffer[]; body?: { items?: BoardOffer[] } }>(`/v1/sales/orders/${selectedOrderId}/offers`, { accessToken });
      const items = Array.isArray(offersPayload.items) ? offersPayload.items : Array.isArray(offersPayload.body?.items) ? offersPayload.body.items : [];
      setOffers(items);
      setStatus("Предложение отправлено.");
    } catch (error) {
      setStatus(`Не удалось создать предложение: ${toErrorText(error)}`);
    }
  }

  return (
    <>
      <Panel title="Заявки и предложения" eyebrow="Sales board">
        <p className="muted">Отдельная страница для просмотра заявок (orders) и предложений (offers) без ручных вызовов API.</p>
        <div className="form-row">
          <label className="form-label" htmlFor="offers-requests-org">
            Организация
          </label>
          <select id="offers-requests-org" className="text-input" value={creatorOrganizationId} onChange={(event) => setCreatorOrganizationId(event.target.value)}>
            <option value="">Выберите организацию</option>
            {organizations.map((organization) => (
              <option key={organization.id} value={organization.id}>
                {organization.name} ({organization.membershipRole})
              </option>
            ))}
          </select>
        </div>
        <div className="button-row">
          <button className={`button ${createMode === "order" ? "primary" : "secondary"}`} type="button" onClick={() => setCreateMode("order")}>
            Создать заявку
          </button>
          <button className={`button ${createMode === "offer" ? "primary" : "secondary"}`} type="button" onClick={() => setCreateMode("offer")}>
            Создать предложение
          </button>
        </div>
        {createMode === "order" ? (
          <form className="form-grid" onSubmit={handleCreateOrder}>
            <h3>Новая заявка</h3>
            <input
              className="text-input"
              value={orderTitle}
              onChange={(event) => setOrderTitle(event.target.value)}
              placeholder="Название заявки"
              required
            />
            <textarea
              className="textarea"
              rows={3}
              value={orderDescription}
              onChange={(event) => setOrderDescription(event.target.value)}
              placeholder="Описание"
            />
            <textarea
              className="textarea"
              rows={3}
              value={orderItemsText}
              onChange={(event) => setOrderItemsText(event.target.value)}
              placeholder="Позиции заявки (по одной в строке)"
            />
            <textarea
              className="textarea"
              rows={2}
              value={orderComment}
              onChange={(event) => setOrderComment(event.target.value)}
              placeholder="Комментарий"
            />
            <div className="button-row">
              <button className="button primary" type="submit" disabled={!creatorOrganizationId}>
                Создать заявку
              </button>
            </div>
          </form>
        ) : (
          <form className="form-grid" onSubmit={handleCreateOffer}>
            <h3>Новое предложение</h3>
            <p className="muted">
              Выбранная заявка: {selectedOrder ? `${selectedOrder.title} (${selectedOrder.number})` : "не выбрана"}
            </p>
            <textarea
              className="textarea"
              rows={3}
              value={offerItemsText}
              onChange={(event) => setOfferItemsText(event.target.value)}
              placeholder="Позиции предложения (по одной в строке)"
            />
            <textarea
              className="textarea"
              rows={2}
              value={offerComment}
              onChange={(event) => setOfferComment(event.target.value)}
              placeholder="Комментарий к предложению"
            />
            <div className="button-row">
              <button className="button primary" type="submit" disabled={!creatorOrganizationId || !selectedOrderId}>
                Создать предложение
              </button>
            </div>
          </form>
        )}
        <div className="status-card info">
          <strong>Статус</strong>
          <p className="status-copy">{status}</p>
        </div>
      </Panel>

      <section className="split">
        <Panel title="Заявки" eyebrow="GET /v1/sales/orders">
          <div className="selection-list">
            {orders.map((order) => (
              <button
                key={order.id}
                type="button"
                className={`selection-card ${selectedOrderId === order.id ? "active" : ""}`}
                onClick={() => setSelectedOrderId(order.id)}
              >
                <strong>{order.title}</strong>
                <span className="muted">
                  <code>{order.number}</code> · {order.status}
                </span>
              </button>
            ))}
          </div>
          {orders.length === 0 ? <p className="muted">Пока нет доступных заявок.</p> : null}
        </Panel>

        <Panel title="Предложения по заявке" eyebrow={selectedOrder?.title || "Выберите заявку"}>
          {!selectedOrder ? (
            <p className="muted">Выберите заявку слева, чтобы увидеть предложения.</p>
          ) : (
            <div className="domain-list">
              <div className="mini-card">
                <h3>{selectedOrder.title}</h3>
                <p className="muted">
                  {selectedOrder.status} · {formatDate(selectedOrder.createdAt)}
                </p>
                <p className="muted">Позиции: {selectedOrder.items?.length || 0}</p>
                <p className="muted">Комментарии: {selectedOrder.comments?.length || 0}</p>
              </div>

              {offers.map((offer) => (
                <div key={offer.id} className="inline-panel">
                  <strong>{offer.comment?.trim() ? offer.comment : `Предложение ${offer.id.slice(0, 8)}`}</strong>
                  <span className="muted">
                    {offer.organizationId} · {offer.status} · {formatDate(offer.createdAt)}
                  </span>
                  <p className="muted">Позиции: {offer.items?.length || 0}</p>
                  {offer.comments?.length ? <p className="muted">Комментариев: {offer.comments.length}</p> : null}
                </div>
              ))}
              {offers.length === 0 ? <p className="muted">По этой заявке пока нет предложений.</p> : null}
            </div>
          )}
        </Panel>
      </section>
    </>
  );
}
