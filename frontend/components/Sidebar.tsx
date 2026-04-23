"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useState } from "react";
import { useAuth } from "@/hooks/useAuth";
import {
   LayoutDashboard,
   Gamepad2,
   Receipt,
   Wallet,
   UserCircle,
   LogOut,
   Menu,
   X,
   TrendingUp,
} from "lucide-react";

const navItems = [
   { href: "/dashboard", label: "Dashboard", icon: LayoutDashboard },
   { href: "/games", label: "Games", icon: Gamepad2 },
   { href: "/bets", label: "Bets", icon: Receipt },
   { href: "/bankroll", label: "Bankroll", icon: Wallet },
   { href: "/account", label: "Account", icon: UserCircle },
];

export default function Sidebar() {
   const [mobileOpen, setMobileOpen] = useState(false);
   const pathname = usePathname();
   const { logout, user } = useAuth();

   const handleLogout = async () => {
      await logout();
      window.location.href = "/login";
   };

   return (
      <>
         {/* Mobile header */}
         <div className="lg:hidden fixed top-0 left-0 right-0 z-50 bg-gray-900 border-b border-gray-800 px-4 py-3 flex items-center justify-between">
            <div className="flex items-center gap-2">
               <TrendingUp className="h-6 w-6 text-orange-500" />
               <span className="text-lg font-bold text-white">NBA Bet Insights</span>
            </div>
            <button
               onClick={() => setMobileOpen(!mobileOpen)}
               className="text-gray-400 hover:text-white"
               aria-label="Toggle menu"
            >
               {mobileOpen ? <X className="h-6 w-6" /> : <Menu className="h-6 w-6" />}
            </button>
         </div>

         {/* Mobile overlay */}
         {mobileOpen && (
            <div
               className="lg:hidden fixed inset-0 z-40 bg-black/50"
               onClick={() => setMobileOpen(false)}
            />
         )}

         {/* Sidebar */}
         <aside
            className={`fixed top-0 left-0 z-40 h-full w-64 bg-gray-900 border-r border-gray-800 transition-transform duration-300 ${mobileOpen ? "translate-x-0" : "-translate-x-full"
               } lg:translate-x-0`}
         >
            <div className="flex flex-col h-full">
               {/* Logo */}
               <div className="hidden lg:flex items-center gap-2 px-6 py-5 border-b border-gray-800">
                  <TrendingUp className="h-7 w-7 text-orange-500" />
                  <span className="text-xl font-bold text-white">NBA Bet Insights</span>
               </div>

               {/* Nav */}
               <nav className="flex-1 py-6 px-3 mt-14 lg:mt-0">
                  <ul className="space-y-1">
                     {navItems.map((item) => {
                        const isActive = pathname === item.href || pathname?.startsWith(item.href + "/");
                        const Icon = item.icon;
                        return (
                           <li key={item.href}>
                              <Link
                                 href={item.href}
                                 onClick={() => setMobileOpen(false)}
                                 className={`flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium transition-colors ${isActive
                                       ? "bg-orange-500/10 text-orange-500"
                                       : "text-gray-400 hover:text-white hover:bg-gray-800"
                                    }`}
                              >
                                 <Icon className="h-5 w-5" />
                                 {item.label}
                              </Link>
                           </li>
                        );
                     })}
                  </ul>
               </nav>

               {/* User section */}
               <div className="border-t border-gray-800 px-3 py-4">
                  <div className="px-3 py-2 text-sm text-gray-500 truncate">
                     {user?.email}
                  </div>
                  <button
                     onClick={handleLogout}
                     className="flex items-center gap-3 w-full px-3 py-2.5 rounded-lg text-sm font-medium text-gray-400 hover:text-red-400 hover:bg-gray-800 transition-colors"
                  >
                     <LogOut className="h-5 w-5" />
                     Logout
                  </button>
               </div>
            </div>
         </aside>
      </>
   );
}
