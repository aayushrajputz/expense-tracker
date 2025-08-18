
'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import Sidebar from '../../components/Sidebar';
import Navbar from '../../components/Navbar';
import LoadingSpinner from '../../components/LoadingSpinner';
import ErrorMessage from '../../components/ErrorMessage';
import { expenseAPI } from '../../lib/api';
import ProtectedRoute from '../../components/ProtectedRoute';

export default function AddExpense() {
  const [formData, setFormData] = useState({
    type: 'expense',
    amount: '',
    title: '',
    category: '',
    date: new Date().toISOString().split('T')[0]
  });
  const [categories, setCategories] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const router = useRouter();

  useEffect(() => {
    fetchCategories();
  }, []);

  const fetchCategories = async () => {
    try {
      const data = await expenseAPI.getCategories();
      setCategories(data.categories || []);
    } catch (err) {
      console.error('Failed to fetch categories:', err);
    }
  };

  const handleSubmit = async (e) => {
    e.preventDefault();

    if (!formData.amount || !formData.title || !formData.category) {
      setError('Please fill in all required fields');
      return;
    }

    try {
      setLoading(true);
      setError(null);

      const expenseData = {
        ...formData,
        amount: parseFloat(formData.amount)
      };

      await expenseAPI.addExpense(expenseData);
      router.push('/');
    } catch (err) {
      setError(err.response?.data?.message || 'Failed to add transaction');
    } finally {
      setLoading(false);
    }
  };

  const handleChange = (e) => {
    const { name, value } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: value
    }));
  };

  return (
    <ProtectedRoute>
      <div className="min-h-screen bg-gradient-to-br from-[#0d0d0d] via-[#1a1a1a] to-[#0d0d0d]">
        <Sidebar isOpen={sidebarOpen} onClose={() => setSidebarOpen(false)} />
        <div className="flex-1 flex flex-col">
          <Navbar onSidebarToggle={() => setSidebarOpen(!sidebarOpen)} />
          <main className="p-4 md:p-6">
            <div className="mb-6 md:mb-8">
              <h1 className="text-2xl md:text-3xl font-bold bg-gradient-to-r from-[#FFD700] to-[#00FFFF] bg-clip-text text-transparent mb-2">
                Add Transaction
              </h1>
              <p className="text-[#a0a0a0]">Record a new expense or income</p>
            </div>

            <div className="max-w-2xl mx-auto">
              <div className="bg-[#1a1a1a]/80 backdrop-blur-xl rounded-2xl border border-[#333333]/50 p-6 md:p-8">
                {error && (
                  <div className="mb-6">
                    <ErrorMessage message={error} />
                  </div>
                )}

                <form onSubmit={handleSubmit} className="space-y-6">
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                    <div>
                      <label className="block text-sm font-medium text-[#e0e0e0] mb-3">
                        Type
                      </label>
                      <div className="flex space-x-4">
                        <label className="flex items-center cursor-pointer">
                          <input
                            type="radio"
                            name="type"
                            value="expense"
                            checked={formData.type === 'expense'}
                            onChange={handleChange}
                            className="mr-2 text-[#FFD700] bg-[#0d0d0d] border-[#333333] focus:ring-[#FFD700]"
                          />
                          <span className="text-[#e0e0e0]">Expense</span>
                        </label>
                        <label className="flex items-center cursor-pointer">
                          <input
                            type="radio"
                            name="type"
                            value="income"
                            checked={formData.type === 'income'}
                            onChange={handleChange}
                            className="mr-2 text-[#FFD700] bg-[#0d0d0d] border-[#333333] focus:ring-[#FFD700]"
                          />
                          <span className="text-[#e0e0e0]">Income</span>
                        </label>
                      </div>
                    </div>

                    <div>
                      <label className="block text-sm font-medium text-[#e0e0e0] mb-3">
                        Amount *
                      </label>
                      <div className="relative">
                        <span className="absolute left-3 top-1/2 transform -translate-y-1/2 text-[#FFD700] font-bold">$</span>
                        <input
                          type="number"
                          name="amount"
                          value={formData.amount}
                          onChange={handleChange}
                          step="0.01"
                          min="0"
                          required
                          className="w-full pl-8 pr-3 py-3 bg-[#0d0d0d]/60 border border-[#333333] rounded-xl text-white focus:outline-none focus:ring-2 focus:ring-[#FFD700] focus:border-[#FFD700] transition-all duration-300"
                          placeholder="0.00"
                        />
                      </div>
                    </div>
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-[#e0e0e0] mb-3">
                      Description *
                    </label>
                    <input
                      type="text"
                      name="title"
                      value={formData.title}
                      onChange={handleChange}
                      required
                      className="w-full px-4 py-3 bg-[#0d0d0d]/60 border border-[#333333] rounded-xl text-white focus:outline-none focus:ring-2 focus:ring-[#FFD700] focus:border-[#FFD700] transition-all duration-300"
                      placeholder="Enter description"
                    />
                  </div>

                  <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                    <div>
                      <label className="block text-sm font-medium text-[#e0e0e0] mb-3">
                        Category *
                      </label>
                      <div className="relative">
                        <select
                          name="category"
                          value={formData.category}
                          onChange={handleChange}
                          required
                          className="w-full px-4 py-3 pr-10 bg-[#0d0d0d]/60 border border-[#333333] rounded-xl text-white focus:outline-none focus:ring-2 focus:ring-[#FFD700] focus:border-[#FFD700] transition-all duration-300 appearance-none"
                        >
                          <option value="">Select category</option>
                          {categories.map((category) => (
                            <option key={category} value={category}>
                              {category}
                            </option>
                          ))}
                        </select>
                        <i className="ri-arrow-down-s-line absolute right-3 top-1/2 transform -translate-y-1/2 text-[#a0a0a0] pointer-events-none"></i>
                      </div>
                    </div>

                    <div>
                      <label className="block text-sm font-medium text-[#e0e0e0] mb-3">
                        Date *
                      </label>
                      <input
                        type="date"
                        name="date"
                        value={formData.date}
                        onChange={handleChange}
                        required
                        className="w-full px-4 py-3 bg-[#0d0d0d]/60 border border-[#333333] rounded-xl text-white focus:outline-none focus:ring-2 focus:ring-[#FFD700] focus:border-[#FFD700] transition-all duration-300"
                      />
                    </div>
                  </div>

                  <div className="flex flex-col md:flex-row justify-end space-y-4 md:space-y-0 md:space-x-4 pt-6">
                    <button
                      type="button"
                      onClick={() => router.push('/')}
                      className="px-6 py-3 text-[#a0a0a0] bg-[#333333]/50 rounded-xl hover:bg-[#333333]/70 transition-colors cursor-pointer whitespace-nowrap"
                    >
                      Cancel
                    </button>
                    <button
                      type="submit"
                      disabled={loading}
                      className="px-6 py-3 bg-gradient-to-r from-[#FFD700] to-[#00FFFF] text-[#0d0d0d] rounded-xl font-bold hover:shadow-lg hover:shadow-[#FFD700]/25 disabled:opacity-50 disabled:cursor-not-allowed transition-all duration-300 cursor-pointer whitespace-nowrap flex items-center justify-center space-x-2"
                    >
                      {loading ? (
                        <>
                          <LoadingSpinner size="sm" />
                          <span>Adding...</span>
                        </>
                      ) : (
                        <>
                          <i className="ri-add-line text-lg"></i>
                          <span>Add Transaction</span>
                        </>
                      )}
                    </button>
                  </div>
                </form>
              </div>
            </div>
          </main>
        </div>
      </div>
    </ProtectedRoute>
  );
}
