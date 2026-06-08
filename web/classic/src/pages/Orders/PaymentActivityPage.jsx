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
import { Button, Card, Form, Spin, Toast, Typography } from '@douyinfe/semi-ui';
import { useTranslation } from 'react-i18next';
import { paymentOrdersApi } from '../../helpers/paymentOrders';

const { Text } = Typography;

const defaultConfig = {
  lucky_wheel: {
    eligible_order_types: ['balance', 'subscription'],
    tiers: [{ id: 'default', name: 'default', min_amount: 0, chances: 1 }],
    prizes: [],
  },
  recharge_activity: {
    first_recharge_enabled: false,
    member_level_enabled: false,
    tiers: [{ id: 'default', name: 'default', pay_amount: 0, bonus_amount: 0 }],
    member_levels: [],
  },
};

const PaymentActivityPage = ({ activityType, title, description }) => {
  const { t } = useTranslation();
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [enabled, setEnabled] = useState(false);
  const [configText, setConfigText] = useState('');
  const [stats, setStats] = useState({});

  const fallbackConfig = useMemo(
    () => defaultConfig[activityType] || {},
    [activityType],
  );

  const load = async () => {
    setLoading(true);
    try {
      const [configRes, statsRes] = await Promise.all([
        paymentOrdersApi.getActivityConfig(activityType),
        paymentOrdersApi.getActivityStats(activityType),
      ]);
      if (configRes.data.success) {
        setEnabled(Boolean(configRes.data.data.enabled));
        setConfigText(
          JSON.stringify(configRes.data.data.config || fallbackConfig, null, 2),
        );
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
  }, [activityType]);

  const save = async () => {
    let config;
    try {
      config = JSON.parse(configText || '{}');
    } catch (error) {
      Toast.error({ content: t('JSON 格式错误') });
      return;
    }
    setSaving(true);
    try {
      const res = await paymentOrdersApi.updateActivityConfig(activityType, {
        enabled,
        config,
      });
      if (res.data.success) {
        Toast.success({ content: t('保存成功') });
        load();
      } else {
        Toast.error({ content: res.data.message || t('保存失败') });
      }
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className='mt-[60px] px-2'>
      <Card title={t(title)}>
        <Spin spinning={loading}>
          <Text type='tertiary'>{t(description)}</Text>
          <div className='grid grid-cols-1 md:grid-cols-4 gap-3 my-4'>
            <Card>
              <Text type='tertiary'>{t('总机会')}</Text>
              <div className='text-2xl font-semibold'>{stats.total_chances || 0}</div>
            </Card>
            <Card>
              <Text type='tertiary'>{t('待使用机会')}</Text>
              <div className='text-2xl font-semibold'>{stats.pending_chances || 0}</div>
            </Card>
            <Card>
              <Text type='tertiary'>{t('已参与订单')}</Text>
              <div className='text-2xl font-semibold'>{stats.drawn_chances || 0}</div>
            </Card>
            <Card>
              <Text type='tertiary'>{t('待履约')}</Text>
              <div className='text-2xl font-semibold'>
                {stats.pending_fulfillments || 0}
              </div>
            </Card>
          </div>
          <Form layout='vertical'>
            <Form.Switch
              label={t('启用活动')}
              field='enabled'
              checked={enabled}
              onChange={setEnabled}
            />
            <Form.TextArea
              label={t('活动配置 JSON')}
              field='config'
              value={configText}
              autosize={{ minRows: 14 }}
              onChange={setConfigText}
            />
          </Form>
          <Button type='primary' loading={saving} onClick={save}>
            {t('保存')}
          </Button>
        </Spin>
      </Card>
    </div>
  );
};

export default PaymentActivityPage;
