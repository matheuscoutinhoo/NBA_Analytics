"use client";

import type { AIPrediction } from "@/types";
import { TrendingUp, Shield, Zap, ChevronDown, ChevronUp, Target, ArrowUpDown, Scale } from "lucide-react";
import { useState } from "react";

/* ── JSON parsing helpers ────────────────────────────────── */

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
   if (jsonStr.startsWith("```")) {
      jsonStr = jsonStr.replace(/^```(?:json)?\s*/, "").replace(/```\s*$/, "");
   }
   const start = jsonStr.indexOf("{");
   const end = jsonStr.lastIndexOf("}");
   if (start === -1 || end === -1) return null;
   try {
      return JSON.parse(jsonStr.substring(start, end + 1));
   } catch {
      return null;
   }
}

function parseFactors(raw: string): string[] {
   if (!raw) return [];
   try {
      const parsed = JSON.parse(raw);
      if (Array.isArray(parsed)) return parsed.map(String);
   } catch {
      /* not JSON */
   }
   if (raw.includes(",")) return raw.split(",").map((s) => s.trim()).filter(Boolean);
   return raw ? [raw] : [];
}

function confBadge(conf: string) {
   switch (conf?.toLowerCase()) {
      case "high":
         return { color: "text-green-400", bg: "bg-green-500/10 border-green-500/20", Icon: Zap };
      case "medium":
         return { color: "text-yellow-400", bg: "bg-yellow-500/10 border-yellow-500/20", Icon: Shield };
      default:
         return { color: "text-gray-400", bg: "bg-gray-500/10 border-gray-500/20", Icon: Shield };
   }
}

function riskBadge(level: string) {
   switch (level?.toLowerCase()) {
      case "low":
         return "bg-green-500/10 border-green-500/20 text-green-400";
      case "medium":
         return "bg-yellow-500/10 border-yellow-500/20 text-yellow-400";
      default:
         return "bg-red-500/10 border-red-500/20 text-red-400";
   }
}

function marketIcon(market: string) {
   switch (market?.toLowerCase()) {
      case "moneyline":
         return Target;
      case "over_under":
         return ArrowUpDown;
      case "handicap":
      case "spread":
         return Scale;
      default:
         return TrendingUp;
   }
}

function marketLabel(market: string) {
   switch (market?.toLowerCase()) {
      case "moneyline":
         return "MONEYLINE";
      case "over_under":
         return "OVER/UNDER";
      case "handicap":
      case "spread":
         return "HANDICAP";
      default:
         return market?.toUpperCase() ?? "";
   }
}

/* ── Component ───────────────────────────────────────────── */

