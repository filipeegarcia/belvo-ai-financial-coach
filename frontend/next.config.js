/** @type {import('next').NextConfig} */
const nextConfig = {
  basePath: process.env.NODE_ENV === 'production' ? '/belvo' : '',
  assetPrefix: process.env.NODE_ENV === 'production' ? '/belvo' : '',
  trailingSlash: true,
  output: 'standalone',
  
  // API configuration for production
  async rewrites() {
    return [
      {
        source: '/api/:path*',
        destination: process.env.NODE_ENV === 'production' 
          ? `${process.env.NEXT_PUBLIC_API_URL || 'https://api.filipegarcia.co'}/api/:path*`
          : 'http://localhost:8000/api/:path*',
      },
    ]
  },

  // Environment variables
  env: {
    NEXT_PUBLIC_API_URL: process.env.NEXT_PUBLIC_API_URL,
    NEXT_PUBLIC_ENVIRONMENT: process.env.NODE_ENV,
  }
}

module.exports = nextConfig
