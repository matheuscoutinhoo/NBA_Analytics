"use client";

import { useEffect, useState } from "react";
import ProtectedLayout from "@/components/ProtectedLayout";
import apiClient from "@/lib/api-client";
import type { NBAGame, BankrollStats, AIPrediction } from "@/types";
import { TrendingUp, TrendingDown, Target, DollarSign, Brain } from "lucide-react";
import Link from "next/link";
import PredictionCard from "@/components/PredictionCard";

export default function DashboardPage() {
   const [upcomingGames, setUpcomingGames] = useState<NBAGame[]>([]);
   const [recentGames, setRecentGames] = useState<NBAGame[]>([]);
   const [stats, setStats] = useState<BankrollStats | null>(null);
   const [predictions, setPredictions] = useState<AIPrediction[]>([]);
   const [loading, setLoading] = useState(true);

   useEffect(() => {
      async function fetchData() {
         try {
            const [upcoming, recent, bankrollStats, preds] = await Promise.all([
               apiClient.getUpcomingGames().catch(() => []),
               apiClient.getRecentGames().catch(() => []),
               apiClient.getBankrollStats().catch(() => null),
               apiClient.getPredictions().catch(() => []),
            ]);
            setUpcomingGames(upcoming);
            setRecentGames(recent.slice(0, 10));
            setStats(bankrollStats);
            setPredictions(preds);
         } finally {
            setLoading(false);
         }
      }
      fetchData();
   }, []);

   return (
      <ProtectedLayout>
         <div className="space-y-6">
            <h1 className="text-2xl font-bold text-white">Dashboard</h1>

            {/* Stats cards */}
            <div className="grid grid-cols-1 sm:grid-cols-2 xl:grid-cols-4 gap-4">
               <StatCard
                  title="Win Rate"
                  value={stats ? `${stats.win_rate}%` : "—"}
                  icon={<Target className="h-5 w-5" />}
                  color="text-green-400"
               />
               <StatCard
                  title="ROI"
                  value={stats ? `${stats.roi}%` : "—"}
                  icon={stats && stats.roi >= 0 ? <TrendingUp className="h-5 w-5" /> : <TrendingDown className="h-5 w-5" />}
                  color={stats && stats.roi >= 0 ? "text-green-400" : "text-red-400"}
               />
               <StatCard
                  title="P&L"
                  value={stats ? `$${stats.profit_loss.toFixed(2)}` : "—"}
                  icon={<DollarSign className="h-5 w-5" />}
                  color={stats && stats.profit_loss >= 0 ? "text-green-400" : "text-red-400"}
               />
               <StatCard
                  title="Total Bets"
                  value={stats ? stats.total_bets.toString() : "0"}
                  icon={<Receipt className="h-5 w-5" />}
                  color="text-orange-400"
               />
            </div>

            {/* Upcoming games */}
            <div className="bg-gray-900 rounded-xl border border-gray-800 p-6">
               <div className="flex items-center justify-between mb-4">
                  <h2 className="text-lg font-semibold text-white">Upcoming Games</h2>
                  <Link href="/games" className="text-sm text-orange-500 hover:text-orange-400">
                     View all →
                  </Link>
               </div>
               {loading ? (
                  <div className="flex justify-center py-8">
                     <div className="animate-spin rounded-full h-8 w-8 border-t-2 border-b-2 border-orange-500" />
                  </div>
               ) : upcomingGames.length === 0 ? (
                  <p className="text-gray-500 text-center py-8">No upcoming games found</p>
               ) : (
                  <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-3">
                     {upcomingGames.slice(0, 6).map((game) => (
                        <GameCard key={game.id} game={game} />
                     ))}
                  </div>
               )}
            </div>

            {/* AI Win Predictions */}
            <div className="bg-gray-900 rounded-xl border border-gray-800 p-6">
               <div className="flex items-center justify-between mb-4">
                  <div className="flex items-center gap-2">
                     <Brain className="h-5 w-5 text-purple-400" />
                     <h2 className="text-lg font-semibold text-white">AI Win Predictions</h2>
                  </div>
                  {predictions.length > 0 && (
                     <span className="text-xs text-gray-500">
                        Updated {new Date(predictions[0].analyzed_at).toLocaleString()}
                     </span>
                  )}
               </div>
               {loading ? (
                  <div className="flex justify-center py-8">
                     <div className="animate-spin rounded-full h-8 w-8 border-t-2 border-b-2 border-purple-500" />
                  </div>
               ) : predictions.length === 0 ? (
                  <div className="text-center py-8">
                     <Brain className="h-10 w-10 text-purple-500/30 mx-auto mb-3" />
                     <p className="text-gray-400 font-medium">AI analysis in progress...</p>
                     <p className="text-gray-600 text-xs mt-1">Predictions update every hour using recent games + Bet365 odds</p>
                  </div>
               ) : (
                  (() => {
                     const topPreds = predictions
                        .filter((p) => Math.max(p.home_win_prob, p.away_win_prob) >= 0.70)
                        .sort((a, b) => Math.max(b.home_win_prob, b.away_win_prob) - Math.max(a.home_win_prob, a.away_win_prob));
                     return topPreds.length === 0 ? (
                        <div className="text-center py-8">
                           <Brain className="h-10 w-10 text-purple-500/30 mx-auto mb-3" />
                           <p className="text-gray-400 font-medium">No high-confidence predictions right now</p>
                           <p className="text-gray-600 text-xs mt-1">Showing only predictions with 70%+ win probability</p>
                        </div>
                     ) : (
                        <div className="grid grid-cols-1 lg:grid-cols-2 gap-3">
                           {topPreds.map((pred) => (
                              <PredictionCard key={pred.id} pred={pred} />
                           ))}
                        </div>
                     );
                  })()
               )}
            </div>

            {/* Recent results */}
            <div className="bg-gray-900 rounded-xl border border-gray-800 p-6">
               <div className="flex items-center justify-between mb-4">
                  <h2 className="text-lg font-semibold text-white">Recent Results</h2>
                  <Link href="/games" className="text-sm text-orange-500 hover:text-orange-400">
                     View all →
                  </Link>
               </div>
               {loading ? (
                  <div className="flex justify-center py-8">
                     <div className="animate-spin rounded-full h-8 w-8 border-t-2 border-b-2 border-orange-500" />
                  </div>
               ) : (
                  <div className="overflow-x-auto">
                     <table className="w-full text-sm">
                        <thead>
                           <tr className="text-gray-400 border-b border-gray-800">
                              <th className="text-left py-2 px-3">Date</th>
                              <th className="text-left py-2 px-3">Matchup</th>
                              <th className="text-center py-2 px-3">Score</th>
                              <th className="text-center py-2 px-3">Status</th>
                           </tr>
                        </thead>
                        <tbody>
                           {recentGames.map((game) => (
                              <tr key={game.id} className="border-b border-gray-800/50 hover:bg-gray-800/30">
                                 <td className="py-2 px-3 text-gray-400">
                                    {new Date(game.game_date).toLocaleDateString()}
                                 </td>
                                 <td className="py-2 px-3 text-white">
                                    {game.away_team} @ {game.home_team}
                                 </td>
                                 <td className="py-2 px-3 text-center text-white">
                                    {game.status === "final"
                                       ? `${game.away_score} - ${game.home_score}`
                                       : "—"}
                                 </td>
                                 <td className="py-2 px-3 text-center">
                                    <span className={`px-2 py-0.5 rounded-full text-xs ${game.status === "final" ? "bg-green-500/10 text-green-400" :
                                       game.status === "live" ? "bg-red-500/10 text-red-400" :
                                          "bg-gray-500/10 text-gray-400"
                                       }`}>
                                       {game.status}
                                    </span>
                                 </td>
                              </tr>
                           ))}
                        </tbody>
                     </table>
                  </div>
               )}
            </div>
         </div>
      </ProtectedLayout>
   );
}

