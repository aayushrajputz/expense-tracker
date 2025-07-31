'use client';

import { useState, useEffect } from 'react';
import Sidebar from '../../components/Sidebar';
import Navbar from '../../components/Navbar';
import LoadingSpinner from '../../components/LoadingSpinner';
import ProtectedRoute from '../../components/ProtectedRoute';
import { expenseAPI, clearAllCaches } from '../../lib/api';
import ErrorMessage from '../../components/ErrorMessage';
import Link from 'next/link';

export default function TransactionsPage() {
  const [transactions, setTransactions] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const [filter, setFilter] = useState('all');
  const [sortBy, setSortBy] = useState('date');
  const [searchTerm, setSearchTerm] = useState('');
  const [deleteConfirmId, setDeleteConfirmId] = useState(null);

  useEffect(() => {
    fetchTransactions();
  }, []);

  // Refresh data when page becomes visible
  useEffect(() => {
    const handleVisibilityChange = () => {
      if (!document.hidden) {
        fetchTransactions();
      }
    };

    document.addEventListener('visibilitychange', handleVisibilityChange);
    return () => {
      document.removeEventListener('visibilitychange', handleVisibilityChange);
    };
  }, []);

  const fetchTransactions = async () => {
    try {
      setLoading(true);
      setError('');
      const data = await expenseAPI.getExpenses();
      
      // Ensure all transactions have proper IDs
      const transactionsWithIds = Array.isArray(data) ? data.map((transaction, index) => ({
        ...transaction,
        id: transaction.id || `temp-${Date.now()}-${index}`
      })) : [];
      
      setTransactions(transactionsWithIds);
    } catch (err) {
      setError('Failed to fetch transactions');
      // Add mock data for development when backend is not available
      const mockTransactions = [
        {
          id: 1,
          title: 'Salary',
          amount: 20000.00,
          type: 'income',
          category: 'Income',
          date: '2025-07-29',
          time: '09:00:00',
          description: 'Monthly salary'
        },
        {
          id: 2,
          title: 'Movie',
          amount: 250.00,
          type: 'expense',
          category: 'Entertainment',
          date: '2025-07-29',
          time: '20:30:00',
          description: 'Movie tickets'
        },
        {
          id: 3,
          title: 'Food',
          amount: 400.00,
          type: 'expense',
          category: 'Food',
          date: '2025-06-15',
          time: '19:15:00',
          description: 'Restaurant dinner'
        },
        {
          id: 4,
          title: 'Fuel',
          amount: 70.00,
          type: 'expense',
          category: 'Transportation',
          date: '2025-05-20',
          time: '14:45:00',
          description: 'Gas station'
        },
        {
          id: 5,
          title: 'Freelance Work',
          amount: 500.00,
          type: 'income',
          category: 'Income',
          date: '2024-12-10',
          time: '16:20:00',
          description: 'Web development project'
        },
        {
          id: 6,
          title: 'Grocery Shopping',
          amount: 150.00,
          type: 'expense',
          category: 'Food',
          date: '2024-11-25',
          time: '10:30:00',
          description: 'Weekly groceries'
        }
      ];
      setTransactions(mockTransactions);
    } finally {
      setLoading(false);
    }
  };

  const filteredTransactions = transactions.filter(transaction => {
    const matchesFilter = filter === 'all' || transaction.type === filter;
    const matchesSearch = transaction.title.toLowerCase().includes(searchTerm.toLowerCase()) ||
                         transaction.category.toLowerCase().includes(searchTerm.toLowerCase());
    return matchesFilter && matchesSearch;
  });

  const sortedTransactions = filteredTransactions.sort((a, b) => {
    switch (sortBy) {
      case 'date':
        // Ensure newest transactions come first with full timestamp
        const timeA = new Date(a.date + ' ' + (a.time || '00:00:00')).getTime();
        const timeB = new Date(b.date + ' ' + (b.time || '00:00:00')).getTime();
        return timeB - timeA; // Newest first
      case 'month':
        // Sort by month (newest month first)
        const monthA = new Date(a.date).getMonth();
        const monthB = new Date(b.date).getMonth();
        const yearA = new Date(a.date).getFullYear();
        const yearB = new Date(b.date).getFullYear();
        if (yearA !== yearB) return yearB - yearA; // Newer year first
        return monthB - monthA; // Newer month first
      case 'year':
        // Sort by year (newest year first)
        const yearA2 = new Date(a.date).getFullYear();
        const yearB2 = new Date(b.date).getFullYear();
        return yearB2 - yearA2; // Newest year first
      case 'amount':
        return b.amount - a.amount;
      case 'category':
        return a.category.localeCompare(b.category);
      default:
        // Default to newest timestamp first
        const defaultTimeA = new Date(a.date + ' ' + (a.time || '00:00:00')).getTime();
        const defaultTimeB = new Date(b.date + ' ' + (b.time || '00:00:00')).getTime();
        return defaultTimeB - defaultTimeA;
    }
  });

  const getTypeColor = (type) => {
    return type === 'income' ? 'text-green-400' : 'text-red-400';
  };

  const getStatusColor = (status) => {
    switch(status) {
      case 'completed': return 'text-green-400 bg-green-400/20';
      case 'pending': return 'text-yellow-400 bg-yellow-400/20';
      case 'failed': return 'text-red-400 bg-red-400/20';
      default: return 'text-gray-400 bg-gray-400/20';
    }
  };

  const getCategoryIcon = (category) => {
    const icons = {
      'Food': 'ri-restaurant-line',
      'Transportation': 'ri-car-line',
      'Shopping': 'ri-shopping-bag-line',
      'Entertainment': 'ri-gamepad-line',
      'Healthcare': 'ri-heart-pulse-line',
      'Education': 'ri-book-line',
      'Bills': 'ri-bill-line',
      'Travel': 'ri-plane-line',
      'Income': 'ri-money-dollar-circle-line'
    };
    return icons[category] || 'ri-wallet-line';
  };

  const showDeleteConfirmation = (id) => {
    setDeleteConfirmId(id);
  };

  const hideDeleteConfirmation = () => {
    setDeleteConfirmId(null);
  };

  const handleDeleteTransaction = async (id) => {
    console.log('Delete button clicked with ID:', id);
    
    if (!id) {
      console.log('No ID provided, cannot delete');
      alert('Cannot delete transaction: No ID found');
      return;
    }
    
    // Check if it's a temporary ID (from mock data or generated)
    const isTemporaryId = typeof id === 'string' && id.startsWith('temp-');
    
    // Simple confirmation
    if (!confirm('Are you sure you want to delete this transaction?')) {
      console.log('Delete cancelled by user');
      return;
    }
    
    console.log('Proceeding with delete for ID:', id);
    
    try {
      // Only call API if it's not a temporary ID
      if (!isTemporaryId) {
        console.log('Calling deleteExpense API...');
        await expenseAPI.deleteExpense(id);
        console.log('Delete API call successful');
      } else {
        console.log('Skipping API call for temporary ID');
      }
      
      // Remove the transaction from the local state
      setTransactions(prevTransactions => 
        prevTransactions.filter(transaction => transaction.id !== id)
      );
      
      // Clear cache to ensure dashboard is updated
      clearAllCaches();
      
      alert(isTemporaryId ? 'Transaction deleted successfully! (Local only)' : 'Transaction deleted successfully!');
      
    } catch (error) {
      console.error('Failed to delete transaction:', error);
      
      // If it's a network error (backend not running), still remove from local state
      if (error.message && error.message.includes('Backend server is not running')) {
        console.log('Backend not running, removing from local state only');
        // Remove the transaction from the local state even if backend is not available
        setTransactions(prevTransactions => 
          prevTransactions.filter(transaction => transaction.id !== id)
        );
        
        // Clear cache to ensure dashboard is updated
        clearAllCaches();
        
        alert('Transaction deleted successfully! (Offline mode)');
      } else {
        alert('Failed to delete transaction. Please try again.');
      }
    }
  };

  const confirmDelete = async () => {
    if (!deleteConfirmId) return;
    
    try {
      // Call the delete API
      await expenseAPI.deleteExpense(deleteConfirmId);
      
      // Remove the transaction from the local state
      setTransactions(prevTransactions => 
        prevTransactions.filter(transaction => transaction.id !== deleteConfirmId)
      );
      
      // Hide confirmation and show success
      setDeleteConfirmId(null);
      alert('Transaction deleted successfully!');
      
    } catch (error) {
      console.error('Failed to delete transaction:', error);
      
      // If it's a network error (backend not running), still remove from local state
      if (error.message && error.message.includes('Backend server is not running')) {
        // Remove the transaction from the local state even if backend is not available
        setTransactions(prevTransactions => 
          prevTransactions.filter(transaction => transaction.id !== deleteConfirmId)
        );
        
        // Hide confirmation and show success
        setDeleteConfirmId(null);
        alert('Transaction deleted successfully! (Offline mode)');
      } else {
        alert('Failed to delete transaction. Please try again.');
      }
    }
  };

  return (
    <ProtectedRoute>
      <div className="flex h-screen bg-gradient-to-br from-[#0d0d0d] via-[#1a1a1a] to-[#0d0d0d] min-h-screen">
        <Sidebar isOpen={sidebarOpen} onClose={() => setSidebarOpen(false)} />
        <div className="flex-1 flex flex-col bg-gradient-to-br from-[#0d0d0d] via-[#1a1a1a] to-[#0d0d0d] min-h-screen">
          <Navbar onSidebarToggle={() => setSidebarOpen(!sidebarOpen)} />
          <main className="p-4 md:p-6 flex-1 bg-gradient-to-br from-[#0d0d0d] via-[#1a1a1a] to-[#0d0d0d] min-h-full">
            <div className="mb-6 md:mb-8">
              <div className="flex items-center justify-between">
                <div>
                  <h1 className="text-2xl md:text-3xl font-bold bg-gradient-to-r from-[#FFD700] to-[#00FFFF] bg-clip-text text-transparent mb-2">
                    Transactions
                  </h1>
                  <p className="text-[#a0a0a0]">View and manage your transaction history</p>
                </div>
                <button
                  onClick={fetchTransactions}
                  disabled={loading}
                  className="px-4 py-2 bg-gradient-to-r from-[#FFD700] to-[#00FFFF] text-[#0d0d0d] rounded-xl font-bold hover:shadow-lg hover:shadow-[#FFD700]/25 disabled:opacity-50 disabled:cursor-not-allowed transition-all duration-300 cursor-pointer flex items-center space-x-2"
                >
                  <i className={`ri-refresh-line text-lg ${loading ? 'animate-spin' : ''}`}></i>
                  <span>{loading ? 'Refreshing...' : 'Refresh'}</span>
                </button>
              </div>
            </div>

            {error && (
              <div className="mb-6">
                <ErrorMessage message={error} onRetry={fetchTransactions} />
              </div>
            )}

            {/* Delete Confirmation Section */}
            {deleteConfirmId && (
              <div className="bg-[#1a1a1a]/80 backdrop-blur-xl rounded-2xl border border-red-500/30 p-6 mb-6">
                <div className="flex items-center justify-between">
                  <div className="flex-1">
                    <h3 className="text-lg font-semibold text-white mb-2">Delete this transaction</h3>
                    <p className="text-[#a0a0a0]">Once you delete this transaction, there is no going back. Please be certain.</p>
                  </div>
                  <div className="flex items-center space-x-3">
                    <button
                      onClick={hideDeleteConfirmation}
                      className="px-4 py-2 text-[#a0a0a0] bg-[#333333]/50 rounded-xl hover:bg-[#333333]/70 transition-colors cursor-pointer"
                    >
                      Cancel
                    </button>
                    <button
                      onClick={confirmDelete}
                      className="px-4 py-2 bg-red-600 text-white rounded-xl hover:bg-red-700 transition-colors cursor-pointer font-medium"
                    >
                      Delete this transaction
                    </button>
                  </div>
                </div>
                <div className="w-full h-0.5 bg-red-500/50 mt-4 rounded-full"></div>
              </div>
            )}

            {/* Filters and Search */}
            <div className="bg-[#1a1a1a]/80 backdrop-blur-xl rounded-2xl border border-[#333333]/50 p-6 mb-6">
              <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
                <div>
                  <label className="block text-sm font-medium text-[#e0e0e0] mb-2">Search</label>
                  <div className="relative">
                    <input
                      type="text"
                      value={searchTerm}
                      onChange={(e) => setSearchTerm(e.target.value)}
                      placeholder="Search transactions..."
                      className="w-full px-4 py-2 pl-10 bg-[#0d0d0d]/60 border border-[#333333] rounded-xl text-white focus:outline-none focus:ring-2 focus:ring-[#FFD700] focus:border-[#FFD700]"
                    />
                    <i className="ri-search-line absolute left-3 top-1/2 transform -translate-y-1/2 text-[#a0a0a0]"></i>
                  </div>
                </div>

                <div>
                  <label className="block text-sm font-medium text-[#e0e0e0] mb-2">Filter by Type</label>
                  <div className="relative">
                    <select
                      value={filter}
                      onChange={(e) => setFilter(e.target.value)}
                      className="w-full px-3 py-2 pr-8 bg-[#0d0d0d]/60 border border-[#333333] rounded-xl text-white focus:outline-none focus:ring-2 focus:ring-[#FFD700] focus:border-[#FFD700] appearance-none"
                    >
                      <option value="all">All Types</option>
                      <option value="income">Income</option>
                      <option value="expense">Expense</option>
                    </select>
                    <i className="ri-arrow-down-s-line absolute right-3 top-1/2 transform -translate-y-1/2 text-[#a0a0a0] pointer-events-none"></i>
                  </div>
                </div>

                <div>
                  <label className="block text-sm font-medium text-[#e0e0e0] mb-2">Sort by</label>
                  <div className="relative">
                    <select
                      value={sortBy}
                      onChange={(e) => setSortBy(e.target.value)}
                      className="w-full px-3 py-2 pr-8 bg-[#0d0d0d]/60 border border-[#333333] rounded-xl text-white focus:outline-none focus:ring-2 focus:ring-[#FFD700] focus:border-[#FFD700] appearance-none"
                    >
                      <option value="date">Date</option>
                      <option value="month">Month</option>
                      <option value="year">Year</option>
                      <option value="amount">Amount</option>
                      <option value="category">Category</option>
                    </select>
                    <i className="ri-arrow-down-s-line absolute right-3 top-1/2 transform -translate-y-1/2 text-[#a0a0a0] pointer-events-none"></i>
                  </div>
                </div>

                <div className="flex items-end">
                  <Link href="/add" className="w-full">
                    <button className="w-full px-4 py-2 bg-gradient-to-r from-[#FFD700] to-[#00FFFF] text-[#0d0d0d] rounded-xl font-bold hover:shadow-lg hover:shadow-[#FFD700]/25 transition-all duration-300 cursor-pointer whitespace-nowrap">
                      Add Transaction
                    </button>
                  </Link>
                </div>
              </div>
            </div>

            {/* Transaction Statistics */}
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4 md:gap-6 mb-6">
              <div className="bg-[#1a1a1a]/80 backdrop-blur-xl rounded-2xl border border-[#333333]/50 p-6">
                <div className="flex items-center justify-between mb-4">
                  <h3 className="text-lg font-semibold text-white">Total Income</h3>
                  <div className="w-10 h-10 bg-gradient-to-r from-green-500/20 to-green-600/20 rounded-full flex items-center justify-center">
                    <i className="ri-arrow-up-line text-green-400 text-xl"></i>
                  </div>
                </div>
                <p className="text-2xl font-bold text-green-400">
                  ${transactions.filter(t => t.type === 'income').reduce((sum, t) => sum + t.amount, 0).toFixed(2)}
                </p>
              </div>

              <div className="bg-[#1a1a1a]/80 backdrop-blur-xl rounded-2xl border border-[#333333]/50 p-6">
                <div className="flex items-center justify-between mb-4">
                  <h3 className="text-lg font-semibold text-white">Total Expenses</h3>
                  <div className="w-10 h-10 bg-gradient-to-r from-red-500/20 to-red-600/20 rounded-full flex items-center justify-center">
                    <i className="ri-arrow-down-line text-red-400 text-xl"></i>
                  </div>
                </div>
                <p className="text-2xl font-bold text-red-400">
                  ${transactions.filter(t => t.type === 'expense').reduce((sum, t) => sum + t.amount, 0).toFixed(2)}
                </p>
              </div>

              <div className="bg-[#1a1a1a]/80 backdrop-blur-xl rounded-2xl border border-[#333333]/50 p-6">
                <div className="flex items-center justify-between mb-4">
                  <h3 className="text-lg font-semibold text-white">Net Balance</h3>
                  <div className="w-10 h-10 bg-gradient-to-r from-[#FFD700]/20 to-[#FFA500]/20 rounded-full flex items-center justify-center">
                    <i className="ri-wallet-line text-[#FFD700] text-xl"></i>
                  </div>
                </div>
                <p className="text-2xl font-bold text-[#FFD700]">
                  ${(transactions.filter(t => t.type === 'income').reduce((sum, t) => sum + t.amount, 0) - 
                     transactions.filter(t => t.type === 'expense').reduce((sum, t) => sum + t.amount, 0)).toFixed(2)}
                </p>
              </div>
            </div>

            {/* Transactions List */}
            <div className="bg-[#1a1a1a]/80 backdrop-blur-xl rounded-2xl border border-[#333333]/50 p-6">
              <div className="flex items-center justify-between mb-6">
                <h3 className="text-lg font-semibold text-white">Transaction History</h3>
                <span className="text-[#808080] text-sm">{sortedTransactions.length} transactions</span>
              </div>

              {loading ? (
                <div className="flex justify-center items-center h-64">
                  <LoadingSpinner size="lg" />
                </div>
              ) : (
                <div className="space-y-3">
                  {sortedTransactions.map((transaction, index) => (
                    <div
                      key={transaction.id || `transaction-${index}-${transaction.title}-${transaction.date}`}
                      className="flex items-center justify-between p-4 bg-[#0d0d0d]/30 rounded-xl hover:bg-[#0d0d0d]/50 transition-all duration-300"
                    >
                      <div className="flex items-center space-x-4">
                        <div className={`w-12 h-12 rounded-full flex items-center justify-center ${
                          transaction.type === 'income' ? 'bg-green-500/20' : 'bg-red-500/20'
                        }`}>
                          <i className={`${getCategoryIcon(transaction.category)} ${getTypeColor(transaction.type)} text-xl`}></i>
                        </div>
                        <div>
                          <h4 className="text-white font-medium">{transaction.title}</h4>
                          <div className="flex items-center space-x-2 text-sm text-[#a0a0a0]">
                            <span>{transaction.category}</span>
                            <span>•</span>
                            <span>{transaction.date}</span>
                          </div>
                        </div>
                      </div>

                      <div className="flex items-center space-x-4">
                        <div className="text-right">
                          <p className={`font-bold text-lg ${getTypeColor(transaction.type)}`}>
                            {transaction.type === 'income' ? '+' : '-'}${transaction.amount.toFixed(2)}
                          </p>
                        </div>
                        <button
                          onClick={(e) => {
                            e.preventDefault();
                            e.stopPropagation();
                            console.log('Delete button clicked for transaction:', transaction);
                            console.log('Transaction ID:', transaction.id);
                            handleDeleteTransaction(transaction.id);
                          }}
                          className="p-2 text-red-400 hover:bg-red-400/20 rounded-lg transition-colors cursor-pointer"
                        >
                          <i className="ri-delete-bin-line text-sm"></i>
                        </button>
                      </div>
                    </div>
                  ))}

                  {sortedTransactions.length === 0 && (
                    <div className="text-center py-12">
                      <div className="w-16 h-16 bg-[#333333]/30 rounded-full flex items-center justify-center mx-auto mb-4">
                        <i className="ri-file-list-line text-[#808080] text-2xl"></i>
                      </div>
                      <p className="text-[#808080] text-lg">No transactions found</p>
                      <p className="text-[#808080] text-sm mt-2">Try adjusting your search or filter criteria</p>
                    </div>
                  )}
                </div>
              )}
            </div>
          </main>
        </div>
      </div>
    </ProtectedRoute>
  );
}