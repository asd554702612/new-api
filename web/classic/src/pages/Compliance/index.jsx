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

import React, { useCallback, useEffect, useMemo, useState } from 'react';
import {
  Button,
  Empty,
  Modal,
  Select,
  Space,
  Switch,
  TabPane,
  Tabs,
  Tag,
  TextArea,
  Typography,
} from '@douyinfe/semi-ui';
import dayjs from 'dayjs';
import { useTranslation } from 'react-i18next';
import { API, showError, showSuccess } from '../../helpers';
import CardPro from '../../components/common/ui/CardPro';
import CardTable from '../../components/common/ui/CardTable';
import { createCardProPagination } from '../../helpers/utils';
import { useIsMobile } from '../../hooks/common/useIsMobile';

const PAGE_SIZE_OPTIONS = [10, 20, 50, 100];

const unwrapApiData = (res) => {
  const body = res?.data;
  return body?.data ?? body;
};

const normalizeListPayload = (payload) => {
  if (Array.isArray(payload)) {
    return { items: payload, total: payload.length };
  }

  const items =
    payload?.items ||
    payload?.requests ||
    payload?.feedbacks ||
    payload?.list ||
    [];

  return {
    items: Array.isArray(items) ? items : [],
    total:
      payload?.total ??
      payload?.count ??
      payload?.total_count ??
      (Array.isArray(items) ? items.length : 0),
  };
};

const formatDate = (value) => {
  if (!value) return '-';
  const normalized =
    typeof value === 'number' && value < 10000000000 ? value * 1000 : value;
  const parsed = dayjs(normalized);
  return parsed.isValid() ? parsed.format('YYYY-MM-DD HH:mm') : String(value);
};

const getRecordId = (record) =>
  record?.id ?? record?.request_id ?? record?.feedback_id;

