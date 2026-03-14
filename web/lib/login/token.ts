import { readFile } from "node:fs/promises";

let cachedServiceToken: string | null = null;

export async function getServiceUserToken(): Promise<string> {
  if (cachedServiceToken) {
    return cachedServiceToken;
  }
  if (process.env.WEB_LOGIN_SERVICE_USER_TOKEN?.trim()) {
    cachedServiceToken = process.env.WEB_LOGIN_SERVICE_USER_TOKEN.trim();
    return cachedServiceToken;
  }
  const tokenFile = process.env.WEB_LOGIN_SERVICE_USER_TOKEN_FILE?.trim();
  if (!tokenFile) {
    throw new Error("WEB_LOGIN_SERVICE_USER_TOKEN or WEB_LOGIN_SERVICE_USER_TOKEN_FILE is required");
  }
  const raw = await readFile(tokenFile, "utf8");
  const token = raw.trim();
  if (!token) {
    throw new Error("WEB_LOGIN service PAT file is empty");
  }
  cachedServiceToken = token;
  return token;
}
