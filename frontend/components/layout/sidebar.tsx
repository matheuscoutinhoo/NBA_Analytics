"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import {
   BarChart3,
   Home,
   Wallet,
   Settings,
   Trophy,
   TrendingUp,
   Shield,
   Menu,
   X,
   LogOut,
} from "lucide-react";
import { cn } from "@/lib/utils";
import { useAuthStore, useUIStore } from "@/store";
import { Button } from "@/components/ui/button";
import { api } from "@/lib/api";

const navItems = [
   { href: "/dashboard", label: "Dashboard", icon: Home },
   { href: "/dashboard/games", label: "Games & Analysis", icon: BarChart3 },
   { href: "/dashboard/bankroll", label: "Bankroll", icon: Wallet },
   { href: "/dashboard/performance", label: "Performance", icon: TrendingUp },
   { href: "/dashboard/achievements", label: "Achievements", icon: Trophy },
   { href: "/dashboard/settings", label: "Settings", icon: Settings },
];

export function Sidebar() {
   const pathname = usePathname();
   const { user, logout } = useAuthStore();
   const { sidebarOpen, toggleSidebar } = useUIStore();

   const handleLogout = async () => {
      try {
         await api.logout();
      } catch {
         // Ignore errors
      }
      logout();
      window.location.href = "/login";
   };

   return (
      <>
         {/* Mobile toggle */}
         <button
            onClick={toggleSidebar}
            className="fixed top-4 left-4 z-50 lg:hidden bg-card rounded-md p-2 border"
         >
            {sidebarOpen ? <X size={20} /> : <Menu size={20} />}
         </button>

         {/* Sidebar */}
         <aside
            className={cn(
               "fixed inset-y-0 left-0 z-40 w-64 bg-card border-r transform transition-transform duration-200 ease-in-out lg:translate-x-0",
               sidebarOpen ? "translate-x-0" : "-translate-x-full"
            )}
         >
            <div className="flex flex-col h-full">
               {/* Logo */}
               <div className="flex items-center gap-2 p-6 border-b">
                  <div className="w-8 h-8 bg-primary rounded-lg flex items-center justify-center">
                     <BarChart3 className="w-5 h-5 text-primary-foreground" />
                  </div>
                  <div>
                     <h1 className="font-bold text-lg">BETTER</h1>
                     <p className="text-xs text-muted-foreground">NBA Analytics</p>
                  </div>
               </div>

               {/* Navigation */}
               <nav className="flex-1 px-3 py-4 space-y-1 overflow-y-auto">
                  {navItems.map((item) => {
                     const isActive = pathname === item.href;
                     return (
                        <Link
                           key={item.href}
                           href={item.href}
                           className={cn(
                              "flex items-center gap-3 px-3 py-2.5 rounded-md text-sm font-medium transition-colors",
                              isActive
                                 ? "bg-primary/10 text-primary"
                                 : "text-muted-foreground hover:bg-accent hover:text-foreground"
                           )}
                        >
                           <item.icon size={18} />
                           {item.label}
                        </Link>
                     );
                  })}
               </nav>

               {/* Disclaimer */}
               <div className="px-4 py-3 mx-3 mb-3 bg-muted/50 rounded-md">
                  <div className="flex items-center gap-1.5 text-xs text-muted-foreground">
                     <Shield size={12} />
                     <span className="font-medium">Informational Only</span>
                  </div>
                  <p className="text-[10px] text-muted-foreground mt-1">
                     No financial advice. No guaranteed returns. 18+ only.
                  </p>
               </div>

               {/* User section */}
               <div className="border-t p-4">
                  <div className="flex items-center justify-between">
                     <div className="flex items-center gap-2 min-w-0">
                        <div className="w-8 h-8 rounded-full bg-primary/20 flex items-center justify-center text-xs font-bold text-primary">
                           {user?.username?.charAt(0).toUpperCase()}
                        </div>
                        <div className="truncate">
                           <p className="text-sm font-medium truncate">{user?.username}</p>
                           <p className="text-xs text-muted-foreground">{user?.xp_points} XP</p>
                        </div>
                     </div>
                     <Button variant="ghost" size="icon" onClick={handleLogout}>
                        <LogOut size={16} />
                     </Button>
                  </div>
               </div>
            </div>
         </aside>

         {/* Mobile overlay */}
         {sidebarOpen && (
            <div
               className="fixed inset-0 bg-black/50 z-30 lg:hidden"
               onClick={toggleSidebar}
            />
         )}
      </>
   );
}
