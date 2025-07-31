
'use client';

import { useState } from 'react';
import { useAuth } from '../../lib/AuthContext';
import Link from 'next/link';
import LoadingSpinner from '../../components/LoadingSpinner';

export default function LoginPage() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const { login } = useAuth();

  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);
    setError('');

    const result = await login(email, password);
    
    if (!result.success) {
      setError(result.error);
    }
    
    setLoading(false);
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-[#0d0d0d] via-[#1a1a1a] to-[#0d0d0d] relative overflow-hidden">
      {/* Animated background elements */}
      <div className="absolute inset-0 bg-gradient-to-r from-[#FFD700]/5 via-transparent to-[#00FFFF]/5 animate-pulse"></div>
      <div className="absolute top-1/4 left-1/4 w-96 h-96 bg-[#FFD700]/10 rounded-full blur-3xl animate-pulse"></div>
      <div className="absolute bottom-1/4 right-1/4 w-96 h-96 bg-[#00FFFF]/10 rounded-full blur-3xl animate-pulse"></div>
      
      <div className="max-w-md w-full space-y-8 p-10 bg-[#1a1a1a]/80 backdrop-blur-xl rounded-3xl border border-[#333333]/50 shadow-2xl relative z-10">
        {/* Company branding */}
        <div className="text-center">
          <div className="relative">
            <h1 className="text-5xl font-bold bg-gradient-to-r from-[#FFD700] via-[#FFA500] to-[#00FFFF] bg-clip-text text-transparent mb-4 tracking-wide" 
                style={{ fontFamily: 'Orbitron, monospace' }}>
              BucksInfo
            </h1>
            <div className="absolute -bottom-2 left-1/2 transform -translate-x-1/2 w-20 h-0.5 bg-gradient-to-r from-[#FFD700] to-[#00FFFF] rounded-full"></div>
            <div className="absolute -bottom-1 left-1/2 transform -translate-x-1/2 w-8 h-0.5 bg-gradient-to-r from-[#00FFFF] to-[#FFD700] rounded-full blur-sm"></div>
          </div>
          
          <div className="mt-8 mb-6">
            <h2 className="text-3xl font-bold text-white tracking-wide">
              Welcome Back
            </h2>
            <p className="mt-3 text-[#a0a0a0] text-lg">
              Access your premium dashboard
            </p>
          </div>
          
          <p className="text-[#808080]">
            New to BucksInfo?{' '}
            <Link href="/signup" className="font-semibold text-[#00FFFF] hover:text-[#FFD700] cursor-pointer transition-all duration-300 hover:underline">
              Create Account
            </Link>
          </p>
        </div>
        
        <form className="mt-10 space-y-6" onSubmit={handleSubmit}>
          <div className="space-y-5">
            <div>
              <label htmlFor="email" className="text-sm font-semibold text-[#e0e0e0] block mb-3 tracking-wide">
                EMAIL ADDRESS
              </label>
              <input
                id="email"
                name="email"
                type="email"
                required
                className="appearance-none rounded-2xl relative block w-full px-6 py-4 bg-[#0d0d0d]/60 border-2 border-[#333333] placeholder-[#808080] text-white focus:outline-none focus:ring-2 focus:ring-[#00FFFF] focus:border-[#00FFFF] transition-all duration-300 text-lg shadow-inner hover:border-[#FFD700]/50"
                placeholder="Enter your email address"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
              />
            </div>
            <div>
              <label htmlFor="password" className="text-sm font-semibold text-[#e0e0e0] block mb-3 tracking-wide">
                PASSWORD
              </label>
              <input
                id="password"
                name="password"
                type="password"
                required
                className="appearance-none rounded-2xl relative block w-full px-6 py-4 bg-[#0d0d0d]/60 border-2 border-[#333333] placeholder-[#808080] text-white focus:outline-none focus:ring-2 focus:ring-[#00FFFF] focus:border-[#00FFFF] transition-all duration-300 text-lg shadow-inner hover:border-[#FFD700]/50"
                placeholder="Enter your password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
              />
            </div>
          </div>

          {error && (
            <div className="bg-red-900/30 border-2 border-red-500/50 rounded-2xl p-4 backdrop-blur-sm">
              <div className="flex items-center">
                <div className="w-6 h-6 flex items-center justify-center">
                  <i className="ri-error-warning-line text-red-400 text-lg"></i>
                </div>
                <div className="ml-3">
                  <p className="text-red-300 font-medium">{error}</p>
                </div>
              </div>
            </div>
          )}

          <div className="pt-4">
            <button
              type="submit"
              disabled={loading}
              className="group relative w-full flex justify-center py-4 px-6 border-2 border-transparent text-lg font-bold rounded-2xl text-[#0d0d0d] bg-gradient-to-r from-[#FFD700] via-[#FFA500] to-[#FFD700] hover:from-[#00FFFF] hover:via-[#40E0D0] hover:to-[#00FFFF] focus:outline-none focus:ring-4 focus:ring-[#FFD700]/50 disabled:opacity-50 disabled:cursor-not-allowed whitespace-nowrap cursor-pointer transition-all duration-500 shadow-lg hover:shadow-2xl hover:shadow-[#FFD700]/25 transform hover:scale-[1.02] active:scale-[0.98]"
            >
              {loading ? <LoadingSpinner size="small" /> : 'ACCESS DASHBOARD'}
            </button>
          </div>
        </form>

        <div className="text-center pt-8 border-t border-[#333333]/50">
          <div className="flex items-center justify-center space-x-4 text-[#808080]">
            <div className="w-2 h-2 bg-[#00FFFF] rounded-full animate-pulse"></div>
            <p className="text-xs tracking-wider">
              ENTERPRISE GRADE SECURITY
            </p>
            <div className="w-2 h-2 bg-[#FFD700] rounded-full animate-pulse"></div>
          </div>
        </div>
      </div>
    </div>
  );
}
