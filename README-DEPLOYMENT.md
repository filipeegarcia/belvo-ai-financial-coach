# 🚀 Deploy AI Financial Coach to filipegarcia.co/belvo

## Quick Start (5 minutes to production!)

### 1. Backend → Railway 
```bash
1. Go to railway.app → "New Project" → "Deploy from GitHub repo"
2. Select this repository
3. Set environment variables:
   - BELVO_SECRET_ID=your-secret-id
   - BELVO_SECRET_PASSWORD=your-secret-password  
   - OPENAI_API_KEY=your-openai-key
4. Custom domain: api.filipegarcia.co
```

### 2. Frontend → Vercel
```bash
1. Go to vercel.com → "New Project" → Import from GitHub
2. Root Directory: frontend
3. Set environment variables:
   - NEXT_PUBLIC_API_URL=https://api.filipegarcia.co
4. Custom domain: filipegarcia.co with path /belvo
```

### 3. Auto-Deploy Setup ✅
- Push to `main` branch = automatic deployment!
- GitHub Actions handles CI/CD
- No manual steps needed after initial setup

## 🌐 Final URLs
- **App:** https://filipegarcia.co/belvo
- **API:** https://api.filipegarcia.co/health

## 🔧 Environment Variables Checklist

**Backend (Railway/Render):**
- ✅ `BELVO_SECRET_ID` 
- ✅ `BELVO_SECRET_PASSWORD`
- ✅ `OPENAI_API_KEY`
- ✅ `PORT=8000`

**Frontend (Vercel):**
- ✅ `NEXT_PUBLIC_API_URL=https://api.filipegarcia.co`

That's it! Your AI Financial Coach will be live! 🎉
