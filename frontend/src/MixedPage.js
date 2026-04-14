import React, { useMemo } from 'react';
import { BarChart, Bar, XAxis, YAxis, Tooltip, CartesianGrid, ResponsiveContainer, PieChart, Pie, Cell } from "recharts";
import { useNetworkData } from './useNetworkData';

function MixedPage() {
  const { flows, latencySrc, loading, error } = useNetworkData();

  const volumeInData = useMemo(() => {
    const grouped = new Map();
    flows.forEach((flow) => {
      const key = flow?.dstLabel || flow?.dstName || 'Unknown';
      const current = grouped.get(key) || 0;
      grouped.set(key, current + (flow?.bytes || 0));
    });

    return Array.from(grouped.entries())
      .map(([name, value]) => ({ name, value }))
      .sort((a, b) => b.value - a.value)
      .slice(0, 6);
  }, [flows]);

  const volumeOutData = useMemo(() => {
    const grouped = new Map();
    flows.forEach((flow) => {
      const key = flow?.srcLabel || flow?.srcName || 'Unknown';
      const current = grouped.get(key) || 0;
      grouped.set(key, current + (flow?.bytes || 0));
    });

    return Array.from(grouped.entries())
      .map(([name, value]) => ({ name, value }))
      .sort((a, b) => b.value - a.value)
      .slice(0, 6);
  }, [flows]);

  const protocolData = useMemo(() => {
    const grouped = new Map();
    flows.forEach((flow) => {
      const key = flow?.l7ProtoName || flow?.protocolName || 'UNKNOWN';
      const current = grouped.get(key) || 0;
      grouped.set(key, current + (flow?.bytes || 0));
    });

    return Array.from(grouped.entries())
      .map(([name, value]) => ({ name, value }))
      .sort((a, b) => b.value - a.value)
      .slice(0, 6);
  }, [flows]);

  const latencyData = useMemo(() => {
    return latencySrc.map((item) => ({
      name: item?.name || 'Unknown',
      value: Number(item?.avgLatencyMs || 0)
    }));
  }, [latencySrc]);

  const formatMs = (value) => `${Number(value).toFixed(1)} ms`;
  const chartTooltipStyle = {
    contentStyle: {
      backgroundColor: '#0a1b35',
      border: '1px solid #30507f',
      borderRadius: '8px',
    },
    labelStyle: {
      color: '#b8c9e6',
      fontWeight: 600,
    },
    itemStyle: {
      color: '#e6efff',
    },
  };
      
        const COLORS = ["#FF6384", "#36A2EB", "#fe9c00", "#4BC0C0", "#ffe100"];

        const pieInData = volumeInData.length > 0 ? volumeInData : [{ name: 'No data', value: 1 }];
        const pieOutData = volumeOutData.length > 0 ? volumeOutData : [{ name: 'No data', value: 1 }];
        const pieProtocolData = protocolData.length > 0 ? protocolData : [{ name: 'No data', value: 1 }];
        const barLatencyData = latencyData.length > 0 ? latencyData : [{ name: 'No data', value: 0 }];
      
        return (
          <div className="dashboard-page">
            <h1 className="dashboard-heading">Incoming/Outgoing Network Overview</h1>
            {(loading || error) && (
              <p className="status-banner">
                {error ? `Backend connection error: ${error}` : 'Loading backend network data...'}
              </p>
            )}
  
            <div className="dashboard-grid-three">
              <div className="dashboard-card dashboard-card-compact">
                <p className="dashboard-card-title">Volume In</p>
                <div className="dashboard-chart">
                  <ResponsiveContainer>
                    <PieChart>
                      <Pie data={pieInData} dataKey="value" nameKey="name" outerRadius="72%" label cx="50%" cy="52%">
                        {pieInData.map((entry, index) => (
                          <Cell key={index} fill={COLORS[index % COLORS.length]} />
                        ))}
                      </Pie>
                      <Tooltip
                        contentStyle={chartTooltipStyle.contentStyle}
                        labelStyle={chartTooltipStyle.labelStyle}
                        itemStyle={chartTooltipStyle.itemStyle}
                      />
                    </PieChart>
                  </ResponsiveContainer>
                </div>
              </div>

  
              <div className="dashboard-card dashboard-card-compact">
                <p className="dashboard-card-title">Volume Out</p>
                <div className="dashboard-chart">
                  <ResponsiveContainer>
                    <PieChart>
                      <Pie data={pieOutData} dataKey="value" nameKey="name" outerRadius="72%" label cx="50%" cy="52%">
                        {pieOutData.map((entry, index) => (
                          <Cell key={index} fill={COLORS[index % COLORS.length]} />
                        ))}
                      </Pie>
                      <Tooltip
                        contentStyle={chartTooltipStyle.contentStyle}
                        labelStyle={chartTooltipStyle.labelStyle}
                        itemStyle={chartTooltipStyle.itemStyle}
                      />
                    </PieChart>
                  </ResponsiveContainer>
                </div>
              </div>

              <div className="dashboard-card dashboard-card-compact">
                <p className="dashboard-card-title">Networking Protocol</p>
                <div className="dashboard-chart">
                  <ResponsiveContainer>
                    <PieChart>
                      <Pie data={pieProtocolData} dataKey="value" nameKey="name" outerRadius="72%" label cx="50%" cy="52%">
                        {pieProtocolData.map((entry, index) => (
                          <Cell key={index} fill={COLORS[index % COLORS.length]} />
                        ))}
                      </Pie>
                      <Tooltip
                        contentStyle={chartTooltipStyle.contentStyle}
                        labelStyle={chartTooltipStyle.labelStyle}
                        itemStyle={chartTooltipStyle.itemStyle}
                      />
                    </PieChart>
                  </ResponsiveContainer>
                </div>
              </div>
            </div>

            <div className="dashboard-card dashboard-card-full dashboard-section-gap">
              <p className="dashboard-card-title">Latency</p>
              <div className="dashboard-chart">
                <ResponsiveContainer>
                  <BarChart data={barLatencyData} layout="vertical">
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis type="number" tickFormatter={formatMs} tick={{ fill: '#b8c9e6' }} />
                    <YAxis type="category" dataKey="name" width={90} tick={{ fill: '#b8c9e6' }} />
                    <Tooltip
                      formatter={(value) => formatMs(value)}
                      contentStyle={chartTooltipStyle.contentStyle}
                      labelStyle={chartTooltipStyle.labelStyle}
                      itemStyle={chartTooltipStyle.itemStyle}
                      cursor={{ fill: 'rgba(59, 130, 246, 0.18)' }}
                    />
                    <Bar dataKey="value" fill="#36A2EB" />
                  </BarChart>
                </ResponsiveContainer>
              </div>
            </div>
          </div>
        );
      }
export default MixedPage;
