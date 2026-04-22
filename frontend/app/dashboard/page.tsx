"use client";

import { useQuery } from "@tanstack/react-query";
import { BarChart3, TrendingUp, Wallet, Trophy, AlertTriangle } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { useAuthStore } from "@/store";
import { api } from "@/lib/api";
import { formatCurrency, formatPercentage } from "@/lib/utils";
import Link from "next/link";

export default function DashboardPage() {
   const { user } = useAuthStore();

   const { data: gamesData } = useQuery({
      queryKey: ["todaysGames"],
      queryFn: () => api.getTodaysGames(),
   });

   const { data: dashboardData } = useQuery({
      queryKey: ["dashboard"],
      queryFn: () => api.getDashboard(),
   });

   const games = (gamesData as any)?.data?.games || [];
   const stats = (dashboardData as any)?.data;

   return (
      <div className="space-y-6">
         {/* Welcome header */}
         <div>
            <h1 className="text-2xl font-bold">
               Welcome, {user?.username}
            </h1>
            <p className="text-muted-foreground">
               Here&apos;s your analytics overview for today
            </p>
         </div>

         {/* Stats grid */}
         <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
            <Card>
               <CardContent className="pt-6">
                  <div className="flex items-center gap-2 text-muted-foreground text-sm mb-1">
                     <BarChart3 size={14} />
                     <span>Today&apos;s Games</span>
                  </div>
                  <p className="text-3xl font-bold font-mono">{games.length}</p>
               </CardContent>
            </Card>

            <Card>
               <CardContent className="pt-6">
                  <div className="flex items-center gap-2 text-muted-foreground text-sm mb-1">
                     <Wallet size={14} />
                     <span>Balance</span>
                  </div>
                  <p className="text-3xl font-bold font-mono">
                     {stats ? formatCurrency(stats.current_balance) : "$0.00"}
                  </p>
               </CardContent>
            </Card>

            <Card>
               <CardContent className="pt-6">
                  <div className="flex items-center gap-2 text-muted-foreground text-sm mb-1">
                     <TrendingUp size={14} />
                     <span>ROI</span>
                  </div>
                  <p className={`text-3xl font-bold font-mono ${stats?.roi > 0 ? "text-profit" : stats?.roi < 0 ? "text-loss" : ""}`}>
                     {stats ? formatPercentage(stats.roi) : "0.00%"}
                  </p>
               </CardContent>
            </Card>

            <Card>
               <CardContent className="pt-6">
                  <div className="flex items-center gap-2 text-muted-foreground text-sm mb-1">
                     <Trophy size={14} />
                     <span>Win Rate</span>
                  </div>
                  <p className="text-3xl font-bold font-mono">
                     {stats ? formatPercentage(stats.win_rate) : "0.00%"}
                  </p>
               </CardContent>
            </Card>
         </div>

         {/* Today's games preview */}
         <Card>
            <CardHeader className="flex flex-row items-center justify-between">
               <CardTitle className="text-lg">Today&apos;s Games</CardTitle>
               <Link href="/dashboard/games" className="text-sm text-primary hover:underline">
                  View All
               </Link>
            </CardHeader>
            <CardContent>
               {games.length === 0 ? (
                  <div className="text-center py-8 text-muted-foreground">
                     <BarChart3 className="w-10 h-10 mx-auto mb-2 opacity-50" />
                     <p>No games scheduled for today</p>
                  </div>
               ) : (
                  <div className="space-y-3">
                     {games.slice(0, 5).map((game: any) => (
                        <Link
                           key={game.id}
                           href={`/dashboard/games/${game.id}`}
                           className="flex items-center justify-between p-3 rounded-lg border hover:border-primary/50 transition-colors"
                        >
                           <div className="flex items-center gap-4">
                              <div className="text-right min-w-[80px]">
                                 <p className="font-semibold text-sm">
                                    {game.home_team?.abbreviation}
                                 </p>
                              </div>
                              <span className="text-xs text-muted-foreground">vs</span>
                              <div className="min-w-[80px]">
                                 <p className="font-semibold text-sm">
                                    {game.away_team?.abbreviation}
                                 </p>
                              </div>
                           </div>
                           <div className="flex items-center gap-4">
                              {game.analysis && (
                                 <div className="text-right">
                                    <p className="text-xs text-muted-foreground">Confidence</p>
                                    <p className={`text-sm font-semibold ${game.analysis.confidence === "HIGH" ? "text-profit" :
                                          game.analysis.confidence === "MEDIUM" ? "text-yellow-500" : "text-loss"
                                       }`}>
                                       {game.analysis.confidence}
                                    </p>
                                 </div>
                              )}
                              <div className="text-right">
                                 <p className="text-xs text-muted-foreground">Status</p>
                                 <p className="text-sm">{game.status}</p>
                              </div>
                           </div>
                        </Link>
                     ))}
                  </div>
               )}
            </CardContent>
         </Card>

         {/* Disclaimer */}
         <div className="flex items-start gap-2 p-4 bg-muted/30 rounded-lg">
            <AlertTriangle size={16} className="text-yellow-500 mt-0.5 flex-shrink-0" />
            <p className="text-xs text-muted-foreground">
               <strong>Disclaimer:</strong> This platform provides statistical analysis for informational
               purposes only. It does not constitute financial advice and does not guarantee returns.
               Past performance is not indicative of future results.
            </p>
         </div>
      </div>
   );
}
