import { NextRequest, NextResponse } from 'next/server';
import { AnalyticsInputSchema, AiInsightsSchema } from '../../../../lib/ai/schema';
import { generateInsights } from '../../../../lib/ai/client';
import { rateLimiter } from '../../../../lib/rateLimit';

export const dynamic = 'force-dynamic';

export async function POST(request: NextRequest) {
  // Check if AI insights are enabled
  if (process.env.AI_INSIGHTS_ENABLED !== 'true') {
    return NextResponse.json(
      { error: 'AI Insights feature is disabled' },
      { status: 503 }
    );
  }

  try {
    const body = await request.json();
    
    // Validate input
    const validatedInput = AnalyticsInputSchema.parse(body);
    
    // Get client IP for rate limiting
    const clientIp = request.headers.get('x-forwarded-for') || 
                    request.headers.get('x-real-ip') || 
                    'unknown';
    
    const userId = validatedInput.userId || clientIp;
    const month = validatedInput.month;
    
    // Check for fresh parameter
    const url = new URL(request.url);
    const fresh = url.searchParams.get('fresh') === '1';
    
    // Rate limiting
    const rateLimitKey = `ai:insights:${userId}`;
    
    // 1 call per 30 seconds
    if (!rateLimiter.limitPerWindow(rateLimitKey, 1, 30 * 1000)) {
      const remainingTime = rateLimiter.getRemainingTime(rateLimitKey);
      return NextResponse.json(
        { 
          error: 'Rate limit exceeded. Please wait before requesting new insights.',
          retryAfter: Math.ceil(remainingTime / 1000)
        },
        { 
          status: 429,
          headers: {
            'Retry-After': Math.ceil(remainingTime / 1000).toString()
          }
        }
      );
    }
    
    // 50 calls per day
    const dailyKey = `ai:insights:daily:${userId}`;
    if (!rateLimiter.limitPerWindow(dailyKey, 50, 24 * 60 * 60 * 1000)) {
      return NextResponse.json(
        { error: 'Daily rate limit exceeded. Please try again tomorrow.' },
        { status: 429 }
      );
    }
    
    // Build prompt for Hugging Face AI
    const prompt = `Analyze this financial data and provide 3 personalized insights:

Financial Summary:
- Month: ${validatedInput.month}
- Total Income: $${validatedInput.incomeTotal}
- Total Expenses: $${validatedInput.expenseTotal}
- Savings Rate: ${validatedInput.incomeTotal > 0 ? ((validatedInput.incomeTotal - validatedInput.expenseTotal) / validatedInput.incomeTotal * 100).toFixed(1) : 0}%

Category Breakdown:
${validatedInput.categories.map(cat => `- ${cat.name}: $${cat.amount} (${(cat.amount / validatedInput.expenseTotal * 100).toFixed(1)}%)`).join('\n')}

Weekly Trends:
${validatedInput.weeklySeries.map(week => `- ${week.week}: Income $${week.income}, Expenses $${week.expense}`).join('\n')}

Please provide 3 actionable financial insights with specific recommendations. Focus on:
1. Savings opportunities
2. Spending patterns
3. Financial health improvements

Format as JSON:
{
  "summary": {"month": "${validatedInput.month}", "income": ${validatedInput.incomeTotal}, "expenses": ${validatedInput.expenseTotal}, "savingsRate": ${validatedInput.incomeTotal > 0 ? (validatedInput.incomeTotal - validatedInput.expenseTotal) / validatedInput.incomeTotal : 0}},
  "topCategories": [{"name": "category", "share": 0.0}],
  "insights": [{"id": "1", "severity": "positive|warning|critical", "title": "Title", "message": "Analysis", "action": "Recommendation"}]
}`;

    // Generate insights using Hugging Face AI
    let aiResponse;
    try {
      aiResponse = await generateInsights(prompt, validatedInput);
    } catch (aiError) {
      console.error('AI generation error:', aiError);
      
      // Check for API errors and provide fallback
      if (aiError instanceof Error && (aiError.message.includes('quota') || aiError.message.includes('429') || aiError.message.includes('401'))) {
        console.log('API error, using fallback insights');
        const mockInsights = {
          summary: {
            month: validatedInput.month,
            income: validatedInput.incomeTotal,
            expenses: validatedInput.expenseTotal,
            savingsRate: validatedInput.incomeTotal > 0 ? Math.max(0, (validatedInput.incomeTotal - validatedInput.expenseTotal) / validatedInput.incomeTotal) : 0
          },
          topCategories: validatedInput.categories
            .sort((a, b) => b.amount - a.amount)
            .slice(0, 3)
            .map(cat => ({
              name: cat.name,
              share: cat.amount / validatedInput.expenseTotal
            })),
          insights: [
            {
              id: "api-fallback-1",
              severity: "warning",
              title: "AI Service Temporarily Unavailable",
              message: "Hugging Face API is not configured. Please add your API key to enable AI insights.",
              action: "Add HUGGINGFACE_API_KEY to your environment variables."
            },
            {
              id: "api-fallback-2",
              severity: "positive",
              title: "Data Analysis Complete",
              message: `Your monthly summary: $${validatedInput.incomeTotal} income, $${validatedInput.expenseTotal} expenses.`,
              action: "Consider reviewing your top spending categories for potential savings."
            }
          ],
          generatedAt: new Date().toISOString(),
          version: "v1"
        };
        
        return NextResponse.json(mockInsights);
      }
      
      // Re-throw other errors
      throw aiError;
    }
    
    // Parse and validate response
    let parsedResponse;
    try {
      parsedResponse = JSON.parse(aiResponse);
    } catch (parseError) {
      console.error('Failed to parse AI response:', parseError);
      return NextResponse.json(
        { error: 'AI temporarily unavailable' },
        { status: 503 }
      );
    }
    
    // Augment with metadata
    const insights = {
      ...parsedResponse,
      generatedAt: new Date().toISOString(),
      version: 'v1'
    };
    
    // Validate final response
    const validatedInsights = AiInsightsSchema.parse(insights);
    
    console.log(`Generated fresh insights for ${userId}:${month}`);
    
    return NextResponse.json(validatedInsights);
    
  } catch (error) {
    console.error('AI Insights API error:', error);
    
    if (error instanceof Error) {
      if (error.message.includes('validation')) {
        return NextResponse.json(
          { error: 'Invalid input data' },
          { status: 400 }
        );
      }
      
      // Check for specific API errors
      if (error.message.includes('Invalid API key')) {
        return NextResponse.json(
          { error: 'Invalid Hugging Face API key. Please check your configuration.' },
          { status: 500 }
        );
      }
      
      if (error.message.includes('rate limit')) {
        return NextResponse.json(
          { error: 'API rate limit exceeded. Please try again later.' },
          { status: 429 }
        );
      }
      
      if (error.message.includes('temporarily unavailable')) {
        return NextResponse.json(
          { error: 'AI service temporarily unavailable. Please try again.' },
          { status: 503 }
        );
      }
    }
    
    return NextResponse.json(
      { error: 'AI temporarily unavailable' },
      { status: 503 }
    );
  }
}