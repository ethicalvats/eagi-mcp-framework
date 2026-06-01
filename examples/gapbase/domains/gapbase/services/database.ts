import { defineService } from '@eagi/sdk';
import { readFileSync } from 'node:fs';
import { join } from 'node:path';

const DATA_DIR = join(process.cwd(), 'data');

export const GapBaseService = defineService('gapbase', async () => {
  const gaps = JSON.parse(readFileSync(join(DATA_DIR, 'gaps.json'), 'utf8'));
  const trends = JSON.parse(readFileSync(join(DATA_DIR, 'trends.json'), 'utf8'));

  const VALID_INDUSTRIES = [
    "accounting",
    "dental",
    "ecommerce",
    "healthcare",
    "legal",
    "property",
    "veterinary",
  ];

  const UPGRADE_NOTE =
    "Full blueprint lives at the `full_blueprint` URL. Visit thevibepreneur.com.";

  function industryCounts() {
    const counts: Record<string, number> = {};
    for (const g of gaps) {
      counts[g.industry] = (counts[g.industry] || 0) + 1;
    }
    return Object.entries(counts)
      .map(([industry, count]) => ({ industry, count }))
      .sort((a, b) => b.count - a.count);
  }

  function scoreGap(gap: any, query: string) {
    if (!query) return 0;
    const q = query.toLowerCase();
    const pain = (gap.pain || "").toLowerCase();
    const solution = (gap.solution || "").toLowerCase();
    const role = (gap.role || "").toLowerCase();

    let score = 0;
    if (pain.includes(q)) score += 10;
    if (pain.startsWith(q)) score += 5;
    if (solution.includes(q)) score += 6;
    if (role.includes(q)) score += 4;

    const tokens = q.split(/\s+/).filter((t: string) => t.length > 2);
    for (const tok of tokens) {
      if (pain.includes(tok)) score += 2;
      if (solution.includes(tok)) score += 1;
    }
    return score;
  }

  return {
    gaps,
    trends,
    VALID_INDUSTRIES,
    UPGRADE_NOTE,
    industryCounts,
    scoreGap,
  };
});
