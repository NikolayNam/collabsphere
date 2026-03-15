"use client";

import { FormEvent, useEffect, useMemo, useState } from "react";

import { Panel } from "@/components/panel";
import { APIError, apiFetch } from "@/lib/api";
import { readStoredTokens } from "@/lib/auth";

type CatalogCategory = {
  id: string;
  name: string;
  code: string;
};

type CatalogProduct = {
  id: string;
  name: string;
  categoryId?: string | null;
};

type BoardOrder = {
  id: string;
  organizationId: string;
  number: string;
  title: string;
  description?: string;
  status: string;
  budgetAmount?: string;
  currencyCode?: string;
  createdAt: string;
  items?: Array<{
    id: string;
    categoryId?: string;
    productName?: string;
    quantity?: string;
    unit?: string;
    note?: string;
  }>;
  comments?: Array<{
    id: string;
    organizationId: string;
    organizationName?: string;
    comment: string;
    createdAt: string;
  }>;
};

type BoardOffer = {
  id: string;
  orderId: string;
  organizationId: string;
  status: string;
  comment?: string;
  createdAt: string;
  items?: Array<{
    id: string;
    categoryId?: string;
    productId?: string;
    customTitle?: string;
    quantity?: string;
    unit?: string;
    priceAmount?: string;
    currencyCode?: string;
    note?: string;
  }>;
  comments?: Array<{
    id: string;
    organizationId: string;
    organizationName?: string;
    comment: string;
    createdAt: string;
  }>;
};

function errorText(error: unknown): string {
  if (error instanceof APIError) {
    return `${error.message}${error.code ? ` (${error.code})` : ""}`;
  }
  if (error instanceof Error) {
    return error.message;
  }
  return "Unknown error";
}

