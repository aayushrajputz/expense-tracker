'use client';

import React, { createContext, useContext, useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { authAPI } from './api';
import { clearAllCaches } from './api';

const AuthContext = createContext();

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};

export const AuthProvider = ({ children }) => {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);
  const [token, setToken] = useState(null);
  const router = useRouter();

  const clearUserData = () => {
    // Clear all cached data when user changes
    clearAllCaches();
    // Clear all localStorage data except for the current user's data
    const currentToken = localStorage.getItem('token');
    const currentUser = localStorage.getItem('user');
    
    // Clear everything except current user data
    localStorage.clear();
    
    // Restore current user data if it exists
    if (currentToken) localStorage.setItem('token', currentToken);
    if (currentUser) localStorage.setItem('user', currentUser);
  };

  const fetchUserData = async (token) => {
    try {
      // Get real user data from localStorage that was stored during signup/login
      const storedUser = localStorage.getItem('user');
      if (storedUser) {
        const userData = JSON.parse(storedUser);
        setUser(userData);
        return;
      }
      
      // Fallback to stored signup data if available
      const signupName = localStorage.getItem('signupName');
      const signupEmail = localStorage.getItem('signupEmail');
      if (signupName && signupEmail) {
        const userData = {
          id: 1,
          name: signupName,
          email: signupEmail,
          budget: 50000,
          createdAt: new Date().toISOString()
        };
        setUser(userData);
        localStorage.setItem('user', JSON.stringify(userData));
        return;
      }
      
      // If no user data found, create a default user object
      const defaultUser = {
        id: 1,
        name: 'User',
        email: 'user@example.com',
        budget: 50000,
        createdAt: new Date().toISOString()
      };
      setUser(defaultUser);
      localStorage.setItem('user', JSON.stringify(defaultUser));
    } catch (error) {
      console.error('Failed to fetch user data:', error);
      // Set a default user even if there's an error
      const defaultUser = {
        id: 1,
        name: 'User',
        email: 'user@example.com',
        budget: 50000,
        createdAt: new Date().toISOString()
      };
      setUser(defaultUser);
    }
  };

  useEffect(() => {
    const initializeAuth = async () => {
      const storedToken = localStorage.getItem('token');
      if (storedToken) {
        setToken(storedToken);
        await fetchUserData(storedToken);
      } else {
        // Even without token, try to load user data
        await fetchUserData();
      }
      setLoading(false);
    };

    initializeAuth();
  }, []);

  const login = async (email, password) => {
    try {
      // Clear all cached data before login to ensure fresh data
      clearUserData();
      
      const response = await authAPI.login(email, password);
      const { token, user } = response;
      
      localStorage.setItem('token', token);
      localStorage.setItem('user', JSON.stringify(user));
      setToken(token);
      setUser(user);
      
      router.push('/');
      return { success: true };
    } catch (error) {
      return { 
        success: false, 
        error: error.response?.data?.message || 'Login failed' 
      };
    }
  };

  const signup = async (name, email, password, confirmPassword) => {
    try {
      if (password !== confirmPassword) {
        return { success: false, error: 'Passwords do not match' };
      }
      
      // Clear all cached data before signup to ensure fresh data
      clearUserData();
      
      const response = await authAPI.signup(name, email, password);
      
      // Store user data from signup
      const userData = {
        id: 1,
        name: name,
        email: email,
        budget: 50000,
        createdAt: new Date().toISOString()
      };
      
      localStorage.setItem('user', JSON.stringify(userData));
      setUser(userData);
      
      return { success: true, data: response };
    } catch (error) {
      return { 
        success: false, 
        error: error.response?.data?.message || 'Signup failed' 
      };
    }
  };

  const verifyOtp = async (email, otp) => {
    try {
      const response = await authAPI.verifyOtp(email, otp);
      return { success: true, data: response };
    } catch (error) {
      return { 
        success: false, 
        error: error.response?.data?.message || 'OTP verification failed' 
      };
    }
  };

  const logout = () => {
    // Clear all cached data and localStorage on logout
    clearAllCaches();
    localStorage.removeItem('token');
    localStorage.removeItem('user');
    localStorage.removeItem('signupName');
    localStorage.removeItem('signupEmail');
    setToken(null);
    setUser(null);
    router.push('/login');
  };

  const isAuthenticated = () => {
    return !!token;
  };

  const value = {
    user,
    token,
    loading,
    login,
    signup,
    verifyOtp,
    logout,
    isAuthenticated
  };

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  );
};