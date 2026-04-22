"use client";

import { useQuery } from "@tanstack/react-query";
import { useParams } from "next/navigation";
import {
   ArrowLeft,
   AlertTriangle,
   TrendingUp,
   TrendingDown,
   Shield,
   Activity,
} from "lucide-react";
import Link from "next/link";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Progress } from "@/components/ui/progress";
import { api } from "@/lib/api";
import { formatPercentage, formatDateTime, getConfidenceColor } from "@/lib/utils";

export default function GameAnalysisPage() {
   const params = useParams();
   const gameId = params.gameId as string;

   const { data, isLoading } = useQuery({
      queryKey: ["gameAnalysis", gameId],
      queryFn: () => api.getGameAnalysis(gameId),
   });

   const gameData = (data as any)?.data;
   const game = gameData?.game;
   const homeValue = gameData?.home_value;
   const awayValue = gameData?.away_value;

   if (isLoading) {
      return (
         <div className="flex items-center justify-center py-20">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary" />
         </div>
      );
   }

   if (!game) {
      return (
         <div className="text-center py-20">
            <p className="text-muted-foreground">Game not found</p>
            <Link href="/dashboard/games" className="text-primary hover:underline mt-2 block">
               Back to Games
            </Link>
         </div>
      );
   }

   const analysis = game.analysis;

   return (
      <div className="space-y-6 max-w-4xl mx-auto">
         {/* Back button */}
         <Link
            href="/dashboard/games"
            className="inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground"
         >
            <ArrowLeft size={16} />
            Back to Games
         </Link>

         {/* Game header */}
         <Card>
            <CardContent className="pt-6">
               <div className="flex items-center justify-between flex-wrap gap-4">
                  <div className="flex items-center gap-6">
                     <div className="text-center">
                        <div className="w-16 h-16 rounded-xl bg-primary/10 flex items-center justify-center mb-1 mx-auto">
                           <span className="text-xl font-bold">{game.home_team?.abbreviation}</span>
                        </div>
                        <p className="text-sm text-muted-foreground">{game.home_team?.city}</p>
                        <p className="font-semibold">{game.home_team?.name}</p>
                        {game.status === "FINAL" && (
                           <p className="text-3xl font-bold font-mono mt-1">{game.home_score}</p>
                        )}
                     </div>
                     <div className="text-center">
                        <p className="text-xl font-bold text-muted-foreground">VS</p>
                        <p className="text-xs text-muted-foreground mt-1">{formatDateTime(game.game_date)}</p>
                        <p className="text-xs font-medium mt-1">{game.status}</p>
                     </div>
                     <div className="text-center">
                        <div className="w-16 h-16 rounded-xl bg-primary/10 flex items-center justify-center mb-1 mx-auto">
                           <span className="text-xl font-bold">{game.away_team?.abbreviation}</span>
                        </div>
                        <p className="text-sm text-muted-foreground">{game.away_team?.city}</p>
                        <p className="font-semibold">{game.away_team?.name}</p>
                        {game.status === "FINAL" && (
                           <p className="text-3xl font-bold font-mono mt-1">{game.away_score}</p>
                        )}
                     </div>
                  </div>
               </div>
            </CardContent>
         </Card>

         {/* Analysis */}
         {analysis && (
            <div className="grid gap-4 md:grid-cols-2">
               {/* Win Probability */}
               <Card>
                  <CardHeader>
                     <CardTitle className="text-base flex items-center gap-2">
                        <Activity size={16} />
                        Win Probability
                     </CardTitle>
                  </CardHeader>
                  <CardContent className="space-y-4">
                     <div>
                        <div className="flex justify-between text-sm mb-1">
                           <span>{game.home_team?.abbreviation}</span>
                           <span className="font-mono font-semibold">
                              {formatPercentage(analysis.home_win_prob * 100)}
                           </span>
                        </div>
                        <Progress
                           value={analysis.home_win_prob * 100}
                           className="h-3"
                           indicatorClassName="bg-primary"
                        />
                     </div>
                     <div>
                        <div className="flex justify-between text-sm mb-1">
                           <span>{game.away_team?.abbreviation}</span>
                           <span className="font-mono font-semibold">
                              {formatPercentage(analysis.away_win_prob * 100)}
                           </span>
                        </div>
                        <Progress
                           value={analysis.away_win_prob * 100}
                           className="h-3"
                           indicatorClassName="bg-blue-500"
                        />
                     </div>
                     <div className="flex items-center justify-between pt-2 border-t">
                        <span className="text-sm text-muted-foreground">Confidence</span>
                        <span className={`font-semibold ${getConfidenceColor(analysis.confidence)}`}>
                           {analysis.confidence}
                        </span>
                     </div>
                  </CardContent>
               </Card>

               {/* Value Analysis */}
               <Card>
                  <CardHeader>
                     <CardTitle className="text-base flex items-center gap-2">
                        <TrendingUp size={16} />
                        Value Analysis
                     </CardTitle>
                  </CardHeader>
                  <CardContent>
                     <div className="space-y-4">
                        {homeValue && (
                           <div className="space-y-2">
                              <p className="text-sm font-medium">{game.home_team?.abbreviation}</p>
                              <div className="grid grid-cols-2 gap-2 text-sm">
                                 <div className="bg-muted/50 rounded p-2">
                                    <p className="text-xs text-muted-foreground">Model</p>
                                    <p className="font-mono">{formatPercentage(homeValue.model_prob * 100)}</p>
                                 </div>
                                 <div className="bg-muted/50 rounded p-2">
                                    <p className="text-xs text-muted-foreground">Market</p>
                                    <p className="font-mono">{formatPercentage(homeValue.implied_prob * 100)}</p>
                                 </div>
                                 <div className="bg-muted/50 rounded p-2">
                                    <p className="text-xs text-muted-foreground">EV</p>
                                    <p className={`font-mono ${homeValue.ev > 0 ? "text-profit" : "text-loss"}`}>
                                       {homeValue.ev > 0 ? "+" : ""}{(homeValue.ev * 100).toFixed(2)}%
                                    </p>
                                 </div>
                                 <div className="bg-muted/50 rounded p-2">
                                    <p className="text-xs text-muted-foreground">Divergence</p>
                                    <p className="font-mono">{homeValue.divergence_pct}%</p>
                                 </div>
                              </div>
                              {homeValue.is_value && (
                                 <div className="flex items-center gap-1 text-xs text-profit">
                                    <TrendingUp size={12} />
                                    <span>Potential value detected</span>
                                 </div>
                              )}
                           </div>
                        )}

                        {awayValue && (
                           <div className="space-y-2 pt-2 border-t">
                              <p className="text-sm font-medium">{game.away_team?.abbreviation}</p>
                              <div className="grid grid-cols-2 gap-2 text-sm">
                                 <div className="bg-muted/50 rounded p-2">
                                    <p className="text-xs text-muted-foreground">Model</p>
                                    <p className="font-mono">{formatPercentage(awayValue.model_prob * 100)}</p>
                                 </div>
                                 <div className="bg-muted/50 rounded p-2">
                                    <p className="text-xs text-muted-foreground">Market</p>
                                    <p className="font-mono">{formatPercentage(awayValue.implied_prob * 100)}</p>
                                 </div>
                                 <div className="bg-muted/50 rounded p-2">
                                    <p className="text-xs text-muted-foreground">EV</p>
                                    <p className={`font-mono ${awayValue.ev > 0 ? "text-profit" : "text-loss"}`}>
                                       {awayValue.ev > 0 ? "+" : ""}{(awayValue.ev * 100).toFixed(2)}%
                                    </p>
                                 </div>
                                 <div className="bg-muted/50 rounded p-2">
                                    <p className="text-xs text-muted-foreground">Divergence</p>
                                    <p className="font-mono">{awayValue.divergence_pct}%</p>
                                 </div>
                              </div>
                              {awayValue.is_value && (
                                 <div className="flex items-center gap-1 text-xs text-profit">
                                    <TrendingUp size={12} />
                                    <span>Potential value detected</span>
                                 </div>
                              )}
                           </div>
                        )}
                     </div>
                  </CardContent>
               </Card>
            </div>
         )}

         {/* Justification */}
         {analysis?.justification && (
            <Card>
               <CardHeader>
                  <CardTitle className="text-base">Analysis Summary</CardTitle>
               </CardHeader>
               <CardContent>
                  <p className="text-sm text-muted-foreground leading-relaxed">
                     {analysis.justification}
                  </p>
               </CardContent>
            </Card>
         )}

         {/* Risk Alerts */}
         {analysis?.risk_alerts && analysis.risk_alerts.length > 0 && (
            <Card>
               <CardHeader>
                  <CardTitle className="text-base flex items-center gap-2">
                     <AlertTriangle size={16} className="text-yellow-500" />
                     Risk Alerts
                  </CardTitle>
               </CardHeader>
               <CardContent>
                  <ul className="space-y-2">
                     {analysis.risk_alerts.map((alert: string, i: number) => (
                        <li key={i} className="flex items-start gap-2 text-sm">
                           <AlertTriangle size={14} className="text-yellow-500 mt-0.5 flex-shrink-0" />
                           <span className="text-muted-foreground">{alert}</span>
                        </li>
                     ))}
                  </ul>
               </CardContent>
            </Card>
         )}

         {/* Odds */}
         {game.latest_odds && (
            <Card>
               <CardHeader>
                  <CardTitle className="text-base">Market Odds</CardTitle>
               </CardHeader>
               <CardContent>
                  <div className="overflow-x-auto">
                     <table className="w-full text-sm">
                        <thead>
                           <tr className="border-b">
                              <th className="text-left py-2 text-muted-foreground font-medium">Market</th>
                              <th className="text-center py-2 text-muted-foreground font-medium">{game.home_team?.abbreviation}</th>
                              <th className="text-center py-2 text-muted-foreground font-medium">{game.away_team?.abbreviation}</th>
                           </tr>
                        </thead>
                        <tbody className="font-mono">
                           <tr className="border-b">
                              <td className="py-2">Moneyline</td>
                              <td className="text-center py-2">{game.latest_odds.home_moneyline?.toFixed(2)}</td>
                              <td className="text-center py-2">{game.latest_odds.away_moneyline?.toFixed(2)}</td>
                           </tr>
                           <tr className="border-b">
                              <td className="py-2">Spread</td>
                              <td className="text-center py-2">{game.latest_odds.home_spread > 0 ? "+" : ""}{game.latest_odds.home_spread}</td>
                              <td className="text-center py-2">{game.latest_odds.away_spread > 0 ? "+" : ""}{game.latest_odds.away_spread}</td>
                           </tr>
                           <tr className="border-b">
                              <td className="py-2">Implied Probability</td>
                              <td className="text-center py-2">{formatPercentage(game.latest_odds.home_implied_prob * 100)}</td>
                              <td className="text-center py-2">{formatPercentage(game.latest_odds.away_implied_prob * 100)}</td>
                           </tr>
                           <tr>
                              <td className="py-2">Over/Under</td>
                              <td className="text-center py-2" colSpan={2}>{game.latest_odds.over_under}</td>
                           </tr>
                        </tbody>
                     </table>
                  </div>
                  <p className="text-xs text-muted-foreground mt-3">
                     Source: {game.latest_odds.source} · Updated: {formatDateTime(game.latest_odds.fetched_at)}
                  </p>
               </CardContent>
            </Card>
         )}

         {/* Injuries */}
         {((game.home_injuries?.length > 0) || (game.away_injuries?.length > 0)) && (
            <Card>
               <CardHeader>
                  <CardTitle className="text-base flex items-center gap-2">
                     <TrendingDown size={16} className="text-loss" />
                     Injury Report
                  </CardTitle>
               </CardHeader>
               <CardContent>
                  <div className="grid md:grid-cols-2 gap-4">
                     <div>
                        <p className="text-sm font-medium mb-2">{game.home_team?.name}</p>
                        {game.home_injuries?.length > 0 ? (
                           <ul className="space-y-1">
                              {game.home_injuries.map((inj: any, i: number) => (
                                 <li key={i} className="text-sm text-muted-foreground">
                                    <span className="font-medium text-foreground">{inj.player_name}</span>
                                    {" — "}<span className="text-loss">{inj.status}</span>
                                    {inj.description && ` (${inj.description})`}
                                 </li>
                              ))}
                           </ul>
                        ) : (
                           <p className="text-sm text-muted-foreground">No reported injuries</p>
                        )}
                     </div>
                     <div>
                        <p className="text-sm font-medium mb-2">{game.away_team?.name}</p>
                        {game.away_injuries?.length > 0 ? (
                           <ul className="space-y-1">
                              {game.away_injuries.map((inj: any, i: number) => (
                                 <li key={i} className="text-sm text-muted-foreground">
                                    <span className="font-medium text-foreground">{inj.player_name}</span>
                                    {" — "}<span className="text-loss">{inj.status}</span>
                                    {inj.description && ` (${inj.description})`}
                                 </li>
                              ))}
                           </ul>
                        ) : (
                           <p className="text-sm text-muted-foreground">No reported injuries</p>
                        )}
                     </div>
                  </div>
               </CardContent>
            </Card>
         )}

         {/* Disclaimer */}
         <div className="flex items-start gap-2 p-4 bg-muted/30 rounded-lg">
            <Shield size={16} className="text-muted-foreground mt-0.5 flex-shrink-0" />
            <div className="text-xs text-muted-foreground space-y-1">
               <p>{gameData?.disclaimer}</p>
               <p>{gameData?.responsible_gaming}</p>
            </div>
         </div>
      </div>
   );
}
