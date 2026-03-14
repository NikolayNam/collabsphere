const fallbackLoginBaseURL = "http://auth.localhost:3000";
const fallbackLoginBasePath = "/ui/v2/login";
const fallbackZitadelPublicAPIURL = "http://auth.localhost:8090";
const fallbackZitadelInternalAPIURL = "http://auth.localhost:8080";
const fallbackSessionCookieName = "collabsphere.login.session";
const fallbackSessionSecret = "collabsphere-dev-login-session-secret";

function trimTrailingSlash(value: string): string {
  return value.replace(/\/+$/, "");
}

export function normalizeHost(value: string | null | undefined): string {
  return (value || "").trim().toLowerCase();
}

export function getLoginBaseURL(): string {
  return trimTrailingSlash(process.env.WEB_LOGIN_BASE_URL || fallbackLoginBaseURL);
}

export function createLoginURL(pathname: string): URL {
  const url = new URL(getLoginBaseURL());
  url.pathname = pathname.startsWith("/") ? pathname : `/${pathname}`;
  url.search = "";
  return url;
}

export function getLoginBasePath(): string {
  const value = (process.env.WEB_LOGIN_BASE_PATH || fallbackLoginBasePath).trim();
  return value.startsWith("/") ? value.replace(/\/+$/, "") || "/" : `/${value.replace(/\/+$/, "")}`;
}

export function getLoginAllowedHost(): string {
  if (process.env.WEB_LOGIN_ALLOWED_HOST?.trim()) {
    return normalizeHost(process.env.WEB_LOGIN_ALLOWED_HOST);
  }
  return normalizeHost(new URL(getLoginBaseURL()).host);
}

export function isLoginHostValue(value: string | null | undefined): boolean {
  const normalized = normalizeHost(value);
  const allowedHost = getLoginAllowedHost();
  if (normalized === allowedHost) {
    return true;
  }

  const loginHostConfigured = Boolean(
    process.env.WEB_LOGIN_ALLOWED_HOST?.trim() || process.env.WEB_LOGIN_BASE_URL?.trim(),
  );
  if (!loginHostConfigured) {
    return false;
  }

  const allowed = new URL(getLoginBaseURL());
  const allowedPort = allowed.port || (allowed.protocol === "https:" ? "443" : "80");
  return normalized === `localhost:${allowedPort}` || normalized === `127.0.0.1:${allowedPort}`;
}

export function getZitadelPublicAPIURL(): string {
  return trimTrailingSlash(process.env.WEB_LOGIN_ZITADEL_PUBLIC_API_URL || fallbackZitadelPublicAPIURL);
}

export function getZitadelInternalAPIURL(): string {
  return trimTrailingSlash(
    process.env.WEB_LOGIN_ZITADEL_INTERNAL_API_URL ||
      process.env.WEB_LOGIN_ZITADEL_API_URL ||
      fallbackZitadelInternalAPIURL,
  );
}

export function getZitadelRuntimeAPIURL(): string {
  return trimTrailingSlash(process.env.WEB_LOGIN_ZITADEL_RUNTIME_API_URL || getZitadelInternalAPIURL());
}

export function getZitadelInstanceHost(): string {
  return normalizeHost(new URL(getZitadelPublicAPIURL()).host);
}

export function getLoginPublicHost(): string {
  return normalizeHost(new URL(getLoginBaseURL()).host);
}

export function getSessionCookieName(): string {
  return process.env.WEB_LOGIN_SESSION_COOKIE_NAME?.trim() || fallbackSessionCookieName;
}

export function getSessionSecret(): string {
  return process.env.WEB_LOGIN_SESSION_SECRET?.trim() || fallbackSessionSecret;
}
