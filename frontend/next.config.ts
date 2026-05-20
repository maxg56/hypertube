import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  experimental: {
    trustHostHeader: true,
  },
};

export default nextConfig;
