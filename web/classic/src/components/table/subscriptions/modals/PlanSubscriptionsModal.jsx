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

import React, { useEffect, useMemo, useRef, useState } from 'react';
import {
  Button,
  Col,
  Empty,
  Form,
  Input,
  Modal,
  Row,
  Select,
  SideSheet,
  Space,
  Tag,
  Typography,
} from '@douyinfe/semi-ui';
import { IconRefresh, IconSearch } from '@douyinfe/semi-icons';
import {
  IllustrationNoResult,
  IllustrationNoResultDark,
} from '@douyinfe/semi-illustrations';
import { API, showError, showSuccess } from '../../../../helpers';
import { useIsMobile } from '../../../../hooks/common/useIsMobile';
import CardTable from '../../../common/ui/CardTable';

const { Text } = Typography;

function formatTs(ts) {
  if (!ts) return '-';
  return new Date(ts * 1000).toLocaleString();
}

function formatUsagePercent(used, total) {
  return `${((used / total) * 100).toFixed(2)}% (${used}/${total})`;
}

function formatDateTimeLocal(unixSeconds) {
  const value = Number(unixSeconds || 0);
  if (value <= 0) return '';
  const date = new Date(value * 1000);
  const pad = (num) => String(num).padStart(2, '0');
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(
    date.getDate(),
  )}T${pad(date.getHours())}:${pad(date.getMinutes())}`;
}

function parseDateTimeLocal(value) {
  if (!value) return 0;
  const timestamp = new Date(value).getTime();
  if (!Number.isFinite(timestamp)) return 0;
  return Math.floor(timestamp / 1000);
}

function renderStatusTag(sub, t) {
  const now = Date.now() / 1000;
  const end = sub?.end_time || 0;
  const status = sub?.status || '';

  if (status === 'active' && end > now) {
    return (
      <Tag color='green' shape='circle' size='small'>
        {t('生效')}
      </Tag>
    );
  }
  if (status === 'cancelled') {
    return (
      <Tag color='grey' shape='circle' size='small'>
        {t('已作废')}
      </Tag>
    );
  }
  return (
    <Tag color='grey' shape='circle' size='small'>
      {t('已过期')}
    </Tag>
  );
}

const DetailItem = ({ label, value }) => (
  <div className='min-w-0'>
    <div className='text-xs text-gray-500 mb-1'>{label}</div>
    <div className='text-sm break-all'>{value ?? '-'}</div>
  </div>
);

const PlanSubscriptionsModal = ({ visible, onCancel, planRecord, t }) => {
  const isMobile = useIsMobile();
  const editFormRef = useRef(null);
  const [loading, setLoading] = useState(false);
  const [records, setRecords] = useState([]);
  const [total, setTotal] = useState(0);
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [keyword, setKeyword] = useState('');
  const [status, setStatus] = useState('');
  const [plans, setPlans] = useState([]);
  const [plansLoading, setPlansLoading] = useState(false);
  const [detailRecord, setDetailRecord] = useState(null);
  const [editingRecord, setEditingRecord] = useState(null);
  const [saving, setSaving] = useState(false);

  const plan = planRecord?.plan;

  const statusOptions = useMemo(
    () => [
      { label: t('全部状态'), value: '' },
      { label: t('生效'), value: 'active' },
      { label: t('已过期'), value: 'expired' },
      { label: t('已作废'), value: 'cancelled' },
    ],
    [t],
  );

  const editStatusOptions = useMemo(
    () => [
      { label: t('生效'), value: 'active' },
      { label: t('已过期'), value: 'expired' },
      { label: t('已作废'), value: 'cancelled' },
    ],
    [t],
  );

  const planOptions = useMemo(
    () =>
      (plans || []).map((item) => ({
        label: `${item?.plan?.title || `#${item?.plan?.id || '-'}`} (ID: ${
          item?.plan?.id || '-'
        })`,
        value: item?.plan?.id,
      })),
    [plans],
  );

  const planTitleMap = useMemo(() => {
    const map = new Map();
    (plans || []).forEach((item) => {
      const id = item?.plan?.id;
      if (id) map.set(id, item?.plan?.title || `#${id}`);
    });
    if (plan?.id && !map.has(plan.id))
      map.set(plan.id, plan.title || `#${plan.id}`);
    return map;
  }, [plans, plan?.id, plan?.title]);

  const loadSubscriptions = async (
    page = currentPage,
    size = pageSize,
    nextStatus = status,
    nextKeyword = keyword,
  ) => {
    if (!plan?.id) return;
    setLoading(true);
    try {
      const params = new URLSearchParams();
      params.set('p', String(page));
      params.set('page_size', String(size));
      if (nextStatus) params.set('status', nextStatus);
      if (nextKeyword) params.set('keyword', nextKeyword.trim());
      const res = await API.get(
        `/api/subscription/admin/plans/${plan.id}/subscriptions?${params.toString()}`,
      );
      if (res.data?.success) {
        const data = res.data.data || {};
        setRecords(data.items || []);
        setTotal(Number(data.total || 0));
        setCurrentPage(Number(data.page || page));
        setPageSize(Number(data.page_size || size));
      } else {
        showError(res.data?.message || t('加载失败'));
      }
    } catch (e) {
      showError(t('请求失败'));
    } finally {
      setLoading(false);
    }
  };

  const loadPlans = async () => {
    setPlansLoading(true);
    try {
      const res = await API.get('/api/subscription/admin/plans');
      if (res.data?.success) {
        setPlans(res.data.data || []);
      } else {
        showError(res.data?.message || t('加载失败'));
      }
    } catch (e) {
      showError(t('请求失败'));
    } finally {
      setPlansLoading(false);
    }
  };

  useEffect(() => {
    if (!visible) return;
    setKeyword('');
    setStatus('');
    setCurrentPage(1);
    setPageSize(10);
    setDetailRecord(null);
    setEditingRecord(null);
    loadSubscriptions(1, 10, '', '');
    loadPlans();
  }, [visible, plan?.id]);

  const handleSearch = () => {
    setCurrentPage(1);
    loadSubscriptions(1, pageSize, status, keyword);
  };

  const handleReset = () => {
    setKeyword('');
    setStatus('');
    setCurrentPage(1);
    loadSubscriptions(1, pageSize, '', '');
  };

  const handleStatusChange = (value) => {
    setStatus(value || '');
    setCurrentPage(1);
    loadSubscriptions(1, pageSize, value || '', keyword);
  };

  const handlePageChange = (page) => {
    setCurrentPage(page);
    loadSubscriptions(page, pageSize, status, keyword);
  };

  const handlePageSizeChange = (size) => {
    setPageSize(size);
    setCurrentPage(1);
    loadSubscriptions(1, size, status, keyword);
  };

  const refreshCurrentPage = () => {
    loadSubscriptions(currentPage, pageSize, status, keyword);
  };

  const getEditInitialValues = () => {
    const sub = editingRecord?.subscription || {};
    return {
      plan_id: sub.plan_id,
      status: sub.status || 'active',
      start_time_local: formatDateTimeLocal(sub.start_time),
      end_time_local: formatDateTimeLocal(sub.end_time),
      amount_total: Number(sub.amount_total || 0),
      amount_used: Number(sub.amount_used || 0),
      next_reset_time_local: formatDateTimeLocal(sub.next_reset_time),
    };
  };

  const submitEdit = async (values) => {
    const subId = editingRecord?.subscription?.id;
    if (!subId) return;
    const payload = {
      plan_id: Number(values.plan_id || 0),
      status: values.status || '',
      start_time: parseDateTimeLocal(values.start_time_local),
      end_time: parseDateTimeLocal(values.end_time_local),
      amount_total: Number(values.amount_total || 0),
      amount_used: Number(values.amount_used || 0),
      next_reset_time: parseDateTimeLocal(values.next_reset_time_local),
    };
    setSaving(true);
    try {
      const res = await API.put(
        `/api/subscription/admin/user_subscriptions/${subId}`,
        payload,
      );
      if (res.data?.success) {
        const msg = res.data?.data?.message;
        showSuccess(msg ? msg : t('更新成功'));
        setEditingRecord(null);
        refreshCurrentPage();
      } else {
        showError(res.data?.message || t('更新失败'));
      }
    } catch (e) {
      showError(t('请求失败'));
    } finally {
      setSaving(false);
    }
  };

  const invalidateSubscription = (record) => {
    const subId = record?.subscription?.id;
    if (!subId) return;
    Modal.confirm({
      title: t('确认作废'),
      content: t('作废后该订阅将立即失效，历史记录不受影响。是否继续？'),
      centered: true,
      onOk: async () => {
        try {
          const res = await API.post(
            `/api/subscription/admin/user_subscriptions/${subId}/invalidate`,
          );
          if (res.data?.success) {
            const msg = res.data?.data?.message;
            showSuccess(msg ? msg : t('已作废'));
            refreshCurrentPage();
          } else {
            showError(res.data?.message || t('操作失败'));
          }
        } catch (e) {
          showError(t('请求失败'));
        }
      },
    });
  };

  const deleteSubscription = (record) => {
    const subId = record?.subscription?.id;
    if (!subId) return;
    Modal.confirm({
      title: t('确认删除'),
      content: t('删除会彻底移除该订阅记录（含权益明细）。是否继续？'),
      centered: true,
      okType: 'danger',
      onOk: async () => {
        try {
          const res = await API.delete(
            `/api/subscription/admin/user_subscriptions/${subId}`,
          );
          if (res.data?.success) {
            const msg = res.data?.data?.message;
            showSuccess(msg ? msg : t('已删除'));
            refreshCurrentPage();
          } else {
            showError(res.data?.message || t('删除失败'));
          }
        } catch (e) {
          showError(t('请求失败'));
        }
      },
    });
  };

  const columns = [
    {
      title: 'ID',
      dataIndex: ['subscription', 'id'],
      key: 'id',
      width: 70,
      render: (text) => <Text type='tertiary'>#{text}</Text>,
    },
    {
      title: t('用户'),
      key: 'user',
      width: 220,
      render: (_, record) => {
        const user = record?.user || {};
        return (
          <div className='min-w-0'>
            <div className='font-medium truncate'>
              {user.display_name || user.username || `#${user.id || '-'}`}
            </div>
            <div className='text-xs text-gray-500 truncate'>
              {user.username || '-'} · ID: {user.id || '-'}
            </div>
            {user.email && (
              <div className='text-xs text-gray-500 truncate'>{user.email}</div>
            )}
          </div>
        );
      },
    },
    {
      title: t('状态'),
      key: 'status',
      width: 90,
      render: (_, record) => renderStatusTag(record?.subscription, t),
    },
    {
      title: t('有效期'),
      key: 'validity',
      width: 210,
      render: (_, record) => {
        const sub = record?.subscription;
        return (
          <div className='text-xs text-gray-600'>
            <div>
              {t('开始')}: {formatTs(sub?.start_time)}
            </div>
            <div>
              {t('结束')}: {formatTs(sub?.end_time)}
            </div>
          </div>
        );
      },
    },
    {
      title: t('用量'),
      key: 'usage',
      width: 180,
      render: (_, record) => {
        const sub = record?.subscription;
        const totalAmount = Number(sub?.amount_total || 0);
        const used = Number(sub?.amount_used || 0);
        return (
          <Text type={totalAmount > 0 ? 'secondary' : 'tertiary'}>
            {totalAmount > 0
              ? formatUsagePercent(used, totalAmount)
              : t('不限')}
          </Text>
        );
      },
    },
    {
      title: t('来源'),
      dataIndex: ['subscription', 'source'],
      key: 'source',
      width: 90,
      render: (text) => <Text type='tertiary'>{text || '-'}</Text>,
    },
    {
      title: t('创建时间'),
      dataIndex: ['subscription', 'created_at'],
      key: 'created_at',
      width: 170,
      render: (text) => <Text type='tertiary'>{formatTs(text)}</Text>,
    },
    {
      title: t('操作'),
      key: 'operate',
      width: 230,
      fixed: 'right',
      render: (_, record) => {
        const sub = record?.subscription;
        const now = Date.now() / 1000;
        const isActive = sub?.status === 'active' && (sub?.end_time || 0) > now;
        return (
          <Space spacing={6}>
            <Button
              size='small'
              theme='light'
              type='secondary'
              onClick={() => setDetailRecord(record)}
            >
              {t('详情')}
            </Button>
            <Button
              size='small'
              theme='light'
              type='tertiary'
              onClick={() => setEditingRecord(record)}
            >
              {t('修改')}
            </Button>
            <Button
              size='small'
              theme='light'
              type='warning'
              disabled={!isActive}
              onClick={() => invalidateSubscription(record)}
            >
              {t('作废')}
            </Button>
            <Button
              size='small'
              theme='light'
              type='danger'
              onClick={() => deleteSubscription(record)}
            >
              {t('删除')}
            </Button>
          </Space>
        );
      },
    },
  ];

  return (
    <>
      <SideSheet
        visible={visible}
        placement='right'
        width={isMobile ? '100%' : 1120}
        bodyStyle={{ padding: 0 }}
        onCancel={onCancel}
        title={
          <Space>
            <Tag color='blue' shape='circle'>
              {t('查看')}
            </Tag>
            <Typography.Title heading={4} className='m-0'>
              {t('购买用户')}
            </Typography.Title>
            <Text type='tertiary' className='ml-2'>
              {plan?.title || '-'} (ID: {plan?.id || '-'})
            </Text>
          </Space>
        }
      >
        <div className='p-4'>
          <div className='flex flex-col md:flex-row md:items-center md:justify-between gap-3 mb-4'>
            <div className='flex flex-col md:flex-row gap-2 flex-1'>
              <Input
                prefix={<IconSearch />}
                value={keyword}
                onChange={setKeyword}
                onEnterPress={handleSearch}
                showClear
                placeholder={t('搜索用户ID、用户名或邮箱')}
                style={{ minWidth: isMobile ? undefined : 260 }}
              />
              <Select
                value={status}
                onChange={handleStatusChange}
                optionList={statusOptions}
                style={{ minWidth: isMobile ? undefined : 160 }}
              />
            </div>
            <Space>
              <Button type='tertiary' onClick={handleSearch} loading={loading}>
                {t('查询')}
              </Button>
              <Button type='tertiary' onClick={handleReset} disabled={loading}>
                {t('重置')}
              </Button>
              <Button
                theme='light'
                type='tertiary'
                icon={<IconRefresh />}
                onClick={() => loadSubscriptions()}
                loading={loading}
              >
                {t('刷新')}
              </Button>
            </Space>
          </div>

          <CardTable
            columns={columns}
            dataSource={records}
            rowKey={(row) => row?.subscription?.id}
            loading={loading}
            scroll={{ x: 'max-content' }}
            hidePagination={false}
            pagination={{
              currentPage,
              pageSize,
              total,
              pageSizeOpts: [10, 20, 50, 100],
              showSizeChanger: true,
              onPageChange: handlePageChange,
              onPageSizeChange: handlePageSizeChange,
            }}
            empty={
              <Empty
                image={
                  <IllustrationNoResult style={{ width: 150, height: 150 }} />
                }
                darkModeImage={
                  <IllustrationNoResultDark
                    style={{ width: 150, height: 150 }}
                  />
                }
                description={t('暂无购买记录')}
                style={{ padding: 30 }}
              />
            }
            size='middle'
          />
        </div>
      </SideSheet>

      <Modal
        title={t('订阅详情')}
        visible={!!detailRecord}
        onCancel={() => setDetailRecord(null)}
        footer={null}
        width={isMobile ? '92%' : 680}
      >
        {detailRecord && (
          <div className='grid grid-cols-1 md:grid-cols-2 gap-4'>
            <DetailItem
              label={t('订阅ID')}
              value={`#${detailRecord.subscription?.id || '-'}`}
            />
            <DetailItem
              label={t('套餐')}
              value={
                planTitleMap.get(detailRecord.subscription?.plan_id) ||
                `#${detailRecord.subscription?.plan_id || '-'}`
              }
            />
            <DetailItem
              label={t('用户')}
              value={`${
                detailRecord.user?.display_name ||
                detailRecord.user?.username ||
                '-'
              } (ID: ${detailRecord.user?.id || '-'})`}
            />
            <DetailItem
              label={t('邮箱')}
              value={detailRecord.user?.email || '-'}
            />
            <DetailItem
              label={t('状态')}
              value={detailRecord.subscription?.status || '-'}
            />
            <DetailItem
              label={t('来源')}
              value={detailRecord.subscription?.source || '-'}
            />
            <DetailItem
              label={t('开始时间')}
              value={formatTs(detailRecord.subscription?.start_time)}
            />
            <DetailItem
              label={t('结束时间')}
              value={formatTs(detailRecord.subscription?.end_time)}
            />
            <DetailItem
              label={t('已用额度')}
              value={detailRecord.subscription?.amount_used}
            />
            <DetailItem
              label={t('总额度')}
              value={
                Number(detailRecord.subscription?.amount_total || 0) > 0
                  ? detailRecord.subscription?.amount_total
                  : t('不限')
              }
            />
            <DetailItem
              label={t('上次重置时间')}
              value={formatTs(detailRecord.subscription?.last_reset_time)}
            />
            <DetailItem
              label={t('下次重置时间')}
              value={formatTs(detailRecord.subscription?.next_reset_time)}
            />
            <DetailItem
              label={t('升级分组')}
              value={detailRecord.subscription?.upgrade_group || t('不升级')}
            />
            <DetailItem
              label={t('升级前分组')}
              value={detailRecord.subscription?.prev_user_group || '-'}
            />
            <DetailItem
              label={t('创建时间')}
              value={formatTs(detailRecord.subscription?.created_at)}
            />
            <DetailItem
              label={t('更新时间')}
              value={formatTs(detailRecord.subscription?.updated_at)}
            />
          </div>
        )}
      </Modal>

      <Modal
        title={t('修改用户订阅')}
        visible={!!editingRecord}
        onCancel={() => setEditingRecord(null)}
        onOk={() => editFormRef.current?.submitForm()}
        confirmLoading={saving}
        width={isMobile ? '92%' : 720}
      >
        {editingRecord && (
          <Form
            key={editingRecord.subscription?.id}
            initValues={getEditInitialValues()}
            getFormApi={(api) => (editFormRef.current = api)}
            onSubmit={submitEdit}
          >
            <Row gutter={12}>
              <Col span={12}>
                <Form.Select
                  field='plan_id'
                  label={t('套餐')}
                  loading={plansLoading}
                  filter
                  required
                  rules={[{ required: true, message: t('请选择订阅套餐') }]}
                >
                  {planOptions.map((item) => (
                    <Select.Option key={item.value} value={item.value}>
                      {item.label}
                    </Select.Option>
                  ))}
                </Form.Select>
              </Col>
              <Col span={12}>
                <Form.Select
                  field='status'
                  label={t('状态')}
                  required
                  rules={[{ required: true, message: t('请选择状态') }]}
                >
                  {editStatusOptions.map((item) => (
                    <Select.Option key={item.value} value={item.value}>
                      {item.label}
                    </Select.Option>
                  ))}
                </Form.Select>
              </Col>
              <Col span={12}>
                <Form.Input
                  field='start_time_local'
                  label={t('开始时间')}
                  type='datetime-local'
                  required
                  rules={[{ required: true, message: t('请选择开始时间') }]}
                />
              </Col>
              <Col span={12}>
                <Form.Input
                  field='end_time_local'
                  label={t('结束时间')}
                  type='datetime-local'
                  required
                  rules={[{ required: true, message: t('请选择结束时间') }]}
                />
              </Col>
              <Col span={12}>
                <Form.InputNumber
                  field='amount_total'
                  label={t('总额度')}
                  min={0}
                  precision={0}
                  style={{ width: '100%' }}
                  extraText={t('0 表示不限')}
                />
              </Col>
              <Col span={12}>
                <Form.InputNumber
                  field='amount_used'
                  label={t('已用额度')}
                  min={0}
                  precision={0}
                  style={{ width: '100%' }}
                />
              </Col>
              <Col span={24}>
                <Form.Input
                  field='next_reset_time_local'
                  label={t('下次重置时间')}
                  type='datetime-local'
                  extraText={t('留空表示不自动重置')}
                />
              </Col>
            </Row>
          </Form>
        )}
      </Modal>
    </>
  );
};

export default PlanSubscriptionsModal;
