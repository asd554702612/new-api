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
import { Gift } from 'lucide-react';
import { paymentOrdersApi } from '../../helpers/paymentOrders';

const { Text, Title } = Typography;

const formatQuota = (value) => Number(value || 0).toLocaleString();

const LuckyWheel = () => {
  const { t } = useTranslation();
  const [loading, setLoading] = useState(false);
  const [drawing, setDrawing] = useState(false);
  const [summary, setSummary] = useState(null);
  const [result, setResult] = useState(null);

  const load = async () => {
    setLoading(true);
    try {
      const res = await paymentOrdersApi.getLuckyWheelSummary();
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

  const activeSession = summary?.active_session;
  const pendingSessions = useMemo(
    () => summary?.pending_sessions || [],
    [summary],
  );
  const historySessions = useMemo(
    () => summary?.history_sessions || [],
    [summary],
  );

  const draw = async (sessionId) => {
    setDrawing(true);
    try {
      const res = await paymentOrdersApi.drawLuckyWheel(sessionId);
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
              {t('转盘活动')}
            </Title>
            <Text type='tertiary'>
              {summary?.config?.intro_text || t('完成支付后可获得转盘机会。')}
            </Text>
          </div>
          <Tag color={summary?.enabled ? 'green' : 'grey'}>
            {summary?.enabled ? t('已开启') : t('未开启')}
          </Tag>
        </div>

        <div className='grid grid-cols-1 lg:grid-cols-3 gap-4'>
          <section className='lg:col-span-2 rounded border border-semi-color-border bg-semi-color-bg-2 p-4'>
            <div className='flex items-center justify-between gap-3 mb-4'>
              <div>
                <Text type='tertiary'>{t('当前可抽机会')}</Text>
                <Title heading={4} className='!my-1'>
                  {activeSession
                    ? `${activeSession.remaining_draws || 0} ${t('次')}`
                    : t('暂无机会')}
                </Title>
              </div>
              <Gift size={32} />
            </div>
            {activeSession ? (
              <Space wrap>
                <Tag>{activeSession.matched_tier_name}</Tag>
                <Tag>
                  {t('实付金额')} ¥
                  {Number(activeSession.source_pay_amount || 0).toFixed(2)}
                </Tag>
                <Tag>
                  {t('倍率范围')} {activeSession.min_multiplier}-
                  {activeSession.max_multiplier}
                </Tag>
                <Button
                  type='primary'
                  loading={drawing}
                  onClick={() => draw(activeSession.id)}
                >
                  {t('立即抽奖')}
                </Button>
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

        <Card title={t('待使用机会')} className='mt-4'>
          <Table
            size='small'
            pagination={false}
            dataSource={pendingSessions}
            rowKey='id'
            columns={[
              { title: t('订单号'), dataIndex: 'source_order_trade_no' },
              { title: t('档位'), dataIndex: 'matched_tier_name' },
              {
                title: t('剩余次数'),
                dataIndex: 'remaining_draws',
              },
              {
                title: t('最佳倍率'),
                dataIndex: 'best_multiplier',
              },
              {
                title: t('操作'),
                render: (_, record) => (
                  <Button
                    size='small'
                    disabled={(record.remaining_draws || 0) <= 0}
                    loading={drawing && activeSession?.id === record.id}
                    onClick={() => draw(record.id)}
                  >
                    {t('抽奖')}
                  </Button>
                ),
              },
            ]}
          />
        </Card>

        <Card title={t('历史记录')} className='mt-4'>
          <Table
            size='small'
            pagination={false}
            dataSource={historySessions}
            rowKey='id'
            columns={[
              { title: t('订单号'), dataIndex: 'source_order_trade_no' },
              { title: t('最佳倍率'), dataIndex: 'best_multiplier' },
              {
                title: t('奖励额度'),
                dataIndex: 'settled_bonus_quota',
                render: formatQuota,
              },
              {
                title: t('状态'),
                render: (_, record) =>
                  record.settled ? (
                    <Tag color='green'>{t('已结算')}</Tag>
                  ) : (
                    <Tag>{t('待抽奖')}</Tag>
                  ),
              },
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
        {result && (
          <Space vertical align='start'>
            <Title heading={4}>
              {t('本次倍率')} {result.draw_record?.final_multiplier}
            </Title>
            <Text>
              {t('当前最佳倍率')} {result.best_multiplier}
            </Text>
            <Text>
              {t('剩余次数')} {result.remaining_draws}
            </Text>
            {result.settled && (
              <Text>
                {t('已结算奖励额度')} {formatQuota(result.settled_bonus_quota)}
              </Text>
            )}
          </Space>
        )}
      </Modal>
    </div>
  );
};

export default LuckyWheel;
