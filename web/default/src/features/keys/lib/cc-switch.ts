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
const CC_SWITCH_USAGE_AUTO_INTERVAL_MINUTES = '30'

type BuildCCSwitchUrlInput = {
  app: string
  name: string
  models: Record<string, string>
  apiKey: string
  serverAddress?: string
}

function encodeUrlSafeBase64(value: string): string {
  const bytes = new TextEncoder().encode(value)
  let binary = ''
  for (const byte of bytes) {
    binary += String.fromCharCode(byte)
  }
  return btoa(binary).replaceAll(/\+/g, '-').replaceAll(/\//g, '_').replace(/=+$/, '')
}

export function getServerAddress(): string {
  try {
    const raw = localStorage.getItem('status')
    if (raw) {
      const status = JSON.parse(raw)
      if (status.server_address) return status.server_address
    }
  } catch {
    /* empty */
  }
  return window.location.origin
}

export function buildCCSwitchUsageScript(): string {
  return `({
  request: {
    url: "{{baseUrl}}/api/usage/token/",
    method: "GET",
    headers: { "Authorization": "Bearer {{apiKey}}" }
  },
  extractor: function(response) {
    const payload = typeof response === "string" ? JSON.parse(response) : response;
    const data = payload && payload.data;
    if (!payload || payload.code !== true || !data) {
      return {
        isValid: false,
        invalidMessage: (payload && payload.message) || "Query failed"
      };
    }
    if (data.unlimited_quota === true) {
      return {
        isValid: true,
        unlimited: true,
        extra: "Unlimited"
      };
    }
    let quotaPerUnit = Number(data.quota_per_unit || 500000);
    if (!Number.isFinite(quotaPerUnit) || quotaPerUnit <= 0) {
      quotaPerUnit = 500000;
    }
    return {
      isValid: true,
      remaining: Number(data.total_available ?? 0) / quotaPerUnit,
      used: Number(data.total_used ?? 0) / quotaPerUnit,
      total: Number(data.total_granted ?? 0) / quotaPerUnit,
      unit: "USD"
    };
  }
})`
}

export function buildCCSwitchURL(input: BuildCCSwitchUrlInput): string {
  const serverAddress = input.serverAddress ?? getServerAddress()
  const endpoint = input.app === 'codex' ? `${serverAddress  }/v1` : serverAddress
  const params = new URLSearchParams()
  params.set('resource', 'provider')
  params.set('app', input.app)
  params.set('name', input.name)
  params.set('endpoint', endpoint)
  params.set('apiKey', input.apiKey)
  for (const [k, v] of Object.entries(input.models)) {
    if (v) params.set(k, v)
  }
  params.set('homepage', serverAddress)
  params.set('enabled', 'true')
  params.set('usageEnabled', 'true')
  params.set('usageBaseUrl', serverAddress)
  params.set('usageApiKey', input.apiKey)
  params.set('usageAutoInterval', CC_SWITCH_USAGE_AUTO_INTERVAL_MINUTES)
  params.set('usageScript', encodeUrlSafeBase64(buildCCSwitchUsageScript()))
  return `ccswitch://v1/import?${params.toString()}`
}
