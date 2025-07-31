# CI/CD Setup Guide for Expense Tracker

This guide will help you set up Continuous Integration and Continuous Deployment for your expense tracker application.

## Overview

The CI/CD pipeline includes:
- **Frontend CI/CD**: Next.js application testing, building, and deployment
- **Backend CI/CD**: Go API testing, building, and deployment
- **Full Stack Deploy**: Coordinated deployment of both frontend and backend

## Prerequisites

1. **GitHub Repository**: Your code should be in a GitHub repository
2. **Hosting Platforms**: Choose your deployment platforms:
   - **Frontend**: Vercel, Netlify, or AWS S3 + CloudFront
   - **Backend**: Railway, Heroku, AWS ECS, or Google Cloud Run
3. **Database**: PostgreSQL database (managed service recommended)

## Step 1: Set Up GitHub Secrets

Go to your GitHub repository → Settings → Secrets and variables → Actions, and add the following secrets:

### For Frontend Deployment (Vercel)
```
VERCEL_TOKEN=your_vercel_token
VERCEL_ORG_ID=your_vercel_org_id
VERCEL_PROJECT_ID=your_vercel_project_id
```

### For Backend Deployment (Railway)
```
RAILWAY_TOKEN=your_railway_token
```

### For Environment Variables
```
API_URL=https://your-backend-url.com
DATABASE_URL=postgresql://username:password@host:port/database
JWT_SECRET=your_jwt_secret
SMTP_HOST=your_smtp_host
SMTP_PORT=587
SMTP_USER=your_smtp_user
SMTP_PASS=your_smtp_password
```

## Step 2: Configure Frontend Deployment

### Option A: Vercel (Recommended for Next.js)

1. **Install Vercel CLI**:
   ```bash
   npm i -g vercel
   ```

2. **Login to Vercel**:
   ```bash
   vercel login
   ```

3. **Deploy from frontend directory**:
   ```bash
   cd frontend
   vercel --prod
   ```

4. **Get deployment tokens**:
   - Go to Vercel Dashboard → Settings → Tokens
   - Create a new token
   - Add to GitHub secrets as `VERCEL_TOKEN`

### Option B: Netlify

1. **Build command**: `npm run build`
2. **Publish directory**: `.next`
3. **Environment variables**: Add all required environment variables

## Step 3: Configure Backend Deployment

### Option A: Railway (Recommended for Go apps)

1. **Connect GitHub repository** to Railway
2. **Set build command**: `go build -o expense-tracker-api ./cmd/server`
3. **Set start command**: `./expense-tracker-api`
4. **Add environment variables** in Railway dashboard

### Option B: Heroku

1. **Create Heroku app**:
   ```bash
   heroku create your-app-name
   ```

2. **Set buildpacks**:
   ```bash
   heroku buildpacks:set heroku/go
   ```

3. **Deploy**:
   ```bash
   git push heroku main
   ```

### Option C: AWS ECS

1. **Create ECR repository** for Docker images
2. **Set up ECS cluster and service**
3. **Configure AWS credentials** in GitHub secrets

## Step 4: Database Setup

### Option A: Railway PostgreSQL
1. Create PostgreSQL service in Railway
2. Get connection string from Railway dashboard
3. Add to environment variables

### Option B: AWS RDS
1. Create PostgreSQL RDS instance
2. Configure security groups
3. Get connection string and add to secrets

### Option C: Supabase
1. Create Supabase project
2. Get connection string from Settings → Database
3. Add to environment variables

## Step 5: Environment Configuration

### Frontend Environment Variables
Create `frontend/.env.local`:
```env
NEXT_PUBLIC_API_URL=https://your-backend-url.com
NEXT_PUBLIC_APP_NAME=Expense Tracker
```

### Backend Environment Variables
Create `backend/.env`:
```env
DATABASE_URL=postgresql://username:password@host:port/database
JWT_SECRET=your_jwt_secret
SMTP_HOST=your_smtp_host
SMTP_PORT=587
SMTP_USER=your_smtp_user
SMTP_PASS=your_smtp_password
```

## Step 6: Testing Setup

### Frontend Testing
Add to `frontend/package.json`:
```json
{
  "scripts": {
    "test": "jest",
    "test:watch": "jest --watch",
    "test:coverage": "jest --coverage"
  },
  "devDependencies": {
    "@testing-library/jest-dom": "^5.16.4",
    "@testing-library/react": "^13.3.0",
    "jest": "^28.1.0",
    "jest-environment-jsdom": "^28.1.0"
  }
}
```

### Backend Testing
Add to `backend/go.mod`:
```go
require (
    github.com/stretchr/testify v1.8.4
)
```

## Step 7: Monitoring and Logging

### Frontend Monitoring
- **Vercel Analytics**: Built-in with Vercel
- **Sentry**: Add Sentry SDK for error tracking

### Backend Monitoring
- **Railway Logs**: Built-in logging
- **Sentry**: Add Sentry SDK for Go

## Step 8: Security

### Environment Variables
- Never commit `.env` files
- Use GitHub secrets for sensitive data
- Rotate secrets regularly

### API Security
- Use HTTPS in production
- Implement rate limiting
- Add CORS configuration
- Use JWT tokens for authentication

## Troubleshooting

### Common Issues

1. **Build Failures**:
   - Check Node.js/Go version compatibility
   - Verify all dependencies are installed
   - Check for syntax errors

2. **Deployment Failures**:
   - Verify environment variables are set
   - Check database connectivity
   - Review deployment logs

3. **Database Connection Issues**:
   - Verify connection string format
   - Check network security groups
   - Ensure database is accessible

### Debug Commands

```bash
# Check frontend build locally
cd frontend
npm run build

# Check backend build locally
cd backend
go build ./cmd/server

# Test database connection
go run ./cmd/server --test-db
```

## Next Steps

1. **Set up staging environment** for testing
2. **Configure branch protection rules**
3. **Add automated testing** for pull requests
4. **Set up monitoring and alerting**
5. **Implement blue-green deployments**

## Support

For issues with:
- **GitHub Actions**: Check GitHub Actions documentation
- **Vercel**: Check Vercel documentation
- **Railway**: Check Railway documentation
- **Go**: Check Go documentation 