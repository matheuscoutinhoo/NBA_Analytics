"use client";

import { Trophy, Award } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { Progress } from "@/components/ui/progress";
import { useAuthStore } from "@/store";

const badges = [
   { name: "First Analysis", description: "View your first game analysis", icon: "🏆", xp: 10, unlocked: false },
   { name: "Analyst Rookie", description: "View 10 game analyses", icon: "📊", xp: 25, unlocked: false },
   { name: "Data Driven", description: "View 50 game analyses", icon: "🗄️", xp: 50, unlocked: false },
   { name: "Bankroll Starter", description: "Make your first bankroll entry", icon: "💰", xp: 10, unlocked: false },
   { name: "Consistent Tracker", description: "Track 30 consecutive days", icon: "📅", xp: 100, unlocked: false },
   { name: "Profit Master", description: "Positive ROI over 50 entries", icon: "📈", xp: 75, unlocked: false },
   { name: "Risk Manager", description: "Set personal limits", icon: "🛡️", xp: 15, unlocked: false },
   { name: "Value Hunter", description: "Identify 10 value opportunities", icon: "🔍", xp: 30, unlocked: false },
];

const levels = [
   { level: 1, name: "Rookie", minXP: 0 },
   { level: 2, name: "Analyst", minXP: 50 },
   { level: 3, name: "Strategist", minXP: 150 },
   { level: 4, name: "Expert", minXP: 350 },
   { level: 5, name: "Master", minXP: 700 },
];

export default function AchievementsPage() {
   const { user } = useAuthStore();
   const xp = user?.xp_points || 0;

   const currentLevel = levels.reduce((acc, lvl) => (xp >= lvl.minXP ? lvl : acc), levels[0]);
   const nextLevel = levels.find((lvl) => lvl.minXP > xp) || levels[levels.length - 1];
   const progressToNext = nextLevel.minXP > currentLevel.minXP
      ? ((xp - currentLevel.minXP) / (nextLevel.minXP - currentLevel.minXP)) * 100
      : 100;

   return (
      <div className="space-y-6">
         <div>
            <h1 className="text-2xl font-bold">Achievements</h1>
            <p className="text-muted-foreground">Track your progress and earn badges</p>
         </div>

         {/* Level progress */}
         <Card>
            <CardContent className="pt-6">
               <div className="flex items-center justify-between mb-4">
                  <div className="flex items-center gap-3">
                     <div className="w-12 h-12 rounded-full bg-primary/20 flex items-center justify-center text-xl">
                        <Trophy className="w-6 h-6 text-primary" />
                     </div>
                     <div>
                        <p className="font-bold text-lg">Level {currentLevel.level}: {currentLevel.name}</p>
                        <p className="text-sm text-muted-foreground">{xp} XP total</p>
                     </div>
                  </div>
                  <div className="text-right">
                     <p className="text-sm text-muted-foreground">Next: {nextLevel.name}</p>
                     <p className="text-xs text-muted-foreground">{nextLevel.minXP - xp} XP needed</p>
                  </div>
               </div>
               <Progress value={progressToNext} className="h-3" />
            </CardContent>
         </Card>

         {/* Badges */}
         <div>
            <h2 className="text-lg font-semibold mb-4 flex items-center gap-2">
               <Award size={20} />
               Badges
            </h2>
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
               {badges.map((badge) => (
                  <Card
                     key={badge.name}
                     className={badge.unlocked ? "" : "opacity-60"}
                  >
                     <CardContent className="pt-6 text-center">
                        <div className="text-3xl mb-2">{badge.icon}</div>
                        <p className="font-semibold text-sm">{badge.name}</p>
                        <p className="text-xs text-muted-foreground mt-1">{badge.description}</p>
                        <div className="mt-2">
                           <span className="text-xs bg-primary/10 text-primary px-2 py-0.5 rounded-full">
                              +{badge.xp} XP
                           </span>
                        </div>
                        {badge.unlocked && (
                           <p className="text-xs text-profit mt-2 font-medium">Unlocked!</p>
                        )}
                     </CardContent>
                  </Card>
               ))}
            </div>
         </div>
      </div>
   );
}
