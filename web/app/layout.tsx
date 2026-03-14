import type { Metadata } from "next";
import type { PropsWithChildren } from "react";
import { headers } from "next/headers";

import { AppShell } from "@/components/app-shell";
import { LoginShell } from "@/components/login-shell";
import { isLoginHostValue } from "@/lib/login/env";
import "./globals.css";

export const metadata: Metadata = {
  title: "CollabSphere Web",
  description: "Next.js frontend shell for the CollabSphere Go backend.",
};

export default async function RootLayout({ children }: PropsWithChildren) {
  const headerList = await headers();
  const host = headerList.get("x-collabsphere-request-host") || headerList.get("host");
  const shell = isLoginHostValue(host) ? <LoginShell>{children}</LoginShell> : <AppShell>{children}</AppShell>;

  return (
    <html lang="ru">
      <body>{shell}</body>
    </html>
  );
}
