import { defineTool } from '@eagi/sdk';
import { z } from 'zod';

export default defineTool({
  name: "list_industries",
  description: "List the 7 industries covered by GapBase and the number of validated gaps in each.",
  input: z.object({}),
  handler: async (input, ctx) => {
    const db = ctx.services.gapbase;
    return {
      industries: db.industryCounts(),
      total_gaps: db.gaps.length,
      _note: db.UPGRADE_NOTE,
    };
  }
});
