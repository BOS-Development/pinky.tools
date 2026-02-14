import type { NextConfig } from "next";
import { headers } from "next/headers";

const nextConfig: NextConfig = {
  async rewrites() {
    return [
      {
        source: "/backend/:path*",
        destination: "http://localhost:8081/:path*",
      },
    ];
  },
  async headers() {
    return [
      {
        source: "/backend/:path*",
        headers: [
          {
            key: "BACKEND-KEY",
            value: process.env.BACKEND_KEY as string,
          },
        ],
      },
    ];
  },
  reactStrictMode: true,
};

export default nextConfig;
