'use client'

import { useState, useEffect, useRef } from 'react'
import { useRouter } from 'next/navigation'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'

interface Message {
  id: string
  role: 'user' | 'assistant'
  content: string
  timestamp: Date
}

interface BelvoSession {
  secretId: string
  secretKey: string
  linkId?: string
  authenticated: boolean
  dataStatus?: string
  timestamp: number
  method?: 'direct' | 'widget' | 'link_selection'
  widgetUrl?: string
}

interface BelvoLink {
  id: string
  institution: string
  display_name: string
  status: string
  created_at: string
  customer_number: number
  short_id: string
}

interface DetailedBelvoLink {
  link_id: string
  owner_name: string
  account_count: number
  accounts: any[]
  account_categories: any
  accounts_by_category: any
  transaction_count: number
  recent_transactions: any[]
  total_balance: number
  currency: string
  has_data: boolean
  financial_summary: any
  ai_context_summary: string
  data_scope?: string
}

export default function ChatPage() {
  const [messages, setMessages] = useState<Message[]>([])
  const [inputMessage, setInputMessage] = useState('')
  const [isLoading, setIsLoading] = useState(false)
  const [session, setSession] = useState<BelvoSession | null>(null)
  const [language, setLanguage] = useState<'en' | 'pt'>('en')
  const [availableLinks, setAvailableLinks] = useState<BelvoLink[]>([])
  const [selectedLink, setSelectedLink] = useState<DetailedBelvoLink | null>(null)
  const [showLinkSelection, setShowLinkSelection] = useState(false)
  const [isLoadingDetails, setIsLoadingDetails] = useState(false)
  const messagesEndRef = useRef<HTMLDivElement>(null)
  const router = useRouter()

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }

  useEffect(() => {
    scrollToBottom()
  }, [messages])

  useEffect(() => {
    // Check if user is authenticated (try both session keys)
    const sessionData = sessionStorage.getItem('belvo_session') || sessionStorage.getItem('belvo_credentials')
    if (!sessionData) {
      router.push('/')
      return
    }

    const parsedSession = JSON.parse(sessionData) as BelvoSession
    setSession(parsedSession)

    // If link_selection method, fetch available links
    if (parsedSession.method === 'link_selection') {
      fetchAvailableLinks(parsedSession)
      setShowLinkSelection(true)
      return
    }

    // Welcome message
    const welcomeMessage: Message = {
      id: Date.now().toString(),
      role: 'assistant',
      content: language === 'en' ? 
        `👋 Welcome to your AI Financial Coach!\n\n` +
        `I'm here to help you make informed financial decisions based on:\n` +
        `• Your ${parsedSession.method === 'widget' ? 'real erebor_br_retail data (via widget)' : parsedSession.method === 'direct_link' ? 'real erebor_br_retail data (via direct API)' : 'financial data (with fallback to mock)'}\n` +
        `• Real-time market data (stocks, crypto, Brazilian rates)\n` +
        `• Advanced AI analysis of your spending and investment patterns\n\n` +
        `I can help you with:\n` +
        `• Investment recommendations based on your financial profile\n` +
        `• Budget analysis and spending insights\n` +
        `• Portfolio suggestions (Conservative/Balanced/Aggressive)\n` +
        `• Market opportunities tailored to your situation\n\n` +
        `🔒 Disclaimer: I'm an AI assistant, not a licensed financial advisor. My recommendations are for educational purposes only.\n\n` +
        `How can I help you today?` :
        `👋 Bem-vindo ao seu Coach Financeiro com IA!\n\n` +
        `Estou aqui para ajudá-lo a tomar decisões financeiras informadas com base em:\n` +
        `• Seus ${parsedSession.method === 'widget' ? 'dados reais do erebor_br_retail (via widget)' : parsedSession.method === 'direct_link' ? 'dados reais do erebor_br_retail (via API direta)' : 'dados financeiros (com fallback para mock)'}\n` +
        `• Dados de mercado em tempo real (ações, cripto, taxas brasileiras)\n` +
        `• Análise avançada de IA dos seus padrões de gastos e investimentos\n\n` +
        `Posso ajudar você com:\n` +
        `• Recomendações de investimento baseadas no seu perfil financeiro\n` +
        `• Análise de orçamento e insights de gastos\n` +
        `• Sugestões de portfólio (Conservador/Moderado/Agressivo)\n` +
        `• Oportunidades de mercado personalizadas para sua situação\n\n` +
        `🔒 Aviso: Sou um assistente de IA, não um consultor financeiro licenciado. Minhas recomendações são apenas para fins educacionais.\n\n` +
        `Como posso ajudar você hoje?`,
      timestamp: new Date()
    }
    setMessages([welcomeMessage])
  }, [language])

  const sendMessage = async () => {
    if (!inputMessage.trim() || !session) return

    const userMessage: Message = {
      id: Date.now().toString(),
      role: 'user',
      content: inputMessage,
      timestamp: new Date()
    }

    setMessages(prev => [...prev, userMessage])
    const messageText = inputMessage
    setInputMessage('')
    setIsLoading(true)

    console.log('🚀 Sending message to AI:', {
      message: messageText,
      language: language === 'en' ? 'en' : 'pt',
      credential_mode: 'custom',
      link_id: session.linkId,
      session
    })

    try {
      // First check if backend is running
      console.log('🔍 Step 1: Checking backend health...')
      const healthResponse = await fetch('http://localhost:8000/health')
      
      if (!healthResponse.ok) {
        throw new Error('Backend server is not running. Please start the GoFr backend.')
      }
      
      console.log('✅ Backend is running')

      // Send the AI chat request
      console.log('🤖 Step 2: Sending AI chat request...')
      const response = await fetch('http://localhost:8000/api/ai/chat', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          message: messageText,
          language: language === 'en' ? 'en' : 'pt',
          credential_mode: 'custom',
          link_id: selectedLink?.link_id || session.linkId,
          secret_id: session.secretId,
          secret_key: session.secretKey
        })
      })

      console.log('📡 Response status:', response.status)
      console.log('📡 Response headers:', Object.fromEntries(response.headers.entries()))

      if (!response.ok) {
        const errorText = await response.text()
        console.error('❌ AI API Error:', {
          status: response.status,
          statusText: response.statusText,
          body: errorText
        })
        throw new Error(`AI API failed with status ${response.status}: ${errorText}`)
      }

      const data = await response.json()
      console.log('📦 AI Response data:', data)
      
      // Try different response formats
      let aiContent = 
        data.data?.chat_response?.message ||
        data.chat_response?.message ||
        data.data?.chat_response?.response ||
        data.chat_response?.response ||
        data.chat_response ||
        data.data?.response ||
        data.response ||
        data.message ||
        'Sorry, I could not process your request.'

      console.log('💬 Extracted AI content:', aiContent)
      
      const aiMessage: Message = {
        id: (Date.now() + 1).toString(),
        role: 'assistant',
        content: aiContent,
        timestamp: new Date()
      }

      setMessages(prev => [...prev, aiMessage])

    } catch (error) {
      console.error('💥 Chat error:', error)
      const errorMessage: Message = {
        id: (Date.now() + 1).toString(),
        role: 'assistant',
        content: language === 'en' ?
          `Sorry, I encountered an error: ${(error as Error).message}` :
          `Desculpe, encontrei um erro: ${(error as Error).message}`,
        timestamp: new Date()
      }
      setMessages(prev => [...prev, errorMessage])
    } finally {
      setIsLoading(false)
    }
  }

  const fetchAvailableLinks = async (session: BelvoSession) => {
    try {
      console.log('🔍 Fetching available links...')
      const response = await fetch('http://localhost:8000/api/belvo/links/for-selection', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          secret_id: session.secretId,
          secret_key: session.secretKey
        })
      })

      console.log('📡 Response status:', response.status)
      
      if (response.ok) {
        const data = await response.json()
        console.log('📦 Received links data:', data)
        const links = data.data?.links || []
        console.log('🔗 Setting available links:', links.length, 'links')
        setAvailableLinks(links)
      } else {
        console.error('❌ Failed to fetch links, status:', response.status)
      }
    } catch (error) {
      console.error('💥 Error fetching links:', error)
    }
  }

  const handleLinkSelection = async (link: BelvoLink) => {
    setIsLoadingDetails(true)
    setShowLinkSelection(false)
    
    // Show enhanced loading message with progress indicators
    const loadingMessage: Message = {
      id: Date.now().toString(),
      role: 'assistant',
      content: language === 'en' ? 
        `🔄 **Analyzing Financial Data for ${link.display_name}**\n\n` +
        `📡 **Step 1/4**: Connecting to bank systems...\n` +
        `🏦 **Step 2/4**: Retrieving account information...\n` +
        `💳 **Step 3/4**: Collecting transaction history...\n` +
        `🤖 **Step 4/4**: Preparing AI financial analysis...\n\n` +
        `⏱️ **Estimated time**: 30-60 seconds\n\n` +
        `💡 **What's happening**: I'm securely gathering your complete financial profile to provide personalized coaching and insights.` :
        `🔄 **Analisando Dados Financeiros para ${link.display_name}**\n\n` +
        `📡 **Etapa 1/4**: Conectando aos sistemas bancários...\n` +
        `🏦 **Etapa 2/4**: Recuperando informações da conta...\n` +
        `💳 **Etapa 3/4**: Coletando histórico de transações...\n` +
        `🤖 **Etapa 4/4**: Preparando análise financeira de IA...\n\n` +
        `⏱️ **Tempo estimado**: 30-60 segundos\n\n` +
        `💡 **O que está acontecendo**: Estou coletando com segurança seu perfil financeiro completo para fornecer coaching e insights personalizados.`,
      timestamp: new Date()
    }
    setMessages([loadingMessage])

    try {
      // Load detailed data for selected customer
      console.log('🔍 Loading detailed data for:', link.id)
      const response = await fetch(`http://localhost:8000/api/belvo/links/detailed-info/${link.id}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          secret_id: session!.secretId,
          secret_key: session!.secretKey
        })
      })

      if (response.ok) {
        const data = await response.json()
        const detailedLink: DetailedBelvoLink = data.data
        console.log('✅ Detailed data loaded:', detailedLink)
        
        setSelectedLink(detailedLink)
        
        // Cache ALL Belvo data for complete AI context
        console.log('💾 Caching ALL Belvo financial data for AI...')
        try {
          await fetch('http://localhost:8000/api/ai/cache-context', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
              link_id: detailedLink.link_id,
              owner_name: detailedLink.owner_name,
              financial_summary: {
                ...detailedLink.financial_summary,
                // Core transaction and account data
                recent_transactions: detailedLink.recent_transactions,
                accounts: detailedLink.accounts,
                transaction_count: detailedLink.transaction_count,
                account_count: detailedLink.account_count,
                // Additional context data for AI analysis
                account_categories: detailedLink.account_categories,
                accounts_by_category: detailedLink.accounts_by_category,
                total_balance: detailedLink.total_balance,
                currency: detailedLink.currency,
                has_data: detailedLink.has_data,
                ai_context_summary: detailedLink.ai_context_summary,
                data_scope: detailedLink.data_scope || "recent"
              }
            })
          })
          console.log('✅ ALL Belvo data cached successfully for AI analysis')
        } catch (error) {
          console.warn('⚠️ Failed to cache complete financial context:', error)
        }
        
        // Update session with selected link
        const updatedSession = {
          ...session!,
          linkId: link.id,
          method: 'selected_link' as const
        }
        setSession(updatedSession)
        sessionStorage.setItem('belvo_session', JSON.stringify(updatedSession))

        // Show success message with comprehensive data
        const monthlyIncome = detailedLink.financial_summary?.monthly_income || 0
        const monthlyExpenses = detailedLink.financial_summary?.monthly_variable_expenses || 0
        const savingsRate = monthlyIncome > 0 ? ((monthlyIncome - monthlyExpenses) / monthlyIncome * 100) : 0
        
        const welcomeMessage: Message = {
          id: Date.now().toString(),
          role: 'assistant',
          content: language === 'en' ? 
            `🎉 **Financial Analysis Complete for ${detailedLink.owner_name}!**\n\n` +
            `✅ **Data Successfully Loaded:**\n` +
            `• **${detailedLink.account_count} accounts** with total balance: **${detailedLink.total_balance.toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' })}**\n` +
            `• **${detailedLink.transaction_count} recent transactions** analyzed\n` +
            `• **Monthly income**: ${monthlyIncome.toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' })}\n` +
            `• **Monthly expenses**: ${monthlyExpenses.toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' })}\n` +
            `• **Savings rate**: ${savingsRate.toFixed(1)}%\n\n` +
            `🚀 **AI Financial Coach Ready!**\n\n` +
            `I can now provide instant, personalized advice on:\n` +
            `• 💰 **Budget optimization** based on real spending patterns\n` +
            `• 📈 **Investment strategies** tailored to your financial profile\n` +
            `• 🎯 **Savings goals** and actionable plans\n` +
            `• 📊 **Expense analysis** and cost-cutting opportunities\n\n` +
            `🔒 *Disclaimer: I'm an AI assistant, not a licensed financial advisor. My recommendations are for educational purposes only.*\n\n` +
            `💬 **What would you like to know about ${detailedLink.owner_name}'s financial situation?**` :
            `🎉 **Análise Financeira Completa para ${detailedLink.owner_name}!**\n\n` +
            `✅ **Dados Carregados com Sucesso:**\n` +
            `• **${detailedLink.account_count} contas** com saldo total: **${detailedLink.total_balance.toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' })}**\n` +
            `• **${detailedLink.transaction_count} transações recentes** analisadas\n` +
            `• **Renda mensal**: ${monthlyIncome.toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' })}\n` +
            `• **Gastos mensais**: ${monthlyExpenses.toLocaleString('pt-BR', { style: 'currency', currency: 'BRL' })}\n` +
            `• **Taxa de poupança**: ${savingsRate.toFixed(1)}%\n\n` +
            `🚀 **Coach Financeiro IA Pronto!**\n\n` +
            `Agora posso fornecer conselhos instantâneos e personalizados sobre:\n` +
            `• 💰 **Otimização de orçamento** baseada em padrões reais de gastos\n` +
            `• 📈 **Estratégias de investimento** personalizadas para seu perfil\n` +
            `• 🎯 **Objetivos de poupança** e planos acionáveis\n` +
            `• 📊 **Análise de despesas** e oportunidades de redução de custos\n\n` +
            `🔒 *Aviso: Sou um assistente de IA, não um consultor financeiro licenciado. Minhas recomendações são apenas para fins educacionais.*\n\n` +
            `💬 **O que você gostaria de saber sobre a situação financeira de ${detailedLink.owner_name}?**`,
          timestamp: new Date()
        }

        setMessages([welcomeMessage])
      } else {
        throw new Error('Failed to load detailed customer data')
      }
    } catch (error) {
      console.error('💥 Error loading detailed data:', error)
      const errorMessage: Message = {
        id: Date.now().toString(),
        role: 'assistant',
        content: language === 'en' ? 
          `❌ Sorry, I couldn't load the detailed financial data for ${link.display_name}.\n\nError: ${(error as Error).message}\n\nPlease try selecting a different customer link or contact support.` :
          `❌ Desculpe, não consegui carregar os dados financeiros detalhados para ${link.display_name}.\n\nErro: ${(error as Error).message}\n\nPor favor, tente selecionar um link de cliente diferente ou entre em contato com o suporte.`,
        timestamp: new Date()
      }
      setMessages([errorMessage])
      setShowLinkSelection(true) // Go back to selection
    } finally {
      setIsLoadingDetails(false)
    }
  }

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      sendMessage()
    }
  }

  if (!session) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto mb-4"></div>
          <p>{language === 'en' ? 'Checking authentication...' : 'Verificando autenticação...'}</p>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-50 via-gray-100 to-gray-200 flex flex-col relative overflow-hidden">
      {/* Background Effects */}
      <div className="absolute inset-0 bg-grid-pattern opacity-3"></div>
      <div className="absolute top-0 left-0 w-full h-full bg-gradient-to-br from-gray-900/3 to-transparent"></div>

      {/* Header */}
      <div className="bg-white/80 backdrop-blur-xl border-b border-gray-200/50 px-6 py-4 flex items-center justify-between shadow-sm relative z-10">
        <div>
          <h1 className="text-2xl font-bold bg-gradient-to-r from-gray-800 to-gray-600 bg-clip-text text-transparent">
            🤖 AI Financial Coach
          </h1>
          <p className="text-sm text-gray-600 font-medium">
            {selectedLink ? 
              `🏦 ${selectedLink.owner_name} (${selectedLink.institution})` :
              session?.dataStatus || (language === 'en' ? 'Select a customer account' : 'Selecione uma conta de cliente')
            }
          </p>
        </div>
        <div className="flex items-center gap-3">
          <button
            onClick={() => setLanguage(prev => prev === 'en' ? 'pt' : 'en')}
            className="bg-gray-100/80 hover:bg-gray-200/80 border border-gray-200 text-gray-600 hover:text-gray-800 px-3 py-2 rounded-full text-sm font-medium transition-all duration-200 backdrop-blur-sm"
          >
            {language === 'en' ? '🇧🇷 PT' : '🇺🇸 EN'}
          </button>
          <button
            onClick={() => {
              sessionStorage.removeItem('belvo_session')
              router.push('/')
            }}
            className="bg-gray-100/80 hover:bg-gray-200/80 border border-gray-200 text-gray-600 hover:text-gray-800 px-3 py-2 rounded-full text-sm font-medium transition-all duration-200 backdrop-blur-sm"
          >
            {language === 'en' ? '🚪 Logout' : '🚪 Sair'}
          </button>
        </div>
      </div>

      {/* Link Selection Screen */}
      {showLinkSelection && (
        <div className="flex-1 overflow-y-auto p-6 space-y-6 relative z-10">
          <div className="max-w-4xl mx-auto space-y-6">
            <div className="bg-white/80 backdrop-blur-xl rounded-2xl shadow-xl border border-white/30 p-8">
              <h2 className="text-2xl font-bold text-gray-800 mb-4 text-center">
                {language === 'en' ? '🔗 Choose a Customer Link' : '🔗 Escolha um Link de Cliente'}
              </h2>
              <p className="text-gray-600 text-center mb-6">
                {language === 'en' 
                  ? 'Select which customer link you want to analyze. Each link represents a connection to a customer\'s financial data:' 
                  : 'Selecione qual link de cliente você deseja analisar. Cada link representa uma conexão com os dados financeiros de um cliente:'}
              </p>
              
              <div className="grid gap-4 max-h-96 overflow-y-auto">
                {console.log('🎯 Rendering links:', availableLinks.length, availableLinks)}
                {availableLinks.length === 0 ? (
                                      <div className="text-center text-gray-500 py-8">
                      <p>{language === 'en' ? 'Loading customer links...' : 'Carregando links de clientes...'}</p>
                      <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600 mx-auto mt-4"></div>
                    </div>
                ) : (
                  availableLinks.map((link) => (
                  <button
                    key={link.id}
                    onClick={() => handleLinkSelection(link)}
                    className="bg-white/60 hover:bg-white/80 border border-gray-200/50 hover:border-gray-300/50 rounded-xl p-4 text-left transition-all duration-200 hover:shadow-lg hover:scale-[1.02]"
                  >
                    <div className="flex items-center justify-between">
                      <div className="flex-1">
                        <h3 className="font-semibold text-gray-800 text-lg">{link.display_name}</h3>
                        <p className="text-sm text-gray-600 mt-1">
                          📅 {language === 'en' ? 'Created:' : 'Criado:'} {new Date(link.created_at).toLocaleDateString()}
                        </p>
                        <p className="text-sm text-gray-600 mt-1">
                          🏦 {language === 'en' ? 'Institution:' : 'Instituição:'} {link.institution}
                        </p>
                        <p className="text-sm text-gray-500 mt-1">
                          ✅ {language === 'en' ? 'Status:' : 'Status:'} {link.status} • 🔗 {language === 'en' ? 'ID:' : 'ID:'} {link.short_id}
                        </p>
                      </div>
                      <div className="flex items-center gap-2">
                        <span className="bg-green-100 text-green-700 px-2 py-1 rounded-full text-xs font-medium">
                          {language === 'en' ? '⚡ Instant' : '⚡ Instantâneo'}
                        </span>
                        <span className="text-gray-400">→</span>
                      </div>
                    </div>
                  </button>
                  ))
                )}
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Chat Messages */}
      {!showLinkSelection && (
        <div className="flex-1 overflow-y-auto p-6 space-y-6 relative z-10">
        <div className="max-w-4xl mx-auto space-y-6">
          {messages.map((message) => (
            <div
              key={message.id}
              className={`flex items-start gap-4 ${message.role === 'user' ? 'flex-row-reverse' : 'flex-row'} group`}
            >
              {/* Avatar */}
              <div className={`flex-shrink-0 w-10 h-10 rounded-full flex items-center justify-center shadow-lg ${
                message.role === 'user' 
                  ? 'bg-gradient-to-br from-gray-700 to-gray-800' 
                  : 'bg-gradient-to-br from-gray-100 to-gray-200'
              }`}>
                <span className="text-lg">
                  {message.role === 'user' ? '👤' : '🤖'}
                </span>
              </div>

              {/* Message Bubble */}
              <div
                className={`max-w-2xl px-6 py-4 rounded-2xl shadow-lg transition-all duration-200 ${
                  message.role === 'user'
                    ? 'bg-blue-600 text-white'
                    : 'bg-white border border-gray-300 text-gray-900'
                } group-hover:shadow-xl`}
              >
                <div className="text-sm leading-relaxed font-medium prose prose-sm max-w-none">
                  <ReactMarkdown remarkPlugins={[remarkGfm]}>
                    {message.content}
                  </ReactMarkdown>
                </div>
                <div className={`text-xs mt-2 ${
                  message.role === 'user' ? 'text-blue-100' : 'text-gray-600'
                }`}>
                  {message.timestamp.toLocaleTimeString()}
                </div>
              </div>
            </div>
          ))}
          
          {isLoading && (
            <div className="flex items-start gap-4 group">
              {/* AI Avatar */}
              <div className="flex-shrink-0 w-10 h-10 rounded-full bg-gradient-to-br from-gray-100 to-gray-200 flex items-center justify-center shadow-lg">
                <span className="text-lg">🤖</span>
              </div>
              
              {/* Typing Indicator */}
              <div className="bg-white border border-gray-300 rounded-2xl px-6 py-4 shadow-lg">
                <div className="flex items-center space-x-3">
                  <div className="flex space-x-1">
                    <div className="w-2 h-2 bg-blue-400 rounded-full animate-bounce"></div>
                    <div className="w-2 h-2 bg-blue-400 rounded-full animate-bounce" style={{animationDelay: '0.1s'}}></div>
                    <div className="w-2 h-2 bg-blue-400 rounded-full animate-bounce" style={{animationDelay: '0.2s'}}></div>
                  </div>
                  <span className="text-sm text-gray-700 font-medium">
                    {language === 'en' ? 'AI is thinking...' : 'IA está pensando...'}
                  </span>
                </div>
              </div>
            </div>
          )}
        </div>
        
        <div ref={messagesEndRef} />
      </div>
      )}

      {/* Input */}
      {!showLinkSelection && (
      <div className="border-t border-gray-200/50 p-6 bg-white/80 backdrop-blur-xl relative z-10">
        <div className="max-w-4xl mx-auto">
          <div className="bg-white border border-gray-300 rounded-2xl p-4 shadow-lg">
            <div className="flex space-x-4">
              <textarea
                value={inputMessage}
                onChange={(e) => setInputMessage(e.target.value)}
                onKeyPress={handleKeyPress}
                placeholder={language === 'en' ? 
                  "Ask about your finances, investment opportunities, budget analysis..." :
                  "Pergunte sobre suas finanças, oportunidades de investimento, análise de orçamento..."
                }
                className="flex-1 bg-white text-gray-900 border-none px-0 py-2 text-sm placeholder-gray-500 focus:outline-none resize-none h-12 max-h-32"
                rows={1}
              />
              <button
                onClick={sendMessage}
                disabled={isLoading || !inputMessage.trim()}
                className="bg-blue-600 hover:bg-blue-700 disabled:bg-gray-400 text-white px-6 py-3 rounded-xl disabled:cursor-not-allowed transition-all duration-200 text-sm font-semibold flex-shrink-0 shadow-lg hover:shadow-xl"
              >
                {isLoading ? (
                  <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white"></div>
                ) : (
                  language === 'en' ? 'Send' : 'Enviar'
                )}
              </button>
            </div>
          </div>
        </div>
      </div>
      )}
    </div>
  )
}