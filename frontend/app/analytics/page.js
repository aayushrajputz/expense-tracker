
'use client';

import { useState, useEffect, useRef } from 'react';
import Sidebar from '../../components/Sidebar';
import Navbar from '../../components/Navbar';
import InsightBox from '../../components/InsightBox';
import Chart from '../../components/Chart';
import LoadingSpinner from '../../components/LoadingSpinner';
import ErrorMessage from '../../components/ErrorMessage';
import { expenseAPI } from '../../lib/api';
import ProtectedRoute from '../../components/ProtectedRoute';
import { useSettings } from '../../lib/SettingsContext';

export default function Analytics() {
  const { currency } = useSettings();
  const [insights, setInsights] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [expenses, setExpenses] = useState([]);
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const [isRealTimeEnabled, setIsRealTimeEnabled] = useState(true);
  const [lastUpdate, setLastUpdate] = useState(null);
  const [aiProcessing, setAiProcessing] = useState(false);
  
  // Refs for real-time functionality
  const pollingIntervalRef = useRef(null);
  const lastActivityRef = useRef(Date.now());
  const isPageVisibleRef = useRef(true);

  // AI-powered insight generation
  const generateAIInsights = (expenseData) => {
    if (!expenseData || expenseData.length === 0) {
      return [];
    }

    const insights = [];
    const now = new Date();
    const currentMonth = now.getMonth();
    const currentYear = now.getFullYear();
    
    // Calculate monthly totals
    const monthlyData = {};
    expenseData.forEach(expense => {
      const date = new Date(expense.date);
      const monthKey = `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}`;
      
      if (!monthlyData[monthKey]) {
        monthlyData[monthKey] = { income: 0, expenses: 0, transactions: 0 };
      }
      
      if (expense.type === 'income') {
        monthlyData[monthKey].income += expense.amount;
      } else {
        monthlyData[monthKey].expenses += expense.amount;
      }
      monthlyData[monthKey].transactions++;
    });

    // Get current and previous month data
    const currentMonthKey = `${currentYear}-${String(currentMonth + 1).padStart(2, '0')}`;
    const prevMonthKey = `${currentMonth === 0 ? currentYear - 1 : currentYear}-${String(currentMonth === 0 ? 12 : currentMonth).padStart(2, '0')}`;
    
    const currentMonthData = monthlyData[currentMonthKey] || { income: 0, expenses: 0, transactions: 0 };
    const prevMonthData = monthlyData[prevMonthKey] || { income: 0, expenses: 0, transactions: 0 };

    // Calculate category spending
    const categorySpending = {};
    expenseData.filter(e => e.type === 'expense').forEach(expense => {
      categorySpending[expense.category] = (categorySpending[expense.category] || 0) + expense.amount;
    });

    // 1. Spending Trend Analysis
    if (prevMonthData.expenses > 0) {
      const spendingChange = ((currentMonthData.expenses - prevMonthData.expenses) / prevMonthData.expenses) * 100;
      const trendType = spendingChange > 10 ? 'warning' : spendingChange > 0 ? 'info' : 'success';
      
      insights.push({
        type: trendType,
        title: 'Spending Trend Analysis',
        description: `Your spending has ${spendingChange > 0 ? 'increased' : 'decreased'} by ${Math.abs(spendingChange).toFixed(1)}% this month compared to last month.`,
        suggestion: spendingChange > 10 ? 'Consider reviewing your budget to control spending growth.' : 
                   spendingChange > 0 ? 'Your spending is growing moderately. Keep monitoring.' : 
                   'Great job! You\'ve reduced your spending this month.'
      });
    }

    // 2. Budget Recommendation
    const totalIncome = expenseData.filter(e => e.type === 'income').reduce((sum, e) => sum + e.amount, 0);
    const totalExpenses = expenseData.filter(e => e.type === 'expense').reduce((sum, e) => sum + e.amount, 0);
    const savingsRate = totalIncome > 0 ? ((totalIncome - totalExpenses) / totalIncome) * 100 : 0;
    
    if (savingsRate < 20) {
      insights.push({
        type: 'warning',
        title: 'Savings Rate Alert',
        description: `Your current savings rate is ${savingsRate.toFixed(1)}%, which is below the recommended 20%.`,
        suggestion: 'Consider reducing discretionary spending or increasing your income to improve your savings rate.'
      });
    } else {
      insights.push({
        type: 'success',
        title: 'Excellent Savings Rate',
        description: `Your savings rate of ${savingsRate.toFixed(1)}% is excellent!`,
        suggestion: 'Keep up the great work! Consider investing your savings for long-term growth.'
      });
    }

    // 3. Category Spending Analysis
    const topCategory = Object.entries(categorySpending).sort(([,a], [,b]) => b - a)[0];
    if (topCategory) {
      const categoryPercentage = (topCategory[1] / totalExpenses) * 100;
      
      if (categoryPercentage > 40) {
        insights.push({
          type: 'warning',
          title: 'High Category Concentration',
          description: `${topCategory[0]} accounts for ${categoryPercentage.toFixed(1)}% of your total spending.`,
          suggestion: 'Consider diversifying your spending across different categories to reduce risk.'
        });
      } else {
        insights.push({
          type: 'info',
          title: 'Balanced Spending',
          description: `Your spending is well-distributed across categories. ${topCategory[0]} is your highest at ${categoryPercentage.toFixed(1)}%.`,
          suggestion: 'Maintain this balanced approach to your spending habits.'
        });
      }
    }

    // 4. Monthly Forecast
    const avgDailySpending = currentMonthData.expenses / Math.min(now.getDate(), 30);
    const projectedMonthlySpending = avgDailySpending * 30;
    
    insights.push({
      type: 'prediction',
      title: 'Monthly Spending Forecast',
      description: `Based on your current spending patterns, you're projected to spend ${projectedMonthlySpending.toLocaleString(undefined, { style: 'currency', currency })} this month.`,
      suggestion: projectedMonthlySpending > totalIncome * 0.8 ? 
                 'Your projected spending is high. Consider reducing expenses to stay within budget.' :
                 'Your spending is on track for this month.'
    });

    // 5. Transaction Frequency Analysis
    const avgTransactionsPerDay = currentMonthData.transactions / Math.min(now.getDate(), 30);
    
    if (avgTransactionsPerDay > 3) {
      insights.push({
        type: 'tip',
        title: 'High Transaction Frequency',
        description: `You're making an average of ${avgTransactionsPerDay.toFixed(1)} transactions per day this month.`,
        suggestion: 'Consider consolidating small purchases to reduce transaction fees and better track your spending.'
      });
    }

    // 6. Income vs Expenses Analysis
    if (currentMonthData.income > 0) {
      const expenseToIncomeRatio = (currentMonthData.expenses / currentMonthData.income) * 100;
      
      if (expenseToIncomeRatio > 90) {
        insights.push({
          type: 'warning',
          title: 'High Expense-to-Income Ratio',
          description: `Your expenses are ${expenseToIncomeRatio.toFixed(1)}% of your income this month.`,
          suggestion: 'This ratio is quite high. Focus on reducing expenses or increasing income to improve your financial health.'
        });
      } else if (expenseToIncomeRatio < 60) {
        insights.push({
          type: 'success',
          title: 'Excellent Financial Health',
          description: `Your expenses are only ${expenseToIncomeRatio.toFixed(1)}% of your income this month.`,
          suggestion: 'Excellent! You have a healthy balance between income and expenses.'
        });
      }
    }

    // 7. Seasonal Spending Pattern
    const recentExpenses = expenseData.filter(e => e.type === 'expense' && new Date(e.date) >= new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000));
    const weeklySpending = recentExpenses.reduce((sum, e) => sum + e.amount, 0);
    const avgWeeklySpending = currentMonthData.expenses / 4;
    
    if (weeklySpending > avgWeeklySpending * 1.5) {
      insights.push({
        type: 'warning',
        title: 'High Weekly Spending',
        description: `You've spent ${weeklySpending.toLocaleString(undefined, { style: 'currency', currency })} this week, which is higher than your average.`,
        suggestion: 'Review your recent purchases and consider if they align with your financial goals.'
      });
    }

    // 11. Category Budget Recommendations
    const topCategories = Object.entries(categorySpending).sort(([,a], [,b]) => b - a).slice(0, 3);
    if (topCategories.length > 0) {
      const topCategory = topCategories[0];
      const categoryBudget = totalExpenses * 0.3; // 30% rule for top category
      
      if (topCategory[1] > categoryBudget) {
        insights.push({
          type: 'tip',
          title: 'Category Budget Alert',
          description: `${topCategory[0]} spending is ${((topCategory[1] / totalExpenses) * 100).toFixed(1)}% of your total expenses.`,
          suggestion: `Consider setting a monthly budget of ${categoryBudget.toLocaleString(undefined, { style: 'currency', currency })} for ${topCategory[0]} to better control spending.`
        });
      }
    }

    // 12. Savings Opportunity Analysis
    const discretionaryCategories = ['Entertainment', 'Shopping', 'Food'];
    const discretionarySpending = expenseData
      .filter(e => e.type === 'expense' && discretionaryCategories.includes(e.category))
      .reduce((sum, e) => sum + e.amount, 0);
    
    if (discretionarySpending > totalExpenses * 0.4) {
      const potentialSavings = discretionarySpending * 0.2; // 20% reduction potential
      insights.push({
        type: 'tip',
        title: 'Savings Opportunity',
        description: `Discretionary spending accounts for ${((discretionarySpending / totalExpenses) * 100).toFixed(1)}% of your expenses.`,
        suggestion: `You could save ${potentialSavings.toLocaleString(undefined, { style: 'currency', currency })} monthly by reducing discretionary spending by 20%.`
      });
    }

    return insights.slice(0, 6); // Return top 6 insights
  };

  // Real-time data fetching
  const fetchRealTimeData = async (showLoading = false) => {
    try {
      if (showLoading) {
        setLoading(true);
      }
      setError(null);
      setAiProcessing(true);

      const [analyticsData, expensesData, aiInsightsData] = await Promise.all([
        expenseAPI.getAnalytics(),
        expenseAPI.getExpenses(),
        expenseAPI.getAIInsights()
      ]);

      const processedExpenses = Array.isArray(expensesData) ? expensesData : [];
      setExpenses(processedExpenses);

      // Use real AI insights from OpenAI GPT-3.5
      if (aiInsightsData && aiInsightsData.insights) {
        setInsights(aiInsightsData.insights);
        setLastUpdate(new Date());
      } else {
        // Fallback to rule-based insights if AI fails
        const ruleBasedInsights = generateAIInsights(processedExpenses);
        setInsights(ruleBasedInsights);
      }

    } catch (err) {
      console.error('Error fetching analytics data:', err);
      setError('Failed to fetch analytics data');
      
      // Generate insights from mock data when API is not available
      const mockExpenses = [
        {
          id: 1,
          title: 'Grocery Shopping',
          amount: 150.50,
          type: 'expense',
          category: 'Food',
          date: '2024-01-15',
          description: 'Weekly groceries'
        },
        {
          id: 2,
          title: 'Salary',
          amount: 3000.00,
          type: 'income',
          category: 'Income',
          date: '2024-01-01',
          description: 'Monthly salary'
        },
        {
          id: 3,
          title: 'Gas Station',
          amount: 45.00,
          type: 'expense',
          category: 'Transportation',
          date: '2024-01-14',
          description: 'Fuel for car'
        },
        {
          id: 4,
          title: 'Restaurant',
          amount: 85.00,
          type: 'expense',
          category: 'Food',
          date: '2024-01-13',
          description: 'Dinner with friends'
        },
        {
          id: 5,
          title: 'Freelance Work',
          amount: 500.00,
          type: 'income',
          category: 'Income',
          date: '2024-01-10',
          description: 'Web development project'
        },
        {
          id: 6,
          title: 'Movie Tickets',
          amount: 25.00,
          type: 'expense',
          category: 'Entertainment',
          date: '2024-01-12',
          description: 'Weekend movie'
        },
        {
          id: 7,
          title: 'Electric Bill',
          amount: 120.00,
          type: 'expense',
          category: 'Bills',
          date: '2024-01-05',
          description: 'Monthly electricity'
        },
        {
          id: 8,
          title: 'Online Shopping',
          amount: 75.00,
          type: 'expense',
          category: 'Shopping',
          date: '2024-01-08',
          description: 'New clothes'
        }
      ];
      
      setExpenses(mockExpenses);
      const mockInsights = generateAIInsights(mockExpenses);
      setInsights(mockInsights);
    } finally {
      if (showLoading) {
        setLoading(false);
      }
      setAiProcessing(false);
    }
  };

  // Start real-time polling
  const startRealTimePolling = () => {
    if (pollingIntervalRef.current) {
      clearInterval(pollingIntervalRef.current);
    }
    
    if (isRealTimeEnabled) {
      pollingIntervalRef.current = setInterval(() => {
        if (isPageVisibleRef.current && (Date.now() - lastActivityRef.current) < 300000) {
          fetchRealTimeData(false);
        }
      }, 30000); // Poll every 30 seconds
    }
  };

  // Stop real-time polling
  const stopRealTimePolling = () => {
    if (pollingIntervalRef.current) {
      clearInterval(pollingIntervalRef.current);
      pollingIntervalRef.current = null;
    }
  };

  // Handle page visibility changes
  useEffect(() => {
    const handleVisibilityChange = () => {
      isPageVisibleRef.current = !document.hidden;
      
      if (isPageVisibleRef.current && isRealTimeEnabled) {
        fetchRealTimeData(false);
        startRealTimePolling();
      } else {
        stopRealTimePolling();
      }
    };

    const handleUserActivity = () => {
      lastActivityRef.current = Date.now();
    };

    const handleFocus = () => {
      isPageVisibleRef.current = true;
      if (isRealTimeEnabled) {
        fetchRealTimeData(false);
        startRealTimePolling();
      }
    };

    const handleBlur = () => {
      isPageVisibleRef.current = false;
      stopRealTimePolling();
    };

    document.addEventListener('visibilitychange', handleVisibilityChange);
    document.addEventListener('mousemove', handleUserActivity);
    document.addEventListener('keydown', handleUserActivity);
    document.addEventListener('click', handleUserActivity);
    window.addEventListener('focus', handleFocus);
    window.addEventListener('blur', handleBlur);

    return () => {
      document.removeEventListener('visibilitychange', handleVisibilityChange);
      document.removeEventListener('mousemove', handleUserActivity);
      document.removeEventListener('keydown', handleUserActivity);
      document.removeEventListener('click', handleUserActivity);
      window.removeEventListener('focus', handleFocus);
      window.removeEventListener('blur', handleBlur);
    };
  }, [isRealTimeEnabled]);

  // Initial data fetch and real-time setup
  useEffect(() => {
    fetchRealTimeData(true);
    if (isRealTimeEnabled) {
      startRealTimePolling();
    }

    return () => {
      stopRealTimePolling();
    };
  }, [isRealTimeEnabled]);

  // Toggle real-time updates
  const toggleRealTime = () => {
    setIsRealTimeEnabled(!isRealTimeEnabled);
    if (!isRealTimeEnabled) {
      startRealTimePolling();
    } else {
      stopRealTimePolling();
    }
  };

  // Format last update time
  const formatLastUpdate = () => {
    if (!lastUpdate) return 'Never';
    const now = new Date();
    const diffInSeconds = Math.floor((now.getTime() - lastUpdate.getTime()) / 1000);
    
    if (diffInSeconds < 60) return `${diffInSeconds}s ago`;
    if (diffInSeconds < 3600) return `${Math.floor(diffInSeconds / 60)}m ago`;
    return `${Math.floor(diffInSeconds / 3600)}h ago`;
  };

  const getMonthlyTrends = () => {
    if (!expenses.length) return [];

    const monthlyData = {};
    expenses.forEach(expense => {
      const date = new Date(expense.date);
      const monthKey = `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}`;

      if (!monthlyData[monthKey]) {
        monthlyData[monthKey] = { income: 0, expenses: 0 };
      }

      if (expense.type === 'income') {
        monthlyData[monthKey].income += expense.amount;
      } else {
        monthlyData[monthKey].expenses += expense.amount;
      }
    });

    return Object.entries(monthlyData)
      .sort(([a], [b]) => a.localeCompare(b))
      .map(([month, data]) => ({
        name: new Date(month + '-01').toLocaleDateString('en-US', { month: 'short', year: '2-digit' }),
        income: data.income,
        expenses: data.expenses,
        savings: data.income - data.expenses
      }));
  };

  const getCategoryComparison = () => {
    if (!expenses.length) return [];

    const categoryTotals = {};
    expenses.filter(e => e.type === 'expense').forEach(expense => {
      categoryTotals[expense.category] = (categoryTotals[expense.category] || 0) + expense.amount;
    });

    return Object.entries(categoryTotals)
      .sort(([,a], [,b]) => b - a)
      .slice(0, 6)
      .map(([name, value]) => ({ name, value }));
  };

  const getWeeklyTrends = () => {
    if (!expenses.length) return [];

    const weeklyData = {};
    const now = new Date();
    const oneWeekAgo = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000);

    expenses.forEach(expense => {
      const date = new Date(expense.date);
      if (date >= oneWeekAgo) {
        const dayKey = date.toLocaleDateString('en-US', { weekday: 'short' });
        if (!weeklyData[dayKey]) {
          weeklyData[dayKey] = { income: 0, expenses: 0 };
        }

        if (expense.type === 'income') {
          weeklyData[dayKey].income += expense.amount;
        } else {
          weeklyData[dayKey].expenses += expense.amount;
        }
      }
    });

    const daysOfWeek = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'];
    return daysOfWeek.map(day => ({
      name: day,
      income: weeklyData[day]?.income || 0,
      expenses: weeklyData[day]?.expenses || 0
    }));
  };

  return (
    <ProtectedRoute>
      <div className="flex h-screen bg-gradient-to-br from-[#0d0d0d] via-[#1a1a1a] to-[#0d0d0d]">
        <Sidebar isOpen={sidebarOpen} onClose={() => setSidebarOpen(false)} />
        <div className="flex-1 flex flex-col">
          <Navbar onSidebarToggle={() => setSidebarOpen(!sidebarOpen)} />
          <main className="flex-1 p-4 md:p-6 overflow-auto">
            <div className="max-w-7xl mx-auto">
              <div className="mb-6 md:mb-8">
                <div className="flex items-center justify-between">
                  <div>
                    <h1 className="text-2xl md:text-3xl font-bold bg-gradient-to-r from-[#FFD700] to-[#00FFFF] bg-clip-text text-transparent mb-2">
                      AI Analytics
                    </h1>
                    <p className="text-[#a0a0a0]">Get smart insights about your spending patterns</p>
                    
                    {/* Real-time AI Status */}
                    <div className="flex items-center space-x-3 mt-2">
                      <div className="flex items-center space-x-2">
                        <div className={`w-2 h-2 rounded-full ${isRealTimeEnabled ? 'bg-green-400 animate-pulse' : 'bg-gray-400'}`}></div>
                        <span className="text-sm text-[#a0a0a0]">
                          {isRealTimeEnabled ? 'AI insights live' : 'AI insights offline'}
                        </span>
                      </div>
                      <span className="text-xs text-[#808080]">
                        Last update: {formatLastUpdate()}
                      </span>
                      {aiProcessing && (
                        <span className="text-xs text-[#FFD700] animate-pulse">
                          <i className="ri-brain-line mr-1"></i>
                          AI processing...
                        </span>
                      )}
                    </div>
                  </div>
                  <div className="flex space-x-3">
                    {/* Real-time Toggle */}
                    <button
                      onClick={toggleRealTime}
                      className={`px-4 py-2 rounded-lg font-semibold transition-all duration-300 shadow-lg flex items-center space-x-2 ${
                        isRealTimeEnabled 
                          ? 'bg-gradient-to-r from-green-500 to-emerald-500 text-white hover:from-emerald-500 hover:to-green-500' 
                          : 'bg-gradient-to-r from-gray-500 to-gray-600 text-white hover:from-gray-600 hover:to-gray-500'
                      }`}
                    >
                      <i className={`ri-${isRealTimeEnabled ? 'brain' : 'brain-line'}-line`}></i>
                      <span>{isRealTimeEnabled ? 'AI Live' : 'AI Off'}</span>
                    </button>
                    
                    <button
                      onClick={() => fetchRealTimeData(true)}
                      disabled={loading || aiProcessing}
                      className="px-4 py-2 bg-gradient-to-r from-[#FFD700] to-[#00FFFF] text-[#0d0d0d] rounded-xl font-bold hover:shadow-lg hover:shadow-[#FFD700]/25 disabled:opacity-50 disabled:cursor-not-allowed transition-all duration-300 cursor-pointer flex items-center space-x-2"
                    >
                      <i className={`ri-refresh-line text-lg ${loading || aiProcessing ? 'animate-spin' : ''}`}></i>
                      <span>{loading || aiProcessing ? 'Processing...' : 'Refresh AI'}</span>
                    </button>
                  </div>
                </div>
              </div>

              {loading && (
                <div className="flex justify-center items-center h-64">
                  <LoadingSpinner size="lg" />
                </div>
              )}

              {error && (
                <ErrorMessage message={error} onRetry={fetchRealTimeData} />
              )}

              {!loading && !error && (
                <>
                  <div className="grid grid-cols-1 xl:grid-cols-2 gap-4 md:gap-6 mb-6 md:mb-8">
                    <div className="bg-[#1a1a1a]/80 backdrop-blur-xl rounded-2xl border border-[#333333]/50 p-4 md:p-6 hover:border-[#FFD700]/30 transition-all duration-300">
                      <h3 className="text-lg font-semibold text-white mb-4 flex items-center">
                        <div className="w-6 h-6 bg-gradient-to-r from-[#FFD700] to-[#00FFFF] rounded-full flex items-center justify-center mr-3">
                          <i className="ri-bar-chart-line text-[#0d0d0d] text-sm"></i>
                        </div>
                        Monthly Income vs Expenses
                      </h3>
                      <div className="bg-[#0d0d0d]/50 rounded-xl p-4">
                        <Chart 
                          data={getMonthlyTrends()} 
                          type="bar" 
                          title="Monthly Income vs Expenses" 
                        />
                      </div>
                    </div>

                    <div className="bg-[#1a1a1a]/80 backdrop-blur-xl rounded-2xl border border-[#333333]/50 p-4 md:p-6 hover:border-[#00FFFF]/30 transition-all duration-300">
                      <h3 className="text-lg font-semibold text-white mb-4 flex items-center">
                        <div className="w-6 h-6 bg-gradient-to-r from-[#00FFFF] to-[#40E0D0] rounded-full flex items-center justify-center mr-3">
                          <i className="ri-pie-chart-line text-[#0d0d0d] text-sm"></i>
                        </div>
                        Top Spending Categories
                      </h3>
                      <div className="bg-[#0d0d0d]/50 rounded-xl p-4">
                        <Chart 
                          data={getCategoryComparison()} 
                          type="pie" 
                          title="Top Spending Categories" 
                        />
                      </div>
                    </div>
                  </div>

                  <div className="grid grid-cols-1 xl:grid-cols-2 gap-4 md:gap-6 mb-6 md:mb-8">
                    <div className="bg-[#1a1a1a]/80 backdrop-blur-xl rounded-2xl border border-[#333333]/50 p-4 md:p-6 hover:border-[#FFD700]/30 transition-all duration-300">
                      <h3 className="text-lg font-semibold text-white mb-4 flex items-center">
                        <div className="w-6 h-6 bg-gradient-to-r from-[#FFD700] to-[#FFA500] rounded-full flex items-center justify-center mr-3">
                          <i className="ri-line-chart-line text-[#0d0d0d] text-sm"></i>
                        </div>
                        Weekly Spending Trends
                      </h3>
                      <div className="bg-[#0d0d0d]/50 rounded-xl p-4">
                        <Chart 
                          data={getWeeklyTrends()} 
                          type="area" 
                          title="Weekly Spending Trends" 
                        />
                      </div>
                    </div>

                    <div className="bg-[#1a1a1a]/80 backdrop-blur-xl rounded-2xl border border-[#333333]/50 p-4 md:p-6 hover:border-[#00FFFF]/30 transition-all duration-300">
                      <h3 className="text-lg font-semibold text-white mb-4 flex items-center">
                        <div className="w-6 h-6 bg-gradient-to-r from-purple-500 to-pink-500 rounded-full flex items-center justify-center mr-3">
                          <i className="ri-dashboard-line text-white text-sm"></i>
                        </div>
                        Quick Stats
                      </h3>
                      <div className="grid grid-cols-1 md:grid-cols-3 gap-4 md:gap-6">
                        <div className="text-center p-4 bg-[#0d0d0d]/30 rounded-xl">
                          <div className="w-12 h-12 bg-gradient-to-r from-[#FFD700]/20 to-[#FFA500]/20 rounded-full flex items-center justify-center mx-auto mb-3">
                            <i className="ri-calendar-line text-[#FFD700] text-xl"></i>
                          </div>
                          <p className="text-2xl font-bold text-white">
                            {expenses.length > 0 ? Math.round(expenses.reduce((sum, e) => sum + e.amount, 0) / expenses.length).toLocaleString(undefined, { style: 'currency', currency }) : 0}
                          </p>
                          <p className="text-sm text-[#a0a0a0]">Avg Transaction</p>
                        </div>
                        
                        <div className="text-center p-4 bg-[#0d0d0d]/30 rounded-xl">
                          <div className="w-12 h-12 bg-gradient-to-r from-[#00FFFF]/20 to-[#40E0D0]/20 rounded-full flex items-center justify-center mx-auto mb-3">
                            <i className="ri-trophy-line text-[#00FFFF] text-xl"></i>
                          </div>
                          <p className="text-2xl font-bold text-white">
                            {getCategoryComparison()[0]?.name || 'N/A'}
                          </p>
                          <p className="text-sm text-[#a0a0a0]">Top Category</p>
                        </div>
                        
                        <div className="text-center p-4 bg-[#0d0d0d]/30 rounded-xl">
                          <div className="w-12 h-12 bg-gradient-to-r from-purple-500/20 to-pink-500/20 rounded-full flex items-center justify-center mx-auto mb-3">
                            <i className="ri-time-line text-purple-400 text-xl"></i>
                          </div>
                          <p className="text-2xl font-bold text-white">
                            {getMonthlyTrends().length}
                          </p>
                          <p className="text-sm text-[#a0a0a0]">Active Months</p>
                        </div>
                      </div>
                    </div>
                  </div>

                  <div className="mb-6 md:mb-8">
                    <h2 className="text-xl font-semibold bg-gradient-to-r from-[#FFD700] to-[#00FFFF] bg-clip-text text-transparent mb-4">
                      AI Insights
                    </h2>
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4 md:gap-6">
                      {insights.length > 0 ? (
                        insights.map((insight, index) => (
                          <InsightBox key={index} insight={insight} />
                        ))
                      ) : (
                        <div className="col-span-full bg-[#1a1a1a]/80 backdrop-blur-xl rounded-2xl border border-[#333333]/50 p-8 text-center">
                          <div className="w-16 h-16 bg-[#333333]/30 rounded-full flex items-center justify-center mx-auto mb-4">
                            <i className="ri-brain-line text-[#808080] text-2xl"></i>
                          </div>
                          <h3 className="text-lg font-medium text-white mb-2">No Insights Available</h3>
                          <p className="text-[#a0a0a0]">Add more transactions to get AI-powered insights about your spending patterns.</p>
                        </div>
                      )}
                    </div>
                  </div>
                </>
              )}
            </div>
          </main>
        </div>
      </div>
    </ProtectedRoute>
  );
}
