import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  output: "standalone",
  typescript: {
    // Évite les conflits de types générés par Turbopack lors du build Docker
    ignoreBuildErrors: true,
  },
};

export default nextConfig;
