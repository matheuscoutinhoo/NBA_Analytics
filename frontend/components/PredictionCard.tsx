"use client";

import type { AIPrediction } from "@/types";
import { Brain, TrendingUp, Shield, Zap, ChevronDown, ChevronUp } from "lucide-react";
import { useState } from "react";

/** Try to parse a string that might be JSON or a JSON-embedded summary */
function parseSummary(raw: string): string {
   if (!raw) return "";
   // If it looks like JSON, try to extract the summary field
   const trimmed = raw.trim();
   if (trimmed.startsWith("{") || trimmed.startsWith("```")) {
      try {
         let jsonStr = trimmed;
         // Strip markdown code fences
         if (jsonStr.startsWith("```")) {
            jsonStr = jsonStr.replace(/^```(?:json)?\s*/, "").replace(/```\s*$/, "");
         }
         const parsed = JSON.parse(jsonStr);
         if (parsed.summary) return parsed.summary;
         // Fallback: build a summary from what we find
         const parts: string[] = [];
         if (parsed.recommended_bets?.length) {
            parts.push(parsed.recommended_bets[0].reasoning || "");
         }
         if (parsed.risk_level) {
            parts.push(`Risk: ${parsed.risk_level}`);
         }
         return parts.filter(Boolean).join(". ") || "AI analysis complete.";
      } catch {
         // Not valid JSON, return as-is but clean up
      }
   }
   return raw;
}

/** Try to parse key_factors from JSON array string */
function parseFactors(raw: string): string[] {
   if (!raw) return [];
   try {
      const parsed = JSON.parse(raw);
      if (Array.isArray(parsed)) return parsed.map(String);
   } catch {
      // Not JSON array
   }
   // Maybe it's a comma-separated string
   if (raw.includes(",")) return raw.split(",").map((s) => s.trim()).filter(Boolean);
   return raw ? [raw] : [];
}

function getConfidenceConfig(confidence: string) {
   switch (confidence?.toLowerCase()) {
      case "high":
         return { color: "text-green-400", bg: "bg-green-500/10 border-green-500/20", icon: Zap, label: "High" };
      case "medium":
         return { color: "text-yellow-400", bg: "bg-yellow-500/10 border-yellow-500/20", icon: Shield, label: "Medium" };
      default:
         return { color: "text-gray-400", bg: "bg-gray-500/10 border-gray-500/20", icon: Shield, label: confidence || "Low" };
   }
}

function getProbColor(pct: number): string {
   if (pct >= 70) return "text-green-400";
   if (pct >= 60) return "text-emerald-400";
   if (pct >= 55) return "text-yellow-400";
   return "text-gray-400";
}

