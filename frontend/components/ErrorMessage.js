
'use client';

export default function ErrorMessage({ message, onRetry }) {
  return (
    <div className="bg-red-900/30 border border-red-500/50 rounded-2xl p-4 backdrop-blur-sm">
      <div className="flex items-center justify-between">
        <div className="flex items-center">
          <div className="w-6 h-6 flex items-center justify-center text-red-400 mr-3">
            <i className="ri-error-warning-line text-lg"></i>
          </div>
          <div>
            <p className="text-red-300 font-medium">{message}</p>
          </div>
        </div>
        {onRetry && (
          <button
            onClick={onRetry}
            className="ml-4 px-4 py-2 bg-red-500/20 text-red-300 rounded-xl hover:bg-red-500/30 transition-colors cursor-pointer whitespace-nowrap"
          >
            Try Again
          </button>
        )}
      </div>
    </div>
  );
}
