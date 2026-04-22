"use client";

import { useQuery } from "@tanstack/react-query";
import Link from "next/link";
import { BarChart3, TrendingUp, AlertTriangle, ArrowRight } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { api } from "@/lib/api";
import { formatPercentage, formatDateTime } from "@/lib/utils";

export default function GamesPage() {
   const { data, isLoading } = useQuery({
      queryKey: ["todaysGames"],
      queryFn: () => api.getTodaysGames(),
   });

   const games = (data as any)?.data?.games || [];
   const date = (data as any)?.data?.date;

   if (isLoading) {
      return (
         <div className="flex items-center justify-center py-20">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary" />
         </div>
      );
   }

   return (
      <div className="space-y-6">
         <div>
            <h1 className="text-2xl font-bold">Games & Analysis</h1>
            <p className="text-muted-foreground">
               {date || "Today"} — {games.length} game{games.length !== 1 ? "s" : ""} scheduled
            </p>
         </div>

         {games.length === 0 ? (
            <Card>
               <CardContent className="py-16 text-center">
                  <BarChart3 className="w-16 h-16 mx-auto mb-4 text-muted-foreground opacity-50" />
                  <h3 className="text-lg font-semibold mb-2">No Games Today</h3>
                  <p className="text-muted-foreground">Check back when there are scheduled games.</p>
               </CardContent>
            </Card>
         ) : (
            <div className="grid gap-4 md:grid-cols-2">
               {games.map((game: any) => (
                  <Link key={game.id} href={`/dashboard/games/${game.id}`}>
                     <Card className="hover:border-primary/50 transition-colors cursor-pointer h-full">
                        <CardHeader className="pb-2">
                           <div className="flex items-center justify-between">
                              <CardTitle className="text-base">
                                 {game.home_team?.city} {game.home_team?.name} vs {game.away_team?.city} {game.away_team?.name}
                              </CardTitle>
                              <ArrowRight size={16} className="text-muted-foreground" />
                           </div>
                           <p className="text-xs text-muted-foreground">
                              {formatDateTime(game.game_date)} · {game.status}
                           </p>
                        </CardHeader>
                        <CardContent>
                           <div className="space-y-3">
                              {/* Score if final */}
                              {game.status === "FINAL" && game.home_score != null && (
                                 <div className="flex items-center justify-center gap-4 py-2">
                                    <div className="text-center">
                                       <p className="text-2xl font-bold font-mono">{game.home_score}</p>
                                       <p className="text-xs text-muted-foreground">{game.home_team?.abbreviation}</p>
                                    </div>
                                    <span className="text-muted-foreground">-</span>
                                    <div className="text-center">
                                       <p className="text-2xl font-bold font-mono">{game.away_score}</p>
                                       <p className="text-xs text-muted-foreground">{game.away_team?.abbreviation}</p>
                                    </div>
                                 </div>
                              )}

                              {/* Analysis preview */}
                              {game.analysis && (
                                 <div className="bg-muted/50 rounded-md p-3">
                                    <div className="flex items-center justify-between mb-2">
                                       <span className="text-xs text-muted-foreground">Win Probability</span>
                                       <span className={`text-xs font-semibold ${game.analysis.confidence === "HIGH" ? "text-profit" :
                                             game.analysis.confidence === "MEDIUM" ? "text-yellow-500" : "text-loss"
                                          }`}>
                                          {game.analysis.confidence}
                                       </span>
                                    </div>
                                    <div className="flex justify-between text-sm">
                                       <div>
                                          <span className="text-muted-foreground">{game.home_team?.abbreviation}: </span>
                                          <span className="font-mono font-semibold">
                                             {formatPercentage(game.analysis.home_win_prob * 100)}
                                          </span>
                                       </div>
                                       <div>
                                          <span className="text-muted-foreground">{game.away_team?.abbreviation}: </span>
                                          <span className="font-mono font-semibold">
                                             {formatPercentage(game.analysis.away_win_prob * 100)}
                                          </span>
                                       </div>
                                    </div>
                                 </div>
                              )}

                              {/* Odds preview */}
                              {game.latest_odds && (
                                 <div className="flex justify-between text-xs text-muted-foreground">
                                    <span>ML: {game.latest_odds.home_moneyline?.toFixed(2)} / {game.latest_odds.away_moneyline?.toFixed(2)}</span>
                                    <span>O/U: {game.latest_odds.over_under}</span>
                                 </div>
                              )}

                              {/* Value indicators */}
                              {game.analysis && (game.analysis.is_home_value || game.analysis.is_away_value) && (
                                 <div className="flex items-center gap-1.5 text-xs">
                                    <TrendingUp size={12} className="text-profit" />
                                    <span className="text-profit font-medium">
                                       Value detected: {game.analysis.is_home_value ? game.home_team?.abbreviation : ''}
                                       {game.analysis.is_home_value && game.analysis.is_away_value ? ' & ' : ''}
                                       {game.analysis.is_away_value ? game.away_team?.abbreviation : ''}
                                    </span>
                                 </div>
                              )}
                           </div>
                        </CardContent>
                     </Card>
                  </Link>
               ))}
            </div>
         )}

         {/* Disclaimer */}
         <div className="flex items-start gap-2 p-4 bg-muted/30 rounded-lg">
            <AlertTriangle size={16} className="text-yellow-500 mt-0.5 flex-shrink-0" />
            <p className="text-xs text-muted-foreground">
               Odds displayed are from publicly authorized providers. Analysis is probabilistic and
               for informational purposes only. No guarantees of accuracy or returns.
            </p>
         </div>
      </div>
   );
}
