import { type ClassValue, clsx } from "clsx";
import { twMerge } from "tailwind-merge";

export function cn(...inputs: ClassValue[]) {
   return twMerge(clsx(inputs));
}

export function formatCurrency(value: number): string {
   return new Intl.NumberFormat("en-US", {
      style: "currency",
      currency: "USD",
   }).format(value);
}

export function formatPercentage(value: number): string {
   return `${value.toFixed(2)}%`;
}

export function formatDate(date: string | Date): string {
   return new Intl.DateTimeFormat("en-US", {
      month: "short",
      day: "numeric",
      year: "numeric",
   }).format(new Date(date));
}

export function formatDateTime(date: string | Date): string {
   return new Intl.DateTimeFormat("en-US", {
      month: "short",
      day: "numeric",
      hour: "2-digit",
      minute: "2-digit",
   }).format(new Date(date));
}

export function getConfidenceColor(confidence: string): string {
   switch (confidence) {
      case "HIGH":
         return "text-profit";
      case "MEDIUM":
         return "text-yellow-500";
      case "LOW":
         return "text-loss";
      default:
         return "text-neutral";
   }
}

export function getEVColor(ev: number): string {
   if (ev > 0.05) return "text-profit";
   if (ev > 0) return "text-yellow-500";
   return "text-loss";
}
