"use client";

import { useEffect, useState } from "react";
import ProtectedLayout from "@/components/ProtectedLayout";
import apiClient from "@/lib/api-client";
import type { NBAGame, GameOdds, AIInsight } from "@/types";
import { Loader2, Sparkles, TrendingUp, Shield, Zap, ChevronDown, ChevronUp } from "lucide-react";

export default function GamesPage() {
   const [upcoming, setUpcoming] = useState<NBAGame[]>([]);
   const [recent, setRecent] = useState<NBAGame[]>([]);
   const [tab, setTab] = useState<"upcoming" | "recent">("upcoming");
   const [loading, setLoading] = useState(true);
   const [selectedGame, setSelectedGame] = useState<NBAGame | null>(null);
   const [odds, setOdds] = useState<GameOdds[]>([]);
   const [insight, setInsight] = useState<AIInsight | null>(null);
   const [insightLoading, setInsightLoading] = useState(false);

   useEffect(() => {
      async function fetchGames() {
         try {
            const [upcomingData, recentData] = await Promise.all([
               apiClient.getUpcomingGames().catch(() => []),
               apiClient.getRecentGames().catch(() => []),
            ]);
            setUpcoming(upcomingData);
            setRecent(recentData);
         } finally {
            setLoading(false);
         }
      }
      fetchGames();
   }, []);

   const handleSelectGame = async (game: NBAGame) => {
      setSelectedGame(game);
      setInsight(null);
      try {
         const oddsData = await apiClient.getGameOdds(game.id).catch(() => []);
         setOdds(oddsData);
      } catch {
         setOdds([]);
      }
   };

   const handleGenerateInsight = async () => {
      if (!selectedGame) return;
      setInsightLoading(true);
      try {
         const data = await apiClient.generateInsights(selectedGame.id);
         setInsight(data);
      } catch (err) {
         alert(err instanceof Error ? err.message : "Failed to generate insight");
      } finally {
         setInsightLoading(false);
      }
   };

   const games = tab === "upcoming" ? upcoming : recent;

   return (
      <ProtectedLayout>
         <div className="space-y-6">
            <h1 className="text-2xl font-bold text-white">Games</h1>

            {/* Tabs */}
            <div className="flex gap-2">
               <button
                  onClick={() => setTab("upcoming")}
                  className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${tab === "upcoming" ? "bg-orange-500 text-white" : "bg-gray-800 text-gray-400 hover:text-white"
                     }`}
               >
                  Upcoming (7 days)
               </button>
               <button
                  onClick={() => setTab("recent")}
                  className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${tab === "recent" ? "bg-orange-500 text-white" : "bg-gray-800 text-gray-400 hover:text-white"
                     }`}
               >
                  Recent (30 days)
               </button>
            </div>

            <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
               {/* Games list */}
               <div className="lg:col-span-2">
                  {loading ? (
                     <div className="flex justify-center py-16">
                        <Loader2 className="h-8 w-8 animate-spin text-orange-500" />
                     </div>
                  ) : games.length === 0 ? (
                     <div className="bg-gray-900 rounded-xl border border-gray-800 p-8 text-center text-gray-500">
                        No games found
                     </div>
                  ) : (
                     <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                        {games.map((game) => (
                           <button
                              key={game.id}
                              onClick={() => handleSelectGame(game)}
                              className={`text-left bg-gray-900 rounded-xl border p-4 transition-colors ${selectedGame?.id === game.id
                                 ? "border-orange-500"
                                 : "border-gray-800 hover:border-gray-700"
                                 }`}
                           >
                              <div className="text-xs text-gray-500 mb-2">
                                 {new Date(game.game_date).toLocaleDateString(undefined, {
                                    weekday: "short", month: "short", day: "numeric",
                                 })}
                                 <span className={`ml-2 px-2 py-0.5 rounded-full ${game.status === "final" ? "bg-green-500/10 text-green-400" :
                                    game.status === "scheduled" ? "bg-blue-500/10 text-blue-400" :
                                       "bg-yellow-500/10 text-yellow-400"
                                    }`}>
                                    {game.status}
                                 </span>
                              </div>
                              <div className="flex items-center justify-between">
                                 <span className="font-medium text-white text-sm truncate">{game.away_team}</span>
                                 <span className="text-gray-500 mx-2 text-xs">@</span>
                                 <span className="font-medium text-white text-sm truncate text-right">{game.home_team}</span>
                              </div>
                              {game.status === "final" && game.home_score !== null && (
                                 <div className="flex items-center justify-between mt-1 text-gray-400 text-sm">
                                    <span>{game.away_score}</span>
                                    <span>{game.home_score}</span>
                                 </div>
                              )}
                           </button>
                        ))}
                     </div>
                  )}
               </div>

               {/* Game detail sidebar */}
               <div className="space-y-4">
                  {selectedGame ? (
                     <>
                        <div className="bg-gray-900 rounded-xl border border-gray-800 p-5">
                           <h3 className="font-semibold text-white mb-3">
                              {selectedGame.away_team} @ {selectedGame.home_team}
                           </h3>
                           <p className="text-sm text-gray-400 mb-1">
                              {new Date(selectedGame.game_date).toLocaleString()}
                           </p>
                           <p className="text-sm text-gray-400">Season: {selectedGame.season}</p>

                           {odds.length > 0 && (
                              <div className="mt-4 space-y-2">
                                 <h4 className="text-sm font-medium text-gray-300">Odds</h4>
                                 {odds.map((o) => (
                                    <div key={o.id} className="text-xs text-gray-400 bg-gray-800 rounded p-2">
                                       <span className="text-gray-500">{o.bookmaker} ({o.market_type})</span>
                                       <div className="flex justify-between mt-1">
                                          {o.home_odd && <span>Home: {o.home_odd.toFixed(2)}</span>}
                                          {o.away_odd && <span>Away: {o.away_odd.toFixed(2)}</span>}
                                       </div>
                                       {o.over_under_line && (
                                          <div className="flex justify-between mt-1">
                                             <span>O/U {o.over_under_line}</span>
                                             {o.over_odd && <span>O: {o.over_odd.toFixed(2)}</span>}
                                             {o.under_odd && <span>U: {o.under_odd.toFixed(2)}</span>}
                                          </div>
                                       )}
                                    </div>
                                 ))}
                              </div>
                           )}
                        </div>

                        {/* AI Insights */}
                        <div className="bg-gray-900 rounded-xl border border-gray-800 p-5">
                           <div className="flex items-center justify-between mb-3">
                              <h4 className="font-semibold text-white flex items-center gap-2">
                                 <Sparkles className="h-4 w-4 text-orange-500" />
                                 AI Insights
                              </h4>
                           </div>
                           {insight ? (
                              <InsightDisplay text={insight.insight_text} />
                           ) : (
                              <button
                                 onClick={handleGenerateInsight}
                                 disabled={insightLoading}
                                 className="w-full py-2 bg-orange-500 hover:bg-orange-600 disabled:bg-orange-500/50 text-white text-sm font-medium rounded-lg transition-colors flex items-center justify-center gap-2"
                              >
                                 {insightLoading ? (
                                    <>
                                       <Loader2 className="h-4 w-4 animate-spin" />
                                       Analyzing...
                                    </>
                                 ) : (
                                    <>
                                       <Sparkles className="h-4 w-4" />
                                       Generate Analysis
                                    </>
                                 )}
                              </button>
                           )}
                        </div>
                     </>
                  ) : (
                     <div className="bg-gray-900 rounded-xl border border-gray-800 p-8 text-center text-gray-500">
                        Select a game to view details
                     </div>
                  )}
               </div>
            </div>
         </div>
      </ProtectedLayout>
   );
}

