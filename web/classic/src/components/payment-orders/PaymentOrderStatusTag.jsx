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

import React from 'react';
import { Tag } from '@douyinfe/semi-ui';

export const ORDER_STATUS_META = {
  PENDING: { color: 'orange', label: '待支付' },
  COMPLETED: { color: 'green', label: '已完成' },
  FAILED: { color: 'red', label: '失败' },
  EXPIRED: { color: 'grey', label: '已过期' },
  CANCELLED: { color: 'grey', label: '已取消' },
  REFUND_REQUESTED: { color: 'blue', label: '退款申请中' },
  REFUNDED: { color: 'purple', label: '已登记退款' },
  REFUND_FAILED: { color: 'red', label: '退款失败' },
};

const PaymentOrderStatusTag = ({ status, t }) => {
  const meta = ORDER_STATUS_META[status] || {
    color: 'white',
    label: status || '-',
  };
  return (
    <Tag color={meta.color} shape='circle' size='small'>
      {t ? t(meta.label) : meta.label}
    </Tag>
  );
};

export default PaymentOrderStatusTag;
