import { getAPIProxyPrefix } from "@/lib/env";

type RequestOptions = Omit<RequestInit, "headers"> & {
  accessToken?: string | null;
  bodyJSON?: unknown;
  headers?: HeadersInit;
};

export class APIError extends Error {
  status: number;
  code?: string;
  detail?: string;

  constructor(message: string, status: number, code?: string, detail?: string) {
    super(message);
    this.name = "APIError";
    this.status = status;
    this.code = code;
    this.detail = detail;
  }
}

export async function apiFetch<T>(path: string, options: RequestOptions = {}): Promise<T> {
  const headers = new Headers(options.headers || {});

  if (options.bodyJSON !== undefined) {
    headers.set("Content-Type", "application/json");
  }
  if (options.accessToken) {
    headers.set("Authorization", `Bearer ${options.accessToken}`);
  }

  const url = path.startsWith("/api/") ? path : `${getAPIProxyPrefix()}${path}`;
  const response = await fetch(url, {
    ...options,
    headers,
    body: options.bodyJSON !== undefined ? JSON.stringify(options.bodyJSON) : options.body,
    cache: "no-store",
  });

  if (response.status === 204) {
    return undefined as T;
  }

  const text = await response.text();
  const payload = text ? safeParseJSON(text) : null;

  if (!response.ok) {
    const detail = getProblemField(payload, ["detail", "title", "error"]);
    const code = getProblemField(payload, ["code"]);
    throw new APIError(detail || `HTTP ${response.status}`, response.status, code, detail);
  }

  return payload as T;
}

function safeParseJSON(value: string): unknown {
  try {
    return JSON.parse(value);
  } catch {
    return value;
  }
}

function getProblemField(value: unknown, keys: string[]): string | undefined {
  if (!value || typeof value !== "object") {
    return undefined;
  }
  const record = value as Record<string, unknown>;
  for (const key of keys) {
    const candidate = record[key];
    if (typeof candidate === "string" && candidate.trim() !== "") {
      return candidate.trim();
    }
  }
  return undefined;
}
