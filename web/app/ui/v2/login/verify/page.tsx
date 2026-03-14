import Link from "next/link";

import { Panel } from "@/components/panel";

type SearchParams = Promise<Record<string, string | string[] | undefined>>;

function readParam(params: Record<string, string | string[] | undefined>, name: string): string {
  const value = params[name];
  return typeof value === "string" ? value.trim() : "";
}

export default async function VerifyEmailPage({ searchParams }: { searchParams: SearchParams }) {
  const params = await searchParams;
  const userId = readParam(params, "userId");
  const loginHint = readParam(params, "loginHint");
  const authRequest = readParam(params, "authRequest");
  const verificationCode = readParam(params, "code");
  const message = readParam(params, "message");
  const error = readParam(params, "error");

  return (
    <>
      <Panel title="Verify email" eyebrow="Account activation">
        <div className={`status-card ${error ? "error" : message ? "success" : "info"}`}>
          <strong>{error ? "Не удалось подтвердить email" : message ? "Код обновлён" : "Подтвердите email перед первым login"}</strong>
          <p className="status-copy">{error || message || "Этот экран используется и для signup, и для ручной активации по verification code."}</p>
        </div>
      </Panel>

      <section className="split">
        <Panel title="Verification code" eyebrow="User Service V2">
          <form className="form-grid" action="/ui/v2/login/actions/verify" method="post">
            {authRequest ? <input type="hidden" name="authRequest" value={authRequest} /> : null}
            {loginHint ? <input type="hidden" name="loginHint" value={loginHint} /> : null}
            <div className="form-row">
              <label className="form-label" htmlFor="userId">
                ZITADEL user ID
              </label>
              <input id="userId" className="text-input" defaultValue={userId} name="userId" required type="text" />
            </div>
            <div className="form-row">
              <label className="form-label" htmlFor="verificationCode">
                Verification code
              </label>
              <input id="verificationCode" className="text-input" defaultValue={verificationCode} name="verificationCode" required type="text" />
            </div>
            <div className="button-row">
              <button className="button primary" type="submit">
                Verify email
              </button>
            </div>
          </form>
        </Panel>

        <Panel title="Resend code" eyebrow="Manual recovery">
          <form className="form-grid" action="/ui/v2/login/actions/resend-email" method="post">
            {authRequest ? <input type="hidden" name="authRequest" value={authRequest} /> : null}
            {loginHint ? <input type="hidden" name="loginHint" value={loginHint} /> : null}
            <div className="form-row">
              <label className="form-label" htmlFor="resendUserId">
                ZITADEL user ID
              </label>
              <input id="resendUserId" className="text-input" defaultValue={userId} name="userId" required type="text" />
            </div>
            <div className="button-row">
              <button className="button secondary" type="submit">
                Request new code
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
