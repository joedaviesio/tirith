import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  output: "export",
  turbopack: {
    root: __dirname,
  },
  // Proxy API requests to Go dashboard server during development.
  async rewrites() {
    return [
      {
        source: "/api/:path*",
        destination: "http://localhost:5556/api/:path*",
      },
    ];
  },
};

export default nextConfig;
