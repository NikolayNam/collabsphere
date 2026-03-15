/**
 * Proxy for multipart/form-data uploads.
 * Next.js rewrites can fail to forward multipart bodies correctly;
 * this route explicitly streams the request to the backend.
 */
import { NextRequest, NextResponse } from "next/server";

const internalAPIBaseURL = (
  process.env.NEXT_INTERNAL_API_BASE_URL || process.env.NEXT_PUBLIC_API_BASE_URL || "http://api:8080"
).replace(/\/+$/, "");

const FORWARD_HEADERS = ["authorization", "content-type", "content-length"];

export async function POST(
  request: NextRequest,
  { params }: { params: Promise<{ path: string[] }> }
) {
  const { path } = await params;
  const pathStr = path.join("/");
  const backendURL = `${internalAPIBaseURL}/${pathStr}`;

  const headers = new Headers();
  for (const name of FORWARD_HEADERS) {
    const value = request.headers.get(name);
    if (value) headers.set(name, value);
  }

  try {
    const fetchOpts: RequestInit & { duplex?: "half" } = {
      method: "POST",
      headers,
      body: request.body,
      ...(request.body && { duplex: "half" as const }),
    };

    const response = await fetch(backendURL, fetchOpts);

    const responseHeaders = new Headers();
    response.headers.forEach((value, key) => {
      if (key.toLowerCase() !== "transfer-encoding") {
        responseHeaders.set(key, value);
      }
    });

    return new NextResponse(response.body, {
      status: response.status,
      statusText: response.statusText,
      headers: responseHeaders,
    });
  } catch (err) {
    console.error("[upload-proxy] fetch failed:", err);
    return NextResponse.json(
      { detail: "Upload proxy failed", error: String(err) },
      { status: 502 }
    );
  }
}
