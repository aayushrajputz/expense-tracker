'use client';

import { useSettings } from './SettingsContext';
import { useState, useEffect } from 'react';

// Import all translation files
import enTranslations from './translations/en.js';
import esTranslations from './translations/es.js';

const translations = {
  en: enTranslations,
  es: esTranslations,
  // Add more languages as needed
  fr: enTranslations, // Fallback to English for now
  de: enTranslations, // Fallback to English for now
  it: enTranslations, // Fallback to English for now
};

export function useTranslation() {
  const { language } = useSettings();
  const [currentTranslations, setCurrentTranslations] = useState(translations.en);

  useEffect(() => {
    const translation = translations[language] || translations.en;
    setCurrentTranslations(translation);
  }, [language]);

  const t = (key) => {
    return currentTranslations[key] || key;
  };

  return { t, language };
}