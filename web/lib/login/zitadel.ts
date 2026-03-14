import {
  getLoginBaseURL,
  getLoginPublicHost,
  getZitadelInstanceHost,
  getZitadelPublicAPIURL,
  getZitadelRuntimeAPIURL,
} from "@/lib/login/env";
import type { LoginSessionState } from "@/lib/login/session";
import { getServiceUserToken } from "@/lib/login/token";

type RequestOptions = {
  method?: string;
  bodyJSON?: unknown;
  headers?: HeadersInit;
};

export class ZitadelAPIError extends Error {
  status: number;
  detail?: string;

  constructor(message: string, status: number, detail?: string) {
    super(message);
    this.name = "ZitadelAPIError";
    this.status = status;
    this.detail = detail;
  }
}

export type AuthRequestState = {
  id: string;
  loginHint?: string;
  prompt?: string;
  raw: Record<string, unknown>;
};

export type SignupResult = {
  userId: string;
  verificationCode?: string;
};

export type PasswordResetResult = {
  verificationCode?: string;
};

export type SessionState = LoginSessionState;

export async function getAuthRequest(authRequestId: string): Promise<AuthRequestState> {
  const payload = await zitadelFetch<Record<string, unknown>>(`/v2/oidc/auth_requests/${encodeURIComponent(authRequestId)}`);
  return {
    id: authRequestId,
    loginHint: readString(payload, ["loginHint", "login_name", "loginName"]),
    prompt: readString(payload, ["prompt"]),
    raw: payload,
  };
}

export async function createSessionForLoginName(loginName: string): Promise<SessionState> {
  const resolvedLoginName = await resolvePreferredLoginName(loginName);
  const payload = await zitadelFetch<Record<string, unknown>>("/v2/sessions", {
    method: "POST",
    bodyJSON: {
      checks: {
        user: {
          loginName: resolvedLoginName,
        },
      },
    },
  });
  const session = toSessionState(payload);
  return {
    ...session,
    loginName: session.loginName || resolvedLoginName,
  };
}

export async function createSessionForUserID(userId: string): Promise<SessionState> {
  const payload = await zitadelFetch<Record<string, unknown>>("/v2/sessions", {
    method: "POST",
    bodyJSON: {
      checks: {
        user: {
          userId,
        },
      },
    },
  });
  return toSessionState(payload);
}

export async function applyPasswordCheck(session: SessionState, password: string): Promise<SessionState> {
  const payload = await zitadelFetch<Record<string, unknown>>(`/v2/sessions/${encodeURIComponent(session.sessionId)}`, {
    method: "PATCH",
    bodyJSON: {
      checks: {
        password: {
          password,
        },
      },
    },
  });
  const next = toSessionState(payload);
  return {
    ...next,
    sessionId: next.sessionId || session.sessionId,
    sessionToken: next.sessionToken || session.sessionToken,
    userId: next.userId || session.userId,
    loginName: next.loginName || session.loginName,
    displayName: next.displayName || session.displayName,
  };
}

export async function createCallback(authRequestId: string, session: SessionState): Promise<string> {
  const payload = await zitadelFetch<Record<string, unknown>>(`/v2/oidc/auth_requests/${encodeURIComponent(authRequestId)}`, {
    method: "POST",
    bodyJSON: {
      session: {
        sessionId: session.sessionId,
        sessionToken: session.sessionToken,
      },
    },
  });
  const callbackURL = readString(payload, ["callbackUrl"]);
  if (!callbackURL) {
    throw new Error("ZITADEL CreateCallback did not return callbackUrl");
  }
  return callbackURL;
}

export async function terminateSession(session: SessionState): Promise<void> {
  await zitadelFetch(`/v2/sessions/${encodeURIComponent(session.sessionId)}`, {
    method: "DELETE",
    bodyJSON: {
      sessionToken: session.sessionToken,
    },
  });
}

