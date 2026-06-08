import assert from 'node:assert/strict';
import {
  buildHeaderNavModulesWithRankings,
  formatCompactNumber,
  normalizeRankingsModule,
} from './utils.js';

assert.deepEqual(normalizeRankingsModule(undefined), {
  enabled: true,
  requireAuth: true,
});

assert.deepEqual(normalizeRankingsModule(false), {
  enabled: false,
  requireAuth: true,
});

assert.deepEqual(
  normalizeRankingsModule({
    enabled: true,
    requireAuth: true,
  }),
  {
    enabled: true,
    requireAuth: true,
  },
);

assert.deepEqual(
  buildHeaderNavModulesWithRankings({
    home: true,
    console: true,
    pricing: {
      enabled: true,
      requireAuth: false,
    },
    docs: true,
    about: true,
  }).rankings,
  {
    enabled: true,
    requireAuth: true,
  },
);

assert.equal(formatCompactNumber(0), '0');
assert.equal(formatCompactNumber(999), '999');
assert.equal(formatCompactNumber(1530), '1.5K');
assert.equal(formatCompactNumber(2_500_000), '2.5M');

console.log('rankings utils tests passed');
