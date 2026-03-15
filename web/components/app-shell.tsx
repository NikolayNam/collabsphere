import Link from "next/link";
import type { PropsWithChildren } from "react";

const navItems = [
  { href: "/", label: "Overview" },
  { href: "/login", label: "Login" },
  { href: "/me", label: "Me" },
  { href: "/organizations", label: "Organizations" },
  { href: "/admin/users", label: "Admin Users" },
  { href: "/admin/limits", label: "Admin Limits" },
  { href: "/chat", label: "Chat" },
];

export function AppShell({ children }: PropsWithChildren) {
  return (
    <div className="app-shell">
      <header className="topbar">
        <div>
          <p className="eyebrow brand-line">
            <img src="/favicon.svg" alt="CollabSphere icon" className="brand-icon" />
            <span>CollabSphere Web</span>
          </p>
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
