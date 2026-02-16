/**
 * Utility functions for formatting financial and numeric data
 */

/**
 * Format ISK values with K, M, B, T suffixes
 */
export function formatISK(value: number, decimals: number = 2): string {
  if (value === 0) return '0 ISK';

  const absValue = Math.abs(value);
  const sign = value < 0 ? '-' : '';

  if (absValue >= 1e12) {
    return `${sign}${(absValue / 1e12).toFixed(decimals)}T ISK`;
  } else if (absValue >= 1e9) {
    return `${sign}${(absValue / 1e9).toFixed(decimals)}B ISK`;
  } else if (absValue >= 1e6) {
    return `${sign}${(absValue / 1e6).toFixed(decimals)}M ISK`;
  } else if (absValue >= 1e3) {
    return `${sign}${(absValue / 1e3).toFixed(decimals)}K ISK`;
  } else {
    return `${sign}${absValue.toLocaleString()} ISK`;
  }
}

/**
 * Format ISK with commas (no suffix)
 */
export function formatISKDetailed(value: number): string {
  return `${value.toLocaleString()} ISK`;
}

/**
 * Format quantity with K, M, B suffixes
 */
export function formatQuantity(value: number, decimals: number = 1): string {
  if (value === 0) return '0';

  const absValue = Math.abs(value);
  const sign = value < 0 ? '-' : '';

  if (absValue >= 1e9) {
    return `${sign}${(absValue / 1e9).toFixed(decimals)}B`;
  } else if (absValue >= 1e6) {
    return `${sign}${(absValue / 1e6).toFixed(decimals)}M`;
  } else if (absValue >= 1e3) {
    return `${sign}${(absValue / 1e3).toFixed(decimals)}K`;
  } else {
    return `${sign}${absValue.toLocaleString()}`;
  }
}

/**
 * Format percentage
 */
export function formatPercent(value: number, decimals: number = 2): string {
  return `${value >= 0 ? '+' : ''}${value.toFixed(decimals)}%`;
}

/**
 * Get color based on value (for profit/loss, etc.)
 */
export function getValueColor(value: number): string {
  if (value > 0) return '#10b981'; // green
  if (value < 0) return '#ef4444'; // red
  return '#94a3b8'; // gray
}

/**
 * Get trend indicator
 */
export function getTrendIndicator(value: number): string {
  if (value > 0) return '↑';
  if (value < 0) return '↓';
  return '→';
}

/**
 * Format number with commas
 */
export function formatNumber(value: number, decimals: number = 0): string {
  return value.toLocaleString(undefined, {
    minimumFractionDigits: decimals,
    maximumFractionDigits: decimals,
  });
}

/**
 * Format compact number (no decimals, with suffixes)
 */
export function formatCompact(value: number): string {
  const absValue = Math.abs(value);
  const sign = value < 0 ? '-' : '';

  if (absValue >= 1e12) return `${sign}${(absValue / 1e12).toFixed(1)}T`;
  if (absValue >= 1e9) return `${sign}${(absValue / 1e9).toFixed(1)}B`;
  if (absValue >= 1e6) return `${sign}${(absValue / 1e6).toFixed(1)}M`;
  if (absValue >= 1e3) return `${sign}${(absValue / 1e3).toFixed(1)}K`;
  return `${sign}${absValue}`;
}
