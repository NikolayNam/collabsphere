import Link from "next/link";
import type { PropsWithChildren } from "react";

const navItems = [
  { href: "/", label: "Overview" },
  { href: "/login", label: "Login" },
  { href: "/me", label: "Me" },
  { href: "/organizations", label: "Organizations" },
  { href: "/chat", label: "Chat" },
];

export function AppShell({ children }: PropsWithChildren) {
  return (
    <div className="app-shell">
      <header className="topbar">
        <div>
          <p className="eyebrow">CollabSphere Web</p>
          <h1 className="brand">Next.js shell over the Go backend</h1>
        </div>
        <nav className="nav">
          {navItems.map((item) => (
            <Link key={item.href} href={item.href} className="nav-link">
              {item.label}
            </Link>
          ))}
        </nav>
      </header>
      <main className="page-grid">{children}</main>
    </div>
  );
}
