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
  Col,
  Empty,
  Radio,
  Row,
  Space,
  Table,
  TabPane,
  Tabs,
  Typography,
} from '@douyinfe/semi-ui';
import { IconRefresh } from '@douyinfe/semi-icons';
import { useTranslation } from 'react-i18next';
import { API, renderQuota, showError } from '../../helpers';
import { formatCompactNumber } from './utils';

const { Text, Title } = Typography;
const MODEL_PERIODS = ['today', 'week', 'month', 'year', 'all'];
const USER_PERIODS = ['today', 'yesterday'];

const periodLabelKeys = {
  today: '今日',
  yesterday: '昨日',
  week: '本周',
  month: '本月',
  year: '今年',
  all: '全部',
};

const Rankings = () => {
  const { t } = useTranslation();
  const [activeTab, setActiveTab] = useState('models');
  const [modelPeriod, setModelPeriod] = useState('week');
  const [userPeriod, setUserPeriod] = useState('today');
  const [modelsLoading, setModelsLoading] = useState(false);
  const [usersLoading, setUsersLoading] = useState(false);
  const [modelsSnapshot, setModelsSnapshot] = useState(null);
  const [userSnapshot, setUserSnapshot] = useState(null);

  const timezone = useMemo(() => {
    return Intl.DateTimeFormat().resolvedOptions().timeZone || '';
  }, []);

  const loadModels = async () => {
    setModelsLoading(true);
    try {
      const res = await API.get('/api/rankings', {
        params: {
          period: modelPeriod,
        },
      });
      const { success, message, data } = res.data;
      if (!success) {
        showError(message || t('加载排行榜失败'));
        return;
      }
      setModelsSnapshot(data);
    } catch (error) {
      showError(error);
    } finally {
      setModelsLoading(false);
    }
  };

  const loadUsers = async () => {
    setUsersLoading(true);
    try {
      const res = await API.get('/api/rankings/users', {
        params: {
          period: userPeriod,
          timezone,
          limit: 20,
        },
      });
      const { success, message, data } = res.data;
      if (!success) {
        showError(message || t('加载用户消耗榜单失败'));
        return;
      }
      setUserSnapshot(data);
    } catch (error) {
      showError(error);
    } finally {
      setUsersLoading(false);
    }
  };

  useEffect(() => {
    loadModels();
  }, [modelPeriod]);

  useEffect(() => {
    loadUsers();
  }, [userPeriod, timezone]);

  const modelColumns = [
    {
      title: t('排名'),
      dataIndex: 'rank',
      width: 90,
      render: (rank) => <Text strong>#{rank}</Text>,
    },
    {
      title: t('模型'),
      dataIndex: 'model_name',
      render: (modelName, record) => (
        <div>
          <Text strong>{modelName || '-'}</Text>
          <div>
            <Text type='tertiary' size='small'>
              {record.vendor || t('未知')}
            </Text>
          </div>
        </div>
      ),
    },
    {
      title: t('分类'),
      dataIndex: 'category',
      width: 120,
      render: (category) => category || '-',
    },
    {
      title: t('Tokens'),
      dataIndex: 'total_tokens',
      align: 'right',
      width: 140,
      render: (value) => formatCompactNumber(value),
    },
    {
      title: t('占比'),
      dataIndex: 'share',
      align: 'right',
      width: 100,
      render: (value) => `${Number(value || 0).toFixed(1)}%`,
    },
  ];

  const vendorColumns = [
    {
      title: t('排名'),
      dataIndex: 'rank',
      width: 70,
      render: (rank) => <Text strong>#{rank}</Text>,
    },
    {
      title: t('厂商'),
      dataIndex: 'vendor',
      render: (vendor) => vendor || t('未知'),
    },
    {
      title: t('模型数'),
      dataIndex: 'models_count',
      align: 'right',
      width: 88,
    },
    {
      title: t('Tokens'),
      dataIndex: 'total_tokens',
      align: 'right',
      width: 110,
      render: (value) => formatCompactNumber(value),
    },
    {
      title: t('占比'),
      dataIndex: 'share',
      align: 'right',
      width: 76,
      render: (value) => `${Number(value || 0).toFixed(1)}%`,
    },
  ];

  const userColumns = [
    {
      title: t('排名'),
      dataIndex: 'rank',
      width: 90,
      render: (rank) => <Text strong>#{rank}</Text>,
    },
    {
      title: t('用户'),
      dataIndex: 'display_name',
      render: (displayName, record) => (
        <div>
          <Text strong>{displayName || `User #${record.user_id}`}</Text>
          <div>
            <Text type='tertiary' size='small'>
              ID: {record.user_id}
            </Text>
          </div>
        </div>
      ),
    },
    {
      title: t('Tokens'),
      dataIndex: 'tokens',
      align: 'right',
      width: 140,
      render: (value) => formatCompactNumber(value),
    },
    {
      title: t('请求数'),
      dataIndex: 'requests',
      align: 'right',
      width: 120,
      render: (value) => Number(value || 0).toLocaleString(),
    },
    {
      title: t('消耗额度'),
      dataIndex: 'quota',
      align: 'right',
      width: 140,
      render: (value) => renderQuota(value, 6),
    },
  ];

  return (
    <div className='rankings-page mt-[72px] px-4 pb-8'>
      <div style={{ maxWidth: 1200, margin: '0 auto' }}>
        <div style={{ marginBottom: 20 }}>
          <Title heading={2} style={{ marginBottom: 8 }}>
            {t('排行榜')}
          </Title>
          <Text type='secondary'>{t('查看模型、厂商与用户消耗排行')}</Text>
        </div>

        <Tabs type='card' activeKey={activeTab} onChange={setActiveTab}>
          <TabPane itemKey='models' tab={t('模型榜单')}>
            <Card
              title={t('模型与厂商排行榜')}
              headerExtraContent={
                <Space>
                  <Radio.Group
                    type='button'
                    buttonSize='small'
                    value={modelPeriod}
                    onChange={(event) => setModelPeriod(event.target.value)}
                  >
                    {MODEL_PERIODS.map((period) => (
                      <Radio key={period} value={period}>
                        {t(periodLabelKeys[period])}
                      </Radio>
                    ))}
                  </Radio.Group>
                  <Button
                    icon={<IconRefresh />}
                    loading={modelsLoading}
                    onClick={loadModels}
                  >
                    {t('刷新')}
                  </Button>
                </Space>
              }
            >
              <Space
                vertical
                spacing='medium'
                align='stretch'
                style={{ width: '100%' }}
              >
                <RankingTableSection title={t('模型')}>
                  <Table
                    rowKey='model_name'
                    columns={modelColumns}
                    dataSource={modelsSnapshot?.models || []}
                    loading={modelsLoading}
                    pagination={false}
                    scroll={{ x: 720 }}
                    empty={<Empty title={t('暂无榜单数据')} />}
                  />
                </RankingTableSection>
                <RankingTableSection title={t('厂商')}>
                  <Table
                    rowKey='vendor'
                    columns={vendorColumns}
                    dataSource={modelsSnapshot?.vendors || []}
                    loading={modelsLoading}
                    pagination={false}
                    scroll={{ x: 560 }}
                    empty={<Empty title={t('暂无榜单数据')} />}
                  />
                </RankingTableSection>
              </Space>
            </Card>
          </TabPane>

          <TabPane itemKey='users' tab={t('用户消耗榜单')}>
            <Space
              vertical
              spacing='medium'
              align='stretch'
              style={{ width: '100%' }}
            >
              <Row gutter={[16, 16]}>
                <Col xs={24} md={8}>
                  <MetricCard
                    title={t('总 Tokens')}
                    value={formatCompactNumber(userSnapshot?.total_tokens)}
                  />
                </Col>
                <Col xs={24} md={8}>
                  <MetricCard
                    title={t('总请求数')}
                    value={Number(
                      userSnapshot?.total_requests || 0,
                    ).toLocaleString()}
                  />
                </Col>
                <Col xs={24} md={8}>
                  <MetricCard
                    title={t('总消耗额度')}
                    value={renderQuota(userSnapshot?.total_quota || 0, 6)}
                  />
                </Col>
              </Row>

              <Card
                title={t('用户消耗榜单')}
                headerExtraContent={
                  <Space>
                    <Text type='tertiary'>
                      {userSnapshot?.start_date || '-'}
                    </Text>
                    <Radio.Group
                      type='button'
                      buttonSize='small'
                      value={userPeriod}
                      onChange={(event) => setUserPeriod(event.target.value)}
                    >
                      {USER_PERIODS.map((period) => (
                        <Radio key={period} value={period}>
                          {t(periodLabelKeys[period])}
                        </Radio>
                      ))}
                    </Radio.Group>
                    <Button
                      icon={<IconRefresh />}
                      loading={usersLoading}
                      onClick={loadUsers}
                    >
                      {t('刷新')}
                    </Button>
                  </Space>
                }
              >
                <Table
                  rowKey='user_id'
                  columns={userColumns}
                  dataSource={userSnapshot?.ranking || []}
                  loading={usersLoading}
                  pagination={false}
                  scroll={{ x: 'max-content' }}
                  empty={<Empty title={t('暂无榜单数据')} />}
                />
              </Card>
            </Space>
          </TabPane>
        </Tabs>
      </div>
    </div>
  );
};

const MetricCard = ({ title, value }) => {
  return (
    <Card bodyStyle={{ padding: 18 }}>
      <Text type='secondary'>{title}</Text>
      <div style={{ marginTop: 8, fontSize: 24, fontWeight: 700 }}>{value}</div>
    </Card>
  );
};

const RankingTableSection = ({ title, children }) => {
  return (
    <section className='rankings-table-section'>
      <div className='rankings-table-section-title'>
        <Text strong>{title}</Text>
      </div>
      {children}
    </section>
  );
};

export default Rankings;
