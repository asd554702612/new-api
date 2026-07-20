import assert from 'node:assert/strict';
import {
  buildHeaderNavLinks,
  normalizeHeaderNavModules,
} from './useNavigation.js';

const t = (key) => key;

const defaultLinks = buildHeaderNavLinks(t, 'https://docs.example.com');

assert.ok(
  defaultLinks.some(
    (link) =>
      link.itemKey === 'feedback' &&
      link.to === '/feedback' &&
      link.text === '投诉反馈',
  ),
  'default header navigation exposes the public feedback entry',
);

assert.equal(
  normalizeHeaderNavModules({
    home: true,
    console: true,
    pricing: {
      enabled: true,
      requireAuth: false,
    },
    docs: true,
    about: true,
  }).feedback,
  true,
  'legacy header navigation config enables feedback by default',
);

assert.equal(
  buildHeaderNavLinks(t, '', { feedback: false }).some(
    (link) => link.itemKey === 'feedback',
  ),
  false,
  'feedback entry can be hidden through header navigation config',
);

console.log('useNavigation tests passed');
