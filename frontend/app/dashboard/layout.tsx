"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { Sidebar } from "@/components/layout/sidebar";
import { MobileNav } from "@/components/layout/mobile-nav";
import { useAuthStore } from "@/store";
import { api } from "@/lib/api";

export default function DashboardLayout({
   children,
}: {
   children: React.ReactNode;
}) {
   const { isAuthenticated, isLoading, setUser, setLoading } = useAuthStore();
   const router = useRouter();

   useEffect(() => {
      const checkAuth = async () => {
         const token = localStorage.getItem("access_token");
         if (!token) {
            setLoading(false);
            router.push("/login");
            return;
         }

         try {
            const response = await api.getMe();
            setUser((response as any).data);
         } catch {
            setUser(null);
            router.push("/login");
         }
      };

      checkAuth();
   }, [router, setUser, setLoading]);

   if (isLoading) {
      return (
         <div className="flex items-center justify-center min-h-screen">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary" />
         </div>
      );
   }

   if (!isAuthenticated) {
      return null;
   }

   return (
      <div className="min-h-screen bg-background">
         <Sidebar />
         <main className="lg:ml-64 min-h-screen">
            <div className="p-4 md:p-6 lg:p-8 pb-20 lg:pb-8">{children}</div>
         </main>
         <MobileNav />
      </div>
   );
}
