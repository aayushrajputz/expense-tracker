import { NextResponse } from 'next/server';

export async function GET() {
  return NextResponse.json({
    environment: process.env.NODE_ENV,
    aiSystem: 'Free AI Financial Analysis',
    aiFeatures: [
      'Advanced savings rate analysis',
      'Category concentration detection', 
      'Expense ratio optimization',
      'Weekly trend analysis',
      'Personalized recommendations'
    ],
    noExternalAPI: true,
    aiInsightsEnabled: process.env.AI_INSIGHTS_ENABLED === 'true',
    nodeEnv: process.env.NODE_ENV,
    version: 'v2'
  });
}