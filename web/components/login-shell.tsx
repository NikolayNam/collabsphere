import Link from "next/link";
import type { PropsWithChildren } from "react";

export function LoginShell({ children }: PropsWithChildren) {
  return (
    <div className="login-shell">
      <header className="login-topbar">
        <div>
          <p className="eyebrow brand-line">
            <img src="/favicon.svg" alt="CollabSphere icon" className="brand-icon" />
            <span>CollabSphere Login</span>
          </p>
          <h1 className="brand">Custom ZITADEL login inside `web/`</h1>
        </div>
        <nav className="nav">
          <Link href="/ui/v2/login/login" className="nav-link">
            Login
          </Link>
          <Link href="/ui/v2/login/login?mode=signup" className="nav-link">
            Signup
          </Link>
          <Link href="/ui/v2/login/password-reset" className="nav-link">
            Reset Password
          </Link>
        </nav>
      </header>
      <main className="page-grid">{children}</main>
    </div>
  );
}
