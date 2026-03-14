import { NextResponse } from "next/server";

import { createLoginURL } from "@/lib/login/env";
import { resendEmailCode, ZitadelAPIError } from "@/lib/login/zitadel";

export async function POST(request: Request): Promise<Response> {
  const form = await request.formData();
  const userId = String(form.get("userId") || "").trim();
  const loginHint = String(form.get("loginHint") || "").trim();
  const authRequest = String(form.get("authRequest") || "").trim();

  const target = createLoginURL("/ui/v2/login/verify");
  target.searchParams.set("userId", userId);
  if (loginHint) {
    target.searchParams.set("loginHint", loginHint);
  }
  if (authRequest) {
    target.searchParams.set("authRequest", authRequest);
  }

  if (!userId) {
    target.searchParams.set("error", "userId is required");
    return NextResponse.redirect(target, { status: 303 });
  }

  try {
    const code = await resendEmailCode(userId);
    target.searchParams.set("message", "A new verification code was requested.");
    if (code) {
      target.searchParams.set("code", code);
    }
    return NextResponse.redirect(target, { status: 303 });
  } catch (error) {
    target.searchParams.set(
      "error",
      error instanceof ZitadelAPIError ? error.message : error instanceof Error ? error.message : "Resend email failed",
    );
    return NextResponse.redirect(target, { status: 303 });
  }
}
