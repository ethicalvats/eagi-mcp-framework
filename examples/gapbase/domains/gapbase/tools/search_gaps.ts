import { defineTool } from '@eagi/sdk';
import { z } from 'zod';

export default defineTool({
  name: "search_gaps",
  description: "PRIMARY TOOL for finding startup ideas. Searches validated startup gaps across 7 industries.",
  input: z.object({
    query: z.string().optional().describe("Free-text keyword to search across pain points and solutions."),
    industry: z.string().optional().describe("Filter to one industry."),
    limit: z.number().default(10).describe("Max results to return (1-25, default 10)."),
  }),
  handler: async (input, ctx) => {
    const db = ctx.services.gapbase;
    const { query, industry, limit } = input;
    
    const safeLimit = Math.max(1, Math.min(25, Number(limit) || 10));
    let pool = db.gaps;

    if (industry) {
      const ind = String(industry).toLowerCase();
      if (!db.VALID_INDUSTRIES.includes(ind)) {
        return {
          error: `Invalid industry '${industry}'. Valid: ${db.VALID_INDUSTRIES.join(", ")}`,
        };
      }
      pool = pool.filter((g: any) => g.industry === ind);
    }

    let results;
    if (query && query.trim()) {
      const scored = pool
        .map((g: any) => ({ gap: g, score: db.scoreGap(g, query) }))
        .filter((s: any) => s.score > 0)
        .sort((a: any, b: any) => b.score - a.score);
      results = scored.slice(0, safeLimit).map((s: any) => s.gap);
    } else {
      results = pool.slice(0, safeLimit);
    }

    return {
      total_in_database: db.gaps.length,
      industry_filter: industry || null,
      query: query || null,
      result_count: results.length,
      results,
      _note: db.UPGRADE_NOTE,
    };
  }
});
