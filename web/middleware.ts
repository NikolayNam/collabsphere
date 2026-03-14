import type { NextRequest } from "next/server";
import { NextResponse } from "next/server";

import { getLoginBaseURL, isLoginHostValue, normalizeHost } from "@/lib/login/env";
import { isLoginProxyPath, isLoginUIPath, isStaticAssetPath } from "@/lib/login/surface";

export function middleware(request: NextRequest) {
  const host = normalizeHost(request.headers.get("host"));
  const pathname = request.nextUrl.pathname;
  const search = request.nextUrl.search;
  const nextHeaders = new Headers(request.headers);
  nextHeaders.set("x-collabsphere-request-host", host);

  if (isStaticAssetPath(pathname)) {
    return NextResponse.next({
      request: {
        headers: nextHeaders,
      },
    });
  }

  if (isLoginHostValue(host)) {
    if (!isLoginUIPath(pathname) && !isLoginProxyPath(pathname)) {
      const target = new URL(getLoginBaseURL());
      target.pathname = "/ui/v2/login/login";
      return NextResponse.redirect(target);
    }
    return NextResponse.next({
      request: {
        headers: nextHeaders,
      },
    });
  }

  if (isLoginUIPath(pathname) || isLoginProxyPath(pathname)) {
    const target = new URL(getLoginBaseURL());
    target.pathname = pathname;
    target.search = search;
    return NextResponse.redirect(target);
  }

  return NextResponse.next({
    request: {
      headers: nextHeaders,
    },
  });
}

export const config = {
  matcher: ["/((?!_next/image|robots.txt).*)"],
};