function Receipt({ className }: { className?: string }) {
   return (
      <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className={className}>
         <path d="M4 2v20l2-1 2 1 2-1 2 1 2-1 2 1 2-1 2 1V2l-2 1-2-1-2 1-2-1-2 1-2-1-2 1Z" />
         <path d="M14 8H8" /><path d="M16 12H8" /><path d="M13 16H8" />
      </svg>
   );
}

function StatCard({ title, value, icon, color }: { title: string; value: string; icon: React.ReactNode; color: string }) {
   return (
      <div className="bg-gray-900 rounded-xl border border-gray-800 p-4">
         <div className="flex items-center justify-between mb-2">
            <span className="text-sm text-gray-400">{title}</span>
            <span className={color}>{icon}</span>
         </div>
         <p className={`text-2xl font-bold ${color}`}>{value}</p>
      </div>
   );
}

function GameCard({ game }: { game: NBAGame }) {
   return (
      <Link href={`/games`} className="block bg-gray-800/50 rounded-lg p-4 hover:bg-gray-800 transition-colors border border-gray-700/50">
         <div className="text-xs text-gray-500 mb-2">
            {new Date(game.game_date).toLocaleDateString(undefined, {
               weekday: "short",
               month: "short",
               day: "numeric",
               hour: "2-digit",
               minute: "2-digit",
            })}
         </div>
         <div className="flex items-center justify-between">
            <div className="text-sm font-medium text-white truncate flex-1">{game.away_team}</div>
            <span className="text-xs text-gray-500 mx-2">@</span>
            <div className="text-sm font-medium text-white truncate flex-1 text-right">{game.home_team}</div>
         </div>
      </Link>
   );
}
