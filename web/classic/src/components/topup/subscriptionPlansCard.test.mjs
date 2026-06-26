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
import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const cardSource = fs.readFileSync(
  path.join(__dirname, 'SubscriptionPlansCard.jsx'),
  'utf8',
);
const modalSource = fs.readFileSync(
  path.join(__dirname, 'modals', 'SubscriptionPurchaseModal.jsx'),
  'utf8',
);
const rechargeSource = fs.readFileSync(
  path.join(__dirname, 'RechargeCard.jsx'),
  'utf8',
);

assert.match(cardSource, /userQuota/);
assert.match(cardSource, /\/api\/subscription\/balance\/pay/);
assert.match(cardSource, /reloadSubscriptionSelf\?\.\(\)/);
assert.match(modalSource, /allow_balance_pay/);
assert.match(modalSource, /余额支付/);
assert.match(rechargeSource, /userQuota=\{userState\?\.user\?\.quota/);
