
'use client';

import { useSettings } from '../lib/SettingsContext';

export default function ExpenseCard({ expense, onDelete }) {
  const { language, currency } = useSettings();
  
  console.log('ExpenseCard - Current currency:', currency);
  
  const getCategoryIcon = (category) => {
    const icons = {
      'Food': 'ri-restaurant-line',
      'Transportation': 'ri-car-line',
      'Shopping': 'ri-shopping-bag-line',
      'Entertainment': 'ri-gamepad-line',
      'Healthcare': 'ri-heart-pulse-line',
      'Education': 'ri-book-line',
      'Bills': 'ri-bill-line',
      'Travel': 'ri-plane-line'
    };
    return icons[category] || 'ri-wallet-line';
  };

  const getTypeColor = (type) => {
    return type === 'income' ? 'text-green-400' : 'text-red-400';
  };

  const handleDelete = (e) => {
    e.preventDefault();
    e.stopPropagation();
    if (onDelete && expense.id) {
      // Simple confirmation
      if (confirm('Are you sure you want to delete this transaction?')) {
        onDelete(expense.id);
      }
    }
  };

  return (
    <>
      <div className="flex items-center justify-between p-4 bg-[#0d0d0d]/30 rounded-xl hover:bg-[#0d0d0d]/50 transition-all duration-300">
        <div className="flex items-center space-x-3">
          <div className={`w-10 h-10 rounded-full flex items-center justify-center ${
            expense.type === 'income' ? 'bg-green-500/20' : 'bg-red-500/20'
          }`}>
            <i className={`${getCategoryIcon(expense.category)} ${getTypeColor(expense.type)} text-lg`}></i>
          </div>
          <div>
            <h4 className="text-white font-medium">{expense.title}</h4>
            <p className="text-[#a0a0a0] text-sm">{expense.category}</p>
          </div>
        </div>
        <div className="flex items-center space-x-3">
          <div className="text-right">
            <p className={`font-bold ${getTypeColor(expense.type)}`}>
              {expense.type === 'income' ? '+' : '-'}
              {expense.amount.toLocaleString(undefined, { style: 'currency', currency })}
            </p>
            <p className="text-[#808080] text-sm">{expense.date}</p>
          </div>
          {onDelete && expense.id && (
            <button
              onClick={handleDelete}
              className="p-2 text-red-400 hover:bg-red-400/20 rounded-lg transition-colors cursor-pointer"
              title="Delete transaction"
            >
              <i className="ri-delete-bin-line text-sm"></i>
            </button>
          )}
        </div>
      </div>
    </>
  );
}
