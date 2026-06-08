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
  Button,
  Card,
  DatePicker,
  Empty,
  Input,
  InputNumber,
  Modal,
  Select,
  Space,
  Table,
  TabPane,
  Tabs,
  Tag,
  Typography,
} from '@douyinfe/semi-ui';
import {
  IllustrationNoResult,
  IllustrationNoResultDark,
} from '@douyinfe/semi-illustrations';
import { Search, RefreshCw } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import {
  API,
  renderQuota,
  showError,
  showSuccess,
  timestamp2string,
} from '../../helpers';

const { Text } = Typography;

const TAB_ENDPOINTS = {
  invites: '/api/user/affiliates/invites',
  rebates: '/api/user/affiliates/rebates',
  transfers: '/api/user/affiliates/transfers',
  withdrawals: '/api/user/affiliates/withdrawals',
  users: '/api/user/affiliates/users',
  fingerprints: '/api/user/affiliates/fingerprints',
};

const WITHDRAWAL_STATUS_COLORS = {
  pending_review: 'orange',
  approved: 'blue',
  paid: 'green',
  rejected: 'red',
  failed: 'red',
  cancelled: 'grey',
};

const DEFAULT_IDENTITY_CONFIG = {
  inviter_rate_multiplier: 1.5,
  invitee_rate_multiplier: 1.4,
  duration_hours: 720,
  qualified_invitee_count: 0,
  qualified_pay_amount: 50,
  eligible_order_types: ['topup', 'subscription'],
  fingerprint_enforcement_enabled: true,
  max_accounts_per_fingerprint_hash: 3,
};

const positiveNumberOrDefault = (value, fallback) => {
  const parsed = Number(value);
  return Number.isFinite(parsed) && parsed > 0 ? parsed : fallback;
};

const nonNegativeNumberOrDefault = (value, fallback) => {
  const parsed = Number(value);
  return Number.isFinite(parsed) && parsed >= 0 ? parsed : fallback;
};

const normalizeIdentityConfig = (config = {}) => {
  const merged = {
    ...DEFAULT_IDENTITY_CONFIG,
    ...config,
  };
  return {
    ...merged,
    inviter_rate_multiplier: positiveNumberOrDefault(
      merged.inviter_rate_multiplier,
      DEFAULT_IDENTITY_CONFIG.inviter_rate_multiplier,
    ),
    invitee_rate_multiplier: positiveNumberOrDefault(
      merged.invitee_rate_multiplier,
      DEFAULT_IDENTITY_CONFIG.invitee_rate_multiplier,
    ),
    duration_hours: positiveNumberOrDefault(
      merged.duration_hours,
      DEFAULT_IDENTITY_CONFIG.duration_hours,
    ),
    qualified_invitee_count: nonNegativeNumberOrDefault(
      merged.qualified_invitee_count,
      DEFAULT_IDENTITY_CONFIG.qualified_invitee_count,
    ),
    qualified_pay_amount: nonNegativeNumberOrDefault(
      merged.qualified_pay_amount,
      DEFAULT_IDENTITY_CONFIG.qualified_pay_amount,
    ),
    eligible_order_types:
      Array.isArray(merged.eligible_order_types) && merged.eligible_order_types.length
        ? merged.eligible_order_types
        : DEFAULT_IDENTITY_CONFIG.eligible_order_types,
    fingerprint_enforcement_enabled: Boolean(merged.fingerprint_enforcement_enabled),
    max_accounts_per_fingerprint_hash: positiveNumberOrDefault(
      merged.max_accounts_per_fingerprint_hash,
      DEFAULT_IDENTITY_CONFIG.max_accounts_per_fingerprint_hash,
    ),
  };
};

const formatDateValue = (value) => {
  if (!value) return '';
  if (typeof value === 'string') return value.slice(0, 10);
  if (value instanceof Date && !Number.isNaN(value.getTime())) {
    return value.toISOString().slice(0, 10);
  }
  return String(value).slice(0, 10);
};

const PAGE_SIZE_OPTIONS = [10, 20, 50, 100];
const PAGE_SIZE_OPTION_STRINGS = PAGE_SIZE_OPTIONS.map(String);
const TABLE_MIN_WIDTHS = {
  invites: 1000,
  rebates: 1400,
  transfers: 1200,
  withdrawals: 1500,
  users: 1250,
  fingerprints: 1100,
};

