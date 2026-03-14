import { NextResponse } from "next/server";

import { createLoginURL } from "@/lib/login/env";
import { createHumanUser, ZitadelAPIError } from "@/lib/login/zitadel";

export async function POST(request: Request): Promise<Response> {
  const form = await request.formData();
  const displayName = String(form.get("displayName") || "").trim();
  const email = String(form.get("email") || "").trim();
  const password = String(form.get("password") || "");
  const authRequest = String(form.get("authRequest") || "").trim();

  if (!displayName || !email || !password) {
    const target = createLoginURL("/ui/v2/login/login");
    target.searchParams.set("mode", "signup");
    target.searchParams.set("error", "displayName, email and password are required");
    if (authRequest) {
      target.searchParams.set("authRequest", authRequest);
    }
    if (email) {
      target.searchParams.set("loginHint", email);
    }
    return NextResponse.redirect(target, { status: 303 });
  }

  try {
    const created = await createHumanUser({ displayName, email, password });
    const target = createLoginURL("/");
    target.searchParams.set("signup", "created");
    if (created.verificationCode) {
      target.searchParams.set("verificationCode", created.verificationCode);
    }
    return NextResponse.redirect(target, { status: 303 });
  } catch (error) {
    const message =
      error instanceof ZitadelAPIError
        ? error.message
        : error instanceof Error
          ? error.message
          : "Signup failed";
    const target = createLoginURL("/ui/v2/login/login");
    target.searchParams.set("mode", "signup");
    target.searchParams.set("error", message);
    target.searchParams.set("loginHint", email);
    if (authRequest) {
      target.searchParams.set("authRequest", authRequest);
    }
    return NextResponse.redirect(target, { status: 303 });
  }
}
