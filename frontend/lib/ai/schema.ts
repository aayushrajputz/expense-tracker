import { z } from 'zod';

// Input schema for analytics data
export const AnalyticsInputSchema = z.object({
  month: z.string().regex(/^\d{4}-\d{2}$/, 'Month must be in YYYY-MM format'),
  incomeTotal: z.number().min(0, 'Income total must be non-negative'),
  expenseTotal: z.number().min(0, 'Expense total must be non-negative'),
  categories: z.array(z.object({
    name: z.string().min(1, 'Category name is required'),
    amount: z.number().min(0, 'Category amount must be non-negative')
  })),
  weeklySeries: z.array(z.object({
    week: z.string(),
    income: z.number().min(0),
    expense: z.number().min(0)
  })),
  userId: z.union([z.string(), z.number()]).optional()
});

// AI Insights response schema
export const AiInsightsSchema = z.object({
  summary: z.object({
    month: z.string(),
    income: z.number().min(0),
    expenses: z.number().min(0),
    savingsRate: z.number().min(0).max(1, 'Savings rate must be between 0 and 1')
  }),
  topCategories: z.array(z.object({
    name: z.string().min(1),
    share: z.number().min(0).max(1, 'Share must be between 0 and 1')
  })),
  insights: z.array(z.object({
    id: z.string().min(1),
    severity: z.enum(['positive', 'warning', 'critical']),
    title: z.string().min(1),
    message: z.string().min(1),
    action: z.string().optional()
  })).max(3, 'Maximum 3 insights allowed'),
  generatedAt: z.string().datetime(),
  version: z.literal('v1')
});

// TypeScript types
export type AnalyticsInput = z.infer<typeof AnalyticsInputSchema>;
export type AiInsights = z.infer<typeof AiInsightsSchema>;