const getTableScrollX = (tabKey) => {
  const minWidth = TABLE_MIN_WIDTHS[tabKey] || 1000;
  return `max(100%, ${minWidth}px)`;
};

const AffiliateAdminPage = () => {
  const { t } = useTranslation();
  const [activeTab, setActiveTab] = useState('invites');
  const [records, setRecords] = useState([]);
  const [loading, setLoading] = useState(false);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);
  const [total, setTotal] = useState(0);
  const [search, setSearch] = useState('');
  const [dateRange, setDateRange] = useState([]);
  const [status, setStatus] = useState('');
  const [selectedRowKeys, setSelectedRowKeys] = useState([]);
  const [identityConfig, setIdentityConfig] = useState({
    enabled: false,
    config: DEFAULT_IDENTITY_CONFIG,
  });

  const loadRecords = async () => {
    if (activeTab === 'identity_config') {
      await loadIdentityConfig();
      return;
    }
    setLoading(true);
    try {
      const params = new URLSearchParams({
        p: String(page),
        page_size: String(pageSize),
      });
      if (search.trim()) params.set('search', search.trim());
      if (dateRange?.[0]) params.set('start_at', dateRange[0]);
      if (dateRange?.[1]) params.set('end_at', dateRange[1]);
      if ((activeTab === 'withdrawals' || activeTab === 'fingerprints') && status) {
        params.set('status', status);
      }
      const res = await API.get(`${TAB_ENDPOINTS[activeTab]}?${params}`);
      if (res.data?.success) {
        setRecords(res.data.data?.items || []);
        setTotal(res.data.data?.total || 0);
      } else {
        showError(res.data?.message || t('加载失败'));
      }
    } catch (error) {
      showError(t('加载失败'));
    } finally {
      setLoading(false);
    }
  };

  const loadIdentityConfig = async () => {
    setLoading(true);
    try {
      const res = await API.get('/api/user/affiliates/identity-config');
      if (res.data?.success) {
        setIdentityConfig({
          enabled: res.data.data?.enabled === true,
          config: normalizeIdentityConfig(res.data.data?.config),
        });
      } else {
        showError(res.data?.message || t('加载失败'));
      }
    } catch {
      showError(t('加载失败'));
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadRecords();
  }, [activeTab, page, pageSize, status]);

  const reloadFromFirstPage = () => {
    if (page === 1) {
      loadRecords();
    } else {
      setPage(1);
    }
  };

  const updateIdentityConfigField = (field, value) => {
    setIdentityConfig((prev) => ({
      ...prev,
      config: {
        ...prev.config,
        [field]: value,
      },
    }));
  };

  const renderUser = (id, username, email) => (
    <div className='min-w-[160px]'>
      <Text strong>#{id || '-'}</Text>
      <div className='text-sm'>{username || '-'}</div>
      <div className='text-xs text-gray-500 break-all'>{email || '-'}</div>
    </div>
  );

  const renderTime = (value) => (value ? timestamp2string(value) : '-');

  const performWithdrawalAction = async (record, action, payload = {}) => {
    try {
      const res = await API.post(
        `/api/user/affiliates/withdrawals/${record.id}/${action}`,
        payload,
      );
      if (res.data?.success) {
        showSuccess(t('操作成功'));
        loadRecords();
      } else {
        showError(res.data?.message || t('操作失败'));
      }
    } catch (error) {
      showError(t('操作失败'));
    }
  };

  const requestWithdrawalApprove = (record) => {
    let note = '';
    Modal.confirm({
      title: t('审核通过'),
      centered: true,
      content: (
        <Input
          prefix={t('管理员备注')}
          onChange={(value) => {
            note = value;
          }}
        />
      ),
      onOk: () => performWithdrawalAction(record, 'approve', { note }),
    });
  };

  const requestWithdrawalReasonAction = (record, action) => {
    let reason = '';
    const actionTitles = {
      reject: t('拒绝提现申请'),
      fail: t('标记打款失败'),
    };
    Modal.confirm({
      title: actionTitles[action],
      centered: true,
      content: (
        <Input
          prefix={action === 'reject' ? t('拒绝原因') : t('失败原因')}
          onChange={(value) => {
            reason = value;
          }}
        />
      ),
      onOk: () => {
        if (!reason.trim()) {
          showError(t('请填写原因'));
          return false;
        }
        return performWithdrawalAction(record, action, { reason });
      },
    });
  };

  const markWithdrawalPaid = (record) => {
    let payoutChannel = '';
    let payoutTradeNo = '';
    let adminNote = '';
    Modal.confirm({
      title: t('标记已打款'),
      centered: true,
      content: (
        <div className='space-y-3'>
          <Input
            prefix={t('打款渠道')}
            placeholder='wechat'
            onChange={(value) => {
              payoutChannel = value;
            }}
          />
          <Input
            prefix={t('打款流水号')}
            onChange={(value) => {
              payoutTradeNo = value;
            }}
          />
          <Input
            prefix={t('管理员备注')}
            onChange={(value) => {
              adminNote = value;
            }}
          />
        </div>
      ),
      onOk: () => {
        if (!payoutChannel.trim()) {
          showError(t('请填写打款渠道'));
          return false;
        }
        return performWithdrawalAction(record, 'paid', {
          payout_channel: payoutChannel,
          payout_trade_no: payoutTradeNo,
          admin_note: adminNote,
        });
      },
    });
  };

  const requestManualInviteRelation = () => {
    let inviterUserId = '';
    let inviteeUserId = '';
    let overwrite = false;
    Modal.confirm({
      title: t('手动创建邀请关系'),
      centered: true,
      content: (
        <div className='space-y-3'>
          <Input placeholder={t('邀请人用户ID')} onChange={(value) => (inviterUserId = value)} />
          <Input placeholder={t('被邀请人用户ID')} onChange={(value) => (inviteeUserId = value)} />
          <Select
            style={{ width: '100%' }}
            defaultValue='false'
            onChange={(value) => {
              overwrite = value === 'true';
            }}
          >
            <Select.Option value='false'>{t('不覆盖已有关系')}</Select.Option>
            <Select.Option value='true'>{t('覆盖已有关系')}</Select.Option>
          </Select>
        </div>
      ),
      onOk: async () => {
        const res = await API.post('/api/user/affiliates/invites', {
          inviter_user_id: Number(inviterUserId),
          invitee_user_id: Number(inviteeUserId),
          overwrite,
        });
        if (res.data?.success) {
          showSuccess(t('操作成功'));
          loadRecords();
        } else {
          showError(res.data?.message || t('操作失败'));
          return false;
        }
      },
    });
  };

  const requestEditAffiliateUser = (record = {}) => {
    let affCode = record.aff_code || '';
    let rate =
      record.aff_rebate_rate_percent === null ||
      record.aff_rebate_rate_percent === undefined
        ? undefined
        : Number(record.aff_rebate_rate_percent);
    Modal.confirm({
      title: record.user_id ? t('编辑专属用户') : t('添加专属用户'),
      centered: true,
      content: (
        <div className='space-y-3'>
          {!record.user_id && (
            <Input placeholder={t('用户ID')} onChange={(value) => (record.user_id = Number(value))} />
          )}
          <Input defaultValue={affCode} placeholder={t('专属邀请码')} onChange={(value) => (affCode = value)} />
          <InputNumber
            defaultValue={rate}
            min={0}
            max={100}
            suffix='%'
            placeholder={t('专属返利比例')}
            style={{ width: '100%' }}
            onChange={(value) => (rate = value)}
          />
        </div>
      ),
      onOk: async () => {
        if (!record.user_id) {
          showError(t('请输入用户ID'));
          return false;
        }
        const payload = { aff_code: affCode };
        if (rate !== undefined && rate !== null && rate !== '') {
          payload.aff_rebate_rate_percent = Number(rate);
        }
        const res = await API.put(`/api/user/affiliates/users/${record.user_id}`, payload);
        if (res.data?.success) {
          showSuccess(t('操作成功'));
          loadRecords();
        } else {
          showError(res.data?.message || t('操作失败'));
          return false;
        }
      },
    });
  };

  const clearAffiliateUser = async (record) => {
    const res = await API.delete(`/api/user/affiliates/users/${record.user_id}`);
    if (res.data?.success) {
      showSuccess(t('操作成功'));
      loadRecords();
    } else {
      showError(res.data?.message || t('操作失败'));
    }
  };

  const requestBatchRate = () => {
    let rate = undefined;
    Modal.confirm({
      title: t('批量设置返利比例'),
      centered: true,
      content: (
        <InputNumber
          min={0}
          max={100}
          suffix='%'
          style={{ width: '100%' }}
          placeholder={t('专属返利比例')}
          onChange={(value) => (rate = value)}
        />
      ),
      onOk: async () => {
        if (!selectedRowKeys.length) {
          showError(t('请先选择用户'));
          return false;
        }
        const res = await API.post('/api/user/affiliates/users/batch-rate', {
          user_ids: selectedRowKeys,
          aff_rebate_rate_percent: Number(rate),
        });
        if (res.data?.success) {
          showSuccess(t('操作成功'));
          setSelectedRowKeys([]);
          loadRecords();
        } else {
          showError(res.data?.message || t('操作失败'));
          return false;
        }
      },
    });
  };

  const saveIdentityConfig = async () => {
    try {
      const config = normalizeIdentityConfig(identityConfig.config);
      const res = await API.put('/api/user/affiliates/identity-config', {
        enabled: identityConfig.enabled,
        config,
      });
      if (res.data?.success) {
        showSuccess(t('保存成功'));
        loadIdentityConfig();
      } else {
        showError(res.data?.message || t('保存失败'));
      }
    } catch {
      showError(t('保存失败'));
    }
  };

  const columns = useMemo(() => {
    if (activeTab === 'invites') {
      return [
        {
          title: t('邀请人'),
          key: 'inviter',
          width: 220,
          render: (_, record) =>
            renderUser(record.inviter_id, record.inviter_username, record.inviter_email),
        },
        {
          title: t('被邀请人'),
          key: 'invitee',
          width: 220,
          render: (_, record) => renderUser(record.user_id, record.username, record.email),
        },
        { title: t('邀请码'), dataIndex: 'aff_code', key: 'aff_code', width: 140, render: (value) => value || '-' },
        { title: t('注册时间'), dataIndex: 'created_at', key: 'created_at', width: 180, render: renderTime },
      ];
    }
    if (activeTab === 'users') {
      return [
        { title: t('用户'), key: 'user', width: 220, render: (_, record) => renderUser(record.user_id, record.username, record.email) },
        { title: t('邀请码'), dataIndex: 'aff_code', key: 'aff_code', width: 180, render: (value, record) => (
          <Space>
            <Text>{value || '-'}</Text>
            {record.aff_code_custom && <Tag color='blue'>{t('专属')}</Tag>}
          </Space>
        ) },
        { title: t('专属返利比例'), dataIndex: 'aff_rebate_rate_percent', key: 'aff_rebate_rate_percent', width: 150, render: (value) => value === null || value === undefined ? '-' : `${value}%` },
        { title: t('邀请数'), dataIndex: 'aff_count', key: 'aff_count', width: 100 },
        { title: t('待使用返利'), dataIndex: 'aff_quota', key: 'aff_quota', width: 140, render: (value) => renderQuota(value || 0) },
        { title: t('累计返利'), dataIndex: 'aff_history_quota', key: 'aff_history_quota', width: 140, render: (value) => renderQuota(value || 0) },
        {
          title: t('操作'),
          key: 'action',
          dataIndex: 'action',
          fixed: 'right',
          width: 160,
          render: (_, record) => (
            <Space>
              <Button size='small' onClick={() => requestEditAffiliateUser(record)}>{t('编辑')}</Button>
              <Button size='small' type='danger' onClick={() => clearAffiliateUser(record)}>{t('清除')}</Button>
            </Space>
          ),
        },
      ];
    }
    if (activeTab === 'fingerprints') {
      return [
        { title: t('用户'), key: 'user', width: 220, render: (_, record) => renderUser(record.user_id, record.username, record.email) },
        { title: t('综合指纹'), dataIndex: 'composite_hash', key: 'composite_hash', width: 280, render: (value) => value || '-' },
        { title: t('重复数'), dataIndex: 'duplicate_count', key: 'duplicate_count', width: 100 },
        { title: t('风险'), dataIndex: 'risk_flagged', key: 'risk_flagged', width: 160, render: (value, record) => value ? <Tag color='red'>{record.risk_reason || t('风险')}</Tag> : <Tag color='green'>{t('正常')}</Tag> },
        { title: t('时间'), dataIndex: 'created_at', key: 'created_at', width: 180, render: renderTime },
      ];
    }
    if (activeTab === 'rebates' || activeTab === 'transfers') {
      return [
        {
          title: activeTab === 'rebates' ? t('邀请人') : t('用户'),
          key: 'user',
          width: 220,
          render: (_, record) => renderUser(record.user_id, record.username, record.email),
        },
        ...(activeTab === 'rebates'
          ? [
              {
                title: t('被邀请人'),
                key: 'related_user',
                width: 220,
                render: (_, record) =>
                  renderUser(
                    record.related_user_id,
                    record.related_username,
                    record.related_email,
                  ),
              },
            ]
          : []),
        { title: t('额度'), dataIndex: 'quota', key: 'quota', width: 140, render: (value) => renderQuota(value || 0) },
        {
          title: t('订单/备注'),
          key: 'source',
          width: 220,
          render: (_, record) => record.source_order_trade_no || record.remark || '-',
        },
        { title: t('支付方式'), dataIndex: 'payment_method', key: 'payment_method', width: 120, render: (value) => value || '-' },
        {
          title: t('当前待使用'),
          dataIndex: 'balance_after',
          key: 'balance_after',
          width: 140,
          render: (value) => renderQuota(value || 0),
        },
        { title: t('时间'), dataIndex: 'created_at', key: 'created_at', width: 180, render: renderTime },
      ];
    }
    return [
      {
        title: t('用户'),
        key: 'user',
        width: 220,
        render: (_, record) => renderUser(record.user_id, record.username, record.email),
      },
      { title: t('提现额度'), dataIndex: 'quota', key: 'quota', width: 140, render: (value) => renderQuota(value || 0) },
      {
        title: t('状态'),
        dataIndex: 'status',
        key: 'status',
        width: 130,
        render: (value) => (
          <Tag color={WITHDRAWAL_STATUS_COLORS[value] || 'grey'}>{t(value || '-')}</Tag>
        ),
      },
      { title: t('收款方式'), dataIndex: 'payout_method', key: 'payout_method', width: 140, render: (value) => value || '-' },
      {
        title: t('收款说明'),
        dataIndex: 'payout_account_note',
        key: 'payout_account_note',
        width: 220,
        render: (value) => <Text ellipsis={{ showTooltip: true }}>{value || '-'}</Text>,
      },
      { title: t('打款渠道'), dataIndex: 'payout_channel', key: 'payout_channel', width: 140, render: (value) => value || '-' },
      { title: t('打款流水号'), dataIndex: 'payout_trade_no', key: 'payout_trade_no', width: 180, render: (value) => value || '-' },
      { title: t('申请时间'), dataIndex: 'created_at', key: 'created_at', width: 180, render: renderTime },
      {
        title: t('操作'),
        key: 'action',
        dataIndex: 'action',
        fixed: 'right',
        width: 180,
        render: (_, record) => (
          <Space>
            {record.status === 'pending_review' && (
              <>
                <Button
                  size='small'
                  onClick={() => requestWithdrawalApprove(record)}
                >
                  {t('通过')}
                </Button>
                <Button
                  size='small'
                  type='danger'
                  onClick={() => requestWithdrawalReasonAction(record, 'reject')}
                >
                  {t('拒绝')}
                </Button>
              </>
            )}
            {record.status === 'approved' && (
              <>
                <Button size='small' type='primary' onClick={() => markWithdrawalPaid(record)}>
                  {t('已打款')}
                </Button>
                <Button
                  size='small'
                  type='danger'
                  onClick={() => requestWithdrawalReasonAction(record, 'fail')}
                >
                  {t('失败')}
                </Button>
              </>
            )}
            {!['pending_review', 'approved'].includes(record.status) && (
              <Text type='tertiary'>{t('无操作')}</Text>
            )}
          </Space>
        ),
      },
    ];
  }, [activeTab, t]);

  const tabTitle = {
    invites: t('邀请关系'),
    rebates: t('返利记录'),
    transfers: t('划转记录'),
    withdrawals: t('提现记录'),
    users: t('专属用户'),
    fingerprints: t('指纹风险'),
    identity_config: t('身份配置'),
  }[activeTab];
  const tableScrollX = getTableScrollX(activeTab);

  return (
    <div className='mt-[60px] px-2'>
      <Card className='!rounded-2xl'>
        <div className='flex flex-col lg:flex-row lg:items-center lg:justify-between gap-3 mb-4'>
          <div>
            <Typography.Title heading={4}>{t('邀请返利')}</Typography.Title>
            <Text type='tertiary'>{tabTitle}</Text>
          </div>
          <Space wrap>
            <Input
              prefix={<Search size={14} />}
              placeholder={t('搜索用户ID、用户名或邮箱')}
              value={search}
              onChange={setSearch}
              onEnterPress={reloadFromFirstPage}
              showClear
            />
            <DatePicker
              type='dateRange'
              onChange={(value) => {
                const range = Array.isArray(value) ? value : [];
                setDateRange(range.map(formatDateValue));
              }}
              placeholder={[t('开始日期'), t('结束日期')]}
            />
            {activeTab === 'withdrawals' && (
              <Select
                value={status}
                onChange={(value) => {
                  setStatus(value || '');
                  setPage(1);
                }}
                style={{ width: 160 }}
              >
                <Select.Option value=''>{t('全部状态')}</Select.Option>
                {Object.keys(WITHDRAWAL_STATUS_COLORS).map((item) => (
                  <Select.Option key={item} value={item}>
                    {t(item)}
                  </Select.Option>
                ))}
              </Select>
            )}
            {activeTab === 'fingerprints' && (
              <Select
                value={status}
                onChange={(value) => {
                  setStatus(value || '');
                  setPage(1);
                }}
                style={{ width: 140 }}
              >
                <Select.Option value=''>{t('全部')}</Select.Option>
                <Select.Option value='risk'>{t('仅风险')}</Select.Option>
              </Select>
            )}
            {activeTab === 'invites' && (
              <Button onClick={requestManualInviteRelation}>{t('手动关系')}</Button>
            )}
            {activeTab === 'users' && (
              <>
                <Button onClick={() => requestEditAffiliateUser()}>{t('添加专属用户')}</Button>
                <Button onClick={requestBatchRate}>{t('批量比例')}</Button>
              </>
            )}
            <Button icon={<RefreshCw size={14} />} onClick={reloadFromFirstPage}>
              {t('查询')}
            </Button>
          </Space>
        </div>

        <Tabs
          type='card'
          activeKey={activeTab}
          onChange={(key) => {
            setActiveTab(key);
            setPage(1);
            setRecords([]);
          }}
        >
          <TabPane itemKey='invites' tab={t('邀请关系')} />
          <TabPane itemKey='rebates' tab={t('返利记录')} />
          <TabPane itemKey='transfers' tab={t('划转记录')} />
          <TabPane itemKey='withdrawals' tab={t('提现记录')} />
          <TabPane itemKey='users' tab={t('专属用户')} />
          <TabPane itemKey='fingerprints' tab={t('指纹风险')} />
          <TabPane itemKey='identity_config' tab={t('身份配置')} />
        </Tabs>

        {activeTab === 'identity_config' ? (
          <div className='space-y-4 max-w-[920px]'>
            <div className='grid grid-cols-1 md:grid-cols-2 gap-4'>
              <div>
                <Text strong>{t('身份倍率')}</Text>
                <div className='grid grid-cols-1 md:grid-cols-2 gap-3 mt-2'>
                  <InputNumber
                    value={identityConfig.config.inviter_rate_multiplier}
                    min={0.01}
                    step={0.1}
                    precision={2}
                    prefix={t('邀请人')}
                    style={{ width: '100%' }}
                    onChange={(value) => updateIdentityConfigField('inviter_rate_multiplier', value)}
                  />
                  <InputNumber
                    value={identityConfig.config.invitee_rate_multiplier}
                    min={0.01}
                    step={0.1}
                    precision={2}
                    prefix={t('被邀请人')}
                    style={{ width: '100%' }}
                    onChange={(value) => updateIdentityConfigField('invitee_rate_multiplier', value)}
                  />
                </div>
              </div>
              <div>
                <Text strong>{t('身份有效期')}</Text>
                <InputNumber
                  value={identityConfig.config.duration_hours}
                  min={1}
                  step={24}
                  suffix={t('小时')}
                  style={{ width: '100%', marginTop: 8 }}
                  onChange={(value) => updateIdentityConfigField('duration_hours', value)}
                />
              </div>
              <div>
                <Text strong>{t('资格门槛')}</Text>
                <div className='grid grid-cols-1 md:grid-cols-2 gap-3 mt-2'>
                  <InputNumber
                    value={identityConfig.config.qualified_invitee_count}
                    min={0}
                    step={1}
                    suffix={t('人')}
                    prefix={t('邀请人数')}
                    style={{ width: '100%' }}
                    onChange={(value) => updateIdentityConfigField('qualified_invitee_count', value)}
                  />
                  <InputNumber
                    value={identityConfig.config.qualified_pay_amount}
                    min={0}
                    step={1}
                    prefix={t('支付额度')}
                    style={{ width: '100%' }}
                    onChange={(value) => updateIdentityConfigField('qualified_pay_amount', value)}
                  />
                </div>
              </div>
              <div>
                <Text strong>{t('合格订单类型')}</Text>
                <Select
                  multiple
                  value={identityConfig.config.eligible_order_types}
                  style={{ width: '100%', marginTop: 8 }}
                  onChange={(value) => updateIdentityConfigField('eligible_order_types', value)}
                >
                  <Select.Option value='topup'>{t('充值')}</Select.Option>
                  <Select.Option value='subscription'>{t('订阅')}</Select.Option>
                </Select>
              </div>
              <div>
                <Text strong>{t('指纹风控')}</Text>
                <div className='grid grid-cols-1 md:grid-cols-2 gap-3 mt-2'>
                  <Select
                    value={identityConfig.config.fingerprint_enforcement_enabled ? 'true' : 'false'}
                    style={{ width: '100%' }}
                    onChange={(value) => updateIdentityConfigField('fingerprint_enforcement_enabled', value === 'true')}
                  >
                    <Select.Option value='true'>{t('启用')}</Select.Option>
                    <Select.Option value='false'>{t('关闭')}</Select.Option>
                  </Select>
                  <InputNumber
                    value={identityConfig.config.max_accounts_per_fingerprint_hash}
                    min={1}
                    step={1}
                    prefix={t('同指纹上限')}
                    style={{ width: '100%' }}
                    onChange={(value) => updateIdentityConfigField('max_accounts_per_fingerprint_hash', value)}
                  />
                </div>
              </div>
              <div>
                <Text strong>{t('启用状态')}</Text>
                <Select
                  value={identityConfig.enabled ? 'true' : 'false'}
                  style={{ width: '100%', marginTop: 8 }}
                  onChange={(value) => setIdentityConfig({ ...identityConfig, enabled: value === 'true' })}
                >
                  <Select.Option value='true'>{t('启用')}</Select.Option>
                  <Select.Option value='false'>{t('关闭')}</Select.Option>
                </Select>
              </div>
            </div>
            <Button type='primary' onClick={saveIdentityConfig}>{t('保存')}</Button>
          </div>
        ) : (
          <Table
            columns={columns}
            dataSource={records}
            loading={loading}
            rowKey={activeTab === 'invites' ? 'user_id' : activeTab === 'users' ? 'user_id' : 'id'}
            rowSelection={activeTab === 'users' ? {
              selectedRowKeys,
              onChange: setSelectedRowKeys,
            } : undefined}
            scroll={{ x: tableScrollX }}
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
            size='middle'
            className='overflow-hidden'
            empty={
              <Empty
                image={<IllustrationNoResult style={{ width: 150, height: 150 }} />}
                darkModeImage={
                  <IllustrationNoResultDark style={{ width: 150, height: 150 }} />
                }
                description={t('暂无记录')}
                style={{ padding: 30 }}
              />
            }
          />
        )}
      </Card>
    </div>
  );
};

export default AffiliateAdminPage;
