
import axios from 'axios';
import { expenseCache, summaryCache, userCache, cacheUtils } from './cache.js';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'https://expense-tracker-tzun.onrender.com';

// Request deduplication map
const pendingRequests = new Map();

const api = axios.create({
  baseURL: API_BASE_URL,
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Add token to requests if available
api.interceptors.request.use(
  (config) => {
    if (typeof window !== 'undefined') {
      const token = localStorage.getItem('token');
      if (token) {
        config.headers.Authorization = `Bearer ${token}`;
      }
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      if (typeof window !== 'undefined') {
        localStorage.removeItem('token');
        window.location.href = '/login';
      }
    }
    
    // Handle network errors more gracefully
    if (error.code === 'ERR_NETWORK' || error.message === 'Network Error') {
      console.error('Network Error: Backend server is not accessible');
      error.message = 'Unable to connect to server. Please check if the backend is running.';
    }
    
    console.error('API Error:', error);
    return Promise.reject(error);
  }
);

// Request deduplication utility
const deduplicateRequest = (key, requestFn) => {
  if (pendingRequests.has(key)) {
    return pendingRequests.get(key);
  }
  
  const promise = requestFn().finally(() => {
    pendingRequests.delete(key);
  });
  
  pendingRequests.set(key, promise);
  return promise;
};

export const authAPI = {
  signup: async (name, email, password, budget = 50000) => {
    try {
      const response = await api.post('/api/signup', { name, email, password, budget });
      return response.data;
    } catch (error) {
      if (error.code === 'ERR_NETWORK') {
        throw new Error('Backend server is not running. Please start your Golang backend server.');
      }
      throw error;
    }
  },

  verifyOtp: async (email, otp) => {
    try {
      const response = await api.post('/api/verify-otp', { email, otp });
      return response.data;
    } catch (error) {
      if (error.code === 'ERR_NETWORK') {
        throw new Error('Backend server is not running. Please start your Golang backend server.');
      }
      throw error;
    }
  },

  login: async (email, password) => {
    try {
      const response = await api.post('/api/login', { email, password });
      return response.data;
    } catch (error) {
      if (error.code === 'ERR_NETWORK') {
        throw new Error('Backend server is not running. Please start your Golang backend server.');
      }
      throw error;
    }
  },
};

export const expenseAPI = {
  getExpenses: async (limit = null) => {
    const cacheKey = cacheUtils.getExpenseKey(limit);
    
    return cacheUtils.getCachedOrFetch(expenseCache, cacheKey, async () => {
      try {
        const url = limit ? `/api/expenses?limit=${limit}` : '/api/expenses';
        const response = await deduplicateRequest(`getExpenses:${limit}`, () => api.get(url));
        return response.data;
      } catch (error) {
        if (error.code === 'ERR_NETWORK') {
          throw new Error('Backend server is not running. Please start your Golang backend server.');
        }
        throw error;
      }
    });
  },

  addExpense: async (expense) => {
    try {
      const response = await api.post('/api/expenses', expense);
      // Invalidate expense cache after adding new expense
      expenseCache.clear();
      return response.data;
    } catch (error) {
      if (error.code === 'ERR_NETWORK') {
        throw new Error('Backend server is not running. Please start your Golang backend server.');
      }
      throw error;
    }
  },

  updateExpense: async (id, expense) => {
    try {
      const response = await api.put(`/api/expenses/${id}`, expense);
      // Invalidate expense cache after updating
      expenseCache.clear();
      return response.data;
    } catch (error) {
      if (error.code === 'ERR_NETWORK') {
        throw new Error('Backend server is not running. Please start your Golang backend server.');
      }
      throw error;
    }
  },

  deleteExpense: async (id) => {
    try {
      const response = await api.delete(`/api/expenses/${id}`);
      // Invalidate expense cache after deleting
      expenseCache.clear();
      return response.data;
    } catch (error) {
      if (error.code === 'ERR_NETWORK') {
        throw new Error('Backend server is not running. Please start your Golang backend server.');
      }
      throw error;
    }
  },

  getExpensesByDateRange: async (startDate, endDate) => {
    const cacheKey = `expenses_range:${cacheUtils.getCurrentUserId()}:${startDate}:${endDate}`;
    
    return cacheUtils.getCachedOrFetch(expenseCache, cacheKey, async () => {
      try {
        const response = await deduplicateRequest(`getExpensesByDateRange:${startDate}:${endDate}`, () => 
          api.get(`/api/expenses/range?start_date=${startDate}&end_date=${endDate}`)
        );
        return response.data;
      } catch (error) {
        if (error.code === 'ERR_NETWORK') {
          throw new Error('Backend server is not running. Please start your Golang backend server.');
        }
        throw error;
      }
    });
  },

  getExpensesByCategory: async (category) => {
    const cacheKey = `expenses_category:${cacheUtils.getCurrentUserId()}:${category}`;
    
    return cacheUtils.getCachedOrFetch(expenseCache, cacheKey, async () => {
      try {
        const response = await deduplicateRequest(`getExpensesByCategory:${category}`, () => 
          api.get(`/api/expenses/category/${encodeURIComponent(category)}`)
        );
        return response.data;
      } catch (error) {
        if (error.code === 'ERR_NETWORK') {
          throw new Error('Backend server is not running. Please start your Golang backend server.');
        }
        throw error;
      }
    });
  },

  recategorizeExpenses: async () => {
    try {
      const response = await api.post('/api/expenses/recategorize');
      // Invalidate all caches after recategorization
      expenseCache.clear();
      summaryCache.clear();
      return response.data;
    } catch (error) {
      if (error.code === 'ERR_NETWORK') {
        throw new Error('Backend server is not running. Please start your Golang backend server.');
      }
      throw error;
    }
  },

  getAnalytics: async () => {
    const cacheKey = `analytics:${cacheUtils.getCurrentUserId()}`;
    
    return cacheUtils.getCachedOrFetch(expenseCache, cacheKey, async () => {
      try {
        const response = await deduplicateRequest('getAnalytics', () => api.get('/api/analytics'));
        return response.data;
      } catch (error) {
        if (error.code === 'ERR_NETWORK') {
          // Return mock analytics data for development
          return {
            insights: [
              {
                type: 'trend',
                title: 'Spending Trend',
                content: 'Your spending has increased by 15% this month compared to last month.'
              },
              {
                type: 'advice',
                title: 'Budget Recommendation',
                content: 'Consider reducing dining out expenses to save $200 monthly.'
              },
              {
                type: 'prediction',
                title: 'Monthly Forecast',
                content: 'Based on current patterns, you might spend $1,200 this month.'
              }
            ]
          };
        }
        throw error;
      }
    });
  },

  // New AI-powered insights endpoint
  getAIInsights: async () => {
    const cacheKey = `ai_insights:${cacheUtils.getCurrentUserId()}`;
    
    return cacheUtils.getCachedOrFetch(expenseCache, cacheKey, async () => {
      try {
        const response = await deduplicateRequest('getAIInsights', () => api.get('/api/ai-insights'));
        return response.data;
      } catch (error) {
        if (error.code === 'ERR_NETWORK') {
          throw new Error('Backend server is not running. Please start your Golang backend server.');
        }
        throw error;
      }
    });
  },

  getCategories: async () => {
    const cacheKey = `categories:${cacheUtils.getCurrentUserId()}`;
    
    return cacheUtils.getCachedOrFetch(expenseCache, cacheKey, async () => {
      try {
        const response = await deduplicateRequest('getCategories', () => api.get('/api/categories'));
        // Backend returns array directly, so we need to wrap it in an object
        return { categories: response.data };
      } catch (error) {
        if (error.code === 'ERR_NETWORK') {
          // Return mock categories for development
          return {
            categories: [
              'Food',
              'Transportation',
              'Entertainment',
              'Shopping',
              'Bills',
              'Healthcare',
              'Income',
              'Other'
            ]
          };
        }
        throw error;
      }
    });
  },

  // New API functions for profile data
  getProfileData: async () => {
    const cacheKey = cacheUtils.getUserKey('profile');
    
    return cacheUtils.getCachedOrFetch(userCache, cacheKey, async () => {
      try {
        const response = await deduplicateRequest('getProfileData', () => api.get('/api/profile'));
        return response.data;
      } catch (error) {
        if (error.code === 'ERR_NETWORK') {
          throw new Error('Backend server is not running. Please start your Golang backend server.');
        }
        throw error;
      }
    });
  },

  getFinancialSummary: async (budget = null) => {
    const cacheKey = cacheUtils.getSummaryKey('monthly', { budget });
    
    return cacheUtils.getCachedOrFetch(summaryCache, cacheKey, async () => {
      try {
        const url = budget ? `/api/summary?budget=${budget}` : '/api/summary';
        const response = await deduplicateRequest(`getFinancialSummary:${budget}`, () => api.get(url));
        return response.data;
      } catch (error) {
        if (error.code === 'ERR_NETWORK') {
          throw new Error('Backend server is not running. Please start your Golang backend server.');
        }
        throw error;
      }
    });
  },

  getLifetimeSummary: async () => {
    const cacheKey = cacheUtils.getSummaryKey('lifetime');
    
    return cacheUtils.getCachedOrFetch(summaryCache, cacheKey, async () => {
      try {
        const response = await deduplicateRequest('getLifetimeSummary', () => api.get('/api/summary/lifetime'));
        return response.data;
      } catch (error) {
        if (error.code === 'ERR_NETWORK') {
          throw new Error('Backend server is not running. Please start your Golang backend server.');
        }
        throw error;
      }
    });
  },

  getCategoryBreakdown: async (startDate, endDate) => {
    const cacheKey = cacheUtils.getSummaryKey('category_breakdown', { startDate, endDate });
    
    return cacheUtils.getCachedOrFetch(summaryCache, cacheKey, async () => {
      try {
        const response = await deduplicateRequest(`getCategoryBreakdown:${startDate}:${endDate}`, () => 
          api.get(`/api/summary/category-breakdown?start_date=${startDate}&end_date=${endDate}`)
        );
        return response.data;
      } catch (error) {
        if (error.code === 'ERR_NETWORK') {
          throw new Error('Backend server is not running. Please start your Golang backend server.');
        }
        throw error;
      }
    });
  },

  getRecentTransactions: async (limit = 10) => {
    return expenseAPI.getExpenses(limit);
  },
};

// Utility function to clear all caches
export const clearAllCaches = () => {
  expenseCache.clear();
  summaryCache.clear();
  userCache.clear();
};

export default api;