export default function PredictionCard({ pred }: { pred: AIPrediction }) {
   const [expanded, setExpanded] = useState(false);

   const parsed = tryParseInsight(pred.summary);
   const factors = parsed?.key_factors ?? parseFactors(pred.key_factors);

   const homeProb = parsed?.win_probability
      ? parsed.win_probability.home * 100
      : pred.home_win_prob * 100;
   const awayProb = parsed?.win_probability
      ? parsed.win_probability.away * 100
      : pred.away_win_prob * 100;

   const bets = parsed?.recommended_bets ?? [];
   const riskLevel = parsed?.risk_level ?? "";
   const summaryText = parsed?.summary ?? "";

   const gameDate = new Date(pred.game_date);
   const isToday = new Date().toDateString() === gameDate.toDateString();

   return (
      <div className="bg-gray-800/60 rounded-xl border border-gray-700/40 overflow-hidden hover:border-gray-600/60 transition-colors">
         {/* Header — Team matchup */}
         <div className="px-4 pt-4 pb-3">
            <div className="flex items-center justify-between mb-1">
               <h3 className="text-sm font-bold text-white truncate">
                  {pred.home_team} <span className="text-gray-500 font-normal">vs</span> {pred.away_team}
               </h3>
               {isToday ? (
                  <span className="text-[10px] uppercase tracking-wider font-semibold px-2 py-0.5 rounded-full bg-orange-500/10 text-orange-400 border border-orange-500/20 shrink-0 ml-2">
                     Today
                  </span>
               ) : (
                  <span className="text-[10px] uppercase tracking-wider text-gray-500 shrink-0 ml-2">
                     {gameDate.toLocaleDateString(undefined, { weekday: "short", month: "short", day: "numeric" })}
                  </span>
               )}
            </div>
         </div>

         <div className="px-4 pb-4 space-y-4">
            {/* Win Probability */}
            <div>
               <p className="text-[10px] text-gray-500 uppercase tracking-wider mb-2">Win Probability</p>
               <div className="space-y-2">
                  <div className="flex items-center gap-3">
                     <span className="text-xs text-gray-400 w-12 text-right">Home</span>
                     <div className="flex-1 h-2.5 rounded-full bg-gray-700 overflow-hidden">
                        <div
                           className="h-full rounded-full bg-linear-to-r from-green-500 to-emerald-400 transition-all duration-500"
                           style={{ width: `${homeProb}%` }}
                        />
                     </div>
                     <span className={`text-sm font-bold tabular-nums w-12 ${homeProb >= awayProb ? "text-green-400" : "text-gray-400"}`}>
                        {homeProb.toFixed(0)}%
                     </span>
                  </div>
                  <div className="flex items-center gap-3">
                     <span className="text-xs text-gray-400 w-12 text-right">Away</span>
                     <div className="flex-1 h-2.5 rounded-full bg-gray-700 overflow-hidden">
                        <div
                           className="h-full rounded-full bg-linear-to-r from-blue-500 to-cyan-400 transition-all duration-500"
                           style={{ width: `${awayProb}%` }}
                        />
                     </div>
                     <span className={`text-sm font-bold tabular-nums w-12 ${awayProb > homeProb ? "text-blue-400" : "text-gray-400"}`}>
                        {awayProb.toFixed(0)}%
                     </span>
                  </div>
               </div>
            </div>

            {/* Recommended Bets — show ALL */}
            {bets.length > 0 && (
               <div>
                  <p className="text-[10px] text-gray-500 uppercase tracking-wider mb-2">Recommended Bets</p>
                  <div className="space-y-2">
                     {bets.map((bet, i) => {
                        const badge = confBadge(bet.confidence);
                        const MktIcon = marketIcon(bet.market);
                        return (
                           <div key={i} className="bg-gray-900/60 rounded-lg p-3 border border-gray-700/30">
                              <div className="flex items-center justify-between mb-1.5">
                                 <div className="flex items-center gap-2">
                                    <MktIcon className="h-3.5 w-3.5 text-purple-400" />
                                    <span className="text-sm font-medium text-white">{bet.pick}</span>
                                 </div>
                                 <div className="flex items-center gap-2">
                                    <span className="text-[10px] uppercase tracking-wider text-gray-500">{marketLabel(bet.market)}</span>
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

            {/* Fallback: no parsed bets, show AI pick from DB fields */}
            {bets.length === 0 && pred.recommended_pick && (
               <div>
                  <p className="text-[10px] text-gray-500 uppercase tracking-wider mb-2">Recommended Bet</p>
                  <div className="bg-gray-900/60 rounded-lg p-3 border border-gray-700/30">
                     <div className="flex items-center justify-between">
                        <div className="flex items-center gap-2">
                           <TrendingUp className="h-3.5 w-3.5 text-purple-400" />
                           <span className="text-sm font-medium text-white">{pred.recommended_pick}</span>
                        </div>
                        <div className="flex items-center gap-2">
                           <span className="text-[10px] uppercase tracking-wider text-gray-500">MONEYLINE</span>
                           {(() => {
                              const badge = confBadge(pred.confidence);
                              return (
                                 <span className={`inline-flex items-center gap-1 text-[10px] uppercase tracking-wider font-semibold px-2 py-0.5 rounded-full border ${badge.bg}`}>
                                    <badge.Icon className="h-3 w-3" />
                                    <span className={badge.color}>{pred.confidence}</span>
                                 </span>
                              );
                           })()}
                        </div>
                     </div>
                  </div>
               </div>
            )}

            {/* Risk Level */}
            {riskLevel && (
               <div className="flex items-center gap-2">
                  <span className="text-[10px] text-gray-500 uppercase tracking-wider">Risk:</span>
                  <span className={`text-xs font-semibold px-2.5 py-0.5 rounded-full border ${riskBadge(riskLevel)}`}>
                     {riskLevel}
                  </span>
               </div>
            )}

            {/* Key Factors */}
            {factors.length > 0 && (
               <div>
                  <p className="text-[10px] text-gray-500 uppercase tracking-wider mb-2">Key Factors</p>
                  <div className={`flex flex-wrap gap-1.5 ${expanded ? "" : "max-h-16 overflow-hidden"}`}>
                     {factors.map((f, i) => (
                        <span
                           key={i}
                           className="inline-flex items-center text-[11px] px-2.5 py-1 rounded-md bg-gray-900/60 text-gray-300 border border-gray-700/40"
                        >
                           {f}
                        </span>
                     ))}
                  </div>
               </div>
            )}

            {/* Summary */}
            {summaryText && expanded && (
               <div className="bg-purple-500/5 border border-purple-500/10 rounded-lg p-3">
                  <p className="text-xs text-gray-300 leading-relaxed">{summaryText}</p>
               </div>
            )}

            {/* Expand/Collapse */}
            {(factors.length > 3 || summaryText) && (
               <button
                  onClick={() => setExpanded(!expanded)}
                  className="flex items-center gap-1 text-[11px] text-gray-500 hover:text-gray-300 transition-colors"
               >
                  {expanded ? <ChevronUp className="h-3 w-3" /> : <ChevronDown className="h-3 w-3" />}
                  {expanded ? "Show less" : "Show more"}
               </button>
            )}
         </div>
      </div>
   );
}
