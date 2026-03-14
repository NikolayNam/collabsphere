import { NextResponse } from "next/server";

import { getLoginBaseURL } from "@/lib/login/env";
import { clearLoginSession } from "@/lib/login/session";

export async function GET(request: Request): Promise<Response> {
  const url = new URL(request.url);
  const postLogoutRedirect = url.searchParams.get("post_logout_redirect") || "/ui/v2/login/login";
  await clearLoginSession();
  return NextResponse.redirect(new URL(postLogoutRedirect, getLoginBaseURL()));
}
