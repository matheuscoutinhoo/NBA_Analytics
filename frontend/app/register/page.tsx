"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { BarChart3, Eye, EyeOff, Shield } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { useAuthStore } from "@/store";
import { api } from "@/lib/api";

export default function RegisterPage() {
   const [username, setUsername] = useState("");
   const [email, setEmail] = useState("");
   const [password, setPassword] = useState("");
   const [showPassword, setShowPassword] = useState(false);
   const [ageVerified, setAgeVerified] = useState(false);
   const [consent, setConsent] = useState(false);
   const [error, setError] = useState("");
   const [loading, setLoading] = useState(false);
   const { setUser, setTokens } = useAuthStore();
   const router = useRouter();

   const handleSubmit = async (e: React.FormEvent) => {
      e.preventDefault();
      setError("");

      if (!ageVerified) {
         setError("You must verify you are 18 or older");
         return;
      }
      if (!consent) {
         setError("You must consent to data processing");
         return;
      }

      setLoading(true);

      try {
         const response = await api.register({
            username,
            email,
            password,
            age_verified: ageVerified,
            consent,
         });
         const data = (response as any).data;
         setUser(data.user);
         setTokens(data.tokens.access_token, data.tokens.refresh_token);
         router.push("/dashboard");
      } catch (err: any) {
         setError(err.message || "Registration failed");
      } finally {
         setLoading(false);
      }
   };

   return (
      <div className="min-h-screen flex items-center justify-center bg-background px-4 py-8">
         <Card className="w-full max-w-md">
            <CardHeader className="text-center">
               <div className="flex justify-center mb-4">
                  <div className="w-12 h-12 bg-primary rounded-xl flex items-center justify-center">
                     <BarChart3 className="w-7 h-7 text-primary-foreground" />
                  </div>
               </div>
               <CardTitle className="text-2xl">Create Account</CardTitle>
               <CardDescription>Start your NBA analytics journey</CardDescription>
            </CardHeader>
            <CardContent>
               <form onSubmit={handleSubmit} className="space-y-4">
                  {error && (
                     <div className="bg-destructive/10 text-destructive text-sm p-3 rounded-md">
                        {error}
                     </div>
                  )}
                  <div className="space-y-2">
                     <Label htmlFor="username">Username</Label>
                     <Input
                        id="username"
                        type="text"
                        placeholder="analyst_pro"
                        value={username}
                        onChange={(e) => setUsername(e.target.value)}
                        required
                        minLength={3}
                        maxLength={30}
                        pattern="[a-zA-Z0-9_]+"
                        autoComplete="username"
                     />
                  </div>
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
                           placeholder="Min 8 characters"
                           value={password}
                           onChange={(e) => setPassword(e.target.value)}
                           required
                           minLength={8}
                           maxLength={72}
                           autoComplete="new-password"
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

                  {/* Age verification */}
                  <div className="flex items-start gap-2">
                     <input
                        type="checkbox"
                        id="age"
                        checked={ageVerified}
                        onChange={(e) => setAgeVerified(e.target.checked)}
                        className="mt-1 rounded border-input"
                     />
                     <label htmlFor="age" className="text-sm text-muted-foreground">
                        I confirm I am <strong>18 years or older</strong>
                     </label>
                  </div>

                  {/* Consent - LGPD */}
                  <div className="flex items-start gap-2">
                     <input
                        type="checkbox"
                        id="consent"
                        checked={consent}
                        onChange={(e) => setConsent(e.target.checked)}
                        className="mt-1 rounded border-input"
                     />
                     <label htmlFor="consent" className="text-sm text-muted-foreground">
                        I consent to the processing of my data as described in the{" "}
                        <Link href="/privacy" className="text-primary hover:underline">
                           Privacy Policy
                        </Link>
                        . I can withdraw consent and request data deletion at any time.
                     </label>
                  </div>

                  {/* Disclaimer */}
                  <div className="bg-muted/50 p-3 rounded-md flex items-start gap-2">
                     <Shield size={16} className="text-muted-foreground mt-0.5 flex-shrink-0" />
                     <p className="text-xs text-muted-foreground">
                        This platform is for informational purposes only. It does not accept,
                        intermediate, or execute bets. Past performance does not guarantee future results.
                     </p>
                  </div>

                  <Button type="submit" className="w-full" disabled={loading}>
                     {loading ? "Creating account..." : "Create Account"}
                  </Button>
               </form>
               <div className="mt-4 text-center text-sm text-muted-foreground">
                  Already have an account?{" "}
                  <Link href="/login" className="text-primary hover:underline">
                     Sign in
                  </Link>
               </div>
            </CardContent>
         </Card>
      </div>
   );
}
