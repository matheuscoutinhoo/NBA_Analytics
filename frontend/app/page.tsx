import Link from "next/link";
import { BarChart3, Shield, TrendingUp, Wallet } from "lucide-react";

export default function HomePage() {
   return (
      <div className="min-h-screen bg-background">
         {/* Header */}
         <header className="border-b">
            <div className="container mx-auto flex items-center justify-between py-4 px-4">
               <div className="flex items-center gap-2">
                  <div className="w-8 h-8 bg-primary rounded-lg flex items-center justify-center">
                     <BarChart3 className="w-5 h-5 text-primary-foreground" />
                  </div>
                  <span className="font-bold text-xl">BETTER</span>
               </div>
               <div className="flex items-center gap-4">
                  <Link
                     href="/login"
                     className="text-sm text-muted-foreground hover:text-foreground transition-colors"
                  >
                     Sign In
                  </Link>
                  <Link
                     href="/register"
                     className="bg-primary text-primary-foreground px-4 py-2 rounded-md text-sm font-medium hover:bg-primary/90 transition-colors"
                  >
                     Get Started
                  </Link>
               </div>
            </div>
         </header>

         {/* Hero */}
         <section className="container mx-auto px-4 py-20 text-center">
            <h1 className="text-4xl md:text-6xl font-bold mb-6 bg-gradient-to-r from-primary to-blue-400 bg-clip-text text-transparent">
               NBA Analytics & Intelligence
            </h1>
            <p className="text-lg md:text-xl text-muted-foreground max-w-2xl mx-auto mb-8">
               Transform historical NBA data into actionable statistical insights.
               Advanced probabilistic analysis powered by AI.
            </p>
            <div className="flex flex-col sm:flex-row gap-4 justify-center">
               <Link
                  href="/register"
                  className="bg-primary text-primary-foreground px-8 py-3 rounded-lg text-lg font-semibold hover:bg-primary/90 transition-colors"
               >
                  Start Analyzing
               </Link>
               <Link
                  href="#features"
                  className="border px-8 py-3 rounded-lg text-lg font-semibold hover:bg-accent transition-colors"
               >
                  Learn More
               </Link>
            </div>

            {/* Disclaimer */}
            <div className="mt-8 inline-flex items-center gap-2 text-sm text-muted-foreground bg-muted/50 px-4 py-2 rounded-full">
               <Shield size={14} />
               <span>Informational purposes only. No betting intermediation.</span>
            </div>
         </section>

         {/* Features */}
         <section id="features" className="container mx-auto px-4 py-16">
            <div className="grid md:grid-cols-3 gap-8">
               <div className="bg-card rounded-xl border p-8">
                  <div className="w-12 h-12 bg-primary/10 rounded-lg flex items-center justify-center mb-4">
                     <BarChart3 className="w-6 h-6 text-primary" />
                  </div>
                  <h3 className="text-xl font-semibold mb-2">Probabilistic Analysis</h3>
                  <p className="text-muted-foreground">
                     AI-powered game analysis with win probabilities, confidence levels,
                     and detailed statistical breakdowns.
                  </p>
               </div>

               <div className="bg-card rounded-xl border p-8">
                  <div className="w-12 h-12 bg-primary/10 rounded-lg flex items-center justify-center mb-4">
                     <TrendingUp className="w-6 h-6 text-primary" />
                  </div>
                  <h3 className="text-xl font-semibold mb-2">Value Detection</h3>
                  <p className="text-muted-foreground">
                     Compare model probabilities with market odds to identify potential
                     value opportunities through EV calculations.
                  </p>
               </div>

               <div className="bg-card rounded-xl border p-8">
                  <div className="w-12 h-12 bg-primary/10 rounded-lg flex items-center justify-center mb-4">
                     <Wallet className="w-6 h-6 text-primary" />
                  </div>
                  <h3 className="text-xl font-semibold mb-2">Bankroll Management</h3>
                  <p className="text-muted-foreground">
                     Track your performance manually with ROI metrics, drawdown analysis,
                     profit evolution charts, and personal limits.
                  </p>
               </div>
            </div>
         </section>

         {/* Responsible Gaming */}
         <section className="bg-muted/30 py-16">
            <div className="container mx-auto px-4 text-center">
               <Shield className="w-10 h-10 mx-auto mb-4 text-muted-foreground" />
               <h2 className="text-2xl font-bold mb-4">Responsible Gaming</h2>
               <div className="max-w-xl mx-auto space-y-2 text-sm text-muted-foreground">
                  <p>This platform is for informational purposes only.</p>
                  <p>We do not accept, intermediate, or execute bets.</p>
                  <p>Past performance does not guarantee future results.</p>
                  <p>Must be 18+ to use this platform.</p>
                  <p>If you need help, visit <a href="https://www.ncpgambling.org" className="text-primary hover:underline" target="_blank" rel="noopener noreferrer">ncpgambling.org</a></p>
               </div>
            </div>
         </section>

         {/* Footer */}
         <footer className="border-t py-8">
            <div className="container mx-auto px-4 text-center text-sm text-muted-foreground">
               <p>&copy; {new Date().getFullYear()} Better Analytics. All rights reserved.</p>
               <div className="flex justify-center gap-4 mt-2">
                  <Link href="/privacy" className="hover:text-foreground">Privacy Policy</Link>
                  <Link href="/terms" className="hover:text-foreground">Terms of Service</Link>
               </div>
            </div>
         </footer>
      </div>
   );
}
