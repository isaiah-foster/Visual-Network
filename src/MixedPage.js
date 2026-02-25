import React from 'react';
import { BarChart, Bar, XAxis, YAxis, Tooltip, CartesianGrid, ResponsiveContainer, PieChart, Pie, Cell } from "recharts";

function MixedPage() {
   const data = [
          { name: "Red", value: 700 },
          { name: "Blue", value: 300 },
          { name: "Orange", value: 300 },
          { name: "Green", value: 200 },
          { name: "Yellow", value: 200 }
          ];
      
          const barData = [
          { name: "Example 1", value: 100 },
          { name: "Example 2", value: 300 },
          { name: "Example 3", value: 300 },
          { name: "Example 4", value: 200 },
          { name: "Example 5", value: 200 }
          ]
      
        const COLORS = ["#FF6384", "#36A2EB", "#fe9c00", "#4BC0C0", "#ffe100"];
      
        return (
          <div className="w-full max-w-6xl mx-auto mt-8 px-4">
            <h1 className="text-2xl font-bold text-slate-800 mb-6 text-center ">Incoming/Outcoming Network Overview</h1>
  
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div className="h-96 rounded-xl border border-slate-200 bg-gray-300 p-4 shadow-sm flex flex-col hover:shadow-2xl">
                <p className="text-center mb-2 font-semibold text-lg text-slate-800">Volume In</p>
                <div className="flex-1">
                  <ResponsiveContainer>
                    <PieChart>
                      <Pie data={data} dataKey="value" nameKey="name" outerRadius="72%" label cx="50%" cy="52%">
                        {data.map((entry, index) => (
                          <Cell key={index} fill={COLORS[index % COLORS.length]} />
                        ))}
                      </Pie>
                      <Tooltip />
                    </PieChart>
                  </ResponsiveContainer>
                </div>
              </div>

  
              <div className="h-96 rounded-xl border border-slate-200 bg-gray-300 p-4 shadow-sm flex flex-col hover:shadow-2xl">
                <p className="text-center mb-2 font-semibold text-lg text-slate-800"> Volume Out</p>
                <div className="flex-1">
                  <ResponsiveContainer>
                    <PieChart>
                      <Pie data={data} dataKey="value" nameKey="name" outerRadius="72%" label cx="50%" cy="52%">
                        {data.map((entry, index) => (
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
                    <BarChart data={barData} layout="vertical">
                      <CartesianGrid strokeDasharray="3 3" />
                      <XAxis type="number" />
                      <YAxis type="category" dataKey="name" width={90} />
                      <Tooltip />
                      <Bar dataKey="value" fill="#36A2EB" />
                    </BarChart>
                  </ResponsiveContainer>
                </div>
              </div>

              <div className="h-96 md:col-span-2 md:max-w-2xl md:mx-auto w-full rounded-xl border border-slate-200 bg-gray-300 p-4 shadow-sm flex flex-col hover:shadow-2xl">
                <p className="text-center mb-2 font-semibold text-lg text-slate-800 ">Networking Portocol </p>
                <div className="flex-1">
                  <ResponsiveContainer>
                    <PieChart>
                      <Pie data={data} dataKey="value" nameKey="name" outerRadius="72%" label cx="50%" cy="52%">
                        {data.map((entry, index) => (
                          <Cell key={index} fill={COLORS[index % COLORS.length]} />
                        ))}
                      </Pie>
                      <Tooltip />
                    </PieChart>
                  </ResponsiveContainer>
                </div>
              </div>
            </div>
          </div>
        );
      }
export default MixedPage;
