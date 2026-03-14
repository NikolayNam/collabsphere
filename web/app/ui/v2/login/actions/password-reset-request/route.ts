import { NextResponse } from "next/server";

import { createLoginURL } from "@/lib/login/env";
import { createSessionForLoginName, requestPasswordReset, terminateSession, ZitadelAPIError } from "@/lib/login/zitadel";

export async function POST(request: Request): Promise<Response> {
  const form = await request.formData();
  const loginHint = String(form.get("loginHint") || "").trim();
  const authRequest = String(form.get("authRequest") || "").trim();

  const target = createLoginURL("/ui/v2/login/password-reset");
  target.searchParams.set("loginHint", loginHint);
  if (authRequest) {
    target.searchParams.set("authRequest", authRequest);
  }

  if (!loginHint) {
    target.searchParams.set("error", "loginHint is required");
    return NextResponse.redirect(target, { status: 303 });
  }

  try {
    const session = await createSessionForLoginName(loginHint);
    if (!session.userId) {
      throw new Error("ZITADEL session did not expose userId for password reset");
    }
    const requested = await requestPasswordReset(session.userId);
    await terminateSession(session).catch(() => undefined);
    target.searchParams.set("userId", session.userId);
    target.searchParams.set("message", "Verification code requested.");
    if (requested.verificationCode) {
      target.searchParams.set("code", requested.verificationCode);
    }
    return NextResponse.redirect(target, { status: 303 });
  } catch (error) {
    target.searchParams.set(
      "error",
      error instanceof ZitadelAPIError ? error.message : error instanceof Error ? error.message : "Password reset failed",
    );
    return NextResponse.redirect(target, { status: 303 });
  }
}
