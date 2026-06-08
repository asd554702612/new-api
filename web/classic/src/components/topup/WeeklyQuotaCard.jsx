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
  Avatar,
  Button,
  Card,
  Modal,
  Space,
  Spin,
  Tag,
  Typography,
} from '@douyinfe/semi-ui';
import { CheckCircle, Clock, Gift, Phone, TimerReset } from 'lucide-react';
import Turnstile from 'react-turnstile';
import {
  API,
  renderQuota,
  showError,
  showSuccess,
  timestamp2string,
} from '../../helpers';

const { Text } = Typography;

const WeeklyQuotaCard = ({
  t,
  userState,
  userDispatch,
  turnstileEnabled,
  turnstileSiteKey,
}) => {
  const [loading, setLoading] = useState(false);
  const [claiming, setClaiming] = useState(false);
  const [status, setStatus] = useState(null);
  const [turnstileModalVisible, setTurnstileModalVisible] = useState(false);
  const [turnstileWidgetKey, setTurnstileWidgetKey] = useState(0);

  const fetchStatus = async () => {
    setLoading(true);
    try {
      const res = await API.get('/api/user/weekly_quota');
      const { success, data, message } = res.data;
      if (success) {
        setStatus(data);
      } else {
        setStatus(null);
        showError(message || t('获取周额度领取状态失败'));
      }
    } catch (error) {
      setStatus(null);
      showError(t('获取周额度领取状态失败'));
    } finally {
      setLoading(false);
    }
  };

  const shouldTriggerTurnstile = (message) => {
    if (!turnstileEnabled) return false;
    if (typeof message !== 'string') return true;
    return message.includes('Turnstile');
  };

  const postClaim = async (token) => {
    const url = token
      ? `/api/user/weekly_quota?turnstile=${encodeURIComponent(token)}`
      : '/api/user/weekly_quota';
    return API.post(url);
  };

  const claimWeeklyQuota = async (token) => {
    setClaiming(true);
    try {
      const res = await postClaim(token);
      const { success, data, message } = res.data;
      if (success) {
        showSuccess(
          t('周额度领取成功，获得') + ' ' + renderQuota(data.quota_awarded),
        );
        if (userState?.user && typeof data.new_quota === 'number') {
          userDispatch({
            type: 'login',
            payload: {
              ...userState.user,
              quota: data.new_quota,
            },
          });
        }
        setTurnstileModalVisible(false);
        fetchStatus();
      } else {
        if (!token && shouldTriggerTurnstile(message)) {
          if (!turnstileSiteKey) {
            showError('Turnstile is enabled but site key is empty.');
            return;
          }
          setTurnstileModalVisible(true);
          return;
        }
        if (token && shouldTriggerTurnstile(message)) {
          setTurnstileWidgetKey((v) => v + 1);
        }
        showError(message || t('周额度领取失败'));
      }
    } catch (error) {
      showError(t('周额度领取失败'));
    } finally {
      setClaiming(false);
    }
  };

  useEffect(() => {
    fetchStatus();
  }, []);

  const statusMeta = useMemo(() => {
    switch (status?.status) {
      case 'claimable':
        return { text: t('可领取'), color: 'green' };
      case 'claimed':
        return { text: t('已领取'), color: 'blue' };
      case 'phone_required':
        return { text: t('需绑定手机号'), color: 'orange' };
      default:
        return { text: t('未启用'), color: 'grey' };
    }
  }, [status?.status, t]);

  if (!loading && !status?.enabled) {
    return null;
  }

  const currentWindow =
    status?.window_started_at && status?.window_ends_at
      ? `${timestamp2string(status.window_started_at)} - ${timestamp2string(status.window_ends_at)}`
      : '-';
  const nextClaimAt = status?.next_claim_at
    ? timestamp2string(status.next_claim_at)
    : status?.status === 'claimable'
      ? t('现在可领取')
      : '-';
  const isClaimable = status?.status === 'claimable';
  const isPhoneRequired = status?.status === 'phone_required';

  return (
    <Card className='!rounded-xl w-full'>
      <Modal
        title='Security Check'
        visible={turnstileModalVisible}
        footer={null}
        centered
        onCancel={() => {
          setTurnstileModalVisible(false);
          setTurnstileWidgetKey((v) => v + 1);
        }}
      >
        <div className='flex justify-center py-2'>
          <Turnstile
            key={turnstileWidgetKey}
            sitekey={turnstileSiteKey}
            onVerify={(token) => {
              claimWeeklyQuota(token);
            }}
            onExpire={() => {
              setTurnstileWidgetKey((v) => v + 1);
            }}
          />
        </div>
      </Modal>

      <Spin spinning={loading}>
        <div className='flex items-start justify-between gap-3'>
          <div className='flex items-center gap-3'>
            <Avatar size='small' color='amber'>
              <Gift size={16} />
            </Avatar>
            <div>
              <Typography.Text className='text-lg font-medium'>
                {t('周额度领取')}
              </Typography.Text>
              <div>
                <Text type='tertiary' size='small'>
                  {t('每 7 天可领取一次固定额度奖励')}
                </Text>
              </div>
            </div>
          </div>
          <Tag color={statusMeta.color}>{statusMeta.text}</Tag>
        </div>

        <div className='grid grid-cols-1 sm:grid-cols-2 gap-3 mt-4'>
          <div className='rounded-lg border border-[var(--semi-color-border)] p-3'>
            <Space spacing={8}>
              <Gift size={16} />
              <Text type='secondary'>{t('本次可领取')}</Text>
            </Space>
            <div className='mt-2 text-lg font-semibold'>
              {renderQuota(status?.amount || 0)}
            </div>
          </div>
          <div className='rounded-lg border border-[var(--semi-color-border)] p-3'>
            <Space spacing={8}>
              <CheckCircle size={16} />
              <Text type='secondary'>{t('累计领取')}</Text>
            </Space>
            <div className='mt-2 text-sm font-medium'>
              {t('次数')} {status?.total_claim_count || 0} ·{' '}
              {renderQuota(status?.total_claim_quota || 0)}
            </div>
          </div>
          <div className='rounded-lg border border-[var(--semi-color-border)] p-3 sm:col-span-2'>
            <Space spacing={8}>
              <Clock size={16} />
              <Text type='secondary'>{t('当前周期')}</Text>
            </Space>
            <div className='mt-2 text-sm'>{currentWindow}</div>
          </div>
          <div className='rounded-lg border border-[var(--semi-color-border)] p-3 sm:col-span-2'>
            <Space spacing={8}>
              <TimerReset size={16} />
              <Text type='secondary'>{t('下次可领取')}</Text>
            </Space>
            <div className='mt-2 text-sm'>{nextClaimAt}</div>
          </div>
        </div>

        <div className='mt-4 flex justify-end'>
          {isPhoneRequired ? (
            <Button
              icon={<Phone size={16} />}
              onClick={() => {
                window.location.href = '/console/personal';
              }}
            >
              {t('前往绑定手机号')}
            </Button>
          ) : (
            <Button
              type='primary'
              icon={<Gift size={16} />}
              loading={claiming}
              disabled={!isClaimable}
              onClick={() => claimWeeklyQuota()}
            >
              {isClaimable ? t('领取周额度') : t('本周期已领取')}
            </Button>
          )}
        </div>
      </Spin>
    </Card>
  );
};

export default WeeklyQuotaCard;
