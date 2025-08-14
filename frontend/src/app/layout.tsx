import type { Metadata } from 'next'
import { Inter } from 'next/font/google'
import './globals.css'

const inter = Inter({ subsets: ['latin'] })

export const metadata: Metadata = {
  title: 'AI Financial Coach - Belvo Integration',
  description: 'Intelligent financial coaching powered by Belvo API and advanced AI analysis',
  keywords: 'financial coach, AI, Belvo, investment, budget analysis, fintech',
  authors: [{ name: 'AI Financial Coach Team' }],
  viewport: 'width=device-width, initial-scale=1',
  robots: 'index, follow',
  openGraph: {
    title: 'AI Financial Coach',
    description: 'Get personalized financial insights with AI-powered analysis of your Belvo data',
    type: 'website',
    locale: 'en_US',
  },
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="en" className="smooth-scroll">
      <body className={`${inter.className} antialiased`}>
        <div className="min-h-screen">
          {children}
        </div>
      </body>
    </html>
  )
}