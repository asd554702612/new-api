const DISPLAY_AMOUNT_DIGITS = 2;
const EPSILON = 1e-9;

export function formatLargeNumber(value) {
  const numeric = Number(value);
  if (!Number.isFinite(numeric)) {
    return String(value);
  }

  return new Intl.NumberFormat(undefined, {
    minimumFractionDigits: 0,
    maximumFractionDigits: DISPLAY_AMOUNT_DIGITS,
    useGrouping: false,
  }).format(numeric);
}

export function formatPayAmount(value) {
  const numeric = Number(value);
  if (!Number.isFinite(numeric)) {
    return '0.00';
  }

  return numeric.toFixed(DISPLAY_AMOUNT_DIGITS);
}

export function normalizeCurrencyDelta(value) {
  const numeric = Number(value);
  if (!Number.isFinite(numeric) || numeric <= EPSILON) {
    return 0;
  }

  return Number(numeric.toFixed(DISPLAY_AMOUNT_DIGITS));
}
