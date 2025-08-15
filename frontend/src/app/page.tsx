'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'

const API_URL = (process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8000').replace(/\/$/, '')

export default function AuthPage() {
  const [secretId, setSecretId] = useState('')
  const [secretKey, setSecretKey] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const router = useRouter()

  const handleTestCredentials = () => {
    setSecretId('397581e3-22a5-4872-b11e-f12ff3c654b4')
    setSecretKey('ifN@BQCu9s3xaad38j_*rNj@IbWEIK7LoAWXlH-pxhiPcOZfvKWKbivBeDlAv0k1')
  }



  const handleConnect = async () => {
    if (!secretId.trim() || !secretKey.trim()) {
      setError('Please enter both Secret ID and Secret Key')
      return
    }

    setLoading(true)
    setError('')

    try {
      console.log('🔐 Step 1: Testing Belvo authentication...')
      
      // First, test Belvo connection with provided credentials
      const authResponse = await fetch(`${API_URL}/api/belvo/test-connection`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          secret_id: secretId,
          secret_key: secretKey
        })
      })

      if (!authResponse.ok) {
        throw new Error('Invalid Belvo credentials. Please check your Secret ID and Secret Key.')
      }

      const authData = await authResponse.json()
      if (!authData.data?.connected && authData.data?.status !== 'success') {
        throw new Error('Failed to authenticate with Belvo API')
      }

      console.log('✅ Belvo authentication successful')

                     console.log('🔍 Step 2: Credentials verified, proceeding to link selection...')

                     // Store session with credentials only - link selection will happen in chat
               sessionStorage.setItem('belvo_session', JSON.stringify({
                   secretId,
                   secretKey,
                   authenticated: true,
                   method: 'link_selection',
                   timestamp: Date.now()
               }))

               console.log('🎉 Authentication complete, redirecting to AI assistant...')
               router.push('/chat')

    } catch (error) {
      console.error('❌ Connection error:', error)
      setError('Connection failed: ' + (error as Error).message)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen bg-gray-100 flex items-center justify-center p-4">
      <div className="bg-white rounded-lg shadow-lg p-8 w-full max-w-md">
        <div className="text-center mb-8">
          <h1 className="text-3xl font-bold text-gray-900 mb-2">
            🤖 AI Financial Coach
          </h1>
          <p className="text-gray-600">
            Connect to Belvo Sandbox
          </p>
        </div>

        <div className="space-y-4">
          {/* Secret ID Input */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Belvo Secret ID
            </label>
            <input
              type="text"
              value={secretId}
              onChange={(e) => setSecretId(e.target.value)}
              placeholder="Enter your Belvo Secret ID"
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>

          {/* Secret Key Input */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Belvo Secret Key
            </label>
            <input
              type="password"
              value={secretKey}
              onChange={(e) => setSecretKey(e.target.value)}
              placeholder="Enter your Belvo Secret Key"
              className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>

          {/* Test Credentials Button */}
          <button
            onClick={handleTestCredentials}
            type="button"
            className="w-full bg-gray-200 text-gray-700 py-2 px-4 rounded-md hover:bg-gray-300 transition-colors"
          >
            🔧 Use Test Credentials
          </button>

          {/* Single Connect Button */}
          <button
            onClick={handleConnect}
            disabled={loading || !secretId.trim() || !secretKey.trim()}
            className="w-full bg-blue-600 text-white py-2 px-4 rounded-md hover:bg-blue-700 disabled:bg-gray-400 disabled:cursor-not-allowed transition-colors"
          >
            {loading ? 'Connecting...' : '🚀 Connect to AI Financial Coach'}
          </button>

          {/* Error Message */}
          {error && (
            <div className="bg-red-100 border border-red-300 text-red-700 px-4 py-3 rounded">
              {error}
            </div>
          )}

          {/* Info */}
          <div className="bg-blue-50 border border-blue-200 rounded-md p-3">
            <p className="text-blue-800 text-sm">
              ℹ️ <strong>Smart Connection:</strong><br/>
              🔐 <strong>Validates your Belvo API credentials</strong><br/>
              🏦 <strong>Discovers all customer accounts linked to your credentials</strong><br/>
              🤖 <strong>Provides personalized AI coaching for each customer</strong>
            </p>
          </div>
          

        </div>
      </div>
    </div>
  )
}