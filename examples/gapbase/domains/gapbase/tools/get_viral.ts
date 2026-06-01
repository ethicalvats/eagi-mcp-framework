import { defineTool } from '@eagi/sdk';
import { z } from 'zod';

export default defineTool({
  name: "get_viral_social_gaps",
  description: "SPECIALIZED TOOL — only use when the user explicitly asks for VIRAL, SOCIAL, TIKTOK, CONSUMER, TREND-BASED, or SHORT-WINDOW build opportunities.",
  input: z.object({}),
  handler: async (input, ctx) => {
    const db = ctx.services.gapbase;
    return {
      viral_social_gaps: db.trends,
      _note: "These are VIRAL CONSUMER trends with short peak windows (days to weeks). For serious B2B / validated business opportunities, use the `search_gaps` tool.",
    };
  }
});
