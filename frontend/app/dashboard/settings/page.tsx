"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import {
   Settings,
   Shield,
   Download,
   Trash2,
   AlertTriangle,
   Bell,
} from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { api } from "@/lib/api";
import { useAuthStore } from "@/store";

export default function SettingsPage() {
   const { user, logout } = useAuthStore();
   const queryClient = useQueryClient();
   const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);
   const [deleteConfirmText, setDeleteConfirmText] = useState("");

   const { data: prefsData } = useQuery({
      queryKey: ["preferences"],
      queryFn: () => api.getPreferences(),
   });

   const prefs = (prefsData as any)?.data;

   const updatePrefsMutation = useMutation({
      mutationFn: (data: Record<string, any>) => api.updatePreferences(data),
      onSuccess: () => {
         queryClient.invalidateQueries({ queryKey: ["preferences"] });
      },
   });

   const handleExport = async () => {
      try {
         const data = await api.exportData();
         const blob = new Blob([JSON.stringify((data as any).data, null, 2)], {
            type: "application/json",
         });
         const url = window.URL.createObjectURL(blob);
         const a = document.createElement("a");
         a.href = url;
         a.download = `better-export-${new Date().toISOString().split("T")[0]}.json`;
         document.body.appendChild(a);
         a.click();
         window.URL.revokeObjectURL(url);
         document.body.removeChild(a);
      } catch (err) {
         alert("Failed to export data");
      }
   };

   const handleDeleteAccount = async () => {
      if (deleteConfirmText !== "DELETE") return;
      try {
         await api.deleteAccount();
         logout();
         window.location.href = "/";
      } catch (err) {
         alert("Failed to delete account");
      }
   };

   return (
      <div className="space-y-6 max-w-2xl">
         <div>
            <h1 className="text-2xl font-bold">Settings</h1>
            <p className="text-muted-foreground">Manage your account and preferences</p>
         </div>

         {/* Account info */}
         <Card>
            <CardHeader>
               <CardTitle className="text-base flex items-center gap-2">
                  <Settings size={16} />
                  Account
               </CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
               <div className="grid grid-cols-2 gap-4 text-sm">
                  <div>
                     <p className="text-muted-foreground">Username</p>
                     <p className="font-medium">{user?.username}</p>
                  </div>
                  <div>
                     <p className="text-muted-foreground">Email</p>
                     <p className="font-medium">{user?.email}</p>
                  </div>
                  <div>
                     <p className="text-muted-foreground">Role</p>
                     <p className="font-medium">{user?.role}</p>
                  </div>
                  <div>
                     <p className="text-muted-foreground">XP</p>
                     <p className="font-medium">{user?.xp_points} pts</p>
                  </div>
               </div>
            </CardContent>
         </Card>

         {/* Personal limits */}
         <Card>
            <CardHeader>
               <CardTitle className="text-base flex items-center gap-2">
                  <Shield size={16} />
                  Responsible Gaming Limits
               </CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
               <p className="text-sm text-muted-foreground">
                  Set personal soft limits to help manage your activity. These are for your awareness only.
               </p>
               <div className="grid gap-4 sm:grid-cols-3">
                  <div className="space-y-2">
                     <Label htmlFor="daily">Daily Limit ($)</Label>
                     <Input
                        id="daily"
                        type="number"
                        step="1"
                        min="0"
                        placeholder="0.00"
                        defaultValue={prefs?.daily_limit || ""}
                        onBlur={(e) => {
                           const val = e.target.value ? parseFloat(e.target.value) : null;
                           updatePrefsMutation.mutate({ daily_limit: val });
                        }}
                     />
                  </div>
                  <div className="space-y-2">
                     <Label htmlFor="weekly">Weekly Limit ($)</Label>
                     <Input
                        id="weekly"
                        type="number"
                        step="1"
                        min="0"
                        placeholder="0.00"
                        defaultValue={prefs?.weekly_limit || ""}
                        onBlur={(e) => {
                           const val = e.target.value ? parseFloat(e.target.value) : null;
                           updatePrefsMutation.mutate({ weekly_limit: val });
                        }}
                     />
                  </div>
                  <div className="space-y-2">
                     <Label htmlFor="monthly">Monthly Limit ($)</Label>
                     <Input
                        id="monthly"
                        type="number"
                        step="1"
                        min="0"
                        placeholder="0.00"
                        defaultValue={prefs?.monthly_limit || ""}
                        onBlur={(e) => {
                           const val = e.target.value ? parseFloat(e.target.value) : null;
                           updatePrefsMutation.mutate({ monthly_limit: val });
                        }}
                     />
                  </div>
               </div>
               <div className="space-y-2">
                  <Label htmlFor="alert">Alert Threshold (%)</Label>
                  <Input
                     id="alert"
                     type="number"
                     step="5"
                     min="0"
                     max="100"
                     placeholder="80"
                     defaultValue={prefs?.alert_threshold || 80}
                     onBlur={(e) => {
                        updatePrefsMutation.mutate({ alert_threshold: parseFloat(e.target.value) });
                     }}
                  />
                  <p className="text-xs text-muted-foreground">Alert when you reach this % of your limit</p>
               </div>
            </CardContent>
         </Card>

         {/* Notifications */}
         <Card>
            <CardHeader>
               <CardTitle className="text-base flex items-center gap-2">
                  <Bell size={16} />
                  Notifications
               </CardTitle>
            </CardHeader>
            <CardContent>
               <div className="flex items-center justify-between">
                  <div>
                     <p className="text-sm font-medium">Enable Notifications</p>
                     <p className="text-xs text-muted-foreground">Get alerts about limits and new analyses</p>
                  </div>
                  <button
                     onClick={() => {
                        updatePrefsMutation.mutate({
                           notifications_enabled: !prefs?.notifications_enabled,
                        });
                     }}
                     className={`w-11 h-6 rounded-full transition-colors ${prefs?.notifications_enabled ? "bg-primary" : "bg-muted"
                        }`}
                  >
                     <div
                        className={`w-5 h-5 rounded-full bg-white transition-transform ${prefs?.notifications_enabled ? "translate-x-5" : "translate-x-0.5"
                           }`}
                     />
                  </button>
               </div>
            </CardContent>
         </Card>

         {/* Data management (LGPD) */}
         <Card>
            <CardHeader>
               <CardTitle className="text-base flex items-center gap-2">
                  <Download size={16} />
                  Data Management (LGPD)
               </CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
               <div>
                  <p className="text-sm text-muted-foreground mb-3">
                     You have the right to access, export, and delete all your personal data.
                  </p>
                  <Button variant="outline" onClick={handleExport}>
                     <Download size={16} className="mr-2" />
                     Export All My Data
                  </Button>
               </div>
            </CardContent>
         </Card>

         {/* Delete account */}
         <Card className="border-destructive/50">
            <CardHeader>
               <CardTitle className="text-base text-destructive flex items-center gap-2">
                  <Trash2 size={16} />
                  Delete Account
               </CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
               <p className="text-sm text-muted-foreground">
                  Permanently delete your account and all associated data. This action cannot be undone.
               </p>
               {showDeleteConfirm ? (
                  <div className="space-y-3">
                     <div className="flex items-start gap-2 p-3 bg-destructive/10 rounded-md">
                        <AlertTriangle size={16} className="text-destructive mt-0.5 flex-shrink-0" />
                        <p className="text-sm text-destructive">
                           This will permanently delete all your data including bankroll entries,
                           preferences, and account information. Type <strong>DELETE</strong> to confirm.
                        </p>
                     </div>
                     <Input
                        type="text"
                        placeholder="Type DELETE to confirm"
                        value={deleteConfirmText}
                        onChange={(e) => setDeleteConfirmText(e.target.value)}
                     />
                     <div className="flex gap-2">
                        <Button variant="outline" onClick={() => {
                           setShowDeleteConfirm(false);
                           setDeleteConfirmText("");
                        }}>
                           Cancel
                        </Button>
                        <Button
                           variant="destructive"
                           disabled={deleteConfirmText !== "DELETE"}
                           onClick={handleDeleteAccount}
                        >
                           Permanently Delete
                        </Button>
                     </div>
                  </div>
               ) : (
                  <Button variant="destructive" onClick={() => setShowDeleteConfirm(true)}>
                     Delete My Account
                  </Button>
               )}
            </CardContent>
         </Card>
      </div>
   );
}
