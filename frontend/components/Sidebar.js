
'use client';

import Link from 'next/link';
import { usePathname, useRouter } from 'next/navigation';
import { useTranslation } from '../lib/useTranslation';
import { useAuth } from '../lib/AuthContext';
import { useState, useEffect, useRef } from 'react';

export default function Sidebar({ isOpen, onClose }) {
  const pathname = usePathname();
  const { t } = useTranslation();
  const { logout, user } = useAuth();
  const [dropdownOpen, setDropdownOpen] = useState(false);
  const dropdownRef = useRef(null);
  const router = useRouter();

  const handleDropdownClick = (e) => {
    e.preventDefault();
    e.stopPropagation();
    setDropdownOpen(!dropdownOpen);
  };

  const handleDropdownClose = () => {
    setDropdownOpen(false);
  };

  const handleProfileClick = () => {
    handleDropdownClose();
    router.push('/profile');
  };

  const handleLogout = () => {
    handleDropdownClose();
    logout();
  };

  // Handle clicking outside dropdown to close it
  useEffect(() => {
    const handleClickOutside = (event) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target)) {
        setDropdownOpen(false);
      }
    };

    if (dropdownOpen) {
      document.addEventListener('mousedown', handleClickOutside);
    }

    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, [dropdownOpen]);

  const menuItems = [
    { name: t('dashboard'), path: '/', icon: 'ri-dashboard-line' },
    { name: t('transactions'), path: '/transactions', icon: 'ri-exchange-line' },
    { name: 'AI Insights', path: '/ai', icon: 'ri-robot-line' },
    { name: t('analytics'), path: '/analytics', icon: 'ri-bar-chart-line' },
    { name: t('profile'), path: '/profile', icon: 'ri-user-line' },
    { name: t('settings'), path: '/settings', icon: 'ri-settings-line' },
  ];

  return (
    <>
      {/* Mobile overlay */}
      {isOpen && (
        <div 
          className="fixed inset-0 bg-black/50 z-40"
          onClick={onClose}
        />
      )}
      
      {/* Sidebar */}
      <div className={`w-64 bg-gradient-to-b from-[#0d0d0d] to-[#1a1a1a] shadow-2xl h-screen fixed left-0 top-0 border-r border-[#333333]/50 backdrop-blur-xl z-50 transition-transform duration-300 ${
        isOpen ? 'translate-x-0' : '-translate-x-full'
      }`}>
        <div className="p-6 border-b border-[#333333]/50">
          <div className="relative">
            <h1 className="text-3xl font-bold bg-gradient-to-r from-[#FFD700] via-[#FFA500] to-[#00FFFF] bg-clip-text text-transparent tracking-wide" 
                style={{ fontFamily: 'Orbitron, monospace' }}>
              BucksInfo
            </h1>
            <div className="absolute -bottom-2 left-0 w-16 h-0.5 bg-gradient-to-r from-[#FFD700] to-[#00FFFF] rounded-full"></div>
          </div>
        </div>
        
        <nav className="mt-8 px-3">
          {menuItems.map((item) => (
            <Link
              key={item.path}
              href={item.path}
              onClick={onClose}
              className={`flex items-center px-4 py-3 mx-2 my-1 rounded-xl text-white hover:bg-gradient-to-r hover:from-[#FFD700]/20 hover:to-[#00FFFF]/20 transition-all duration-300 cursor-pointer group ${
                pathname === item.path 
                  ? 'bg-gradient-to-r from-[#FFD700]/30 to-[#00FFFF]/30 border border-[#FFD700]/50 shadow-lg shadow-[#FFD700]/25' 
                  : 'hover:border hover:border-[#333333]/50'
              }`}
            >
              <div className={`w-6 h-6 flex items-center justify-center mr-3 ${
                pathname === item.path ? 'text-[#FFD700]' : 'text-[#a0a0a0] group-hover:text-[#00FFFF]'
              }`}>
                <i className={`${item.icon} text-lg`}></i>
              </div>
              <span className={`font-medium tracking-wide ${
                pathname === item.path ? 'text-white' : 'text-[#a0a0a0] group-hover:text-white'
              }`}>
                {item.name}
              </span>
            </Link>
          ))}
        </nav>
        
        {/* Avatar Section */}
        <div className="absolute bottom-4 left-4 right-4">
          <div className="relative" ref={dropdownRef}>
            <button
              onClick={handleDropdownClick}
              onMouseDown={(e) => e.preventDefault()}
              className="w-full bg-[#1a1a1a]/50 rounded-xl p-4 border border-[#333333]/30 hover:bg-[#1a1a1a]/70 transition-all duration-300 cursor-pointer flex items-center space-x-3"
            >
              <div className="w-10 h-10 bg-gradient-to-r from-[#FFD700] to-[#00FFFF] rounded-full flex items-center justify-center">
                <span className="text-[#0d0d0d] text-sm font-bold">
                  {(user?.name || user?.email || 'U').charAt(0).toUpperCase()}
                </span>
              </div>
              <div className="flex-1 text-left">
                <p className="text-white font-semibold text-sm">
                  {user?.name || user?.email?.split('@')[0] || 'User'}
                </p>
                <p className="text-[#808080] text-xs">{user?.email || 'user@example.com'}</p>
              </div>
              <div className="w-2 h-2 bg-[#00FFFF] rounded-full animate-pulse"></div>
            </button>
            
            {dropdownOpen && (
              <>
                <div className="fixed inset-0 bg-black/20 z-[99998]" onClick={handleDropdownClose}></div>
                <div className="absolute bottom-20 left-4 right-4 w-auto bg-[#1a1a1a] rounded-xl shadow-2xl border border-[#333333]/50 backdrop-blur-xl z-[99999]">
                  {/* Menu Items */}
                  <div className="py-2">
                    <button onClick={handleProfileClick} className="w-full">
                      <div className="w-full px-4 py-3 text-[#a0a0a0] hover:text-white hover:bg-gradient-to-r hover:from-[#FFD700]/20 hover:to-[#00FFFF]/20 cursor-pointer transition-all duration-300 flex items-center space-x-3">
                        <div className="w-5 h-5 flex items-center justify-center">
                          <i className="ri-user-line"></i>
                        </div>
                        <span className="text-sm">Profile</span>
                      </div>
                    </button>
                    
                    <div className="px-4 py-3 text-[#a0a0a0] hover:text-white hover:bg-gradient-to-r hover:from-[#FFD700]/20 hover:to-[#00FFFF]/20 cursor-pointer transition-all duration-300 flex items-center space-x-3">
                      <div className="w-5 h-5 flex items-center justify-center">
                        <i className="ri-shield-check-line"></i>
                      </div>
                      <span className="text-sm">Security</span>
                    </div>
                    
                    <div className="px-4 py-3 text-[#a0a0a0] hover:text-white hover:bg-gradient-to-r hover:from-[#FFD700]/20 hover:to-[#00FFFF]/20 cursor-pointer transition-all duration-300 flex items-center space-x-3">
                      <div className="w-5 h-5 flex items-center justify-center">
                        <i className="ri-customer-service-line"></i>
                      </div>
                      <span className="text-sm">Support</span>
                    </div>
                  </div>

                  {/* Divider */}
                  <div className="border-t border-[#333333]/50"></div>

                  {/* Logout */}
                  <div className="py-2">
                    <button
                      onClick={handleLogout}
                      className="w-full text-left px-4 py-3 text-red-400 hover:text-red-300 hover:bg-red-500/20 cursor-pointer transition-all duration-300 flex items-center space-x-3"
                    >
                      <div className="w-5 h-5 flex items-center justify-center">
                        <i className="ri-logout-circle-line"></i>
                      </div>
                      <span className="text-sm">Sign out</span>
                    </button>
                  </div>
                </div>
              </>
            )}
          </div>
        </div>
      </div>
    </>
  );
}
