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

export function normalizeRankingsModule(moduleConfig) {
  if (typeof moduleConfig === 'boolean') {
    return {
      enabled: moduleConfig,
      requireAuth: true,
    };
  }

  if (moduleConfig && typeof moduleConfig === 'object') {
    return {
      enabled: moduleConfig.enabled !== false,
      requireAuth: moduleConfig.requireAuth === true,
    };
  }

  return {
    enabled: true,
    requireAuth: true,
  };
}

export function buildHeaderNavModulesWithRankings(modules) {
  return {
    ...modules,
    rankings: normalizeRankingsModule(modules?.rankings),
  };
}

export function formatCompactNumber(value) {
  const num = Number(value || 0);
  if (!Number.isFinite(num)) return '0';
  const abs = Math.abs(num);
  if (abs >= 1_000_000_000) return `${trimFixed(num / 1_000_000_000)}B`;
  if (abs >= 1_000_000) return `${trimFixed(num / 1_000_000)}M`;
  if (abs >= 1_000) return `${trimFixed(num / 1_000)}K`;
  return Math.round(num).toLocaleString();
}

function trimFixed(value) {
  return value.toFixed(1).replace(/\.0$/, '');
}
