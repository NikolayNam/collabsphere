"use client";

import { Suspense, useEffect, useState } from "react";
import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";

import { Panel } from "@/components/panel";
import { APIError } from "@/lib/api";
import { exchangeBrowserTicket, storeTokens, type TokenBundle } from "@/lib/auth";

type CallbackState =
  | { kind: "working"; title: string; description: string }
  | { kind: "success"; title: string; description: string; tokens: TokenBundle }
  | { kind: "error"; title: string; description: string };

const initialState: CallbackState = {
  kind: "working",
  title: "Ожидаем browser callback",
  description: "Если backend уже вернул `ticket`, frontend сейчас обменяет его на локальные токены.",
};

function CallbackFallback() {
  return (
    <Panel title="Auth callback" eyebrow="Ticket exchange">
      <div className="status-card info">
        <strong>{initialState.title}</strong>
        <p className="status-copy">{initialState.description}</p>
      </div>
    </Panel>
  );
}

function AuthCallbackContent() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const [state, setState] = useState<CallbackState>(initialState);

  useEffect(() => {
    const providerError = searchParams.get("error");
    const providerDescription = searchParams.get("error_description");

    if (providerError) {
      setState({
        kind: "error",
        title: "Внешний login завершился ошибкой",
        description: providerDescription || providerError,
      });
      return;
    }

    const ticket = searchParams.get("ticket");
    if (!ticket) {
      setState({
        kind: "error",
        title: "Ticket не найден",
        description: "Откройте /login и запустите browser flow заново.",
      });
      return;
    }

    let cancelled = false;
    void (async () => {
      setState({
        kind: "working",
        title: "Завершаем вход",
        description: "Frontend вызывает /v1/auth/exchange через Next.js proxy и ждёт локальные access/refresh токены.",
      });

      try {
        const tokens = await exchangeBrowserTicket(ticket);
        if (cancelled) {
          return;
        }
        storeTokens(tokens);
        setState({
          kind: "success",
          title: "Browser login завершён",
          description: "Токены сохранены. Через пару секунд frontend вернёт пользователя на домашнюю страницу приложения.",
          tokens,
        });
        setTimeout(() => {
          router.replace("/");
        }, 1400);
      } catch (error) {
        if (cancelled) {
          return;
        }
        const message =
          error instanceof APIError
            ? `${error.message}${error.code ? ` (${error.code})` : ""}`
            : error instanceof Error
              ? error.message
              : "Unknown exchange error";
        setState({
          kind: "error",
          title: "Exchange не удался",
          description: message,
        });
      }
    })();

    return () => {
      cancelled = true;
    };
  }, [router, searchParams]);

  return (
    <>
      <Panel title="Auth callback" eyebrow="Ticket exchange">
        <div className={`status-card ${state.kind === "error" ? "error" : state.kind === "success" ? "success" : "info"}`}>
          <strong>{state.title}</strong>
          <p className="status-copy">{state.description}</p>
        </div>
      </Panel>

      {state.kind === "success" ? (
        <section className="split">
          <Panel title="Access token" eyebrow="Dev-visible token">
            <textarea className="code-block" readOnly value={state.tokens.accessToken} />
          </Panel>
          <Panel title="Refresh token" eyebrow="Dev-visible token">
            <textarea className="code-block" readOnly value={state.tokens.refreshToken} />
          </Panel>
        </section>
      ) : null}

      <Panel title="Next steps" eyebrow="Where to go next">
        <div className="button-row">
          <Link href="/login" className="button-link secondary">
            Назад к login
          </Link>
          <Link href="/me" className="button-link primary">
            Открыть /me
          </Link>
          <Link href="/" className="button-link secondary">
            Открыть главную
          </Link>
          <Link href="/organizations" className="button-link secondary">
            Открыть /organizations
          </Link>
        </div>
      </Panel>
    </>
  );
}

export default function AuthCallbackPage() {
  return (
    <Suspense fallback={<CallbackFallback />}>
      <AuthCallbackContent />
    </Suspense>
  );
}
