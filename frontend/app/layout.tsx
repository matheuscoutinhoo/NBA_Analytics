import type { Metadata } from "next";
import { Inter, JetBrains_Mono } from "next/font/google";
import "./globals.css";
import { Providers } from "@/components/providers";

const inter = Inter({
   subsets: ["latin"],
   variable: "--font-sans",
});

const jetbrainsMono = JetBrains_Mono({
   subsets: ["latin"],
   variable: "--font-mono",
});

export const metadata: Metadata = {
   title: "Better - NBA Analytics Platform",
   description:
      "Advanced NBA probabilistic analysis platform for sports analytics and intelligence. Informational purposes only.",
   keywords: ["NBA", "analytics", "statistics", "basketball", "sports intelligence"],
};

export default function RootLayout({
   children,
}: {
   children: React.ReactNode;
}) {
   return (
      <html lang="en" className="dark" suppressHydrationWarning>
         <body className={`${inter.variable} ${jetbrainsMono.variable} font-sans antialiased`}>
            <Providers>{children}</Providers>
         </body>
      </html>
   );
}
