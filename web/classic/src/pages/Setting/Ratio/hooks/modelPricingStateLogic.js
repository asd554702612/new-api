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

import { combineBillingExpr } from '../components/requestRuleExpr.js';

export const hasValue = (value) =>
  value !== '' && value !== null && value !== undefined && value !== false;

export const VIDEO_RESOLUTION_PRESETS = ['480p', '720p', '1080p'];

const toNumberOrNull = (value) => {
  if (!hasValue(value) && value !== 0) {
    return null;
  }
  const num = Number(value);
  return Number.isFinite(num) ? num : null;
};

const formatNumber = (value) => {
  const num = toNumberOrNull(value);
  if (num === null) {
    return '';
  }
  return parseFloat(num.toFixed(12)).toString();
};

const compareVideoResolutionKeys = (left, right) => {
  const leftIndex = VIDEO_RESOLUTION_PRESETS.indexOf(left);
  const rightIndex = VIDEO_RESOLUTION_PRESETS.indexOf(right);

  if (leftIndex !== -1 || rightIndex !== -1) {
    if (leftIndex === -1) return 1;
    if (rightIndex === -1) return -1;
    return leftIndex - rightIndex;
  }

  return left.localeCompare(right);
};

export const getVideoResolutionMultiplierEntries = (model) => {
  const multipliers = model?.videoResolutionMultipliers;
  if (!multipliers || typeof multipliers !== 'object') {
    return [];
  }

  return Object.entries(multipliers)
    .map(([resolution, value]) => [resolution, formatNumber(value)])
    .filter(([, value]) => hasValue(value))
    .sort(([left], [right]) => compareVideoResolutionKeys(left, right));
};

export const hasVideoResolutionMultipliers = (model) =>
  getVideoResolutionMultiplierEntries(model).length > 0;

export const buildVideoBillingRule = (model) => {
  if (
    model?.billingMode !== 'video_per_second' ||
    !hasValue(model.videoSecondPrice)
  ) {
    return {
      rule: null,
      invalidResolution: null,
    };
  }

  const basePrice = toNumberOrNull(model.videoSecondPrice);
  if (basePrice === null) {
    return {
      rule: null,
      invalidResolution: null,
    };
  }

  const resolutionMultipliers = {};
  const multipliers = model.videoResolutionMultipliers || {};
  Object.keys(multipliers)
    .sort(compareVideoResolutionKeys)
    .forEach((resolution) => {
      const value = multipliers[resolution];
      if (!hasValue(value)) {
        return;
      }

      const multiplier = toNumberOrNull(value);
      if (multiplier === null || multiplier <= 0) {
        resolutionMultipliers[resolution] = null;
        return;
      }

      resolutionMultipliers[resolution] = Number(formatNumber(multiplier));
    });

  const invalidResolution = Object.entries(resolutionMultipliers).find(
    ([, value]) => value === null,
  )?.[0];
  if (invalidResolution) {
    return {
      rule: null,
      invalidResolution,
    };
  }

  if (Object.keys(resolutionMultipliers).length > 0) {
    return {
      rule: {
        mode: 'matrix',
        base_price: Number(formatNumber(basePrice)),
        multipliers: {
          resolution: resolutionMultipliers,
        },
      },
      invalidResolution: null,
    };
  }

  return {
    rule: {
      mode: 'per_second',
      base_price: Number(formatNumber(basePrice)),
    },
    invalidResolution: null,
  };
};

