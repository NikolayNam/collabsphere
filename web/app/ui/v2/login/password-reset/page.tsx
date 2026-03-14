import Link from "next/link";

import { Panel } from "@/components/panel";

type SearchParams = Promise<Record<string, string | string[] | undefined>>;

function readParam(params: Record<string, string | string[] | undefined>, name: string): string {
  const value = params[name];
  return typeof value === "string" ? value.trim() : "";
}

export default async function PasswordResetPage({ searchParams }: { searchParams: SearchParams }) {
  const params = await searchParams;
  const authRequest = readParam(params, "authRequest");
  const loginHint = readParam(params, "loginHint");
  const userId = readParam(params, "userId");
  const verificationCode = readParam(params, "code");
  const error = readParam(params, "error");
  const message = readParam(params, "message");

  return (
    <>
      <Panel title="Password reset" eyebrow="User Service V2">
        <div className={`status-card ${error ? "error" : message ? "success" : "info"}`}>
          <strong>{error ? "Password reset flow failed" : message ? "Code created" : "Reset password through verification code"}</strong>
          <p className="status-copy">
            {error ||
              message ||
              "Сначала запрашивается verification code по loginName, затем новый пароль подтверждается этим кодом."}
          </p>
        </div>
      </Panel>

      <section className="split">
        <Panel title="Request code" eyebrow="Resolve user by loginName">
          <form className="form-grid" action="/ui/v2/login/actions/password-reset-request" method="post">
            {authRequest ? <input type="hidden" name="authRequest" value={authRequest} /> : null}
            <div className="form-row">
              <label className="form-label" htmlFor="resetLoginHint">
                Email or username
              </label>
              <input
                id="resetLoginHint"
                className="text-input"
                autoComplete="username"
                defaultValue={loginHint}
                name="loginHint"
                placeholder="user@example.com"
                required
                type="text"
              />
            </div>
            <div className="button-row">
              <button className="button secondary" type="submit">
                Request verification code
              </button>
            </div>
          </form>
        </Panel>

        <Panel title="Apply new password" eyebrow="Verification code">
          <form className="form-grid" action="/ui/v2/login/actions/password-reset-confirm" method="post">
            {authRequest ? <input type="hidden" name="authRequest" value={authRequest} /> : null}
            <div className="form-row">
              <label className="form-label" htmlFor="resetUserId">
                ZITADEL user ID
              </label>
              <input id="resetUserId" className="text-input" defaultValue={userId} name="userId" required type="text" />
            </div>
            <div className="form-row">
              <label className="form-label" htmlFor="resetCode">
                Verification code
              </label>
              <input id="resetCode" className="text-input" defaultValue={verificationCode} name="verificationCode" required type="text" />
            </div>
            <div className="form-row">
              <label className="form-label" htmlFor="newPassword">
                New password
              </label>
              <input id="newPassword" className="text-input" autoComplete="new-password" name="password" required type="password" />
            </div>
            <div className="button-row">
              <button className="button primary" type="submit">
                Set new password
              </button>
              <Link
                className="button-link secondary"
                href={`/ui/v2/login/login${authRequest ? `?authRequest=${encodeURIComponent(authRequest)}` : ""}${loginHint ? `${authRequest ? "&" : "?"}loginHint=${encodeURIComponent(loginHint)}` : ""}`}
              >
                Back to login
              </Link>
            </div>
          </form>
        </Panel>
      </section>
    </>
  );
}
