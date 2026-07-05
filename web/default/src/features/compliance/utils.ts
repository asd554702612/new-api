/*
Copyright (C) 2023-2026 QuantumNous

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
import dayjs from '@/lib/dayjs'
import type { PagePayload } from './types'

export function normalizePagePayload<T>(
  payload: PagePayload<T> | T[] | undefined
): Required<PagePayload<T>> {
  if (Array.isArray(payload)) {
    return {
      page: 1,
      page_size: payload.length,
      total: payload.length,
      items: payload,
    }
  }

  const items = Array.isArray(payload?.items) ? payload.items : []
  return {
    page: payload?.page ?? 1,
    page_size: payload?.page_size ?? items.length,
    total: payload?.total ?? items.length,
    items,
  }
}

export function formatComplianceTimestamp(value?: number | string): string {
  if (!value) return '-'
  const normalized =
    typeof value === 'number' && value > 0 && value < 10000000000
      ? value * 1000
      : value
  const parsed = dayjs(normalized)
  return parsed.isValid() ? parsed.format('YYYY-MM-DD HH:mm') : String(value)
}