interface ParsedInsight {
   win_probability?: { home: number; away: number };
   recommended_bets?: Array<{
      market: string;
      pick: string;
      confidence: string;
      reasoning: string;
   }>;
   risk_level?: string;
   key_factors?: string[];
   summary?: string;
}

function tryParseInsight(text: string): ParsedInsight | null {
   if (!text) return null;
   let jsonStr = text.trim();
   // Strip markdown code fences
   if (jsonStr.startsWith("```")) {
      jsonStr = jsonStr.replace(/^```(?:json)?\s*/, "").replace(/```\s*$/, "");
   }
   // Find JSON object
   const start = jsonStr.indexOf("{");
   const end = jsonStr.lastIndexOf("}");
   if (start === -1 || end === -1) return null;
   try {
      return JSON.parse(jsonStr.substring(start, end + 1));
   } catch {
      return null;
   }
}

function getConfBadge(conf: string) {
   switch (conf?.toLowerCase()) {
      case "high":
         return { color: "text-green-400", bg: "bg-green-500/10 border-green-500/20", Icon: Zap };
      case "medium":
         return { color: "text-yellow-400", bg: "bg-yellow-500/10 border-yellow-500/20", Icon: Shield };
      default:
         return { color: "text-gray-400", bg: "bg-gray-500/10 border-gray-500/20", Icon: Shield };
   }
}

