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
  multiplier_step: 0.1,
  global_max_multiplier: 2,
  intro_text: '完成支付后可获得转盘机会。',
  rules_title: '活动规则',
  rules_items: ['支付完成后按实付金额获得抽奖机会。', '多次抽奖取最高倍率结算。'],
  amount_tiers: [
    {
      id: 'default',
      name: '默认档位',
      min_amount: 0,
      max_amount: null,
      min_multiplier: 0.1,
      max_multiplier: 1,
      draw_count: 1,
    },
  ],
  invite_bonus: {
    enabled: false,
    qualifying_amount: 20,
    bonus_per_invitee: 0.2,
    max_bonus: 1,
    consume_policy: 'next_session_once',
  },
  golden_window: {
    enabled: false,
    timezone: 'Asia/Shanghai',
    start_time: '20:00',
    end_time: '22:00',
    min_amount: 50,
    extra_draws: 1,
    daily_quota: 5,
  },
};

const numberValue = (value, fallback = 0) =>
  value === undefined || value === null || Number.isNaN(Number(value))
    ? fallback
    : Number(value);

const AdminLuckyWheel = () => {
  const { t } = useTranslation();
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [enabled, setEnabled] = useState(false);
  const [config, setConfig] = useState(fallbackConfig);
  const [stats, setStats] = useState({});

  const load = async () => {
    setLoading(true);
    try {
      const [configRes, statsRes] = await Promise.all([
        paymentOrdersApi.getLuckyWheelConfig(),
        paymentOrdersApi.getLuckyWheelStats(),
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
  }, []);

  const updateConfig = (key, value) => {
    setConfig((prev) => ({ ...prev, [key]: value }));
  };

  const updateNested = (key, field, value) => {
    setConfig((prev) => ({
      ...prev,
      [key]: { ...(prev[key] || {}), [field]: value },
    }));
  };

  const updateTier = (index, field, value) => {
    setConfig((prev) => {
      const amount_tiers = [...(prev.amount_tiers || [])];
      amount_tiers[index] = { ...amount_tiers[index], [field]: value };
      return { ...prev, amount_tiers };
    });
  };

  const addTier = () => {
    setConfig((prev) => ({
      ...prev,
      amount_tiers: [
        ...(prev.amount_tiers || []),
        {
          id: `tier_${Date.now()}`,
          name: t('新档位'),
          min_amount: 0,
          max_amount: null,
          min_multiplier: 0.1,
          max_multiplier: 1,
          draw_count: 1,
        },
      ],
    }));
  };

  const removeTier = (index) => {
    setConfig((prev) => ({
      ...prev,
      amount_tiers: (prev.amount_tiers || []).filter((_, i) => i !== index),
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
      const res = await paymentOrdersApi.updateLuckyWheelConfig({
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

  const statCards = [
    [t('总会话'), stats.total_sessions || 0],
    [t('待抽奖'), stats.pending_sessions || 0],
    [t('已结算'), stats.settled_sessions || 0],
    [t('已发奖励额度'), Number(stats.total_bonus_quota || 0).toLocaleString()],
  ];

  return (
    <div className='mt-[60px] px-2'>
      <Spin spinning={loading}>
        <div className='mb-4 flex items-center justify-between gap-3'>
          <div>
            <Title heading={3} className='!mb-1'>
              {t('转盘活动')}
            </Title>
            <Text type='tertiary'>{t('支付完成后的转盘机会配置')}</Text>
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
            <label>
              <Text>{t('倍率步长')}</Text>
              <InputNumber
                value={config.multiplier_step}
                onChange={(value) => updateConfig('multiplier_step', numberValue(value, 0.1))}
              />
            </label>
            <label>
              <Text>{t('全局最大倍率')}</Text>
              <InputNumber
                value={config.global_max_multiplier}
                onChange={(value) => updateConfig('global_max_multiplier', numberValue(value, 2))}
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
          title={t('金额档位')}
          className='mt-4'
          headerExtraContent={<Button onClick={addTier}>{t('新增档位')}</Button>}
        >
          <Table
            size='small'
            pagination={false}
            dataSource={config.amount_tiers || []}
            rowKey='id'
            columns={[
              {
                title: t('档位ID'),
                render: (_, record, index) => (
                  <Input value={record.id} onChange={(value) => updateTier(index, 'id', value)} />
                ),
              },
              {
                title: t('名称'),
                render: (_, record, index) => (
                  <Input
                    value={record.name}
                    onChange={(value) => updateTier(index, 'name', value)}
                  />
                ),
              },
              {
                title: t('最低金额'),
                render: (_, record, index) => (
                  <InputNumber
                    value={record.min_amount}
                    onChange={(value) => updateTier(index, 'min_amount', numberValue(value))}
                  />
                ),
              },
              {
                title: t('最高金额'),
                render: (_, record, index) => (
                  <InputNumber
                    value={record.max_amount}
                    placeholder={t('不限')}
                    onChange={(value) =>
                      updateTier(index, 'max_amount', value === null ? null : numberValue(value))
                    }
                  />
                ),
              },
              {
                title: t('倍率范围'),
                render: (_, record, index) => (
                  <Space>
                    <InputNumber
                      value={record.min_multiplier}
                      onChange={(value) =>
                        updateTier(index, 'min_multiplier', numberValue(value))
                      }
                    />
                    <InputNumber
                      value={record.max_multiplier}
                      onChange={(value) =>
                        updateTier(index, 'max_multiplier', numberValue(value))
                      }
                    />
                  </Space>
                ),
              },
              {
                title: t('次数'),
                render: (_, record, index) => (
                  <InputNumber
                    value={record.draw_count}
                    onChange={(value) => updateTier(index, 'draw_count', numberValue(value, 1))}
                  />
                ),
              },
              {
                title: t('操作'),
                render: (_, __, index) => (
                  <Button type='danger' onClick={() => removeTier(index)}>
                    {t('删除')}
                  </Button>
                ),
              },
            ]}
          />
        </Card>

        <div className='grid grid-cols-1 lg:grid-cols-2 gap-4 mt-4'>
          <Card title={t('邀请加成')}>
            <Space vertical align='start'>
              <Switch
                checked={Boolean(config.invite_bonus?.enabled)}
                onChange={(value) => updateNested('invite_bonus', 'enabled', value)}
              />
              <InputNumber
                prefix={t('合格实付')}
                value={config.invite_bonus?.qualifying_amount}
                onChange={(value) =>
                  updateNested('invite_bonus', 'qualifying_amount', numberValue(value))
                }
              />
              <InputNumber
                prefix={t('每人加成')}
                value={config.invite_bonus?.bonus_per_invitee}
                onChange={(value) =>
                  updateNested('invite_bonus', 'bonus_per_invitee', numberValue(value))
                }
              />
              <InputNumber
                prefix={t('加成上限')}
                value={config.invite_bonus?.max_bonus}
                onChange={(value) =>
                  updateNested('invite_bonus', 'max_bonus', numberValue(value))
                }
              />
            </Space>
          </Card>

          <Card title={t('黄金窗口')}>
            <Space vertical align='start'>
              <Switch
                checked={Boolean(config.golden_window?.enabled)}
                onChange={(value) => updateNested('golden_window', 'enabled', value)}
              />
              <Input
                prefix={t('时区')}
                value={config.golden_window?.timezone}
                onChange={(value) => updateNested('golden_window', 'timezone', value)}
              />
              <Space>
                <Input
                  prefix={t('开始')}
                  value={config.golden_window?.start_time}
                  onChange={(value) => updateNested('golden_window', 'start_time', value)}
                />
                <Input
                  prefix={t('结束')}
                  value={config.golden_window?.end_time}
                  onChange={(value) => updateNested('golden_window', 'end_time', value)}
                />
              </Space>
              <InputNumber
                prefix={t('最低实付')}
                value={config.golden_window?.min_amount}
                onChange={(value) =>
                  updateNested('golden_window', 'min_amount', numberValue(value))
                }
              />
              <InputNumber
                prefix={t('额外次数')}
                value={config.golden_window?.extra_draws}
                onChange={(value) =>
                  updateNested('golden_window', 'extra_draws', numberValue(value, 1))
                }
              />
              <InputNumber
                prefix={t('每日名额')}
                value={config.golden_window?.daily_quota}
                onChange={(value) =>
                  updateNested('golden_window', 'daily_quota', numberValue(value, 1))
                }
              />
            </Space>
          </Card>
        </div>

        <Card title={t('最近会话')} className='mt-4'>
          <Table
            size='small'
            pagination={false}
            dataSource={stats.recent_sessions || []}
            rowKey='id'
            columns={[
              { title: t('用户ID'), dataIndex: 'user_id' },
              { title: t('订单号'), dataIndex: 'source_order_trade_no' },
              { title: t('档位'), dataIndex: 'matched_tier_name' },
              { title: t('最佳倍率'), dataIndex: 'best_multiplier' },
              {
                title: t('状态'),
                render: (_, record) =>
                  record.settled ? (
                    <Tag color='green'>{t('已结算')}</Tag>
                  ) : (
                    <Tag color='orange'>{t('待抽奖')}</Tag>
                  ),
              },
            ]}
          />
        </Card>
      </Spin>
    </div>
  );
};

export default AdminLuckyWheel;
