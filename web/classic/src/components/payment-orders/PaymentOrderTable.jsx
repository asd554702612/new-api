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

import React, { useMemo } from 'react';
import { Button, Space, Table, Tag, Typography } from '@douyinfe/semi-ui';
import { timestamp2string } from '../../helpers';
import PaymentOrderStatusTag from './PaymentOrderStatusTag';
import { formatPaymentOrderMoney } from './paymentOrderMoney';

const { Text } = Typography;

const PaymentOrderTable = ({
  orders,
  loading,
  pagination,
  admin = false,
  onPageChange,
  onPageSizeChange,
  onDetail,
  onCancel,
  onRetry,
  onRefund,
  t,
}) => {
  const columns = useMemo(
    () => [
      ...(admin
        ? [
            {
              title: t('用户ID'),
              dataIndex: 'user_id',
              width: 90,
            },
          ]
        : []),
      {
        title: t('订单号'),
        dataIndex: 'trade_no',
        render: (tradeNo) => <Text copyable>{tradeNo}</Text>,
      },
      {
        title: t('订单类型'),
        dataIndex: 'order_type',
        width: 110,
        render: (value) => (
          <Tag
            color={value === 'subscription' ? 'purple' : 'blue'}
            shape='circle'
          >
            {t(value === 'subscription' ? '订阅订单' : '充值订单')}
          </Tag>
        ),
      },
      {
        title: t('支付方式'),
        dataIndex: 'payment_method',
        width: 140,
        render: (value, record) => value || record.payment_provider || '-',
      },
      {
        title: t('金额'),
        dataIndex: 'pay_amount',
        width: 110,
        render: (value, record) => formatPaymentOrderMoney(value, record),
      },
      {
        title: t('状态'),
        dataIndex: 'status',
        width: 130,
        render: (status) => <PaymentOrderStatusTag status={status} t={t} />,
      },
      {
        title: t('创建时间'),
        dataIndex: 'created_time',
        width: 180,
        render: (value) => (value ? timestamp2string(value) : '-'),
      },
      {
        title: t('操作'),
        dataIndex: 'operate',
        width: admin ? 250 : 170,
        render: (_, record) => (
          <Space wrap>
            <Button size='small' onClick={() => onDetail(record)}>
              {t('详情')}
            </Button>
            {record.status === 'PENDING' && (
              <Button
                size='small'
                type='warning'
                onClick={() => onCancel(record)}
              >
                {t('取消')}
              </Button>
            )}
            {admin && ['PENDING', 'FAILED'].includes(record.status) && (
              <Button
                size='small'
                type='primary'
                onClick={() => onRetry(record)}
              >
                {t('补发')}
              </Button>
            )}
            {['COMPLETED', 'REFUND_REQUESTED'].includes(record.status) && (
              <Button
                size='small'
                type='danger'
                onClick={() => onRefund(record)}
              >
                {admin ? t('退款') : t('申请退款')}
              </Button>
            )}
          </Space>
        ),
      },
    ],
    [admin, onCancel, onDetail, onRefund, onRetry, t],
  );

  return (
    <Table
      rowKey={(record) => `${record.order_type}-${record.id}`}
      loading={loading}
      columns={columns}
      dataSource={orders}
      pagination={{
        currentPage: pagination.page,
        pageSize: pagination.pageSize,
        total: pagination.total,
        showSizeChanger: true,
        onPageChange,
        onPageSizeChange,
      }}
    />
  );
};

export default PaymentOrderTable;
