export const LOGIN_UI_PREFIX = "/ui/v2/login";
export const LOGIN_PROXY_PREFIXES = ["/.well-known", "/oauth", "/oidc", "/idps"];

export function isLoginUIPath(pathname: string): boolean {
  return pathname === LOGIN_UI_PREFIX || pathname.startsWith(`${LOGIN_UI_PREFIX}/`);
}

export function isLoginProxyPath(pathname: string): boolean {
  return LOGIN_PROXY_PREFIXES.some((prefix) => pathname === prefix || pathname.startsWith(`${prefix}/`));
}

export function isStaticAssetPath(pathname: string): boolean {
  return pathname.startsWith("/_next/") || pathname === "/favicon.ico";
}

export function isLoginSurfacePath(pathname: string): boolean {
  return isLoginUIPath(pathname) || isLoginProxyPath(pathname) || isStaticAssetPath(pathname);
}

export function appendSearch(pathname: string, search: string): string {
  return search ? `${pathname}${search}` : pathname;
}
