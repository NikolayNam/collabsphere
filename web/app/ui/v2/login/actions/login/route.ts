import { NextResponse } from "next/server";

import { createLoginURL } from "@/lib/login/env";
import { clearLoginSession, writeLoginSession } from "@/lib/login/session";
import { applyPasswordCheck, createCallback, createSessionForLoginName, ZitadelAPIError } from "@/lib/login/zitadel";

function redirectToLogin(message: string, authRequest: string, loginHint: string): NextResponse {
  const target = createLoginURL("/ui/v2/login/login");
  target.searchParams.set("error", message);
  if (authRequest) {
    target.searchParams.set("authRequest", authRequest);
  }
  if (loginHint) {
    target.searchParams.set("loginHint", loginHint);
  }
  return NextResponse.redirect(target, { status: 303 });
}

export async function POST(request: Request): Promise<Response> {
  const form = await request.formData();
  const loginName = String(form.get("loginName") || "").trim();
  const password = String(form.get("password") || "");
  const authRequest = String(form.get("authRequest") || "").trim();

  if (!loginName || !password) {
    return redirectToLogin("loginName and password are required", authRequest, loginName);
  }

  try {
    const baseSession = await createSessionForLoginName(loginName);
    const passwordSession = await applyPasswordCheck(baseSession, password);
    await writeLoginSession(passwordSession);
    if (authRequest) {
      const callbackURL = await createCallback(authRequest, passwordSession);
      await clearLoginSession();
      return NextResponse.redirect(callbackURL, { status: 303 });
    }
    const target = createLoginURL("/ui/v2/login/login");
    target.searchParams.set("message", "Login succeeded");
    return NextResponse.redirect(target, { status: 303 });
  } catch (error) {
    const message =
      error instanceof ZitadelAPIError
        ? error.message
        : error instanceof Error
          ? error.message
          : "Login failed";
    return redirectToLogin(message, authRequest, loginName);
  }
}
