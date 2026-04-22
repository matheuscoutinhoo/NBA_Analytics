"use client";

import { useEffect, useState } from "react";
import ProtectedLayout from "@/components/ProtectedLayout";
import apiClient from "@/lib/api-client";
import type { Bankroll, BankrollTransaction, BankrollStats } from "@/types";
import { DollarSign, TrendingUp, TrendingDown, ArrowUpCircle, ArrowDownCircle, Loader2 } from "lucide-react";
import { AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from "recharts";

export default function BankrollPage() {
  const [bankroll, setBankroll] = useState<Bankroll | null>(null);
  const [transactions, setTransactions] = useState<BankrollTransaction[]>([]);
  const [stats, setStats] = useState<BankrollStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [amount, setAmount] = useState("");
  const [actionType, setActionType] = useState<"deposit" | "withdraw" | null>(null);

  useEffect(() => {
    fetchData();
  }, []);

  const fetchData = async () => {
    try {
      const [bankrollData, statsData] = await Promise.all([
        apiClient.getBankroll().catch(() => ({ bankroll: null, transactions: [] })),
        apiClient.getBankrollStats().catch(() => null),
      ]);
      setBankroll(bankrollData.bankroll);
      setTransactions(bankrollData.transactions);
      setStats(statsData);
    } finally {
      setLoading(false);
    }
  };

  const handleAction = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!actionType || !amount) return;
    try {
      if (actionType === "deposit") {
        await apiClient.deposit(parseFloat(amount));
      } else {
        await apiClient.withdraw(parseFloat(amount));
      }
      setAmount("");
      setActionType(null);
      fetchData();
    } catch (err) {
      alert(err instanceof Error ? err.message : "Action failed");
    }
  };

  const chartData = transactions
    .slice()
    .reverse()
    .map((tx) => ({
      date: new Date(tx.created_at).toLocaleDateString(),
      balance: tx.balance_after,
    }));

  return (
    <ProtectedLayout>
      <div className="space-y-6">
        <h1 className="text-2xl font-bold text-white">Bankroll Management</h1>

        {loading ? (
          <div className="flex justify-center py-16">
            <Loader2 className="h-8 w-8 animate-spin text-orange-500" />
          </div>
        ) : (
          <>
            {/* Stats cards */}
            <div className="grid grid-cols-1 sm:grid-cols-2 xl:grid-cols-4 gap-4">
              <div className="bg-gray-900 rounded-xl border border-gray-800 p-4">
                <div className="flex items-center justify-between mb-2">
                  <span className="text-sm text-gray-400">Current Balance</span>
                  <DollarSign className="h-5 w-5 text-orange-400" />
                </div>
                <p className="text-2xl font-bold text-white">
                  ${bankroll?.current_amount?.toFixed(2) || "0.00"}
                </p>
              </div>
              <div className="bg-gray-900 rounded-xl border border-gray-800 p-4">
                <div className="flex items-center justify-between mb-2">
                  <span className="text-sm text-gray-400">Win Rate</span>
                  <TrendingUp className="h-5 w-5 text-green-400" />
                </div>
                <p className="text-2xl font-bold text-green-400">{stats?.win_rate || 0}%</p>
              </div>
              <div className="bg-gray-900 rounded-xl border border-gray-800 p-4">
                <div className="flex items-center justify-between mb-2">
                  <span className="text-sm text-gray-400">ROI</span>
                  {stats && stats.roi >= 0 ? (
                    <TrendingUp className="h-5 w-5 text-green-400" />
                  ) : (
                    <TrendingDown className="h-5 w-5 text-red-400" />
                  )}
                </div>
                <p className={`text-2xl font-bold ${stats && stats.roi >= 0 ? "text-green-400" : "text-red-400"}`}>
                  {stats?.roi || 0}%
                </p>
              </div>
              <div className="bg-gray-900 rounded-xl border border-gray-800 p-4">
                <div className="flex items-center justify-between mb-2">
                  <span className="text-sm text-gray-400">P&L</span>
                  <DollarSign className="h-5 w-5 text-orange-400" />
                </div>
                <p className={`text-2xl font-bold ${stats && stats.profit_loss >= 0 ? "text-green-400" : "text-red-400"}`}>
                  {stats && stats.profit_loss >= 0 ? "+" : ""}${stats?.profit_loss?.toFixed(2) || "0.00"}
                </p>
              </div>
            </div>

            {/* Actions */}
            <div className="flex gap-3 flex-wrap">
              <button
                onClick={() => setActionType(actionType === "deposit" ? null : "deposit")}
                className={`flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
                  actionType === "deposit" ? "bg-green-500 text-white" : "bg-gray-800 text-gray-400 hover:text-white"
                }`}
              >
                <ArrowUpCircle className="h-4 w-4" /> Deposit
              </button>
              <button
                onClick={() => setActionType(actionType === "withdraw" ? null : "withdraw")}
                className={`flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
                  actionType === "withdraw" ? "bg-red-500 text-white" : "bg-gray-800 text-gray-400 hover:text-white"
                }`}
              >
                <ArrowDownCircle className="h-4 w-4" /> Withdraw
              </button>
            </div>

            {actionType && (
              <form onSubmit={handleAction} className="bg-gray-900 rounded-xl border border-gray-800 p-4 flex items-end gap-3">
                <div className="flex-1">
                  <label className="block text-sm text-gray-400 mb-1">Amount ($)</label>
                  <input
                    type="number"
                    step="0.01"
                    min="0.01"
                    value={amount}
                    onChange={(e) => setAmount(e.target.value)}
                    required
                    className="w-full px-3 py-2 bg-gray-800 border border-gray-700 rounded-lg text-white placeholder-gray-500"
                    placeholder="100.00"
                  />
                </div>
                <button
                  type="submit"
                  className={`px-6 py-2 font-medium rounded-lg text-white transition-colors ${
                    actionType === "deposit" ? "bg-green-500 hover:bg-green-600" : "bg-red-500 hover:bg-red-600"
                  }`}
                >
                  {actionType === "deposit" ? "Deposit" : "Withdraw"}
                </button>
              </form>
            )}

            {/* Chart */}
            {chartData.length > 1 && (
              <div className="bg-gray-900 rounded-xl border border-gray-800 p-6">
                <h2 className="text-lg font-semibold text-white mb-4">Balance History</h2>
                <div className="h-64">
                  <ResponsiveContainer width="100%" height="100%">
                    <AreaChart data={chartData}>
                      <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
                      <XAxis dataKey="date" stroke="#9CA3AF" fontSize={12} />
                      <YAxis stroke="#9CA3AF" fontSize={12} />
                      <Tooltip
                        contentStyle={{ backgroundColor: "#1F2937", border: "1px solid #374151", borderRadius: "8px" }}
                        labelStyle={{ color: "#9CA3AF" }}
                        itemStyle={{ color: "#F97316" }}
                      />
                      <Area type="monotone" dataKey="balance" stroke="#F97316" fill="#F97316" fillOpacity={0.1} />
                    </AreaChart>
                  </ResponsiveContainer>
                </div>
              </div>
            )}

            {/* Transactions */}
            <div className="bg-gray-900 rounded-xl border border-gray-800 p-6">
              <h2 className="text-lg font-semibold text-white mb-4">Transaction History</h2>
              {transactions.length === 0 ? (
                <p className="text-gray-500 text-center py-4">No transactions yet</p>
              ) : (
                <div className="overflow-x-auto">
                  <table className="w-full text-sm">
                    <thead>
                      <tr className="text-gray-400 border-b border-gray-800">
                        <th className="text-left py-2 px-3">Date</th>
                        <th className="text-left py-2 px-3">Type</th>
                        <th className="text-right py-2 px-3">Amount</th>
                        <th className="text-right py-2 px-3">Balance After</th>
                      </tr>
                    </thead>
                    <tbody>
                      {transactions.map((tx) => (
                        <tr key={tx.id} className="border-b border-gray-800/50">
                          <td className="py-2 px-3 text-gray-400">{new Date(tx.created_at).toLocaleString()}</td>
                          <td className="py-2 px-3">
                            <span className={`px-2 py-0.5 rounded-full text-xs ${
                              tx.type === "deposit" || tx.type === "bet_won" ? "bg-green-500/10 text-green-400" :
                              tx.type === "withdraw" || tx.type === "bet_placed" || tx.type === "bet_lost" ? "bg-red-500/10 text-red-400" :
                              "bg-gray-500/10 text-gray-400"
                            }`}>{tx.type.replace("_", " ")}</span>
                          </td>
                          <td className={`py-2 px-3 text-right ${tx.amount >= 0 ? "text-green-400" : "text-red-400"}`}>
                            {tx.amount >= 0 ? "+" : ""}${tx.amount.toFixed(2)}
                          </td>
                          <td className="py-2 px-3 text-right text-white">${tx.balance_after.toFixed(2)}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}
            </div>
          </>
        )}
      </div>
    </ProtectedLayout>
  );
}
