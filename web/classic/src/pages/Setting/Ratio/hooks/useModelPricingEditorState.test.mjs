/*
Copyright (C) 2025 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
import assert from 'node:assert/strict';
import { describe, test } from 'node:test';

import {
  buildVideoBillingRule,
  buildPreviewRows,
  buildSummaryText,
} from './modelPricingStateLogic.js';

const t = (key, params) => {
  if (!params) return key;
  return Object.entries(params).reduce(
    (text, [name, value]) => text.replace(`{{${name}}}`, value),
    key,
  );
};

describe('video per-second pricing summary', () => {
  test('summarizes video billing as per-second pricing', () => {
    const summary = buildSummaryText(
      {
        billingMode: 'video_per_second',
        videoSecondPrice: '0.03',
      },
      t,
    );

    assert.equal(summary, '视频计费 $0.03 / 秒');
  });

  test('previews the video billing rule fields', () => {
    const rows = buildPreviewRows(
      {
        billingMode: 'video_per_second',
        videoSecondPrice: '0.03',
      },
      t,
    );

    assert.deepEqual(rows, [
      {
        key: 'VideoBillingMode',
        label: 'VideoBillingMode',
        value: 'per_second',
      },
      {
        key: 'VideoBasePrice',
        label: 'VideoBasePrice',
        value: '0.03',
      },
    ]);
  });

  test('summarizes video billing with resolution multipliers', () => {
    const summary = buildSummaryText(
      {
        billingMode: 'video_per_second',
        videoSecondPrice: '0.03',
        videoResolutionMultipliers: {
          '480p': '1',
          '720p': '1.5',
          '1080p': '2',
        },
      },
      t,
    );

    assert.equal(summary, '视频计费 $0.03 / 秒，分辨率倍率 3');
  });

  test('previews the video billing matrix rule fields', () => {
    const rows = buildPreviewRows(
      {
        billingMode: 'video_per_second',
        videoSecondPrice: '0.03',
        videoResolutionMultipliers: {
          '480p': '1',
          '720p': '1.5',
          '1080p': '2',
        },
      },
      t,
    );

    assert.deepEqual(rows, [
      {
        key: 'VideoBillingMode',
        label: 'VideoBillingMode',
        value: 'matrix',
      },
      {
        key: 'VideoBasePrice',
        label: 'VideoBasePrice',
        value: '0.03',
      },
      {
        key: 'VideoResolutionMultipliers',
        label: 'VideoResolutionMultipliers',
        value: '480p: 1, 720p: 1.5, 1080p: 2',
      },
    ]);
  });

  test('builds a video billing matrix rule when resolution multipliers exist', () => {
    const result = buildVideoBillingRule({
      billingMode: 'video_per_second',
      videoSecondPrice: '0.03',
      videoResolutionMultipliers: {
        '480p': '1',
        '720p': '1.5',
        '1080p': '2',
      },
    });

    assert.deepEqual(result, {
      rule: {
        mode: 'matrix',
        base_price: 0.03,
        multipliers: {
          resolution: {
            '480p': 1,
            '720p': 1.5,
            '1080p': 2,
          },
        },
      },
      invalidResolution: null,
    });
  });
});
