# Deployment Guide for Expense Tracker Backend

## Overview
This guide helps you deploy the Expense Tracker Backend to Render.com.

## Prerequisites
- A Render.com account
- A PostgreSQL database (can be provisioned on Render)
- Environment variables configured

## Quick Deployment

### 1. Connect to Render
1. Go to [Render.com](https://render.com)
2. Click "New +" and select "Web Service"
3. Connect your GitHub repository

### 2. Configure the Service
- **Name**: `expense-tracker-backend`
- **Environment**: `Docker`
- **Region**: Choose closest to your users
- **Branch**: `main` (or your preferred branch)
- **Root Directory**: `backend` (if your backend is in a subdirectory)
- **Dockerfile Path**: `./Dockerfile`
- **Docker Context**: `.`

### 3. Environment Variables
Set the following environment variables in Render:

#### Required Variables:
```
PORT=8080
APP_ENV=production
JWT_SECRET=<generate-a-secure-random-string>
WEBHOOK_SECRET=<generate-a-secure-random-string>
```

#### Database Variables:
```
DB_DSN=postgres://username:password@host:port/database?sslmode=require
```

#### Optional Variables:
```
AA_BASE_URL=https://sandbox.example-aa.com
AA_PROVIDER=mock
BANK_VERIFICATION_ENABLED=false
```

### 4. Health Check
- **Health Check Path**: `/health`
- **Health Check Timeout**: 180 seconds

## Troubleshooting

### Port Binding Issues
If you see "Port scan timeout reached, no open ports detected":

1. **Check Environment Variables**: Ensure `PORT` is set to `8080`
2. **Check Logs**: Look for startup errors in the Render logs
3. **Database Connection**: Ensure the database connection string is correct
4. **Build Issues**: Check if the application builds successfully

### Common Issues

#### 1. Database Connection Failed
- Verify the `DB_DSN` environment variable
- Ensure the database is accessible from Render
- Check if SSL mode is required (`sslmode=require`)

#### 2. Application Won't Start
- Check the build logs for compilation errors
- Verify all required environment variables are set
- Look for missing dependencies

#### 3. Health Check Failing
- The application provides two health endpoints:
  - `/health` - Full health check
  - `/ping` - Simple connectivity check
- Check if the application is binding to the correct port

## Local Testing

Before deploying, test locally:

```bash
# Set environment variables
export PORT=8080
export APP_ENV=development
export DB_DSN="postgres://postgres:postgres@localhost:5432/expensedb?sslmode=disable"
export JWT_SECRET="test-secret"

# Build and run
go build -o main ./cmd/server
./main
```

## Monitoring

### Logs
- Check Render logs for startup messages
- Look for "Starting server" and "Application initialized successfully" messages
- Monitor for database connection retries

### Health Endpoints
- `GET /health` - Application health status
- `GET /ping` - Simple connectivity test
- `GET /api/health` - API health status

## Security Notes

1. **JWT Secret**: Generate a secure random string for production
2. **Database**: Use SSL connections in production
3. **Environment**: Set `APP_ENV=production` for production deployments
4. **Secrets**: Never commit secrets to version control

## Support

If you continue to have issues:
1. Check the Render documentation: https://render.com/docs
2. Review the application logs in Render dashboard
3. Test the application locally first
4. Verify all environment variables are correctly set 