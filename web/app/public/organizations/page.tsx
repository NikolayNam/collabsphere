"use client";

import { useEffect, useState } from "react";

import { Panel } from "@/components/panel";
import { APIError, apiFetch } from "@/lib/api";

type PublicOrganization = {
  id: string;
  name: string;
  slug: string;
  description?: string | null;
  website?: string | null;
  industry?: string | null;
  primaryDomain?: string | null;
  kycLevelCode: string;
  kycLevelName?: string | null;
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

export default function PublicOrganizationsPage() {
  const [items, setItems] = useState<PublicOrganization[]>([]);
  const [status, setStatus] = useState("Загружаем публичный список организаций...");

  useEffect(() => {
    let cancelled = false;
    async function load() {
      try {
        const payload = await apiFetch<{ items?: PublicOrganization[] }>("/v1/organizations/public/kyc-directory?limit=200");
        if (cancelled) {
          return;
        }
        const list = Array.isArray(payload.items) ? payload.items : [];
        setItems(list);
        setStatus(list.length > 0 ? `Найдено организаций: ${list.length}` : "Список пока пуст");
      } catch (error) {
        if (cancelled) {
          return;
        }
        setItems([]);
        setStatus(`Не удалось загрузить список: ${errorText(error)}`);
      }
    }
    void load();
    return () => {
      cancelled = true;
    };
  }, []);

  return (
    <Panel title="Публичный реестр организаций" eyebrow="KYC directory">
      <p className="muted">
        В реестр попадают организации с уровнем <code>public_directory_org_verified</code>, подтвержденным учредительным документом,
        заполненным юридическим наименованием и подтвержденным основным доменом.
      </p>
      <div className="status-card info">
        <strong>Статус</strong>
        <p className="status-copy">{status}</p>
      </div>
      {items.length > 0 ? (
        <div className="domain-list">
          {items.map((item) => (
            <div key={item.id} className="inline-panel">
              <strong>{item.name}</strong>
              <p className="muted">
                <code>{item.slug}</code> · {item.kycLevelName || item.kycLevelCode}
              </p>
              {item.primaryDomain ? <p className="muted">Домен: {item.primaryDomain}</p> : null}
              {item.website ? (
                <p className="muted">
                  Сайт:{" "}
                  <a href={item.website} target="_blank" rel="noreferrer">
                    {item.website}
                  </a>
                </p>
              ) : null}
              {item.industry ? <p className="muted">Отрасль: {item.industry}</p> : null}
              {item.description ? <p className="muted">{item.description}</p> : null}
            </div>
          ))}
        </div>
      ) : null}
    </Panel>
  );
}
