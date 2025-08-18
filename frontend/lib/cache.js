// Frontend LRU Cache implementation for performance optimization
class LRUCache {
  constructor(capacity, expirationMinutes = 15) {
    this.capacity = capacity;
    this.expirationMinutes = expirationMinutes;
    this.cache = new Map();
    this.accessOrder = [];
  }

  // Get item from cache
  get(key) {
    if (this.cache.has(key)) {
      const item = this.cache.get(key);
      
      // Check if item has expired
      if (Date.now() > item.expiration) {
        this.delete(key);
        return null;
      }
      
      // Move to front (most recently used)
      this.moveToFront(key);
      return item.value;
    }
    return null;
  }

  // Set item in cache
  set(key, value) {
    const expiration = Date.now() + (this.expirationMinutes * 60 * 1000);
    
    if (this.cache.has(key)) {
      // Update existing item
      this.cache.set(key, { value, expiration });
      this.moveToFront(key);
    } else {
      // Add new item
      this.cache.set(key, { value, expiration });
      this.accessOrder.unshift(key);
      
      // Remove least recently used if capacity exceeded
      if (this.cache.size > this.capacity) {
        this.removeLRU();
      }
    }
  }

  // Delete item from cache
  delete(key) {
    if (this.cache.has(key)) {
      this.cache.delete(key);
      const index = this.accessOrder.indexOf(key);
      if (index > -1) {
        this.accessOrder.splice(index, 1);
      }
    }
  }

  // Clear all items from cache
  clear() {
    this.cache.clear();
    this.accessOrder = [];
  }

  // Get cache size
  size() {
    return this.cache.size;
  }

  // Move item to front of access order
  moveToFront(key) {
    const index = this.accessOrder.indexOf(key);
    if (index > -1) {
      this.accessOrder.splice(index, 1);
      this.accessOrder.unshift(key);
    }
  }

  // Remove least recently used item
  removeLRU() {
    if (this.accessOrder.length > 0) {
      const lruKey = this.accessOrder.pop();
      this.cache.delete(lruKey);
    }
  }

  // Clean up expired items
  cleanup() {
    const now = Date.now();
    for (const [key, item] of this.cache.entries()) {
      if (now > item.expiration) {
        this.delete(key);
      }
    }
  }

  // Start automatic cleanup
  startCleanup(intervalMinutes = 5) {
    setInterval(() => {
      this.cleanup();
    }, intervalMinutes * 60 * 1000);
  }
}

// Create global cache instances
export const expenseCache = new LRUCache(100, 10); // 100 items, 10 minutes
export const summaryCache = new LRUCache(50, 5);   // 50 items, 5 minutes
export const userCache = new LRUCache(10, 30);     // 10 items, 30 minutes

// Start cleanup for all caches
expenseCache.startCleanup(5);
summaryCache.startCleanup(5);
userCache.startCleanup(5);

// Cache utility functions
export const cacheUtils = {
  // Get current user ID from localStorage
  getCurrentUserId: () => {
    if (typeof window !== 'undefined') {
      const user = localStorage.getItem('user');
      if (user) {
        try {
          const userData = JSON.parse(user);
          return userData.id || userData.email || 'default';
        } catch (e) {
          return 'default';
        }
      }
    }
    return 'default';
  },

  // Generate cache key for expenses with user ID
  getExpenseKey: (limit = null) => {
    const userId = cacheUtils.getCurrentUserId();
    return `expenses:${userId}:${limit || 'all'}`;
  },

  // Generate cache key for summary with user ID
  getSummaryKey: (type = 'monthly', params = {}) => {
    const userId = cacheUtils.getCurrentUserId();
    const paramStr = Object.keys(params).length > 0 
      ? ':' + Object.entries(params).map(([k, v]) => `${k}=${v}`).join(':')
      : '';
    return `summary:${userId}:${type}${paramStr}`;
  },

  // Generate cache key for user data with user ID
  getUserKey: (type = 'profile') => {
    const userId = cacheUtils.getCurrentUserId();
    return `user:${userId}:${type}`;
  },

  // Invalidate all cache entries for current user
  invalidateUserCache: () => {
    // Clear all caches when user data changes
    expenseCache.clear();
    summaryCache.clear();
    userCache.clear();
  },

  // Get cached data with fallback
  getCachedOrFetch: async (cache, key, fetchFunction) => {
    const cached = cache.get(key);
    if (cached) {
      return cached;
    }
    
    const data = await fetchFunction();
    cache.set(key, data);
    return data;
  }
};

export default LRUCache;