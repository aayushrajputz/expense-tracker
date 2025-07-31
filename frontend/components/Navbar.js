
'use client';

import Link from 'next/link';

export default function Navbar({ onSidebarToggle }) {
  return (
    <nav className="bg-gradient-to-r from-[#0d0d0d] to-[#1a1a1a] shadow-2xl border-b border-[#333333]/50 px-6 py-4 backdrop-blur-xl">
      <div className="flex items-center justify-between">
        <div className="flex items-center space-x-4">
          <button
            onClick={onSidebarToggle}
            className="p-2 text-[#a0a0a0] hover:text-[#FFD700] transition-all duration-300 cursor-pointer hover:bg-[#333333]/30 rounded-xl"
          >
            <i className="ri-menu-line text-xl"></i>
          </button>
          <h2 className="text-2xl font-bold bg-gradient-to-r from-[#FFD700] to-[#00FFFF] bg-clip-text text-transparent tracking-wide">
            Welcome back!
          </h2>
          <div className="w-2 h-2 bg-[#00FFFF] rounded-full animate-pulse"></div>
        </div>
        
        <div className="flex items-center space-x-6">
          <Link href="/add">
            <button className="px-4 py-2 bg-gradient-to-r from-[#FFD700] to-[#00FFFF] text-[#0d0d0d] rounded-xl font-bold hover:shadow-lg hover:shadow-[#FFD700]/25 transition-all duration-300 cursor-pointer flex items-center space-x-2">
              <i className="ri-add-line text-lg"></i>
              <span>Add Transaction</span>
            </button>
          </Link>
          
          <button className="relative p-3 text-[#a0a0a0] hover:text-[#FFD700] transition-all duration-300 cursor-pointer hover:bg-[#333333]/30 rounded-xl">
            <div className="w-6 h-6 flex items-center justify-center">
              <i className="ri-notification-line text-xl"></i>
            </div>
            <span className="absolute -top-1 -right-1 w-4 h-4 bg-gradient-to-r from-[#FFD700] to-[#FFA500] rounded-full text-[#0d0d0d] text-xs flex items-center justify-center font-bold">
              3
            </span>
          </button>
        </div>
      </div>
    </nav>
  );
}
