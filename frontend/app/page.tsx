
'use client';

import ProtectedRoute from '../components/ProtectedRoute';
import Sidebar from '../components/Sidebar';
import Navbar from '../components/Navbar';
import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { expenseAPI, clearAllCaches } from '../lib/api';
import LoadingSpinner from '../components/LoadingSpinner';
import ErrorMessage from '../components/ErrorMessage';
import ExpenseCard from '../components/ExpenseCard';
import Chart from '../components/Chart';
import { useTranslation } from '../lib/useTranslation';

export default function DashboardPage() {
  const [expenses, setExpenses] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const router = useRouter();
  const { t } = useTranslation();

  useEffect(() => {
    fetchExpenses();
  }, []);

  // Monitor expenses changes and log for debugging
  useEffect(() => {
    console.log('Expenses updated:', expenses);
    console.log('Total expenses:', expenses.filter(expense => expense.type === 'expense').reduce((sum, expense) => sum + expense.amount, 0));
    console.log('Total income:', expenses.filter(expense => expense.type === 'income').reduce((sum, expense) => sum + expense.amount, 0));
    console.log('Transaction count:', expenses.length);
  }, [expenses]);

  // Refresh data when page becomes visible
  useEffect(() => {
    const handleVisibilityChange = () => {
      if (!document.hidden) {
        fetchExpenses();
      }
    };

    document.addEventListener('visibilitychange', handleVisibilityChange);
    return () => {
      document.removeEventListener('visibilitychange', handleVisibilityChange);
    };
  }, []);

  const fetchExpenses = async () => {
    try {
      setLoading(true);
      const data = await expenseAPI.getExpenses();
      // The API returns expenses array directly, not wrapped in an object
      const expensesArray = Array.isArray(data) ? data : [];
      console.log('Fetched expenses:', expensesArray);
      setExpenses(expensesArray);
    } catch (err) {
      setError(t('failedToFetchExpenses'));
      // Add mock data for development when backend is not available
      const mockExpenses = [
        {
          id: 1,
          title: t('groceryShopping'),
          amount: 150.50,
          type: 'expense',
          category: t('food'),
          date: '2024-01-15',
          description: t('weeklyGroceries')
        },
        {
          id: 2,
          title: t('salary'),
          amount: 3000.00,
          type: 'income',
          category: t('income'),
          date: '2024-01-01',
          description: t('monthlySalary')
        },
        {
          id: 3,
          title: t('gasStation'),
          amount: 45.00,
          type: 'expense',
          category: t('transportation'),
          date: '2024-01-14',
          description: t('fuelForCar')
        },
        {
          id: 4,
          title: t('restaurant'),
          amount: 85.00,
          type: 'expense',
          category: t('food'),
          date: '2024-01-13',
          description: t('dinnerWithFriends')
        },
        {
          id: 5,
          title: t('freelanceWork'),
          amount: 500.00,
          type: 'income',
          category: t('income'),
          date: '2024-01-10',
          description: t('webDevelopmentProject')
        }
      ];
      console.log('Using mock expenses:', mockExpenses);
      setExpenses(mockExpenses);
    } finally {
      setLoading(false);
    }
  };

  const forceResetDashboard = () => {
    console.log('Force resetting dashboard');
    setExpenses([]);
    clearAllCaches();
  };

  const getSpendingTrendsData = () => {
    if (!expenses.length) return [];

    // Group expenses by date and calculate daily totals
    const dailyData: { [key: string]: { income: number; expenses: number } } = {};
    expenses.forEach(expense => {
      const date = new Date(expense.date);
      const dateKey = date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
      
      if (!dailyData[dateKey]) {
        dailyData[dateKey] = { income: 0, expenses: 0 };
      }

      if (expense.type === 'income') {
        dailyData[dateKey].income += expense.amount;
      } else {
        dailyData[dateKey].expenses += expense.amount;
      }
    });

    // Convert to array format expected by Chart component
    return Object.entries(dailyData)
      .sort(([a], [b]) => new Date(a).getTime() - new Date(b).getTime())
      .map(([name, data]) => ({
        name,
        income: data.income,
        expenses: data.expenses
      }));
  };

  const totalExpenses = expenses.length > 0 ? expenses.filter(expense => expense.type === 'expense').reduce((sum, expense) => sum + expense.amount, 0) : 0;
  const totalIncome = expenses.length > 0 ? expenses.filter(expense => expense.type === 'income').reduce((sum, expense) => sum + expense.amount, 0) : 0;
  const netBalance = totalIncome - totalExpenses;
  
  // Sort expenses by newest first, then take the first 5 for recent transactions
  const recentExpenses = expenses
    .sort((a, b) => {
      // Sort by full timestamp (date + time) to ensure most recent transaction appears first
      const timeA = new Date(a.date + ' ' + (a.time || '00:00:00')).getTime();
      const timeB = new Date(b.date + ' ' + (b.time || '00:00:00')).getTime();
      return timeB - timeA; // Newest first
    })
    .slice(0, 5);

  const handleDeleteExpense = async (id) => {
    if (!confirm('Are you sure you want to delete this transaction?')) {
      return;
    }
    
    try {
      await expenseAPI.deleteExpense(id);
      
      // Remove the expense from the local state and force refresh
      setExpenses(prevExpenses => {
        const updatedExpenses = prevExpenses.filter(expense => expense.id !== id);
        console.log('Updated expenses after deletion:', updatedExpenses);
        
        // Force a complete refresh if all transactions are deleted
        if (updatedExpenses.length === 0) {
          console.log('All transactions deleted, forcing refresh');
          setTimeout(() => {
            forceResetDashboard();
          }, 100);
        }
        
        return updatedExpenses;
      });
      
      // Clear cache to ensure fresh data
      clearAllCaches();
      
      alert('Transaction deleted successfully!');
    } catch (error) {
      console.error('Failed to delete transaction:', error);
      
      // If it's a network error (backend not running), still remove from local state
      if (error.message && error.message.includes('Backend server is not running')) {
        // Remove the expense from the local state even if backend is not available
        setExpenses(prevExpenses => {
          const updatedExpenses = prevExpenses.filter(expense => expense.id !== id);
          console.log('Updated expenses after deletion (offline):', updatedExpenses);
          
          // Force a complete refresh if all transactions are deleted
          if (updatedExpenses.length === 0) {
            console.log('All transactions deleted (offline), forcing refresh');
            setTimeout(() => {
              forceResetDashboard();
            }, 100);
          }
          
          return updatedExpenses;
        });
        
        // Clear cache to ensure fresh data
        clearAllCaches();
        
        alert('Transaction deleted successfully! (Offline mode)');
      } else {
        alert('Failed to delete transaction. Please try again.');
      }
    }
  };

  if (loading) {
    return (
      <ProtectedRoute>
        <div className="min-h-screen bg-gradient-to-br from-[#0d0d0d] via-[#1a1a1a] to-[#0d0d0d] flex items-center justify-center">
          <LoadingSpinner size="lg" />
        </div>
      </ProtectedRoute>
    );
  }

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
                      {t('dashboard')}
                    </h1>
                    <p className="text-[#a0a0a0]">{t('monitorFinancialInsights')}</p>
                  </div>
                  <button
                    onClick={fetchExpenses}
                    disabled={loading}
                    className="px-4 py-2 bg-gradient-to-r from-[#FFD700] to-[#00FFFF] text-[#0d0d0d] rounded-xl font-bold hover:shadow-lg hover:shadow-[#FFD700]/25 disabled:opacity-50 disabled:cursor-not-allowed transition-all duration-300 cursor-pointer flex items-center space-x-2"
                  >
                    <i className={`ri-refresh-line text-lg ${loading ? 'animate-spin' : ''}`}></i>
                    <span>{loading ? t('refreshing') : t('refresh')}</span>
                  </button>
                </div>
              </div>
              
              {error && (
                <ErrorMessage message={error} onRetry={fetchExpenses} />
              )}
              
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 md:gap-6 mb-6 md:mb-8">
                <div className="bg-[#1a1a1a]/80 backdrop-blur-xl rounded-2xl border border-[#333333]/50 p-6 hover:border-[#FFD700]/30 transition-all duration-300">
                  <div className="flex items-center justify-between mb-4">
                    <h3 className="text-lg font-semibold text-white">{t('totalExpenses')}</h3>
                    <div className="w-10 h-10 bg-gradient-to-r from-red-500/20 to-red-600/20 rounded-full flex items-center justify-center">
                      <i className="ri-arrow-down-line text-red-400 text-xl"></i>
                    </div>
                  </div>
                  <p className="text-2xl md:text-3xl font-bold text-red-400">${totalExpenses.toFixed(2)}</p>
                  <p className="text-[#808080] text-sm mt-2">{t('currentMonth')}</p>
                </div>
                
                <div className="bg-[#1a1a1a]/80 backdrop-blur-xl rounded-2xl border border-[#333333]/50 p-6 hover:border-[#00FFFF]/30 transition-all duration-300">
                  <div className="flex items-center justify-between mb-4">
                    <h3 className="text-lg font-semibold text-white">{t('totalIncome')}</h3>
                    <div className="w-10 h-10 bg-gradient-to-r from-green-500/20 to-green-600/20 rounded-full flex items-center justify-center">
                      <i className="ri-arrow-up-line text-green-400 text-xl"></i>
                    </div>
                  </div>
                  <p className="text-2xl md:text-3xl font-bold text-green-400">${totalIncome.toFixed(2)}</p>
                  <p className="text-[#808080] text-sm mt-2">{t('currentMonth')}</p>
                </div>
                
                <div className="bg-[#1a1a1a]/80 backdrop-blur-xl rounded-2xl border border-[#333333]/50 p-6 hover:border-[#FFD700]/30 transition-all duration-300">
                  <div className="flex items-center justify-between mb-4">
                    <h3 className="text-lg font-semibold text-white">{t('totalBalance')}</h3>
                    <div className="w-10 h-10 bg-gradient-to-r from-[#FFD700]/20 to-[#FFA500]/20 rounded-full flex items-center justify-center">
                      <i className="ri-wallet-line text-[#FFD700] text-xl"></i>
                    </div>
                  </div>
                  <p className={`text-2xl md:text-3xl font-bold ${netBalance >= 0 ? 'text-green-400' : 'text-red-400'}`}>
                    ${netBalance.toFixed(2)}
                  </p>
                  <p className="text-[#808080] text-sm mt-2">{t('incomeMinusExpenses')}</p>
                </div>

                <div className="bg-[#1a1a1a]/80 backdrop-blur-xl rounded-2xl border border-[#333333]/50 p-6 hover:border-[#00FFFF]/30 transition-all duration-300">
                  <div className="flex items-center justify-between mb-4">
                    <h3 className="text-lg font-semibold text-white">{t('transactions')}</h3>
                    <div className="w-10 h-10 bg-gradient-to-r from-[#00FFFF]/20 to-[#40E0D0]/20 rounded-full flex items-center justify-center">
                      <i className="ri-exchange-line text-[#00FFFF] text-xl"></i>
                    </div>
                  </div>
                  <p className="text-2xl md:text-3xl font-bold text-[#00FFFF]">{expenses.length}</p>
                  <p className="text-[#808080] text-sm mt-2">{t('totalCount')}</p>
                </div>
              </div>

              <div className="grid grid-cols-1 xl:grid-cols-2 gap-4 md:gap-6">
                <div className="bg-[#1a1a1a]/80 backdrop-blur-xl rounded-2xl border border-[#333333]/50 p-4 md:p-6 hover:border-[#FFD700]/30 transition-all duration-300">
                  <h3 className="text-lg font-semibold text-white mb-4 flex items-center">
                    <div className="w-6 h-6 bg-gradient-to-r from-[#FFD700] to-[#00FFFF] rounded-full flex items-center justify-center mr-3">
                      <i className="ri-line-chart-line text-[#0d0d0d] text-sm"></i>
                    </div>
                    {t('spendingTrends')}
                  </h3>
                  <div className="bg-[#0d0d0d]/50 rounded-xl p-4">
                    <Chart 
                      data={getSpendingTrendsData()} 
                      type="line" 
                      title={t('spendingTrends')}
                    />
                  </div>
                </div>
                
                <div className="bg-[#1a1a1a]/80 backdrop-blur-xl rounded-2xl border border-[#333333]/50 p-4 md:p-6 hover:border-[#00FFFF]/30 transition-all duration-300">
                  <h3 className="text-lg font-semibold text-white mb-4 flex items-center">
                    <div className="w-6 h-6 bg-gradient-to-r from-[#00FFFF] to-[#40E0D0] rounded-full flex items-center justify-center mr-3">
                      <i className="ri-history-line text-[#0d0d0d] text-sm"></i>
                    </div>
                    {t('recentTransactions')}
                  </h3>
                  <div className="space-y-3 max-h-80 overflow-y-auto">
                    {recentExpenses.length > 0 ? (
                      recentExpenses.map((expense, index) => (
                        <ExpenseCard key={expense.id || `expense-${index}`} expense={expense} onDelete={handleDeleteExpense} />
                      ))
                    ) : (
                      <div className="text-center py-8">
                        <div className="w-16 h-16 bg-[#333333]/30 rounded-full flex items-center justify-center mx-auto mb-4">
                          <i className="ri-wallet-line text-[#808080] text-2xl"></i>
                        </div>
                        <p className="text-[#808080]">No transactions yet</p>
                      </div>
                    )}
                  </div>
                </div>
              </div>
            </div>
          </main>
        </div>
      </div>
    </ProtectedRoute>
  );
}
