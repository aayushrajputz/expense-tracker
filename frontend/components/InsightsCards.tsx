'use client';

import { useState } from 'react';
import { AiInsights } from '../lib/ai/schema';

interface InsightsCardsProps {
  insights: AiInsights;
  onRefresh?: () => void;
  isRefreshing?: boolean;
}

export default function InsightsCards({ insights, onRefresh, isRefreshing = false }: InsightsCardsProps) {
  const [expandedInsight, setExpandedInsight] = useState<string | null>(null);

  const getSeverityColor = (severity: string) => {
    switch (severity) {
      case 'positive':
        return 'border-green-500/30 bg-green-500/10';
      case 'warning':
        return 'border-yellow-500/30 bg-yellow-500/10';
      case 'critical':
        return 'border-red-500/30 bg-red-500/10';
      default:
        return 'border-gray-500/30 bg-gray-500/10';
    }
  };

  const getSeverityIcon = (severity: string) => {
    switch (severity) {
      case 'positive':
        return 'ri-check-line text-green-400';
      case 'warning':
        return 'ri-alert-line text-yellow-400';
      case 'critical':
        return 'ri-error-warning-line text-red-400';
      default:
        return 'ri-information-line text-gray-400';
    }
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleString();
  };

  const formatPercentage = (value: number) => {
    return `${(value * 100).toFixed(1)}%`;
  };

  return (
    <div className="space-y-6">
      {/* Header with refresh button */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold bg-gradient-to-r from-[#FFD700] to-[#00FFFF] bg-clip-text text-transparent">
            AI Financial Insights
          </h2>
          <p className="text-[#a0a0a0] text-sm">
            Last updated: {formatDate(insights.generatedAt)}
          </p>
        </div>
        {onRefresh && (
          <button
            onClick={onRefresh}
            disabled={isRefreshing}
            className="px-4 py-2 bg-gradient-to-r from-[#FFD700] to-[#00FFFF] text-[#0d0d0d] rounded-xl font-bold hover:shadow-lg hover:shadow-[#FFD700]/25 disabled:opacity-50 disabled:cursor-not-allowed transition-all duration-300 cursor-pointer flex items-center space-x-2"
          >
            <i className={`ri-refresh-line text-lg ${isRefreshing ? 'animate-spin' : ''}`}></i>
            <span>{isRefreshing ? 'Refreshing...' : 'Refresh AI'}</span>
          </button>
        )}
      </div>

      {/* Summary Card */}
      <div className="bg-[#1a1a1a]/80 backdrop-blur-xl rounded-2xl border border-[#333333]/50 p-6">
        <h3 className="text-lg font-semibold text-white mb-4">Monthly Summary</h3>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div className="text-center">
            <p className="text-[#a0a0a0] text-sm">Income</p>
            <p className="text-2xl font-bold text-green-400">${insights.summary.income.toFixed(2)}</p>
          </div>
          <div className="text-center">
            <p className="text-[#a0a0a0] text-sm">Expenses</p>
            <p className="text-2xl font-bold text-red-400">${insights.summary.expenses.toFixed(2)}</p>
          </div>
          <div className="text-center">
            <p className="text-[#a0a0a0] text-sm">Savings Rate</p>
            <p className="text-2xl font-bold text-[#FFD700]">{formatPercentage(insights.summary.savingsRate)}</p>
          </div>
        </div>
      </div>

      {/* Top Categories */}
      {insights.topCategories.length > 0 && (
        <div className="bg-[#1a1a1a]/80 backdrop-blur-xl rounded-2xl border border-[#333333]/50 p-6">
          <h3 className="text-lg font-semibold text-white mb-4">Top Spending Categories</h3>
          <div className="space-y-3">
            {insights.topCategories.map((category, index) => (
              <div key={index} className="flex items-center justify-between">
                <span className="text-[#e0e0e0]">{category.name}</span>
                <div className="flex items-center space-x-3">
                  <div className="w-32 bg-[#333333] rounded-full h-2">
                    <div 
                      className="bg-gradient-to-r from-[#FFD700] to-[#00FFFF] h-2 rounded-full transition-all duration-300"
                      style={{ width: `${category.share * 100}%` }}
                    ></div>
                  </div>
                  <span className="text-[#FFD700] font-medium w-12 text-right">
                    {formatPercentage(category.share)}
                  </span>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Insights */}
      <div className="space-y-4">
        <h3 className="text-lg font-semibold text-white">AI Insights</h3>
        {insights.insights.map((insight) => (
          <div
            key={insight.id}
            className={`bg-[#1a1a1a]/80 backdrop-blur-xl rounded-2xl border p-6 transition-all duration-300 cursor-pointer hover:scale-[1.02] ${getSeverityColor(insight.severity)}`}
            onClick={() => setExpandedInsight(expandedInsight === insight.id ? null : insight.id)}
          >
            <div className="flex items-start space-x-4">
              <div className={`w-10 h-10 rounded-full flex items-center justify-center ${getSeverityColor(insight.severity)}`}>
                <i className={`${getSeverityIcon(insight.severity)} text-xl`}></i>
              </div>
              <div className="flex-1">
                <h4 className="text-lg font-semibold text-white mb-2">{insight.title}</h4>
                <p className="text-[#a0a0a0] mb-3">{insight.message}</p>
                {insight.action && (
                  <div className={`transition-all duration-300 ${expandedInsight === insight.id ? 'max-h-20 opacity-100' : 'max-h-0 opacity-0 overflow-hidden'}`}>
                    <p className="text-[#FFD700] font-medium">💡 {insight.action}</p>
                  </div>
                )}
              </div>
              <div className="flex items-center space-x-2">
                <span className={`px-2 py-1 rounded-full text-xs font-medium ${
                  insight.severity === 'positive' ? 'bg-green-500/20 text-green-400' :
                  insight.severity === 'warning' ? 'bg-yellow-500/20 text-yellow-400' :
                  'bg-red-500/20 text-red-400'
                }`}>
                  {insight.severity}
                </span>
                {insight.action && (
                  <i className={`ri-arrow-down-s-line text-[#a0a0a0] transition-transform duration-300 ${
                    expandedInsight === insight.id ? 'rotate-180' : ''
                  }`}></i>
                )}
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}