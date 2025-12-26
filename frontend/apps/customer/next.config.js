/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,

  // Enable Turbopack for faster dev server (bun compatible)
  experimental: {
    // Turbopack for faster builds
    turbo: {
      rules: {
        // Add any custom file handling here
      },
    },
  },

  // Transpile internal packages for monorepo
  transpilePackages: [
    "@api-billing/ui",
    "@api-billing/auth",
    "@api-billing/query",
    "@api-billing/api-client",
    "@repo/i18n",
  ],

  // Optimize images
  images: {
    formats: ["image/avif", "image/webp"],
    deviceSizes: [640, 750, 828, 1080, 1200, 1920, 2048, 3840],
    imageSizes: [16, 32, 48, 64, 96, 128, 256, 384],
  },

  // Reduce bundle size
  modularizeImports: {
    "@tanstack/react-query": {
      transform: "@tanstack/react-query/{{member}}",
    },
  },

  // Webpack optimizations (used when not using Turbopack)
  webpack: (config, { dev, isServer }) => {
    // Enable tree shaking for production
    if (!dev) {
      config.optimization = {
        ...config.optimization,
        usedExports: true,
        sideEffects: true,
      };
    }
    return config;
  },

  // Output configuration for optimized builds
  output: "standalone",
};

module.exports = nextConfig;