export default function PredictionCard({ pred }: { pred: AIPrediction }) {
   const [expanded, setExpanded] = useState(false);

   const homeProb = pred.home_win_prob * 100;
   const awayProb = pred.away_win_prob * 100;
   const winProb = Math.max(homeProb, awayProb);
   const favored = homeProb >= awayProb ? pred.home_team : pred.away_team;
   const underdog = homeProb >= awayProb ? pred.away_team : pred.home_team;
   const favoredProb = Math.max(homeProb, awayProb);
   const underdogProb = Math.min(homeProb, awayProb);
   const favoredOdd = homeProb >= awayProb ? pred.home_odd : pred.away_odd;
   const underdogOdd = homeProb >= awayProb ? pred.away_odd : pred.home_odd;

   const summary = parseSummary(pred.summary);
   const factors = parseFactors(pred.key_factors);
   const conf = getConfidenceConfig(pred.confidence);
   const ConfIcon = conf.icon;

   const gameDate = new Date(pred.game_date);
   const isToday = new Date().toDateString() === gameDate.toDateString();

   return (
      <div className="bg-gray-800/60 rounded-xl border border-gray-700/40 overflow-hidden hover:border-gray-600/60 transition-colors">
         {/* Header */}
         <div className="px-4 pt-4 pb-3">
            <div className="flex items-center justify-between mb-3">
               <div className="flex items-center gap-2">
                  {isToday ? (
                     <span className="text-[10px] uppercase tracking-wider font-semibold px-2 py-0.5 rounded-full bg-orange-500/10 text-orange-400 border border-orange-500/20">
                        Today
                     </span>
                  ) : (
                     <span className="text-[10px] uppercase tracking-wider text-gray-500">
                        {gameDate.toLocaleDateString(undefined, { weekday: "short", month: "short", day: "numeric" })}
                     </span>
                  )}
                  <span className={`inline-flex items-center gap-1 text-[10px] uppercase tracking-wider font-semibold px-2 py-0.5 rounded-full border ${conf.bg}`}>
                     <ConfIcon className="h-3 w-3" />
                     <span className={conf.color}>{conf.label}</span>
                  </span>
               </div>
               {pred.bookmaker && (
                  <span className="text-[10px] text-gray-600 uppercase tracking-wider">{pred.bookmaker}</span>
               )}
            </div>

            {/* Teams + Probabilities */}
            <div className="flex items-center gap-3">
               {/* Home team */}
               <div className="flex-1 min-w-0">
                  <div className="flex items-baseline justify-between mb-1">
                     <span className={`text-sm font-medium truncate ${homeProb >= awayProb ? "text-white" : "text-gray-400"}`}>
                        {pred.home_team}
                     </span>
                     <span className={`text-lg font-bold tabular-nums ml-2 ${getProbColor(homeProb)}`}>
                        {homeProb.toFixed(0)}%
                     </span>
                  </div>
                  <div className="h-1.5 rounded-full bg-gray-700 overflow-hidden">
                     <div
                        className="h-full rounded-full bg-linear-to-r from-green-500 to-emerald-400 transition-all duration-500"
                        style={{ width: `${homeProb}%` }}
                     />
                  </div>
               </div>

               <span className="text-xs text-gray-600 font-medium shrink-0">vs</span>

               {/* Away team */}
               <div className="flex-1 min-w-0">
                  <div className="flex items-baseline justify-between mb-1">
                     <span className={`text-sm font-medium truncate ${awayProb > homeProb ? "text-white" : "text-gray-400"}`}>
                        {pred.away_team}
                     </span>
                     <span className={`text-lg font-bold tabular-nums ml-2 ${getProbColor(awayProb)}`}>
                        {awayProb.toFixed(0)}%
                     </span>
                  </div>
                  <div className="h-1.5 rounded-full bg-gray-700 overflow-hidden">
                     <div
                        className="h-full rounded-full bg-linear-to-r from-blue-500 to-cyan-400 transition-all duration-500"
                        style={{ width: `${awayProb}%` }}
                     />
                  </div>
               </div>
            </div>
         </div>

         {/* Pick strip */}
         <div className="px-4 py-2.5 bg-purple-500/5 border-t border-b border-purple-500/10 flex items-center justify-between">
            <div className="flex items-center gap-2">
               <TrendingUp className="h-3.5 w-3.5 text-purple-400" />
               <span className="text-xs text-gray-400">AI Pick:</span>
               <span className="text-xs font-semibold text-purple-300">{favored}</span>
            </div>
            <div className="flex items-center gap-3">
               {favoredOdd != null && (
                  <span className="text-xs font-mono text-gray-400">
                     Odds <span className="text-white">{favoredOdd.toFixed(2)}</span>
                  </span>
               )}
               {pred.over_under_line != null && (
                  <span className="text-xs font-mono text-gray-400">
                     O/U <span className="text-yellow-400">{pred.over_under_line.toFixed(1)}</span>
                  </span>
               )}
            </div>
         </div>

         {/* Summary + Factors */}
         <div className="px-4 py-3">
            {summary && (
               <p className={`text-xs text-gray-400 leading-relaxed ${expanded ? "" : "line-clamp-2"}`}>
                  {summary}
               </p>
            )}

            {factors.length > 0 && (
               <div className={`flex flex-wrap gap-1.5 mt-2 ${expanded ? "" : "max-h-7 overflow-hidden"}`}>
                  {factors.map((f, i) => (
                     <span
                        key={i}
                        className="inline-flex items-center text-[11px] px-2 py-0.5 rounded-md bg-gray-700/40 text-gray-300 border border-gray-700/50"
                     >
                        {f}
                     </span>
                  ))}
               </div>
            )}

            {(summary.length > 120 || factors.length > 3) && (
               <button
                  onClick={() => setExpanded(!expanded)}
                  className="flex items-center gap-1 text-[11px] text-gray-500 hover:text-gray-300 mt-2 transition-colors"
               >
                  {expanded ? <ChevronUp className="h-3 w-3" /> : <ChevronDown className="h-3 w-3" />}
                  {expanded ? "Show less" : "Show more"}
               </button>
            )}
         </div>
      </div>
   );
}
