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
  Empty,
  Modal,
  Space,
  Spin,
  Table,
  Tag,
  Toast,
  Typography,
} from '@douyinfe/semi-ui';
import { useTranslation } from 'react-i18next';
import { paymentOrdersApi } from '../../helpers/paymentOrders';

const { Text, Title } = Typography;

const RechargeActivity = () => {
  const { t } = useTranslation();
  const [loading, setLoading] = useState(false);
  const [drawing, setDrawing] = useState(false);
  const [summary, setSummary] = useState(null);
  const [result, setResult] = useState(null);

  const load = async () => {
    setLoading(true);
    try {
      const res = await paymentOrdersApi.getRechargeActivitySummary();
      if (res.data.success) {
        setSummary(res.data.data);
      } else {
        Toast.error({ content: res.data.message || t('加载失败') });
      }
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load();
  }, []);

  const pendingChances = useMemo(
    () => summary?.pending_chances || [],
    [summary],
  );
  const historyRecords = useMemo(
    () => summary?.history_records || [],
    [summary],
  );

  const draw = async (chanceId) => {
    setDrawing(true);
    try {
      const res = await paymentOrdersApi.drawRechargeActivity(chanceId);
      if (res.data.success) {
        setResult(res.data.data);
        await load();
      } else {
        Toast.error({ content: res.data.message || t('抽奖失败') });
      }
    } finally {
      setDrawing(false);
    }
  };

  return (
    <div className='mt-[60px] px-2'>
      <Spin spinning={loading}>
        <div className='mb-4 flex flex-col md:flex-row md:items-center md:justify-between gap-3'>
          <div>
            <Title heading={3} className='!mb-1'>
              {t('充值活动')}
            </Title>
            <Text type='tertiary'>
              {summary?.config?.intro_text || t('完成支付后可获得充值活动抽奖机会。')}
            </Text>
          </div>
          <Tag color={summary?.enabled ? 'green' : 'grey'}>
            {summary?.enabled ? t('已开启') : t('未开启')}
          </Tag>
        </div>

        <div className='grid grid-cols-1 lg:grid-cols-3 gap-4'>
          <section className='lg:col-span-2 rounded border border-semi-color-border bg-semi-color-bg-2 p-4'>
            <Text type='tertiary'>{t('待抽奖机会')}</Text>
            <Title heading={4} className='!my-1'>
              {pendingChances.length}
            </Title>
            {pendingChances.length > 0 ? (
              <Space wrap>
                <Button
                  type='primary'
                  loading={drawing}
                  onClick={() => draw(pendingChances[0].id)}
                >
                  {t('立即抽奖')}
                </Button>
                <Text type='tertiary'>
                  {t('将使用最早的一次待抽奖机会')}
                </Text>
              </Space>
            ) : (
              <Empty description={t('支付完成后会在这里显示抽奖机会')} />
            )}
          </section>

          <section className='rounded border border-semi-color-border bg-semi-color-bg-2 p-4'>
            <Title heading={5}>{summary?.config?.rules_title || t('活动规则')}</Title>
            {(summary?.config?.rules_items || []).map((item) => (
              <div key={item} className='mt-2'>
                <Text>{item}</Text>
              </div>
            ))}
          </section>
        </div>

        <Card title={t('奖品列表')} className='mt-4'>
          <Table
            size='small'
            pagination={false}
            dataSource={summary?.config?.prizes || []}
            rowKey='id'
            columns={[
              { title: t('奖品名称'), dataIndex: 'name' },
              {
                title: t('最低实付金额'),
                dataIndex: 'min_pay_amount',
                render: (value) => `¥${Number(value || 0).toFixed(2)}`,
              },
              {
                title: t('概率'),
                dataIndex: 'probability',
                render: (value) => `${value}%`,
              },
              { title: t('奖励说明'), dataIndex: 'reward_description' },
            ]}
          />
        </Card>

        <Card title={t('待抽奖机会')} className='mt-4'>
          <Table
            size='small'
            pagination={false}
            dataSource={pendingChances}
            rowKey='id'
            columns={[
              { title: t('订单号'), dataIndex: 'source_order_trade_no' },
              {
                title: t('实付金额'),
                dataIndex: 'source_pay_amount',
                render: (value) => `¥${Number(value || 0).toFixed(2)}`,
              },
              {
                title: t('操作'),
                render: (_, record) => (
                  <Button size='small' loading={drawing} onClick={() => draw(record.id)}>
                    {t('抽奖')}
                  </Button>
                ),
              },
            ]}
          />
        </Card>

        <Card title={t('中奖记录')} className='mt-4'>
          <Table
            size='small'
            pagination={false}
            dataSource={historyRecords}
            rowKey='id'
            columns={[
              { title: t('奖品名称'), dataIndex: 'prize_name' },
              { title: t('奖励说明'), dataIndex: 'reward_description' },
              {
                title: t('履约状态'),
                dataIndex: 'fulfillment_status',
                render: (value) =>
                  value === 'fulfilled' ? (
                    <Tag color='green'>{t('已发放')}</Tag>
                  ) : (
                    <Tag color='orange'>{t('待发放')}</Tag>
                  ),
              },
              { title: t('备注'), dataIndex: 'fulfillment_note' },
            ]}
          />
        </Card>
      </Spin>

      <Modal
        title={t('抽奖结果')}
        visible={Boolean(result)}
        onCancel={() => setResult(null)}
        footer={<Button onClick={() => setResult(null)}>{t('关闭')}</Button>}
      >
        {result?.record && (
          <Space vertical align='start'>
            <Title heading={4}>{result.record.prize_name}</Title>
            <Text>{result.record.reward_description}</Text>
            <Tag color='orange'>{t('等待管理员发放')}</Tag>
          </Space>
        )}
      </Modal>
    </div>
  );
};

export default RechargeActivity;
