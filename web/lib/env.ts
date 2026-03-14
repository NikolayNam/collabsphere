const fallbackAPIBaseURL = "http://api.localhost:8080";
const fallbackAppBaseURL = "http://collabsphere.localhost:3002";

function trimTrailingSlash(value: string): string {
  return value.replace(/\/+$/, "");
}

export function getAPIBaseURL(): string {
  return trimTrailingSlash(process.env.NEXT_PUBLIC_API_BASE_URL || fallbackAPIBaseURL);
}

export function getAppBaseURL(): string {
  return trimTrailingSlash(process.env.NEXT_PUBLIC_APP_BASE_URL || fallbackAppBaseURL);
}

export function getBrowserCallbackURL(): string {
  return `${getAppBaseURL()}/auth/callback`;
}

export function getBrowserLoginURL(intent: "login" | "signup" = "login"): string {
  const route = intent === "signup" ? "/v1/auth/zitadel/signup" : "/v1/auth/zitadel/login";
  const url = new URL(`${getAPIBaseURL()}${route}`);
  url.searchParams.set("return_to", getBrowserCallbackURL());
  return url.toString();
}

export function getAPIProxyPrefix(): string {
  return "/api/backend";
}
