"use client";

import { useEffect, useState } from "react";
import ProtectedLayout from "@/components/ProtectedLayout";
import apiClient from "@/lib/api-client";
import type { UserBet, NBAGame } from "@/types";
import { Plus, Trash2, Check, X, Loader2 } from "lucide-react";

export default function BetsPage() {
   const [bets, setBets] = useState<UserBet[]>([]);
   const [games, setGames] = useState<NBAGame[]>([]);
   const [loading, setLoading] = useState(true);
   const [showForm, setShowForm] = useState(false);
   const [formData, setFormData] = useState({
      game_id: 0,
      bet_type: "moneyline",
      selection: "",
      odd_value: "",
      stake: "",
   });
   const [formError, setFormError] = useState("");

   useEffect(() => {
      fetchBets();
      apiClient.getUpcomingGames().then(setGames).catch(() => setGames([]));
   }, []);

   const fetchBets = async () => {
      try {
         const data = await apiClient.getBets();
         setBets(data);
      } catch {
         setBets([]);
      } finally {
         setLoading(false);
      }
   };

   const handleCreateBet = async (e: React.FormEvent) => {
      e.preventDefault();
      setFormError("");
      try {
         await apiClient.createBet({
            game_id: formData.game_id,
            bet_type: formData.bet_type,
            selection: formData.selection,
            odd_value: parseFloat(formData.odd_value),
            stake: parseFloat(formData.stake),
         });
         setShowForm(false);
         setFormData({ game_id: 0, bet_type: "moneyline", selection: "", odd_value: "", stake: "" });
         fetchBets();
      } catch (err) {
         setFormError(err instanceof Error ? err.message : "Failed to create bet");
      }
   };

   const handleUpdateResult = async (betId: number, result: string) => {
      try {
         await apiClient.updateBetResult(betId, result);
         fetchBets();
      } catch (err) {
         alert(err instanceof Error ? err.message : "Failed to update");
      }
   };

   const handleDeleteBet = async (betId: number) => {
      if (!confirm("Delete this bet?")) return;
      try {
         await apiClient.deleteBet(betId);
         fetchBets();
      } catch (err) {
         alert(err instanceof Error ? err.message : "Failed to delete");
      }
   };

   const pendingBets = bets.filter((b) => b.result === "pending");
   const settledBets = bets.filter((b) => b.result !== "pending");

   return (
      <ProtectedLayout>
         <div className="space-y-6">
            <div className="flex items-center justify-between">
               <h1 className="text-2xl font-bold text-white">Bet Tracking</h1>
               <button
                  onClick={() => setShowForm(!showForm)}
                  className="flex items-center gap-2 px-4 py-2 bg-orange-500 hover:bg-orange-600 text-white text-sm font-medium rounded-lg transition-colors"
               >
                  <Plus className="h-4 w-4" />
                  New Bet
               </button>
            </div>

            {/* New bet form */}
            {showForm && (
               <form onSubmit={handleCreateBet} className="bg-gray-900 rounded-xl border border-gray-800 p-6 space-y-4">
                  {formError && (
                     <div className="bg-red-500/10 border border-red-500/20 text-red-400 px-4 py-2 rounded-lg text-sm">{formError}</div>
                  )}
                  <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
                     <div>
                        <label className="block text-sm text-gray-400 mb-1">Game</label>
                        <select
                           value={formData.game_id}
                           onChange={(e) => setFormData({ ...formData, game_id: parseInt(e.target.value) })}
                           required
                           className="w-full px-3 py-2 bg-gray-800 border border-gray-700 rounded-lg text-white"
                        >
                           <option value={0}>Select game</option>
                           {games.map((g) => (
                              <option key={g.id} value={g.id}>
                                 {g.away_team} @ {g.home_team}
                              </option>
                           ))}
                        </select>
                     </div>
                     <div>
                        <label className="block text-sm text-gray-400 mb-1">Type</label>
                        <select
                           value={formData.bet_type}
                           onChange={(e) => setFormData({ ...formData, bet_type: e.target.value })}
                           className="w-full px-3 py-2 bg-gray-800 border border-gray-700 rounded-lg text-white"
                        >
                           <option value="moneyline">Moneyline</option>
                           <option value="spread">Spread</option>
                           <option value="over_under">Over/Under</option>
                           <option value="prop">Prop</option>
                        </select>
                     </div>
                     <div>
                        <label className="block text-sm text-gray-400 mb-1">Selection</label>
                        <input
                           type="text"
                           value={formData.selection}
                           onChange={(e) => setFormData({ ...formData, selection: e.target.value })}
                           required
                           placeholder="e.g., Lakers ML"
                           className="w-full px-3 py-2 bg-gray-800 border border-gray-700 rounded-lg text-white placeholder-gray-500"
                        />
                     </div>
                     <div>
                        <label className="block text-sm text-gray-400 mb-1">Odds</label>
                        <input
                           type="number"
                           step="0.01"
                           min="1.01"
                           value={formData.odd_value}
                           onChange={(e) => setFormData({ ...formData, odd_value: e.target.value })}
                           required
                           placeholder="1.85"
                           className="w-full px-3 py-2 bg-gray-800 border border-gray-700 rounded-lg text-white placeholder-gray-500"
                        />
                     </div>
                     <div>
                        <label className="block text-sm text-gray-400 mb-1">Stake ($)</label>
                        <input
                           type="number"
                           step="0.01"
                           min="0.01"
                           value={formData.stake}
                           onChange={(e) => setFormData({ ...formData, stake: e.target.value })}
                           required
                           placeholder="100.00"
                           className="w-full px-3 py-2 bg-gray-800 border border-gray-700 rounded-lg text-white placeholder-gray-500"
                        />
                     </div>
                     <div className="flex items-end">
                        <button
                           type="submit"
                           className="w-full py-2 bg-orange-500 hover:bg-orange-600 text-white font-medium rounded-lg transition-colors"
                        >
                           Place Bet
                        </button>
                     </div>
                  </div>
               </form>
            )}

            {loading ? (
               <div className="flex justify-center py-16">
                  <Loader2 className="h-8 w-8 animate-spin text-orange-500" />
               </div>
            ) : (
               <>
                  {/* Pending bets */}
                  <div className="bg-gray-900 rounded-xl border border-gray-800 p-6">
                     <h2 className="text-lg font-semibold text-white mb-4">Pending ({pendingBets.length})</h2>
                     {pendingBets.length === 0 ? (
                        <p className="text-gray-500 text-center py-4">No pending bets</p>
                     ) : (
                        <div className="overflow-x-auto">
                           <table className="w-full text-sm">
                              <thead>
                                 <tr className="text-gray-400 border-b border-gray-800">
                                    <th className="text-left py-2 px-3">Date</th>
                                    <th className="text-left py-2 px-3">Type</th>
                                    <th className="text-left py-2 px-3">Selection</th>
                                    <th className="text-right py-2 px-3">Odds</th>
                                    <th className="text-right py-2 px-3">Stake</th>
                                    <th className="text-right py-2 px-3">Pot. Return</th>
                                    <th className="text-center py-2 px-3">Actions</th>
                                 </tr>
                              </thead>
                              <tbody>
                                 {pendingBets.map((bet) => (
                                    <tr key={bet.id} className="border-b border-gray-800/50">
                                       <td className="py-2 px-3 text-gray-400">{new Date(bet.placed_at).toLocaleDateString()}</td>
                                       <td className="py-2 px-3 text-white">{bet.bet_type}</td>
                                       <td className="py-2 px-3 text-white">{bet.selection}</td>
                                       <td className="py-2 px-3 text-right text-white">{bet.odd_value.toFixed(2)}</td>
                                       <td className="py-2 px-3 text-right text-white">${bet.stake.toFixed(2)}</td>
                                       <td className="py-2 px-3 text-right text-orange-400">${bet.potential_return.toFixed(2)}</td>
                                       <td className="py-2 px-3">
                                          <div className="flex items-center justify-center gap-1">
                                             <button onClick={() => handleUpdateResult(bet.id, "won")} className="p-1 text-green-400 hover:bg-green-500/10 rounded" title="Won">
                                                <Check className="h-4 w-4" />
                                             </button>
                                             <button onClick={() => handleUpdateResult(bet.id, "lost")} className="p-1 text-red-400 hover:bg-red-500/10 rounded" title="Lost">
                                                <X className="h-4 w-4" />
                                             </button>
                                             <button onClick={() => handleDeleteBet(bet.id)} className="p-1 text-gray-400 hover:bg-gray-700 rounded" title="Delete">
                                                <Trash2 className="h-4 w-4" />
                                             </button>
                                          </div>
                                       </td>
                                    </tr>
                                 ))}
                              </tbody>
                           </table>
                        </div>
                     )}
                  </div>

                  {/* Settled bets */}
                  <div className="bg-gray-900 rounded-xl border border-gray-800 p-6">
                     <h2 className="text-lg font-semibold text-white mb-4">Settled ({settledBets.length})</h2>
                     {settledBets.length === 0 ? (
                        <p className="text-gray-500 text-center py-4">No settled bets</p>
                     ) : (
                        <div className="overflow-x-auto">
                           <table className="w-full text-sm">
                              <thead>
                                 <tr className="text-gray-400 border-b border-gray-800">
                                    <th className="text-left py-2 px-3">Date</th>
                                    <th className="text-left py-2 px-3">Type</th>
                                    <th className="text-left py-2 px-3">Selection</th>
                                    <th className="text-right py-2 px-3">Odds</th>
                                    <th className="text-right py-2 px-3">Stake</th>
                                    <th className="text-center py-2 px-3">Result</th>
                                    <th className="text-right py-2 px-3">P&L</th>
                                 </tr>
                              </thead>
                              <tbody>
                                 {settledBets.map((bet) => {
                                    const pl = bet.result === "won" ? bet.potential_return - bet.stake : bet.result === "void" ? 0 : -bet.stake;
                                    return (
                                       <tr key={bet.id} className="border-b border-gray-800/50">
                                          <td className="py-2 px-3 text-gray-400">{new Date(bet.placed_at).toLocaleDateString()}</td>
                                          <td className="py-2 px-3 text-white">{bet.bet_type}</td>
                                          <td className="py-2 px-3 text-white">{bet.selection}</td>
                                          <td className="py-2 px-3 text-right text-white">{bet.odd_value.toFixed(2)}</td>
                                          <td className="py-2 px-3 text-right text-white">${bet.stake.toFixed(2)}</td>
                                          <td className="py-2 px-3 text-center">
                                             <span className={`px-2 py-0.5 rounded-full text-xs ${bet.result === "won" ? "bg-green-500/10 text-green-400" :
                                                   bet.result === "lost" ? "bg-red-500/10 text-red-400" :
                                                      "bg-gray-500/10 text-gray-400"
                                                }`}>{bet.result}</span>
                                          </td>
                                          <td className={`py-2 px-3 text-right ${pl >= 0 ? "text-green-400" : "text-red-400"}`}>
                                             {pl >= 0 ? "+" : ""}${pl.toFixed(2)}
                                          </td>
                                       </tr>
                                    );
                                 })}
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
