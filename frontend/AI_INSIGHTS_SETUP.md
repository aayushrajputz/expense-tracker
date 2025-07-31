# AI Insights Feature Setup

## Overview
This feature provides AI-powered financial insights using OpenAI's GPT model to analyze your expense data and provide personalized recommendations.

## Features
- 🤖 **AI-Powered Analysis**: Uses OpenAI GPT-3.5-turbo for intelligent financial insights
- 📊 **Monthly Aggregates**: Analyzes income, expenses, and category breakdowns
- 🎯 **Personalized Insights**: Provides actionable recommendations based on your spending patterns
- ⚡ **Smart Caching**: 6-hour cache to reduce API calls and improve performance
- 🛡️ **Rate Limiting**: Prevents abuse with 1 call/30s and 50 calls/day limits
- 🔒 **Security**: API keys never reach the client, all processing server-side

## Setup Instructions

### 1. Environment Variables
Create or update your `.env.local` file in the frontend directory:

```bash
# AI Insights Configuration
OPENAI_API_KEY=your_openai_api_key_here
AI_INSIGHTS_ENABLED=true
NEXT_PUBLIC_BASE_URL=http://localhost:3000
```

### 2. Install Dependencies
The required packages are already added to `package.json`. Install them:

```bash
npm install
```

### 3. Get OpenAI API Key
1. Visit [OpenAI Platform](https://platform.openai.com/)
2. Create an account or sign in
3. Navigate to API Keys section
4. Create a new API key
5. Copy the key and add it to your `.env.local` file

### 4. Enable the Feature
Set `AI_INSIGHTS_ENABLED=true` in your `.env.local` file to enable the feature.

## Usage

### Accessing AI Insights
1. Navigate to the AI Insights page via the sidebar menu
2. The system will automatically analyze your current month's data
3. View personalized insights and recommendations
4. Click "Refresh AI" to get fresh insights (respects rate limits)

### API Endpoints
- `POST /api/ai/insights` - Generate AI insights
- Query parameter `?fresh=1` - Bypass cache for fresh insights

### Rate Limits
- **Per User**: 1 request per 30 seconds
- **Daily**: 50 requests per day per user
- **Cache**: 6 hours TTL for repeated requests

## Architecture

### Files Created
1. `lib/ai/schema.ts` - Zod schemas and TypeScript types
2. `lib/ai/client.ts` - OpenAI client configuration
3. `lib/cache.ts` - In-memory TTL cache
4. `lib/rateLimit.ts` - Rate limiting utility
5. `app/api/ai/insights/route.ts` - API endpoint
6. `components/InsightsCards.tsx` - UI component
7. `app/ai/page.tsx` - AI insights page

### Data Flow
1. Frontend fetches monthly aggregates (currently mock data)
2. Sends data to `/api/ai/insights` endpoint
3. Server validates input, checks rate limits and cache
4. If cache miss, calls OpenAI API with structured prompt
5. Validates AI response and caches result
6. Returns structured insights to frontend
7. Frontend renders insights with severity-based styling

## Security Features
- ✅ API keys never exposed to client
- ✅ Input validation with Zod schemas
- ✅ Rate limiting per user/IP
- ✅ Structured JSON responses only
- ✅ Error handling with safe fallbacks
- ✅ Kill switch via environment variable

## Troubleshooting

### Common Issues

1. **"AI Insights feature is disabled"**
   - Check `AI_INSIGHTS_ENABLED=true` in `.env.local`

2. **"OPENAI_API_KEY environment variable is required"**
   - Add your OpenAI API key to `.env.local`

3. **"Rate limit exceeded"**
   - Wait 30 seconds between requests
   - Check daily limit (50 requests/day)

4. **"AI temporarily unavailable"**
   - Check OpenAI API key validity
   - Verify internet connection
   - Check OpenAI service status

### Debug Mode
Enable debug logging by checking browser console and server logs for detailed error messages.

## Customization

### Modifying Insights
Edit the prompt in `app/api/ai/insights/route.ts` to customize the type of insights generated.

### Styling
Modify `components/InsightsCards.tsx` to change the visual appearance of insight cards.

### Cache Duration
Adjust the TTL in `app/api/ai/insights/route.ts` (currently 6 hours).

### Rate Limits
Modify rate limiting parameters in `app/api/ai/insights/route.ts`.

## Production Considerations

1. **Environment Variables**: Use proper secret management in production
2. **Monitoring**: Add logging for API usage and errors
3. **Backup**: Implement fallback insights when AI is unavailable
4. **Scaling**: Consider Redis for distributed caching in production
5. **Costs**: Monitor OpenAI API usage and costs

## Support
For issues or questions about the AI Insights feature, check the console logs and ensure all environment variables are properly configured.