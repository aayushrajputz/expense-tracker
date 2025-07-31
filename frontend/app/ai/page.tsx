'use client';

import { useState, useEffect } from 'react';
import { useAuth } from '../../lib/AuthContext';
import ProtectedRoute from '../../components/ProtectedRoute';
import Sidebar from '../../components/Sidebar';
import Navbar from '../../components/Navbar';
import InsightsCards from '../../components/InsightsCards';
import LoadingSpinner from '../../components/LoadingSpinner';
import ErrorMessage from '../../components/ErrorMessage';
import { AiInsights } from '../../lib/ai/schema';
import { expenseAPI } from '../../lib/api';

export default function AIPage() {
  const { user } = useAuth();
  const [insights, setInsights] = useState<AiInsights | null>(null);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [sidebarOpen, setSidebarOpen] = useState(false);

  // Get current month in YYYY-MM format
  const getCurrentMonth = () => {
    const now = new Date();
    return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`;
  };

  // Fetch real monthly aggregates from backend
  const fetchMonthlyAggregates = async () => {
    try {
      const currentMonth = getCurrentMonth();
      const startDate = `${currentMonth}-01`;
      const endDate = `${currentMonth}-31`;
      
      // Get all expenses for the current month
      const expenses = await expenseAPI.getExpensesByDateRange(startDate, endDate);
      
      // Calculate totals and categories
      const incomeTotal = expenses
        .filter(exp => exp.type === 'income')
        .reduce((sum, exp) => sum + exp.amount, 0);
      
      const expenseTotal = expenses
        .filter(exp => exp.type === 'expense')
        .reduce((sum, exp) => sum + exp.amount, 0);
      
      // Group by categories
      const categoryMap = new Map();
      expenses
        .filter(exp => exp.type === 'expense')
        .forEach(exp => {
          const category = exp.category || 'Other';
          categoryMap.set(category, (categoryMap.get(category) || 0) + exp.amount);
        });
      
      // Convert to array format
      const categories = Array.from(categoryMap.entries()).map(([name, amount]) => ({
        name,
        amount
      }));
      
      // Generate weekly series (simplified - you can enhance this)
      const weeklySeries = [
        { week: 'Week 1', income: incomeTotal / 4, expense: expenseTotal / 4 },
        { week: 'Week 2', income: incomeTotal / 4, expense: expenseTotal / 4 },
        { week: 'Week 3', income: incomeTotal / 4, expense: expenseTotal / 4 },
        { week: 'Week 4', income: incomeTotal / 4, expense: expenseTotal / 4 }
      ];
      
      const aggregates = {
        month: currentMonth,
        incomeTotal,
        expenseTotal,
        categories,
        weeklySeries,
        userId: user?.id || user?.email
      };
      
      console.log('Real financial data:', aggregates);
      return aggregates;
      
    } catch (error) {
      console.error('Error fetching real financial data:', error);
      // Fallback to mock data if backend is unavailable
      const currentMonth = getCurrentMonth();
      return {
        month: currentMonth,
        incomeTotal: 0,
        expenseTotal: 0,
        categories: [],
        weeklySeries: [],
        userId: user?.id || user?.email
      };
    }
  };

  const fetchInsights = async (fresh = false) => {
    try {
      setError(null);
      const aggregates = await fetchMonthlyAggregates();
      
      // Only proceed if we have real data
      if (aggregates.incomeTotal === 0 && aggregates.expenseTotal === 0) {
        setError('No financial data available for AI analysis. Please add some transactions first.');
        setLoading(false);
        setRefreshing(false);
        return;
      }
      
      const url = fresh ? '/api/ai/insights?fresh=1' : '/api/ai/insights';
      
      const response = await fetch(url, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(aggregates),
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || 'Failed to fetch insights');
      }

      const data = await response.json();
      setInsights(data);
    } catch (err) {
      console.error('Error fetching insights:', err);
      setError(err instanceof Error ? err.message : 'Failed to fetch insights');
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  };

  const handleRefresh = async () => {
    setRefreshing(true);
    await fetchInsights(true);
  };

  useEffect(() => {
    fetchInsights();
  }, []);

  if (loading) {
    return (
      <ProtectedRoute>
              <div className="min-h-screen bg-gray-900">
        <Sidebar isOpen={sidebarOpen} onClose={() => setSidebarOpen(false)} />
        <div className="lg:ml-64">
          <Navbar onSidebarToggle={() => setSidebarOpen(!sidebarOpen)} />
          <div className="p-6">
            <div className="flex items-center justify-center h-64">
              <LoadingSpinner />
            </div>
          </div>
        </div>
      </div>
      </ProtectedRoute>
    );
  }

  return (
    <ProtectedRoute>
      <div className="min-h-screen bg-gray-900">
        <Sidebar isOpen={sidebarOpen} onClose={() => setSidebarOpen(false)} />
        <div className="lg:ml-64">
          <Navbar onSidebarToggle={() => setSidebarOpen(!sidebarOpen)} />
          <div className="p-6">
            <div className="flex justify-between items-center mb-6">
              <div>
                <h1 className="text-3xl font-bold text-transparent bg-clip-text bg-gradient-to-r from-yellow-400 to-green-400">
                  AI Financial Insights
                </h1>
                <p className="text-gray-400 mt-2">
                  Last updated: {insights?.generatedAt ? new Date(insights.generatedAt).toLocaleString() : 'Never'}
                </p>
              </div>
              <button
                onClick={handleRefresh}
                disabled={refreshing}
                className="bg-yellow-500 hover:bg-yellow-600 disabled:bg-gray-600 text-black font-semibold py-2 px-4 rounded-lg flex items-center space-x-2 transition-colors"
              >
                <svg
                  className={`w-5 h-5 ${refreshing ? 'animate-spin' : ''}`}
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
                  />
                </svg>
                <span>Refresh AI</span>
              </button>
            </div>

            {error && <ErrorMessage message={error} />}

            {insights && <InsightsCards insights={insights} />}
          </div>
        </div>
      </div>
    </ProtectedRoute>
  );
}