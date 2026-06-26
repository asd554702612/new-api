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

const DEFAULT_USD_EXCHANGE_RATE = 7;

function getUsdExchangeRate(storage) {
  const fallback = DEFAULT_USD_EXCHANGE_RATE;
  if (!storage) {
    return fallback;
  }

  try {
    const status = JSON.parse(storage.getItem('status') || '{}');
    const rate = Number(status?.usd_exchange_rate);
    return rate > 0 ? rate : fallback;
  } catch {
    return fallback;
  }
}

export function formatPaymentOrderMoney(
  value,
  order = {},
  storage = globalThis.localStorage,
) {
  const amount = Number(value || 0);
  const orderType = String(order?.order_type || '').trim();
  const payCurrency = String(order?.pay_currency || '')
    .trim()
    .toUpperCase();
  const shouldConvertUsdSubscription =
    orderType === 'subscription' && payCurrency === 'USD';
  const displayAmount = shouldConvertUsdSubscription
    ? amount * getUsdExchangeRate(storage)
    : amount;

  return `¥${displayAmount.toFixed(2)}`;
}
