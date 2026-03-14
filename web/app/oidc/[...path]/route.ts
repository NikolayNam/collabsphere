import { proxyZitadelRequest } from "@/lib/login/zitadel";

type Params = { params: Promise<{ path: string[] }> };

async function handle(request: Request, context: Params): Promise<Response> {
  const { path } = await context.params;
  return proxyZitadelRequest(request, `/oidc/${path.join("/")}`, new URL(request.url).search);
}

export const GET = handle;
export const POST = handle;
export const PUT = handle;
export const PATCH = handle;
export const DELETE = handle;
export const OPTIONS = handle;
export const HEAD = handle;
