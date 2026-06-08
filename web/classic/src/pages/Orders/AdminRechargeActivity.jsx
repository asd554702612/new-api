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
  Button,
  Card,
  Checkbox,
  Input,
  InputNumber,
  Modal,
  Space,
  Spin,
  Switch,
  Table,
  Tag,
  TextArea,
  Toast,
  Typography,
} from '@douyinfe/semi-ui';
import { useTranslation } from 'react-i18next';
import { paymentOrdersApi } from '../../helpers/paymentOrders';

const { Text, Title } = Typography;

const fallbackConfig = {
  eligible_order_types: ['balance', 'subscription'],
  intro_text: '完成支付后可获得充值活动抽奖机会。',
  rules_title: '活动规则',
  rules_items: ['支付完成后获得一次抽奖机会。', '中奖后由管理员人工发放。'],
  prizes: [
    {
      id: 'default',
      name: '默认奖品',
      reward_amount: 0,
      reward_description: '请联系管理员领取奖励',
      probability: 100,
      min_pay_amount: 0,
      enabled: true,
      sort_order: 1,
    },
  ],
};

const numberValue = (value, fallback = 0) =>
  value === undefined || value === null || Number.isNaN(Number(value))
    ? fallback
    : Number(value);

const AdminRechargeActivity = () => {
  const { t } = useTranslation();
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [enabled, setEnabled] = useState(false);
  const [config, setConfig] = useState(fallbackConfig);
  const [stats, setStats] = useState({});
  const [page, setPage] = useState(1);
  const [fulfillmentRecord, setFulfillmentRecord] = useState(null);
  const [fulfillmentNote, setFulfillmentNote] = useState('');
  const [fulfillmentLoading, setFulfillmentLoading] = useState(false);

  const load = async () => {
    setLoading(true);
    try {
      const [configRes, statsRes] = await Promise.all([
        paymentOrdersApi.getRechargeActivityConfig(),
        paymentOrdersApi.getRechargeActivityStats({ p: page, page_size: 10 }),
      ]);
      if (configRes.data.success) {
        setEnabled(Boolean(configRes.data.data.enabled));
        setConfig({ ...fallbackConfig, ...(configRes.data.data.config || {}) });
      }
      if (statsRes.data.success) {
        setStats(statsRes.data.data || {});
      }
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load();
  }, [page]);

  const updateConfig = (key, value) => {
    setConfig((prev) => ({ ...prev, [key]: value }));
  };

  const updatePrize = (index, field, value) => {
    setConfig((prev) => {
      const prizes = [...(prev.prizes || [])];
      prizes[index] = { ...prizes[index], [field]: value };
      return { ...prev, prizes };
    });
  };

  const addPrize = () => {
    setConfig((prev) => ({
      ...prev,
      prizes: [
        ...(prev.prizes || []),
        {
          id: `prize_${Date.now()}`,
          name: t('新奖品'),
          reward_amount: 0,
          reward_description: '',
          probability: 0,
          min_pay_amount: 0,
          enabled: true,
          sort_order: (prev.prizes || []).length + 1,
        },
      ],
    }));
  };

  const removePrize = (index) => {
    setConfig((prev) => ({
      ...prev,
      prizes: (prev.prizes || []).filter((_, i) => i !== index),
    }));
  };

  const toggleOrderType = (orderType, checked) => {
    const current = new Set(config.eligible_order_types || []);
    if (checked) {
      current.add(orderType);
    } else {
      current.delete(orderType);
    }
    updateConfig('eligible_order_types', Array.from(current));
  };

  const save = async () => {
    setSaving(true);
    try {
      const payload = {
        ...config,
        rules_items: String(config.rules_items_text || config.rules_items?.join('\n') || '')
          .split('\n')
          .map((item) => item.trim())
          .filter(Boolean),
      };
      delete payload.rules_items_text;
      const res = await paymentOrdersApi.updateRechargeActivityConfig({
        enabled,
        config: payload,
      });
      if (res.data.success) {
        Toast.success({ content: t('保存成功') });
        await load();
      } else {
        Toast.error({ content: res.data.message || t('保存失败') });
      }
    } finally {
      setSaving(false);
    }
  };

  const openFulfillment = (record) => {
    setFulfillmentRecord(record);
    setFulfillmentNote(record.fulfillment_note || '');
  };

  const submitFulfillment = async (status) => {
    setFulfillmentLoading(true);
    try {
      const res = await paymentOrdersApi.updateRechargeActivityFulfillment(
        fulfillmentRecord.id,
        { status, note: fulfillmentNote },
      );
      if (res.data.success) {
        Toast.success({ content: t('操作成功') });
        setFulfillmentRecord(null);
        await load();
      } else {
        Toast.error({ content: res.data.message || t('操作失败') });
      }
    } finally {
      setFulfillmentLoading(false);
    }
  };

  const enabledProbability = (config.prizes || [])
    .filter((prize) => prize.enabled)
    .reduce((sum, prize) => sum + numberValue(prize.probability), 0);

  const statCards = [
    [t('总机会'), stats.total_chances || 0],
    [t('待抽奖'), stats.pending_chances || 0],
    [t('已抽奖'), stats.drawn_chances || 0],
    [t('待发放'), stats.pending_fulfillments || 0],
  ];

  return (
    <div className='mt-[60px] px-2'>
      <Spin spinning={loading}>
        <div className='mb-4 flex items-center justify-between gap-3'>
          <div>
            <Title heading={3} className='!mb-1'>
              {t('充值活动')}
            </Title>
            <Text type='tertiary'>{t('充值奖励和人工履约活动配置')}</Text>
          </div>
          <Space>
            <Switch checked={enabled} onChange={setEnabled} />
            <Button type='primary' loading={saving} onClick={save}>
              {t('保存')}
            </Button>
          </Space>
        </div>

        <div className='grid grid-cols-1 md:grid-cols-4 gap-3 mb-4'>
          {statCards.map(([label, value]) => (
            <section
              key={label}
              className='rounded border border-semi-color-border bg-semi-color-bg-2 p-4'
            >
              <Text type='tertiary'>{label}</Text>
              <Title heading={4} className='!my-1'>
                {value}
              </Title>
            </section>
          ))}
        </div>

        <Card title={t('基础配置')}>
          <div className='grid grid-cols-1 md:grid-cols-2 gap-4'>
            <label>
              <Text>{t('活动说明')}</Text>
              <Input
                value={config.intro_text}
                onChange={(value) => updateConfig('intro_text', value)}
              />
            </label>
            <label>
              <Text>{t('规则标题')}</Text>
              <Input
                value={config.rules_title}
                onChange={(value) => updateConfig('rules_title', value)}
              />
            </label>
          </div>
          <Space className='mt-4'>
            <Checkbox
              checked={(config.eligible_order_types || []).includes('balance')}
              onChange={(event) => toggleOrderType('balance', event.target.checked)}
            >
              {t('充值订单')}
            </Checkbox>
            <Checkbox
              checked={(config.eligible_order_types || []).includes('subscription')}
              onChange={(event) => toggleOrderType('subscription', event.target.checked)}
            >
              {t('订阅订单')}
            </Checkbox>
          </Space>
          <div className='mt-4'>
            <Text>{t('规则内容')}</Text>
            <TextArea
              autosize={{ minRows: 3 }}
              value={config.rules_items_text ?? (config.rules_items || []).join('\n')}
              onChange={(value) => updateConfig('rules_items_text', value)}
            />
          </div>
        </Card>

        <Card
          title={`${t('奖品配置')} (${t('启用概率合计')} ${enabledProbability}%)`}
          className='mt-4'
          headerExtraContent={<Button onClick={addPrize}>{t('新增奖品')}</Button>}
        >
          <Table
            size='small'
            pagination={false}
            dataSource={config.prizes || []}
            rowKey='id'
            columns={[
              {
                title: t('启用'),
                render: (_, record, index) => (
                  <Switch
                    checked={Boolean(record.enabled)}
                    onChange={(value) => updatePrize(index, 'enabled', value)}
                  />
                ),
              },
              {
                title: t('奖品ID'),
                render: (_, record, index) => (
                  <Input value={record.id} onChange={(value) => updatePrize(index, 'id', value)} />
                ),
              },
              {
                title: t('名称'),
                render: (_, record, index) => (
                  <Input
                    value={record.name}
                    onChange={(value) => updatePrize(index, 'name', value)}
                  />
                ),
              },
              {
                title: t('概率'),
                render: (_, record, index) => (
                  <InputNumber
                    value={record.probability}
                    onChange={(value) => updatePrize(index, 'probability', numberValue(value))}
                  />
                ),
              },
              {
                title: t('最低实付'),
                render: (_, record, index) => (
                  <InputNumber
                    value={record.min_pay_amount}
                    onChange={(value) => updatePrize(index, 'min_pay_amount', numberValue(value))}
                  />
                ),
              },
              {
                title: t('奖励说明'),
                render: (_, record, index) => (
                  <Input
                    value={record.reward_description}
                    onChange={(value) =>
                      updatePrize(index, 'reward_description', value)
                    }
                  />
                ),
              },
              {
                title: t('操作'),
                render: (_, __, index) => (
                  <Button type='danger' onClick={() => removePrize(index)}>
                    {t('删除')}
                  </Button>
                ),
              },
            ]}
          />
        </Card>

        <Card title={t('中奖记录')} className='mt-4'>
          <Table
            size='small'
            dataSource={stats.recent_records || []}
            rowKey='id'
            pagination={{
              currentPage: page,
              pageSize: stats.recent_records_page_size || 10,
              total: stats.recent_records_total || 0,
              onPageChange: setPage,
            }}
            columns={[
              { title: t('用户ID'), dataIndex: 'user_id' },
              { title: t('用户'), dataIndex: 'user_name' },
              { title: t('奖品名称'), dataIndex: 'prize_name' },
              { title: t('奖励说明'), dataIndex: 'reward_description' },
              {
                title: t('状态'),
                dataIndex: 'fulfillment_status',
                render: (value) =>
                  value === 'fulfilled' ? (
                    <Tag color='green'>{t('已发放')}</Tag>
                  ) : (
                    <Tag color='orange'>{t('待发放')}</Tag>
                  ),
              },
              {
                title: t('操作'),
                render: (_, record) => (
                  <Button size='small' onClick={() => openFulfillment(record)}>
                    {t('履约')}
                  </Button>
                ),
              },
            ]}
          />
        </Card>
      </Spin>

      <Modal
        title={t('人工发放')}
        visible={Boolean(fulfillmentRecord)}
        onCancel={() => setFulfillmentRecord(null)}
        footer={
          <Space>
            <Button
              loading={fulfillmentLoading}
              onClick={() => submitFulfillment('pending')}
            >
              {t('标记待发放')}
            </Button>
            <Button
              type='primary'
              loading={fulfillmentLoading}
              onClick={() => submitFulfillment('fulfilled')}
            >
              {t('标记已发放')}
            </Button>
          </Space>
        }
      >
        <Space vertical align='start' style={{ width: '100%' }}>
          <Text>{fulfillmentRecord?.prize_name}</Text>
          <TextArea
            autosize={{ minRows: 3 }}
            value={fulfillmentNote}
            placeholder={t('填写发放备注')}
            onChange={setFulfillmentNote}
          />
        </Space>
      </Modal>
    </div>
  );
};

export default AdminRechargeActivity;
