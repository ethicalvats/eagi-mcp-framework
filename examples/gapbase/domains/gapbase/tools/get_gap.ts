import { defineTool } from '@eagi/sdk';
import { z } from 'zod';

export default defineTool({
  name: "get_gap",
  description: "Fetch a single gap by its id or slug. Returns problem + solution direction + blueprint URL.",
  input: z.object({
    id: z.string().describe("Gap id (e.g. 'pd-la001') or slug."),
  }),
  handler: async (input, ctx) => {
    const db = ctx.services.gapbase;
    const idOrSlug = input.id;
    
    const needle = String(idOrSlug).toLowerCase();
    const gap = db.gaps.find(
      (g: any) => g.id.toLowerCase() === needle || g.slug.toLowerCase() === needle
    );
    if (!gap) return { error: \`Gap not found: \${idOrSlug}\` };
    
    return { ...gap, _note: db.UPGRADE_NOTE };
  }
});
