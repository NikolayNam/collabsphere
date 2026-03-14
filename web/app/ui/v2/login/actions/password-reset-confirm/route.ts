import { NextResponse } from "next/server";

import { createLoginURL } from "@/lib/login/env";
import { changePasswordWithCode, ZitadelAPIError } from "@/lib/login/zitadel";

export async function POST(request: Request): Promise<Response> {
  const form = await request.formData();
  const userId = String(form.get("userId") || "").trim();
  const verificationCode = String(form.get("verificationCode") || "").trim();
  const password = String(form.get("password") || "");
  const authRequest = String(form.get("authRequest") || "").trim();

  const failureTarget = createLoginURL("/ui/v2/login/password-reset");
  failureTarget.searchParams.set("userId", userId);
  if (verificationCode) {
    failureTarget.searchParams.set("code", verificationCode);
  }
  if (authRequest) {
    failureTarget.searchParams.set("authRequest", authRequest);
  }

  if (!userId || !verificationCode || !password) {
    failureTarget.searchParams.set("error", "userId, verificationCode and password are required");
    return NextResponse.redirect(failureTarget, { status: 303 });
  }

  try {
    await changePasswordWithCode(userId, verificationCode, password);
    const target = createLoginURL("/ui/v2/login/login");
    target.searchParams.set("message", "Password changed. Continue with login.");
    if (authRequest) {
      target.searchParams.set("authRequest", authRequest);
    }
    return NextResponse.redirect(target, { status: 303 });
  } catch (error) {
    failureTarget.searchParams.set(
      "error",
      error instanceof ZitadelAPIError ? error.message : error instanceof Error ? error.message : "Password change failed",
    );
    return NextResponse.redirect(failureTarget, { status: 303 });
  }
}
