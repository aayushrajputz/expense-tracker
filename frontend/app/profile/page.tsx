'use client';

import { useState, useEffect, useRef } from 'react';
import { useAuth } from '../../lib/AuthContext';
import { expenseAPI } from '../../lib/api';
import Sidebar from '../../components/Sidebar';
import Navbar from '../../components/Navbar';

export default function ProfilePage() {
  const { user, loading } = useAuth();
  const [sidebarOpen, setSidebarOpen] = useState(true);
  const [profileData, setProfileData] = useState<any>(null);
  const [financialData, setFinancialData] = useState<any>(null);
  const [recentTransactions, setRecentTransactions] = useState<any[]>([]);
  const [dataLoading, setDataLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [downloadLoading, setDownloadLoading] = useState(false);
  const [lastUpdate, setLastUpdate] = useState<Date | null>(null);
  const [isRealTimeEnabled, setIsRealTimeEnabled] = useState(true);
  
  // Refs for real-time functionality
  const pollingIntervalRef = useRef<NodeJS.Timeout | null>(null);
  const lastActivityRef = useRef<number>(Date.now());
  const isPageVisibleRef = useRef<boolean>(true);

  // Real-time data fetching function
  const fetchRealTimeData = async (showLoading = false) => {
    if (!user) return;
    
    try {
      if (showLoading) {
        setDataLoading(true);
      }
      setError(null);

      // Check if user has a token
      const token = localStorage.getItem('token');
      if (!token) {
        // No token, use fallback data
        setProfileData({
          name: user.name || 'User',
          email: user.email || 'user@example.com',
          total_transactions: 0,
          member_since: user.createdAt || new Date().toISOString(),
          account_status: 'Active'
        });
        
        setFinancialData({
          total_income: 0,
          total_expenses: 0,
          net_balance: 0,
          top_categories: {}
        });
        
        setRecentTransactions([]);
        return;
      }

      // Try to fetch data from API, but don't fail if it doesn't work
      try {
        const [profileResponse, summaryResponse, transactionsResponse] = await Promise.all([
          expenseAPI.getProfileData(),
          expenseAPI.getLifetimeSummary(),
          expenseAPI.getRecentTransactions(5)
        ]);

        setProfileData(profileResponse);
        setFinancialData(summaryResponse);
        setRecentTransactions(transactionsResponse);
        setLastUpdate(new Date());
      } catch (apiError: any) {
        console.log('API not available, using fallback data:', apiError.message);
        
        // Use fallback data when API is not available
        setProfileData({
          name: user.name || 'User',
          email: user.email || 'user@example.com',
          total_transactions: 0,
          member_since: user.createdAt || new Date().toISOString(),
          account_status: 'Active'
        });
        
        setFinancialData({
          total_income: 0,
          total_expenses: 0,
          net_balance: 0,
          top_categories: {}
        });
        
        setRecentTransactions([]);
      }

    } catch (err: any) {
      console.error('Error in profile data setup:', err);
      setError('Unable to load profile data. Please try again later.');
      
      // Set fallback data
      setProfileData({
        name: user.name || 'User',
        email: user.email || 'user@example.com',
        total_transactions: 0,
        member_since: user.createdAt || new Date().toISOString(),
        account_status: 'Active'
      });
      
      setFinancialData({
        total_income: 0,
        total_expenses: 0,
        net_balance: 0,
        top_categories: {}
      });
      
      setRecentTransactions([]);
    } finally {
      if (showLoading) {
        setDataLoading(false);
      }
    }
  };

  // Start real-time polling
  const startRealTimePolling = () => {
    if (pollingIntervalRef.current) {
      clearInterval(pollingIntervalRef.current);
    }
    
    if (isRealTimeEnabled) {
      pollingIntervalRef.current = setInterval(() => {
        // Only fetch if page is visible and user has been active recently
        if (isPageVisibleRef.current && (Date.now() - lastActivityRef.current) < 300000) { // 5 minutes
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
        // Page became visible, refresh data immediately
        fetchRealTimeData(false);
        startRealTimePolling();
      } else {
        // Page hidden, stop polling
        stopRealTimePolling();
      }
    };

    // Handle user activity
    const handleUserActivity = () => {
      lastActivityRef.current = Date.now();
    };

    // Handle focus/blur events
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

    // Add event listeners
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
    const initializeData = async () => {
      await fetchRealTimeData(true);
      if (isRealTimeEnabled) {
        startRealTimePolling();
      }
    };

    if (user) {
      initializeData();
    }

    return () => {
      stopRealTimePolling();
    };
  }, [user, isRealTimeEnabled]);

  // Handle clicking outside dropdown
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      const dropdown = document.getElementById('download-dropdown');
      const button = document.querySelector('[data-download-button]');
      
      if (dropdown && !dropdown.contains(event.target as Node) && !button?.contains(event.target as Node)) {
        dropdown.classList.add('hidden');
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, []);

  // Show loading state while data is being fetched
  if (loading || dataLoading) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-[#0d0d0d] to-[#1a1a1a] flex items-center justify-center">
        <div className="text-white text-xl">Loading profile...</div>
      </div>
    );
  }

  // Function to manually refresh user data
  const refreshUserData = async () => {
    if (!user) return;
    
    try {
      setDataLoading(true);
      setError(null);

      // Check if user has a token
      const token = localStorage.getItem('token');
      if (!token) {
        setError('No authentication token found. Please log in again.');
        return;
      }

      try {
        const [profileResponse, summaryResponse, transactionsResponse] = await Promise.all([
          expenseAPI.getProfileData(),
          expenseAPI.getLifetimeSummary(),
          expenseAPI.getRecentTransactions(5)
        ]);

        setProfileData(profileResponse);
        setFinancialData(summaryResponse);
        setRecentTransactions(transactionsResponse);
        setLastUpdate(new Date());
        setError(null); // Clear any previous errors
      } catch (apiError: any) {
        console.log('API not available during refresh:', apiError.message);
        setError('Backend server is not available. Using cached data.');
      }
    } catch (err: any) {
      console.error('Error refreshing data:', err);
      setError('Failed to refresh data. Please try again later.');
    } finally {
      setDataLoading(false);
    }
  };

  // Toggle real-time updates
  const toggleRealTime = () => {
    setIsRealTimeEnabled(!isRealTimeEnabled);
    if (!isRealTimeEnabled) {
      startRealTimePolling();
    } else {
      stopRealTimePolling();
    }
  };

  // Format date for display
  const formatDate = (dateString: string) => {
    const date = new Date(dateString);
    return date.toLocaleDateString('en-US', { 
      year: 'numeric', 
      month: 'long', 
      day: 'numeric' 
    });
  };

  // Format time ago
  const formatTimeAgo = (dateString: string) => {
    const date = new Date(dateString);
    const now = new Date();
    const diffInHours = Math.floor((now.getTime() - date.getTime()) / (1000 * 60 * 60));
    
    if (diffInHours < 1) return 'Just now';
    if (diffInHours < 24) return `${diffInHours} hours ago`;
    if (diffInHours < 48) return '1 day ago';
    return `${Math.floor(diffInHours / 24)} days ago`;
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

  // Download functions
  const downloadDataAsCSV = (data: any[], period: string) => {
    if (!data || data.length === 0) {
      alert('No data available for download');
      return;
    }

    const headers = ['Date', 'Title', 'Category', 'Type', 'Amount', 'Payment Method', 'Notes'];
    const csvContent = [
      headers.join(','),
      ...data.map(expense => [
        expense.date,
        `"${expense.title}"`,
        expense.category || 'Other',
        expense.type,
        expense.amount,
        expense.payment_method || '',
        `"${expense.notes || ''}"`
      ].join(','))
    ].join('\n');

    const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
    const link = document.createElement('a');
    const url = URL.createObjectURL(blob);
    link.setAttribute('href', url);
    link.setAttribute('download', `expense_data_${period}_${new Date().toISOString().split('T')[0]}.csv`);
    link.style.visibility = 'hidden';
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
  };

  const downloadDataAsJSON = (data: any[], period: string) => {
    if (!data || data.length === 0) {
      alert('No data available for download');
      return;
    }

    const jsonData = {
      period: period,
      generatedAt: new Date().toISOString(),
      user: {
        name: profileData?.name || user?.name,
        email: profileData?.email || user?.email
      },
      summary: {
        totalTransactions: data.length,
        totalIncome: data.filter(t => t.type === 'income').reduce((sum, t) => sum + t.amount, 0),
        totalExpenses: data.filter(t => t.type === 'expense').reduce((sum, t) => sum + t.amount, 0)
      },
      transactions: data
    };

    const blob = new Blob([JSON.stringify(jsonData, null, 2)], { type: 'application/json' });
    const link = document.createElement('a');
    const url = URL.createObjectURL(blob);
    link.setAttribute('href', url);
    link.setAttribute('download', `expense_data_${period}_${new Date().toISOString().split('T')[0]}.json`);
    link.style.visibility = 'hidden';
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
  };

  const downloadData = async (period: string, format: 'csv' | 'json') => {
    if (!user) return;
    
    try {
      setDownloadLoading(true);
      
      let startDate = '';
      const endDate = new Date().toISOString().split('T')[0];
      
      // Calculate start date based on period
      const now = new Date();
      switch (period) {
        case 'week':
          startDate = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000).toISOString().split('T')[0];
          break;
        case 'month':
          startDate = new Date(now.getFullYear(), now.getMonth(), 1).toISOString().split('T')[0];
          break;
        case 'year':
          startDate = new Date(now.getFullYear(), 0, 1).toISOString().split('T')[0];
          break;
        case 'all':
          startDate = '';
          break;
        default:
          startDate = '';
      }

      // Fetch all transactions for the period
      const allTransactions = await expenseAPI.getExpenses();
      
      // Filter transactions by date if needed
      let filteredTransactions = allTransactions;
      if (startDate) {
        filteredTransactions = allTransactions.filter((t: any) => t.date >= startDate);
      }

      if (format === 'csv') {
        downloadDataAsCSV(filteredTransactions, period);
      } else {
        downloadDataAsJSON(filteredTransactions, period);
      }
      
    } catch (err: any) {
      console.error('Error downloading data:', err);
      alert('Failed to download data. Please try again.');
    } finally {
      setDownloadLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-[#0d0d0d] to-[#1a1a1a]">
      <Sidebar isOpen={sidebarOpen} onClose={() => setSidebarOpen(false)} />
      <div className={`transition-all duration-300 ${sidebarOpen ? 'ml-64' : 'ml-0'}`}>
        <Navbar onSidebarToggle={() => setSidebarOpen(!sidebarOpen)} />
        
        <div className="p-6">
          {/* Profile Header */}
          <div className="mb-8 flex justify-between items-center">
            <div>
              <h1 className="text-3xl font-bold bg-gradient-to-r from-[#FFD700] to-[#00FFFF] bg-clip-text text-transparent mb-2">
                Profile
              </h1>
              <p className="text-[#a0a0a0]">Manage your account and view your financial overview</p>
              
              {/* Real-time Status */}
              <div className="flex items-center space-x-3 mt-2">
                <div className="flex items-center space-x-2">
                  <div className={`w-2 h-2 rounded-full ${isRealTimeEnabled ? 'bg-green-400 animate-pulse' : 'bg-gray-400'}`}></div>
                  <span className="text-sm text-[#a0a0a0]">
                    {isRealTimeEnabled ? 'Real-time updates enabled' : 'Real-time updates disabled'}
                  </span>
                </div>
                <span className="text-xs text-[#808080]">
                  Last updated: {formatLastUpdate()}
                </span>
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
                <i className={`ri-${isRealTimeEnabled ? 'wifi' : 'wifi-off'}-line`}></i>
                <span>{isRealTimeEnabled ? 'Live' : 'Offline'}</span>
              </button>
              
              <button
                onClick={refreshUserData}
                disabled={dataLoading}
                className="px-4 py-2 bg-gradient-to-r from-[#FFD700] to-[#00FFFF] text-[#0d0d0d] rounded-lg font-semibold hover:from-[#00FFFF] hover:to-[#FFD700] transition-all duration-300 shadow-lg disabled:opacity-50"
              >
                {dataLoading ? 'Refreshing...' : 'Refresh Data'}
              </button>
              
              {/* Download Dropdown */}
              <div className="relative">
                <button
                  disabled={downloadLoading}
                  data-download-button
                  className="px-4 py-2 bg-gradient-to-r from-[#00FFFF] to-[#40E0D0] text-[#0d0d0d] rounded-lg font-semibold hover:from-[#40E0D0] hover:to-[#00FFFF] transition-all duration-300 shadow-lg disabled:opacity-50 flex items-center space-x-2"
                  onClick={() => {
                    const dropdown = document.getElementById('download-dropdown');
                    if (dropdown) {
                      dropdown.classList.toggle('hidden');
                    }
                  }}
                >
                  <i className="ri-download-line"></i>
                  <span>{downloadLoading ? 'Downloading...' : 'Download Data'}</span>
                  <i className="ri-arrow-down-s-line"></i>
                </button>
                
                {/* Download Options Dropdown */}
                <div id="download-dropdown" className="hidden absolute right-0 mt-2 w-64 bg-[#1a1a1a] border border-[#333333]/50 rounded-lg shadow-2xl z-50">
                  <div className="p-4">
                    <h3 className="text-white font-semibold mb-3">Download Options</h3>
                    
                    {/* Time Periods */}
                    <div className="mb-4">
                      <h4 className="text-[#a0a0a0] text-sm mb-2">Time Period:</h4>
                      <div className="grid grid-cols-2 gap-2">
                        <button
                          onClick={() => downloadData('week', 'csv')}
                          disabled={downloadLoading}
                          className="px-3 py-2 bg-[#333333] hover:bg-[#444444] text-white text-sm rounded transition-colors disabled:opacity-50"
                        >
                          Last Week (CSV)
                        </button>
                        <button
                          onClick={() => downloadData('week', 'json')}
                          disabled={downloadLoading}
                          className="px-3 py-2 bg-[#333333] hover:bg-[#444444] text-white text-sm rounded transition-colors disabled:opacity-50"
                        >
                          Last Week (JSON)
                        </button>
                        <button
                          onClick={() => downloadData('month', 'csv')}
                          disabled={downloadLoading}
                          className="px-3 py-2 bg-[#333333] hover:bg-[#444444] text-white text-sm rounded transition-colors disabled:opacity-50"
                        >
                          Last Month (CSV)
                        </button>
                        <button
                          onClick={() => downloadData('month', 'json')}
                          disabled={downloadLoading}
                          className="px-3 py-2 bg-[#333333] hover:bg-[#444444] text-white text-sm rounded transition-colors disabled:opacity-50"
                        >
                          Last Month (JSON)
                        </button>
                        <button
                          onClick={() => downloadData('year', 'csv')}
                          disabled={downloadLoading}
                          className="px-3 py-2 bg-[#333333] hover:bg-[#444444] text-white text-sm rounded transition-colors disabled:opacity-50"
                        >
                          This Year (CSV)
                        </button>
                        <button
                          onClick={() => downloadData('year', 'json')}
                          disabled={downloadLoading}
                          className="px-3 py-2 bg-[#333333] hover:bg-[#444444] text-white text-sm rounded transition-colors disabled:opacity-50"
                        >
                          This Year (JSON)
                        </button>
                        <button
                          onClick={() => downloadData('all', 'csv')}
                          disabled={downloadLoading}
                          className="px-3 py-2 bg-[#333333] hover:bg-[#444444] text-white text-sm rounded transition-colors disabled:opacity-50"
                        >
                          All Time (CSV)
                        </button>
                        <button
                          onClick={() => downloadData('all', 'json')}
                          disabled={downloadLoading}
                          className="px-3 py-2 bg-[#333333] hover:bg-[#444444] text-white text-sm rounded transition-colors disabled:opacity-50"
                        >
                          All Time (JSON)
                        </button>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>

          {error && (
            <div className="mb-6 p-4 bg-red-500/20 border border-red-500/50 rounded-lg">
              <p className="text-red-400">{error}</p>
            </div>
          )}

          <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
            {/* Profile Card */}
            <div className="lg:col-span-1">
              <div className="bg-[#1a1a1a] rounded-xl p-6 border border-[#333333]/50 shadow-2xl">
                <div className="text-center mb-6">
                  <div className="w-24 h-24 bg-gradient-to-r from-[#FFD700] to-[#00FFFF] rounded-full flex items-center justify-center mx-auto mb-4">
                    <span className="text-[#0d0d0d] text-3xl font-bold">
                      {profileData?.name ? 
                        profileData.name.charAt(0).toUpperCase() :
                        (profileData?.email ? 
                          profileData.email.charAt(0).toUpperCase() : 
                          'U'
                        )
                      }
                    </span>
                  </div>
                  <h2 className="text-xl font-bold text-white mb-1">
                    {profileData?.name ? 
                      (profileData.name.length > 20 ? 
                        profileData.name.substring(0, 20) + '...' : 
                        profileData.name
                      ) : 
                      (profileData?.email ? 
                        profileData.email.split('@')[0].substring(0, 15) : 
                        'User'
                      )
                    }
                  </h2>
                  <p className="text-[#808080] text-sm mb-3">
                    {profileData?.email || 'user@example.com'}
                  </p>
                  <div className="flex items-center justify-center space-x-2">
                    <div className="w-2 h-2 bg-[#00FFFF] rounded-full"></div>
                    <span className="text-[#00FFFF] text-sm font-medium">Premium Account</span>
                  </div>
                </div>

                <div className="space-y-4">
                  <div className="flex justify-between items-center py-2 border-b border-[#333333]/30">
                    <span className="text-[#a0a0a0]">Member since</span>
                    <span className="text-white font-medium">
                      {profileData?.member_since ? formatDate(profileData.member_since) : 'N/A'}
                    </span>
                  </div>
                  <div className="flex justify-between items-center py-2 border-b border-[#333333]/30">
                    <span className="text-[#a0a0a0]">Total Transactions</span>
                    <span className="text-white font-medium">{profileData?.total_transactions || 0}</span>
                  </div>
                  <div className="flex justify-between items-center py-2">
                    <span className="text-[#a0a0a0]">Account Status</span>
                    <span className="text-green-400 font-medium">{profileData?.account_status || 'Active'}</span>
                  </div>
                </div>
              </div>
            </div>

            {/* Financial Overview */}
            <div className="lg:col-span-2">
              <div className="grid grid-cols-1 md:grid-cols-2 gap-6 mb-6">
                <div className="bg-[#1a1a1a] rounded-xl p-6 border border-[#333333]/50 shadow-2xl">
                  <div className="flex items-center justify-between mb-4">
                    <h3 className="text-lg font-semibold text-white">Total Income</h3>
                    <div className="w-10 h-10 bg-green-500/20 rounded-full flex items-center justify-center">
                      <i className="ri-arrow-up-line text-green-400 text-lg"></i>
                    </div>
                  </div>
                  <p className="text-2xl font-bold text-green-400 mb-2">
                    ${financialData?.total_income?.toFixed(2) || '0.00'}
                  </p>
                  <p className="text-[#a0a0a0] text-sm">Lifetime earnings</p>
                </div>

                <div className="bg-[#1a1a1a] rounded-xl p-6 border border-[#333333]/50 shadow-2xl">
                  <div className="flex items-center justify-between mb-4">
                    <h3 className="text-lg font-semibold text-white">Total Spent</h3>
                    <div className="w-10 h-10 bg-red-500/20 rounded-full flex items-center justify-center">
                      <i className="ri-arrow-down-line text-red-400 text-lg"></i>
                    </div>
                  </div>
                  <p className="text-2xl font-bold text-red-400 mb-2">
                    ${financialData?.total_expenses?.toFixed(2) || '0.00'}
                  </p>
                  <p className="text-[#a0a0a0] text-sm">Lifetime expenses</p>
                </div>
              </div>

              {/* Spending by Category */}
              {financialData?.top_categories && Object.keys(financialData.top_categories).length > 0 && (
                <div className="bg-[#1a1a1a] rounded-xl p-6 border border-[#333333]/50 shadow-2xl mb-6">
                  <h3 className="text-lg font-semibold text-white mb-4">Spending by Category</h3>
                  <div className="space-y-3">
                    {Object.entries(financialData.top_categories).map(([category, amount], index) => {
                      const colors = ['bg-red-400', 'bg-cyan-400', 'bg-blue-400', 'bg-green-400', 'bg-yellow-400'];
                      const color = colors[index % colors.length];
                      
                      return (
                        <div key={category} className="flex items-center justify-between">
                          <div className="flex items-center space-x-3">
                            <div className={`w-3 h-3 rounded-full ${color}`}></div>
                            <span className="text-white">{category}</span>
                          </div>
                          <div className="text-right">
                            <p className="text-white font-medium">${(amount as number).toFixed(2)}</p>
                            <p className="text-[#a0a0a0] text-sm">Category spending</p>
                          </div>
                        </div>
                      );
                    })}
                  </div>
                </div>
              )}

              {/* Recent Activity */}
              {recentTransactions.length > 0 && (
                <div className="bg-[#1a1a1a] rounded-xl p-6 border border-[#333333]/50 shadow-2xl">
                  <h3 className="text-lg font-semibold text-white mb-4">Recent Activity</h3>
                  <div className="space-y-3">
                    {recentTransactions.map((transaction, index) => (
                      <div key={transaction.id} className={`flex items-center justify-between py-2 ${index < recentTransactions.length - 1 ? 'border-b border-[#333333]/30' : ''}`}>
                        <div className="flex items-center space-x-3">
                          <div className={`w-8 h-8 rounded-full flex items-center justify-center ${transaction.type === 'income' ? 'bg-green-500/20' : 'bg-red-500/20'}`}>
                            <i className={`ri-arrow-${transaction.type === 'income' ? 'up' : 'down'}-line ${transaction.type === 'income' ? 'text-green-400' : 'text-red-400'}`}></i>
                          </div>
                          <div>
                            <p className="text-white font-medium">
                              {transaction.category && transaction.category !== 'Other' 
                                ? transaction.category 
                                : (transaction.title && transaction.title.length > 0 
                                    ? transaction.title 
                                    : 'Transaction'
                                  )
                              }
                            </p>
                            <p className="text-[#a0a0a0] text-sm">{formatTimeAgo(transaction.date)}</p>
                          </div>
                        </div>
                        <span className={`font-bold ${transaction.type === 'income' ? 'text-green-400' : 'text-red-400'}`}>
                          {transaction.type === 'income' ? '+' : '-'}${transaction.amount.toFixed(2)}
                        </span>
                      </div>
                    ))}
                  </div>
                </div>
              )}

              {/* No data message */}
              {(!financialData?.top_categories || Object.keys(financialData.top_categories).length === 0) && recentTransactions.length === 0 && (
                <div className="bg-[#1a1a1a] rounded-xl p-6 border border-[#333333]/50 shadow-2xl text-center">
                  <p className="text-[#a0a0a0] mb-4">No financial data available yet</p>
                  <p className="text-[#808080] text-sm">Start adding expenses and income to see your financial overview</p>
                </div>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}