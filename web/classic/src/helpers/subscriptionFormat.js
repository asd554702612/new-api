import { getCurrencyConfig } from './render';

export function formatSubscriptionDuration(plan, t) {
  const unit = plan?.duration_unit || 'month';
  const value = plan?.duration_value || 1;
  const unitLabels = {
    year: t('年'),
    month: t('个月'),
    day: t('天'),
    hour: t('小时'),
    custom: t('自定义'),
  };
  if (unit === 'custom') {
    const seconds = plan?.custom_seconds || 0;
    if (seconds >= 86400) return `${Math.floor(seconds / 86400)} ${t('天')}`;
    if (seconds >= 3600) return `${Math.floor(seconds / 3600)} ${t('小时')}`;
    return `${seconds} ${t('秒')}`;
  }
  return `${value} ${unitLabels[unit] || unit}`;
}

export function formatSubscriptionResetPeriod(plan, t) {
  const period = plan?.quota_reset_period || 'never';
  if (period === 'never') return t('不重置');
  if (period === 'daily') return t('每天');
  if (period === 'weekly') return t('每周');
  if (period === 'monthly') return t('每月');
  if (period === 'custom') {
    const seconds = Number(plan?.quota_reset_custom_seconds || 0);
    if (seconds >= 86400) return `${Math.floor(seconds / 86400)} ${t('天')}`;
    if (seconds >= 3600) return `${Math.floor(seconds / 3600)} ${t('小时')}`;
    if (seconds >= 60) return `${Math.floor(seconds / 60)} ${t('分钟')}`;
    return `${seconds} ${t('秒')}`;
  }
  return t('不重置');
}

export function formatSubscriptionPrice(plan, digits = 2) {
  const priceUSD = subscriptionPriceToUSD(plan);
  const { symbol, rate, type } = getCurrencyConfig();
  if (type === 'CNY' || type === 'CUSTOM') {
    return `${symbol}${formatSubscriptionMoney(priceUSD * (rate || 1), digits)}`;
  }
  return `$${formatSubscriptionMoney(priceUSD, digits)}`;
}

export function subscriptionPriceToDisplayAmount(plan) {
  const priceUSD = subscriptionPriceToUSD(plan);
  const { rate, type } = getCurrencyConfig();
  if (type === 'CNY' || type === 'CUSTOM') {
    return priceUSD * (rate || 1);
  }
  return priceUSD;
}

export function subscriptionDisplayAmountToUSD(amount) {
  const value = Number(amount || 0);
  if (!Number.isFinite(value) || value <= 0) return 0;
  const { rate, type } = getCurrencyConfig();
  if (type === 'CNY' || type === 'CUSTOM') {
    return value / (rate || 1);
  }
  return value;
}

function subscriptionPriceToUSD(plan) {
  const price = Number(plan?.price_amount || 0);
  if (!Number.isFinite(price) || price <= 0) return 0;
  const currency = String(plan?.currency || 'USD').trim().toUpperCase();
  if (currency === 'CNY') {
    return price / getUSDExchangeRate();
  }
  return price;
}

function getUSDExchangeRate() {
  try {
    const status = JSON.parse(localStorage.getItem('status') || '{}');
    const rate = Number(status?.usd_exchange_rate || 0);
    return Number.isFinite(rate) && rate > 0 ? rate : 1;
  } catch (e) {
    return 1;
  }
}

function formatSubscriptionMoney(amount, digits) {
  const value = Number(amount || 0);
  return value.toFixed(Number.isInteger(value) ? 0 : digits);
}