const getPrivacyTypeLabel = (t, type) => {
  const labels = {
    access: t('查阅复制'),
    correction: t('更正补充'),
    deletion: t('删除注销'),
  };
  return labels[type] || type || '-';
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
    completed: t('已完成'),
    resolved: t('已解决'),
    closed: t('已关闭'),
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
    resolved: 'green',
    closed: 'grey',
    rejected: 'red',
    cancelled: 'grey',
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

const buildOptions = (items) =>
  items.map(([value, label]) => ({ value, label }));

const Compliance = () => {
  const { t } = useTranslation();
  const isMobile = useIsMobile();
  const [activeTab, setActiveTab] = useState('privacy');
  const [filters, setFilters] = useState({ status: '', type: '' });
  const [records, setRecords] = useState([]);
  const [loading, setLoading] = useState(false);
  const [activePage, setActivePage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [total, setTotal] = useState(0);
  const [detailVisible, setDetailVisible] = useState(false);
  const [detailLoading, setDetailLoading] = useState(false);
  const [currentRecord, setCurrentRecord] = useState(null);
  const [updating, setUpdating] = useState(false);
  const [nextStatus, setNextStatus] = useState('processing');
  const [processNote, setProcessNote] = useState('');
  const [executeAccountDeletion, setExecuteAccountDeletion] = useState(false);

  const isPrivacyTab = activeTab === 'privacy';
  const listEndpoint = isPrivacyTab
    ? '/api/privacy/admin/requests'
    : '/api/feedback/admin';

  const statusOptions = useMemo(
    () =>
      buildOptions([
        ['', t('全部状态')],
        ['pending', t('待处理')],
        ['processing', t('处理中')],
        [
          isPrivacyTab ? 'completed' : 'resolved',
          isPrivacyTab ? t('已完成') : t('已解决'),
        ],
        ['rejected', t('已驳回')],
        [
          isPrivacyTab ? 'cancelled' : 'closed',
          isPrivacyTab ? t('已取消') : t('已关闭'),
        ],
      ]),
    [isPrivacyTab, t],
  );

  const typeOptions = useMemo(
    () =>
      isPrivacyTab
        ? buildOptions([
            ['', t('全部类型')],
            ['access', t('查阅复制')],
            ['correction', t('更正补充')],
            ['deletion', t('删除注销')],
          ])
        : buildOptions([
            ['', t('全部类型')],
            ['complaint', t('投诉')],
            ['feedback', t('反馈')],
            ['other', t('其他')],
          ]),
    [isPrivacyTab, t],
  );

  const updateStatusOptions = useMemo(
    () =>
      isPrivacyTab
        ? buildOptions([
            ['pending', t('待处理')],
            ['processing', t('处理中')],
            ['completed', t('已完成')],
            ['rejected', t('已驳回')],
            ['cancelled', t('已取消')],
          ])
        : buildOptions([
            ['pending', t('待处理')],
            ['processing', t('处理中')],
            ['resolved', t('已解决')],
            ['closed', t('已关闭')],
            ['rejected', t('已驳回')],
          ]),
    [isPrivacyTab, t],
  );

  const fetchRecords = useCallback(async () => {
    setLoading(true);
    try {
      const typeFilterKey = isPrivacyTab ? 'request_type' : 'feedback_type';
      const res = await API.get(listEndpoint, {
        params: {
          page: activePage,
          p: activePage,
          page_size: pageSize,
          size: pageSize,
          status: filters.status || undefined,
          [typeFilterKey]: filters.type || undefined,
        },
        disableDuplicate: true,
      });
      const { success, message } = res.data || {};
      if (success === false) {
        showError(message || t('列表加载失败'));
        return;
      }

      const payload = normalizeListPayload(unwrapApiData(res));
      setRecords(payload.items);
      setTotal(payload.total);
    } catch (error) {
      showError(error?.message || t('列表加载失败'));
    } finally {
      setLoading(false);
    }
  }, [
    activePage,
    filters.status,
    filters.type,
    isPrivacyTab,
    listEndpoint,
    pageSize,
    t,
  ]);

  useEffect(() => {
    fetchRecords();
  }, [fetchRecords]);

  const handleTabChange = (key) => {
    setActiveTab(key);
    setFilters({ status: '', type: '' });
    setActivePage(1);
    setRecords([]);
    setTotal(0);
  };

  const handleFilterChange = (field, value) => {
    setFilters((prev) => ({ ...prev, [field]: value || '' }));
    setActivePage(1);
  };

  const openDetail = async (record) => {
    const recordId = getRecordId(record);
    setCurrentRecord(record);
    setNextStatus(
      record.status || (isPrivacyTab ? 'processing' : 'processing'),
    );
    setProcessNote(record.admin_note || '');
    setExecuteAccountDeletion(Boolean(record.execute_account_deletion));
    setDetailVisible(true);

    if (!recordId) return;

    setDetailLoading(true);
    try {
      const res = await API.get(`${listEndpoint}/${recordId}`, {
        skipErrorHandler: true,
      });
      if (res.data?.success === false) return;

      const detail = unwrapApiData(res);
      const normalizedDetail =
        detail?.item || detail?.request || detail?.feedback || detail;
      if (normalizedDetail && typeof normalizedDetail === 'object') {
        setCurrentRecord({ ...record, ...normalizedDetail });
        setNextStatus(normalizedDetail.status || record.status || 'processing');
        setProcessNote(normalizedDetail.admin_note || '');
        setExecuteAccountDeletion(
          Boolean(normalizedDetail.execute_account_deletion),
        );
      }
    } catch (error) {
      setCurrentRecord(record);
    } finally {
      setDetailLoading(false);
    }
  };

  const updateRecord = async () => {
    const recordId = getRecordId(currentRecord);
    if (!recordId) {
      showError(t('无法识别记录'));
      return;
    }

    setUpdating(true);
    try {
      const payload = {
        status: nextStatus,
        admin_note: processNote.trim(),
      };

      if (isPrivacyTab) {
        payload.execute_account_deletion = executeAccountDeletion;
      }

      const res = await API.patch(`${listEndpoint}/${recordId}`, payload, {
        skipErrorHandler: true,
      });
      if (res.data?.success === false) {
        throw new Error(res.data?.message || t('更新失败，请重试'));
      }
      showSuccess(t('更新成功'));
      setDetailVisible(false);
      await fetchRecords();
    } catch (error) {
      showError(
        error?.response?.data?.message ||
          error?.message ||
          t('更新失败，请重试'),
      );
    } finally {
      setUpdating(false);
    }
  };

  const renderType = (record) => {
    const type = record.request_type || record.feedback_type;
    return isPrivacyTab
      ? getPrivacyTypeLabel(t, type)
      : getFeedbackTypeLabel(t, type);
  };

  const getRecordTitle = (record) => record.title || record.content || '-';

  const columns = [
    {
      title: t('编号'),
      dataIndex: 'id',
      width: 90,
      render: (_, record, index) => getRecordId(record) || index + 1,
    },
    {
      title: t('追踪码'),
      dataIndex: 'tracking_code',
      render: (value) => value || '-',
    },
    {
      title: t('类型'),
      dataIndex: 'type',
      render: (_, record) => <Tag color='blue'>{renderType(record)}</Tag>,
    },
    {
      title: t('状态'),
      dataIndex: 'status',
      render: (value) => (
        <Tag color={getStatusColor(value)}>{getStatusLabel(t, value)}</Tag>
      ),
    },
    {
      title: isPrivacyTab ? t('申请人') : t('联系人'),
      dataIndex: 'contact',
      render: (_, record) =>
        isPrivacyTab
          ? record.username || record.user_id || '-'
          : record.contact_name || record.contact_email || '-',
    },
    {
      title: t('标题或说明'),
      dataIndex: 'title',
      render: (_, record) => (
        <span className='line-clamp-2'>{getRecordTitle(record)}</span>
      ),
    },
    {
      title: t('提交时间'),
      dataIndex: 'created_at',
      render: (value) => formatDate(value),
    },
    {
      title: t('操作'),
      dataIndex: 'operate',
      fixed: isMobile ? undefined : 'right',
      render: (_, record) => (
        <Button size='small' type='tertiary' onClick={() => openDetail(record)}>
          {t('详情')}
        </Button>
      ),
    },
  ];

  const searchArea = (
    <div className='flex flex-col md:flex-row gap-3 md:items-center md:justify-between w-full'>
      <Space wrap>
        <Select
          value={filters.status}
          optionList={statusOptions}
          onChange={(value) => handleFilterChange('status', value)}
          style={{ width: 160 }}
        />
        <Select
          value={filters.type}
          optionList={typeOptions}
          onChange={(value) => handleFilterChange('type', value)}
          style={{ width: 160 }}
        />
      </Space>
      <Button type='tertiary' loading={loading} onClick={fetchRecords}>
        {t('刷新')}
      </Button>
    </div>
  );

  const currentType =
    currentRecord?.request_type || currentRecord?.feedback_type;
  const showDeletionSwitch = isPrivacyTab && currentType === 'deletion';

  return (
    <div className='mt-[60px] px-2'>
      <CardPro
        type='type3'
        tabsArea={
          <Tabs type='button' activeKey={activeTab} onChange={handleTabChange}>
            <TabPane itemKey='privacy' tab={t('个人信息申请')} />
            <TabPane itemKey='feedback' tab={t('投诉反馈')} />
          </Tabs>
        }
        searchArea={searchArea}
        paginationArea={createCardProPagination({
          currentPage: activePage,
          pageSize,
          total,
          onPageChange: setActivePage,
          onPageSizeChange: (size) => {
            setPageSize(size);
            setActivePage(1);
          },
          pageSizeOpts: PAGE_SIZE_OPTIONS,
          isMobile,
          t,
        })}
        t={t}
      >
        <CardTable
          columns={columns}
          dataSource={records}
          rowKey={(record, index) => getRecordId(record) || index}
          loading={loading}
          hidePagination
          scroll={isMobile ? undefined : { x: 'max-content' }}
          empty={<Empty description={t('暂无数据')} />}
        />
      </CardPro>

      <Modal
        title={isPrivacyTab ? t('个人信息申请详情') : t('投诉反馈详情')}
        visible={detailVisible}
        onCancel={() => setDetailVisible(false)}
        footer={
          <Space>
            <Button onClick={() => setDetailVisible(false)}>{t('取消')}</Button>
            <Button type='primary' loading={updating} onClick={updateRecord}>
              {t('保存处理结果')}
            </Button>
          </Space>
        }
        style={{ maxWidth: 720 }}
        width='90%'
      >
        <div className='space-y-4'>
          <div
            className={`grid grid-cols-1 ${isMobile ? '' : 'md:grid-cols-2'} gap-x-6`}
          >
            <DetailItem label={t('编号')} value={getRecordId(currentRecord)} />
            <DetailItem
              label={t('追踪码')}
              value={currentRecord?.tracking_code}
            />
            <DetailItem
              label={t('类型')}
              value={renderType(currentRecord || {})}
            />
            <DetailItem
              label={t('状态')}
              value={getStatusLabel(t, currentRecord?.status)}
            />
            <DetailItem
              label={isPrivacyTab ? t('申请人') : t('联系人')}
              value={
                isPrivacyTab
                  ? currentRecord?.username || currentRecord?.user_id
                  : currentRecord?.contact_name || currentRecord?.contact_email
              }
            />
            {!isPrivacyTab && (
              <>
                <DetailItem
                  label={t('邮箱')}
                  value={currentRecord?.contact_email}
                />
                <DetailItem
                  label={t('手机号')}
                  value={currentRecord?.contact_phone}
                />
              </>
            )}
            <DetailItem
              label={t('提交时间')}
              value={formatDate(currentRecord?.created_at)}
            />
            <DetailItem
              label={t('更新时间')}
              value={formatDate(currentRecord?.updated_at)}
            />
          </div>

          <DetailItem
            label={isPrivacyTab ? t('申请说明') : t('内容')}
            value={currentRecord?.content}
          />

          <div className='rounded-xl border border-gray-100 p-4'>
            <Typography.Text strong>{t('处理')}</Typography.Text>
            <div className='grid grid-cols-1 md:grid-cols-2 gap-4 mt-3'>
              <div>
                <div className='text-sm font-medium mb-2'>{t('状态')}</div>
                <Select
                  value={nextStatus}
                  optionList={updateStatusOptions}
                  onChange={(value) => setNextStatus(value)}
                  style={{ width: '100%' }}
                  loading={detailLoading}
                />
              </div>
              {showDeletionSwitch && (
                <div>
                  <div className='text-sm font-medium mb-2'>
                    {t('执行账号注销')}
                  </div>
                  <Switch
                    checked={executeAccountDeletion}
                    onChange={setExecuteAccountDeletion}
                  />
                </div>
              )}
            </div>
            <div className='mt-4'>
              <div className='text-sm font-medium mb-2'>{t('处理说明')}</div>
              <TextArea
                value={processNote}
                onChange={setProcessNote}
                autosize={{ minRows: 4, maxRows: 8 }}
                placeholder={t('请填写处理说明')}
              />
            </div>
          </div>
        </div>
      </Modal>
    </div>
  );
};

export default Compliance;