export const buildSummaryText = (model, t) => {
  const requestRuleSuffix =
    model.billingMode === 'tiered_expr' && model.requestRuleExpr
      ? `，${t('请求规则')}`
      : '';

  if (model.billingMode === 'video_per_second') {
    const resolutionMultiplierCount =
      getVideoResolutionMultiplierEntries(model).length;
    const resolutionSuffix =
      resolutionMultiplierCount > 0
        ? `，${t('分辨率倍率')} ${resolutionMultiplierCount}`
        : '';
    return hasValue(model.videoSecondPrice)
      ? `${t('视频计费')} $${model.videoSecondPrice} / ${t('秒')}${resolutionSuffix}`
      : `${t('视频计费')}${resolutionSuffix}`;
  }

  if (model.billingMode === 'tiered_expr') {
    const expr = model.billingExpr;
    if (!expr) return `${t('表达式计费')}${requestRuleSuffix}`;
    const tierCount = (expr.match(/tier\(/g) || []).length;
    if (tierCount === 0) {
      return `${t('表达式计费')}${requestRuleSuffix}`;
    }
    return `${t('阶梯计费')} (${tierCount} ${t('档')})${requestRuleSuffix}`;
  }

  if (model.billingMode === 'per-request' && hasValue(model.fixedPrice)) {
    return `${t('按次')} $${model.fixedPrice} / ${t('次')}${requestRuleSuffix}`;
  }

  if (hasValue(model.inputPrice)) {
    const extraCount = [
      model.completionPrice,
      model.cachePrice,
      model.createCachePrice,
      model.imagePrice,
      model.audioInputPrice,
      model.audioOutputPrice,
    ].filter(hasValue).length;
    const extraLabel =
      extraCount > 0 ? `，${t('额外价格项')} ${extraCount}` : '';
    return `${t('输入')} $${model.inputPrice}${extraLabel}${requestRuleSuffix}`;
  }

  return `${t('未设置价格')}${requestRuleSuffix}`;
};

export const buildPreviewRows = (model, t) => {
  if (!model) return [];

  if (model.billingMode === 'video_per_second') {
    const resolutionMultiplierEntries =
      getVideoResolutionMultiplierEntries(model);
    const rows = [
      {
        key: 'VideoBillingMode',
        label: 'VideoBillingMode',
        value: resolutionMultiplierEntries.length > 0 ? 'matrix' : 'per_second',
      },
      {
        key: 'VideoBasePrice',
        label: 'VideoBasePrice',
        value: hasValue(model.videoSecondPrice)
          ? model.videoSecondPrice
          : t('空'),
      },
    ];
    if (resolutionMultiplierEntries.length > 0) {
      rows.push({
        key: 'VideoResolutionMultipliers',
        label: 'VideoResolutionMultipliers',
        value: resolutionMultiplierEntries
          .map(([resolution, multiplier]) => `${resolution}: ${multiplier}`)
          .join(', '),
      });
    }
    return rows;
  }

  const finalBillingExpr = combineBillingExpr(
    model.billingExpr,
    model.requestRuleExpr,
  );

  if (model.billingMode === 'tiered_expr') {
    const rows = [
      {
        key: 'BillingMode',
        label: 'ModelBillingMode',
        value: 'tiered_expr',
      },
    ];
    if (finalBillingExpr) {
      const tierCount = (model.billingExpr.match(/tier\(/g) || []).length;
      rows.push({
        key: 'BillingExpr',
        label: 'ModelBillingExpr',
        value:
          tierCount > 0
            ? `${tierCount} ${t('档')} — ${
                finalBillingExpr.length > 60
                  ? finalBillingExpr.slice(0, 60) + '...'
                  : finalBillingExpr
              }`
            : finalBillingExpr.length > 60
              ? finalBillingExpr.slice(0, 60) + '...'
              : finalBillingExpr,
      });
    }
    return rows;
  }

  if (model.billingMode === 'per-request') {
    const rows = [
      {
        key: 'ModelPrice',
        label: 'ModelPrice',
        value: hasValue(model.fixedPrice) ? model.fixedPrice : t('空'),
      },
    ];
    return rows;
  }

  const inputPrice = toNumberOrNull(model.inputPrice);
  if (inputPrice === null) {
    const rows = [
      {
        key: 'ModelRatio',
        label: 'ModelRatio',
        value: hasValue(model.rawRatios.modelRatio)
          ? model.rawRatios.modelRatio
          : t('空'),
      },
      {
        key: 'CompletionRatio',
        label: 'CompletionRatio',
        value: hasValue(model.rawRatios.completionRatio)
          ? model.rawRatios.completionRatio
          : t('空'),
      },
      {
        key: 'CacheRatio',
        label: 'CacheRatio',
        value: hasValue(model.rawRatios.cacheRatio)
          ? model.rawRatios.cacheRatio
          : t('空'),
      },
      {
        key: 'CreateCacheRatio',
        label: 'CreateCacheRatio',
        value: hasValue(model.rawRatios.createCacheRatio)
          ? model.rawRatios.createCacheRatio
          : t('空'),
      },
      {
        key: 'ImageRatio',
        label: 'ImageRatio',
        value: hasValue(model.rawRatios.imageRatio)
          ? model.rawRatios.imageRatio
          : t('空'),
      },
      {
        key: 'AudioRatio',
        label: 'AudioRatio',
        value: hasValue(model.rawRatios.audioRatio)
          ? model.rawRatios.audioRatio
          : t('空'),
      },
      {
        key: 'AudioCompletionRatio',
        label: 'AudioCompletionRatio',
        value: hasValue(model.rawRatios.audioCompletionRatio)
          ? model.rawRatios.audioCompletionRatio
          : t('空'),
      },
    ];
    return rows;
  }

  const completionPrice = toNumberOrNull(model.completionPrice);
  const cachePrice = toNumberOrNull(model.cachePrice);
  const createCachePrice = toNumberOrNull(model.createCachePrice);
  const imagePrice = toNumberOrNull(model.imagePrice);
  const audioInputPrice = toNumberOrNull(model.audioInputPrice);
  const audioOutputPrice = toNumberOrNull(model.audioOutputPrice);

  const rows = [
    {
      key: 'ModelRatio',
      label: 'ModelRatio',
      value: formatNumber(inputPrice / 2),
    },
    {
      key: 'CompletionRatio',
      label: 'CompletionRatio',
      value: model.completionRatioLocked
        ? `${model.lockedCompletionRatio || t('空')} (${t('后端固定')})`
        : completionPrice !== null
          ? formatNumber(completionPrice / inputPrice)
          : t('空'),
    },
    {
      key: 'CacheRatio',
      label: 'CacheRatio',
      value:
        cachePrice !== null ? formatNumber(cachePrice / inputPrice) : t('空'),
    },
    {
      key: 'CreateCacheRatio',
      label: 'CreateCacheRatio',
      value:
        createCachePrice !== null
          ? formatNumber(createCachePrice / inputPrice)
          : t('空'),
    },
    {
      key: 'ImageRatio',
      label: 'ImageRatio',
      value:
        imagePrice !== null ? formatNumber(imagePrice / inputPrice) : t('空'),
    },
    {
      key: 'AudioRatio',
      label: 'AudioRatio',
      value:
        audioInputPrice !== null
          ? formatNumber(audioInputPrice / inputPrice)
          : t('空'),
    },
    {
      key: 'AudioCompletionRatio',
      label: 'AudioCompletionRatio',
      value:
        audioOutputPrice !== null &&
        audioInputPrice !== null &&
        audioInputPrice !== 0
          ? formatNumber(audioOutputPrice / audioInputPrice)
          : t('空'),
    },
  ];
  return rows;
};
