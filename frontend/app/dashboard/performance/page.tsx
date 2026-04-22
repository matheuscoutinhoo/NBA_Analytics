"use client";

import { useQuery } from "@tanstack/react-query";
import {
   TrendingUp,
   TrendingDown,
   Target,
   DollarSign,
   BarChart3,
   Activity,
} from "lucide-react";
import {
   LineChart,
   Line,
   XAxis,
   YAxis,
   CartesianGrid,
   Tooltip,
   ResponsiveContainer,
   BarChart,
   Bar,
} from "recharts";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { api } from "@/lib/api";
import { formatCurrency, formatPercentage } from "@/lib/utils";

export default function PerformancePage() {
   const { data, isLoading } = useQuery({
      queryKey: ["dashboard"],
      queryFn: () => api.getDashboard(),
   });

   const stats = (data as any)?.data;

   if (isLoading) {
      return (
         <div className="flex items-center justify-center py-20">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary" />
         </div>
      );
   }

   if (!stats) {
      return (
         <div className="text-center py-20">
            <BarChart3 className="w-16 h-16 mx-auto mb-4 text-muted-foreground opacity-50" />
            <h3 className="text-lg font-semibold mb-2">No Data Yet</h3>
            <p className="text-muted-foreground">Start tracking your bankroll to see performance metrics.</p>
         </div>
      );
   }

   return (
      <div className="space-y-6">
         <div>
            <h1 className="text-2xl font-bold">Performance Dashboard</h1>
            <p className="text-muted-foreground">Your analytics and performance metrics</p>
         </div>

         {/* Key metrics */}
         <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
            <Card>
               <CardContent className="pt-6">
                  <div className="flex items-center gap-2 text-muted-foreground text-sm mb-1">
                     <DollarSign size={14} />
                     <span>Balance</span>
                  </div>
                  <p className="text-2xl font-bold font-mono">{formatCurrency(stats.current_balance)}</p>
                  <p className={`text-xs mt-1 ${stats.profit_loss >= 0 ? "text-profit" : "text-loss"}`}>
                     {stats.profit_loss >= 0 ? "+" : ""}{formatCurrency(stats.profit_loss)} P/L
                  </p>
               </CardContent>
            </Card>

            <Card>
               <CardContent className="pt-6">
                  <div className="flex items-center gap-2 text-muted-foreground text-sm mb-1">
                     <TrendingUp size={14} />
                     <span>ROI</span>
                  </div>
                  <p className={`text-2xl font-bold font-mono ${stats.roi >= 0 ? "text-profit" : "text-loss"}`}>
                     {stats.roi >= 0 ? "+" : ""}{formatPercentage(stats.roi)}
                  </p>
               </CardContent>
            </Card>

            <Card>
               <CardContent className="pt-6">
                  <div className="flex items-center gap-2 text-muted-foreground text-sm mb-1">
                     <Target size={14} />
                     <span>Win Rate</span>
                  </div>
                  <p className="text-2xl font-bold font-mono">{formatPercentage(stats.win_rate)}</p>
                  <p className="text-xs text-muted-foreground mt-1">
                     {stats.win_count}W / {stats.loss_count}L
                  </p>
               </CardContent>
            </Card>

            <Card>
               <CardContent className="pt-6">
                  <div className="flex items-center gap-2 text-muted-foreground text-sm mb-1">
                     <TrendingDown size={14} />
                     <span>Max Drawdown</span>
                  </div>
                  <p className="text-2xl font-bold font-mono text-loss">
                     {formatPercentage(stats.max_drawdown)}
                  </p>
               </CardContent>
            </Card>
         </div>

         {/* Secondary metrics */}
         <div className="grid grid-cols-3 gap-4">
            <Card>
               <CardContent className="pt-6">
                  <p className="text-sm text-muted-foreground">Total Deposits</p>
                  <p className="text-lg font-bold font-mono">{formatCurrency(stats.total_deposits)}</p>
               </CardContent>
            </Card>
            <Card>
               <CardContent className="pt-6">
                  <p className="text-sm text-muted-foreground">Total Withdrawals</p>
                  <p className="text-lg font-bold font-mono">{formatCurrency(stats.total_withdrawals)}</p>
               </CardContent>
            </Card>
            <Card>
               <CardContent className="pt-6">
                  <p className="text-sm text-muted-foreground">Total Bets</p>
                  <p className="text-lg font-bold font-mono">{stats.total_bets}</p>
               </CardContent>
            </Card>
         </div>

         {/* Balance evolution chart */}
         {stats.balance_history && stats.balance_history.length > 0 && (
            <Card>
               <CardHeader>
                  <CardTitle className="text-base flex items-center gap-2">
                     <Activity size={16} />
                     Balance Evolution
                  </CardTitle>
               </CardHeader>
               <CardContent>
                  <div className="h-64 sm:h-80">
                     <ResponsiveContainer width="100%" height="100%">
                        <LineChart data={stats.balance_history}>
                           <CartesianGrid strokeDasharray="3 3" stroke="hsl(var(--border))" />
                           <XAxis
                              dataKey="date"
                              tick={{ fontSize: 12 }}
                              stroke="hsl(var(--muted-foreground))"
                           />
                           <YAxis
                              tick={{ fontSize: 12 }}
                              stroke="hsl(var(--muted-foreground))"
                              tickFormatter={(v) => `$${v}`}
                           />
                           <Tooltip
                              contentStyle={{
                                 backgroundColor: "hsl(var(--card))",
                                 border: "1px solid hsl(var(--border))",
                                 borderRadius: "8px",
                                 fontSize: "12px",
                              }}
                              formatter={(value: number) => [formatCurrency(value), "Balance"]}
                           />
                           <Line
                              type="monotone"
                              dataKey="balance"
                              stroke="#3b82f6"
                              strokeWidth={2}
                              dot={false}
                           />
                        </LineChart>
                     </ResponsiveContainer>
                  </div>
               </CardContent>
            </Card>
         )}

         {/* Monthly performance */}
         {stats.monthly_stats && stats.monthly_stats.length > 0 && (
            <Card>
               <CardHeader>
                  <CardTitle className="text-base">Monthly Performance</CardTitle>
               </CardHeader>
               <CardContent>
                  <div className="h-64">
                     <ResponsiveContainer width="100%" height="100%">
                        <BarChart data={stats.monthly_stats}>
                           <CartesianGrid strokeDasharray="3 3" stroke="hsl(var(--border))" />
                           <XAxis
                              dataKey="month"
                              tick={{ fontSize: 12 }}
                              stroke="hsl(var(--muted-foreground))"
                           />
                           <YAxis
                              tick={{ fontSize: 12 }}
                              stroke="hsl(var(--muted-foreground))"
                              tickFormatter={(v) => `$${v}`}
                           />
                           <Tooltip
                              contentStyle={{
                                 backgroundColor: "hsl(var(--card))",
                                 border: "1px solid hsl(var(--border))",
                                 borderRadius: "8px",
                                 fontSize: "12px",
                              }}
                              formatter={(value: number) => [formatCurrency(value), "Profit"]}
                           />
                           <Bar
                              dataKey="profit"
                              fill="#3b82f6"
                              radius={[4, 4, 0, 0]}
                           />
                        </BarChart>
                     </ResponsiveContainer>
                  </div>
               </CardContent>
            </Card>
         )}

         {/* Disclaimer */}
         <p className="text-xs text-muted-foreground text-center">
            Performance data is based on self-reported entries. Past performance does not guarantee future results.
         </p>
      </div>
   );
}