export default function MarketplacePage() {
  const accessToken = useMemo(() => readStoredTokens()?.accessToken || null, []);

  const [orders, setOrders] = useState<BoardOrder[]>([]);
  const [selectedOrderId, setSelectedOrderId] = useState("");
  const [selectedOrder, setSelectedOrder] = useState<BoardOrder | null>(null);
  const [offers, setOffers] = useState<BoardOffer[]>([]);
  const [status, setStatus] = useState("Загрузка доски...");

  const [organizationId, setOrganizationId] = useState("");
  const [orderTitle, setOrderTitle] = useState("");
  const [orderDescription, setOrderDescription] = useState("");
  const [orderComment, setOrderComment] = useState("");
  const [orderItemsText, setOrderItemsText] = useState("");

  const [offerComment, setOfferComment] = useState("");
  const [offerItemsText, setOfferItemsText] = useState("");
  const [newOrderComment, setNewOrderComment] = useState("");
  const [categories, setCategories] = useState<CatalogCategory[]>([]);
  const [products, setProducts] = useState<CatalogProduct[]>([]);
  const [orderNeededCategoryId, setOrderNeededCategoryId] = useState("");
  const [orderNeededCategoryIds, setOrderNeededCategoryIds] = useState<string[]>([]);
  const [offerCategoryId, setOfferCategoryId] = useState("");
  const [offerProductId, setOfferProductId] = useState("");
  const [offerCategoryIds, setOfferCategoryIds] = useState<string[]>([]);
  const [offerProductIds, setOfferProductIds] = useState<string[]>([]);
  const [offerCommentDraftById, setOfferCommentDraftById] = useState<Record<string, string>>({});

  async function loadOrders() {
    if (!accessToken) {
      setStatus("Нужна авторизация (/login).");
      return;
    }
    try {
      const payload = await apiFetch<{ items?: BoardOrder[] }>("/v1/sales/orders?limit=100", { accessToken });
      const items = Array.isArray(payload.items) ? payload.items : [];
      setOrders(items);
      setSelectedOrderId((current) => (current && items.some((item) => item.id === current) ? current : items[0]?.id || ""));
      setStatus(`Заказов: ${items.length}`);
    } catch (error) {
      setStatus(`Ошибка загрузки доски: ${errorText(error)}`);
    }
  }

  async function loadOrderDetails(orderId: string) {
    if (!accessToken || !orderId) {
      setSelectedOrder(null);
      setOffers([]);
      return;
    }
    try {
      const order = await apiFetch<BoardOrder>(`/v1/sales/orders/${orderId}`, { accessToken });
      const offersPayload = await apiFetch<{ items?: BoardOffer[] }>(`/v1/sales/orders/${orderId}/offers`, { accessToken });
      setSelectedOrder(order);
      setOffers(Array.isArray(offersPayload.items) ? offersPayload.items : []);
    } catch (error) {
      setStatus(`Ошибка загрузки деталей: ${errorText(error)}`);
    }
  }

  useEffect(() => {
    void loadOrders();
  }, []);

  useEffect(() => {
    void loadOrderDetails(selectedOrderId);
  }, [selectedOrderId]);

  useEffect(() => {
    let cancelled = false;
    async function loadCatalogForOrganization() {
      if (!accessToken || !organizationId.trim()) {
        setCategories([]);
        setProducts([]);
        return;
      }
      try {
        const [categoriesPayload, productsPayload] = await Promise.all([
          apiFetch<{ items?: CatalogCategory[] }>(`/v1/organizations/${organizationId.trim()}/product-categories`, { accessToken }),
          apiFetch<{ items?: CatalogProduct[] }>(`/v1/organizations/${organizationId.trim()}/products`, { accessToken }),
        ]);
        if (cancelled) {
          return;
        }
        setCategories(Array.isArray(categoriesPayload.items) ? categoriesPayload.items : []);
        setProducts(Array.isArray(productsPayload.items) ? productsPayload.items : []);
      } catch (error) {
        if (!cancelled) {
          setStatus(`Не удалось загрузить категории/продукцию: ${errorText(error)}`);
        }
      }
    }
    void loadCatalogForOrganization();
    return () => {
      cancelled = true;
    };
  }, [accessToken, organizationId]);

  async function handleCreateOrder(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!accessToken) {
      return;
    }
    try {
      const lines = orderItemsText
        .split("\n")
        .map((item) => item.trim())
        .filter(Boolean);
      const items = [
        ...lines.map((line) => ({ productName: line })),
        ...orderNeededCategoryIds.map((categoryId) => ({ categoryId })),
      ];
      await apiFetch("/v1/sales/orders", {
        method: "POST",
        accessToken,
        bodyJSON: {
          organizationId,
          title: orderTitle,
          description: orderDescription || undefined,
          comment: orderComment || undefined,
          items,
        },
      });
      setOrderTitle("");
      setOrderDescription("");
      setOrderComment("");
      setOrderItemsText("");
      setOrderNeededCategoryIds([]);
      await loadOrders();
    } catch (error) {
      setStatus(`Не удалось создать заказ: ${errorText(error)}`);
    }
  }

  async function handleCreateOffer(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!accessToken || !selectedOrderId) {
      return;
    }
    try {
      const lines = offerItemsText
        .split("\n")
        .map((item) => item.trim())
        .filter(Boolean);
      const items = [
        ...lines.map((line) => ({ customTitle: line })),
        ...offerCategoryIds.map((categoryId) => ({ categoryId })),
        ...offerProductIds.map((productId) => ({ productId })),
      ];
      await apiFetch(`/v1/sales/orders/${selectedOrderId}/offers`, {
        method: "POST",
        accessToken,
        bodyJSON: {
          organizationId,
          comment: offerComment || undefined,
          items,
        },
      });
      setOfferComment("");
      setOfferItemsText("");
      setOfferCategoryIds([]);
      setOfferProductIds([]);
      await loadOrderDetails(selectedOrderId);
    } catch (error) {
      setStatus(`Не удалось создать предложение: ${errorText(error)}`);
    }
  }

  async function handleAddOrderComment(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!accessToken || !selectedOrderId || !newOrderComment.trim()) {
      return;
    }
    try {
      await apiFetch(`/v1/sales/orders/${selectedOrderId}/comments`, {
        method: "POST",
        accessToken,
        bodyJSON: {
          organizationId,
          comment: newOrderComment.trim(),
        },
      });
      setNewOrderComment("");
      await loadOrderDetails(selectedOrderId);
    } catch (error) {
      setStatus(`Не удалось добавить комментарий: ${errorText(error)}`);
    }
  }

  async function handleAddOfferComment(offerId: string) {
    if (!accessToken || !offerId || !organizationId.trim()) {
      return;
    }
    const text = (offerCommentDraftById[offerId] || "").trim();
    if (!text) {
      return;
    }
    try {
      await apiFetch(`/v1/sales/offers/${offerId}/comments`, {
        method: "POST",
        accessToken,
        bodyJSON: {
          organizationId: organizationId.trim(),
          comment: text,
        },
      });
      setOfferCommentDraftById((prev) => ({ ...prev, [offerId]: "" }));
      await loadOrderDetails(selectedOrderId);
    } catch (error) {
      setStatus(`Не удалось добавить комментарий к предложению: ${errorText(error)}`);
    }
  }

  function categoryLabel(categoryId?: string) {
    if (!categoryId) {
      return "Категория не указана";
    }
    const found = categories.find((item) => item.id === categoryId);
    return found ? `${found.name} (${found.code})` : categoryId;
  }

  function productLabel(productId?: string) {
    if (!productId) {
      return "Продукция не указана";
    }
    const found = products.find((item) => item.id === productId);
    return found ? found.name : productId;
  }

  return (
    <>
      <Panel title="Доска заказов и предложений" eyebrow="Sales board">
        <p className="muted">
          Используйте Organization ID текущей организации-участника. В заказе задается список потребностей (кастомный), а в предложении — строки по
          категориям/продукции или кастомные позиции.
        </p>
        <div className="form-row">
          <label className="form-label">Organization ID (ваш контекст)</label>
          <input className="text-input" value={organizationId} onChange={(event) => setOrganizationId(event.target.value)} />
        </div>
        <div className="status-card info">
          <strong>Статус</strong>
          <p className="status-copy">{status}</p>
        </div>
      </Panel>

      <section className="split">
        <Panel title="Заказы" eyebrow="GET / POST /v1/sales/orders">
          <div className="selection-list">
            {orders.map((item) => (
              <button key={item.id} type="button" className={`selection-card ${selectedOrderId === item.id ? "active" : ""}`} onClick={() => setSelectedOrderId(item.id)}>
                <strong>{item.title}</strong>
                <span className="muted">
                  <code>{item.number}</code> · {item.status}
                </span>
              </button>
            ))}
          </div>

          <form className="form-grid" onSubmit={handleCreateOrder}>
            <h3>Создать заказ</h3>
            <input className="text-input" value={orderTitle} onChange={(event) => setOrderTitle(event.target.value)} placeholder="Название заказа" required />
            <textarea className="textarea" value={orderDescription} onChange={(event) => setOrderDescription(event.target.value)} placeholder="Описание" rows={3} />
            <div className="form-row two">
              <select className="text-input" value={orderNeededCategoryId} onChange={(event) => setOrderNeededCategoryId(event.target.value)}>
                <option value="">Добавить потребность из категории...</option>
                {categories.map((item) => (
                  <option key={item.id} value={item.id}>
                    {item.name}
                  </option>
                ))}
              </select>
              <button
                className="button secondary"
                type="button"
                onClick={() => {
                  if (!orderNeededCategoryId) return;
                  setOrderNeededCategoryIds((prev) => (prev.includes(orderNeededCategoryId) ? prev : [...prev, orderNeededCategoryId]));
                  setOrderNeededCategoryId("");
                }}
              >
                Добавить категорию в заказ
              </button>
            </div>
            {orderNeededCategoryIds.length > 0 ? (
              <div className="domain-list">
                {orderNeededCategoryIds.map((item) => (
                  <div key={item} className="inline-panel">
                    <strong>{categoryLabel(item)}</strong>
                    <button className="button secondary" type="button" onClick={() => setOrderNeededCategoryIds((prev) => prev.filter((id) => id !== item))}>
                      Убрать
                    </button>
                  </div>
                ))}
              </div>
            ) : null}
            <textarea
              className="textarea"
              value={orderItemsText}
              onChange={(event) => setOrderItemsText(event.target.value)}
              placeholder="Что нужно (по одной позиции в строке)"
              rows={4}
            />
            <textarea className="textarea" value={orderComment} onChange={(event) => setOrderComment(event.target.value)} placeholder="Первый комментарий" rows={2} />
            <div className="button-row">
              <button className="button primary" type="submit">
                Создать заказ
              </button>
            </div>
          </form>
        </Panel>

        <Panel title="Детали заказа / предложения" eyebrow={selectedOrder?.title || "Выберите заказ"}>
          {selectedOrder ? (
            <>
              <div className="mini-card">
                <h3>Нужно по заказу</h3>
                <div className="domain-list">
                  {(selectedOrder.items || []).map((item) => (
                    <div key={item.id} className="inline-panel">
                      <strong>{item.productName || categoryLabel(item.categoryId) || "item"}</strong>
                      {item.note ? <p className="muted">{item.note}</p> : null}
                    </div>
                  ))}
                </div>
              </div>

              <form className="form-grid" onSubmit={handleAddOrderComment}>
                <h3>Комментарии к заказу</h3>
                <textarea className="textarea" value={newOrderComment} onChange={(event) => setNewOrderComment(event.target.value)} rows={2} />
                <div className="button-row">
                  <button className="button secondary" type="submit">
                    Добавить комментарий
                  </button>
                </div>
                <div className="domain-list">
                  {(selectedOrder.comments || []).map((item) => (
                    <div key={item.id} className="inline-panel">
                      <strong>{item.organizationName || item.organizationId}</strong>
                      <p className="muted">{item.comment}</p>
                    </div>
                  ))}
                </div>
              </form>

              <form className="form-grid" onSubmit={handleCreateOffer}>
                <h3>Создать предложение</h3>
                <div className="form-row two">
                  <select className="text-input" value={offerCategoryId} onChange={(event) => setOfferCategoryId(event.target.value)}>
                    <option value="">Добавить позицию по категории...</option>
                    {categories.map((item) => (
                      <option key={item.id} value={item.id}>
                        {item.name}
                      </option>
                    ))}
                  </select>
                  <button
                    className="button secondary"
                    type="button"
                    onClick={() => {
                      if (!offerCategoryId) return;
                      setOfferCategoryIds((prev) => (prev.includes(offerCategoryId) ? prev : [...prev, offerCategoryId]));
                      setOfferCategoryId("");
                    }}
                  >
                    Добавить категорию
                  </button>
                </div>
                <div className="form-row two">
                  <select className="text-input" value={offerProductId} onChange={(event) => setOfferProductId(event.target.value)}>
                    <option value="">Добавить позицию по продукции...</option>
                    {products.map((item) => (
                      <option key={item.id} value={item.id}>
                        {item.name}
                      </option>
                    ))}
                  </select>
                  <button
                    className="button secondary"
                    type="button"
                    onClick={() => {
                      if (!offerProductId) return;
                      setOfferProductIds((prev) => (prev.includes(offerProductId) ? prev : [...prev, offerProductId]));
                      setOfferProductId("");
                    }}
                  >
                    Добавить продукцию
                  </button>
                </div>
                {offerCategoryIds.length > 0 || offerProductIds.length > 0 ? (
                  <div className="domain-list">
                    {offerCategoryIds.map((item) => (
                      <div key={`c-${item}`} className="inline-panel">
                        <strong>{categoryLabel(item)}</strong>
                        <button className="button secondary" type="button" onClick={() => setOfferCategoryIds((prev) => prev.filter((id) => id !== item))}>
                          Убрать
                        </button>
                      </div>
                    ))}
                    {offerProductIds.map((item) => (
                      <div key={`p-${item}`} className="inline-panel">
                        <strong>{productLabel(item)}</strong>
                        <button className="button secondary" type="button" onClick={() => setOfferProductIds((prev) => prev.filter((id) => id !== item))}>
                          Убрать
                        </button>
                      </div>
                    ))}
                  </div>
                ) : null}
                <textarea
                  className="textarea"
                  value={offerItemsText}
                  onChange={(event) => setOfferItemsText(event.target.value)}
                  rows={4}
                  placeholder="Кастомные позиции предложения (по одной в строке). Для категорий/продукции используйте categoryId/productId API-поля."
                />
                <textarea className="textarea" value={offerComment} onChange={(event) => setOfferComment(event.target.value)} rows={2} placeholder="Комментарий к предложению" />
                <div className="button-row">
                  <button className="button primary" type="submit">
                    Отправить предложение
                  </button>
                </div>
              </form>

              <div className="mini-card">
                <h3>Предложения</h3>
                <div className="domain-list">
                  {offers.map((offer) => (
                    <div key={offer.id} className="inline-panel">
                      <strong>{offer.organizationId}</strong>
                      {offer.comment ? <p className="muted">{offer.comment}</p> : null}
                      <p className="muted">Позиции: {offer.items?.length || 0}</p>
                      <div className="domain-list">
                        {(offer.items || []).map((item) => (
                          <div key={item.id} className="inline-panel">
                            <strong>{item.customTitle || productLabel(item.productId) || categoryLabel(item.categoryId)}</strong>
                            {item.priceAmount ? (
                              <p className="muted">
                                Цена: {item.priceAmount} {item.currencyCode || ""}
                              </p>
                            ) : null}
                          </div>
                        ))}
                      </div>
                      <div className="domain-list">
                        {(offer.comments || []).map((comment) => (
                          <div key={comment.id} className="inline-panel">
                            <strong>{comment.organizationName || comment.organizationId}</strong>
                            <p className="muted">{comment.comment}</p>
                          </div>
                        ))}
                      </div>
                      <div className="form-row two">
                        <input
                          className="text-input"
                          value={offerCommentDraftById[offer.id] || ""}
                          onChange={(event) => setOfferCommentDraftById((prev) => ({ ...prev, [offer.id]: event.target.value }))}
                          placeholder="Комментарий к предложению"
                        />
                        <button className="button secondary" type="button" onClick={() => void handleAddOfferComment(offer.id)}>
                          Комментировать предложение
                        </button>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            </>
          ) : (
            <p className="muted">Выберите заказ слева, чтобы увидеть детали и предложения.</p>
          )}
        </Panel>
      </section>
    </>
  );
}
