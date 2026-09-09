import type { NextConfig } from "next";
const config: NextConfig = {
  output: "standalone",
  async rewrites(){return [
    {source:"/media/:assetID/:recipe",destination:`${process.env.API_ORIGIN ?? "http://localhost:8080"}/v1/public/assets/:assetID/:recipe`},
    {source:"/media/:assetID",destination:`${process.env.API_ORIGIN ?? "http://localhost:8080"}/v1/public/assets/:assetID`},
  ]},
};
export default config;