export async function createHumanUser(input: {
  displayName: string;
  email: string;
  password: string;
}): Promise<SignupResult> {
  const nestedPayload = {
    username: input.email,
    human: {
      profile: {
        givenName: firstWord(input.displayName),
        familyName: remainderWords(input.displayName) || "User",
        displayName: input.displayName,
      },
      email: {
        email: input.email,
        returnCode: {},
      },
      password: {
        password: input.password,
        changeRequired: false,
      },
    },
  };
  const flatPayload = {
    username: input.email,
    profile: {
      givenName: firstWord(input.displayName),
      familyName: remainderWords(input.displayName) || "User",
      displayName: input.displayName,
    },
    email: {
      email: input.email,
      returnCode: {},
    },
    password: {
      password: input.password,
      changeRequired: false,
    },
  };

  let payload: Record<string, unknown>;
  try {
    payload = await zitadelFetch<Record<string, unknown>>("/v2/users/new", {
      method: "POST",
      bodyJSON: nestedPayload,
    });
  } catch (error) {
    // Some ZITADEL setups accept only /v2/users/human with flat payload.
    // Fall back to keep signup flow portable across local versions/configs.
    if (!(error instanceof ZitadelAPIError) || error.status < 400 || error.status >= 500) {
      throw error;
    }
    payload = await zitadelFetch<Record<string, unknown>>("/v2/users/human", {
      method: "POST",
      bodyJSON: flatPayload,
    });
  }
  const userId = readString(payload, ["userId", "id"]);
  if (!userId) {
    throw new Error("ZITADEL CreateUser did not return userId");
  }
  return {
    userId,
    verificationCode: readString(payload, ["verificationCode", "emailCode"]),
  };
}

export async function verifyEmail(userId: string, verificationCode: string): Promise<void> {
  await zitadelFetch(`/v2/users/${encodeURIComponent(userId)}/email/verify`, {
    method: "POST",
    bodyJSON: {
      verificationCode,
    },
  });
}

export async function resendEmailCode(userId: string): Promise<string | undefined> {
  const payload = await zitadelFetch<Record<string, unknown>>(`/v2/users/${encodeURIComponent(userId)}/email/resend`, {
    method: "POST",
  });
  return readString(payload, ["verificationCode"]);
}

export async function requestPasswordReset(userId: string): Promise<PasswordResetResult> {
  const payload = await zitadelFetch<Record<string, unknown>>(`/v2/users/${encodeURIComponent(userId)}/password_reset`, {
    method: "POST",
  });
  return {
    verificationCode: readString(payload, ["verificationCode"]),
  };
}

export async function changePasswordWithCode(userId: string, verificationCode: string, password: string): Promise<void> {
  await zitadelFetch(`/v2/users/${encodeURIComponent(userId)}/password`, {
    method: "POST",
    bodyJSON: {
      newPassword: {
        password,
        changeRequired: false,
      },
      verificationCode,
    },
  });
}

export async function proxyZitadelRequest(request: Request, pathname: string, search: string): Promise<Response> {
  const targetURL = `${getZitadelRuntimeAPIURL()}${pathname}${search}`;
  const headers = new Headers(request.headers);
  headers.set("x-zitadel-public-host", getLoginPublicHost());
  headers.set("x-zitadel-instance-host", getZitadelInstanceHost());
  headers.set("x-forwarded-host", getLoginPublicHost());
  headers.set("x-forwarded-proto", "http");
  headers.delete("host");
  headers.delete("content-length");

  const init: RequestInit & { duplex?: "half" } = {
    method: request.method,
    headers,
    redirect: "manual",
  };
  if (request.method !== "GET" && request.method !== "HEAD") {
    init.body = await request.arrayBuffer();
    init.duplex = "half";
  }

  const upstream = await fetch(targetURL, init);
  const responseHeaders = new Headers(upstream.headers);
  if ("getSetCookie" in upstream.headers && typeof upstream.headers.getSetCookie === "function") {
    responseHeaders.delete("set-cookie");
    for (const cookie of upstream.headers.getSetCookie()) {
      responseHeaders.append("set-cookie", cookie);
    }
  }
  const location = responseHeaders.get("location");
  if (location?.startsWith(getZitadelRuntimeAPIURL())) {
    responseHeaders.set("location", location.replace(getZitadelRuntimeAPIURL(), getLoginBaseOrigin()));
  }
  if (location?.startsWith(getZitadelPublicAPIURL())) {
    responseHeaders.set("location", location.replace(getZitadelPublicAPIURL(), getLoginBaseOrigin()));
  }
  return new Response(upstream.body, {
    status: upstream.status,
    statusText: upstream.statusText,
    headers: responseHeaders,
  });
}

function getLoginBaseOrigin(): string {
  return getLoginBaseURL();
}

