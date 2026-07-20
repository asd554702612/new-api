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
import {
  Avatar,
  Button,
  Card,
  Empty,
  Modal,
  Table,
  Tag,
  Typography,
} from '@douyinfe/semi-ui';
import { Eye, MessageSquare } from 'lucide-react';
import dayjs from 'dayjs';
import { API, showError } from '../../../../helpers';

const unwrapApiData = (res) => {
  const body = res?.data;
  return body?.data ?? body;
};

const normalizeList = (payload) => {
  if (Array.isArray(payload)) return payload;
  if (Array.isArray(payload?.items)) return payload.items;
  if (Array.isArray(payload?.feedbacks)) return payload.feedbacks;
  if (Array.isArray(payload?.list)) return payload.list;
  return [];
};

const formatDate = (value) => {
  if (!value) return '-';
  const normalized =
    typeof value === 'number' && value < 10000000000 ? value * 1000 : value;
  const parsed = dayjs(normalized);
  return parsed.isValid() ? parsed.format('YYYY-MM-DD HH:mm') : String(value);
};

const getFeedbackTypeLabel = (t, type) => {
  const labels = {
    complaint: t('投诉'),
    feedback: t('反馈'),
    other: t('其他'),
  };
  return labels[type] || type || '-';
};

const getStatusLabel = (t, status) => {
  const labels = {
    pending: t('待处理'),
    processing: t('处理中'),
    resolved: t('已解决'),
    closed: t('已关闭'),
    rejected: t('已驳回'),
  };
  return labels[status] || status || '-';
};

const getStatusColor = (status) => {
  const colors = {
    pending: 'orange',
    processing: 'blue',
    resolved: 'green',
    closed: 'grey',
    rejected: 'red',
  };
  return colors[status] || 'grey';
};

const DetailItem = ({ label, value }) => (
  <div className='border-b border-dashed border-gray-200 py-2 last:border-b-0'>
    <div className='text-xs text-gray-500 mb-1'>{label}</div>
    <div className='text-sm text-gray-900 break-all whitespace-pre-wrap'>
      {value || '-'}
    </div>
  </div>
);

const PublicFeedbackStatusCard = ({ t }) => {
  const [records, setRecords] = useState([]);
  const [loading, setLoading] = useState(false);
  const [detailVisible, setDetailVisible] = useState(false);
  const [detailLoading, setDetailLoading] = useState(false);
  const [currentRecord, setCurrentRecord] = useState(null);

  const loadFeedback = async () => {
    setLoading(true);
    try {
      const res = await API.get('/api/feedback/my', {
        params: { p: 1, page_size: 20 },
      });
      const { success, message } = res.data || {};
      if (success === false) {
        showError(message || t('投诉反馈记录加载失败'));
        return;
      }
      setRecords(normalizeList(unwrapApiData(res)));
    } catch (error) {
      showError(error?.message || t('投诉反馈记录加载失败'));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadFeedback();
  }, []);

  const openDetail = async (record) => {
    setCurrentRecord(record);
    setDetailVisible(true);
    setDetailLoading(true);
    try {
      const res = await API.get(`/api/feedback/my/${record.id}`, {
        skipErrorHandler: true,
      });
      if (res.data?.success === false) {
        throw new Error(res.data?.message || t('详情加载失败'));
      }
      const detail = unwrapApiData(res);
      if (detail && typeof detail === 'object') {
        setCurrentRecord(detail);
      }
    } catch (error) {
      showError(
        error?.response?.data?.message ||
          error?.message ||
          t('详情加载失败'),
      );
    } finally {
      setDetailLoading(false);
    }
  };

  const columns = [
    {
      title: t('追踪码'),
      dataIndex: 'tracking_code',
      render: (value) => <span className='break-all'>{value || '-'}</span>,
    },
    {
      title: t('类型'),
      dataIndex: 'feedback_type',
      render: (value) => (
        <Tag color='blue'>{getFeedbackTypeLabel(t, value)}</Tag>
      ),
    },
    {
      title: t('状态'),
      dataIndex: 'status',
      render: (value) => (
        <Tag color={getStatusColor(value)}>{getStatusLabel(t, value)}</Tag>
      ),
    },
    {
      title: t('标题'),
      dataIndex: 'title',
      render: (value) => <span className='line-clamp-2'>{value || '-'}</span>,
    },
    {
      title: t('处理时间'),
      dataIndex: 'handled_at',
      render: (value) => formatDate(value),
    },
    {
      title: t('处理说明'),
      dataIndex: 'admin_note',
      render: (value) => <span className='line-clamp-2'>{value || '-'}</span>,
    },
    {
      title: t('操作'),
      dataIndex: 'operate',
      render: (_, record) => (
        <Button
          size='small'
          type='tertiary'
          icon={<Eye size={14} />}
          onClick={() => openDetail(record)}
        >
          {t('详情')}
        </Button>
      ),
    },
  ];

  return (
    <Card className='!rounded-2xl shadow-sm border-0'>
      <div className='flex items-center justify-between gap-3 mb-4'>
        <div className='flex items-center min-w-0'>
          <Avatar size='small' color='blue' className='mr-3 shadow-md'>
            <MessageSquare size={16} />
          </Avatar>
          <div className='min-w-0'>
            <Typography.Text className='text-lg font-medium'>
              {t('我的投诉反馈')}
            </Typography.Text>
            <div className='text-xs text-gray-600'>
              {t('查看投诉反馈处理状态和处理说明')}
            </div>
          </div>
        </div>
        <Button type='tertiary' size='small' loading={loading} onClick={loadFeedback}>
          {t('刷新')}
        </Button>
      </div>

      <Table
        size='small'
        rowKey={(record, index) => record?.id || index}
        columns={columns}
        dataSource={records}
        loading={loading}
        pagination={false}
        empty={<Empty description={t('暂无投诉反馈记录')} />}
      />

      <Modal
        title={t('投诉反馈详情')}
        visible={detailVisible}
        onCancel={() => setDetailVisible(false)}
        footer={<Button onClick={() => setDetailVisible(false)}>{t('关闭')}</Button>}
        style={{ maxWidth: 720 }}
        width='90%'
      >
        {detailLoading && !currentRecord ? (
          <div className='py-8 text-center text-gray-500'>{t('加载中')}</div>
        ) : (
          <div className='grid grid-cols-1 md:grid-cols-2 gap-x-6'>
            <DetailItem label={t('追踪码')} value={currentRecord?.tracking_code} />
            <DetailItem
              label={t('类型')}
              value={getFeedbackTypeLabel(t, currentRecord?.feedback_type)}
            />
            <DetailItem
              label={t('状态')}
              value={getStatusLabel(t, currentRecord?.status)}
            />
            <DetailItem
              label={t('提交时间')}
              value={formatDate(currentRecord?.created_at)}
            />
            <DetailItem
              label={t('处理时间')}
              value={formatDate(currentRecord?.handled_at)}
            />
            <DetailItem label={t('标题')} value={currentRecord?.title} />
            <div className='md:col-span-2'>
              <DetailItem label={t('内容')} value={currentRecord?.content} />
            </div>
            <div className='md:col-span-2'>
              <DetailItem label={t('处理说明')} value={currentRecord?.admin_note} />
            </div>
          </div>
        )}
      </Modal>
    </Card>
  );
};

export default PublicFeedbackStatusCard;
