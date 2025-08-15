/** @type {import('next').NextConfig} */
const nextConfig = {
  // Configure for /belvo subpath when proxied
  basePath: '/belvo',
  assetPrefix: '/belvo',
  
  // Environment variables
  env: {
    NEXT_PUBLIC_API_URL: process.env.NEXT_PUBLIC_API_URL,
    NEXT_PUBLIC_ENVIRONMENT: process.env.NODE_ENV,
  }
}

module.exports = nextConfig
