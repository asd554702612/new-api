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

import React, { useEffect, useState } from 'react';
import { Empty, Modal, Table, Tag, Typography } from '@douyinfe/semi-ui';
import {
  IllustrationNoResult,
  IllustrationNoResultDark,
} from '@douyinfe/semi-illustrations';
import { API, timestamp2string } from '../../../helpers';

const STATUS_COLORS = {
  pending_review: 'orange',
  approved: 'blue',
  paid: 'green',
  rejected: 'red',
  failed: 'red',
  cancelled: 'grey',
};

const PAGE_SIZE_OPTIONS = [10, 20, 50];
const PAGE_SIZE_OPTION_STRINGS = PAGE_SIZE_OPTIONS.map(String);
const TABLE_SCROLL_X = 'max(100%, 850px)';

const AffiliateWithdrawalsModal = ({ visible, onCancel, t, renderQuota }) => {
  const [loading, setLoading] = useState(false);
  const [records, setRecords] = useState([]);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [total, setTotal] = useState(0);

  const loadRecords = async () => {
    setLoading(true);
    try {
      const res = await API.get(
        `/api/user/aff/withdrawals?p=${page}&page_size=${pageSize}`,
      );
      if (res.data?.success) {
        setRecords(res.data.data?.items || []);
        setTotal(res.data.data?.total || 0);
      }
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (visible) {
      loadRecords();
    }
  }, [visible, page, pageSize]);

  const columns = [
    {
      title: t('提现额度'),
      dataIndex: 'quota',
      key: 'quota',
      width: 140,
      render: (value) => renderQuota(value || 0),
    },
    {
      title: t('状态'),
      dataIndex: 'status',
      key: 'status',
      width: 130,
      render: (value) => (
        <Tag color={STATUS_COLORS[value] || 'grey'}>{t(value || '-')}</Tag>
      ),
    },
    {
      title: t('收款说明'),
      dataIndex: 'payout_account_note',
      key: 'payout_account_note',
      width: 220,
      render: (value) => (
        <Typography.Text ellipsis={{ showTooltip: true }}>
          {value || '-'}
        </Typography.Text>
      ),
    },
    {
      title: t('打款流水号'),
      dataIndex: 'payout_trade_no',
      key: 'payout_trade_no',
      width: 180,
      render: (value) => value || '-',
    },
    {
      title: t('申请时间'),
      dataIndex: 'created_at',
      key: 'created_at',
      width: 180,
      render: (value) => (value ? timestamp2string(value) : '-'),
    },
  ];

  return (
    <Modal
      title={t('提现记录')}
      visible={visible}
      onCancel={onCancel}
      footer={null}
      centered
      size='large'
    >
      <Table
        columns={columns}
        dataSource={records}
        loading={loading}
        rowKey='id'
        pagination={{
          currentPage: page,
          pageSize,
          total,
          showSizeChanger: true,
          showQuickJumper: true,
          pageSizeOpts: PAGE_SIZE_OPTIONS,
          pageSizeOptions: PAGE_SIZE_OPTION_STRINGS,
          onPageChange: setPage,
          onPageSizeChange: (value) => {
            setPageSize(value);
            setPage(1);
          },
          onChange: (nextPage, nextPageSize) => {
            setPage(nextPage);
            if (nextPageSize && nextPageSize !== pageSize) {
              setPageSize(nextPageSize);
            }
          },
          onShowSizeChange: (current, nextPageSize) => {
            setPage(1);
            setPageSize(nextPageSize);
          },
        }}
        scroll={{ x: TABLE_SCROLL_X }}
        size='middle'
        className='overflow-hidden'
        empty={
          <Empty
            image={<IllustrationNoResult style={{ width: 150, height: 150 }} />}
            darkModeImage={
              <IllustrationNoResultDark style={{ width: 150, height: 150 }} />
            }
            description={t('暂无提现记录')}
            style={{ padding: 30 }}
          />
        }
      />
    </Modal>
  );
};

export default AffiliateWithdrawalsModal;
