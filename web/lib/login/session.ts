import { createHmac, timingSafeEqual } from "node:crypto";
import { cookies } from "next/headers";

import { getSessionCookieName, getSessionSecret } from "@/lib/login/env";

export type LoginSessionState = {
  sessionId: string;
  sessionToken: string;
  userId?: string;
  loginName?: string;
  displayName?: string;
};

function toBase64URL(value: string): string {
  return Buffer.from(value, "utf8").toString("base64url");
}

function fromBase64URL(value: string): string {
  return Buffer.from(value, "base64url").toString("utf8");
}

function sign(value: string): string {
  return createHmac("sha256", getSessionSecret()).update(value).digest("base64url");
}

function serialize(payload: LoginSessionState): string {
  const raw = toBase64URL(JSON.stringify(payload));
  return `${raw}.${sign(raw)}`;
}

function deserialize(raw: string | undefined): LoginSessionState | null {
  if (!raw) {
    return null;
  }
  const [payload, signature] = raw.split(".", 2);
  if (!payload || !signature) {
    return null;
  }
  const expected = Buffer.from(sign(payload), "utf8");
  const actual = Buffer.from(signature, "utf8");
  if (expected.length !== actual.length || !timingSafeEqual(expected, actual)) {
    return null;
  }
  try {
    const decoded = JSON.parse(fromBase64URL(payload)) as Partial<LoginSessionState>;
    if (!decoded.sessionId || !decoded.sessionToken) {
      return null;
    }
    return {
      sessionId: decoded.sessionId,
      sessionToken: decoded.sessionToken,
      userId: decoded.userId,
      loginName: decoded.loginName,
      displayName: decoded.displayName,
    };
  } catch {
    return null;
  }
}

export async function readLoginSession(): Promise<LoginSessionState | null> {
  const cookieStore = await cookies();
  return deserialize(cookieStore.get(getSessionCookieName())?.value);
}

export async function writeLoginSession(session: LoginSessionState): Promise<void> {
  const cookieStore = await cookies();
  cookieStore.set(getSessionCookieName(), serialize(session), {
    httpOnly: true,
    sameSite: "lax",
    secure: false,
    path: "/",
    maxAge: 60 * 30,
  });
}

export async function clearLoginSession(): Promise<void> {
  const cookieStore = await cookies();
  cookieStore.set(getSessionCookieName(), "", {
    httpOnly: true,
    sameSite: "lax",
    secure: false,
    path: "/",
    maxAge: 0,
    expires: new Date(0),
  });
}
