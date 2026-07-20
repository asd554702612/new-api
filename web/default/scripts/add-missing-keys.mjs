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
import fs from 'node:fs/promises'
import path from 'node:path'

const LOCALES_DIR = path.resolve('src/i18n/locales')

function stableStringify(obj) {
  return JSON.stringify(obj, null, 2) + '\n'
}

const newKeys = {
  en: {
    'footer.companyCopyright': '© 2018 成都格品科技有限公司版权所有',
    'footer.icpNumber': '蜀ICP备17044249号-1',
  },
  zh: {
    'footer.companyCopyright': '© 2018 成都格品科技有限公司版权所有',
    'footer.icpNumber': '蜀ICP备17044249号-1',
  },
  fr: {
    'footer.companyCopyright': '© 2018 成都格品科技有限公司版权所有',
    'footer.icpNumber': '蜀ICP备17044249号-1',
  },
  ja: {
    'footer.companyCopyright': '© 2018 成都格品科技有限公司版权所有',
    'footer.icpNumber': '蜀ICP备17044249号-1',
  },
  ru: {
    'footer.companyCopyright': '© 2018 成都格品科技有限公司版权所有',
    'footer.icpNumber': '蜀ICP备17044249号-1',
  },
  vi: {
    'footer.companyCopyright': '© 2018 成都格品科技有限公司版权所有',
    'footer.icpNumber': '蜀ICP备17044249号-1',
  },
}

async function main() {
  let totalAdded = 0

  for (const [locale, trans] of Object.entries(newKeys)) {
    const filePath = path.join(LOCALES_DIR, `${locale}.json`)
    const json = JSON.parse(await fs.readFile(filePath, 'utf8'))

    let count = 0
    for (const [key, value] of Object.entries(trans)) {
      if (!Object.prototype.hasOwnProperty.call(json.translation, key)) {
        json.translation[key] = value
        count++
      } else if (json.translation[key] !== value) {
        json.translation[key] = value
        count++
      }
    }

    if (count > 0) {
      json.translation = Object.fromEntries(
        Object.entries(json.translation).sort(([a], [b]) => a.localeCompare(b))
      )
      await fs.writeFile(filePath, stableStringify(json), 'utf8')
    }

    console.log(`${locale}: ${count} translations applied`)
    totalAdded += count
  }

  console.log(`\nTotal: ${totalAdded} translations applied`)
}

main().catch((err) => {
  console.error(err)
  process.exitCode = 1
})
