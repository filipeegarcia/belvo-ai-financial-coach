# 🚀 Production Deployment Guide

## Overview
This guide will help you deploy the AI Financial Coach to production with automatic deployments from GitHub.

**Architecture:**
- **Frontend (Next.js)** → **Vercel** at `filipegarcia.co/belvo`
- **Backend (Go)** → **Railway/Render** at `api.filipegarcia.co`
- **Auto-deploy** → On every push to `main` branch

---

## 🎯 Quick Setup (5 minutes)

### 1. Frontend Deployment (Vercel)

1. **Connect Repository:**
   - Go to [Vercel Dashboard](https://vercel.com/dashboard)
   - Click "New Project" → Import from GitHub
   - Select this repository

2. **Configure Build Settings:**
   ```bash
   Build Command: cd frontend && npm run build
   Output Directory: frontend/.next
   Install Command: cd frontend && npm install
   ```

3. **Set Environment Variables:**
   ```bash
   NEXT_PUBLIC_API_URL=https://api.filipegarcia.co
   NEXT_PUBLIC_ENVIRONMENT=production
   ```

4. **Configure Domain:**
   - In Project Settings → Domains
   - Add custom domain: `filipegarcia.co`
   - Set path prefix: `/belvo`

### 2. Backend Deployment (Choose One)

#### Option A: Railway (Recommended)

1. **Connect Repository:**
   - Go to [Railway](https://railway.app)
   - Create new project from GitHub repo
   - Select this repository

2. **Configure Service:**
   - Service name: `belvo-api`
   - Build command: `go build -o bin/api cmd/api/main.go`
   - Start command: `./bin/api`

3. **Set Environment Variables:**
   ```bash
   PORT=8000
   BELVO_ENVIRONMENT=sandbox
   BELVO_SECRET_ID=your-belvo-secret-id
   BELVO_SECRET_PASSWORD=your-belvo-secret-password
   OPENAI_API_KEY=your-openai-api-key
   ENABLE_METRICS=false
   ```

4. **Configure Domain:**
   - In Railway dashboard → Settings → Domains
   - Add custom domain: `api.filipegarcia.co`

#### Option B: Render

1. **Connect Repository:**
   - Go to [Render](https://render.com)
   - Create new Web Service from GitHub repo

2. **Use render.yaml:**
   - Render will automatically detect `render.yaml`
   - Set environment variables in dashboard

---

## 🔧 Environment Variables Reference

### Backend Variables
```bash
# Required
PORT=8000
BELVO_SECRET_ID=your-belvo-secret-id
BELVO_SECRET_PASSWORD=your-belvo-secret-password
OPENAI_API_KEY=your-openai-api-key

# Optional
BELVO_ENVIRONMENT=sandbox
ENABLE_METRICS=false
GOFR_TELEMETRY=false
```

### Frontend Variables (Vercel)
```bash
NEXT_PUBLIC_API_URL=https://api.filipegarcia.co
NEXT_PUBLIC_ENVIRONMENT=production
```

---

## 🤖 GitHub Actions Setup

The repository includes automatic CI/CD:

1. **Frontend:** Vercel auto-deploys on push to `main`
2. **Backend:** Railway/Render auto-deploys via GitHub Actions

### GitHub Secrets (if using GitHub Actions)
```bash
RAILWAY_TOKEN=your-railway-token
# OR
RENDER_API_KEY=your-render-api-key
RENDER_SERVICE_ID=your-render-service-id
```

---

## 🌐 Domain Configuration

### DNS Settings
Point these DNS records to your hosting providers:

```bash
# A Record
filipegarcia.co → [Vercel IP]

# CNAME Records  
www.filipegarcia.co → filipegarcia.co
api.filipegarcia.co → [Railway/Render URL]
```

### URL Structure
- **Frontend:** `https://filipegarcia.co/belvo`
- **Backend API:** `https://api.filipegarcia.co`
- **Health Check:** `https://api.filipegarcia.co/health`

---

## ✅ Testing Production Deployment

### 1. Test Backend
```bash
curl https://api.filipegarcia.co/health
# Should return: {"data":{"status":"healthy",...}}
```

### 2. Test Frontend
- Visit: `https://filipegarcia.co/belvo`
- Should load the authentication page

### 3. Test End-to-End
- Complete authentication flow
- Verify AI chat works with real Belvo data

---

## 🔍 Troubleshooting

### Common Issues

**CORS Errors:**
- Check `allowedOrigins` in `internal/app/app.go`
- Ensure `https://filipegarcia.co` is included

**Build Failures:**
- Check environment variables are set
- Verify Go version compatibility (1.21+)

**API Connection Issues:**
- Verify `NEXT_PUBLIC_API_URL` matches backend domain
- Check backend health endpoint

**Belvo API Issues:**
- Verify `BELVO_SECRET_ID` and `BELVO_SECRET_PASSWORD`
- Check Belvo sandbox status

---

## 🚀 Deployment Commands

### Manual Deploy Backend (if needed)
```bash
# Railway
railway deploy

# Render
curl -X POST https://api.render.com/v1/services/YOUR_SERVICE_ID/deploys \
  -H "Authorization: Bearer YOUR_API_KEY"
```

### Manual Deploy Frontend
```bash
# Vercel
vercel --prod
```

---

## 📊 Monitoring

### Health Checks
- **Backend:** `https://api.filipegarcia.co/health`
- **Frontend:** `https://filipegarcia.co/belvo`

### Logs
- **Railway:** Dashboard → Service → Logs
- **Render:** Dashboard → Service → Logs  
- **Vercel:** Dashboard → Project → Functions

---

## 🔄 Auto-Deployment Workflow

1. **Push to `main` branch**
2. **GitHub Actions triggers:**
   - Frontend builds and deploys to Vercel
   - Backend builds and deploys to Railway/Render
3. **Vercel handles:**
   - Domain routing (`/belvo` path)
   - SSL certificates
   - CDN distribution
4. **Railway/Render handles:**
   - API service hosting
   - Health checks
   - Auto-scaling

**Your app will be live at `https://filipegarcia.co/belvo` within 2-3 minutes of pushing to main!** 🎉
