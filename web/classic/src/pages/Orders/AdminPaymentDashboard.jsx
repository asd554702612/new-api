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
import { Card, Col, Row, Select, Spin, Table, Typography } from '@douyinfe/semi-ui';
import { useTranslation } from 'react-i18next';
import { paymentOrdersApi } from '../../helpers/paymentOrders';

const { Text, Title } = Typography;

const formatMoney = (value) => `¥${Number(value || 0).toFixed(2)}`;

const StatCard = ({ title, value, extra }) => (
  <Card>
    <Text type='tertiary'>{title}</Text>
    <Title heading={3} className='mt-2 mb-1'>
      {value}
    </Title>
    {extra && <Text type='tertiary'>{extra}</Text>}
  </Card>
);

const AdminPaymentDashboard = () => {
  const { t } = useTranslation();
  const [days, setDays] = useState(30);
  const [loading, setLoading] = useState(false);
  const [stats, setStats] = useState(null);

  useEffect(() => {
    const load = async () => {
      setLoading(true);
      try {
        const res = await paymentOrdersApi.dashboard(days);
        if (res.data.success) {
          setStats(res.data.data);
        }
      } finally {
        setLoading(false);
      }
    };
    load();
  }, [days]);

  const dailyRows = useMemo(() => stats?.daily_series || [], [stats]);

  return (
    <div className='mt-[60px] px-2'>
      <Card
        title={t('支付概览')}
        headerExtraContent={
          <Select
            value={days}
            style={{ width: 140 }}
            onChange={setDays}
            optionList={[
              { label: t('近7天'), value: 7 },
              { label: t('近30天'), value: 30 },
              { label: t('近90天'), value: 90 },
            ]}
          />
        }
      >
        <Spin spinning={loading}>
          <Row gutter={[16, 16]}>
            <Col xs={24} sm={12} lg={6}>
              <StatCard title={t('订单数')} value={stats?.total_orders || 0} />
            </Col>
            <Col xs={24} sm={12} lg={6}>
              <StatCard title={t('支付金额')} value={formatMoney(stats?.total_amount)} />
            </Col>
            <Col xs={24} sm={12} lg={6}>
              <StatCard
                title={t('完成率')}
                value={`${((stats?.completion_rate || 0) * 100).toFixed(1)}%`}
              />
            </Col>
            <Col xs={24} sm={12} lg={6}>
              <StatCard
                title={t('待支付')}
                value={stats?.pending_orders || 0}
                extra={`${t('失败')} ${stats?.failed_orders || 0}`}
              />
            </Col>
          </Row>
          <Row gutter={[16, 16]} className='mt-4'>
            <Col xs={24} lg={12}>
              <Card title={t('支付方式分布')}>
                <Table
                  size='small'
                  pagination={false}
                  dataSource={stats?.payment_methods || []}
                  rowKey='type'
                  columns={[
                    { title: t('支付方式'), dataIndex: 'type' },
                    { title: t('订单数'), dataIndex: 'count' },
                    { title: t('金额'), dataIndex: 'amount', render: formatMoney },
                  ]}
                />
              </Card>
            </Col>
            <Col xs={24} lg={12}>
              <Card title={t('Top 用户')}>
                <Table
                  size='small'
                  pagination={false}
                  dataSource={stats?.top_users || []}
                  rowKey='user_id'
                  columns={[
                    { title: t('用户ID'), dataIndex: 'user_id' },
                    { title: t('订单数'), dataIndex: 'count' },
                    { title: t('金额'), dataIndex: 'amount', render: formatMoney },
                  ]}
                />
              </Card>
            </Col>
          </Row>
          <Card title={t('日收入趋势')} className='mt-4'>
            <Table
              size='small'
              pagination={false}
              dataSource={dailyRows}
              rowKey='date'
              columns={[
                { title: t('日期'), dataIndex: 'date' },
                { title: t('订单数'), dataIndex: 'count' },
                { title: t('金额'), dataIndex: 'amount', render: formatMoney },
              ]}
            />
          </Card>
        </Spin>
      </Card>
    </div>
  );
};

export default AdminPaymentDashboard;
