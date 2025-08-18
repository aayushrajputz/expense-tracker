'use client';

import React, { createContext, useContext, useState, useEffect } from "react";

const SettingsContext = createContext();

export function SettingsProvider({ children }) {
  // Try to load from localStorage, fallback to defaults
  const [language, setLanguage] = useState(() => {
    if (typeof window !== "undefined") {
      return localStorage.getItem("language") || "en";
    }
    return "en";
  });
  const [currency, setCurrency] = useState(() => {
    if (typeof window !== "undefined") {
      return localStorage.getItem("currency") || "USD";
    }
    return "USD";
  });

  const updateLanguage = (newLanguage) => {
    console.log('Updating language to:', newLanguage);
    setLanguage(newLanguage);
    // Save to localStorage immediately when changed
    if (typeof window !== "undefined") {
      localStorage.setItem("language", newLanguage);
      console.log('Language saved to localStorage:', newLanguage);
    }
  };

  const updateCurrency = (newCurrency) => {
    console.log('Updating currency to:', newCurrency);
    setCurrency(newCurrency);
    // Save to localStorage immediately when changed
    if (typeof window !== "undefined") {
      localStorage.setItem("currency", newCurrency);
      console.log('Currency saved to localStorage:', newCurrency);
    }
  };

  // Remove the automatic useEffect hooks that save to localStorage
  // The settings will now be saved only when explicitly called from the settings page

  return (
    <SettingsContext.Provider value={{ 
      language, 
      setLanguage: updateLanguage, 
      currency, 
      setCurrency: updateCurrency 
    }}>
      {children}
    </SettingsContext.Provider>
  );
}

export function useSettings() {
  return useContext(SettingsContext);
}