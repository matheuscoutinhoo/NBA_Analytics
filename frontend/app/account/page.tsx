"use client";

import { useState } from "react";
import ProtectedLayout from "@/components/ProtectedLayout";
import { useAuth } from "@/hooks/useAuth";
import apiClient from "@/lib/api-client";
import { UserCircle, Mail, Lock, Trash2, Loader2 } from "lucide-react";

export default function AccountPage() {
   const { user, deleteAccount } = useAuth();
   const [email, setEmail] = useState(user?.email || "");
   const [currentPassword, setCurrentPassword] = useState("");
   const [newPassword, setNewPassword] = useState("");
   const [confirmPassword, setConfirmPassword] = useState("");
   const [profileMsg, setProfileMsg] = useState("");
   const [passwordMsg, setPasswordMsg] = useState("");
   const [profileError, setProfileError] = useState("");
   const [passwordError, setPasswordError] = useState("");
   const [profileLoading, setProfileLoading] = useState(false);
   const [passwordLoading, setPasswordLoading] = useState(false);
   const [deleteLoading, setDeleteLoading] = useState(false);

   const handleUpdateProfile = async (e: React.FormEvent) => {
      e.preventDefault();
      setProfileMsg("");
      setProfileError("");
      setProfileLoading(true);
      try {
         await apiClient.updateProfile(email);
         setProfileMsg("Profile updated successfully");
      } catch (err) {
         setProfileError(err instanceof Error ? err.message : "Failed to update");
      } finally {
         setProfileLoading(false);
      }
   };

   const handleChangePassword = async (e: React.FormEvent) => {
      e.preventDefault();
      setPasswordMsg("");
      setPasswordError("");

      if (newPassword !== confirmPassword) {
         setPasswordError("Passwords do not match");
         return;
      }

      setPasswordLoading(true);
      try {
         await apiClient.changePassword(currentPassword, newPassword);
         setPasswordMsg("Password changed successfully. Please login again.");
         setCurrentPassword("");
         setNewPassword("");
         setConfirmPassword("");
      } catch (err) {
         setPasswordError(err instanceof Error ? err.message : "Failed to change password");
      } finally {
         setPasswordLoading(false);
      }
   };

   const handleDeleteAccount = async () => {
      if (!confirm("Are you sure you want to delete your account? This action cannot be undone.")) return;
      setDeleteLoading(true);
      try {
         await deleteAccount();
         window.location.href = "/login";
      } catch (err) {
         alert(err instanceof Error ? err.message : "Failed to delete account");
         setDeleteLoading(false);
      }
   };

   return (
      <ProtectedLayout>
         <div className="space-y-6 max-w-2xl">
            <h1 className="text-2xl font-bold text-white">Account</h1>

            {/* Profile */}
            <form onSubmit={handleUpdateProfile} className="bg-gray-900 rounded-xl border border-gray-800 p-6 space-y-4">
               <div className="flex items-center gap-3 mb-2">
                  <UserCircle className="h-6 w-6 text-orange-500" />
                  <h2 className="text-lg font-semibold text-white">Profile</h2>
               </div>

               {profileMsg && <div className="bg-green-500/10 text-green-400 px-4 py-2 rounded-lg text-sm">{profileMsg}</div>}
               {profileError && <div className="bg-red-500/10 text-red-400 px-4 py-2 rounded-lg text-sm">{profileError}</div>}

               <div>
                  <label className="block text-sm text-gray-400 mb-1">
                     <Mail className="h-4 w-4 inline mr-1" /> Email
                  </label>
                  <input
                     type="email"
                     value={email}
                     onChange={(e) => setEmail(e.target.value)}
                     required
                     className="w-full px-3 py-2 bg-gray-800 border border-gray-700 rounded-lg text-white"
                  />
               </div>

               <div className="text-sm text-gray-500">
                  Role: <span className="text-gray-300">{user?.role}</span> · Member since:{" "}
                  <span className="text-gray-300">{user ? new Date(user.created_at).toLocaleDateString() : "—"}</span>
               </div>

               <button
                  type="submit"
                  disabled={profileLoading}
                  className="px-4 py-2 bg-orange-500 hover:bg-orange-600 disabled:bg-orange-500/50 text-white font-medium rounded-lg transition-colors flex items-center gap-2"
               >
                  {profileLoading && <Loader2 className="h-4 w-4 animate-spin" />}
                  Update Profile
               </button>
            </form>

            {/* Password */}
            <form onSubmit={handleChangePassword} className="bg-gray-900 rounded-xl border border-gray-800 p-6 space-y-4">
               <div className="flex items-center gap-3 mb-2">
                  <Lock className="h-6 w-6 text-orange-500" />
                  <h2 className="text-lg font-semibold text-white">Change Password</h2>
               </div>

               {passwordMsg && <div className="bg-green-500/10 text-green-400 px-4 py-2 rounded-lg text-sm">{passwordMsg}</div>}
               {passwordError && <div className="bg-red-500/10 text-red-400 px-4 py-2 rounded-lg text-sm">{passwordError}</div>}

               <div>
                  <label className="block text-sm text-gray-400 mb-1">Current Password</label>
                  <input
                     type="password"
                     value={currentPassword}
                     onChange={(e) => setCurrentPassword(e.target.value)}
                     required
                     className="w-full px-3 py-2 bg-gray-800 border border-gray-700 rounded-lg text-white"
                  />
               </div>
               <div>
                  <label className="block text-sm text-gray-400 mb-1">New Password</label>
                  <input
                     type="password"
                     value={newPassword}
                     onChange={(e) => setNewPassword(e.target.value)}
                     required
                     minLength={8}
                     className="w-full px-3 py-2 bg-gray-800 border border-gray-700 rounded-lg text-white"
                     placeholder="Min 8 chars, upper, lower, digit"
                  />
               </div>
               <div>
                  <label className="block text-sm text-gray-400 mb-1">Confirm New Password</label>
                  <input
                     type="password"
                     value={confirmPassword}
                     onChange={(e) => setConfirmPassword(e.target.value)}
                     required
                     className="w-full px-3 py-2 bg-gray-800 border border-gray-700 rounded-lg text-white"
                  />
               </div>

               <button
                  type="submit"
                  disabled={passwordLoading}
                  className="px-4 py-2 bg-orange-500 hover:bg-orange-600 disabled:bg-orange-500/50 text-white font-medium rounded-lg transition-colors flex items-center gap-2"
               >
                  {passwordLoading && <Loader2 className="h-4 w-4 animate-spin" />}
                  Change Password
               </button>
            </form>

            {/* Delete Account */}
            <div className="bg-gray-900 rounded-xl border border-red-900/50 p-6">
               <div className="flex items-center gap-3 mb-2">
                  <Trash2 className="h-6 w-6 text-red-500" />
                  <h2 className="text-lg font-semibold text-white">Danger Zone</h2>
               </div>
               <p className="text-sm text-gray-400 mb-4">
                  Permanently delete your account and all associated data. This action cannot be undone.
               </p>
               <button
                  onClick={handleDeleteAccount}
                  disabled={deleteLoading}
                  className="px-4 py-2 bg-red-500 hover:bg-red-600 disabled:bg-red-500/50 text-white font-medium rounded-lg transition-colors flex items-center gap-2"
               >
                  {deleteLoading && <Loader2 className="h-4 w-4 animate-spin" />}
                  Delete Account
               </button>
            </div>
         </div>
      </ProtectedLayout>
   );
}
