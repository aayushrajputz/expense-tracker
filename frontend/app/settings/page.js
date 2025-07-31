
'use client';

import { useState, useEffect } from 'react';
import Sidebar from '../../components/Sidebar';
import Navbar from '../../components/Navbar';
import LoadingSpinner from '../../components/LoadingSpinner';
import ErrorMessage from '../../components/ErrorMessage';
import ProtectedRoute from '../../components/ProtectedRoute';
import { useSettings } from '../../lib/SettingsContext';
import { useTranslation } from '../../lib/useTranslation';

export default function Settings() {
  const { language: globalLanguage, setLanguage: setGlobalLanguage, currency: globalCurrency, setCurrency: setGlobalCurrency } = useSettings();
  const { t } = useTranslation();
  
  // Local state for language and currency (only applied when save is clicked)
  const [localLanguage, setLocalLanguage] = useState(globalLanguage);
  const [localCurrency, setLocalCurrency] = useState(globalCurrency);
  
  // Track if there are unsaved changes
  const [hasUnsavedChanges, setHasUnsavedChanges] = useState(false);
  
  console.log('Settings page - Global language:', globalLanguage);
  console.log('Settings page - Global currency:', globalCurrency);
  console.log('Settings page - Local language:', localLanguage);
  console.log('Settings page - Local currency:', localCurrency);
  console.log('Settings page - Has unsaved changes:', hasUnsavedChanges);
  
  const [settings, setSettings] = useState({
    notifications: true,
    darkMode: false,
    autoSync: true,
    budgetAlerts: true,
    weeklyReports: false,
    monthlyReports: true,
    dataExport: 'json',
    backupFrequency: 'weekly'
  });

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [success, setSuccess] = useState(false);
  const [uploadLoading, setUploadLoading] = useState(false);
  const [deleteLoading, setDeleteLoading] = useState(false);
  const [sidebarOpen, setSidebarOpen] = useState(false);

  useEffect(() => {
    loadSettings();
  }, []);

  // Update local state when global settings change
  useEffect(() => {
    setLocalLanguage(globalLanguage);
    setLocalCurrency(globalCurrency);
  }, [globalLanguage, globalCurrency]);

  // Check for unsaved changes
  useEffect(() => {
    const hasChanges = localLanguage !== globalLanguage || localCurrency !== globalCurrency;
    setHasUnsavedChanges(hasChanges);
  }, [localLanguage, localCurrency, globalLanguage, globalCurrency]);

  const loadSettings = () => {
    const savedSettings = localStorage.getItem('expenseTrackerSettings');
    if (savedSettings) {
      setSettings(JSON.parse(savedSettings));
    }
  };

  const handleChange = (key, value) => {
    setSettings(prev => ({ ...prev, [key]: value }));
    setSuccess(false);
    setError(null);
  };

  const handleLanguageChange = (newLanguage) => {
    console.log('Local language changed to:', newLanguage);
    setLocalLanguage(newLanguage);
    setSuccess(false);
    setError(null);
  };

  const handleCurrencyChange = (newCurrency) => {
    console.log('Local currency changed to:', newCurrency);
    setLocalCurrency(newCurrency);
    setSuccess(false);
    setError(null);
  };

  const handleSave = async () => {
    try {
      setLoading(true);
      setError(null);

      // Apply language and currency changes globally
      setGlobalLanguage(localLanguage);
      setGlobalCurrency(localCurrency);

      // Save other settings to localStorage
      localStorage.setItem('expenseTrackerSettings', JSON.stringify(settings));
      
      await new Promise(resolve => setTimeout(resolve, 1000));

      setSuccess(true);
      setHasUnsavedChanges(false);
      setTimeout(() => setSuccess(false), 3000);
    } catch (err) {
      setError(t('failedToSaveSettings'));
    } finally {
      setLoading(false);
    }
  };

  const handleUploadData = async () => {
    try {
      setUploadLoading(true);
      setError(null);
      
      // Create a file input element
      const fileInput = document.createElement('input');
      fileInput.type = 'file';
      fileInput.accept = '.json,.csv,.xlsx';
      fileInput.style.display = 'none';
      
      // Handle file selection
      fileInput.onchange = async (event) => {
        const file = event.target.files[0];
        if (!file) {
          setUploadLoading(false);
          return;
        }
        
        // Validate file size (max 10MB)
        if (file.size > 10 * 1024 * 1024) {
          setError('File size must be less than 10MB');
          setUploadLoading(false);
          return;
        }
        
        // Validate file type
        const allowedTypes = ['application/json', 'text/csv', 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet'];
        if (!allowedTypes.includes(file.type)) {
          setError('Please select a valid file type (.json, .csv, .xlsx)');
          setUploadLoading(false);
          return;
        }
        
        try {
          // Simulate file upload process
          await new Promise(resolve => setTimeout(resolve, 2000));
          
          // Show success message
          setSuccess(true);
          setTimeout(() => setSuccess(false), 3000);
          
          // You can add actual file upload logic here
          console.log('File uploaded:', file.name);
          alert(`File "${file.name}" uploaded successfully!`);
          
        } catch (err) {
          setError(t('failedToUploadData'));
        } finally {
          setUploadLoading(false);
        }
      };
      
      // Handle case when user cancels file selection
      fileInput.oncancel = () => {
        setUploadLoading(false);
      };
      
      // Trigger file selection
      document.body.appendChild(fileInput);
      fileInput.click();
      document.body.removeChild(fileInput);
      
    } catch (err) {
      setError(t('failedToUploadData'));
      setUploadLoading(false);
    }
  };

  const handleDeleteAccount = async () => {
    try {
      setDeleteLoading(true);
      setError(null);
      
      // Call the backend endpoint to delete user and all data
      const response = await fetch('/api/user', {
        method: 'DELETE',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('token')}`
        }
      });

      if (response.ok) {
        // Clear local storage and redirect to login
        localStorage.clear();
        window.location.href = '/login';
      } else {
        const errorData = await response.json();
        setError(errorData.error || 'Failed to delete account');
      }
    } catch (err) {
      setError(t('failedToDeleteAccount'));
    } finally {
      setDeleteLoading(false);
    }
  };

  const handleReset = () => {
    const defaultSettings = {
      notifications: true,
      darkMode: false,
      autoSync: true,
      budgetAlerts: true,
      weeklyReports: false,
      monthlyReports: true,
      dataExport: 'json',
      backupFrequency: 'weekly'
    };
    setSettings(defaultSettings);
    // Reset local language and currency to match global values
    setLocalLanguage(globalLanguage);
    setLocalCurrency(globalCurrency);
    setSuccess(false);
    setError(null);
  };

  return (
    <ProtectedRoute>
      <div className="flex h-screen bg-gradient-to-br from-[#0d0d0d] via-[#1a1a1a] to-[#0d0d0d]">
        <Sidebar isOpen={sidebarOpen} onClose={() => setSidebarOpen(false)} />
        <div className="flex-1 flex flex-col">
          <Navbar onSidebarToggle={() => setSidebarOpen(!sidebarOpen)} />
          <main className="flex-1 p-4 md:p-6 overflow-auto">
            <div className="max-w-4xl mx-auto">
              <div className="mb-6 md:mb-8">
                <h1 className="text-2xl md:text-3xl font-bold bg-gradient-to-r from-[#FFD700] to-[#00FFFF] bg-clip-text text-transparent mb-2">
                  {t('settings')}
                </h1>
                <p className="text-[#a0a0a0]">{t('managePreferences')}</p>
              </div>

              <div className="space-y-6">
                {error && <ErrorMessage message={error} />}

                {success && (
                  <div className="bg-green-900/30 border border-green-500/50 rounded-2xl p-4 backdrop-blur-sm">
                    <div className="flex items-center">
                      <div className="w-6 h-6 flex items-center justify-center text-green-400 mr-3">
                        <i className="ri-check-line text-lg"></i>
                      </div>
                      <p className="text-green-300 font-medium">{t('settingsSaved')}</p>
                    </div>
                  </div>
                )}

                <div className="bg-[#1a1a1a]/80 backdrop-blur-xl rounded-2xl border border-[#333333]/50 p-6">
                  <h2 className="text-lg font-semibold text-white mb-4 flex items-center">
                    <div className="w-6 h-6 bg-gradient-to-r from-[#FFD700] to-[#00FFFF] rounded-full flex items-center justify-center mr-3">
                      <i className="ri-settings-line text-[#0d0d0d] text-sm"></i>
                    </div>
                    {t('generalSettings')}
                  </h2>
                  <div className="space-y-4">
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                      <div>
                        <label className="block text-sm font-medium text-[#e0e0e0] mb-2">
                          {t('currency')}
                          {localCurrency !== globalCurrency && (
                            <span className="ml-2 text-xs text-yellow-400">{t('unsaved')}</span>
                          )}
                        </label>
                        <div className="relative">
                          <select
                            value={localCurrency}
                            onChange={(e) => handleCurrencyChange(e.target.value)}
                            className={`w-full px-3 py-2 pr-8 bg-[#0d0d0d]/60 border rounded-xl text-white focus:outline-none focus:ring-2 focus:ring-[#FFD700] focus:border-[#FFD700] appearance-none ${
                              localCurrency !== globalCurrency 
                                ? 'border-yellow-500/50 focus:ring-yellow-500/50' 
                                : 'border-[#333333]'
                            }`}
                          >
                            <option value="USD">USD ({t('USD')})</option>
                            <option value="EUR">EUR ({t('EUR')})</option>
                            <option value="GBP">GBP ({t('GBP')})</option>
                            <option value="JPY">JPY ({t('JPY')})</option>
                            <option value="CAD">CAD ({t('CAD')})</option>
                          </select>
                          <div className="absolute right-3 top-1/2 transform -translate-y-1/2 w-4 h-4 flex items-center justify-center pointer-events-none">
                            <i className="ri-arrow-down-s-line text-[#a0a0a0]"></i>
                          </div>
                        </div>
                      </div>

                      <div>
                        <label className="block text-sm font-medium text-[#e0e0e0] mb-2">
                          {t('language')}
                          {localLanguage !== globalLanguage && (
                            <span className="ml-2 text-xs text-yellow-400">{t('unsaved')}</span>
                          )}
                        </label>
                        <div className="relative">
                          <select
                            value={localLanguage}
                            onChange={(e) => handleLanguageChange(e.target.value)}
                            className={`w-full px-3 py-2 pr-8 bg-[#0d0d0d]/60 border rounded-xl text-white focus:outline-none focus:ring-2 focus:ring-[#FFD700] focus:border-[#FFD700] appearance-none ${
                              localLanguage !== globalLanguage 
                                ? 'border-yellow-500/50 focus:ring-yellow-500/50' 
                                : 'border-[#333333]'
                            }`}
                          >
                            <option value="en">English</option>
                            <option value="es">Español</option>
                            <option value="fr">Français</option>
                            <option value="de">Deutsch</option>
                            <option value="it">Italiano</option>
                          </select>
                          <div className="absolute right-3 top-1/2 transform -translate-y-1/2 w-4 h-4 flex items-center justify-center pointer-events-none">
                            <i className="ri-arrow-down-s-line text-[#a0a0a0]"></i>
                          </div>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>

                <div className="bg-[#1a1a1a]/80 backdrop-blur-xl rounded-2xl border border-[#333333]/50 p-6">
                  <h2 className="text-lg font-semibold text-white mb-4 flex items-center">
                    <div className="w-6 h-6 bg-gradient-to-r from-[#00FFFF] to-[#40E0D0] rounded-full flex items-center justify-center mr-3">
                      <i className="ri-notification-line text-[#0d0d0d] text-sm"></i>
                    </div>
                    {t('notifications')}
                  </h2>
                  <div className="space-y-4">
                    {[
                      { key: 'notifications', label: t('pushNotifications'), desc: t('receiveImportantUpdates') },
                      { key: 'budgetAlerts', label: t('budgetAlerts'), desc: t('getNotifiedWhenApproachingBudgetLimits') },
                      { key: 'weeklyReports', label: t('weeklyReports'), desc: t('receiveWeeklySpendingSummaries') },
                      { key: 'monthlyReports', label: t('monthlyReports'), desc: t('receiveMonthlyFinancialInsights') }
                    ].map(({ key, label, desc }) => (
                      <div key={key} className="flex items-center justify-between p-4 bg-[#0d0d0d]/30 rounded-xl">
                        <div>
                          <h3 className="font-medium text-white">{label}</h3>
                          <p className="text-sm text-[#a0a0a0]">{desc}</p>
                        </div>
                        <label className="relative inline-flex items-center cursor-pointer">
                          <input
                            type="checkbox"
                            checked={settings[key]}
                            onChange={(e) => handleChange(key, e.target.checked)}
                            className="sr-only"
                          />
                          <div className={`w-11 h-6 rounded-full transition-colors ${
                            settings[key] ? 'bg-gradient-to-r from-[#FFD700] to-[#00FFFF]' : 'bg-[#333333]'
                          }`}>
                            <div className={`w-4 h-4 bg-white rounded-full shadow transform transition-transform ${
                              settings[key] ? 'translate-x-6' : 'translate-x-1'
                            } mt-1`}></div>
                          </div>
                        </label>
                      </div>
                    ))}
                  </div>
                </div>

                <div className="bg-[#1a1a1a]/80 backdrop-blur-xl rounded-2xl border border-[#333333]/50 p-6">
                  <h2 className="text-lg font-semibold text-white mb-4 flex items-center">
                    <div className="w-6 h-6 bg-gradient-to-r from-purple-500 to-pink-500 rounded-full flex items-center justify-center mr-3">
                      <i className="ri-database-line text-white text-sm"></i>
                    </div>
                    {t('dataManagement')}
                  </h2>
                  <div className="space-y-6">
                    <div className="flex items-center justify-between p-4 bg-[#0d0d0d]/30 rounded-xl">
                      <div>
                        <h3 className="font-medium text-white">{t('autoSync')}</h3>
                        <p className="text-sm text-[#a0a0a0]">{t('automaticallySyncData')}</p>
                      </div>
                      <label className="relative inline-flex items-center cursor-pointer">
                        <input
                          type="checkbox"
                          checked={settings.autoSync}
                          onChange={(e) => handleChange('autoSync', e.target.checked)}
                          className="sr-only"
                        />
                        <div className={`w-11 h-6 rounded-full transition-colors ${
                          settings.autoSync ? 'bg-gradient-to-r from-[#FFD700] to-[#00FFFF]' : 'bg-[#333333]'
                        }`}>
                          <div className={`w-4 h-4 bg-white rounded-full shadow transform transition-transform ${
                            settings.autoSync ? 'translate-x-6' : 'translate-x-1'
                          } mt-1`}></div>
                        </div>
                      </label>
                    </div>

                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                      <div>
                        <label className="block text-sm font-medium text-[#e0e0e0] mb-2">
                          {t('dataExportFormat')}
                        </label>
                        <div className="relative">
                          <select
                            value={settings.dataExport}
                            onChange={(e) => handleChange('dataExport', e.target.value)}
                            className="w-full px-3 py-2 pr-8 bg-[#0d0d0d]/60 border border-[#333333] rounded-xl text-white focus:outline-none focus:ring-2 focus:ring-[#FFD700] focus:border-[#FFD700] appearance-none"
                          >
                            <option value="json">{t('json')}</option>
                            <option value="csv">{t('csv')}</option>
                            <option value="excel">{t('excel')}</option>
                            <option value="pdf">{t('pdf')}</option>
                          </select>
                          <div className="absolute right-3 top-1/2 transform -translate-y-1/2 w-4 h-4 flex items-center justify-center pointer-events-none">
                            <i className="ri-arrow-down-s-line text-[#a0a0a0]"></i>
                          </div>
                        </div>
                      </div>

                      <div>
                        <label className="block text-sm font-medium text-[#e0e0e0] mb-2">
                          {t('backupFrequency')}
                        </label>
                        <div className="relative">
                          <select
                            value={settings.backupFrequency}
                            onChange={(e) => handleChange('backupFrequency', e.target.value)}
                            className="w-full px-3 py-2 pr-8 bg-[#0d0d0d]/60 border border-[#333333] rounded-xl text-white focus:outline-none focus:ring-2 focus:ring-[#FFD700] focus:border-[#FFD700] appearance-none"
                          >
                            <option value="daily">{t('daily')}</option>
                            <option value="weekly">{t('weekly')}</option>
                            <option value="monthly">{t('monthly')}</option>
                            <option value="never">{t('never')}</option>
                          </select>
                          <div className="absolute right-3 top-1/2 transform -translate-y-1/2 w-4 h-4 flex items-center justify-center pointer-events-none">
                            <i className="ri-arrow-down-s-line text-[#a0a0a0]"></i>
                          </div>
                        </div>
                      </div>
                    </div>

                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                      <button
                        onClick={handleUploadData}
                        disabled={uploadLoading}
                        className="flex items-center justify-center space-x-2 px-6 py-3 bg-gradient-to-r from-[#00FFFF] to-[#40E0D0] text-[#0d0d0d] rounded-xl font-bold hover:shadow-lg hover:shadow-[#00FFFF]/25 disabled:opacity-50 disabled:cursor-not-allowed transition-all duration-300 cursor-pointer whitespace-nowrap"
                      >
                        {uploadLoading ? (
                          <>
                            <LoadingSpinner size="sm" />
                            <span>{t('uploading')}</span>
                          </>
                        ) : (
                          <>
                            <i className="ri-upload-cloud-line text-lg"></i>
                            <span>{t('uploadData')}</span>
                          </>
                        )}
                      </button>

                      <button
                        onClick={handleDeleteAccount}
                        disabled={deleteLoading}
                        className="flex items-center justify-center space-x-2 px-6 py-3 bg-gradient-to-r from-red-500 to-red-600 text-white rounded-xl font-bold hover:shadow-lg hover:shadow-red-500/25 disabled:opacity-50 disabled:cursor-not-allowed transition-all duration-300 cursor-pointer whitespace-nowrap"
                      >
                        {deleteLoading ? (
                          <>
                            <LoadingSpinner size="sm" />
                            <span>{t('deleting')}</span>
                          </>
                        ) : (
                          <>
                            <i className="ri-delete-bin-line text-lg"></i>
                            <span>{t('deleteAccount')}</span>
                          </>
                        )}
                      </button>
                    </div>
                  </div>
                </div>

                {/* Delete Account Section */}
                <div className="bg-[#1a1a1a]/80 backdrop-blur-xl rounded-2xl border border-red-500/30 p-6 mt-10">
                  <div className="flex items-center justify-between">
                    <div className="flex-1">
                      <h3 className="text-lg font-semibold text-white mb-2">Delete your account</h3>
                      <p className="text-[#a0a0a0]">Once you delete your account, all your data will be permanently removed. There is no going back. Please be certain.</p>
                    </div>
                    <button
                      onClick={() => {
                        if (window.confirm('Are you sure you want to delete your account? This action cannot be undone and all your data will be permanently removed.')) {
                          handleDeleteAccount();
                        }
                      }}
                      disabled={deleteLoading}
                      className="px-6 py-3 bg-red-600 text-white rounded-xl font-bold hover:bg-red-700 transition-colors cursor-pointer whitespace-nowrap disabled:opacity-50 disabled:cursor-not-allowed ml-6"
                    >
                      {deleteLoading ? 'Deleting...' : 'Delete Account'}
                    </button>
                  </div>
                  <div className="w-full h-0.5 bg-red-500/50 mt-4 rounded-full"></div>
                </div>

                <div className="flex flex-col md:flex-row justify-end space-y-4 md:space-y-0 md:space-x-4">
                  <button
                    onClick={handleReset}
                    className="px-6 py-3 text-[#a0a0a0] bg-[#333333]/50 rounded-xl hover:bg-[#333333]/70 transition-colors cursor-pointer whitespace-nowrap"
                  >
                    {t('resetToDefault')}
                  </button>
                  <button
                    onClick={handleSave}
                    disabled={loading || !hasUnsavedChanges}
                    className={`px-6 py-3 rounded-xl font-bold transition-all duration-300 cursor-pointer whitespace-nowrap flex items-center justify-center space-x-2 ${
                      hasUnsavedChanges
                        ? 'bg-gradient-to-r from-[#FFD700] to-[#FFA500] text-[#0d0d0d] hover:shadow-lg hover:shadow-[#FFD700]/25'
                        : 'bg-[#333333]/50 text-[#666666] cursor-not-allowed'
                    } disabled:opacity-50 disabled:cursor-not-allowed`}
                  >
                    {loading ? (
                      <>
                        <LoadingSpinner size="sm" />
                        <span>{t('saving')}</span>
                      </>
                    ) : (
                      <>
                        <i className="ri-save-line text-lg"></i>
                        <span>{hasUnsavedChanges ? t('saveSettings') : t('noChangesToSave')}</span>
                      </>
                    )}
                  </button>
                </div>
              </div>
            </div>
          </main>
        </div>
      </div>
    </ProtectedRoute>
  );
}
