/** @type {import('next').NextConfig} */
const nextConfig = {
  basePath: '/belvo',
  trailingSlash: false,
  env: {
    NEXT_PUBLIC_API_URL: process.env.NEXT_PUBLIC_API_URL,
    NEXT_PUBLIC_ENVIRONMENT: process.env.NODE_ENV,
  },
};

module.exports = nextConfig;
