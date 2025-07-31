
'use client';

import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, BarChart, Bar, PieChart, Pie, Cell, AreaChart, Area, Legend } from 'recharts';
import { useSettings } from '../lib/SettingsContext';

export default function Chart({ data, type, title }) {
  const { currency } = useSettings();
  const colors = {
    primary: '#FFD700',
    secondary: '#00FFFF',
    accent: '#FFA500',
    success: '#10B981',
    warning: '#F59E0B',
    error: '#EF4444',
    purple: '#8B5CF6',
    pink: '#EC4899'
  };

  const pieColors = ['#FFD700', '#00FFFF', '#FFA500', '#10B981', '#F59E0B', '#EF4444', '#8B5CF6', '#EC4899'];

  const CustomTooltip = ({ active, payload, label }) => {
    if (active && payload && payload.length) {
      return (
        <div className="bg-[#1a1a1a] border border-[#333333] rounded-xl p-4 shadow-2xl backdrop-blur-xl">
          <p className="text-white font-semibold mb-2">{label}</p>
          {payload.map((entry, index) => (
            <p key={index} className="text-sm" style={{ color: entry.color }}>
              {entry.name}: {entry.value?.toLocaleString(undefined, { style: 'currency', currency }) || entry.value}
            </p>
          ))}
        </div>
      );
    }
    return null;
  };

  const renderChart = () => {
    switch (type) {
      case 'line':
        return (
          <LineChart data={data}>
            <CartesianGrid strokeDasharray="3 3" stroke="#333333" opacity={0.3} />
            <XAxis 
              dataKey="name" 
              stroke="#a0a0a0"
              fontSize={12}
              tickLine={false}
              axisLine={false}
            />
            <YAxis 
              stroke="#a0a0a0"
              fontSize={12}
              tickLine={false}
              axisLine={false}
              tickFormatter={(value) => value.toLocaleString(undefined, { style: 'currency', currency })}
            />
            <Tooltip content={<CustomTooltip />} />
            <Legend />
            <Line 
              type="monotone" 
              dataKey="income" 
              stroke={colors.success}
              strokeWidth={3}
              dot={{ fill: colors.success, strokeWidth: 2, r: 5 }}
              activeDot={{ r: 8, stroke: colors.success, strokeWidth: 2 }}
            />
            <Line 
              type="monotone" 
              dataKey="expenses" 
              stroke={colors.error}
              strokeWidth={3}
              dot={{ fill: colors.error, strokeWidth: 2, r: 5 }}
              activeDot={{ r: 8, stroke: colors.error, strokeWidth: 2 }}
            />
          </LineChart>
        );
      
      case 'area':
        return (
          <AreaChart data={data}>
            <CartesianGrid strokeDasharray="3 3" stroke="#333333" opacity={0.3} />
            <XAxis 
              dataKey="name" 
              stroke="#a0a0a0"
              fontSize={12}
              tickLine={false}
              axisLine={false}
            />
            <YAxis 
              stroke="#a0a0a0"
              fontSize={12}
              tickLine={false}
              axisLine={false}
              tickFormatter={(value) => value.toLocaleString(undefined, { style: 'currency', currency })}
            />
            <Tooltip content={<CustomTooltip />} />
            <Legend />
            <Area 
              type="monotone" 
              dataKey="income" 
              stackId="1"
              stroke={colors.success}
              fill={colors.success}
              fillOpacity={0.6}
              strokeWidth={2}
            />
            <Area 
              type="monotone" 
              dataKey="expenses" 
              stackId="1"
              stroke={colors.error}
              fill={colors.error}
              fillOpacity={0.6}
              strokeWidth={2}
            />
          </AreaChart>
        );
      
      case 'bar':
        return (
          <BarChart data={data}>
            <CartesianGrid strokeDasharray="3 3" stroke="#333333" opacity={0.3} />
            <XAxis 
              dataKey="name" 
              stroke="#a0a0a0"
              fontSize={12}
              tickLine={false}
              axisLine={false}
            />
            <YAxis 
              stroke="#a0a0a0"
              fontSize={12}
              tickLine={false}
              axisLine={false}
              tickFormatter={(value) => value.toLocaleString(undefined, { style: 'currency', currency })}
            />
            <Tooltip content={<CustomTooltip />} />
            <Legend />
            <Bar 
              dataKey="income" 
              fill={colors.success} 
              radius={[6, 6, 0, 0]}
              name="Income"
            />
            <Bar 
              dataKey="expenses" 
              fill={colors.error} 
              radius={[6, 6, 0, 0]}
              name="Expenses"
            />
            <Bar 
              dataKey="savings" 
              fill={colors.primary} 
              radius={[6, 6, 0, 0]}
              name="Savings"
            />
          </BarChart>
        );
      
      case 'pie':
        return (
          <PieChart>
            <Pie
              data={data}
              cx="50%"
              cy="50%"
              labelLine={false}
              label={({ name, percent }) => `${name} ${(percent * 100).toFixed(0)}%`}
              outerRadius={80}
              fill="#8884d8"
              dataKey="value"
              nameKey="name"
            >
              {data.map((entry, index) => (
                <Cell 
                  key={`cell-${index}`} 
                  fill={pieColors[index % pieColors.length]}
                  stroke="#1a1a1a"
                  strokeWidth={2}
                />
              ))}
            </Pie>
            <Tooltip content={<CustomTooltip />} />
            <Legend />
          </PieChart>
        );
      
      case 'doughnut':
        return (
          <PieChart>
            <Pie
              data={data}
              cx="50%"
              cy="50%"
              labelLine={false}
              label={({ name, percent }) => `${name} ${(percent * 100).toFixed(0)}%`}
              innerRadius={40}
              outerRadius={80}
              fill="#8884d8"
              dataKey="value"
              nameKey="name"
            >
              {data.map((entry, index) => (
                <Cell 
                  key={`cell-${index}`} 
                  fill={pieColors[index % pieColors.length]}
                  stroke="#1a1a1a"
                  strokeWidth={2}
                />
              ))}
            </Pie>
            <Tooltip content={<CustomTooltip />} />
            <Legend />
          </PieChart>
        );
      
      default:
        return null;
    }
  };

  return (
    <div className="w-full h-64 md:h-80">
      <ResponsiveContainer width="100%" height="100%">
        {renderChart()}
      </ResponsiveContainer>
    </div>
  );
}
