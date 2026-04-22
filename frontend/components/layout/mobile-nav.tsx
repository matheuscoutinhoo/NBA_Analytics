"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { Home, BarChart3, Wallet, TrendingUp, Settings } from "lucide-react";
import { cn } from "@/lib/utils";

const mobileNavItems = [
   { href: "/dashboard", label: "Home", icon: Home },
   { href: "/dashboard/games", label: "Games", icon: BarChart3 },
   { href: "/dashboard/bankroll", label: "Bankroll", icon: Wallet },
   { href: "/dashboard/performance", label: "Stats", icon: TrendingUp },
   { href: "/dashboard/settings", label: "Settings", icon: Settings },
];

export function MobileNav() {
   const pathname = usePathname();

   return (
      <nav className="fixed bottom-0 left-0 right-0 z-50 bg-card border-t lg:hidden">
         <div className="flex justify-around py-2">
            {mobileNavItems.map((item) => {
               const isActive = pathname === item.href;
               return (
                  <Link
                     key={item.href}
                     href={item.href}
                     className={cn(
                        "flex flex-col items-center gap-1 px-3 py-1.5 text-xs",
                        isActive ? "text-primary" : "text-muted-foreground"
                     )}
                  >
                     <item.icon size={20} />
                     <span>{item.label}</span>
                  </Link>
               );
            })}
         </div>
      </nav>
   );
}
