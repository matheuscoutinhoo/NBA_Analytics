import {
   formatCurrency,
   formatPercentage,
   getConfidenceColor,
   getEVColor,
} from "@/lib/utils";

describe("formatCurrency", () => {
   it("formats positive values", () => {
      expect(formatCurrency(1234.56)).toBe("$1,234.56");
   });

   it("formats zero", () => {
      expect(formatCurrency(0)).toBe("$0.00");
   });

   it("formats negative values", () => {
      expect(formatCurrency(-500)).toBe("-$500.00");
   });
});

describe("formatPercentage", () => {
   it("formats percentage with 2 decimals", () => {
      expect(formatPercentage(65.123)).toBe("65.12%");
   });

   it("formats zero", () => {
      expect(formatPercentage(0)).toBe("0.00%");
   });
});

describe("getConfidenceColor", () => {
   it('returns green for HIGH', () => {
      expect(getConfidenceColor("HIGH")).toBe("text-profit");
   });

   it('returns yellow for MEDIUM', () => {
      expect(getConfidenceColor("MEDIUM")).toBe("text-yellow-500");
   });

   it('returns red for LOW', () => {
      expect(getConfidenceColor("LOW")).toBe("text-loss");
   });

   it('returns neutral for unknown', () => {
      expect(getConfidenceColor("UNKNOWN")).toBe("text-neutral");
   });
});

describe("getEVColor", () => {
   it('returns green for high positive EV', () => {
      expect(getEVColor(0.10)).toBe("text-profit");
   });

   it('returns yellow for marginal positive EV', () => {
      expect(getEVColor(0.02)).toBe("text-yellow-500");
   });

   it('returns red for negative EV', () => {
      expect(getEVColor(-0.10)).toBe("text-loss");
   });
});
