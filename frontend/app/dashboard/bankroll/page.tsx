"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import {
   Plus,
   Wallet,
   Trash2,
   ArrowUpCircle,
   ArrowDownCircle,
   AlertTriangle,
} from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { api } from "@/lib/api";
import { formatCurrency, formatDate } from "@/lib/utils";

const entryTypes = [
   { value: "DEPOSIT", label: "Deposit", icon: ArrowUpCircle, color: "text-profit" },
   { value: "WITHDRAWAL", label: "Withdrawal", icon: ArrowDownCircle, color: "text-loss" },
   { value: "BET", label: "Bet", icon: Wallet, color: "text-primary" },
   { value: "WIN", label: "Win", icon: ArrowUpCircle, color: "text-profit" },
   { value: "LOSS", label: "Loss", icon: ArrowDownCircle, color: "text-loss" },
];

export default function BankrollPage() {
   const [showForm, setShowForm] = useState(false);
   const [entryType, setEntryType] = useState("DEPOSIT");
   const [stake, setStake] = useState("");
   const [odd, setOdd] = useState("");
   const [betType, setBetType] = useState("");
   const [notes, setNotes] = useState("");
   const [result, setResult] = useState("");
   const [page, setPage] = useState(1);

   const queryClient = useQueryClient();

   const { data: entriesData, isLoading } = useQuery({
      queryKey: ["bankrollEntries", page],
      queryFn: () => api.getBankrollEntries(page),
   });

   const createMutation = useMutation({
      mutationFn: (data: any) => api.createBankrollEntry(data),
      onSuccess: () => {
         queryClient.invalidateQueries({ queryKey: ["bankrollEntries"] });
         queryClient.invalidateQueries({ queryKey: ["dashboard"] });
         resetForm();
      },
   });

   const deleteMutation = useMutation({
      mutationFn: (id: string) => api.deleteBankrollEntry(id),
      onSuccess: () => {
         queryClient.invalidateQueries({ queryKey: ["bankrollEntries"] });
         queryClient.invalidateQueries({ queryKey: ["dashboard"] });
      },
   });

   const entries = (entriesData as any)?.data || [];
   const meta = (entriesData as any)?.meta;

   const resetForm = () => {
      setShowForm(false);
      setEntryType("DEPOSIT");
      setStake("");
      setOdd("");
      setBetType("");
      setNotes("");
      setResult("");
   };

   const handleSubmit = (e: React.FormEvent) => {
      e.preventDefault();
      const data: any = {
         entry_type: entryType,
         stake: parseFloat(stake),
      };
      if (odd) data.odd = parseFloat(odd);
      if (betType) data.bet_type = betType;
      if (notes) data.notes = notes;
      if (result) data.result = parseFloat(result);

      createMutation.mutate(data);
   };

   const getEntryTypeInfo = (type: string) => {
      return entryTypes.find((t) => t.value === type) || entryTypes[0];
   };

   return (
      <div className="space-y-6">
         <div className="flex items-center justify-between">
            <div>
               <h1 className="text-2xl font-bold">Bankroll Management</h1>
               <p className="text-muted-foreground">Track your operations manually</p>
            </div>
            <Button onClick={() => setShowForm(!showForm)}>
               <Plus size={16} className="mr-1" />
               New Entry
            </Button>
         </div>

         {/* Create entry form */}
         {showForm && (
            <Card>
               <CardHeader>
                  <CardTitle className="text-base">Add Entry</CardTitle>
               </CardHeader>
               <CardContent>
                  <form onSubmit={handleSubmit} className="space-y-4">
                     <div className="grid grid-cols-2 sm:grid-cols-5 gap-2">
                        {entryTypes.map((type) => (
                           <button
                              key={type.value}
                              type="button"
                              onClick={() => setEntryType(type.value)}
                              className={`p-2 rounded-md border text-sm font-medium transition-colors ${entryType === type.value
                                    ? "border-primary bg-primary/10 text-primary"
                                    : "hover:border-muted-foreground/50"
                                 }`}
                           >
                              {type.label}
                           </button>
                        ))}
                     </div>

                     <div className="grid gap-4 sm:grid-cols-2">
                        <div className="space-y-2">
                           <Label htmlFor="stake">Amount ($)</Label>
                           <Input
                              id="stake"
                              type="number"
                              step="0.01"
                              min="0.01"
                              placeholder="100.00"
                              value={stake}
                              onChange={(e) => setStake(e.target.value)}
                              required
                           />
                        </div>

                        {(entryType === "BET" || entryType === "WIN" || entryType === "LOSS") && (
                           <>
                              <div className="space-y-2">
                                 <Label htmlFor="odd">Odd (decimal)</Label>
                                 <Input
                                    id="odd"
                                    type="number"
                                    step="0.01"
                                    min="1.01"
                                    placeholder="1.85"
                                    value={odd}
                                    onChange={(e) => setOdd(e.target.value)}
                                 />
                              </div>
                              <div className="space-y-2">
                                 <Label htmlFor="betType">Bet Type</Label>
                                 <Input
                                    id="betType"
                                    type="text"
                                    placeholder="e.g., Moneyline, Spread"
                                    value={betType}
                                    onChange={(e) => setBetType(e.target.value)}
                                 />
                              </div>
                           </>
                        )}

                        {entryType === "WIN" && (
                           <div className="space-y-2">
                              <Label htmlFor="result">Profit Amount ($)</Label>
                              <Input
                                 id="result"
                                 type="number"
                                 step="0.01"
                                 placeholder="85.00"
                                 value={result}
                                 onChange={(e) => setResult(e.target.value)}
                              />
                           </div>
                        )}
                     </div>

                     <div className="space-y-2">
                        <Label htmlFor="notes">Notes (optional)</Label>
                        <Input
                           id="notes"
                           type="text"
                           placeholder="Any notes about this entry..."
                           value={notes}
                           onChange={(e) => setNotes(e.target.value)}
                        />
                     </div>

                     <div className="flex gap-2 justify-end">
                        <Button type="button" variant="outline" onClick={resetForm}>
                           Cancel
                        </Button>
                        <Button type="submit" disabled={createMutation.isPending}>
                           {createMutation.isPending ? "Saving..." : "Save Entry"}
                        </Button>
                     </div>
                  </form>
               </CardContent>
            </Card>
         )}

         {/* Entries list */}
         <Card>
            <CardHeader>
               <CardTitle className="text-base">Transaction History</CardTitle>
            </CardHeader>
            <CardContent>
               {isLoading ? (
                  <div className="flex items-center justify-center py-8">
                     <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-primary" />
                  </div>
               ) : entries.length === 0 ? (
                  <div className="text-center py-8 text-muted-foreground">
                     <Wallet className="w-10 h-10 mx-auto mb-2 opacity-50" />
                     <p>No entries yet. Start by adding a deposit.</p>
                  </div>
               ) : (
                  <>
                     <div className="overflow-x-auto">
                        <table className="w-full text-sm">
                           <thead>
                              <tr className="border-b">
                                 <th className="text-left py-2 text-muted-foreground font-medium">Date</th>
                                 <th className="text-left py-2 text-muted-foreground font-medium">Type</th>
                                 <th className="text-right py-2 text-muted-foreground font-medium">Amount</th>
                                 <th className="text-right py-2 text-muted-foreground font-medium">Odd</th>
                                 <th className="text-right py-2 text-muted-foreground font-medium">Result</th>
                                 <th className="text-right py-2 text-muted-foreground font-medium">Balance</th>
                                 <th className="text-right py-2 text-muted-foreground font-medium"></th>
                              </tr>
                           </thead>
                           <tbody>
                              {entries.map((entry: any) => {
                                 const typeInfo = getEntryTypeInfo(entry.entry_type);
                                 return (
                                    <tr key={entry.id} className="border-b last:border-0">
                                       <td className="py-2.5">{formatDate(entry.created_at)}</td>
                                       <td className={`py-2.5 ${typeInfo.color} font-medium`}>
                                          {typeInfo.label}
                                       </td>
                                       <td className="py-2.5 text-right font-mono">
                                          {formatCurrency(entry.stake)}
                                       </td>
                                       <td className="py-2.5 text-right font-mono text-muted-foreground">
                                          {entry.odd?.toFixed(2) || "-"}
                                       </td>
                                       <td className={`py-2.5 text-right font-mono ${entry.result > 0 ? "text-profit" : entry.result === 0 ? "text-loss" : ""
                                          }`}>
                                          {entry.result != null ? formatCurrency(entry.result) : "-"}
                                       </td>
                                       <td className="py-2.5 text-right font-mono font-medium">
                                          {formatCurrency(entry.balance_after)}
                                       </td>
                                       <td className="py-2.5 text-right">
                                          <Button
                                             variant="ghost"
                                             size="icon"
                                             className="h-7 w-7 text-muted-foreground hover:text-loss"
                                             onClick={() => {
                                                if (confirm("Delete this entry?")) {
                                                   deleteMutation.mutate(entry.id);
                                                }
                                             }}
                                          >
                                             <Trash2 size={14} />
                                          </Button>
                                       </td>
                                    </tr>
                                 );
                              })}
                           </tbody>
                        </table>
                     </div>

                     {/* Pagination */}
                     {meta && meta.total_pages > 1 && (
                        <div className="flex items-center justify-between pt-4">
                           <p className="text-sm text-muted-foreground">
                              Page {meta.page || 1} of {meta.total_pages} ({meta.total} entries)
                           </p>
                           <div className="flex gap-2">
                              <Button
                                 variant="outline"
                                 size="sm"
                                 disabled={page <= 1}
                                 onClick={() => setPage((p) => p - 1)}
                              >
                                 Previous
                              </Button>
                              <Button
                                 variant="outline"
                                 size="sm"
                                 disabled={page >= meta.total_pages}
                                 onClick={() => setPage((p) => p + 1)}
                              >
                                 Next
                              </Button>
                           </div>
                        </div>
                     )}
                  </>
               )}
            </CardContent>
         </Card>

         {/* Disclaimer */}
         <div className="flex items-start gap-2 p-4 bg-muted/30 rounded-lg">
            <AlertTriangle size={16} className="text-yellow-500 mt-0.5 flex-shrink-0" />
            <p className="text-xs text-muted-foreground">
               This is a manual tracking tool. We do not accept, intermediate, or execute bets.
               All entries are self-reported. No guarantee of accuracy or returns.
            </p>
         </div>
      </div>
   );
}
