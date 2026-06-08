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
import { Empty, Modal, Spin, Tag, Typography } from '@douyinfe/semi-ui';
import {
  IllustrationNoResult,
  IllustrationNoResultDark,
} from '@douyinfe/semi-illustrations';
import { ArrowDownToLine, Coins, ReceiptText } from 'lucide-react';
import { timestamp2string } from '../../../helpers';

const { Text } = Typography;

const ACTION_CONFIG = {
  accrue: {
    color: 'green',
    icon: <Coins size={14} />,
    key: '返现',
  },
  transfer: {
    color: 'blue',
    icon: <ArrowDownToLine size={14} />,
    key: '划转',
  },
  signup_bonus: {
    color: 'green',
    icon: <Coins size={14} />,
    key: '注册奖励',
  },
  withdraw_request: {
    color: 'orange',
    icon: <ReceiptText size={14} />,
    key: '提现申请',
  },
  withdraw_paid: {
    color: 'green',
    icon: <ReceiptText size={14} />,
    key: '提现已打款',
  },
  withdraw_reject: {
    color: 'red',
    icon: <ReceiptText size={14} />,
    key: '提现拒绝',
  },
  withdraw_fail: {
    color: 'red',
    icon: <ReceiptText size={14} />,
    key: '提现失败',
  },
};

const SOURCE_TYPE_LABELS = {
  topup: '充值',
  subscription: '订阅',
  manual: '手动',
};

const AffiliateRecordsModal = ({
  visible,
  onCancel,
  t,
  records = [],
  loading = false,
  renderQuota,
}) => {
  const renderRecordQuota = (quota) => {
    if (typeof renderQuota === 'function') {
      return renderQuota(quota || 0);
    }
    return quota || 0;
  };

  const renderSourceType = (sourceType) => {
    if (!sourceType) {
      return '-';
    }
    return t(SOURCE_TYPE_LABELS[sourceType] || sourceType);
  };

  const renderRecordMeta = (record) => {
    const meta = [
      {
        label: t('来源'),
        value: renderSourceType(record.source_order_type),
      },
      {
        label: t('支付方式'),
        value: record.payment_method || '-',
      },
      {
        label: t('当前待使用'),
        value: renderRecordQuota(record.balance_after || 0),
      },
      {
        label: t('累计收益'),
        value: renderRecordQuota(record.history_after || 0),
      },
    ];

    return meta.map((item) => (
      <div key={item.label} className='min-w-0'>
        <div className='text-xs text-gray-500'>{item.label}</div>
        <div className='text-sm mt-1 break-all'>{item.value}</div>
      </div>
    ));
  };

  return (
    <Modal
      title={t('邀请流水')}
      visible={visible}
      onCancel={onCancel}
      footer={null}
      centered
      size='large'
    >
      {loading ? (
        <div className='py-10 text-center'>
          <Spin />
          <div className='mt-3 text-gray-500'>{t('加载中')}</div>
        </div>
      ) : records.length === 0 ? (
        <Empty
          image={<IllustrationNoResult style={{ width: 150, height: 150 }} />}
          darkModeImage={
            <IllustrationNoResultDark style={{ width: 150, height: 150 }} />
          }
          description={t('暂无邀请流水')}
          style={{ padding: 30 }}
        />
      ) : (
        <div className='space-y-3 max-h-[520px] overflow-auto pr-1'>
          {records.map((record) => {
            const action = ACTION_CONFIG[record.action] || {
              color: 'grey',
              icon: <ReceiptText size={14} />,
              key: record.action || '未知',
            };
            const createdAt = record.created_at || record.create_time || 0;
            const sourceText =
              record.source_order_trade_no || record.remark || '-';

            return (
              <div
                key={record.id}
                className='border border-gray-100 rounded-lg p-4 bg-white'
              >
                <div className='flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2'>
                  <div className='flex items-center gap-2 min-w-0'>
                    <Tag
                      color={action.color}
                      shape='circle'
                      prefixIcon={action.icon}
                    >
                      {t(action.key)}
                    </Tag>
                    <Text className='break-all'>
                      {sourceText}
                    </Text>
                  </div>
                  <div className='flex items-center gap-3 shrink-0'>
                    <Text strong>{renderRecordQuota(record.quota || 0)}</Text>
                    <Text type='tertiary' size='small'>
                      {createdAt ? timestamp2string(createdAt) : '-'}
                    </Text>
                  </div>
                </div>

                <div className='grid grid-cols-2 lg:grid-cols-4 gap-3 mt-4'>
                  {renderRecordMeta(record)}
                </div>
              </div>
            );
          })}
        </div>
      )}
    </Modal>
  );
};

export default AffiliateRecordsModal;