function InsightDisplay({ text }: { text: string }) {
   const [expanded, setExpanded] = useState(false);
   const parsed = tryParseInsight(text);

   if (!parsed) {
      // Fallback: plain text
      return (
         <div className="text-sm text-gray-300 whitespace-pre-wrap max-h-96 overflow-y-auto">
            {text}
         </div>
      );
   }

   const homeProb = (parsed.win_probability?.home ?? 0.5) * 100;
   const awayProb = (parsed.win_probability?.away ?? 0.5) * 100;

   return (
      <div className="space-y-4">
         {/* Win Probability */}
         {parsed.win_probability && (
            <div>
               <p className="text-xs text-gray-500 uppercase tracking-wider mb-2">Win Probability</p>
               <div className="space-y-2">
                  <div className="flex items-center gap-3">
                     <span className="text-xs text-gray-400 w-12 text-right">Home</span>
                     <div className="flex-1 h-2 rounded-full bg-gray-700 overflow-hidden">
                        <div className="h-full rounded-full bg-green-500 transition-all duration-500" style={{ width: `${homeProb}%` }} />
                     </div>
                     <span className={`text-sm font-bold tabular-nums w-12 ${homeProb >= awayProb ? "text-green-400" : "text-gray-400"}`}>
                        {homeProb.toFixed(0)}%
                     </span>
                  </div>
                  <div className="flex items-center gap-3">
                     <span className="text-xs text-gray-400 w-12 text-right">Away</span>
                     <div className="flex-1 h-2 rounded-full bg-gray-700 overflow-hidden">
                        <div className="h-full rounded-full bg-blue-500 transition-all duration-500" style={{ width: `${awayProb}%` }} />
                     </div>
                     <span className={`text-sm font-bold tabular-nums w-12 ${awayProb > homeProb ? "text-blue-400" : "text-gray-400"}`}>
                        {awayProb.toFixed(0)}%
                     </span>
                  </div>
               </div>
            </div>
         )}

         {/* Recommended Bets */}
         {parsed.recommended_bets && parsed.recommended_bets.length > 0 && (
            <div>
               <p className="text-xs text-gray-500 uppercase tracking-wider mb-2">Recommended Bets</p>
               <div className="space-y-2">
                  {parsed.recommended_bets.map((bet, i) => {
                     const badge = getConfBadge(bet.confidence);
                     return (
                        <div key={i} className="bg-gray-800/60 rounded-lg p-3 border border-gray-700/40">
                           <div className="flex items-center justify-between mb-1.5">
                              <div className="flex items-center gap-2">
                                 <TrendingUp className="h-3.5 w-3.5 text-purple-400" />
                                 <span className="text-sm font-medium text-white">{bet.pick}</span>
                              </div>
                              <div className="flex items-center gap-2">
                                 <span className="text-[10px] uppercase tracking-wider text-gray-500">{bet.market}</span>
                                 <span className={`inline-flex items-center gap-1 text-[10px] uppercase tracking-wider font-semibold px-2 py-0.5 rounded-full border ${badge.bg}`}>
                                    <badge.Icon className="h-3 w-3" />
                                    <span className={badge.color}>{bet.confidence}</span>
                                 </span>
                              </div>
                           </div>
                           {bet.reasoning && (
                              <p className="text-xs text-gray-400 leading-relaxed">{bet.reasoning}</p>
                           )}
                        </div>
                     );
                  })}
               </div>
            </div>
         )}

         {/* Risk Level */}
         {parsed.risk_level && (
            <div className="flex items-center gap-2">
               <span className="text-xs text-gray-500 uppercase tracking-wider">Risk:</span>
               <span className={`text-xs font-semibold px-2 py-0.5 rounded-full border ${parsed.risk_level.toLowerCase() === "low" ? "bg-green-500/10 border-green-500/20 text-green-400" :
                     parsed.risk_level.toLowerCase() === "medium" ? "bg-yellow-500/10 border-yellow-500/20 text-yellow-400" :
                        "bg-red-500/10 border-red-500/20 text-red-400"
                  }`}>
                  {parsed.risk_level}
               </span>
            </div>
         )}

         {/* Key Factors */}
         {parsed.key_factors && parsed.key_factors.length > 0 && (
            <div>
               <p className="text-xs text-gray-500 uppercase tracking-wider mb-2">Key Factors</p>
               <div className={`flex flex-wrap gap-1.5 ${expanded ? "" : "max-h-16 overflow-hidden"}`}>
                  {parsed.key_factors.map((f, i) => (
                     <span key={i} className="inline-flex items-center text-[11px] px-2 py-1 rounded-md bg-gray-800/80 text-gray-300 border border-gray-700/50">
                        {f}
                     </span>
                  ))}
               </div>
               {parsed.key_factors.length > 4 && (
                  <button onClick={() => setExpanded(!expanded)} className="flex items-center gap-1 text-[11px] text-gray-500 hover:text-gray-300 mt-1.5 transition-colors">
                     {expanded ? <ChevronUp className="h-3 w-3" /> : <ChevronDown className="h-3 w-3" />}
                     {expanded ? "Show less" : `Show all ${parsed.key_factors.length} factors`}
                  </button>
               )}
            </div>
         )}

         {/* Summary */}
         {parsed.summary && (
            <div className="bg-purple-500/5 border border-purple-500/10 rounded-lg p-3">
               <p className="text-xs text-gray-300 leading-relaxed">{parsed.summary}</p>
            </div>
         )}
      </div>
   );
}
