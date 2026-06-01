import { defineTool } from '@eagi/sdk';
import { z } from 'zod';

export default defineTool({
  name: "get_stats",
  description: "Get GapBase database statistics: total gaps, industry breakdown, trending gap count.",
  input: z.object({}),
  handler: async (input, ctx) => {
    const db = ctx.services.gapbase;
    return {
      total_gaps: db.gaps.length,
      industries: db.industryCounts(),
      viral_social_gaps: db.trends.length,
      database: "GapBase",
      website: "https://thevibepreneur.com",
      source: "Validated pain points from Reddit, LinkedIn, and X.",
      _note: db.UPGRADE_NOTE,
    };
  }
});
