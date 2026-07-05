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

import React, { useEffect, useMemo, useState } from 'react';
import {
  Avatar,
  Button,
  Card,
  Empty,
  Input,
  Modal,
  Select,
  Space,
  Table,
  Tag,
  TextArea,
  Typography,
} from '@douyinfe/semi-ui';
import { ShieldCheck } from 'lucide-react';
import dayjs from 'dayjs';
import { API, showError, showSuccess } from '../../../../helpers';

const { Text } = Typography;

const REQUEST_TYPES = ['access', 'correction', 'deletion'];
const CANCELLABLE_STATUSES = ['pending', 'processing'];

const unwrapApiData = (res) => {
  const body = res?.data;
  return body?.data ?? body;
};

const normalizeList = (payload) => {
  if (Array.isArray(payload)) return payload;
  if (Array.isArray(payload?.items)) return payload.items;
  if (Array.isArray(payload?.requests)) return payload.requests;
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

const getRequestId = (record) => record?.id ?? record?.request_id;

const getRequestTypeLabel = (t, type) => {
  const labels = {
    access: t('查阅复制'),
    correction: t('更正补充'),
    deletion: t('删除注销'),
  };
  return labels[type] || type || '-';
};

const getStatusLabel = (t, status) => {
  const labels = {
    pending: t('待处理'),
    processing: t('处理中'),
    completed: t('已完成'),
    rejected: t('已驳回'),
    cancelled: t('已取消'),
  };
  return labels[status] || status || '-';
};

const getStatusColor = (status) => {
  const colors = {
    pending: 'orange',
    processing: 'blue',
    completed: 'green',
    rejected: 'red',
    cancelled: 'grey',
  };
  return colors[status] || 'grey';
};

const SnapshotItem = ({ label, value }) => (
  <div className='rounded-xl border border-gray-100 bg-gray-50 px-3 py-2'>
    <div className='text-xs text-gray-500 mb-1'>{label}</div>
    <div className='text-sm font-medium text-gray-900 break-all'>
      {value || '-'}
    </div>
  </div>
);

const PersonalInfoRightsCard = ({ t }) => {
  const [snapshot, setSnapshot] = useState({});
  const [requests, setRequests] = useState([]);
  const [loading, setLoading] = useState(false);
  const [requestsLoading, setRequestsLoading] = useState(false);
  const [submitLoading, setSubmitLoading] = useState(false);
  const [requestType, setRequestType] = useState('access');
  const [contactName, setContactName] = useState('');
  const [contactEmail, setContactEmail] = useState('');
  const [contactPhone, setContactPhone] = useState('');
  const [description, setDescription] = useState('');

  const requestTypeOptions = useMemo(
    () =>
      REQUEST_TYPES.map((type) => ({
        label: getRequestTypeLabel(t, type),
        value: type,
      })),
    [t],
  );

  const snapshotItems = useMemo(() => {
    const data =
      snapshot?.snapshot || snapshot?.personal_info || snapshot || {};
    return [
      [t('用户 ID'), data.user_id ?? data.id],
      [t('用户名'), data.username],
      [t('显示名称'), data.display_name ?? data.displayName],
      [t('邮箱'), data.email],
      [t('手机号'), data.phone_number ?? data.phone],
      [t('用户分组'), data.group],
      [t('账户状态'), data.status],
      [t('创建时间'), formatDate(data.created_at ?? data.createdAt)],
    ];
  }, [snapshot, t]);

  const loadSnapshot = async () => {
    setLoading(true);
    try {
      const res = await API.get('/api/privacy/personal-info');
      const { success, message } = res.data || {};
      if (success === false) {
        showError(message || t('个人信息快照加载失败'));
        return;
      }
      const payload = unwrapApiData(res) || {};
      const data = payload?.snapshot || payload?.personal_info || payload;
      setSnapshot(payload);
      setContactName(
        (prev) => prev || data.display_name || data.username || '',
      );
      setContactEmail((prev) => prev || data.email || '');
      setContactPhone((prev) => prev || data.phone_number || data.phone || '');
    } catch (error) {
      showError(error?.message || t('个人信息快照加载失败'));
    } finally {
      setLoading(false);
    }
  };

  const loadRequests = async () => {
    setRequestsLoading(true);
    try {
      const res = await API.get('/api/privacy/requests');
      const { success, message } = res.data || {};
      if (success === false) {
        showError(message || t('申请记录加载失败'));
        return;
      }
      setRequests(normalizeList(unwrapApiData(res)));
    } catch (error) {
      showError(error?.message || t('申请记录加载失败'));
    } finally {
      setRequestsLoading(false);
    }
  };

  useEffect(() => {
    loadSnapshot();
    loadRequests();
  }, []);

  const submitRequest = async () => {
    const trimmedDescription = description.trim();
    if (!trimmedDescription) {
      showError(t('请填写申请说明'));
      return;
    }

    setSubmitLoading(true);
    try {
      const res = await API.post('/api/privacy/requests', {
        request_type: requestType,
        contact_name: contactName.trim(),
        contact_email: contactEmail.trim(),
        contact_phone: contactPhone.trim(),
        content: trimmedDescription,
      });
      const { success, message } = res.data || {};
      if (success === false) {
        showError(message || t('申请提交失败'));
        return;
      }
      showSuccess(t('申请已提交'));
      setDescription('');
      setRequestType('access');
      await loadRequests();
    } catch (error) {
      showError(error?.message || t('申请提交失败'));
    } finally {
      setSubmitLoading(false);
    }
  };

  const cancelRequest = async (record) => {
    const requestId = getRequestId(record);
    if (!requestId) {
      showError(t('无法识别申请记录'));
      return;
    }

    try {
      const res = await API.post(
        `/api/privacy/requests/${requestId}/cancel`,
        {},
        { skipErrorHandler: true },
      );
      if (res.data?.success === false) {
        throw new Error(res.data?.message || t('取消失败，请重试'));
      }
      showSuccess(t('申请已取消'));
      await loadRequests();
    } catch (error) {
      showError(
        error?.response?.data?.message ||
          error?.message ||
          t('取消失败，请重试'),
      );
    }
  };

  const confirmCancelRequest = (record) => {
    Modal.confirm({
      title: t('取消申请'),
      content: t('确认取消这条个人信息权利申请吗？'),
      okText: t('确认取消'),
      cancelText: t('返回'),
      onOk: () => cancelRequest(record),
    });
  };

  const columns = [
    {
      title: t('类型'),
      dataIndex: 'request_type',
      render: (value) => (
        <Tag color='blue'>{getRequestTypeLabel(t, value)}</Tag>
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
      title: t('申请说明'),
      dataIndex: 'content',
      render: (value) => <span className='line-clamp-2'>{value || '-'}</span>,
    },
    {
      title: t('提交时间'),
      dataIndex: 'created_at',
      render: (value, record) => formatDate(value || record.createdAt),
    },
    {
      title: t('操作'),
      dataIndex: 'operate',
      render: (_, record) =>
        CANCELLABLE_STATUSES.includes(record.status) ? (
          <Button
            size='small'
            type='tertiary'
            onClick={() => confirmCancelRequest(record)}
          >
            {t('取消')}
          </Button>
        ) : (
          <Text type='tertiary'>-</Text>
        ),
    },
  ];

  return (
    <Card className='!rounded-2xl shadow-sm border-0'>
      <div className='flex items-center justify-between gap-3 mb-4'>
        <div className='flex items-center min-w-0'>
          <Avatar size='small' color='green' className='mr-3 shadow-md'>
            <ShieldCheck size={16} />
          </Avatar>
          <div className='min-w-0'>
            <Typography.Text className='text-lg font-medium'>
              {t('个人信息权利')}
            </Typography.Text>
            <div className='text-xs text-gray-600'>
              {t('查看个人信息快照，并提交查阅、更正或删除申请')}
            </div>
          </div>
        </div>
        <Button
          type='tertiary'
          size='small'
          loading={loading || requestsLoading}
          onClick={() => {
            loadSnapshot();
            loadRequests();
          }}
        >
          {t('刷新')}
        </Button>
      </div>

      <div className='grid grid-cols-1 sm:grid-cols-2 xl:grid-cols-4 gap-3 mb-4'>
        {snapshotItems.map(([label, value]) => (
          <SnapshotItem key={label} label={label} value={value} />
        ))}
      </div>

      <div className='rounded-xl border border-gray-100 p-4 mb-4'>
        <div className='font-semibold text-sm text-gray-900 mb-3'>
          {t('提交权利申请')}
        </div>
        <Space vertical align='start' className='w-full'>
          <Select
            value={requestType}
            optionList={requestTypeOptions}
            onChange={setRequestType}
            style={{ width: 220 }}
          />
          <div className='grid grid-cols-1 md:grid-cols-3 gap-3 w-full'>
            <Input
              value={contactName}
              onChange={setContactName}
              placeholder={t('请输入联系人')}
            />
            <Input
              value={contactEmail}
              onChange={setContactEmail}
              placeholder={t('请输入邮箱')}
            />
            <Input
              value={contactPhone}
              onChange={setContactPhone}
              placeholder={t('请输入手机号')}
            />
          </div>
          <TextArea
            value={description}
            onChange={setDescription}
            autosize={{ minRows: 3, maxRows: 6 }}
            placeholder={t('请说明申请原因和需要处理的个人信息范围')}
            style={{ width: '100%' }}
          />
          <Button
            type='primary'
            loading={submitLoading}
            onClick={submitRequest}
          >
            {t('提交申请')}
          </Button>
        </Space>
      </div>

      <div className='font-semibold text-sm text-gray-900 mb-3'>
        {t('我的申请')}
      </div>
      <Table
        size='small'
        rowKey={(record, index) => getRequestId(record) || index}
        columns={columns}
        dataSource={requests}
        loading={requestsLoading}
        pagination={false}
        empty={<Empty description={t('暂无申请记录')} />}
      />
    </Card>
  );
};

export default PersonalInfoRightsCard;
