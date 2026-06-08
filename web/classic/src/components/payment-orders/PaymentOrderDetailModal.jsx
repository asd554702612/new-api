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
import { Descriptions, Modal, Space, Table, Typography } from '@douyinfe/semi-ui';
import { timestamp2string } from '../../helpers';
import PaymentOrderStatusTag from './PaymentOrderStatusTag';

const { Text } = Typography;

const formatMoney = (value) => `¥${Number(value || 0).toFixed(2)}`;

const PaymentOrderDetailModal = ({ visible, detail, loading, onCancel, t }) => {
  const order = detail?.order;
  const auditLogs = detail?.audit_logs || [];
  const refunds = detail?.refunds || [];

  const data = order
    ? [
        { key: t('订单号'), value: <Text copyable>{order.trade_no}</Text> },
        { key: t('订单类型'), value: t(order.order_type === 'subscription' ? '订阅订单' : '充值订单') },
        { key: t('用户ID'), value: order.user_id },
        { key: t('支付方式'), value: order.payment_method || order.payment_provider || '-' },
        { key: t('状态'), value: <PaymentOrderStatusTag status={order.status} t={t} /> },
        { key: t('支付金额'), value: formatMoney(order.pay_amount) },
        { key: t('站内额度'), value: order.amount || '-' },
        { key: t('订阅套餐'), value: order.plan_title || '-' },
        { key: t('创建时间'), value: order.created_time ? timestamp2string(order.created_time) : '-' },
        { key: t('完成时间'), value: order.complete_time ? timestamp2string(order.complete_time) : '-' },
      ]
    : [];

  return (
    <Modal
      title={t('订单详情')}
      visible={visible}
      onCancel={onCancel}
      footer={null}
      width={760}
    >
      <Space vertical align='stretch' className='w-full'>
        <Descriptions loading={loading} data={data} row />
        <div>
          <div className='mb-2 font-medium'>{t('退款记录')}</div>
          <Table
            size='small'
            pagination={false}
            dataSource={refunds}
            rowKey='id'
            columns={[
              { title: t('金额'), dataIndex: 'amount', render: formatMoney },
              { title: t('原因'), dataIndex: 'reason' },
              { title: t('状态'), dataIndex: 'status' },
              { title: t('申请人'), dataIndex: 'requested_by' },
              {
                title: t('时间'),
                dataIndex: 'create_time',
                render: (value) => (value ? timestamp2string(value) : '-'),
              },
            ]}
          />
        </div>
        <div>
          <div className='mb-2 font-medium'>{t('操作日志')}</div>
          <Table
            size='small'
            pagination={false}
            dataSource={auditLogs}
            rowKey='id'
            columns={[
              { title: t('动作'), dataIndex: 'action' },
              { title: t('详情'), dataIndex: 'detail' },
              { title: t('操作人'), dataIndex: 'operator' },
              {
                title: t('时间'),
                dataIndex: 'create_time',
                render: (value) => (value ? timestamp2string(value) : '-'),
              },
            ]}
          />
        </div>
      </Space>
    </Modal>
  );
};

export default PaymentOrderDetailModal;
