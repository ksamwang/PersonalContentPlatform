import type { NextConfig } from "next";
const nextConfig: NextConfig = {
  output: "standalone",
  async rewrites() {
    return [
      {
        source: "/media/:assetID/:recipe",
        destination: `${process.env.API_ORIGIN ?? "http://localhost:8080"}/v1/public/assets/:assetID/:recipe`,
      },
      {
        source: "/media/:assetID",
        destination: `${process.env.API_ORIGIN ?? "http://localhost:8080"}/v1/public/assets/:assetID`,
      },
      {
        source: "/api/:path*",
        destination: `${process.env.API_ORIGIN ?? "http://localhost:8080"}/:path*`,
      },
    ];
  },
};
export default nextConfig;
