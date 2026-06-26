import assert from 'node:assert/strict';
import {
  formatLargeNumber,
  formatPayAmount,
  normalizeCurrencyDelta,
} from './format.js';

assert.equal(formatLargeNumber(33.849999999999994), '33.85');
assert.equal(formatLargeNumber(20.31), '20.31');
assert.equal(formatLargeNumber(3385), '3385');
assert.equal(formatLargeNumber('not-a-number'), 'not-a-number');

assert.equal(formatPayAmount(6.769999999999999), '6.77');
assert.equal(formatPayAmount('6.7'), '6.70');
assert.equal(formatPayAmount('not-a-number'), '0.00');

assert.equal(normalizeCurrencyDelta(-0.0000000001), 0);
assert.equal(normalizeCurrencyDelta(0.004), 0);
assert.equal(normalizeCurrencyDelta(0.005), 0.01);
