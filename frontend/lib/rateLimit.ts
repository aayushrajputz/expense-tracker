interface RateLimitEntry {
  count: number;
  resetTime: number;
}

class RateLimiter {
  private limits = new Map<string, RateLimitEntry>();

  limitPerWindow(key: string, max: number, windowMs: number): boolean {
    const now = Date.now();
    const entry = this.limits.get(key);

    if (!entry || now > entry.resetTime) {
      // First request or window expired
      this.limits.set(key, {
        count: 1,
        resetTime: now + windowMs
      });
      return true; // Allow request
    }

    if (entry.count >= max) {
      return false; // Rate limit exceeded
    }

    // Increment counter
    entry.count++;
    return true; // Allow request
  }

  getRemainingTime(key: string): number {
    const entry = this.limits.get(key);
    if (!entry) return 0;
    return Math.max(0, entry.resetTime - Date.now());
  }

  // Clean up old entries
  cleanup(): void {
    const now = Date.now();
    for (const [key, entry] of this.limits.entries()) {
      if (now > entry.resetTime) {
        this.limits.delete(key);
      }
    }
  }
}

export const rateLimiter = new RateLimiter();

// Clean up old rate limit entries every minute
if (typeof window === 'undefined') {
  setInterval(() => {
    rateLimiter.cleanup();
  }, 60 * 1000);
}