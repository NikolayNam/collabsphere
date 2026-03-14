import type { NextConfig } from "next";

const publicAPIBaseURL = (process.env.NEXT_PUBLIC_API_BASE_URL || "http://api.localhost:8080").replace(/\/+$/, "");
const internalAPIBaseURL = (process.env.NEXT_INTERNAL_API_BASE_URL || publicAPIBaseURL).replace(/\/+$/, "");

const nextConfig: NextConfig = {
  reactStrictMode: true,
  output: "standalone",
  async rewrites() {
    return [
      {
        source: "/api/backend/:path*",
        destination: `${internalAPIBaseURL}/:path*`,
      },
    ];
  },
};

export default nextConfig;
