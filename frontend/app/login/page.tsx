"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { BarChart3, Eye, EyeOff } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { useAuthStore } from "@/store";
import { api } from "@/lib/api";

export default function LoginPage() {
   const [email, setEmail] = useState("");
   const [password, setPassword] = useState("");
   const [showPassword, setShowPassword] = useState(false);
   const [error, setError] = useState("");
   const [loading, setLoading] = useState(false);
   const { setUser, setTokens } = useAuthStore();
   const router = useRouter();

   const handleSubmit = async (e: React.FormEvent) => {
      e.preventDefault();
      setError("");
      setLoading(true);

      try {
         const response = await api.login({ email, password });
         const data = (response as any).data;
         setUser(data.user);
         setTokens(data.tokens.access_token, data.tokens.refresh_token);
         router.push("/dashboard");
      } catch (err: any) {
         setError(err.message || "Invalid credentials");
      } finally {
         setLoading(false);
      }
   };

   return (
      <div className="min-h-screen flex items-center justify-center bg-background px-4">
         <Card className="w-full max-w-md">
            <CardHeader className="text-center">
               <div className="flex justify-center mb-4">
                  <div className="w-12 h-12 bg-primary rounded-xl flex items-center justify-center">
                     <BarChart3 className="w-7 h-7 text-primary-foreground" />
                  </div>
               </div>
               <CardTitle className="text-2xl">Welcome back</CardTitle>
               <CardDescription>Sign in to your analytics dashboard</CardDescription>
            </CardHeader>
            <CardContent>
               <form onSubmit={handleSubmit} className="space-y-4">
                  {error && (
                     <div className="bg-destructive/10 text-destructive text-sm p-3 rounded-md">
                        {error}
                     </div>
                  )}
                  <div className="space-y-2">
                     <Label htmlFor="email">Email</Label>
                     <Input
                        id="email"
                        type="email"
                        placeholder="you@example.com"
                        value={email}
                        onChange={(e) => setEmail(e.target.value)}
                        required
                        autoComplete="email"
                     />
                  </div>
                  <div className="space-y-2">
                     <Label htmlFor="password">Password</Label>
                     <div className="relative">
                        <Input
                           id="password"
                           type={showPassword ? "text" : "password"}
                           placeholder="••••••••"
                           value={password}
                           onChange={(e) => setPassword(e.target.value)}
                           required
                           autoComplete="current-password"
                        />
                        <button
                           type="button"
                           onClick={() => setShowPassword(!showPassword)}
                           className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
                        >
                           {showPassword ? <EyeOff size={16} /> : <Eye size={16} />}
                        </button>
                     </div>
                  </div>
                  <Button type="submit" className="w-full" disabled={loading}>
                     {loading ? "Signing in..." : "Sign In"}
                  </Button>
               </form>
               <div className="mt-4 text-center text-sm text-muted-foreground">
                  Don&apos;t have an account?{" "}
                  <Link href="/register" className="text-primary hover:underline">
                     Create one
                  </Link>
               </div>
            </CardContent>
         </Card>
      </div>
   );
}
