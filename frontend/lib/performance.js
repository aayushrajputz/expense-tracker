// Performance monitoring and optimization utilities
class PerformanceMonitor {
  constructor() {
    this.metrics = new Map();
    this.observers = new Map();
    this.isEnabled = typeof window !== 'undefined' && process.env.NODE_ENV === 'development';
  }

  // Start timing an operation
  startTimer(operation) {
    if (!this.isEnabled) return;
    
    const startTime = performance.now();
    this.metrics.set(operation, { startTime, endTime: null, duration: null });
    
    return () => this.endTimer(operation);
  }

  // End timing an operation
  endTimer(operation) {
    if (!this.isEnabled) return;
    
    const metric = this.metrics.get(operation);
    if (metric) {
      metric.endTime = performance.now();
      metric.duration = metric.endTime - metric.startTime;
      
      // Log performance data
      console.log(`⏱️ ${operation}: ${metric.duration.toFixed(2)}ms`);
      
      // Store for analysis
      this.storeMetric(operation, metric.duration);
    }
  }

  // Measure async operation
  async measureAsync(operation, asyncFn) {
    const endTimer = this.startTimer(operation);
    try {
      const result = await asyncFn();
      endTimer();
      return result;
    } catch (error) {
      endTimer();
      throw error;
    }
  }

  // Store metric for analysis
  storeMetric(operation, duration) {
    if (!this.observers.has(operation)) {
      this.observers.set(operation, []);
    }
    
    this.observers.get(operation).push({
      timestamp: Date.now(),
      duration
    });
    
    // Keep only last 100 measurements
    const measurements = this.observers.get(operation);
    if (measurements.length > 100) {
      measurements.shift();
    }
  }

  // Get performance statistics
  getStats(operation) {
    const measurements = this.observers.get(operation) || [];
    if (measurements.length === 0) return null;
    
    const durations = measurements.map(m => m.duration);
    const avg = durations.reduce((a, b) => a + b, 0) / durations.length;
    const min = Math.min(...durations);
    const max = Math.max(...durations);
    
    return {
      operation,
      count: measurements.length,
      average: avg.toFixed(2),
      min: min.toFixed(2),
      max: max.toFixed(2),
      last: durations[durations.length - 1].toFixed(2)
    };
  }

  // Get all performance statistics
  getAllStats() {
    const stats = [];
    for (const operation of this.observers.keys()) {
      const stat = this.getStats(operation);
      if (stat) stats.push(stat);
    }
    return stats.sort((a, b) => parseFloat(b.average) - parseFloat(a.average));
  }

  // Monitor component render performance
  monitorComponent(componentName) {
    if (!this.isEnabled) return () => {};
    
    return this.startTimer(`Component Render: ${componentName}`);
  }

  // Monitor API call performance
  monitorAPICall(endpoint) {
    if (!this.isEnabled) return () => {};
    
    return this.startTimer(`API Call: ${endpoint}`);
  }

  // Monitor cache hit/miss ratio
  monitorCache(cacheName) {
    if (!this.isEnabled) return { hit: () => {}, miss: () => {} };
    
    let hits = 0;
    let misses = 0;
    
    return {
      hit: () => {
        hits++;
        this.logCacheStats(cacheName, hits, misses);
      },
      miss: () => {
        misses++;
        this.logCacheStats(cacheName, hits, misses);
      }
    };
  }

  // Log cache statistics
  logCacheStats(cacheName, hits, misses) {
    const total = hits + misses;
    const hitRate = total > 0 ? ((hits / total) * 100).toFixed(1) : 0;
    console.log(`📦 ${cacheName} Cache: ${hits} hits, ${misses} misses (${hitRate}% hit rate)`);
  }

  // Monitor memory usage
  monitorMemory() {
    if (!this.isEnabled || !performance.memory) return;
    
    const memory = performance.memory;
    console.log(`💾 Memory: ${(memory.usedJSHeapSize / 1024 / 1024).toFixed(2)}MB used / ${(memory.totalJSHeapSize / 1024 / 1024).toFixed(2)}MB total`);
  }

  // Monitor network performance
  monitorNetwork() {
    if (!this.isEnabled || !navigator.connection) return;
    
    const connection = navigator.connection;
    console.log(`🌐 Network: ${connection.effectiveType} (${connection.downlink}Mbps)`);
  }

  // Debounce function for performance
  debounce(func, wait) {
    let timeout;
    return function executedFunction(...args) {
      const later = () => {
        clearTimeout(timeout);
        func(...args);
      };
      clearTimeout(timeout);
      timeout = setTimeout(later, wait);
    };
  }

  // Throttle function for performance
  throttle(func, limit) {
    let inThrottle;
    return function executedFunction(...args) {
      if (!inThrottle) {
        func.apply(this, args);
        inThrottle = true;
        setTimeout(() => inThrottle = false, limit);
      }
    };
  }

  // Optimize images
  optimizeImages() {
    if (!this.isEnabled || typeof window === 'undefined') return;
    
    const images = document.querySelectorAll('img');
    images.forEach(img => {
      // Add loading="lazy" for better performance
      if (!img.loading) {
        img.loading = 'lazy';
      }
      
      // Add error handling
      img.onerror = () => {
        console.warn(`Failed to load image: ${img.src}`);
      };
    });
  }

  // Monitor scroll performance
  monitorScroll(callback) {
    if (!this.isEnabled) return;
    
    let ticking = false;
    const scrollHandler = () => {
      if (!ticking) {
        requestAnimationFrame(() => {
          callback();
          ticking = false;
        });
        ticking = true;
      }
    };
    
    window.addEventListener('scroll', scrollHandler, { passive: true });
    
    return () => window.removeEventListener('scroll', scrollHandler);
  }

  // Generate performance report
  generateReport() {
    if (!this.isEnabled) return null;
    
    return {
      timestamp: new Date().toISOString(),
      stats: this.getAllStats(),
      memory: performance.memory ? {
        used: (performance.memory.usedJSHeapSize / 1024 / 1024).toFixed(2),
        total: (performance.memory.totalJSHeapSize / 1024 / 1024).toFixed(2)
      } : null,
      network: navigator.connection ? {
        type: navigator.connection.effectiveType,
        speed: navigator.connection.downlink
      } : null
    };
  }
}

// Create global performance monitor instance
export const performanceMonitor = new PerformanceMonitor();

// Performance optimization utilities
export const performanceUtils = {
  // Debounce utility
  debounce: (func, wait) => performanceMonitor.debounce(func, wait),
  
  // Throttle utility
  throttle: (func, limit) => performanceMonitor.throttle(func, limit),
  
  // Measure async operation
  measureAsync: (operation, asyncFn) => performanceMonitor.measureAsync(operation, asyncFn),
  
  // Monitor component
  monitorComponent: (componentName) => performanceMonitor.monitorComponent(componentName),
  
  // Monitor API call
  monitorAPICall: (endpoint) => performanceMonitor.monitorAPICall(endpoint),
  
  // Monitor cache
  monitorCache: (cacheName) => performanceMonitor.monitorCache(cacheName),
  
  // Generate report
  generateReport: () => performanceMonitor.generateReport()
};

export default performanceMonitor;