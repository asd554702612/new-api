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
import assert from 'node:assert/strict'
import { describe, test } from 'node:test'
import { buildCCSwitchURL } from './cc-switch'

function decodeUrlSafeBase64(value: string): string {
  let normalized = value.replace(/-/g, '+').replace(/_/g, '/')
  while (normalized.length % 4) {
    normalized += '='
  }
  return Buffer.from(normalized, 'base64').toString('utf8')
}

describe('buildCCSwitchURL', () => {
  test('adds token usage query configuration to imported provider', () => {
    const url = buildCCSwitchURL({
      app: 'claude',
      name: 'My Claude',
      models: { model: 'claude-sonnet-4' },
      apiKey: 'sk-test-key',
      serverAddress: 'https://api.example.com',
    })
    const parsed = new URL(url)
    const params = parsed.searchParams

    assert.equal(parsed.protocol, 'ccswitch:')
    assert.equal(params.get('endpoint'), 'https://api.example.com')
    assert.equal(params.get('usageEnabled'), 'true')
    assert.equal(params.get('usageBaseUrl'), 'https://api.example.com')
    assert.equal(params.get('usageApiKey'), 'sk-test-key')
    assert.equal(params.get('usageAutoInterval'), '30')

    const script = decodeUrlSafeBase64(params.get('usageScript') || '')
    assert.match(script, /url:\s*"\{\{baseUrl\}\}\/api\/usage\/token\/"/)
    assert.match(script, /"Authorization":\s*"Bearer \{\{apiKey\}\}"/)
    assert.match(script, /data\.total_available/)
    assert.match(script, /quotaPerUnit\s*=\s*Number\(data\.quota_per_unit/)
    assert.match(
      script,
      /remaining:\s*Number\(data\.total_available\s*\?\?\s*0\)\s*\/\s*quotaPerUnit/
    )
    assert.match(
      script,
      /used:\s*Number\(data\.total_used\s*\?\?\s*0\)\s*\/\s*quotaPerUnit/
    )
    assert.match(
      script,
      /total:\s*Number\(data\.total_granted\s*\?\?\s*0\)\s*\/\s*quotaPerUnit/
    )
    assert.match(script, /unit:\s*"USD"/)
    assert.match(script, /extra:\s*"Unlimited"/)
    assert.doesNotMatch(script, /Number\.MAX_SAFE_INTEGER/)
    assert.doesNotMatch(script, /unit:\s*"quota"/)
  })

  test('keeps Codex provider endpoint OpenAI compatible', () => {
    const url = buildCCSwitchURL({
      app: 'codex',
      name: 'My Codex',
      models: { model: 'gpt-5' },
      apiKey: 'sk-test-key',
      serverAddress: 'https://api.example.com',
    })
    const params = new URL(url).searchParams

    assert.equal(params.get('endpoint'), 'https://api.example.com/v1')
    assert.equal(params.get('usageBaseUrl'), 'https://api.example.com')
  })
})
