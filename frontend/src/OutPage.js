import React, { useMemo } from 'react';
import { BarChart, Bar, XAxis, YAxis, Tooltip, CartesianGrid, ResponsiveContainer, PieChart, Pie, Cell } from "recharts";
import { useNetworkData } from './useNetworkData';

function OutPage() {
  const { flows, latencySrc, loading, error } = useNetworkData();

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
      
        const COLORS = ["#FF6384", "#36A2EB", "#fe9c00", "#4BC0C0", "#ffe100"];

        const pieVolumeData = volumeOutData.length > 0 ? volumeOutData : [{ name: 'No data', value: 1 }];
        const pieProtocolData = protocolData.length > 0 ? protocolData : [{ name: 'No data', value: 1 }];
        const barLatencyData = latencyData.length > 0 ? latencyData : [{ name: 'No data', value: 0 }];
      
        return (
          <div className="w-full max-w-6xl mx-auto mt-8 px-4">
            <h1 className="text-2xl font-bold text-slate-800 mb-6 text-center ">Outgoing Network Overview</h1>
            {(loading || error) && (
              <p className="text-center text-slate-700 mb-4">
                {error ? `Backend connection error: ${error}` : 'Loading backend network data...'}
              </p>
            )}
  
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div className="h-96 rounded-xl border border-slate-200 bg-gray-300 p-4 shadow-sm flex flex-col hover:shadow-2xl">
                <p className="text-center mb-2 font-semibold text-lg text-slate-800">Volume Out</p>
                <div className="flex-1">
                  <ResponsiveContainer>
                    <PieChart>
                      <Pie data={pieVolumeData} dataKey="value" nameKey="name" outerRadius="72%" label cx="50%" cy="52%">
                        {pieVolumeData.map((entry, index) => (
                          <Cell key={index} fill={COLORS[index % COLORS.length]} />
                        ))}
                      </Pie>
                      <Tooltip />
                    </PieChart>
                  </ResponsiveContainer>
                </div>
              </div>
  
              <div className="h-96 rounded-xl border border-slate-200 bg-gray-300 p-4 shadow-sm flex flex-col hover:shadow-2xl">
                <p className="text-center mb-2 font-semibold text-lg text-slate-800">Networking Protocol</p>
                <div className="flex-1">
                  <ResponsiveContainer>
                    <PieChart>
                      <Pie data={pieProtocolData} dataKey="value" nameKey="name" outerRadius="72%" label cx="50%" cy="52%">
                        {pieProtocolData.map((entry, index) => (
                          <Cell key={index} fill={COLORS[index % COLORS.length]} />
                        ))}
                      </Pie>
                      <Tooltip />
                    </PieChart>
                  </ResponsiveContainer>
                </div>
              </div>
  
              <div className="h-96 md:col-span-2 rounded-xl border border-slate-200 bg-gray-300 p-4 shadow-sm flex flex-col hover:shadow-2xl">
                <p className="text-center mb-2 font-semibold text-lg text-slate-800">Latency</p>
                <div className="flex-1">
                  <ResponsiveContainer>
                    <BarChart data={barLatencyData} layout="vertical">
                      <CartesianGrid strokeDasharray="3 3" />
                      <XAxis type="number" tickFormatter={formatMs} />
                      <YAxis type="category" dataKey="name" width={90} />
                      <Tooltip formatter={(value) => formatMs(value)} />
                      <Bar dataKey="value" fill="#36A2EB" />
                    </BarChart>
                  </ResponsiveContainer>
                </div>
              </div>
            </div>
          </div>
        );
      }
export default OutPage;
