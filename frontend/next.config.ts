import type { NextConfig } from "next";
import { headers } from "next/headers";

const nextConfig: NextConfig = {
  async rewrites() {
    return [
      {
        source: "/backend/:path",
        destination: "http://localhost:8081/" + ":path*", // process.env.API_SERVICE + ":path*",
      },
    ];
  },
  async headers() {
    return [
      {
        source: "/backend",
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