async function zitadelFetch<T = unknown>(path: string, options: RequestOptions = {}): Promise<T> {
  const headers = new Headers(options.headers || {});
  headers.set("Accept", "application/json");
  headers.set("Authorization", `Bearer ${await getServiceUserToken()}`);
  headers.set("x-zitadel-public-host", getLoginPublicHost());
  headers.set("x-zitadel-instance-host", getZitadelInstanceHost());
  headers.set("x-forwarded-host", getLoginPublicHost());
  headers.set("x-forwarded-proto", "http");
  if (options.bodyJSON !== undefined) {
    headers.set("Content-Type", "application/json");
  }

  const response = await fetch(`${getZitadelRuntimeAPIURL()}${path}`, {
    method: options.method || "GET",
    headers,
    body: options.bodyJSON !== undefined ? JSON.stringify(options.bodyJSON) : undefined,
    cache: "no-store",
  });

  const text = await response.text();
  const payload = text ? safeParseJSON(text) : null;
  if (!response.ok) {
    throw new ZitadelAPIError(readString(payload, ["message", "detail"]) || `HTTP ${response.status}`, response.status, text || undefined);
  }
  return payload as T;
}

function toSessionState(payload: Record<string, unknown>): SessionState {
  const sessionId = readString(payload, ["sessionId"]);
  const sessionToken = readString(payload, ["sessionToken"]);
  if (!sessionId && !sessionToken) {
    throw new Error("ZITADEL session response did not return sessionId/sessionToken");
  }
  const userId =
    readString(payload, ["session.factors.user.id"]) ||
    readString(payload, ["factors.user.id"]) ||
    readString(payload, ["userId"]);
  const loginName =
    readString(payload, ["session.factors.user.loginName"]) ||
    readString(payload, ["factors.user.loginName"]) ||
    readString(payload, ["loginName"]);
  const displayName =
    readString(payload, ["session.factors.user.displayName"]) ||
    readString(payload, ["factors.user.displayName"]) ||
    readString(payload, ["displayName"]);

  return {
    sessionId: sessionId || "",
    sessionToken: sessionToken || "",
    userId: userId || undefined,
    loginName: loginName || undefined,
    displayName: displayName || undefined,
  };
}

async function resolvePreferredLoginName(input: string): Promise<string> {
  const normalized = input.trim();
  if (!normalized || !normalized.includes("@")) {
    return normalized;
  }

  const payload = await zitadelFetch<Record<string, unknown>>("/v2/users", {
    method: "POST",
    bodyJSON: {
      queries: [
        {
          emailQuery: {
            emailAddress: normalized,
            method: "TEXT_QUERY_METHOD_EQUALS",
          },
        },
      ],
    },
  });

  const result = readFirstObject(payload, ["result"]);
  return (
    readString(result, ["preferredLoginName", "username"]) ||
    readString(result, ["loginNames.0"]) ||
    normalized
  );
}

function safeParseJSON(value: string): unknown {
  try {
    return JSON.parse(value);
  } catch {
    return value;
  }
}

function readString(value: unknown, paths: string[]): string | undefined {
  for (const path of paths) {
    const found = readPath(value, path);
    if (typeof found === "string" && found.trim()) {
      return found.trim();
    }
  }
  return undefined;
}

function readFirstObject(value: unknown, paths: string[]): Record<string, unknown> | undefined {
  for (const path of paths) {
    const found = readPath(value, path);
    if (Array.isArray(found) && found.length > 0) {
      const first = found[0];
      if (first && typeof first === "object") {
        return first as Record<string, unknown>;
      }
    }
  }
  return undefined;
}

function readPath(value: unknown, path: string): unknown {
  const segments = path.split(".");
  let current: unknown = value;
  for (const segment of segments) {
    if (Array.isArray(current)) {
      const index = Number(segment);
      if (!Number.isInteger(index) || index < 0 || index >= current.length) {
        return undefined;
      }
      current = current[index];
      continue;
    }
    if (!current || typeof current !== "object") {
      return undefined;
    }
    current = (current as Record<string, unknown>)[segment];
  }
  return current;
}

function firstWord(value: string): string {
  return value.trim().split(/\s+/, 1)[0] || "User";
}

function remainderWords(value: string): string {
  const parts = value.trim().split(/\s+/);
  return parts.slice(1).join(" ");
}
