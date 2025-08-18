
'use client';

export default function InsightBox({ insight }) {
  const getInsightIcon = (type) => {
    const icons = {
      'warning': 'ri-alert-line',
      'tip': 'ri-lightbulb-line',
      'success': 'ri-check-line',
      'info': 'ri-information-line'
    };
    return icons[type] || 'ri-brain-line';
  };

  const getInsightColor = (type) => {
    const colors = {
      'warning': 'border-yellow-500/50 bg-yellow-500/10 text-yellow-400',
      'tip': 'border-blue-500/50 bg-blue-500/10 text-blue-400',
      'success': 'border-green-500/50 bg-green-500/10 text-green-400',
      'info': 'border-purple-500/50 bg-purple-500/10 text-purple-400'
    };
    return colors[type] || 'border-[#FFD700]/50 bg-[#FFD700]/10 text-[#FFD700]';
  };

  return (
    <div className={`p-6 rounded-2xl border backdrop-blur-sm ${getInsightColor(insight.type)}`}>
      <div className="flex items-start space-x-3">
        <div className="flex-shrink-0">
          <div className={`w-10 h-10 rounded-full flex items-center justify-center ${
            insight.type === 'warning' ? 'bg-yellow-500/20' :
            insight.type === 'tip' ? 'bg-blue-500/20' :
            insight.type === 'success' ? 'bg-green-500/20' :
            insight.type === 'info' ? 'bg-purple-500/20' :
            'bg-[#FFD700]/20'
          }`}>
            <i className={`${getInsightIcon(insight.type)} text-lg`}></i>
          </div>
        </div>
        <div className="flex-1">
          <h3 className="font-semibold text-white mb-2">{insight.title}</h3>
          <p className="text-[#a0a0a0] text-sm">{insight.description}</p>
          {insight.suggestion && (
            <p className="text-white text-sm mt-2 font-medium">
              💡 {insight.suggestion}
            </p>
          )}
        </div>
      </div>
    </div>
  );
}
