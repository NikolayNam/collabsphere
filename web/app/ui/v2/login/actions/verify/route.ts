import { NextResponse } from "next/server";

import { createLoginURL } from "@/lib/login/env";
import { verifyEmail, ZitadelAPIError } from "@/lib/login/zitadel";

export async function POST(request: Request): Promise<Response> {
  const form = await request.formData();
  const userId = String(form.get("userId") || "").trim();
  const verificationCode = String(form.get("verificationCode") || "").trim();
  const loginHint = String(form.get("loginHint") || "").trim();
  const authRequest = String(form.get("authRequest") || "").trim();

  const failureTarget = createLoginURL("/ui/v2/login/verify");
  failureTarget.searchParams.set("userId", userId);
  if (verificationCode) {
    failureTarget.searchParams.set("code", verificationCode);
  }
  if (loginHint) {
    failureTarget.searchParams.set("loginHint", loginHint);
  }
  if (authRequest) {
    failureTarget.searchParams.set("authRequest", authRequest);
  }

  if (!userId || !verificationCode) {
    failureTarget.searchParams.set("error", "userId and verificationCode are required");
    return NextResponse.redirect(failureTarget, { status: 303 });
  }

  try {
    await verifyEmail(userId, verificationCode);
    const target = createLoginURL("/ui/v2/login/login");
    target.searchParams.set("message", "Email verified. Continue with password login.");
    if (loginHint) {
      target.searchParams.set("loginHint", loginHint);
    }
    if (authRequest) {
      target.searchParams.set("authRequest", authRequest);
    }
    return NextResponse.redirect(target, { status: 303 });
  } catch (error) {
    failureTarget.searchParams.set(
      "error",
      error instanceof ZitadelAPIError ? error.message : error instanceof Error ? error.message : "Verify email failed",
    );
    return NextResponse.redirect(failureTarget, { status: 303 });
  }
}
