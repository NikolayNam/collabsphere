import { apiFetch } from "@/lib/api";
import { getBrowserLoginURL } from "@/lib/env";

export type TokenBundle = {
  accessToken: string;
  refreshToken: string;
  tokenType: string;
  expiresIn: number;
  provider?: string;
  intent?: string;
  isNewAccount?: boolean;
};

export type MeResponse = {
  id: string;
  email: string;
  displayName?: string | null;
  avatarObjectId?: string | null;
  bio?: string | null;
  phone?: string | null;
  locale?: string | null;
  timezone?: string | null;
  website?: string | null;
  isActive: boolean;
  createdAt: string;
  updatedAt?: string | null;
};

const STORAGE_KEY = "collabsphere.web.auth";
const ACCESS_KEY = "collabsphere.web.accessToken";
const REFRESH_KEY = "collabsphere.web.refreshToken";

export function beginBrowserLogin(intent: "login" | "signup" = "login"): void {
  window.location.assign(getBrowserLoginURL(intent));
}

export function readStoredTokens(): TokenBundle | null {
  if (typeof window === "undefined") {
    return null;
  }

  const raw = window.localStorage.getItem(STORAGE_KEY);
  if (raw) {
    try {
      const parsed = JSON.parse(raw) as Partial<TokenBundle>;
      if (parsed.accessToken && parsed.refreshToken) {
        return {
          accessToken: parsed.accessToken,
          refreshToken: parsed.refreshToken,
          tokenType: parsed.tokenType || "Bearer",
          expiresIn: Number(parsed.expiresIn || 0),
          provider: parsed.provider,
          intent: parsed.intent,
          isNewAccount: parsed.isNewAccount,
        };
      }
    } catch {
      // ignore broken local storage payloads
    }
  }

  const accessToken = window.localStorage.getItem(ACCESS_KEY);
  const refreshToken = window.localStorage.getItem(REFRESH_KEY);
  if (!accessToken || !refreshToken) {
    return null;
  }
  return {
    accessToken,
    refreshToken,
    tokenType: "Bearer",
    expiresIn: 0,
  };
}

export function storeTokens(tokens: TokenBundle): void {
  if (typeof window === "undefined") {
    return;
  }
  window.localStorage.setItem(STORAGE_KEY, JSON.stringify(tokens));
  window.localStorage.setItem(ACCESS_KEY, tokens.accessToken);
  window.localStorage.setItem(REFRESH_KEY, tokens.refreshToken);
}

export function clearTokens(): void {
  if (typeof window === "undefined") {
    return;
  }
  window.localStorage.removeItem(STORAGE_KEY);
  window.localStorage.removeItem(ACCESS_KEY);
  window.localStorage.removeItem(REFRESH_KEY);
}

export async function exchangeBrowserTicket(ticket: string): Promise<TokenBundle> {
  return apiFetch<TokenBundle>("/v1/auth/exchange", {
    method: "POST",
    bodyJSON: { ticket },
  });
}

export async function loginWithPassword(email: string, password: string): Promise<TokenBundle> {
  return apiFetch<TokenBundle>("/v1/auth/login", {
    method: "POST",
    bodyJSON: { email, password },
  });
}

export async function loadMe(accessToken: string): Promise<MeResponse> {
  return apiFetch<MeResponse>("/v1/auth/me", {
    accessToken,
  });
}

export async function rotateTokens(refreshToken: string): Promise<TokenBundle> {
  return apiFetch<TokenBundle>("/v1/auth/refresh", {
    method: "POST",
    bodyJSON: { refreshToken },
  });
}

export async function logout(refreshToken: string): Promise<void> {
  await apiFetch<void>("/v1/auth/logout", {
    method: "POST",
    bodyJSON: { refreshToken },
  });
}
